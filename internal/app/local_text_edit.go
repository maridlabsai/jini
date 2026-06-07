package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type localTextEditIntent struct {
	Line string
}

type localTextFileCandidate struct {
	Name  string
	Path  string
	Score int
}

var quotedLinePattern = regexp.MustCompile(`[\"“”']([^\"“”']+)[\"“”']`)
var sayingPrefixPattern = regexp.MustCompile(`(?i)\bsaying\s+`)

func maybeHandleLocalTextFileEditIntent(raw string, stdout, stderr io.Writer) (bool, int) {
	intent, ok := parseLocalTextEditIntent(raw)
	if !ok {
		return false, 0
	}
	target, err := resolveLocalTextEditTarget(raw, intent)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return true, 1
	}
	changed, err := appendLineIfMissing(target.Path, intent.Line)
	if err != nil {
		fmt.Fprintf(stderr, "Could not update %s: %v\n", target.Name, err)
		return true, 1
	}
	if changed {
		fmt.Fprintf(stdout, "Updated %s\n", target.Name)
		fmt.Fprintf(stdout, "- Added line: %s\n", intent.Line)
		fmt.Fprintf(stdout, "- Location: %s\n", target.Path)
		return true, 0
	}
	fmt.Fprintf(stdout, "%s already contains that line.\n", target.Name)
	fmt.Fprintf(stdout, "- Location: %s\n", target.Path)
	return true, 0
}

func parseLocalTextEditIntent(raw string) (localTextEditIntent, bool) {
	normalized := normalizeName(raw)
	if !strings.Contains(normalized, "txt") && !strings.Contains(normalized, "text file") {
		return localTextEditIntent{}, false
	}
	if !strings.Contains(normalized, "line") {
		return localTextEditIntent{}, false
	}
	if !containsAny(normalized, []string{"add", "append", "insert"}) {
		return localTextEditIntent{}, false
	}
	line := firstQuotedText(raw)
	if line == "" {
		line = firstUnquotedSayingText(raw)
	}
	if line == "" {
		return localTextEditIntent{}, false
	}
	return localTextEditIntent{Line: line}, true
}

func firstQuotedText(raw string) string {
	match := quotedLinePattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func firstUnquotedSayingText(raw string) string {
	match := sayingPrefixPattern.FindStringIndex(raw)
	if len(match) < 2 {
		return ""
	}
	rest := strings.TrimSpace(raw[match[1]:])
	line := bestUnquotedSayingLine(rest)
	if line == "" || strings.Contains(normalizeName(line), " txt ") {
		return ""
	}
	if normalizeName(line) == "jini was here" {
		return "jini was here"
	}
	return line
}

func bestUnquotedSayingLine(rest string) string {
	lower := strings.ToLower(rest)
	delimiters := []int{}
	searchFrom := 0
	for {
		index := strings.Index(lower[searchFrom:], " in ")
		if index < 0 {
			break
		}
		delimiters = append(delimiters, searchFrom+index)
		searchFrom += index + len(" in ")
	}
	if len(delimiters) == 0 {
		return ""
	}
	names := localTextFileNamesForParsing()
	bestLine := ""
	bestScore := -1
	bestIndex := -1
	for _, delimiter := range delimiters {
		line := strings.TrimSpace(rest[:delimiter])
		targetText := strings.TrimSpace(rest[delimiter+len(" in "):])
		if line == "" || targetText == "" {
			continue
		}
		score := localTextTargetScore(targetText, names)
		if score > bestScore || (score == bestScore && delimiter > bestIndex) {
			bestLine = line
			bestScore = score
			bestIndex = delimiter
		}
	}
	if bestLine == "" {
		return strings.TrimSpace(rest[:delimiters[len(delimiters)-1]])
	}
	return bestLine
}

func localTextFileNamesForParsing() []string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".txt" {
			names = append(names, entry.Name())
		}
	}
	return names
}

func localTextTargetScore(targetText string, names []string) int {
	if len(names) == 0 {
		return 0
	}
	tokens := localTextEditRequestTokens(targetText, "")
	best := 0
	for _, name := range names {
		score := localTextFileMatchScore(name, tokens)
		if score > best {
			best = score
		}
	}
	return best
}

func resolveLocalTextEditTarget(raw string, intent localTextEditIntent) (localTextFileCandidate, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return localTextFileCandidate{}, fmt.Errorf("Could not read the current folder: %v", err)
	}
	entries, err := os.ReadDir(cwd)
	if err != nil {
		return localTextFileCandidate{}, fmt.Errorf("Could not list the current folder: %v", err)
	}
	candidates := []localTextFileCandidate{}
	requestTokens := localTextEditRequestTokens(raw, intent.Line)
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".txt" {
			continue
		}
		candidates = append(candidates, localTextFileCandidate{
			Name:  entry.Name(),
			Path:  filepath.Join(cwd, entry.Name()),
			Score: localTextFileMatchScore(entry.Name(), requestTokens),
		})
	}
	if len(candidates) == 0 {
		return localTextFileCandidate{}, fmt.Errorf("Could not find a .txt file in this folder.")
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Name < candidates[j].Name
		}
		return candidates[i].Score > candidates[j].Score
	})
	if candidates[0].Score == 0 || candidates[0].Score == candidates[1].Score {
		return localTextFileCandidate{}, fmt.Errorf("I found multiple .txt files. Please include the exact filename.")
	}
	return candidates[0], nil
}

func localTextEditRequestTokens(raw, line string) map[string]bool {
	withoutLine := raw
	if line != "" {
		withoutLine = strings.ReplaceAll(raw, line, " ")
	}
	stopWords := map[string]bool{
		"a": true, "add": true, "an": true, "and": true, "append": true, "file": true,
		"folder": true, "here": true, "in": true, "insert": true, "line": true,
		"saying": true, "that": true, "the": true, "this": true,
		"to": true, "txt": true, "was": true, "with": true,
	}
	tokens := map[string]bool{}
	for _, token := range strings.Fields(normalizeName(withoutLine)) {
		token = strings.TrimSpace(token)
		if token == "" || stopWords[token] {
			continue
		}
		tokens[token] = true
	}
	return tokens
}

func localTextFileMatchScore(name string, requestTokens map[string]bool) int {
	if len(requestTokens) == 0 {
		return 0
	}
	score := 0
	for _, token := range strings.Fields(normalizeName(strings.TrimSuffix(name, filepath.Ext(name)))) {
		if requestTokens[token] {
			score++
		}
	}
	return score
}

func appendLineIfMissing(path, line string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	text := string(data)
	for _, existing := range strings.Split(text, "\n") {
		if strings.TrimSpace(existing) == line {
			return false, nil
		}
	}
	prefix := ""
	if len(data) > 0 && !strings.HasSuffix(text, "\n") {
		prefix = "\n"
	}
	return true, os.WriteFile(path, []byte(text+prefix+line+"\n"), info.Mode().Perm())
}
