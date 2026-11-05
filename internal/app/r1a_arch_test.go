package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// r1aProductionSources returns the non-test r1a_*.go source files.
func r1aProductionSources(t *testing.T) map[string]string {
	t.Helper()
	matches, err := filepath.Glob("r1a_*.go")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		data, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		out[m] = string(data)
	}
	if len(out) == 0 {
		t.Fatal("no r1a production sources found")
	}
	return out
}

// TestR1ACoreHasOneMutationPrimitiveAndNoUnsafeFallback proves the new core
// mutates the workspace only via the confined os.Root rename path and never
// falls back to toolEditFile's direct os.WriteFile. Integration is deferred to
// Commit 2; this guards against accidental coupling in the meantime.
func TestR1ACoreHasOneMutationPrimitiveAndNoUnsafeFallback(t *testing.T) {
	srcs := r1aProductionSources(t)

	for name, body := range srcs {
		if strings.Contains(body, "os.WriteFile(") {
			t.Fatalf("%s uses os.WriteFile; the R1A core must write via protected temp+rename / os.Root only", name)
		}
		// A prose comment may name toolEditFile (the step requires documenting
		// that it stays unsafe/unwired); what must not appear is a CALL to it —
		// i.e. coupling or a fallback into the unmediated writer.
		if strings.Contains(body, "toolEditFile(") || strings.Contains(body, "allAgentTools(") {
			t.Fatalf("%s calls into the unmediated edit_file path; the core must not couple to it in Commit 1", name)
		}
		if strings.Contains(body, "Jini_Trust_First_Recovery_Plan") {
			t.Fatalf("%s references the recovery report; it must never be read or touched", name)
		}
		if strings.Contains(body, "exec.Command") || strings.Contains(body, "sh -c") {
			t.Fatalf("%s introduces a subprocess; none is permitted in Commit 1", name)
		}
	}

	txn, ok := srcs["r1a_txn.go"]
	if !ok {
		t.Fatal("r1a_txn.go not found")
	}
	if strings.Count(txn, "root.Rename(") == 0 {
		t.Fatal("the workspace mutation must go through the confined os.Root.Rename")
	}

	// The unmediated path still exists and is intentionally left unchanged and
	// unwired until Commit 2 replaces it.
	edit, err := os.ReadFile("agent_tools.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(edit), "os.WriteFile(path, []byte(body), 0o644)") {
		t.Skip("toolEditFile direct-write signature changed; re-verify Commit-1 boundary assumptions")
	}
}

// TestR1AContainmentIsStructural proves — from source, not behavior — that the
// production mutation path is fail-closed and cannot be re-enabled by any
// runtime input, and that the toolchain floor is wired at the entry point.
func TestR1AContainmentIsStructural(t *testing.T) {
	srcs := r1aProductionSources(t)

	txn, ok := srcs["r1a_txn.go"]
	if !ok {
		t.Fatal("r1a_txn.go not found")
	}
	// The production switch is a compile-time false and is never set true.
	if !strings.Contains(txn, "const r1aProductionMutationEnabled = false") {
		t.Fatal("production mutation switch must be a compile-time false constant")
	}
	if strings.Contains(txn, "r1aProductionMutationEnabled = true") {
		t.Fatal("the production mutation switch must never be set true")
	}
	// Apply and Rollback must route through that constant (no bypass entry).
	if !strings.Contains(txn, "applyBounded(binding, verifier, r1aProductionMutationEnabled") {
		t.Fatal("Apply must pass the production (disabled) mutation switch")
	}
	if !strings.Contains(txn, "rollbackBounded(r1aProductionMutationEnabled") {
		t.Fatal("Rollback must pass the production (disabled) mutation switch")
	}
	// The toolchain floor is enforced at the protected entry point.
	if !strings.Contains(txn, "r1aEnforceGoFloor(") {
		t.Fatal("the Go toolchain floor must be enforced in the preparation entry point")
	}

	// No production R1A source may consult the environment: a flag/env toggle
	// would be a runtime override of the containment guard.
	for name, body := range srcs {
		if strings.Contains(body, "os.Getenv(") || strings.Contains(body, "os.LookupEnv(") {
			t.Fatalf("%s reads the environment; containment must not be toggleable at runtime", name)
		}
		// The experimental write path is enabled only by passing a literal true,
		// which must appear solely in test files (excluded from these sources).
		if strings.Contains(body, "applyBounded(") && strings.Contains(body, ", true,") {
			t.Fatalf("%s appears to enable the experimental mutation path in production", name)
		}
		if strings.Contains(body, "rollbackBounded(true") {
			t.Fatalf("%s appears to enable the experimental rollback path in production", name)
		}
	}
}
