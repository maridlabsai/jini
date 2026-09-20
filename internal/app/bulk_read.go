package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// jini bulk-read — #3 (hook-enforced cheap delegation, the Spotify pattern) of
// specs/pre-viral-readiness.md. Reads files, wraps them into a compact tagged bundle
// with the question, and delegates the summarization to the CHEAPEST route — so a
// premium model (or an expensive agent Jini proxies to) never ingests the raw files
// and follow-up turns cost nothing extra. Returns bullets. A PreToolUse hook can call
// this to keep the premium context lean (~90% fewer tokens on large reads), which is
// also an immediate win for our own dogfood.

const bulkReadMaxBytesPerFile = 200_000

const bulkReadSystemPrompt = "You are a precise summarizer. Answer only from the provided files. Output bullets, one per line, each starting with the file name or a line number. No preamble, no code fences."

// bulkReadDelegate sends the wrapped bundle to a cheap route and returns the summary.
// Overridable so tests never call a provider.
var bulkReadDelegate = bulkReadDelegateLive

func runBulkRead(args []string, stdout, stderr io.Writer) int {
	question := "Summarize the key points, one bullet per line."
	var files []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--question" && i+1 < len(args):
			question = args[i+1]
			i++
		case strings.HasPrefix(a, "--question="):
			question = strings.TrimPrefix(a, "--question=")
		case strings.HasPrefix(a, "--"):
			// ignore unknown flags leniently
		default:
			files = append(files, a)
		}
	}
	if len(files) == 0 {
		fmt.Fprintln(stderr, `Usage: jini bulk-read <file>... [--question "<q>"]`)
		return 1
	}
	bundle, err := bulkReadBundle(files, question)
	if err != nil {
		fmt.Fprintf(stderr, "bulk-read: %v\n", err)
		return 1
	}
	summary, derr := bulkReadDelegate(context.Background(), bundle)
	if derr != nil {
		// No cheap route available: emit the bundle so the caller can still use it
		// (e.g. pipe it to a model itself) rather than failing.
		fmt.Fprintf(stderr, "bulk-read: no cheap route available (%v); emitting the bundle.\n", derr)
		fmt.Fprintln(stdout, bundle)
		return 0
	}
	fmt.Fprintln(stdout, strings.TrimSpace(summary))
	return 0
}

// bulkReadBundle wraps the files in tags with the question — the compact, bounded
// context a cheap model consumes.
func bulkReadBundle(files []string, question string) (string, error) {
	var b strings.Builder
	b.WriteString("Answer the question using only the files below.\n\n")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", f, err)
		}
		if len(data) > bulkReadMaxBytesPerFile {
			data = append(data[:bulkReadMaxBytesPerFile], []byte("\n…(truncated)")...)
		}
		fmt.Fprintf(&b, "<file path=%q>\n%s\n</file>\n\n", filepath.Clean(f), string(data))
	}
	fmt.Fprintf(&b, "Question: %s\n", strings.TrimSpace(question))
	return b.String(), nil
}

func bulkReadDelegateLive(ctx context.Context, bundle string) (string, error) {
	// Resolve the cheapest available route (auto), the same path jini check model
	// uses to reach a live provider.
	decision := enrichRouteDecisionForRequest(providerGenerationRequest{}, detectRouteForToolMode("auto", true))
	provider := providerForDecision(providerGenerationRequest{}, decision)
	if provider.ID == "local-preview" || provider.Status != "ok" {
		return "", fmt.Errorf("no ready cheap route")
	}
	req := providerGenerationRequest{Source: bundle, Title: "bulk-read", Standalone: true}
	return generateProviderText(ctx, provider, req, bulkReadSystemPrompt, bundle)
}
