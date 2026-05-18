package app

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type smartLink struct {
	URL    string
	Labels []string
}

var markdownLinkPattern = regexp.MustCompile(`\[(?P<label>[^\]]+)\]\((?P<url>https?://[^\s)]+)\)`)
var bareURLPattern = regexp.MustCompile(`https?://[^\s)]+`)

func enrichSmartHyperlinksInViews(workDir string, request providerGenerationRequest) error {
	entries, err := os.ReadDir(filepath.Join(workDir, "views"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(workDir, "views", entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		enriched := enrichArtifactMarkdown(request, string(content))
		if enriched == string(content) {
			continue
		}
		if err := os.WriteFile(path, []byte(enriched), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func enrichArtifactMarkdown(request providerGenerationRequest, content string) string {
	links := append(sourceReferenceLinks(request.Source), packSpecificSmartLinks(request)...)
	if len(links) == 0 {
		return content
	}
	return applySmartLinks(content, links)
}

func sourceReferenceLinks(source string) []smartLink {
	if strings.TrimSpace(source) == "" {
		return nil
	}
	dedup := map[string]smartLink{}
	for _, match := range markdownLinkPattern.FindAllStringSubmatch(source, -1) {
		if len(match) < 3 {
			continue
		}
		label := strings.TrimSpace(match[1])
		target := strings.TrimSpace(match[2])
		if label == "" || target == "" {
			continue
		}
		key := strings.ToLower(label) + "|" + target
		dedup[key] = smartLink{URL: target, Labels: []string{label}}
	}
	for _, raw := range bareURLPattern.FindAllString(source, -1) {
		target := strings.TrimSpace(raw)
		if target == "" {
			continue
		}
		if label := hostLabelForURL(target); label != "" {
			key := strings.ToLower(label) + "|" + target
			dedup[key] = smartLink{URL: target, Labels: []string{label}}
		}
	}
	links := make([]smartLink, 0, len(dedup))
	for _, link := range dedup {
		links = append(links, link)
	}
	sort.SliceStable(links, func(i, j int) bool {
		return len(links[i].Labels[0]) > len(links[j].Labels[0])
	})
	return links
}

func hostLabelForURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimPrefix(parsed.Hostname(), "www."))
	switch host {
	case "github.com", "docs.github.com":
		return "GitHub"
	case "figma.com":
		return "Figma"
	case "notion.so", "www.notion.so":
		return "Notion"
	case "linear.app":
		return "Linear"
	case "openai.com", "platform.openai.com":
		return "OpenAI"
	case "anthropic.com", "docs.anthropic.com":
		return "Anthropic"
	case "atlassian.com", "support.atlassian.com":
		return "Atlassian"
	}
	return ""
}

func packSpecificSmartLinks(request providerGenerationRequest) []smartLink {
	switch request.Choice.PackID {
	case "travel-plan":
		return []smartLink{
			{URL: "https://www.louvre.fr/en", Labels: []string{"Louvre"}},
			{URL: "https://www.paris.fr/lieux/jardin-des-tuileries-1710", Labels: []string{"Tuileries", "Tuileries Garden"}},
			{URL: "https://www.paris.fr/pages/la-seine-2077", Labels: []string{"Seine"}},
			{URL: "https://www.sainte-chapelle.fr/en/", Labels: []string{"Sainte-Chapelle"}},
			{URL: "https://www.cathedrale-notredamedeparis.fr/en/", Labels: []string{"Notre-Dame", "Notre-Dame area"}},
			{URL: "https://parisjetaime.com/eng/article/the-latin-quarter-a775", Labels: []string{"Latin Quarter"}},
			{URL: "https://parisjetaime.com/eng/article/montmartre-a043", Labels: []string{"Montmartre"}},
			{URL: "https://www.sacre-coeur-montmartre.com/english/", Labels: []string{"Sacre-Coeur", "Sacré-Cœur"}},
			{URL: "https://en.chateauversailles.fr/", Labels: []string{"Versailles"}},
			{URL: "https://www.musee-orsay.fr/en", Labels: []string{"Musee d'Orsay", "Musée d'Orsay"}},
			{URL: "https://parisjetaime.com/eng/article/le-marais-a057", Labels: []string{"Le Marais"}},
			{URL: "https://parisjetaime.com/eng/article/ile-de-la-cite-and-ile-saint-louis-a051", Labels: []string{"Ile de la Cite", "Île de la Cité"}},
		}
	default:
		return nil
	}
}

func applySmartLinks(content string, links []smartLink) string {
	enriched := content
	for _, link := range links {
		for _, label := range link.Labels {
			updated, changed := replaceFirstPlainLabel(enriched, label, link.URL)
			if changed {
				enriched = updated
				break
			}
		}
	}
	return enriched
}

func replaceFirstPlainLabel(content, label, target string) (string, bool) {
	if strings.TrimSpace(label) == "" || strings.TrimSpace(target) == "" {
		return content, false
	}
	lowerContent := strings.ToLower(content)
	lowerLabel := strings.ToLower(label)
	searchFrom := 0
	for {
		index := strings.Index(lowerContent[searchFrom:], lowerLabel)
		if index == -1 {
			return content, false
		}
		start := searchFrom + index
		end := start + len(label)
		if isPlainLabelMatch(content, start, end) {
			matched := content[start:end]
			replacement := "[" + matched + "](" + target + ")"
			return content[:start] + replacement + content[end:], true
		}
		searchFrom = end
	}
}

func isPlainLabelMatch(content string, start, end int) bool {
	if start < 0 || end > len(content) || start >= end {
		return false
	}
	if start > 0 {
		prev, _ := utf8.DecodeLastRuneInString(content[:start])
		if prev == '[' || prev == '`' || prev == '/' || isWordRune(prev) {
			return false
		}
	}
	if end < len(content) {
		next, _ := utf8.DecodeRuneInString(content[end:])
		if next == '`' || isWordRune(next) {
			return false
		}
		if next == ']' && strings.HasPrefix(content[end:], "](") {
			return false
		}
	}
	if strings.Contains(content[windowStart(start, 8):windowEnd(len(content), end, 8)], "http") {
		return false
	}
	return true
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '-'
}

func windowStart(index, pad int) int {
	start := index - pad
	if start > 0 {
		return start
	}
	return 0
}

func windowEnd(length, index, pad int) int {
	end := index + pad
	if end < length {
		return end
	}
	return length
}
