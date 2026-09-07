package app

import "strings"

// Coding-intent detection — the shared signal that keeps demo/starter intercepts
// (follow-up email, travel plan, repo-review snapshot) from swallowing a real
// coding task. A prompt like "write a Go function to send a follow-up email"
// contains email words but its DELIVERABLE is code, so it must route to the
// coding agent, not a canned email template. Conservative by design: it fires
// only on strong code signals, so a genuine "email summarizing the API changes"
// (which merely mentions code) still reaches the compact email path.

// codingArtifactNouns name a code deliverable strongly enough that the request
// is about producing code. Deliberately excludes ambiguous words (api, email,
// file, service, test) that appear as often in non-coding asks.
var codingArtifactNouns = []string{
	"function", "method", "class", "struct", "interface", "enum",
	"endpoint", "handler", "middleware", "script", "program", "module",
	"component", "goroutine", "closure", "pointer", "compiler", "binary",
	"daemon", "regex", "schema", "query", "migration", "unit test",
}

// codingLanguages are unambiguous language/tech tokens (word-boundary matched).
var codingLanguages = []string{
	"golang", "python", "javascript", "typescript", "rust", "kotlin",
	"swift", "scala", "haskell", "css", "html", "sql",
}

// codingFileExtensions signal a code file target anywhere in the raw text.
var codingFileExtensions = []string{
	".go", ".py", ".js", ".ts", ".rs", ".java", ".rb", ".sql", ".html",
	".css", ".cpp", ".c ", ".sh", ".yaml", ".json", ".tsx", ".jsx",
}

// promptRequestsCodeArtifact reports whether the prompt's DELIVERABLE is code —
// a code-artifact noun, a language, or a file extension. This is the signal the
// demo/starter intercepts use to step aside, because it survives email/travel
// prose: a bare mutation verb ("fix") can appear inside an email's summarized
// content ("shipped login fix"), but "write a function" / ".go" name a code
// deliverable. Deliberately does NOT include mutation verbs for that reason.
func promptRequestsCodeArtifact(source string) bool {
	padded := " " + normalizeName(source) + " "
	for _, noun := range codingArtifactNouns {
		if strings.Contains(padded, " "+noun+" ") {
			return true
		}
	}
	for _, lang := range codingLanguages {
		if strings.Contains(padded, " "+lang+" ") {
			return true
		}
	}
	lower := strings.ToLower(source)
	for _, ext := range codingFileExtensions {
		if strings.Contains(lower, ext) {
			return true
		}
	}
	return false
}

// promptLooksLikeCodingTask is the broad signal (mutation verb OR code
// artifact) for callers that want to know the prompt is about code at all.
func promptLooksLikeCodingTask(source string) bool {
	return promptRequestsCodeMutation(normalizeName(source)) || promptRequestsCodeArtifact(source)
}
