package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func r1aPrepProposed(t *testing.T, ws, rel string, orig, after []byte, mode os.FileMode) *r1aTransaction {
	t.Helper()
	r1aWriteFile(t, ws, rel, orig, mode)
	txn, err := r1aPrepare(ws, rel)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if err := txn.Propose(after); err != nil {
		t.Fatalf("propose: %v", err)
	}
	return txn
}

// r1aApplyExperimental drives the EXPERIMENTAL private mutation primitives via
// the same-package unexported path. This is NOT production-supported mutation:
// production Apply refuses (no supported backend). beforeRename, when non-nil,
// injects a change BEFORE the final re-read (not in the final-read→rename gap),
// so a test can confirm a pre-check change is detected. It exists only to
// exercise the safety machinery under test.
func r1aApplyExperimental(t *testing.T, txn *r1aTransaction, b r1aApprovalBinding, v r1aVerifier, beforeRename func()) (r1aOutcome, error) {
	t.Helper()
	return txn.applyBounded(b, v, true, beforeRename)
}

// r1aRollbackExperimental drives the experimental private rollback path.
func r1aRollbackExperimental(t *testing.T, txn *r1aTransaction, beforeRename func()) (r1aOutcome, error) {
	t.Helper()
	return txn.rollbackBounded(true, beforeRename)
}

func r1aTargetTempFiles(t *testing.T, ws, rel string) []string {
	t.Helper()
	dir := filepath.Dir(filepath.Join(ws, rel))
	base := filepath.Base(rel)
	m, err := filepath.Glob(filepath.Join(dir, "."+base+".*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	return m
}

// ---------------------------------------------------------------------------
// CONTAINMENT: production Apply/Rollback fail closed (no supported backend).
// ---------------------------------------------------------------------------

func TestR1AProductionApplyFailsClosedUnsupportedBackend(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o640)
	before, err := os.Stat(filepath.Join(ws, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}

	out, err := txn.Apply(txn.ApprovalBinding(true), nil)
	if err == nil {
		t.Fatal("production Apply must fail closed with an error")
	}
	if out != r1aDenied {
		t.Fatalf("want denied (unsupported backend), got %s", out)
	}
	if !strings.Contains(err.Error(), "backend") {
		t.Fatalf("error must name the unsupported backend: %v", err)
	}
	if txn.applied {
		t.Fatal("transaction must not be marked applied")
	}
	// Target bytes and permission bits unchanged.
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("target bytes changed: %q", got)
	}
	after, err := os.Stat(filepath.Join(ws, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("target metadata changed: mode %v→%v mtime %v→%v", before.Mode(), after.Mode(), before.ModTime(), after.ModTime())
	}
	// No target-directory temp file created.
	if tmps := r1aTargetTempFiles(t, ws, "f.txt"); len(tmps) != 0 {
		t.Fatalf("denied mutation created temp files: %v", tmps)
	}
	// The denial is recorded truthfully.
	sawDenied := false
	for _, e := range txn.log.snapshotEvents() {
		if e.EventType == "backend.unsupported" && e.Outcome == r1aDenied {
			sawDenied = true
		}
		if e.EventType == "write.applied" {
			t.Fatal("no write may be recorded under containment")
		}
	}
	if !sawDenied {
		t.Fatal("expected a recorded backend.unsupported denial")
	}
}

func TestR1AProductionRollbackFailsClosed(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	// Apply is contained, so nothing is applied; Rollback has nothing to undo.
	if _, err := txn.Apply(txn.ApprovalBinding(true), nil); err == nil {
		t.Fatal("apply should have failed closed")
	}
	out, err := txn.Rollback()
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if out != r1aFailedNoMutation {
		t.Fatalf("want failed_no_mutation (nothing applied), got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("rollback changed target: %q", got)
	}
	if tmps := r1aTargetTempFiles(t, ws, "f.txt"); len(tmps) != 0 {
		t.Fatalf("rollback created temp files: %v", tmps)
	}
}

func TestR1ANoEnvOrConfigReenablesMutation(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	// Set assorted configuration-shaped inputs; none may re-enable mutation.
	t.Setenv("JINI_MODE", "autopilot")
	t.Setenv("JINI_UNSAFE", "1")
	t.Setenv("JINI_ALLOW_WRITE", "true")
	t.Setenv("JINI_TRUST", "autonomous")
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := txn.Apply(txn.ApprovalBinding(true), nil)
	if err == nil || out != r1aDenied {
		t.Fatalf("configuration must not re-enable mutation; got out=%s err=%v", out, err)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("target changed despite containment: %q", got)
	}
}

// ---------------------------------------------------------------------------
// Input/validation handling. Production Apply denies first (no supported
// backend), so these outcomes (rejected / failed_no_mutation / conflict) are
// only reachable on the EXPERIMENTAL path, where a mutation could actually
// occur. They are exercised there, not on the production entry.
// ---------------------------------------------------------------------------

func TestR1AExperimentalApprovalRejectedDoesNotWrite(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(false), nil, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aRejected {
		t.Fatalf("want rejected, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("file should be untouched on rejection: %q", got)
	}
}

func TestR1AExperimentalMalformedBindingFailsClosed(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	bad := txn.ApprovalBinding(true)
	bad.AfterHash = "sha256:wrong"
	out, err := r1aApplyExperimental(t, txn, bad, nil, nil)
	if err == nil {
		t.Fatal("malformed binding must error")
	}
	if out != r1aFailedNoMutation {
		t.Fatalf("want failed_no_mutation, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatal("file should be untouched")
	}
}

func TestR1AExperimentalApplyWithoutProposeFailsNoMutation(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "f.txt", []byte("old"), 0o644)
	txn, err := r1aPrepare(ws, "f.txt")
	if err != nil {
		t.Fatal(err)
	}
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err == nil {
		t.Fatal("apply without propose must error")
	}
	if out != r1aFailedNoMutation {
		t.Fatalf("want failed_no_mutation, got %s", out)
	}
}

func TestR1AExperimentalBeforeHashConflictBlocksWrite(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	// Concurrent external change AFTER preview, BEFORE apply.
	r1aWriteFile(t, ws, "f.txt", []byte("someone-else"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aConflict {
		t.Fatalf("want conflict, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "someone-else" {
		t.Fatalf("concurrent content must be preserved: %q", got)
	}
}

// ---------------------------------------------------------------------------
// EXPERIMENTAL private-primitive behavior (NOT production-supported mutation).
// ---------------------------------------------------------------------------

func TestR1AExperimentalSupervisedEditUnverified(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aCompletedUnverified {
		t.Fatalf("want completed_unverified, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "new" {
		t.Fatalf("file not updated: %q", got)
	}
}

func TestR1AExperimentalEditVerified(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), r1aHashVerifier{}, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aCompletedVerified {
		t.Fatalf("want completed_verified, got %s", out)
	}
}

func TestR1AExperimentalPreservesFileMode(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o640)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(ws, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode not preserved: got %o", info.Mode().Perm())
	}
}

func TestR1AExperimentalUnrelatedFilesUntouched(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "keep-dirty.txt", []byte("dirty"), 0o644)
	r1aWriteFile(t, ws, "sub/untracked.txt", []byte("untracked"), 0o644)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := r1aReadDisk(t, ws, "keep-dirty.txt"); string(got) != "dirty" {
		t.Fatal("unrelated dirty file changed")
	}
	if got := r1aReadDisk(t, ws, "sub/untracked.txt"); string(got) != "untracked" {
		t.Fatal("unrelated untracked file changed")
	}
}

func TestR1AExperimentalRollbackWhenMatchesAfterHash(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil {
		t.Fatal(err)
	}
	out, err := r1aRollbackExperimental(t, txn, nil)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if out != r1aFailedRolledBack {
		t.Fatalf("want failed_rolled_back, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("file not restored: %q", got)
	}
}

func TestR1AExperimentalRollbackRefusesAfterConcurrentEdit(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil {
		t.Fatal(err)
	}
	// User edits the file after our write.
	r1aWriteFile(t, ws, "f.txt", []byte("user-edit"), 0o644)
	out, err := r1aRollbackExperimental(t, txn, nil)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if out != r1aRecoveryRequired {
		t.Fatalf("want recovery_required, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "user-edit" {
		t.Fatalf("concurrent user edit must not be overwritten: %q", got)
	}
}

type r1aFailingVerifier struct{}

func (r1aFailingVerifier) Verify(*r1aTransaction) (bool, error) { return false, nil }

func TestR1AExperimentalVerificationFailureRollsBack(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), r1aFailingVerifier{}, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aFailedRolledBack {
		t.Fatalf("want failed_rolled_back, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "old" {
		t.Fatalf("file must be restored after verify failure: %q", got)
	}
}

func TestR1AExperimentalIdempotentRetry(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out1, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err != nil {
		t.Fatalf("idempotent retry errored: %v", err)
	}
	if out1 != out2 {
		t.Fatalf("retry changed outcome: %s vs %s", out1, out2)
	}
	applied := 0
	for _, e := range txn.log.snapshotEvents() {
		if e.EventType == "write.applied" {
			applied++
		}
	}
	if applied != 1 {
		t.Fatalf("mutation applied %d times, want 1", applied)
	}
}

// TestR1AExperimentalPreCheckInterleaveApplyDetected injects an external write
// BEFORE the final re-read (not in the final-read→rename gap) and proves the
// experimental apply DETECTS that pre-check change, fails closed with conflict,
// and does not overwrite it. It demonstrates detection of a pre-check change
// only; it does NOT exercise or protect the final-read→rename syscall gap, where
// a write would still be silently overwritten (the unmet R1 guarantee, which is
// why production mutation is contained). A passing Go race detector does not
// establish a filesystem compare-and-swap. The injection is passed by value —
// there is no production-global hook.
func TestR1AExperimentalPreCheckInterleaveApplyDetected(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)

	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, func() {
		_ = os.WriteFile(filepath.Join(ws, "f.txt"), []byte("concurrent"), 0o644)
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if out != r1aConflict {
		t.Fatalf("want conflict on pre-check interleave, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "concurrent" {
		t.Fatalf("pre-check write must not be overwritten: %q", got)
	}
}

// TestR1AExperimentalPreCheckInterleaveRollbackDetected injects a write BEFORE
// the final re-read of the restore and proves rollback DETECTS that pre-check
// change, refuses, and returns recovery_required rather than overwriting the
// user's change. As above, it demonstrates pre-check detection only, not
// protection of the final-read→rename gap.
func TestR1AExperimentalPreCheckInterleaveRollbackDetected(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil {
		t.Fatal(err)
	}
	out, err := r1aRollbackExperimental(t, txn, func() {
		_ = os.WriteFile(filepath.Join(ws, "f.txt"), []byte("user-edit"), 0o644)
	})
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if out != r1aRecoveryRequired {
		t.Fatalf("want recovery_required, got %s", out)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "user-edit" {
		t.Fatalf("pre-check user edit must survive rollback refusal: %q", got)
	}
}

// ---------------------------------------------------------------------------
// Protected store, cleanup, persistence, toolchain, boundaries, truthfulness.
// ---------------------------------------------------------------------------

func TestR1AProtectedArtifactsAndPermissions(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)

	info, err := os.Stat(txn.storeDir)
	if err != nil {
		t.Fatalf("store dir missing: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("store dir not owner-only: %o", info.Mode().Perm())
	}
	pre := filepath.Join(txn.storeDir, "snapshot", "preimage")
	pinfo, err := os.Stat(pre)
	if err != nil {
		t.Fatalf("preimage missing: %v", err)
	}
	if pinfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("preimage not owner-only: %o", pinfo.Mode().Perm())
	}
	if got, _ := os.ReadFile(pre); string(got) != "old" {
		t.Fatalf("preimage content wrong: %q", got)
	}
	// Store must be OUTSIDE the workspace.
	if rel, err := filepath.Rel(ws, txn.storeDir); err == nil && !filepath.IsAbs(rel) && rel != ".." && !hasDotDotPrefix(rel) {
		t.Fatalf("store dir must not be inside the workspace: %s", txn.storeDir)
	}
}

func hasDotDotPrefix(rel string) bool {
	return len(rel) >= 2 && rel[0] == '.' && rel[1] == '.'
}

func TestR1ACleanupOnlyAfterExperimentalVerifiedSuccess(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	// Eligible cleanup requires VERIFIED completion (contract §7). Reach
	// completed_verified on the experimental path using the existing in-process
	// hash verifier. This is synthetic test state, NOT evidence that real
	// customer verification is implemented.
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), r1aHashVerifier{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != r1aCompletedVerified {
		t.Fatalf("setup must reach completed_verified, got %s", out)
	}
	store := txn.storeDir
	if err := txn.CleanupSuccess(); err != nil {
		t.Fatalf("cleanup after verified success: %v", err)
	}
	if _, err := os.Stat(store); !os.IsNotExist(err) {
		t.Fatal("verified-success artifacts should be removed after explicit finalize")
	}

	// A failed/recovery outcome must NOT be cleanable.
	txn2 := r1aPrepProposed(t, ws, "g.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn2, txn2.ApprovalBinding(true), r1aFailingVerifier{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := txn2.CleanupSuccess(); err == nil {
		t.Fatal("non-success evidence must be retained (cleanup refused)")
	}
	if _, err := os.Stat(txn2.storeDir); err != nil {
		t.Fatal("failure evidence should remain available")
	}
}

// TestR1ACleanupRefusesUnverifiedCompletion enforces contract §7/§8 through the
// public wrapper: a completed_unverified success (no verifier) is "not success"
// for cleanup purposes, so CleanupSuccess refuses and the snapshot/preimage
// evidence is retained intact.
func TestR1ACleanupRefusesUnverifiedCompletion(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != r1aCompletedUnverified {
		t.Fatalf("setup must reach completed_unverified, got %s", out)
	}
	if err := txn.CleanupSuccess(); err == nil {
		t.Fatal("completed_unverified must not be cleanable (contract §7: verified completion only)")
	}
	if _, err := os.Stat(txn.storeDir); err != nil {
		t.Fatalf("unverified-success evidence must be retained: %v", err)
	}
	if _, err := os.Stat(filepath.Join(txn.storeDir, "snapshot", "preimage")); err != nil {
		t.Fatalf("preimage evidence must be retained: %v", err)
	}
}

// TestR1ACleanupBoundaryRefusesUnverifiedDirectly guards against regressing the
// fix into the wrapper alone: the r1aCleanupSuccess deletion boundary itself
// must refuse completed_unverified (and only completed_verified is eligible),
// even when called directly.
func TestR1ACleanupBoundaryRefusesUnverifiedDirectly(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if err := r1aCleanupSuccess(txn.storeDir, r1aCompletedUnverified); err == nil {
		t.Fatal("deletion boundary must refuse completed_unverified directly")
	}
	if _, err := os.Stat(txn.storeDir); err != nil {
		t.Fatalf("store must remain after refused cleanup: %v", err)
	}
	// Only verified completion is eligible at the boundary.
	if err := r1aCleanupSuccess(txn.storeDir, r1aCompletedVerified); err != nil {
		t.Fatalf("verified completion must be cleanable at the boundary: %v", err)
	}
	if _, err := os.Stat(txn.storeDir); !os.IsNotExist(err) {
		t.Fatal("verified completion must be removable at the boundary")
	}
}

func TestR1APersistenceFailureSurfaced(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	// Destroy the protected store so manifest persistence must fail.
	if err := os.RemoveAll(txn.storeDir); err != nil {
		t.Fatal(err)
	}
	out, err := txn.Apply(txn.ApprovalBinding(true), nil)
	if out != r1aDenied {
		t.Fatalf("want denied outcome, got %s", out)
	}
	if err == nil || !strings.Contains(err.Error(), "persistence") {
		t.Fatalf("persistence failure must be surfaced, got %v", err)
	}
	if txn.persistErr == nil {
		t.Fatal("persistErr must be retained on the transaction")
	}
}

func TestR1AGoFloorEnforcedAtEntry(t *testing.T) {
	// The function r1aPrepare invokes must reject sub-floor toolchains and accept
	// the running one (proving enforcement, not merely a tested helper).
	if err := r1aEnforceGoFloor("go1.26.4"); err == nil {
		t.Fatal("go1.26.4 is below the floor and must be rejected")
	}
	if err := r1aEnforceGoFloor("go1.25.9"); err == nil {
		t.Fatal("go1.25.9 is below the floor and must be rejected")
	}
	if err := r1aEnforceGoFloor("devel go1.27-x"); err == nil {
		t.Fatal("a development toolchain must fail closed")
	}
	if err := r1aEnforceGoFloor(runtime.Version()); err != nil {
		t.Fatalf("running toolchain %q must meet the floor: %v", runtime.Version(), err)
	}
}

func TestR1ACanonicalRootEquivalentSpellings(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "f.txt", []byte("old"), 0o644)
	want := r1aWorkspaceID(ws)
	for _, spelling := range []string{ws + string(os.PathSeparator), ws + "/./", filepath.Join(ws, "nonexistent", "..")} {
		txn, err := r1aPrepare(spelling, "f.txt")
		if err != nil {
			t.Fatalf("prepare %q: %v", spelling, err)
		}
		if txn.req.WorkspaceID != want {
			t.Fatalf("spelling %q yielded workspace id %q, want %q", spelling, txn.req.WorkspaceID, want)
		}
		if txn.req.WorkspaceRoot != ws {
			t.Fatalf("spelling %q captured root %q, want canonical %q", spelling, txn.req.WorkspaceRoot, ws)
		}
	}
}

func TestR1ASymlinkedRootCanonicalized(t *testing.T) {
	r1aSetHome(t)
	real := r1aTestWorkspace(t)
	r1aWriteFile(t, real, "f.txt", []byte("old"), 0o644)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	txn, err := r1aPrepare(link, "f.txt")
	if err != nil {
		t.Fatalf("prepare via symlink: %v", err)
	}
	if txn.req.WorkspaceRoot != real {
		t.Fatalf("symlinked root not canonicalized: got %q want %q", txn.req.WorkspaceRoot, real)
	}
	if txn.req.WorkspaceID != r1aWorkspaceID(real) {
		t.Fatal("workspace id must derive from the canonical root")
	}
}

func TestR1AStoreInsideWorkspaceFailsClosed(t *testing.T) {
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "f.txt", []byte("old"), 0o644)
	// Point the protected home INSIDE the workspace so the store would land inside.
	homeInside := filepath.Join(ws, "home")
	if err := os.MkdirAll(homeInside, 0o755); err != nil {
		t.Fatal(err)
	}
	old := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return homeInside, nil }
	t.Cleanup(func() { executionModeHomeDir = old })

	if _, err := r1aPrepare(ws, "f.txt"); err == nil || !strings.Contains(err.Error(), "inside the workspace") {
		t.Fatalf("store inside the workspace must fail closed, got %v", err)
	}
}

func TestR1AApprovalRefAndOutcomesTruthful(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if _, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), r1aHashVerifier{}, nil); err != nil {
		t.Fatal(err)
	}
	var sawApproved, sawPreApproval bool
	for _, e := range txn.log.snapshotEvents() {
		switch e.EventType {
		case "prepared", "proposed":
			if e.ApprovalRef != "" {
				t.Fatalf("event %q must not carry an approval ref before approval", e.EventType)
			}
			sawPreApproval = true
		case "write.applied", "verify.passed":
			if !strings.HasPrefix(e.ApprovalRef, "approval:sha256:") {
				t.Fatalf("event %q must carry the approval ref, got %q", e.EventType, e.ApprovalRef)
			}
			sawApproved = true
		}
		// Declared-but-unimplemented outcomes must never be emitted.
		if e.Outcome == r1aCancelled || e.Outcome == r1aInterrupted {
			t.Fatalf("event %q emitted a declared-only outcome %q", e.EventType, e.Outcome)
		}
	}
	if !sawPreApproval || !sawApproved {
		t.Fatal("expected both pre-approval and post-approval events")
	}
}

func TestR1ACheckpointerImplementsSeam(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	cp := newR1ACheckpointer(txn)
	if err := cp.Checkpoint("pre-step"); err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(txn.storeDir, "snapshot", "checkpoint-pre-step"))
	if err != nil {
		t.Fatalf("checkpoint artifact missing: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("checkpoint captured wrong bytes: %q", got)
	}
}

// ---------------------------------------------------------------------------
// F-2 behavioral regression: deny-before-success/target-I/O, and persistence
// failures preserved across retry and cleanup.
// ---------------------------------------------------------------------------

// TestR1AProductionDeniesExperimentallyMutatedFixture proves a transaction
// previously mutated through the explicit experimental path cannot make the
// production entries return a cached success: production Apply denies and
// production Rollback refuses to restore.
func TestR1AProductionDeniesExperimentallyMutatedFixture(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	// Explicit experimental mutation (NOT production-supported).
	if out, err := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil); err != nil || out != r1aCompletedUnverified {
		t.Fatalf("experimental apply setup failed: out=%s err=%v", out, err)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "new" {
		t.Fatalf("setup: experimental write expected: %q", got)
	}
	// Production Apply must DENY, not return the cached experimental success.
	out, err := txn.Apply(txn.ApprovalBinding(true), nil)
	if out != r1aDenied || err == nil || !strings.Contains(err.Error(), "backend") {
		t.Fatalf("production Apply must deny on an experimentally-mutated fixture; out=%s err=%v", out, err)
	}
	// Production Rollback must refuse to restore (recovery_required), not revert.
	rout, rerr := txn.Rollback()
	if rout != r1aRecoveryRequired || rerr == nil {
		t.Fatalf("production Rollback must refuse; out=%s err=%v", rout, rerr)
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "new" {
		t.Fatalf("production Rollback must not restore: %q", got)
	}
}

// TestR1AProductionDenialPrecedesTargetInspection proves production Apply takes
// the unsupported-mutation path without inspecting or replacing the target: the
// target is removed after prepare, yet Apply denies rather than erroring on a
// read or recreating the target. (No FIFO is used — removal suffices.)
func TestR1AProductionDenialPrecedesTargetInspection(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	if err := os.Remove(filepath.Join(ws, "f.txt")); err != nil {
		t.Fatal(err)
	}
	out, err := txn.Apply(txn.ApprovalBinding(true), nil)
	if out != r1aDenied || err == nil || !strings.Contains(err.Error(), "backend") {
		t.Fatalf("denial must precede target inspection; out=%s err=%v", out, err)
	}
	if _, serr := os.Stat(filepath.Join(ws, "f.txt")); !os.IsNotExist(serr) {
		t.Fatal("a denied Apply must not recreate the target")
	}
	if tmps := r1aTargetTempFiles(t, ws, "f.txt"); len(tmps) != 0 {
		t.Fatalf("a denied Apply created temp files: %v", tmps)
	}
}

// TestR1APersistenceFailurePreservedAcrossRetryAndCleanup proves an evidence
// persistence failure is surfaced on the first experimental apply and again on
// retry (never masked with nil), that retry does not repeat the mutation, and
// that CleanupSuccess refuses while leaving the recovery evidence intact.
func TestR1APersistenceFailurePreservedAcrossRetryAndCleanup(t *testing.T) {
	r1aSetHome(t)
	ws := r1aTestWorkspace(t)
	txn := r1aPrepProposed(t, ws, "f.txt", []byte("old"), []byte("new"), 0o644)
	// Make the protected store read-only so manifest persistence fails, while the
	// snapshot evidence written during prepare remains intact. Restore perms on
	// cleanup so the temp-dir teardown can remove it.
	if err := os.Chmod(txn.storeDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(txn.storeDir, 0o700) })

	out1, err1 := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if err1 == nil || !strings.Contains(err1.Error(), "persistence") {
		t.Fatalf("initial apply must surface the persistence failure; out=%s err=%v", out1, err1)
	}
	if txn.persistErr == nil {
		t.Fatal("persistErr must be retained on the transaction")
	}
	if got := r1aReadDisk(t, ws, "f.txt"); string(got) != "new" {
		t.Fatalf("experimental write expected: %q", got)
	}
	// Retry must still surface the failure (never nil) and not repeat the write.
	out2, err2 := r1aApplyExperimental(t, txn, txn.ApprovalBinding(true), nil, nil)
	if out2 != out1 {
		t.Fatalf("retry changed outcome: %s vs %s", out2, out1)
	}
	if err2 == nil || !strings.Contains(err2.Error(), "persistence") {
		t.Fatalf("retry must still surface the persistence failure, got %v", err2)
	}
	writes := 0
	for _, e := range txn.log.snapshotEvents() {
		if e.EventType == "write.applied" {
			writes++
		}
	}
	if writes != 1 {
		t.Fatalf("retry repeated the experimental mutation: %d writes", writes)
	}
	// Cleanup must refuse while evidence is not durably persisted.
	if err := txn.CleanupSuccess(); err == nil {
		t.Fatal("cleanup must refuse after a persistence failure")
	}
	// Recovery evidence (the pre-image) must remain intact.
	if _, err := os.Stat(filepath.Join(txn.storeDir, "snapshot", "preimage")); err != nil {
		t.Fatalf("recovery evidence must remain available: %v", err)
	}
}
