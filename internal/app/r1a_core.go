package app

// JINI-R1A Increment 1 — safe single-file mutation CORE (internal primitives).
//
// Scope (specs/r1-safe-execution-contract.md §§2-8A, specs/frugal-control-plane-
// roadmap.md §e): reusable internal safety primitives for ONE supervised
// mutation of exactly ONE existing regular file. Nothing here is wired to the
// CLI, the native agentic loop, or toolEditFile — that integration is Commit 2.
// The direct, unmediated os.WriteFile in toolEditFile (agent_tools.go) REMAINS
// UNSAFE and is NOT reachable from this core; see r1a_arch_test.go.
//
// Deferred by contract (NOT implemented here): AUTO mutation, subprocess/shell/
// network/Git mutation, self-update, multi-file transactions, durable resume
// (R2), routing/failover (R3), verifier depth (R4), remote/UI surfaces.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// r1aSchemaVersion is the versioned, extensible record schema. New OPTIONAL
// fields never require a bump; only an incompatible change does. Later role
// routing / bounded-graph fields (parent op, role, attempt, dependency,
// verification-of, route ref) may be ADDED as optional opaque fields without a
// bump — none are required or emitted by Increment 1.
const r1aSchemaVersion = "r1a.v1"

// r1aOpClassSingleFileWrite is the ONLY operation class this core accepts.
const r1aOpClassSingleFileWrite = "file.write.single"

// r1aCapFSWriteSingle is the ONLY capability this core accepts.
const r1aCapFSWriteSingle = "fs.write.single"

// r1aMaxFileBytes bounds the single-file slice; larger targets fail closed.
const r1aMaxFileBytes = 16 << 20 // 16 MiB

// r1aMode is the execution mode. Increment 1 may mutate only in SUPERVISED;
// PLAN is read-only and AUTO mutation is deferred/denied (contract §2).
type r1aMode string

const (
	r1aModePlan       r1aMode = "plan"
	r1aModeSupervised r1aMode = "supervised"
	r1aModeAuto       r1aMode = "auto"
)

// r1aOutcome is a settled terminal outcome (contract §8). failed_partial is NOT
// here — it belongs to the later multi-file increment.
type r1aOutcome string

const (
	r1aCompletedVerified   r1aOutcome = "completed_verified"
	r1aCompletedUnverified r1aOutcome = "completed_unverified"
	r1aDenied              r1aOutcome = "denied"
	r1aRejected            r1aOutcome = "rejected"
	r1aFailedNoMutation    r1aOutcome = "failed_no_mutation"
	r1aFailedRolledBack    r1aOutcome = "failed_rolled_back"
	r1aRecoveryRequired    r1aOutcome = "recovery_required"
	r1aConflict            r1aOutcome = "conflict"
	// r1aCancelled and r1aInterrupted are part of the SETTLED outcome vocabulary
	// but are DECLARED, not implemented, in Increment 1: no lifecycle path emits
	// them, and recovering from a cancellation or an interruption (durable
	// continuation/resume) is R2 work that is intentionally not built here. They
	// are retained so the vocabulary is complete and stable.
	r1aCancelled   r1aOutcome = "cancelled"
	r1aInterrupted r1aOutcome = "interrupted"
)

func (o r1aOutcome) valid() bool {
	switch o {
	case r1aCompletedVerified, r1aCompletedUnverified, r1aDenied, r1aRejected,
		r1aFailedNoMutation, r1aFailedRolledBack, r1aRecoveryRequired,
		r1aConflict, r1aCancelled, r1aInterrupted:
		return true
	}
	return false
}

// r1aNewID returns a prefixed, cryptographically random identifier, e.g.
// "op_<32 hex>". Stable and unique per operation/run (contract §4, roadmap §e).
func r1aNewID(prefix string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("r1a: identifier generation failed: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(b[:]), nil
}

// r1aValidID reports whether id is a well-formed prefix + 32 hex chars.
func r1aValidID(prefix, id string) bool {
	want := prefix + "_"
	if !strings.HasPrefix(id, want) {
		return false
	}
	rest := id[len(want):]
	if len(rest) != 32 {
		return false
	}
	_, err := hex.DecodeString(rest)
	return err == nil
}

// r1aHashBytes returns a stable content hash reference.
func r1aHashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// r1aOperationRequest is the smallest versioned single-file mutation request.
type r1aOperationRequest struct {
	SchemaVersion string   `json:"schema_version"`
	OperationID   string   `json:"operation_id"`
	RunID         string   `json:"run_id"`
	SessionID     string   `json:"session_id,omitempty"`
	Mode          r1aMode  `json:"mode"`
	OpClass       string   `json:"op_class"`
	Capability    string   `json:"capability"`
	WorkspaceRoot string   `json:"-"` // absolute path; never persisted verbatim
	WorkspaceID   string   `json:"workspace_id"`
	RelTarget     string   `json:"rel_target"`
	WriteScope    []string `json:"write_scope"`
	BeforeHash    string   `json:"before_hash"`
	AfterHash     string   `json:"after_hash"` // proposed
}

// r1aPolicyDecision is the structured allow/deny with a reason.
type r1aPolicyDecision struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason"`
}

// r1aEvaluatePolicy fails closed on any unknown mode, class, capability,
// malformed id, absent/absolute/traversal/root-equivalent target, or a write
// scope that is anything other than exactly the one declared target.
func r1aEvaluatePolicy(req r1aOperationRequest) r1aPolicyDecision {
	if req.SchemaVersion != r1aSchemaVersion {
		return r1aPolicyDecision{false, "unsupported schema version"}
	}
	if req.Mode != r1aModeSupervised {
		return r1aPolicyDecision{false, "only supervised mode may mutate in R1A Increment 1"}
	}
	if req.OpClass != r1aOpClassSingleFileWrite {
		return r1aPolicyDecision{false, "unknown operation class"}
	}
	if req.Capability != r1aCapFSWriteSingle {
		return r1aPolicyDecision{false, "unknown capability"}
	}
	if !r1aValidID("op", req.OperationID) || !r1aValidID("run", req.RunID) {
		return r1aPolicyDecision{false, "malformed operation or run identifier"}
	}
	t := strings.TrimSpace(req.RelTarget)
	if t == "" {
		return r1aPolicyDecision{false, "absent target"}
	}
	if filepath.IsAbs(t) {
		return r1aPolicyDecision{false, "absolute target not allowed"}
	}
	if r1aHasTraversal(t) {
		return r1aPolicyDecision{false, "path traversal not allowed"}
	}
	if clean := filepath.Clean(t); clean == "." || clean == string(filepath.Separator) {
		return r1aPolicyDecision{false, "root-equivalent target not allowed"}
	}
	if len(req.WriteScope) != 1 || req.WriteScope[0] != req.RelTarget {
		return r1aPolicyDecision{false, "write scope must be exactly the one declared target"}
	}
	return r1aPolicyDecision{true, "single-file supervised mutation permitted"}
}

// r1aHasTraversal reports whether any path component is "..".
func r1aHasTraversal(p string) bool {
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

// r1aApprovalBinding is the EXACT binding a caller must supply before the core
// mutates (contract §2: approval binds to the exact operation). Commit 1 does
// NOT obtain approval interactively; tests construct this directly and Commit 2
// will produce it from a real approval flow. Absent/malformed/mismatched
// bindings fail closed; Approved=false on an otherwise-matching binding is a
// user rejection.
type r1aApprovalBinding struct {
	OperationID string   `json:"operation_id"`
	RelTarget   string   `json:"rel_target"`
	BeforeHash  string   `json:"before_hash"`
	AfterHash   string   `json:"after_hash"`
	WriteScope  []string `json:"write_scope"`
	Approved    bool     `json:"approved"`
}

// r1aBindingWellFormed checks structural completeness, independent of a request.
func r1aBindingWellFormed(b r1aApprovalBinding) bool {
	return r1aValidID("op", b.OperationID) &&
		strings.TrimSpace(b.RelTarget) != "" &&
		strings.HasPrefix(b.BeforeHash, "sha256:") &&
		strings.HasPrefix(b.AfterHash, "sha256:") &&
		len(b.WriteScope) == 1
}

// r1aBindingMatches checks the binding is bound to this exact operation.
func r1aBindingMatches(req r1aOperationRequest, b r1aApprovalBinding) bool {
	return b.OperationID == req.OperationID &&
		b.RelTarget == req.RelTarget &&
		b.BeforeHash == req.BeforeHash &&
		b.AfterHash == req.AfterHash &&
		len(b.WriteScope) == 1 && len(req.WriteScope) == 1 &&
		b.WriteScope[0] == req.WriteScope[0]
}

// r1aApprovalRef returns a stable, non-secret reference to an approval binding
// (a hash over its bound fields). It is recorded on events once a well-formed,
// matching binding exists, so records can truthfully cite the approval evidence
// without storing content or secrets.
func r1aApprovalRef(b r1aApprovalBinding) string {
	canon := strings.Join([]string{
		b.OperationID,
		b.RelTarget,
		b.BeforeHash,
		b.AfterHash,
		strings.Join(b.WriteScope, ","),
		fmt.Sprintf("approved=%t", b.Approved),
	}, "\n")
	return "approval:" + r1aHashBytes([]byte(canon))
}

// r1aEvent is one append-only, ordered operation event (roadmap §e). Terminal
// output is a projection of these; events carry references/hashes, never file
// content or secrets.
type r1aEvent struct {
	SchemaVersion string     `json:"schema_version"`
	OperationID   string     `json:"operation_id"`
	RunID         string     `json:"run_id"`
	Sequence      int64      `json:"sequence"`
	Timestamp     string     `json:"timestamp"`
	EventType     string     `json:"event_type"`
	Mode          r1aMode    `json:"mode"`
	OpClass       string     `json:"op_class"`
	RelTarget     string     `json:"rel_target"`
	WriteScope    []string   `json:"write_scope"`
	PolicyReason  string     `json:"policy_reason,omitempty"`
	BeforeHash    string     `json:"before_hash,omitempty"`
	AfterHash     string     `json:"after_hash,omitempty"`
	SnapshotRef   string     `json:"snapshot_ref,omitempty"`
	VerifyRef     string     `json:"verify_ref,omitempty"`
	ApprovalRef   string     `json:"approval_ref,omitempty"`
	Outcome       r1aOutcome `json:"outcome,omitempty"`
	Detail        string     `json:"detail,omitempty"`
}

// r1aEventLog is an in-memory append-only ordered log with a monotonic sequence.
type r1aEventLog struct {
	mu     sync.Mutex
	seq    int64
	events []r1aEvent
}

// append assigns the next monotonic sequence and stores the event (never
// rewrites an earlier one). Detail is redacted at the call site.
func (l *r1aEventLog) append(e r1aEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq++
	e.Sequence = l.seq
	l.events = append(l.events, e)
}

// snapshotEvents returns an ordered copy of the events recorded so far.
func (l *r1aEventLog) snapshotEvents() []r1aEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]r1aEvent, len(l.events))
	copy(out, l.events)
	return out
}

// r1aRedact masks credential-shaped substrings so no secret reaches a record.
// Events are already secret-free by construction (hashes/paths/refs only); this
// is defence in depth for any Detail string.
func r1aRedact(s string) string {
	for _, re := range r1aSecretPatterns {
		s = re.ReplaceAllString(s, "[REDACTED]")
	}
	return s
}

// nowUTC is a seam so tests can assert ordering deterministically if needed.
var nowUTC = func() time.Time { return time.Now().UTC() }
