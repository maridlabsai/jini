package app_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maridlabsai/jini/internal/app"
)

func TestDirectPromptWithConfiguredCLIRoutePrintsAnswerPlainly(t *testing.T) {
	stateDir := t.TempDir()
	fakeBin := t.TempDir()
	argsPath := filepath.Join(t.TempDir(), "args.txt")
	t.Setenv("JINI_TEST_CLI_ARGS_PATH", argsPath)
	writeFakeExecutable(t, fakeBin, "claude", "printf '%s\\n' \"$@\" > \"$JINI_TEST_CLI_ARGS_PATH\"\nprintf 'jini dogfood ok\\n'\n")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"reply", "with", "exactly:", "jini", "dogfood", "ok"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "jini dogfood ok") {
		t.Fatalf("expected routed CLI answer to be printed plainly, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Working on:",
		"Saved.",
		"Next:",
		"Working Draft",
		"Task Snapshot",
		"jini open",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct routed prompt to avoid draft ceremony %q, got:\n%s", unwanted, out)
		}
	}
	args := strings.Split(strings.TrimSpace(mustReadFile(t, argsPath)), "\n")
	wantArgs := []string{"--print", "reply with exactly: jini dogfood ok"}
	if strings.Join(args, "\n") != strings.Join(wantArgs, "\n") {
		t.Fatalf("expected CLI handoff args %#v, got %#v", wantArgs, args)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected direct routed prompt not to create current work, stat error: %v", err)
	}
}

func TestDirectPromptWithBrokenCLIRouteFailsLoudly(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(t.TempDir(), "missing-claude"))
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := app.Run([]string{"summarize", "the", "release", "notes"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected broken configured CLI route to fail loudly, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "CLI handoff needs setup") {
		t.Fatalf("expected setup guidance on stderr, got:\n%s", stderr.String())
	}
	if strings.Contains(stdout.String(), "Working on:") {
		t.Fatalf("expected no draft ceremony for broken route, got:\n%s", stdout.String())
	}
}
