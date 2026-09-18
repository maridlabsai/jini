package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// jini verify — objective, language-aware workspace verification. Phase 1 of the
// confidence-routing + verification design (specs/pre-viral-readiness.md): after a
// coding change, prove it objectively (does it build? do tests pass?) instead of
// trusting the model's word. This is the reusable core the escalation loop will
// call on task completion; on its own it lets anyone verify a workspace and
// doubles as a gate — exit 0 = verified, 1 = a check failed, 2 = nothing runnable.

type verifyCheck struct {
	Name    string   `json:"name"`
	Command []string `json:"command,omitempty"`
	Passed  bool     `json:"passed"`
	Skipped bool     `json:"skipped"`
	Detail  string   `json:"detail,omitempty"`
}

type verifyResult struct {
	Dir      string        `json:"dir"`
	Checks   []verifyCheck `json:"checks"`
	Ran      int           `json:"ran"`
	Failed   int           `json:"failed"`
	Verified bool          `json:"verified"`
}

// verifyCommandRunner runs one verification command in dir and reports pass +
// (on failure) a short detail. Overridable so the suite never shells out.
var verifyCommandRunner = runVerifyCommandLive

func runVerifyCommandLive(ctx context.Context, dir string, args []string) (bool, string) {
	if len(args) == 0 {
		return false, "empty command"
	}
	if _, err := exec.LookPath(args[0]); err != nil {
		return false, fmt.Sprintf("%s not found on PATH", args[0])
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		if detail == "" {
			detail = err.Error()
		}
		if len(detail) > 500 { // keep the tail, where the error usually is
			detail = "…" + detail[len(detail)-500:]
		}
		return false, detail
	}
	return true, ""
}

func verifyHasFile(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func verifyNodeHasScript(dir, script string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, ok := pkg.Scripts[script]
	return ok
}

// verifyPlan builds the ordered objective checks for a workspace by detected
// project type. Buildability is the default signal; includeTests adds the slower
// test checks. First recognized ecosystem wins (avoids double-running in polyglot
// repos; the loop can target a sub-dir).
func verifyPlan(dir string, includeTests bool) []verifyCheck {
	var checks []verifyCheck
	add := func(name string, cmd ...string) {
		checks = append(checks, verifyCheck{Name: name, Command: cmd})
	}
	switch {
	case verifyHasFile(dir, "go.mod"):
		add("go build", "go", "build", "./...")
		add("go vet", "go", "vet", "./...")
		if includeTests {
			add("go test", "go", "test", "./...")
		}
	case verifyHasFile(dir, "Cargo.toml"):
		add("cargo build", "cargo", "build", "--quiet")
		if includeTests {
			add("cargo test", "cargo", "test", "--quiet")
		}
	case verifyHasFile(dir, "package.json"):
		switch {
		case verifyHasFile(dir, "tsconfig.json"):
			add("tsc --noEmit", "npx", "--no-install", "tsc", "--noEmit")
		case verifyNodeHasScript(dir, "build"):
			add("npm run build", "npm", "run", "build")
		}
		if includeTests && verifyNodeHasScript(dir, "test") {
			add("npm test", "npm", "test")
		}
	case verifyHasFile(dir, "pyproject.toml") || verifyHasFile(dir, "setup.py"):
		add("py compile", "python3", "-m", "compileall", "-q", ".")
	}
	return checks
}

// verifyWorkspace runs the plan for dir and aggregates an objective verdict.
func verifyWorkspace(ctx context.Context, dir string, includeTests bool) verifyResult {
	result := verifyResult{Dir: dir}
	plan := verifyPlan(dir, includeTests)
	if len(plan) == 0 {
		result.Checks = []verifyCheck{{
			Name:    "no recognized build system",
			Skipped: true,
			Detail:  "no go.mod / Cargo.toml / package.json / pyproject.toml / setup.py found",
		}}
		return result
	}
	for _, chk := range plan {
		ok, detail := verifyCommandRunner(ctx, dir, chk.Command)
		chk.Passed = ok
		chk.Detail = detail
		result.Ran++
		if !ok {
			result.Failed++
		}
		result.Checks = append(result.Checks, chk)
	}
	result.Verified = result.Ran > 0 && result.Failed == 0
	return result
}

// runVerify implements `jini verify [dir] [--tests] [--format json|--json]`.
// Exit 0 = verified, 1 = a check failed, 2 = nothing runnable (unverifiable).
func runVerify(args []string, stdout, stderr io.Writer) int {
	dir := "."
	includeTests := false
	asJSON := false
	for _, a := range args {
		switch a {
		case "--tests", "--test":
			includeTests = true
		case "--json", "--format=json":
			asJSON = true
		case "--format", "json":
			// lenient: tolerate `--format json`
		default:
			if !strings.HasPrefix(a, "--") {
				dir = a
			}
		}
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}

	result := verifyWorkspace(context.Background(), dir, includeTests)

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
	} else {
		renderVerifyResult(stdout, result)
	}

	switch {
	case result.Ran == 0:
		return 2
	case result.Failed > 0:
		return 1
	default:
		return 0
	}
}

func renderVerifyResult(w io.Writer, r verifyResult) {
	if r.Ran == 0 {
		detail := "no recognized build system"
		if len(r.Checks) > 0 && r.Checks[0].Detail != "" {
			detail = r.Checks[0].Detail
		}
		fmt.Fprintf(w, "Unverifiable — %s.\n", detail)
		return
	}
	for _, c := range r.Checks {
		switch {
		case c.Skipped:
			fmt.Fprintf(w, "  ~ %s (skipped)\n", c.Name)
		case c.Passed:
			fmt.Fprintf(w, "  ✓ %s\n", c.Name)
		default:
			fmt.Fprintf(w, "  ✗ %s\n", c.Name)
			if c.Detail != "" {
				fmt.Fprintf(w, "      %s\n", verifyFirstLine(c.Detail))
			}
		}
	}
	if r.Verified {
		fmt.Fprintf(w, "verified ✓ (%d checks passed)\n", r.Ran)
	} else {
		fmt.Fprintf(w, "NOT verified — %d of %d checks failed.\n", r.Failed, r.Ran)
	}
}

func verifyFirstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}
