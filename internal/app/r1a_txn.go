package app

// R1A single-file transaction lifecycle (contract §§7-8): capture → propose →
// approve-bound apply (revalidate before-state → atomic in-root replace →
// integrity guard → optional deterministic verify) → guarded rollback. No Git
// mutation, no stash/reset/clean, no subprocess.
//
// CONTAINMENT (owner decision, Option 2): no $0/stdlib backend establishes the
// contract's required guarantee that an approved single-file replacement cannot
// SILENTLY OVERWRITE a concurrent external writer. POSIX rename is atomic
// against torn writes but is NOT a compare-and-swap, and there is no portable,
// dependency-free mandatory lock. That guarantee is an R1 mutation-safety
// responsibility (NOT an R2 durable-continuation feature); because it is
// currently UNMET, production Apply and Rollback FAIL CLOSED before creating any
// temp file, replacing/truncating/restoring, or changing target metadata. The
// private replace/restore primitives below remain only so the safety machinery
// can be exercised by same-package tests; that experimental path is NOT
// production-supported mutation and NOTHING (flag, env var, trust grant, repo
// config) re-enables it. Enabling customer mutation requires a future gated
// backend plus a reviewed production call site (Commit 2+).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/maridlabsai/jini/agentloop"
)

// r1aProductionMutationEnabled is the production mutation switch. It is a
// compile-time constant false; no flag, environment variable, trust grant, or
// repository configuration can change it. Production Apply/Rollback pass this
// constant, so they always fail closed. Only a future reviewed source change
// (Commit 2+) that supplies an enforceable backend could flip the production
// path — never a runtime input.
const r1aProductionMutationEnabled = false

// r1aMutationBackendReason is the settled, non-secret reason production refuses
// to mutate: there is no supported backend that guarantees an approved
// single-file replacement cannot silently overwrite a concurrent external
// writer. This is reported with the settled outcome r1aDenied.
const r1aMutationBackendReason = "no supported mutation backend establishes the required concurrent-overwrite guarantee"

// errR1AConflictBeforeRename is returned by the experimental replace primitive
// when a re-read performed just before the atomic rename finds the target no
// longer matches the expected current state. This re-read DETECTS a change that
// occurred up to the moment of the re-read and converts it into a fail-closed
// conflict instead of a silent overwrite. It does NOT cover the gap between that
// re-read and the rename syscall: a write landing in that final-check→rename gap
// is still silently overwritten. A content re-read followed by rename is NOT a
// content-conditional replacement (no filesystem compare-and-swap exists here),
// and a passing Go race detector does not establish one. That irreducible gap is
// precisely why stdlib rename does not satisfy the R1 no-silent-overwrite
// guarantee and why production mutation is contained. This is an unmet R1
// mutation-safety limitation, not a concern transferred to R2.
var errR1AConflictBeforeRename = errors.New("r1a: target changed before the final re-read; conflict")

type r1aTransaction struct {
	req               r1aOperationRequest
	target            *r1aResolvedTarget
	storeDir          string
	log               *r1aEventLog
	snapshotRef       string
	approvalRef       string
	afterBytes        []byte
	producedAfterHash string
	applied           bool
	finalOutcome      r1aOutcome
	persistErr        error
}

// r1aPrepare resolves and snapshots one safe target, returning a transaction
// whose request carries stable ids and the captured before-hash. It does NOT
// mutate; Propose then Apply (with an approval binding) drive the lifecycle
// (which, in production, fails closed at the containment guard).
//
// r1aPrepare is the protected entry point: it enforces the Go toolchain floor,
// canonicalizes and captures the workspace root itself (callers are not trusted
// to have done so), derives workspace identity and the protected-store location
// from that captured root, and validates the store resolves outside the
// workspace before creating any protected state.
//
// UNRESOLVED (integration blockers, NOT corrected in this step):
//   - F-3: target/component identity and file type may change between resolution
//     (r1aResolveTarget's per-component Lstat) and the subsequent open/read. A
//     second content-hash check does not resolve symlink, hard-link, special-file
//     or mount/boundary identity races.
//   - F-4: a store ancestor substitution can redirect directory creation or chmod
//     before the post-creation boundary check rejects it (r1aEnsureProtectedStore).
//
// These are read/prepare-side risks only; production never mutates the target
// (containment), and r1aPrepare has no production caller today. "Currently
// unwired" is containment, not a fix — closing F-3/F-4 is required before any
// application integration (Commit 2) and needs its own authorization. No
// recursive RemoveAll on an uncertain/attacker-influenced path is adopted here.
func r1aPrepare(workspaceRoot, rel string) (*r1aTransaction, error) {
	// M1: enforce the patched-toolchain floor before any operation that depends
	// on the protected filesystem primitives (os.Root confinement assumes the
	// GO-2026-4970 fix). This runs at the real entry point, not only in a helper.
	if err := r1aEnforceGoFloor(runtime.Version()); err != nil {
		return nil, err
	}
	// C: establish the security prerequisite here rather than trusting the caller.
	// Canonicalization yields a stable identity for boundary and store decisions;
	// it does NOT by itself eliminate filesystem races (those are handled by
	// per-operation revalidation and, ultimately, by failing closed).
	canonicalRoot, err := r1aCanonicalRoot(workspaceRoot)
	if err != nil {
		return nil, err
	}
	root, err := r1aOpenWorkspace(canonicalRoot)
	if err != nil {
		return nil, err
	}
	target, err := r1aResolveTarget(root, rel)
	if err != nil {
		return nil, err
	}
	opID, err := r1aNewID("op")
	if err != nil {
		return nil, err
	}
	runID, err := r1aNewID("run")
	if err != nil {
		return nil, err
	}
	wsID := r1aWorkspaceID(canonicalRoot)
	slashTarget := filepath.ToSlash(target.rel)
	req := r1aOperationRequest{
		SchemaVersion: r1aSchemaVersion,
		OperationID:   opID,
		RunID:         runID,
		Mode:          r1aModeSupervised,
		OpClass:       r1aOpClassSingleFileWrite,
		Capability:    r1aCapFSWriteSingle,
		WorkspaceRoot: canonicalRoot,
		WorkspaceID:   wsID,
		RelTarget:     slashTarget,
		WriteScope:    []string{slashTarget},
		BeforeHash:    target.beforeHash,
	}
	// C: validate the store boundary and create protected state in one place,
	// deriving the store location from the captured canonical root.
	storeDir, err := r1aEnsureProtectedStore(canonicalRoot, wsID, runID)
	if err != nil {
		return nil, err
	}
	preImage, err := r1aReadTarget(root, target.rel)
	if err != nil {
		return nil, err
	}
	// H1: the resolve read and this snapshot read are separate reads. If the
	// target was substituted between them the snapshot would be inconsistent with
	// the recorded before-hash, so fail closed rather than persist a mismatched
	// pre-image. (Detection, not race elimination.)
	if r1aHashBytes(preImage) != target.beforeHash {
		return nil, fmt.Errorf("r1a: target changed during preparation; fail closed")
	}
	snapRef, err := r1aSaveSnapshot(storeDir, preImage, r1aSnapshotMeta{
		SchemaVersion: r1aSchemaVersion,
		RelTarget:     slashTarget,
		BeforeHash:    target.beforeHash,
		Mode:          target.mode,
		Size:          target.size,
	})
	if err != nil {
		return nil, err
	}
	txn := &r1aTransaction{req: req, target: target, storeDir: storeDir, log: &r1aEventLog{}, snapshotRef: snapRef}
	txn.emit("prepared", "", "captured pre-image and before-hash", "")
	return txn, nil
}

// Propose sets the full proposed after-content and its hash.
func (txn *r1aTransaction) Propose(afterBytes []byte) error {
	if int64(len(afterBytes)) > r1aMaxFileBytes {
		return fmt.Errorf("r1a: proposed content exceeds the single-file size bound")
	}
	txn.afterBytes = append([]byte(nil), afterBytes...)
	txn.req.AfterHash = r1aHashBytes(txn.afterBytes)
	txn.emit("proposed", "", "proposed after-content hashed", txn.req.AfterHash)
	return nil
}

// ApprovalBinding constructs the exact binding for this operation. In Commit 1
// this is used by tests and later by Commit 2's approval flow; the core never
// fabricates approval.
func (txn *r1aTransaction) ApprovalBinding(approved bool) r1aApprovalBinding {
	return r1aApprovalBinding{
		OperationID: txn.req.OperationID,
		RelTarget:   txn.req.RelTarget,
		BeforeHash:  txn.req.BeforeHash,
		AfterHash:   txn.req.AfterHash,
		WriteScope:  append([]string(nil), txn.req.WriteScope...),
		Approved:    approved,
	}
}

// Apply is the production mutation entry point. There is no supported mutation
// backend, so it FAILS CLOSED first — before any cached/state-derived success
// return, any target I/O, and any experimental write helper — and never reads or
// writes the target. A transaction previously mutated by an experimental fixture
// cannot make Apply report success.
func (txn *r1aTransaction) Apply(binding r1aApprovalBinding, verifier r1aVerifier) (r1aOutcome, error) {
	return txn.applyBounded(binding, verifier, r1aProductionMutationEnabled, nil)
}

// Rollback is the production rollback entry point; like Apply it fails closed
// first (no supported backend performs a restore) and never reads or writes the
// target.
func (txn *r1aTransaction) Rollback() (r1aOutcome, error) {
	return txn.rollbackBounded(r1aProductionMutationEnabled, nil)
}

// applyBounded runs the apply lifecycle. mutationEnabled is false on every
// production path (Apply passes the compile-time constant) and true only on the
// experimental same-package test path. beforeRename is a test-only injection,
// passed by value — there is no production-global hook.
//
// CONTAINMENT: when mutationEnabled is false the function denies FIRST — before
// the idempotency shortcut (so a cached/experimental success is never returned),
// before any target I/O, and before any write helper. The read-only input
// validation and target revalidation live on the experimental path, because they
// are only meaningful when a mutation could actually occur.
func (txn *r1aTransaction) applyBounded(binding r1aApprovalBinding, verifier r1aVerifier, mutationEnabled bool, beforeRename func()) (r1aOutcome, error) {
	if !mutationEnabled {
		txn.emitReason("backend.unsupported", r1aDenied, r1aMutationBackendReason, r1aMutationBackendReason, "")
		return txn.settle(r1aDenied, fmt.Errorf("r1a: %s", r1aMutationBackendReason))
	}

	// ----- EXPERIMENTAL private mutation path (same-package tests only) -----
	if txn.applied {
		// Idempotent retry: return the ACTUAL outcome of the prior apply on this
		// txn. If that apply's evidence failed to persist, surface the failure
		// again (never a nil error); never replay the mutation.
		if txn.persistErr != nil {
			return txn.finalOutcome, fmt.Errorf("r1a: prior outcome %s but evidence persistence failed: %w", txn.finalOutcome, txn.persistErr)
		}
		return txn.finalOutcome, nil
	}
	req := txn.req
	if req.AfterHash == "" {
		txn.emit("proposal.missing", r1aFailedNoMutation, "no proposed content", "")
		return txn.settle(r1aFailedNoMutation, fmt.Errorf("r1a: Propose must be called before Apply"))
	}
	if dec := r1aEvaluatePolicy(req); !dec.Allow {
		txn.emitReason("policy.denied", r1aDenied, dec.Reason, dec.Reason, "")
		return txn.settle(r1aDenied, fmt.Errorf("r1a: policy denied: %s", dec.Reason))
	}
	if !r1aBindingWellFormed(binding) || !r1aBindingMatches(req, binding) {
		txn.emit("approval.invalid", r1aFailedNoMutation, "approval binding absent or malformed", "")
		return txn.settle(r1aFailedNoMutation, fmt.Errorf("r1a: approval binding absent or malformed"))
	}
	// L2: a well-formed, matching binding IS approval evidence; record a
	// non-secret reference so subsequent events carry it truthfully. The ref
	// identifies the binding data; it is NOT independent proof of human consent.
	txn.approvalRef = r1aApprovalRef(binding)
	if !binding.Approved {
		txn.emit("approval.rejected", r1aRejected, "operation rejected at approval", "")
		return txn.settle(r1aRejected, nil)
	}
	// Re-read and compare the target against the approved before-state. This
	// DETECTS a change up to this read and fails closed as a conflict; it is not
	// race-free confinement and is not described as one.
	cur, err := r1aReadTarget(txn.target.root, txn.target.rel)
	if err != nil {
		txn.emit("revalidate.error", r1aFailedNoMutation, "cannot re-read target before write", "")
		return txn.settle(r1aFailedNoMutation, err)
	}
	if r1aHashBytes(cur) != req.BeforeHash {
		txn.emit("conflict.before", r1aConflict, "target changed after preview; not overwriting", "")
		return txn.settle(r1aConflict, nil)
	}
	if err := txn.writeAtomic(beforeRename); err != nil {
		if errors.Is(err, errR1AConflictBeforeRename) {
			txn.emit("conflict.prerename", r1aConflict, "target changed before the final re-read; not overwriting", "")
			return txn.settle(r1aConflict, nil)
		}
		txn.emit("write.error", r1aFailedNoMutation, "atomic write failed before replacement completed", "")
		return txn.settle(r1aFailedNoMutation, err)
	}
	txn.applied = true
	txn.producedAfterHash = req.AfterHash
	txn.emit("write.applied", "", "atomic replacement complete", req.AfterHash)

	// Integrity guard (always): the file now equals EXACTLY the approved bytes.
	now, err := r1aReadTarget(txn.target.root, txn.target.rel)
	if err != nil || r1aHashBytes(now) != req.AfterHash {
		out, _ := txn.rollbackBounded(mutationEnabled, beforeRename)
		txn.emit("integrity.failed", out, "post-write integrity check failed", "")
		return txn.settle(out, fmt.Errorf("r1a: post-write integrity check failed"))
	}

	// Optional deterministic verifier. With none, success is completed_unverified
	// (contract §8) — never a false verified claim. git diff --check is a future
	// verifier (Commit 2) and is intentionally not run here.
	if verifier == nil {
		txn.emit("verify.none", r1aCompletedUnverified, "no deterministic verifier configured (git diff --check deferred)", "")
		return txn.settle(r1aCompletedUnverified, nil)
	}
	ok, verr := verifier.Verify(txn)
	if verr != nil || !ok {
		out, _ := txn.rollbackBounded(mutationEnabled, beforeRename)
		txn.emit("verify.failed", out, "verification failed; rolled back", "")
		return txn.settle(out, verr)
	}
	txn.emit("verify.passed", r1aCompletedVerified, "verification passed", "")
	return txn.settle(r1aCompletedVerified, nil)
}

// rollbackBounded restores the pre-image ONLY when the current target still
// equals the bytes this transaction produced. When mutationEnabled is false
// (production) it denies FIRST, before any target I/O or restore: a transaction
// previously mutated by an experimental fixture is refused with
// recovery_required (never a restore), and an un-applied transaction reports
// failed_no_mutation. If another actor changed the target it returns
// recovery_required, never overwriting that work.
func (txn *r1aTransaction) rollbackBounded(mutationEnabled bool, beforeRename func()) (r1aOutcome, error) {
	if !mutationEnabled {
		if !txn.applied {
			return r1aFailedNoMutation, nil // nothing was applied; no restore to deny
		}
		txn.emitReason("rollback.backend.unsupported", r1aRecoveryRequired, r1aMutationBackendReason, r1aMutationBackendReason, "")
		return r1aRecoveryRequired, fmt.Errorf("r1a: %s", r1aMutationBackendReason)
	}
	// ----- EXPERIMENTAL restore path (same-package tests only) -----
	if !txn.applied {
		return r1aFailedNoMutation, nil
	}
	cur, err := r1aReadTarget(txn.target.root, txn.target.rel)
	if err != nil {
		txn.emit("rollback.read.error", r1aRecoveryRequired, "cannot read target for rollback", "")
		return r1aRecoveryRequired, err
	}
	if r1aHashBytes(cur) != txn.producedAfterHash {
		txn.emit("rollback.conflict", r1aRecoveryRequired, "target changed since our write; not overwriting concurrent work", "")
		return r1aRecoveryRequired, nil
	}
	if err := txn.restorePreImage(beforeRename); err != nil {
		if errors.Is(err, errR1AConflictBeforeRename) {
			txn.emit("rollback.conflict", r1aRecoveryRequired, "target changed immediately before restore; not overwriting concurrent work", "")
			return r1aRecoveryRequired, nil
		}
		txn.emit("rollback.restore.error", r1aRecoveryRequired, "pre-image restore failed", "")
		return r1aRecoveryRequired, err
	}
	restored, err := r1aReadTarget(txn.target.root, txn.target.rel)
	if err != nil || r1aHashBytes(restored) != txn.req.BeforeHash {
		txn.emit("rollback.verify.error", r1aRecoveryRequired, "restored state does not match pre-image", "")
		return r1aRecoveryRequired, fmt.Errorf("r1a: rollback verification failed")
	}
	txn.applied = false
	txn.emit("rollback.done", r1aFailedRolledBack, "rolled back to pre-image", txn.req.BeforeHash)
	return r1aFailedRolledBack, nil
}

// CleanupSuccess removes artifacts, permitted only after a VERIFIED-completion
// success whose evidence was durably persisted. Two independent gates apply:
// persistErr must be clear (checked here), AND finalOutcome must be
// completed_verified (enforced at the r1aCleanupSuccess deletion boundary, which
// refuses completed_unverified per contract §7/§8). A success-shaped or cached
// outcome is never sufficient: on refusal the snapshot/preimage/recovery
// evidence is left intact, and a prior persistence failure is never silently
// cleared by a cleanup attempt.
func (txn *r1aTransaction) CleanupSuccess() error {
	if txn.persistErr != nil {
		return fmt.Errorf("r1a: refusing cleanup; operation evidence was not durably persisted: %w", txn.persistErr)
	}
	return r1aCleanupSuccess(txn.storeDir, txn.finalOutcome)
}

// writeAtomic writes a same-directory temp file and atomically renames it over
// the target, all through the confined root. beforeRename (test-only, may be
// nil) is invoked BEFORE the final re-read below — not in the final-read→rename
// gap — so tests can inject a pre-check change and confirm it is detected.
func (txn *r1aTransaction) writeAtomic(beforeRename func()) error {
	tmpRel, err := txn.tempName(".r1atmp")
	if err != nil {
		return err
	}
	// Expect the target to still equal the approved before-state right up to the
	// rename; fail closed if it changed.
	return txn.writeBytesAtomic(tmpRel, txn.afterBytes, txn.req.BeforeHash, beforeRename)
}

// restorePreImage rewrites the stored pre-image atomically over the target.
func (txn *r1aTransaction) restorePreImage(beforeRename func()) error {
	data, err := r1aReadPreImage(txn.storeDir)
	if err != nil {
		return err
	}
	tmpRel, err := txn.tempName(".r1arb")
	if err != nil {
		return err
	}
	// Expect the target to still equal the bytes we produced right up to the
	// restore rename; fail closed (recovery_required) if a concurrent actor
	// changed it, so rollback never overwrites intervening user work.
	return txn.writeBytesAtomic(tmpRel, data, txn.producedAfterHash, beforeRename)
}

func (txn *r1aTransaction) tempName(suffix string) (string, error) {
	id, err := r1aNewID("tmp")
	if err != nil {
		return "", err
	}
	frag := strings.TrimPrefix(id, "tmp_")
	dir := filepath.Dir(txn.target.rel)
	base := filepath.Base(txn.target.rel)
	return filepath.Join(dir, "."+base+"."+frag+suffix), nil
}

func (txn *r1aTransaction) writeBytesAtomic(tmpRel string, data []byte, expectCurrent string, beforeRename func()) error {
	root := txn.target.root
	f, err := root.OpenFile(tmpRel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, txn.target.mode)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		_ = root.Remove(tmpRel)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		_ = root.Remove(tmpRel)
		return err
	}
	if err := f.Close(); err != nil {
		_ = root.Remove(tmpRel)
		return err
	}
	// L3: only the regular-file PERMISSION BITS are preserved (via the temp's
	// create mode and this Chmod). Ownership, timestamps, ACLs, and extended
	// attributes are NOT preserved. This is a path-based Chmod, which is NOT
	// inherently symlink-race-safe; the mitigations here are the unpredictable
	// in-root temp name and os.Root confinement. Root-relative confinement, an
	// unpredictable name, and path-based Chmod are distinct properties, not
	// interchangeable guarantees.
	if err := root.Chmod(tmpRel, txn.target.mode); err != nil {
		_ = root.Remove(tmpRel)
		return err
	}
	// Test-only injection point: fires BEFORE the final re-read below so a test
	// can inject a pre-check change. It does NOT exercise the final-read→rename
	// gap.
	if beforeRename != nil {
		beforeRename()
	}
	// Final re-read: atomic rename is NOT a compare-and-swap, so re-read and fail
	// closed if the target changed. This DETECTS a change that occurred up to this
	// read; it does NOT cover the gap between this read and the rename syscall — a
	// write landing in that gap is still silently overwritten. A content re-read
	// followed by rename is not content-conditional replacement, which is why
	// stdlib rename does not satisfy the no-silent-overwrite guarantee.
	if expectCurrent != "" {
		cur, rerr := r1aReadTarget(root, txn.target.rel)
		if rerr != nil {
			_ = root.Remove(tmpRel)
			return rerr
		}
		if r1aHashBytes(cur) != expectCurrent {
			_ = root.Remove(tmpRel)
			return errR1AConflictBeforeRename
		}
	}
	if err := root.Rename(tmpRel, txn.target.rel); err != nil {
		_ = root.Remove(tmpRel)
		return err
	}
	return nil
}

// settle records the final outcome, persists the manifest and ordered events,
// and SURFACES any persistence failure. M2: a success is never reported cleanly
// when the durable evidence could not be written; the persistence error is
// merged into the returned error and retained on the transaction.
func (txn *r1aTransaction) settle(outcome r1aOutcome, opErr error) (r1aOutcome, error) {
	txn.finalOutcome = outcome
	if perr := r1aPersistManifest(txn.storeDir, txn.req, txn.log.snapshotEvents(), outcome); perr != nil {
		txn.persistErr = perr
		if opErr != nil {
			return outcome, fmt.Errorf("%w; evidence persistence also failed: %v", opErr, perr)
		}
		return outcome, fmt.Errorf("r1a: outcome %s reached but evidence persistence failed: %w", outcome, perr)
	}
	return outcome, opErr
}

func (txn *r1aTransaction) emit(eventType string, outcome r1aOutcome, detail, afterHash string) {
	txn.emitReason(eventType, outcome, "", detail, afterHash)
}

func (txn *r1aTransaction) emitReason(eventType string, outcome r1aOutcome, policyReason, detail, afterHash string) {
	txn.log.append(r1aEvent{
		SchemaVersion: r1aSchemaVersion,
		OperationID:   txn.req.OperationID,
		RunID:         txn.req.RunID,
		Timestamp:     nowUTC().Format(time.RFC3339Nano),
		EventType:     eventType,
		Mode:          txn.req.Mode,
		OpClass:       txn.req.OpClass,
		RelTarget:     txn.req.RelTarget,
		WriteScope:    txn.req.WriteScope,
		PolicyReason:  policyReason,
		BeforeHash:    txn.req.BeforeHash,
		AfterHash:     afterHash,
		SnapshotRef:   txn.snapshotRef,
		ApprovalRef:   txn.approvalRef,
		Outcome:       outcome,
		Detail:        r1aRedact(detail),
	})
}

// r1aVerifier is a deterministic, in-process minimum verifier. No arbitrary
// command runner is introduced in Commit 1.
type r1aVerifier interface {
	Verify(txn *r1aTransaction) (bool, error)
}

// r1aHashVerifier confirms the on-disk target equals the approved after-hash.
type r1aHashVerifier struct{}

func (r1aHashVerifier) Verify(txn *r1aTransaction) (bool, error) {
	data, err := r1aReadTarget(txn.target.root, txn.target.rel)
	if err != nil {
		return false, err
	}
	return r1aHashBytes(data) == txn.req.AfterHash, nil
}

// r1aCheckpointer implements agentloop.Checkpointer for ONE declared target by
// saving a labeled pre-image into the protected store. It is a best-effort seam
// (the interface's own contract). Commit 1 does NOT register it or wire it to
// the native loop; that decision is Commit 2. The interface is not widened.
type r1aCheckpointer struct {
	root  *os.Root
	rel   string
	store string
}

func newR1ACheckpointer(txn *r1aTransaction) *r1aCheckpointer {
	return &r1aCheckpointer{root: txn.target.root, rel: txn.target.rel, store: txn.storeDir}
}

func (c *r1aCheckpointer) Checkpoint(label string) error {
	data, err := r1aReadTarget(c.root, c.rel)
	if err != nil {
		return err
	}
	return r1aWriteProtected(filepath.Join(c.store, "snapshot", "checkpoint-"+r1aSanitizeLabel(label)), data)
}

// Compile-time proof the seam is implemented without widening the interface.
var _ agentloop.Checkpointer = (*r1aCheckpointer)(nil)
