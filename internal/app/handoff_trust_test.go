package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestTrustStoreRoundTripWithLevels(t *testing.T) {
	withExecutionModeHome(t)
	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelAutonomous, trustAcknowledgement(trustLevelAutonomous)); err != nil {
		t.Fatal(err)
	}
	level, ok := trustedLevelForDir(dir)
	if !ok || level != trustLevelAutonomous {
		t.Fatalf("expected autonomous, got %q ok=%v", level, ok)
	}
	// acknowledged recorded for audit.
	store := loadTrustStore()
	if store == nil || store.Dirs[resolveTrustDir(dir)].Acknowledged != "edits+commands" {
		t.Fatalf("acknowledged not recorded: %+v", store)
	}
	removed, err := removeTrustGrant(dir)
	if err != nil || !removed {
		t.Fatalf("remove failed: removed=%v err=%v", removed, err)
	}
	if _, ok := trustedLevelForDir(dir); ok {
		t.Fatal("still trusted after remove")
	}
}

func TestTrustStoreCorruptReadsUntrusted(t *testing.T) {
	withExecutionModeHome(t)
	path, _ := trustStorePath()
	if err := os.MkdirAll(strings.TrimSuffix(path, "/trusted-dirs.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := trustedLevelForDir(t.TempDir()); ok {
		t.Fatal("corrupt store must read as untrusted")
	}
}

func TestRunTrustConsentGrantsOnYes(t *testing.T) {
	withExecutionModeHome(t)
	withPromptIO(t, "y\n") // sets TTY true + input
	var out bytes.Buffer
	if code := runTrust(nil, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := out.String()
	if !strings.Contains(got, "apply file edits without asking first") || !strings.Contains(got, "won't run commands") {
		t.Fatalf("semi repercussions not shown: %q", got)
	}
	if !strings.Contains(got, "Trusted ") {
		t.Fatalf("grant not confirmed: %q", got)
	}
	cwd, _ := os.Getwd()
	if level, ok := trustedLevelForDir(cwd); !ok || level != trustLevelSemi {
		t.Fatalf("cwd not trusted at semi: %q ok=%v", level, ok)
	}
}

func TestRunTrustBareEnterDeclines(t *testing.T) {
	withExecutionModeHome(t)
	withPromptIO(t, "\n")
	var out bytes.Buffer
	runTrust(nil, &out, io.Discard)
	if !strings.Contains(out.String(), "Nothing changed.") {
		t.Fatalf("bare enter must decline: %q", out.String())
	}
	cwd, _ := os.Getwd()
	if _, ok := trustedLevelForDir(cwd); ok {
		t.Fatal("declined consent must not record a grant")
	}
}

func TestRunTrustNoTTYRecordsNothing(t *testing.T) {
	withExecutionModeHome(t)
	withPromptIO(t, "y\n")
	throttlePromptIsTTY = func() bool { return false }
	var out, errOut bytes.Buffer
	if code := runTrust([]string{"--autonomous"}, &out, &errOut); code != 1 {
		t.Fatalf("no-TTY must exit 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "interactive terminal") {
		t.Fatalf("no-TTY message missing: %q", errOut.String())
	}
	cwd, _ := os.Getwd()
	if _, ok := trustedLevelForDir(cwd); ok {
		t.Fatal("no-TTY must not record a grant")
	}
}

func TestRunTrustYesGrantsNonInteractively(t *testing.T) {
	withExecutionModeHome(t)
	// No TTY at all — --yes must still grant, having shown the disclosure.
	throttlePromptIsTTY = func() bool { return false }
	var out, errOut bytes.Buffer
	if code := runTrust([]string{"--autonomous", "--yes"}, &out, &errOut); code != 0 {
		t.Fatalf("--yes must grant without a TTY, got code=%d err=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "apply file edits and run commands") {
		t.Fatalf("disclosure must still be shown with --yes: %q", out.String())
	}
	if !strings.Contains(out.String(), "confirmed with --yes") {
		t.Fatalf("expected confirmation line, got %q", out.String())
	}
	cwd, _ := os.Getwd()
	if level, ok := trustedLevelForDir(cwd); !ok || level != trustLevelAutonomous {
		t.Fatalf("--yes must record an autonomous grant, got level=%q ok=%v", level, ok)
	}
	t.Cleanup(func() { removeTrustGrant(cwd) })
}

func TestRunTrustAutonomousShowsCommandsCopy(t *testing.T) {
	withExecutionModeHome(t)
	withPromptIO(t, "n\n")
	var out bytes.Buffer
	runTrust([]string{"--autonomous"}, &out, io.Discard)
	if !strings.Contains(out.String(), "apply file edits and run commands") {
		t.Fatalf("autonomous copy must state commands: %q", out.String())
	}
}

func TestTrustCopyIsNeutralNoFearFraming(t *testing.T) {
	for _, level := range []string{trustLevelSemi, trustLevelAutonomous} {
		copy := trustRepercussions(level)
		// No ALL-CAPS capability words (fear framing).
		for _, banned := range []string{"WARNING", "DANGER", "DELETE", "RUN COMMANDS", "IRREVERSIBLE", "CAUTION"} {
			if strings.Contains(copy, banned) {
				t.Fatalf("%s copy contains fear framing %q: %q", level, banned, copy)
			}
		}
	}
}

func TestTrustIsARoutedTopLevelCommand(t *testing.T) {
	if canonicalTopLevelCommand("trust") != "trust" {
		t.Fatal("trust not canonical top-level command")
	}
	if err := validateNativeArgs([]string{"trust", "--autonomous"}); err != nil {
		t.Fatalf("trust --autonomous must validate: %v", err)
	}
}
