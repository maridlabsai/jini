package app

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// `jini feedback [bug|ask] <message>` — frictionless, in-terminal filing that
// bridges to GitHub, where votes and status live. It records the item locally
// (the user's own log, and the seam for opt-in aggregate signal) and prints a
// PREFILLED GitHub link so submitting is one click and others can upvote with a
// reaction. It never auto-submits — filing off-machine is the user's choice.

const feedbackRepoSlug = "maridlabsai/jini"

func runFeedback(args []string, stdout, stderr io.Writer) int {
	kind := "ask"
	rest := args
	if len(args) > 0 {
		switch exactCommandToken(args[0]) {
		case "bug", "issue":
			kind, rest = "bug", args[1:]
		case "ask", "idea", "feature":
			kind, rest = "ask", args[1:]
		}
	}
	message := strings.TrimSpace(strings.Join(rest, " "))
	if message == "" {
		fmt.Fprintln(stderr, `Usage: jini feedback [bug|ask] "<what you hit or want>"`)
		return 1
	}

	// Non-sensitive context only — never the prompt, code, or file contents.
	context := fmt.Sprintf("jini %s (%s) · %s/%s", jiniVersion(), jiniChannel(), runtime.GOOS, runtime.GOARCH)
	if err := appendFeedbackLog(kind, message, context); err != nil {
		// A logging failure must not block the user from filing on GitHub.
		fmt.Fprintf(stderr, "note: could not write local feedback log: %v\n", err)
	}

	title, body, submitURL, browseURL := feedbackGitHubLinks(kind, message, context)
	_ = title
	fmt.Fprintf(stdout, "Recorded your %s locally.\n", kind)
	fmt.Fprintln(stdout, "Submit it so others can upvote:")
	fmt.Fprintf(stdout, "  %s\n", submitURL)
	fmt.Fprintln(stdout, "Or search existing ones first and add a 👍 to upvote:")
	fmt.Fprintf(stdout, "  %s\n", browseURL)
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "%s\n\n%s\n", title, body)
	return 0
}

// feedbackGitHubLinks builds the prefilled submit URL and a browse/upvote URL.
// Bugs go to Issues (labelled `bug`); asks go to Issues labelled `enhancement`
// so they carry 👍 reactions for ranking (Discussions don't accept a prefilled
// body via URL).
func feedbackGitHubLinks(kind, message, context string) (title, body, submitURL, browseURL string) {
	summary := message
	if len(summary) > 72 {
		summary = strings.TrimSpace(summary[:72]) + "…"
	}
	label := "enhancement"
	titlePrefix := "Ask"
	if kind == "bug" {
		label, titlePrefix = "bug", "Bug"
	}
	title = fmt.Sprintf("%s: %s", titlePrefix, summary)
	body = message + "\n\n---\n" + context
	base := "https://github.com/" + feedbackRepoSlug
	q := url.Values{}
	q.Set("title", title)
	q.Set("body", body)
	q.Set("labels", label)
	submitURL = base + "/issues/new?" + q.Encode()
	browseURL = base + "/issues?q=" + url.QueryEscape("is:issue is:open label:"+label+" sort:reactions-+1-desc")
	return title, body, submitURL, browseURL
}

func feedbackLogPath() string {
	return filepath.Join(sessionStateRoot(), "feedback.log")
}

func appendFeedbackLog(kind, message, context string) error {
	path := feedbackLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	line := fmt.Sprintf("%s\t%s\t%s\t%s\n",
		time.Now().UTC().Format(time.RFC3339), kind,
		strings.ReplaceAll(message, "\n", " "), context)
	_, err = f.WriteString(line)
	return err
}
