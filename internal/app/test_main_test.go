package app

import (
	"os"
	"path/filepath"
	"testing"
)

// cliHandoffExecutableEnvVars lists the CLI handoff executable overrides that
// TestMain pins to a missing path so the suite stays hermetic: without this,
// a developer machine with a real signed `claude` (or other) CLI installed
// would let auto-routing hand test prompts to the real, billed CLI. Tests that
// exercise handoffs point these at fake executables (the writeFakeExecutable /
// writeProviderFakeExecutable helpers do this automatically).
var cliHandoffExecutableEnvVars = []string{
	"JINI_CODEX_CLI",
	"JINI_CLAUDE_CODE_CLI",
	"JINI_GEMINI_CLI",
	"JINI_AIDER_CLI",
	"JINI_OPENCODE_CLI",
}

func TestMain(m *testing.M) {
	stateDir, err := os.MkdirTemp("", "jini-app-test-state-")
	if err != nil {
		panic(err)
	}
	previous, hadPrevious := os.LookupEnv("JINI_STATE_DIR")
	if err := os.Setenv("JINI_STATE_DIR", stateDir); err != nil {
		panic(err)
	}
	// Pin the global ~/.jini home to the temp dir so side-effecting writes —
	// the savings ledger written by any successful work task, and mode.json —
	// never touch the developer's real home. Tests that need specific home
	// behavior still override executionModeHomeDir per-test.
	executionModeHomeDir = func() (string, error) { return stateDir, nil }
	// Pinned unconditionally: honoring a pre-existing export would leave the
	// hole this closes, since a developer shell that exports JINI_CLAUDE_CODE_CLI
	// would let auto-routing hand test prompts to the real, billed CLI.
	// Individual tests still override per-test via t.Setenv.
	missingCLI := filepath.Join(stateDir, "missing-cli-handoff-executable")
	for _, env := range cliHandoffExecutableEnvVars {
		if err := os.Setenv(env, missingCLI); err != nil {
			panic(err)
		}
	}
	code := m.Run()
	if hadPrevious {
		_ = os.Setenv("JINI_STATE_DIR", previous)
	} else {
		_ = os.Unsetenv("JINI_STATE_DIR")
	}
	_ = os.RemoveAll(stateDir)
	os.Exit(code)
}
