package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// JINI-R0: the scoped claim-language gate. These tests prove its real behavior —
// the manifest fixtures, the actual public-document discovery path with its
// exclusions, the live runtime-output scan, fail-closed on runtime failure, and
// that exported override variables cannot neuter the production invocation.

func gatePath(root string) string {
	return filepath.Join(root, "tools", "claim_language_gate.sh")
}

func runGate(t *testing.T, root string, env ...string) (int, string) {
	t.Helper()
	cmd := exec.Command("bash", gatePath(root))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), string(out)
	}
	if err != nil {
		t.Fatalf("invoking claim gate: %v\n%s", err, out)
	}
	return 0, string(out)
}

func writeClaimDoc(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// F (positive / negative / qualification): the manifest fixture path.
func TestClaimGateManifestFixtures(t *testing.T) {
	root := repoRootForMigrationTest(t)
	dir := t.TempDir()

	cases := []struct {
		name string
		body string
		want int
	}{
		{"bad", "Jini is un-metered and it saved you ≈ $5.\n", 1},
		{"saved-order", "Total saved ≈ $12 this week.\n", 1},
		{"free-unqualified", "Jini is free to run today.\n", 1},
		// Regression: factual "Saved ≈ $X" is banned even WITH an "imputed" label.
		{"saved-imputed-not-rescued", "Saved ≈ $5 (imputed) this task.\n", 1},
		{"qualified", "Savings are estimated/imputed; free to run is a future target gated on v1.5.\n", 0},
		// The acceptable honest form must pass.
		{"avoided-imputed-ok", "Estimated avoided API spend ≈ $5 (imputed) this task.\n", 0},
		// Ordinary non-financial "saved" must not be flagged.
		{"saved-a-file-ok", "Jini saved a file and saved your draft to disk.\n", 0},
		{"clean", "Jini routes your coding work across models you already have.\n", 0},
	}
	for _, c := range cases {
		p := filepath.Join(dir, c.name+".md")
		writeClaimDoc(t, p, c.body)
		code, out := runGate(t, root,
			"CLAIM_GATE_TESTMODE=1", "CLAIM_GATE_MANIFEST="+p, "CLAIM_GATE_SKIP_RUNTIME=1")
		if code != c.want {
			t.Fatalf("%s: got exit %d want %d\n%s", c.name, code, c.want, out)
		}
	}
}

// F.8: the real public-document discovery path is exercised (recursively), and
// F.9: the recovery-plan / historical / archive exclusions are exercised through
// actual discovery, not by omitting a fixture from a manifest.
func TestClaimGateRealDiscoveryAndExclusions(t *testing.T) {
	root := repoRootForMigrationTest(t)

	// (a) A violation in a NESTED normal doc must be caught by real discovery.
	caught := t.TempDir()
	writeClaimDoc(t, filepath.Join(caught, "index.md"), "All good here.\n")
	writeClaimDoc(t, filepath.Join(caught, "guides", "deep.md"), "Jini is un-metered.\n")
	if code, out := runGate(t, root,
		"CLAIM_GATE_TESTMODE=1", "CLAIM_GATE_DOCROOT="+caught, "CLAIM_GATE_SKIP_RUNTIME=1"); code != 1 {
		t.Fatalf("real discovery must catch a nested-doc violation; got %d\n%s", code, out)
	}

	// (b) Violations present ONLY in excluded files must NOT fail the gate.
	excluded := t.TempDir()
	writeClaimDoc(t, filepath.Join(excluded, "clean.md"), "Honest, qualified copy.\n")
	writeClaimDoc(t, filepath.Join(excluded, "Jini_Trust_First_Recovery_Plan.md"), "un-metered walls avoided $9 saved\n")
	writeClaimDoc(t, filepath.Join(excluded, "old-historical.md"), "un-metered walls avoided\n")
	writeClaimDoc(t, filepath.Join(excluded, "archive", "old.md"), "un-metered walls avoided\n")
	writeClaimDoc(t, filepath.Join(excluded, "bannered.md"), "# Title\n\nHISTORICAL — unverified snapshot\n\nun-metered walls avoided\n")
	if code, out := runGate(t, root,
		"CLAIM_GATE_TESTMODE=1", "CLAIM_GATE_DOCROOT="+excluded, "CLAIM_GATE_SKIP_RUNTIME=1"); code != 0 {
		t.Fatalf("excluded files must be skipped by real discovery; got %d\n%s", code, out)
	}
}

// F.10: the real runtime-output scan runs (no CLAIM_GATE_SKIP_RUNTIME) against the
// actual repo and passes clean through the production path.
func TestClaimGateProductionRunsRuntimeScanClean(t *testing.T) {
	root := repoRootForMigrationTest(t)
	code, out := runGate(t, root) // no test overrides at all
	if code != 0 {
		t.Fatalf("production gate must be clean; got %d\n%s", code, out)
	}
	if !strings.Contains(out, "claim-language gate: clean") || strings.Contains(out, "test manifest") {
		t.Fatalf("expected the production (non-manifest) clean path; got:\n%s", out)
	}
}

// F.11: a runtime scan failure (misbehaving binary) makes the gate fail closed.
func TestClaimGateRuntimeFailureFailsClosed(t *testing.T) {
	root := repoRootForMigrationTest(t)
	dir := t.TempDir()
	fake := filepath.Join(dir, "jini")
	// share exits 0 (should be 1); mcp leaks [ready]; savings presents dollars as saved.
	writeClaimDoc(t, fake, "#!/bin/sh\ncase \"$1\" in\n  share) exit 0 ;;\n  mcp) echo 'x [ready]'; exit 0 ;;\n  savings) echo 'Total saved $9'; exit 0 ;;\n  *) exit 0 ;;\nesac\n")
	if err := os.Chmod(fake, 0o755); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	emptyDocs := t.TempDir() // isolate from real docs to keep the test fast/deterministic
	code, out := runGate(t, root,
		"CLAIM_GATE_TESTMODE=1", "CLAIM_GATE_DOCROOT="+emptyDocs, "CLAIM_GATE_BIN="+fake)
	if code != 1 {
		t.Fatalf("misbehaving runtime binary must fail the gate; got %d\n%s", code, out)
	}
}

// F.12: exported override variables cannot neuter the gate on the production path.
func TestClaimGateProductionIgnoresExportedOverrides(t *testing.T) {
	root := repoRootForMigrationTest(t)
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.md")
	writeClaimDoc(t, bad, "Jini is un-metered.\n")

	// Under TESTMODE the bad manifest is a genuine violation.
	if code, _ := runGate(t, root,
		"CLAIM_GATE_TESTMODE=1", "CLAIM_GATE_MANIFEST="+bad, "CLAIM_GATE_SKIP_RUNTIME=1"); code != 1 {
		t.Fatalf("control: bad manifest under TESTMODE must fail")
	}
	// WITHOUT TESTMODE (production), the same exported overrides are ignored: the
	// gate scans the real (clean) surfaces and never short-circuits on the manifest.
	code, out := runGate(t, root,
		"CLAIM_GATE_MANIFEST="+bad, "CLAIM_GATE_SKIP_RUNTIME=1")
	if code != 0 {
		t.Fatalf("production must ignore exported overrides and scan real surfaces; got %d\n%s", code, out)
	}
	if strings.Contains(out, "test manifest") {
		t.Fatalf("production path must not honor CLAIM_GATE_MANIFEST; got:\n%s", out)
	}
}

// Wiring: the required-gate runner sanitizes every override and the matrix lists it.
func TestClaimGateWiredAndSanitized(t *testing.T) {
	root := repoRootForMigrationTest(t)
	runner := readRepoFile(t, root, "tools/run_required_gates.sh")
	if !strings.Contains(runner, "claim_language_gate.sh") {
		t.Fatal("tools/run_required_gates.sh must invoke claim_language_gate.sh")
	}
	for _, v := range []string{"CLAIM_GATE_TESTMODE", "CLAIM_GATE_MANIFEST", "CLAIM_GATE_SKIP_RUNTIME", "CLAIM_GATE_DOCROOT", "CLAIM_GATE_BIN"} {
		if !strings.Contains(runner, "-u "+v) {
			t.Fatalf("runner must sanitize override %q before invoking the gate", v)
		}
	}
	matrix := readRepoFile(t, root, "specs/engineering-gate-matrix.md")
	if !strings.Contains(matrix, "claim_language_gate.sh") {
		t.Fatal("specs/engineering-gate-matrix.md must register claim_language_gate.sh")
	}
}
