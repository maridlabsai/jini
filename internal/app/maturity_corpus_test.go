package app

// Black-box maturity corpus — drives a battery of request kinds across domains
// and task complexity through the real CLI entrypoint and enforces the hard
// user-facing-output invariants on every result: no ANSI/control chars
// (accessibility) and no leaked credentials (security). Tone and readability
// enforcement over the corpus is phased in separately (see
// specs/product-maturity-coverage.md). Runs fully offline: JINI_PROVIDER is
// pinned to local-preview and TestMain pins hand-off CLIs to a missing path.

import (
	"bytes"
	"strings"
	"testing"
)

type corpusCase struct {
	name string
	args []string // command or direct-task prompt (as arg tokens)
}

func maturityCorpus() []corpusCase {
	prompt := func(name, p string) corpusCase { return corpusCase{name, strings.Fields(p)} }
	cmd := func(name string, a ...string) corpusCase { return corpusCase{name, a} }
	return []corpusCase{
		// Command surfaces (deterministic Jini copy).
		cmd("help", "help"),
		cmd("commands", "commands"),
		cmd("doctor", "doctor"),
		cmd("route-list", "route", "list"),
		cmd("route-status", "route", "status"),
		cmd("status", "status"),
		cmd("savings", "savings"),
		cmd("mode", "mode"),
		cmd("trust", "trust"),
		// Request kinds × domains × complexity (direct-task intake).
		prompt("simple-factual", "what is the capital of France?"),
		prompt("math-symbol", "what is 17 * 23?"),
		prompt("math-word", "what is 17 times 23"),
		prompt("typo-simple", "waht is 2 plus 2"),
		prompt("prose", "write a haiku about autumn leaves"),
		prompt("code-work", "refactor the database connection pool for reuse"),
		prompt("sql", "write a sql query for the top 10 customers by revenue"),
		prompt("regex", "write a regex to match an email address"),
		prompt("git", "how do I undo the last commit"),
		prompt("shell", "find all files larger than 100mb here"),
		prompt("docker", "write a dockerfile for a go service"),
		prompt("k8s", "create a kubernetes deployment for nginx"),
		prompt("security", "how do I store api keys safely"),
		prompt("translate", "translate good morning to japanese"),
		prompt("explain", "explain what a mutex is"),
		prompt("compare", "compare rest and grpc"),
		prompt("debug", "why might a go program deadlock"),
		prompt("data", "summarize the columns in a sales csv"),
		prompt("devops", "set up a github actions workflow for go tests"),
		prompt("multi-step", "add a parser test, run it, and fix any failure"),
		prompt("ambiguous-entity", "React"),
		prompt("file-edit-intent", "add a line saying hello to notes.txt"),
		prompt("unicode-nonenglish", "¿cuál es la capital de España?"),
	}
}

// trivialPrompts resolve locally to a compact answer and must never carry
// work-draft ceremony — the intent-first-cli-parity bar (PRD non-negotiable).
func trivialPrompts() []corpusCase {
	p := func(name, s string) corpusCase { return corpusCase{name, strings.Fields(s)} }
	return []corpusCase{
		p("plus", "what is 2 plus 2"),
		p("times-word", "what is 17 times 23"),
		p("divide-word", "what is 100 divided by 4"),
		p("minus-word", "what is 10 minus 4"),
		p("symbol", "what is 8 * 9"),
		p("capital", "what is the capital of France?"),
	}
}

var ceremonyMarkers = []string{
	"Saved a restorable version", "Versions", "Undo",
	"Start/Keep", "snapshot", "Saved draft", "draft saved",
}

func TestMaturityCorpus_TrivialPromptsAreCompact(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")
	for _, tc := range trivialPrompts() {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			t.Chdir(t.TempDir())

			var stdout, stderr bytes.Buffer
			if code := RunInteractive(tc.args, strings.NewReader(""), &stdout, &stderr); code != 0 {
				t.Fatalf("trivial prompt should answer cleanly, code=%d err=%q", code, stderr.String())
			}
			out := strings.TrimSpace(stdout.String())
			lines := strings.Split(out, "\n")
			if len(lines) != 1 {
				t.Fatalf("trivial answer must be one compact line, got %d:\n%q", len(lines), out)
			}
			for _, m := range ceremonyMarkers {
				if strings.Contains(stdout.String(), m) {
					t.Fatalf("trivial prompt %q leaked ceremony marker %q", tc.name, m)
				}
			}
		})
	}
}

func TestMaturityCorpus_ToneAndReadability(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")
	full := auditOptions{} // tone + readability enabled

	for _, tc := range maturityCorpus() {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			t.Chdir(t.TempDir())

			var stdout, stderr bytes.Buffer
			RunInteractive(tc.args, strings.NewReader(""), &stdout, &stderr)
			combined := stdout.String() + "\n" + stderr.String()

			var soft []outputIssue
			for _, is := range auditUserFacingOutput(combined, full) {
				if is.Kind == issueFearOrHype || is.Kind == issueOverlongLine {
					soft = append(soft, is)
				}
			}
			if len(soft) != 0 {
				t.Fatalf("tone/readability issue for %q:\n%+v\noutput=%q", tc.name, soft, combined)
			}
		})
	}
}

// Token-frugality budgets (chars as a token proxy). Tight where competitors are
// compact (trivial answers); a generous global ceiling catches runaway bloat.
const (
	trivialAnswerCharBudget = 40
	corpusOutputCharCeiling = 4000
)

func TestMaturityCorpus_TokenFrugality(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")

	// Trivial answers stay tiny and never replay the question (transcript-replay
	// avoidance) — the token-frugality-p0 competitive bar.
	for _, tc := range trivialPrompts() {
		t.Run("trivial/"+tc.name, func(t *testing.T) {
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			t.Chdir(t.TempDir())
			var out, errBuf bytes.Buffer
			RunInteractive(tc.args, strings.NewReader(""), &out, &errBuf)
			answer := strings.TrimSpace(out.String())
			if len(answer) > trivialAnswerCharBudget {
				t.Fatalf("trivial answer over budget (%d > %d): %q", len(answer), trivialAnswerCharBudget, answer)
			}
			if strings.Contains(strings.ToLower(answer), "what is") {
				t.Fatalf("trivial answer replays the question: %q", answer)
			}
		})
	}

	// No corpus case may produce a runaway wall of output.
	for _, tc := range maturityCorpus() {
		t.Run("ceiling/"+tc.name, func(t *testing.T) {
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			t.Chdir(t.TempDir())
			var out, errBuf bytes.Buffer
			RunInteractive(tc.args, strings.NewReader(""), &out, &errBuf)
			total := out.Len() + errBuf.Len()
			if total > corpusOutputCharCeiling {
				t.Fatalf("output for %q bloated to %d chars (ceiling %d)", tc.name, total, corpusOutputCharCeiling)
			}
		})
	}
}

func TestMaturityCorpus_HardInvariants(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")
	hard := auditOptions{SkipTone: true, SkipReadab: true}

	for _, tc := range maturityCorpus() {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JINI_STATE_DIR", t.TempDir())
			t.Chdir(t.TempDir())

			var stdout, stderr bytes.Buffer
			RunInteractive(tc.args, strings.NewReader(""), &stdout, &stderr)
			combined := stdout.String() + "\n" + stderr.String()

			issues := auditUserFacingOutput(combined, hard)
			if hasSecurityOrAccessibilityIssue(issues) {
				t.Fatalf("hard invariant violated for %q:\nissues=%+v\noutput=%q", tc.name, issues, combined)
			}
		})
	}
}
