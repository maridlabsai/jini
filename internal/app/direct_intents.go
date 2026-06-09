package app

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

func maybeHandleDirectPromptIntent(source string, stdout, _ io.Writer) (bool, int) {
	switch interactiveCommandName(source) {
	case "route help":
		renderRouteSetupHelp(stdout)
		return true, 0
	}
	if isRepoReviewDirectTask(source) {
		renderRepoReviewDirect(stdout, buildRepoReviewSnapshot())
		return true, 0
	}
	if isDirectTravelPlanPrompt(source) {
		renderDirectTravelPlan(stdout, source)
		return true, 0
	}
	if isDirectFollowupEmailPrompt(source) {
		renderDirectFollowupEmail(stdout, source)
		return true, 0
	}
	return false, 0
}

func isDirectTravelPlanPrompt(source string) bool {
	normalized := normalizeName(source)
	if !clearTravelDraftPrompt(source) {
		return false
	}
	return containsAny(normalized, []string{"plan a", "plan an", "first time visitor", "itinerary"})
}

func renderDirectTravelPlan(w io.Writer, source string) {
	ctx := parseTravelStarterContext(source)
	destination := firstNonEmpty(ctx.Destination, "the destination")
	dayCount := ctx.DayCount
	if dayCount <= 0 {
		dayCount = 3
	}
	titleDestination := titleCase(destination)
	fmt.Fprintf(w, "%d day %s itinerary\n", dayCount, titleDestination)
	fmt.Fprintln(w)
	for day := 1; day <= dayCount; day++ {
		fmt.Fprintf(w, "Day %d: %s\n", day, directTravelDayLine(day, dayCount, titleDestination, ctx.MustDos))
	}
	missing := travelStillToConfirm(ctx)
	if len(missing) > 0 && !(len(missing) == 1 && normalizeName(missing[0]) == "nothing right now") {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "To refine:")
		for _, item := range firstStrings(missing, 4) {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
}

func directTravelDayLine(day, dayCount int, destination string, mustDos []string) string {
	if day == 1 {
		return fmt.Sprintf("arrive, settle in, and take an easy first walk in %s.", destination)
	}
	if day == dayCount {
		return "keep this as a flexible final day with one favorite stop and departure buffer."
	}
	if len(mustDos) > 0 {
		index := (day - 2) % len(mustDos)
		return fmt.Sprintf("anchor the day around %s, then keep the rest local.", mustDos[index])
	}
	return "pick one neighborhood or museum anchor, then leave room for food and wandering."
}

func isDirectFollowupEmailPrompt(source string) bool {
	normalized := normalizeName(source)
	if !strings.Contains(normalized, "email") {
		return false
	}
	if !containsAny(normalized, []string{"write", "draft", "compose"}) {
		return false
	}
	return containsAny(normalized, []string{"follow up", "followup", "standup", "summary", "summarizing", "recap"})
}

func renderDirectFollowupEmail(w io.Writer, source string) {
	items := directFollowupEmailItems(source)
	fmt.Fprintln(w, "Subject: Standup follow-up")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Hi team,")
	fmt.Fprintln(w)
	if len(items) == 0 {
		fmt.Fprintln(w, "Quick follow-up from today: please reply with any correction to the summary, owners, or timing.")
	} else {
		fmt.Fprintln(w, "Quick follow-up from today:")
		for _, item := range items {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Please reply if I missed anything or if any owner/date needs correction.")
}

func directFollowupEmailItems(source string) []string {
	detail := strings.TrimSpace(source)
	if before, after, ok := strings.Cut(detail, ":"); ok && strings.TrimSpace(after) != "" {
		_ = before
		detail = after
	}
	parts := strings.FieldsFunc(detail, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
	items := []string{}
	for _, part := range parts {
		item := cleanDirectEmailItem(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func cleanDirectEmailItem(raw string) string {
	item := strings.TrimSpace(raw)
	if item == "" {
		return ""
	}
	prefixes := []string{
		"write a follow-up email summarizing",
		"write a follow up email summarizing",
		"draft a follow-up email summarizing",
		"draft a follow up email summarizing",
		"compose a follow-up email summarizing",
		"compose a follow up email summarizing",
		"today standup",
		"standup",
	}
	normalized := normalizeName(item)
	for _, prefix := range prefixes {
		normalizedPrefix := normalizeName(prefix)
		if strings.HasPrefix(normalized, normalizedPrefix+" ") {
			item = strings.TrimSpace(item[len(commonRawPrefix(item, prefix)):])
			normalized = normalizeName(item)
		}
	}
	item = strings.Trim(item, " .")
	if item == "" {
		return ""
	}
	if strings.HasSuffix(strings.ToLower(item), " blocked") && !strings.Contains(strings.ToLower(item), " is blocked") {
		item = strings.TrimSuffix(item, " blocked") + " is blocked"
	}
	item = uppercaseFirstRune(item)
	if !strings.HasSuffix(item, ".") && !strings.HasSuffix(item, "!") && !strings.HasSuffix(item, "?") {
		item += "."
	}
	return item
}

func commonRawPrefix(value, prefix string) string {
	if len(value) < len(prefix) {
		return value
	}
	return value[:len(prefix)]
}

func uppercaseFirstRune(value string) string {
	for index, r := range value {
		return string(unicode.ToUpper(r)) + value[index+len(string(r)):]
	}
	return value
}
