package app

import "testing"

// Coding tasks that merely mention email/travel words must route to the agent,
// not get grabbed by a demo template (critique #4 routing-accuracy fix).
func TestDemoInterceptsSkipCodingTasks(t *testing.T) {
	cases := []struct {
		name          string
		prompt        string
		email, travel bool
	}{
		// Coding tasks with incidental email/travel words → must NOT intercept.
		{"code-email-fn", "write a Go function to send a follow-up email via the API", false, false},
		{"code-email-method", "implement a method that composes a summary email from standup notes", false, false},
		{"code-travel-refactor", "plan a refactor of the itinerary service module", false, false},
		{"code-ext", "draft a summary email module in notify.go for the standup recap", false, false},
		// Genuine non-coding asks → still intercept (compact direct path preserved).
		{"genuine-email", "write a follow-up email summarizing today's standup: shipped X, blocked on Y", true, false},
		{"genuine-email-api-mention", "draft a follow-up email summarizing the API changes we shipped", true, false},
		{"genuine-travel", "plan a 3 day trip to Kyoto, first time visitor", false, true},
	}
	for _, c := range cases {
		if got := isDirectFollowupEmailPrompt(c.prompt); got != c.email {
			t.Errorf("email intercept %q = %v, want %v", c.prompt, got, c.email)
		}
		if got := isDirectTravelPlanPrompt(c.prompt); got != c.travel {
			t.Errorf("travel intercept %q = %v, want %v", c.prompt, got, c.travel)
		}
	}
}

// The shared detector must fire on code deliverables and stay quiet on prose.
func TestPromptLooksLikeCodingTask(t *testing.T) {
	coding := []string{
		"write a Go function", "implement a method", "refactor the parser",
		"add a handler for /health", "fix the failing unit test", "update schema.sql",
		"write a python script", "create a react component",
	}
	notCoding := []string{
		"write a follow-up email about the standup", "plan a trip to Kyoto",
		"summarize the API changes for the newsletter", "what is a database index",
	}
	for _, p := range coding {
		if !promptLooksLikeCodingTask(p) {
			t.Errorf("should detect coding: %q", p)
		}
	}
	for _, p := range notCoding {
		if promptLooksLikeCodingTask(p) {
			t.Errorf("should NOT detect coding: %q", p)
		}
	}
}
