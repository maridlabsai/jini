package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeGoModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return dir
}

func stubVerifyRunner(t *testing.T, fn func(args []string) (bool, string)) {
	t.Helper()
	old := verifyCommandRunner
	t.Cleanup(func() { verifyCommandRunner = old })
	verifyCommandRunner = func(_ context.Context, _ string, args []string) (bool, string) {
		return fn(args)
	}
}

func TestVerifyWorkspaceGoPlanBuildAndVet(t *testing.T) {
	dir := writeGoModule(t)
	stubVerifyRunner(t, func(args []string) (bool, string) { return true, "" })

	r := verifyWorkspace(context.Background(), dir, false)
	if r.Ran != 2 {
		t.Fatalf("Go workspace without --tests must run build+vet (2), got %d", r.Ran)
	}
	if !r.Verified {
		t.Fatalf("all-pass workspace must be verified")
	}
}

func TestVerifyWorkspaceFailsWhenACheckFails(t *testing.T) {
	dir := writeGoModule(t)
	// build passes, vet fails.
	stubVerifyRunner(t, func(args []string) (bool, string) {
		if len(args) >= 2 && args[1] == "vet" {
			return false, "vet: suspicious construct"
		}
		return true, ""
	})

	r := verifyWorkspace(context.Background(), dir, false)
	if r.Verified {
		t.Fatalf("a failing check must make the workspace NOT verified")
	}
	if r.Failed != 1 {
		t.Fatalf("expected exactly 1 failed check, got %d", r.Failed)
	}
}

func TestVerifyWorkspaceIncludeTestsAddsTestCheck(t *testing.T) {
	dir := writeGoModule(t)
	stubVerifyRunner(t, func(args []string) (bool, string) { return true, "" })

	r := verifyWorkspace(context.Background(), dir, true)
	if r.Ran != 3 {
		t.Fatalf("Go workspace with --tests must run build+vet+test (3), got %d", r.Ran)
	}
}

func TestVerifyWorkspaceUnverifiableWithoutBuildSystem(t *testing.T) {
	dir := t.TempDir() // empty: no go.mod / package.json / etc.
	r := verifyWorkspace(context.Background(), dir, false)
	if r.Ran != 0 {
		t.Fatalf("empty workspace must run 0 checks, got %d", r.Ran)
	}
	if r.Verified {
		t.Fatalf("an unverifiable workspace must never report verified")
	}
}

func TestRunVerifyExitCodes(t *testing.T) {
	goDir := writeGoModule(t)
	emptyDir := t.TempDir()

	// verified → exit 0
	stubVerifyRunner(t, func(_ []string) (bool, string) { return true, "" })
	var out bytes.Buffer
	if code := runVerify([]string{goDir}, &out, &out); code != 0 {
		t.Fatalf("verified workspace must exit 0, got %d\n%s", code, out.String())
	}

	// a check fails → exit 1
	stubVerifyRunner(t, func(_ []string) (bool, string) { return false, "boom" })
	out.Reset()
	if code := runVerify([]string{goDir}, &out, &out); code != 1 {
		t.Fatalf("failed check must exit 1, got %d", code)
	}

	// nothing runnable → exit 2
	out.Reset()
	if code := runVerify([]string{emptyDir}, &out, &out); code != 2 {
		t.Fatalf("unverifiable workspace must exit 2, got %d", code)
	}
}

func TestRunVerifyJSONOutput(t *testing.T) {
	dir := writeGoModule(t)
	stubVerifyRunner(t, func(_ []string) (bool, string) { return true, "" })
	var out bytes.Buffer
	runVerify([]string{dir, "--json"}, &out, &out)
	if !bytes.Contains(out.Bytes(), []byte(`"verified": true`)) {
		t.Fatalf("--json must emit a machine-readable verdict, got:\n%s", out.String())
	}
}
