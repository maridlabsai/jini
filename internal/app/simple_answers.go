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
