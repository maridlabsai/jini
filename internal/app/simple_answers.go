package app

import (
	"fmt"
	"io"
	"strings"
)

func maybeHandleSimpleAnswer(raw string, stdout io.Writer) bool {
	if answer, ok := simpleCapitalAnswer(raw); ok {
		fmt.Fprintln(stdout, answer)
		return true
	}
	if looksLikeStandaloneQuestion(raw) {
		fmt.Fprintln(stdout, "I don't know locally.")
		return true
	}
	return false
}

func simpleCapitalAnswer(raw string) (string, bool) {
	normalized := normalizeName(raw)
	country := ""
	for _, prefix := range []string{"what is the capital of ", "capital of "} {
		if strings.HasPrefix(normalized, prefix) {
			country = strings.TrimSpace(strings.TrimPrefix(normalized, prefix))
			break
		}
	}
	if country == "" {
		return "", false
	}
	country = strings.TrimSuffix(country, "?")
	capitals := map[string]string{
		"france":         "Paris.",
		"germany":        "Berlin.",
		"italy":          "Rome.",
		"japan":          "Tokyo.",
		"spain":          "Madrid.",
		"united kingdom": "London.",
		"uk":             "London.",
		"united states":  "Washington, DC.",
		"usa":            "Washington, DC.",
	}
	answer, ok := capitals[country]
	return answer, ok
}

func looksLikeStandaloneQuestion(raw string) bool {
	normalized := normalizeName(raw)
	if normalized == "" || looksLikeCurrentWorkQuestion(normalized) {
		return false
	}
	for _, prefix := range []string{
		"what is ",
		"who is ",
		"when is ",
		"where is ",
		"why is ",
		"how many ",
		"how much ",
		"define ",
	} {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return strings.HasSuffix(strings.TrimSpace(raw), "?")
}

func looksLikeCurrentWorkQuestion(normalized string) bool {
	if currentWorkQuestionKind(normalized) != "" {
		return true
	}
	switch normalized {
	case "status", "open", "ready", "artifacts", "artifact", "current work":
		return true
	default:
		return false
	}
}
