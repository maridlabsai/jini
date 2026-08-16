package app

// Attachment intake — lets a prompt carry file/image references alongside text
// (e.g. `jini "summarize @report.md and @chart.png"`). Jini validates the
// references up front (fail closed on a missing path, with an exact message),
// acknowledges them, and forwards the prompt verbatim to the routed CLI —
// Claude Code and Codex read `@path` references (including images) natively, so
// Jini leverages them rather than reimplementing vision. For local/provider
// routes without multimodal wiring, text attachments can be inlined and images
// are flagged honestly as referenced-but-not-read. See
// specs/product-maturity-coverage.md.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Attachment limits keep token cost bounded (P0 frugality) and prevent abuse.
const (
	maxAttachments        = 10               // references honored per prompt
	maxInlineAttachmentB  = 64 * 1024        // per-file cap when inlining text
	maxTotalInlineB       = 192 * 1024       // combined inline budget across files
	inlineTruncatedNotice = "\n…[truncated]" // appended when a file is capped
)

type attachmentRef struct {
	Raw  string // the token as written, e.g. "@report.md"
	Path string // resolved absolute path
	Kind string // "image" | "audio" | "file"
}

// parseAttachmentRefs returns the `@token` references in source: `@` followed by
// a run of non-space characters, with trailing sentence punctuation trimmed so
// "read @notes.txt." yields "notes.txt". "@" alone and "@@x" are skipped.
func parseAttachmentRefs(source string) []string {
	var refs []string
	for _, tok := range strings.Fields(source) {
		if !strings.HasPrefix(tok, "@") || len(tok) < 2 {
			continue
		}
		path := strings.TrimPrefix(tok, "@")
		path = strings.TrimRight(path, ".,;:!?)")
		if path == "" || strings.HasPrefix(path, "@") {
			continue
		}
		refs = append(refs, path)
	}
	return refs
}

// pathLike is true for tokens that clearly denote a file path (extension or
// separator or absolute), so a social handle like "@alice" is never mistaken
// for a missing attachment.
func pathLike(token string) bool {
	return filepath.IsAbs(token) || strings.ContainsAny(token, "/\\") || strings.Contains(token, ".")
}

// resolveAttachments validates the `@path` references in source against cwd. A
// reference that resolves to an existing file becomes an attachment; a
// path-like reference that does not resolve is reported as missing (caller
// fails closed); a non-path-like token that does not resolve (a handle) is
// ignored. At most maxAttachments are honored.
func resolveAttachments(source, cwd string) (found []attachmentRef, missing []string) {
	for _, raw := range parseAttachmentRefs(source) {
		if len(found) >= maxAttachments {
			break
		}
		path := raw
		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			if pathLike(raw) {
				missing = append(missing, raw)
			}
			continue
		}
		found = append(found, attachmentRef{
			Raw:  "@" + raw,
			Path: path,
			Kind: classifyInputKind(filepath.Base(path)),
		})
	}
	return found, missing
}

// inlineTextAttachments appends the text content of file attachments to a prompt
// so a local/provider model without native file access can read them. Images
// and audio are skipped (referenced-but-not-read on non-handoff routes). Each
// file is capped at maxInlineAttachmentB and the combined budget at
// maxTotalInlineB; overflow is truncated with a visible notice.
func inlineTextAttachments(prompt string, found []attachmentRef) string {
	var b strings.Builder
	b.WriteString(prompt)
	budget := maxTotalInlineB
	for _, a := range found {
		if a.Kind != "file" || budget <= 0 {
			continue
		}
		data, err := os.ReadFile(a.Path)
		if err != nil {
			continue
		}
		truncated := false
		limit := maxInlineAttachmentB
		if limit > budget {
			limit = budget
		}
		if len(data) > limit {
			data = data[:limit]
			truncated = true
		}
		budget -= len(data)
		fmt.Fprintf(&b, "\n\n--- %s ---\n%s", filepath.Base(a.Path), string(data))
		if truncated {
			b.WriteString(inlineTruncatedNotice)
		}
	}
	return b.String()
}

// missingAttachmentError is the exact, neutral fail-closed message.
func missingAttachmentError(missing []string) string {
	if len(missing) == 1 {
		return fmt.Sprintf("I couldn't find the attachment %q. Check the path and try again.", missing[0])
	}
	return fmt.Sprintf("I couldn't find these attachments: %s. Check the paths and try again.", strings.Join(quoteAll(missing), ", "))
}

func quoteAll(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = fmt.Sprintf("%q", s)
	}
	return out
}

// attachmentAckLine acknowledges resolved attachments in a single compact line,
// or "" when there are none. Kept terse to respect token frugality.
func attachmentAckLine(found []attachmentRef) string {
	if len(found) == 0 {
		return ""
	}
	parts := make([]string, len(found))
	for i, a := range found {
		parts[i] = fmt.Sprintf("%s (%s)", filepath.Base(a.Path), a.Kind)
	}
	return "Attached: " + strings.Join(parts, ", ")
}

// hasImageAttachment reports whether any resolved attachment is an image, so a
// non-handoff route can note honestly that it won't read the pixels.
func hasImageAttachment(found []attachmentRef) bool {
	for _, a := range found {
		if a.Kind == "image" || a.Kind == "audio" {
			return true
		}
	}
	return false
}
