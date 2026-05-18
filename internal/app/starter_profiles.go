package app

import "strings"

type starterPackProfile struct {
	PackID        string
	ChoiceLabel   string
	DefaultName   string
	State         string
	MenuAliases   []string
	DetectSignals []string
	Clarification *starterClarificationProfile
	Writer        func(workDir, title, source, detail string) error
}

type starterClarificationProfile struct {
	Intro          string
	SkipHint       string
	Example        string
	MinimumMissing int
	Dimensions     []starterScopeDimension
}

type starterScopeDimension struct {
	Label             string
	NormalizedSignals []string
	RawSignals        []string
}

var starterPackProfiles = map[string]starterPackProfile{
	"meeting-followup": {
		PackID:        "meeting-followup",
		ChoiceLabel:   "Turn meeting notes into something I can send",
		DefaultName:   "Meeting Follow-up",
		State:         "decided",
		MenuAliases:   []string{"1", "meeting", "turn meeting notes into something i can send", "follow up after a meeting", "follow up", "meeting follow up", "meeting followup"},
		DetectSignals: []string{"meeting", "follow up", "followup", "action items", "owners", "due dates", "open questions"},
		Writer:        writeMeetingStarterWork,
	},
	"research-prd": {
		PackID:        "research-prd",
		ChoiceLabel:   "Check whether a plan is ready to hand off",
		DefaultName:   "Plan Readiness",
		State:         "awaiting_verification",
		MenuAliases:   []string{"2", "plan", "check whether a plan is ready to hand off", "check if a plan is ready", "spec", "spec readiness"},
		DetectSignals: []string{"prd", "spec", "build readiness", "ready to hand off", "handoff", "hand off", "rollback", "implementation slice"},
		Writer:        writeResearchStarterWork,
	},
	"vendor-selection": {
		PackID:        "vendor-selection",
		ChoiceLabel:   "Compare options and choose one",
		DefaultName:   "Option Review",
		State:         "decided",
		MenuAliases:   []string{"compare options", "compare options and choose one", "vendor"},
		DetectSignals: []string{"vendor", "compare options", "choose one", "recommendation memo"},
		Writer: func(workDir, title, source, detail string) error {
			return writeSimpleStarterWork(workDir, title, "Recommendation Memo", source, []string{
				"Top option",
				"Tradeoffs still to review",
				"Budget or approval boundary",
			})
		},
	},
	"incident-response": {
		PackID:        "incident-response",
		ChoiceLabel:   "Clean up an incident",
		DefaultName:   "Incident Cleanup",
		State:         "incident",
		MenuAliases:   []string{"4", "incident", "clean up an incident"},
		DetectSignals: []string{"incident", "outage", "customer impact", "root cause", "recovery"},
		Writer: func(workDir, title, source, detail string) error {
			return writeSimpleStarterWork(workDir, title, "Closure Checklist", source, []string{
				"Recovery proof",
				"Open follow-up owners",
				"Customer or leadership update status",
			})
		},
	},
	"travel-plan": {
		PackID:        "travel-plan",
		ChoiceLabel:   "Plan a trip",
		DefaultName:   "Trip Plan",
		State:         "decided",
		MenuAliases:   []string{"5", "trip", "plan a trip"},
		DetectSignals: []string{"trip", "travel", "paris", "hotel", "flight", "itinerary"},
		Clarification: &starterClarificationProfile{
			Intro:          "Before I draft it, give me what is still missing in one line:",
			SkipHint:       "Type `skip` if you want a generic draft.",
			Example:        "Example: early October, mixed pace, central hotel area, Louvre and Versailles are must-dos",
			MinimumMissing: 2,
			Dimensions: []starterScopeDimension{
				{
					Label:             "who is going",
					NormalizedSignals: []string{"solo", "couple", "friends", "family", "kids", "children", "parents", "honeymoon", "wife", "husband", "partner"},
				},
				{
					Label:             "rough budget",
					NormalizedSignals: []string{"cheap", "luxury", "midrange", "2500", "3000", "2000", "1500", "4000"},
					RawSignals:        []string{"$", "budget"},
				},
				{
					Label:             "dates or season",
					NormalizedSignals: []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december", "spring", "summer", "fall", "autumn", "winter", "weekend", "weekday", "christmas", "new year"},
				},
				{
					Label:             "trip style",
					NormalizedSignals: []string{"food", "museum", "romantic", "nightlife", "shopping", "family friendly", "mixed", "slow pace", "fast pace", "walking", "architecture", "relaxed", "packed", "kid friendly", "honeymoon", "adventure"},
				},
				{
					Label:             "hotel area, or whether you want help choosing it",
					NormalizedSignals: []string{"hotel", "stay", "marais", "latin quarter", "montmartre", "central", "area", "neighborhood", "neighbourhood", "arrondissement", "left bank", "right bank"},
				},
				{
					Label:             "must-do sights, or whether you want help choosing them",
					NormalizedSignals: []string{"louvre", "versailles", "eiffel", "orsay", "montmartre", "notre dame", "latin quarter", "marais", "disneyland", "seine cruise", "must do", "must see"},
				},
			},
		},
		Writer: writeTravelStarterWork,
	},
	"general-work": {
		PackID:        "general-work",
		ChoiceLabel:   "Something else",
		DefaultName:   "General Work",
		State:         "decided",
		MenuAliases:   []string{"6", "something else", "something"},
		DetectSignals: nil,
		Writer: func(workDir, title, source, detail string) error {
			return writeFirstUsefulPassStarterWork(workDir, title, source)
		},
	},
}

var starterPackDetectionOrder = []string{
	"meeting-followup",
	"research-prd",
	"travel-plan",
	"vendor-selection",
	"incident-response",
}

var starterPackMenuOrder = []string{
	"meeting-followup",
	"research-prd",
	"vendor-selection",
	"incident-response",
	"travel-plan",
	"general-work",
}

func starterProfileForPack(packID string) (starterPackProfile, bool) {
	profile, ok := starterPackProfiles[packID]
	return profile, ok
}

func starterChoiceForPack(packID string) (starterChoice, bool) {
	profile, ok := starterProfileForPack(packID)
	if !ok {
		return starterChoice{}, false
	}
	return starterChoice{
		PackID:      profile.PackID,
		ChoiceLabel: profile.ChoiceLabel,
		DefaultName: profile.DefaultName,
		State:       profile.State,
	}, true
}

func detectStarterPackFromSource(source string) string {
	normalized := normalizeName(source)
	for _, packID := range starterPackDetectionOrder {
		profile, ok := starterProfileForPack(packID)
		if !ok || len(profile.DetectSignals) == 0 {
			continue
		}
		if containsAny(normalized, profile.DetectSignals) {
			return packID
		}
	}
	return "general-work"
}

func clarificationPromptForProfile(profile starterPackProfile, source string) (string, bool) {
	if profile.Clarification == nil {
		return "", false
	}
	missing := missingScopeDimensions(source, profile.Clarification.Dimensions)
	if len(missing) < max(1, profile.Clarification.MinimumMissing) {
		return "", false
	}
	lines := []string{profile.Clarification.Intro}
	for _, item := range missing {
		lines = append(lines, "- "+item)
	}
	if strings.TrimSpace(profile.Clarification.SkipHint) != "" {
		lines = append(lines, profile.Clarification.SkipHint)
	}
	if strings.TrimSpace(profile.Clarification.Example) != "" {
		lines = append(lines, profile.Clarification.Example)
	}
	return strings.Join(lines, "\n"), true
}

func missingScopeDimensions(source string, dimensions []starterScopeDimension) []string {
	normalized := normalizeName(source)
	rawLower := strings.ToLower(strings.TrimSpace(source))
	missing := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		if containsAny(normalized, dimension.NormalizedSignals) {
			continue
		}
		if containsAny(rawLower, dimension.RawSignals) {
			continue
		}
		missing = append(missing, dimension.Label)
	}
	return missing
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
