package app

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func maybeHandleSimpleAnswer(raw string, stdout io.Writer) bool {
	if answer, ok := simpleArithmeticAnswer(raw); ok {
		fmt.Fprintln(stdout, answer)
		return true
	}
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

func maybeHandleAmbiguousBareEntity(raw string, stdout io.Writer) bool {
	subject, ok := ambiguousBareEntitySubject(raw)
	if !ok {
		return false
	}
	fmt.Fprintf(stdout, "What would you like me to do with %s?\n", subject)
	return true
}

var simpleArithmeticPattern = regexp.MustCompile(`^([+-]?(?:\d+(?:\.\d*)?|\.\d+))\s*([+\-*/x])\s*([+-]?(?:\d+(?:\.\d*)?|\.\d+))$`)

func simpleArithmeticAnswer(raw string) (string, bool) {
	expression := simpleArithmeticExpression(raw)
	matches := simpleArithmeticPattern.FindStringSubmatch(expression)
	if len(matches) != 4 {
		return "", false
	}

	left, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", false
	}
	right, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return "", false
	}

	var result float64
	switch matches[2] {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*", "x":
		result = left * right
	case "/":
		if right == 0 {
			return "", false
		}
		result = left / right
	default:
		return "", false
	}
	return strconv.FormatFloat(result, 'f', -1, 64) + ".", true
}

func simpleArithmeticExpression(raw string) string {
	expression := strings.ToLower(strings.TrimSpace(raw))
	expression = strings.TrimSpace(strings.TrimSuffix(expression, "?"))
	expression = strings.ReplaceAll(expression, "×", "x")
	expression = strings.ReplaceAll(expression, "÷", "/")
	for _, prefix := range []string{
		"what is ",
		"what's ",
		"whats ",
		"calculate ",
		"compute ",
		"solve ",
	} {
		if strings.HasPrefix(expression, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(expression, prefix))
		}
	}
	return expression
}

func simpleCapitalAnswer(raw string) (string, bool) {
	country := simpleCapitalCountry(raw)
	if country == "" {
		return "", false
	}
	if correction, ok := capitalQuestionCorrection(country); ok {
		return correction, true
	}
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

func capitalQuestionCorrection(subject string) (string, bool) {
	switch subject {
	case "paris":
		return "Paris is a city, not a country. Paris is the capital of France.", true
	default:
		return "", false
	}
}

func simpleCapitalCountry(raw string) string {
	normalized := normalizeSimpleQuestion(raw)
	for _, prefix := range []string{
		"what is the capital city of ",
		"whats the capital city of ",
		"what is the capital of ",
		"whats the capital of ",
		"what is capital of ",
		"whats capital of ",
		"capital city of ",
		"capital of ",
	} {
		if strings.HasPrefix(normalized, prefix) {
			country := strings.TrimSpace(strings.TrimPrefix(normalized, prefix))
			return strings.TrimSpace(strings.TrimPrefix(country, "the "))
		}
	}
	return ""
}

func normalizeSimpleQuestion(raw string) string {
	normalized := normalizeName(raw)
	replacements := map[string]string{
		"teh":     "the",
		"capitol": "capital",
	}
	words := strings.Fields(normalized)
	for index, word := range words {
		if replacement, ok := replacements[word]; ok {
			words[index] = replacement
		}
	}
	return strings.Join(words, " ")
}

func ambiguousBareEntitySubject(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	normalized := normalizeName(trimmed)
	if normalized == "" || looksLikeStandaloneQuestion(raw) {
		return "", false
	}
	if strings.ContainsAny(trimmed, "\n\r:;?!") {
		return "", false
	}
	words := strings.Fields(normalized)
	if len(words) == 0 || len(words) > 3 {
		return "", false
	}
	if hasStarterIntentSignal(normalized) {
		return "", false
	}
	return trimmed, true
}

func hasStarterIntentSignal(normalized string) bool {
	signals := []string{
		"add", "answer", "book", "build", "change", "choose", "clean", "code",
		"commit", "compare", "configure", "convert", "create", "debug", "delete",
		"deploy", "design", "draft", "edit", "email", "explain", "find", "fix",
		"flight", "follow up", "followup", "generate", "hotel", "implement",
		"incident", "install", "itinerary", "make", "meeting", "memo", "open",
		"plan", "push", "read", "recommend", "refactor", "remove", "research",
		"review", "run", "send", "show", "summarize", "test", "travel", "trip",
		"update", "vendor", "write",
	}
	padded := " " + normalized + " "
	for _, signal := range signals {
		if strings.Contains(padded, " "+signal+" ") {
			return true
		}
	}
	return false
}

func looksLikeStandaloneQuestion(raw string) bool {
	normalized := normalizeName(raw)
	if normalized == "" || looksLikeCurrentWorkQuestion(normalized) {
		return false
	}
	for _, prefix := range []string{
		"what is ",
		"whats ",
		"who is ",
		"whos ",
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
