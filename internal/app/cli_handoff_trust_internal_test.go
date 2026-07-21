package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDarwinCLIHandoffTrustIssueTrustsNonQuarantinedExecutable(t *testing.T) {
	issue := darwinCLIHandoffTrustIssue(
		"/Users/example/.local/bin/locally-built-cli",
		func(string) bool { return false },
		func(string) bool {
			t.Fatalf("signature check must not run for a non-quarantined executable")
			return false
		},
	)
	if issue != "" {
		t.Fatalf("expected non-quarantined executable to be trusted, got %q", issue)
	}
}

func TestDarwinCLIHandoffTrustIssueTrustsQuarantinedIdentitySignedExecutable(t *testing.T) {
	issue := darwinCLIHandoffTrustIssue(
		"/Users/example/Downloads/claude",
		func(string) bool { return true },
		func(string) bool { return true },
	)
	if issue != "" {
		t.Fatalf("expected quarantined identity-signed executable to be trusted, got %q", issue)
	}
}

func TestDarwinCLIHandoffTrustIssueRejectsQuarantinedUnidentifiedExecutable(t *testing.T) {
	// Covers unsigned, broken, and merely ad-hoc signed binaries alike: the
	// identity check is what separates them from Developer ID signatures.
	issue := darwinCLIHandoffTrustIssue(
		"/Users/example/Downloads/claude",
		func(string) bool { return true },
		func(string) bool { return false },
	)
	for _, want := range []string{
		"macOS Gatekeeper rejected CLI executable: /Users/example/Downloads/claude",
		"quarantined download without an identified developer signature",
		"xattr -d com.apple.quarantine /Users/example/Downloads/claude",
	} {
		if !strings.Contains(issue, want) {
			t.Fatalf("expected quarantined unidentified executable issue to contain %q, got %q", want, issue)
		}
	}
	if shipCLIHandoffSetupCategory([]string{issue}) != "macOS Gatekeeper" {
		t.Fatalf("expected quarantine issue to map to the macOS Gatekeeper ship category, got %q", shipCLIHandoffSetupCategory([]string{issue}))
	}
}

func TestDefaultCLIHandoffTrustIssueUsesRealCodesignAndQuarantineChecks(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only integration check")
	}
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "")

	dir := t.TempDir()
	// t.TempDir may itself sit behind a symlink (/var -> /private/var); the
	// trust check reports the fully resolved target path.
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	script := filepath.Join(dir, "local-cli")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write local script: %v", err)
	}

	// Unsigned local script without quarantine metadata is trusted.
	if issue := defaultCLIHandoffTrustIssue(script); issue != "" {
		t.Fatalf("expected unsigned non-quarantined script to be trusted, got %q", issue)
	}

	// An AD-HOC signed Mach-O carrying quarantine metadata must fail closed.
	// `codesign --verify --strict` alone accepts ad-hoc signatures, and the
	// arm64 linker ad-hoc signs every native binary, so this is the case that
	// would silently trust any downloaded arm64 binary if the identity check
	// were dropped.
	adhoc := filepath.Join(dir, "adhoc-cli")
	if err := copyFileForTrustTest("/bin/ls", adhoc); err != nil {
		t.Fatalf("copy binary for ad-hoc signing: %v", err)
	}
	if out, err := exec.Command("codesign", "--force", "-s", "-", adhoc).CombinedOutput(); err != nil {
		t.Skipf("could not ad-hoc sign test binary: %v (%s)", err, out)
	}
	if err := exec.Command("codesign", "--verify", "--strict", adhoc).Run(); err != nil {
		t.Skipf("ad-hoc signature did not verify, cannot exercise the regression: %v", err)
	}
	if out, err := exec.Command("xattr", "-w", "com.apple.quarantine", "0081;00000000;jini-test;", adhoc).CombinedOutput(); err != nil {
		t.Skipf("could not set quarantine xattr: %v (%s)", err, out)
	}
	if issue := defaultCLIHandoffTrustIssue(adhoc); !strings.Contains(issue, "identified developer signature") {
		t.Fatalf("expected quarantined ad-hoc signed binary to fail closed, got %q", issue)
	}

	// The same script carrying quarantine metadata fails closed, and a symlink
	// to it is assessed against the resolved target.
	if out, err := exec.Command("xattr", "-w", "com.apple.quarantine", "0081;00000000;jini-test;", script).CombinedOutput(); err != nil {
		t.Skipf("could not set quarantine xattr: %v (%s)", err, out)
	}
	issue := defaultCLIHandoffTrustIssue(script)
	if !strings.Contains(issue, "macOS Gatekeeper rejected CLI executable: "+script) {
		t.Fatalf("expected quarantined unsigned script to fail closed, got %q", issue)
	}
	link := filepath.Join(dir, "link-cli")
	if err := os.Symlink(script, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	linkIssue := defaultCLIHandoffTrustIssue(link)
	if !strings.Contains(linkIssue, script) {
		t.Fatalf("expected symlink trust check to name resolved target %q, got %q", script, linkIssue)
	}

	// A validly signed system binary is trusted even though spctl's execute
	// assessment rejects standalone CLI binaries.
	if issue := defaultCLIHandoffTrustIssue("/bin/ls"); issue != "" {
		t.Fatalf("expected signed system binary to be trusted, got %q", issue)
	}
}

func copyFileForTrustTest(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o755)
}

func TestDefaultCLIHandoffTrustIssueHonorsSkipEnv(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only check")
	}
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	if issue := defaultCLIHandoffTrustIssue(filepath.Join(t.TempDir(), "does-not-exist")); issue != "" {
		t.Fatalf("expected skip env to bypass trust check, got %q", issue)
	}
}
