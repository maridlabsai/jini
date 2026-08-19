package app

// respguard — user-facing output quality guards. Pure detectors that encode
// Jini's maturity bar for anything printed to a user: accessibility (plain text,
// no ANSI/control chars), security (no leaked secrets), readability (no walls of
// text / overlong lines), and tone (neutral — no fear or hype). See
// specs/product-maturity-coverage.md.
//
// Detection-only: these functions never mutate output. Tests enforce them as
// invariants over real command output; a future redaction pass may reuse the
// secret detector at runtime. Kept conservative to avoid flagging legitimate CLI
// copy — flag names in backticks, technical terms, and placeholders are exempt.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type outputIssueKind string

const (
	issueANSIEscape      outputIssueKind = "ansi-escape"      // accessibility: breaks plain-text/screen readers
	issueControlChar     outputIssueKind = "control-char"     // accessibility: non-printable leaks
	issueLeakedSecret    outputIssueKind = "leaked-secret"    // security: token/key echoed back
	issueOverlongLine    outputIssueKind = "overlong-line"    // readability: unwrapped wall
	issueFearOrHype      outputIssueKind = "tone"             // tone: fear-mongering or marketing hype
	issueBrokenReference outputIssueKind = "broken-reference" // citation: a file Jini names does not exist
)

// outputIssue is one detected problem, 1-indexed line.
type outputIssue struct {
	Kind   outputIssueKind
	Line   int
	Detail string
}

// auditOptions tunes the readability/tone thresholds so callers can enforce the
// hard invariants (security, accessibility) strictly while phasing in the softer
// ones. Zero value = sensible defaults.
type auditOptions struct {
	MaxLineWidth int  // 0 → default 120; readability ceiling for a single line
	SkipTone     bool // when true, do not run the tone lexicon
	SkipReadab   bool // when true, do not run readability checks
}

// ansiEscapePattern matches CSI/OSC escape sequences (ESC [ … or ESC ] …).
var ansiEscapePattern = regexp.MustCompile("\x1b[\\[\\]][0-9;?]*[ -/]*[@-~]?")

// secretPatterns match well-known credential shapes with enough entropy that a
// real value — not a placeholder — is almost certainly present. Placeholders
// (sk-test-…, sk-…, <token>, ***) fall below the length/charset thresholds.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),                                          // OpenAI-style
	regexp.MustCompile(`sk-ant-[A-Za-z0-9_-]{20,}`),                                    // Anthropic-style
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{30,}`),                                   // GitHub tokens
	regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{20,}`),                                 // Slack tokens
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                                             // AWS access key id
	regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),                                        // Google API key
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{5,}`), // JWT
}

// fearHypeWords are marketing/fear terms that neutral CLI copy should never use.
// Deliberately narrow: only words with no legitimate technical meaning in Jini's
// surface, so a match is a real tone regression rather than a false positive.
var fearHypeWords = []string{
	"revolutionary", "blazing", "blazingly", "unleash", "supercharge",
	"game-changer", "game changer", "mind-blowing", "jaw-dropping",
	"catastrophic", "disastrous", "terrifying", "nightmare", "doomed",
	"warning!!!", "danger!!!", "irreversible damage", "you will lose everything",
}

// auditUserFacingOutput returns every quality issue found in text. Empty slice
// means the text meets the bar for the enabled checks.
func auditUserFacingOutput(text string, opts auditOptions) []outputIssue {
	if opts.MaxLineWidth == 0 {
		opts.MaxLineWidth = 120
	}
	var issues []outputIssue
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNo := i + 1

		// Accessibility: ANSI escapes.
		if ansiEscapePattern.MatchString(line) {
			issues = append(issues, outputIssue{issueANSIEscape, lineNo, "ANSI escape sequence in output"})
		}
		// Accessibility: stray control chars (allow tab; \r\n already split out).
		if idx := indexControlChar(line); idx >= 0 {
			issues = append(issues, outputIssue{issueControlChar, lineNo, "non-printable control character"})
		}
		// Security: leaked secrets.
		for _, pat := range secretPatterns {
			if m := pat.FindString(line); m != "" {
				issues = append(issues, outputIssue{issueLeakedSecret, lineNo, "looks like a real credential: " + maskSecret(m)})
				break
			}
		}
		// Readability: a run-on wall — a line that both exceeds the readable
		// width AND packs multiple sentences that should be split for
		// scannability. A single long sentence is fine (the terminal soft-wraps
		// it); unwrappable tokens (URLs/paths/code/table rows) are exempt.
		if !opts.SkipReadab && lineWidth(line) > opts.MaxLineWidth && sentenceCount(line) >= 2 && !isUnwrappableLine(line) {
			issues = append(issues, outputIssue{issueOverlongLine, lineNo, "multi-sentence line should be split for readability"})
		}
		// Tone: fear/hype lexicon (ignore text inside backticks — flag names,
		// identifiers — so `--dangerously-skip-permissions` never trips it).
		if !opts.SkipTone {
			if w := findFearHype(stripInlineCode(line)); w != "" {
				issues = append(issues, outputIssue{issueFearOrHype, lineNo, "fear/hype term: " + w})
			}
		}
	}
	return issues
}

// hasSecurityOrAccessibilityIssue reports the hard-invariant subset that must
// always hold, regardless of readability/tone phase-in.
func hasSecurityOrAccessibilityIssue(issues []outputIssue) bool {
	for _, is := range issues {
		switch is.Kind {
		case issueANSIEscape, issueControlChar, issueLeakedSecret:
			return true
		}
	}
	return false
}

func indexControlChar(line string) int {
	for i, r := range line {
		if r == '\t' {
			continue
		}
		if r == utf8.RuneError || (unicode.IsControl(r)) {
			return i
		}
	}
	return -1
}

func lineWidth(line string) int { return utf8.RuneCountInString(line) }

// sentenceCount counts sentence terminators followed by a space and a capital
// letter — i.e. sentence boundaries inside a single line. A line with 2+ such
// boundaries is a run-on that reads better split across lines.
var sentenceBoundary = regexp.MustCompile(`[.!?]\s+[A-Z(]`)

func sentenceCount(line string) int {
	return len(sentenceBoundary.FindAllString(line, -1))
}

// isUnwrappableLine is true for lines dominated by a single long token (URL,
// path, hash, base64) or a code/table row, which readability can't fault.
func isUnwrappableLine(line string) bool {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "|") || strings.HasPrefix(t, "```") || strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
		return true
	}
	fields := strings.Fields(t)
	for _, f := range fields {
		if len(f) > 60 { // a single unbreakable token
			return true
		}
	}
	return false
}

func findFearHype(line string) string {
	lower := strings.ToLower(line)
	for _, w := range fearHypeWords {
		if strings.Contains(lower, w) {
			return w
		}
	}
	return ""
}

// stripInlineCode removes `backticked` spans so identifiers/flag names inside
// them are exempt from tone checks.
func stripInlineCode(line string) string {
	var b strings.Builder
	inCode := false
	for _, r := range line {
		if r == '`' {
			inCode = !inCode
			continue
		}
		if !inCode {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// fileClaimPattern captures a path from phrases where Jini asserts it acted on
// a specific file ("Updated X", "Added a line to X", "wrote N bytes to X",
// "· edited X"). Only these claim shapes are checked — never arbitrary prose —
// so a reference is flagged solely when Jini vouches for a file that isn't there.
var fileClaimPattern = regexp.MustCompile(`(?:Updated|Edited|edited|Created|Wrote|wrote(?: \d+ bytes)?|Added(?: a)? line to|Appended to|Saved to)(?:\s+to)?\s+(\S+)`)

// concretePathRef is true for a token that denotes a specific file (extension or
// separator), excluding placeholders/URLs/globs, so example or hypothetical
// mentions are not mistaken for a broken citation.
func concretePathRef(token string) bool {
	token = strings.Trim(token, "`\"'.,;:)")
	if token == "" || strings.ContainsAny(token, "<>*?") {
		return false
	}
	if strings.HasPrefix(token, "http://") || strings.HasPrefix(token, "https://") {
		return false
	}
	return strings.ContainsAny(token, "/\\") || strings.Contains(token, ".")
}

// brokenPathReferences flags files Jini claims to have acted on that do not
// exist under baseDir — a citation/reference-integrity guard over Jini's OWN
// output (not model answer prose). Conservative by construction: only explicit
// file-claim phrases are inspected.
func brokenPathReferences(text, baseDir string) []outputIssue {
	var issues []outputIssue
	for i, line := range strings.Split(text, "\n") {
		for _, m := range fileClaimPattern.FindAllStringSubmatch(line, -1) {
			raw := strings.Trim(m[1], "`\"'.,;:)")
			if !concretePathRef(raw) {
				continue
			}
			path := raw
			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, raw)
			}
			if _, err := os.Stat(path); err != nil {
				issues = append(issues, outputIssue{issueBrokenReference, i + 1, "names a file that does not exist: " + raw})
			}
		}
	}
	return issues
}

// maskSecret shows only a short prefix so the detector's own message never
// reprints the full credential.
func maskSecret(s string) string {
	if len(s) <= 6 {
		return "******"
	}
	return s[:4] + "…" + strings.Repeat("*", 3)
}
