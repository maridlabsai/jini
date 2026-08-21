package app

// Functional self-check — the harness as a product capability. `jini check
// functional` drives Jini's own core scenarios end-to-end, offline
// (local-preview) and hermetically (each scenario in a throwaway cwd/home), and
// reports which passed. It lets a user — or Jini itself during a dogfood loop —
// confirm "am I functional across all scenarios?" at any time, with no network
// and no side effects on the real workspace. The same scenario table backs the
// TestFunctionalHarness suite, so the command and the test never drift.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type selfCheckScenario struct {
	Name         string
	Args         []string
	WantExit     int
	Contains     []string          // every substring must appear (stdout+stderr)
	NotContains  []string          // none may appear
	SetupFiles   map[string]string // files written into the scenario cwd first
	FileContains map[string]string // filename → substring that must be present after
}

// selfCheckScenarios is the single source of truth for "is Jini functional".
func selfCheckScenarios() []selfCheckScenario {
	return []selfCheckScenario{
		{Name: "arithmetic-symbol", Args: []string{"what is 17 * 23?"}, Contains: []string{"391."}},
		{Name: "arithmetic-word", Args: []string{"what is 2 plus 2"}, Contains: []string{"4."}},
		{Name: "capital", Args: []string{"what is the capital of France?"}, Contains: []string{"Paris."}},
		{Name: "ambiguous-entity", Args: []string{"React"}, Contains: []string{"Specify what to do with React."}},
		{Name: "standalone-question-offline", Args: []string{"who is the CEO of Apple?"}, Contains: []string{"No configured route"}},
		{Name: "work-task", Args: []string{"refactor the database connection pool for reuse"}, Contains: []string{"Working on:", "Saved."}},
		{Name: "file-edit",
			Args:         []string{`add a line saying "jini was here" in the notes.txt file in this folder`},
			Contains:     []string{"Updated notes.txt"},
			SetupFiles:   map[string]string{"notes.txt": "first\n"},
			FileContains: map[string]string{"notes.txt": "jini was here"}},
		{Name: "attachment-present", Args: []string{"summarize @report.md"},
			Contains:   []string{"Attached: report.md (file)"},
			SetupFiles: map[string]string{"report.md": "q3 up"}},
		{Name: "attachment-missing", Args: []string{"summarize @nope.md"}, WantExit: 1, Contains: []string{"couldn't find the attachment"}},
		{Name: "help", Args: []string{"help"}, Contains: []string{"Usage:"}},
		{Name: "doctor", Args: []string{"doctor"}, Contains: []string{"Provider", "Status"}},
		{Name: "route-list", Args: []string{"route", "list"}, Contains: []string{"claude-code"}},
		{Name: "route-status", Args: []string{"route", "status"}, Contains: []string{"Route and cost"}},
		{Name: "savings", Args: []string{"savings"}, Contains: []string{"savings"}},
		{Name: "trust-list", Args: []string{"trust", "list"}},
		{Name: "slash-unknown", Args: []string{"/florblegorb"}, WantExit: 1, Contains: []string{"Unknown command"}},
		{Name: "unicode-nonenglish", Args: []string{"¿cuál es la capital de España?"}},
	}
}

// checkSelfCheckResult evaluates one scenario's captured result. Pure: shared by
// the command and the test so both judge a scenario identically.
func checkSelfCheckResult(sc selfCheckScenario, exit int, out, dir string) (bool, string) {
	if exit != sc.WantExit {
		return false, fmt.Sprintf("exit=%d want=%d", exit, sc.WantExit)
	}
	for _, want := range sc.Contains {
		if !strings.Contains(out, want) {
			return false, fmt.Sprintf("missing %q", want)
		}
	}
	for _, bad := range sc.NotContains {
		if strings.Contains(out, bad) {
			return false, fmt.Sprintf("unexpected %q", bad)
		}
	}
	for name, want := range sc.FileContains {
		b, _ := os.ReadFile(filepath.Join(dir, name))
		if !strings.Contains(string(b), want) {
			return false, fmt.Sprintf("%s missing %q", name, want)
		}
	}
	return true, ""
}

// runFunctionalSelfCheck runs every scenario and reports health. Returns 0 only
// if all pass. Offline and side-effect-free on the real workspace.
func runFunctionalSelfCheck(stdout, stderr io.Writer) int {
	scenarios := selfCheckScenarios()
	passed := 0
	for _, sc := range scenarios {
		ok, detail := runSelfCheckScenario(sc)
		if ok {
			passed++
			fmt.Fprintf(stdout, "ok    %s\n", sc.Name)
			continue
		}
		fmt.Fprintf(stdout, "FAIL  %s — %s\n", sc.Name, detail)
	}
	fmt.Fprintf(stdout, "\nFunctional self-check: %d/%d scenarios passed.\n", passed, len(scenarios))
	if passed != len(scenarios) {
		return 1
	}
	return 0
}

// runSelfCheckScenario runs one scenario in a throwaway cwd + home with the
// route forced offline, restoring process state afterward.
func runSelfCheckScenario(sc selfCheckScenario) (bool, string) {
	work, err := os.MkdirTemp("", "jini-selfcheck-")
	if err != nil {
		return false, err.Error()
	}
	defer os.RemoveAll(work)
	home, err := os.MkdirTemp("", "jini-selfcheck-home-")
	if err != nil {
		return false, err.Error()
	}
	defer os.RemoveAll(home)
	for name, content := range sc.SetupFiles {
		if err := os.WriteFile(filepath.Join(work, name), []byte(content), 0o644); err != nil {
			return false, err.Error()
		}
	}

	restore := isolateSelfCheckEnv(work, home)
	defer restore()

	var out, errb bytes.Buffer
	exit := RunInteractive(sc.Args, strings.NewReader(""), &out, &errb)
	return checkSelfCheckResult(sc, exit, out.String()+errb.String(), work)
}

// isolateSelfCheckEnv points cwd, home, state, and the route at throwaway,
// offline locations and returns a function that restores the prior values.
func isolateSelfCheckEnv(work, home string) func() {
	prevWd, _ := os.Getwd()
	type saved struct {
		val string
		set bool
	}
	prev := map[string]saved{}
	setEnv := func(k, v string) {
		old, ok := os.LookupEnv(k)
		prev[k] = saved{old, ok}
		_ = os.Setenv(k, v)
	}
	setEnv("JINI_PROVIDER", "local-preview")
	setEnv("JINI_STATE_DIR", filepath.Join(home, "state"))
	setEnv("HOME", home)
	_ = os.Chdir(work)
	return func() {
		_ = os.Chdir(prevWd)
		for k, s := range prev {
			if s.set {
				_ = os.Setenv(k, s.val)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}
