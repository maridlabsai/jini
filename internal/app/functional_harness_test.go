package app

// Functional harness test — drives every self-check scenario through the real
// CLI entrypoint in-process, hermetically (offline via local-preview; TestMain
// pins state/home and missing hand-off CLIs). Shares selfCheckScenarios() with
// the `jini check functional` command (selfcheck.go), so the always-on health
// suite and the runtime self-check never drift. Add a scenario in selfcheck.go.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFunctionalHarness(t *testing.T) {
	for _, sc := range selfCheckScenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			withExecutionModeHome(t) // isolate mode/home per scenario
			t.Setenv("JINI_PROVIDER", "local-preview")
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			dir := t.TempDir()
			t.Chdir(dir)
			for name, content := range sc.SetupFiles {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			var stdout, stderr bytes.Buffer
			exit := RunInteractive(sc.Args, strings.NewReader(""), &stdout, &stderr)
			if ok, detail := checkSelfCheckResult(sc, exit, stdout.String()+stderr.String(), dir); !ok {
				t.Fatalf("scenario %s failed: %s\noutput=%q", sc.Name, detail, stdout.String()+stderr.String())
			}
		})
	}
}
