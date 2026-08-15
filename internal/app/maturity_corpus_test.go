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
		prompt("math", "what is 17 times 23?"),
		prompt("typo-simple", "waht is 2 plus 2"),
		prompt("prose", "write a haiku about autumn leaves"),
		prompt("code-work", "refactor the database connection pool for reuse"),
		prompt("data", "summarize the columns in a sales csv"),
		prompt("devops", "set up a github actions workflow for go tests"),
		prompt("ambiguous-entity", "React"),
		prompt("file-edit-intent", "add a line saying hello to notes.txt"),
		prompt("unicode-nonenglish", "¿cuál es la capital de España?"),
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
