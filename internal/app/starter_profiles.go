package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type starterPackProfile struct {
	PackID        string
	ChoiceLabel   string
	DefaultName   string
	State         string
	WorkClass     string
	RequestCohort string
	ArtifactFamily string
	MenuAliases     []string
	DetectSignals   []string
	PrimaryViewPath  string
	PrimaryViewLabel string
	WorkingWith      string
	Done             []string
	NextStep         string
	PrioritizedViewIDs []string
	SynthesizedViews []starterCatalogSpec
	TasksView        *starterTasksViewProfile
	DoingByState     map[string]string
	SmartLinks       []smartLink
	ProviderArtifactGuidance string
	DraftQualityProfile      draftQualityProfile
	MissingBuilder   func(state string, details []catalogItem, source string) []string
	UncertainBuilder func(source string, missing []string) []string
	AskBuilder       func(summary *workSummary, source string) *threadAsk
	Writer        func(workDir, title, source, detail string) error
}

type scopePlannerProfile struct {
	RequestCohorts []string
	Intro          string
	SkipHint       string
	Example        string
	MinimumMissing int
	Dimensions     []scopePlannerDimension
}

type scopePlannerDimension struct {
	Label             string
	NormalizedSignals []string
	RawSignals        []string
}

type starterCatalogSpec struct {
	ID       string
	FileStem string
	Label    string
	Aliases  []string
}

type starterTasksViewProfile struct {
	ID                   string
	Label                string
	Aliases              []string
	CompanionFileStem    string
	CompanionPresentView *starterCatalogSpec
}

var starterPackProfiles = map[string]starterPackProfile{
	"meeting-followup": {
		PackID:        "meeting-followup",
		ChoiceLabel:   "Turn meeting notes into something I can send",
		DefaultName:   "Meeting Follow-up",
		State:         "decided",
		WorkClass:     "planning",
		RequestCohort: "sendable-followup",
		ArtifactFamily:"narrative-draft",
		MenuAliases:   []string{"1", "meeting", "turn meeting notes into something i can send", "follow up after a meeting", "follow up", "meeting follow up", "meeting followup"},
		DetectSignals: []string{"meeting", "follow up", "followup", "action items", "owners", "due dates", "open questions"},
		PrimaryViewPath:  "followup.md",
		PrimaryViewLabel: "Sendable Follow-Up",
		WorkingWith:      "Meeting notes and follow-up tasks",
		Done:             []string{"Sendable follow-up drafted", "Owners and due points pulled out"},
		NextStep:         "Open Sendable Follow-up",
		PrioritizedViewIDs: []string{"sendable-follow-up", "owners-and-due-points"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "sendable-follow-up", FileStem: "followup", Label: "Sendable Follow-up", Aliases: []string{"follow-up", "followup", "summary"}},
			{ID: "owners-and-due-points", FileStem: "owners-and-due-points", Label: "Owners and Due Points", Aliases: []string{"owners", "due points", "owners and due points"}},
		},
		TasksView: &starterTasksViewProfile{
			ID:                "owners-and-due-points",
			Label:             "Owners and Due Points",
			Aliases:           []string{"tasks", "task list", "owners"},
			CompanionFileStem: "owners-and-due-points",
			CompanionPresentView: &starterCatalogSpec{
				ID:      "task-list",
				FileStem: "tasks",
				Label:   "Task List",
				Aliases: []string{"tasks", "task list"},
			},
		},
		DoingByState: map[string]string{
			"decided": "Turning notes into owners and next steps",
		},
		ProviderArtifactGuidance: strings.Join([]string{
			"Return a sendable follow-up note first.",
			"Use these sections exactly:",
			"- `## Send this note`",
			"- `## Decisions captured from the notes`",
			"- `## Owners and due dates to confirm`",
			"- `## Open questions to close`",
			"- `## Recommended next move`",
			"Do not invent names, dates, or commitments that are not grounded in the source.",
		}, "\n"),
		DraftQualityProfile: draftQualityProfile{
			RequiredHeadings:       []string{"## Send this note", "## Decisions captured from the notes", "## Owners and due dates to confirm", "## Open questions to close", "## Recommended next move"},
			PreferredHeadings:      []string{"## Still to confirm"},
			RequiredHeadingWeight:  6,
			PreferredHeadingWeight: 3,
			UncertaintyWeight:      4,
		},
		MissingBuilder: func(state string, details []catalogItem, source string) []string {
			if state == "decided" || state == "in_make" {
				return []string{"Metric and legal-review decision"}
			}
			return nil
		},
		UncertainBuilder: func(source string, missing []string) []string {
			if len(missing) == 0 {
				return nil
			}
			return []string{"Whether the metric decision also needs legal review"}
		},
		AskBuilder: func(summary *workSummary, source string) *threadAsk {
			return &threadAsk{
				AskID:   "confirm-owners-and-dates",
				Prompt:  "Confirm any missing owner or due date before sending this follow-up.",
				Reason:  "The note is usable now, but it becomes truly sendable only when ownership and timing are explicit.",
				Options: []string{"Add missing owner", "Add due date", "Skip for now"},
				AssumptionsIfSkipped: []string{
					"Jini will keep the follow-up in draft form and leave missing owner or date gaps visible.",
				},
				Blocking: true,
			}
		},
		Writer:        writeMeetingStarterWork,
	},
	"research-prd": {
		PackID:        "research-prd",
		ChoiceLabel:   "Check whether a plan is ready to hand off",
		DefaultName:   "Plan Readiness",
		State:         "awaiting_verification",
		WorkClass:     "code",
		RequestCohort: "build-readiness",
		ArtifactFamily:"structured-check",
		MenuAliases:   []string{"2", "plan", "check whether a plan is ready to hand off", "check if a plan is ready", "spec", "spec readiness"},
		DetectSignals: []string{"prd", "spec", "build readiness", "ready to hand off", "handoff", "hand off", "rollback", "implementation slice"},
		PrimaryViewPath:  "prd.md",
		PrimaryViewLabel: "Build-Readiness Check",
		WorkingWith:      "Latest PRD draft and review comments",
		Done:             []string{"Build-readiness draft created", "Missing build blockers identified"},
		NextStep:         "Open Build-Readiness Check",
		PrioritizedViewIDs: []string{"build-readiness-check", "handoff-brief", "missing-pieces-before-build"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "build-readiness-check", FileStem: "prd", Label: "Build-Readiness Check", Aliases: []string{"readiness", "build readiness check", "check"}},
			{ID: "handoff-brief", FileStem: "prd", Label: "Handoff Brief", Aliases: []string{"prd", "summary", "brief", "handoff"}},
			{ID: "missing-pieces-before-build", FileStem: "missing-pieces-before-build", Label: "Missing Pieces Before Build", Aliases: []string{"missing", "before build", "missing pieces"}},
		},
		TasksView: &starterTasksViewProfile{
			ID:                "missing-pieces-before-build",
			Label:             "Missing Pieces Before Build",
			Aliases:           []string{"tasks", "task list", "missing"},
			CompanionFileStem: "missing-pieces-before-build",
			CompanionPresentView: &starterCatalogSpec{
				ID:      "task-list",
				FileStem: "tasks",
				Label:   "Task List",
				Aliases: []string{"tasks", "task list"},
			},
		},
		ProviderArtifactGuidance: strings.Join([]string{
			"Return a build-readiness artifact, not a vague summary.",
			"Use these sections exactly:",
			"- `## What looks ready now`",
			"- `## Must clear before build`",
			"- `## Recommended first slice`",
			"- `## Who needs to answer what`",
			"- `## Still to confirm`",
			"Do not reduce the answer to a binary verdict. Keep missing proof and approval gaps visible.",
		}, "\n"),
		DraftQualityProfile: draftQualityProfile{
			RequiredHeadings:       []string{"## What looks ready now", "## Must clear before build", "## Recommended first slice", "## Who needs to answer what", "## Still to confirm"},
			PreferredHeadings:      []string{"## Risks", "## Approval gaps"},
			RequiredHeadingWeight:  6,
			PreferredHeadingWeight: 3,
			UncertaintyWeight:      4,
		},
		UncertainBuilder: func(source string, missing []string) []string {
			if len(missing) == 0 {
				return nil
			}
			return []string{"Whether approval was already granted in the review thread"}
		},
		AskBuilder: func(summary *workSummary, source string) *threadAsk {
			return &threadAsk{
				AskID:   "confirm-approval-and-first-slice",
				Prompt:  "Name the approval owner and confirm the first implementation slice.",
				Reason:  "The readiness check is useful now, but build should not start until approval and the first slice are explicit.",
				Options: []string{"Set approval owner", "Set first slice", "Skip for now"},
				AssumptionsIfSkipped: []string{
					"Jini will keep approval and first-slice gaps visible instead of treating the plan as build-ready.",
				},
				Blocking: true,
			}
		},
		Writer:        writeResearchStarterWork,
	},
	"vendor-selection": {
		PackID:        "vendor-selection",
		ChoiceLabel:   "Compare options and choose one",
		DefaultName:   "Option Review",
		State:         "decided",
		WorkClass:     "planning",
		RequestCohort: "option-compare",
		ArtifactFamily:"comparison-matrix",
		MenuAliases:   []string{"compare options", "compare options and choose one", "vendor"},
		DetectSignals: []string{"vendor", "compare options", "choose one", "recommendation memo"},
		PrimaryViewPath:  "recommendation-memo.md",
		PrimaryViewLabel: "Recommendation Memo",
		WorkingWith:      "Vendor notes, tradeoffs, and decision criteria",
		Done:             []string{"Recommendation memo drafted", "Tradeoffs laid out"},
		NextStep:         "Open Recommendation Memo",
		PrioritizedViewIDs: []string{"recommendation-memo"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "recommendation-memo", FileStem: "selection", Label: "Recommendation Memo", Aliases: []string{"selection", "memo"}},
		},
		ProviderArtifactGuidance: strings.Join([]string{
			"Return a recommendation artifact, not a generic comparison summary.",
			"Use these sections exactly:",
			"- `## Recommendation`",
			"- `## Tradeoffs`",
			"- `## Risks`",
			"- `## Next move`",
			"- `## Still to confirm`",
		}, "\n"),
		DraftQualityProfile: draftQualityProfile{
			RequiredHeadings:       []string{"## Recommendation", "## Risks", "## Still to confirm"},
			PreferredHeadings:      []string{"## Tradeoffs", "## Next move"},
			RequiredHeadingWeight:  7,
			PreferredHeadingWeight: 3,
			UncertaintyWeight:      4,
		},
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
		WorkClass:     "code",
		RequestCohort: "incident-cleanup",
		ArtifactFamily:"step-plan",
		MenuAliases:   []string{"4", "incident", "clean up an incident"},
		DetectSignals: []string{"incident", "outage", "customer impact", "root cause", "recovery"},
		PrimaryViewPath:  "closure-checklist.md",
		PrimaryViewLabel: "Closure Checklist",
		WorkingWith:      "Incident notes, timeline, and follow-up tasks",
		Done:             []string{"Closure checklist drafted", "Recovery follow-ups pulled out"},
		NextStep:         "Open Closure Checklist",
		PrioritizedViewIDs: []string{"closure-checklist"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "closure-checklist", FileStem: "response", Label: "Closure Checklist", Aliases: []string{"response", "checklist"}},
		},
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
		WorkClass:     "planning",
		RequestCohort: "trip-itinerary",
		ArtifactFamily:"itinerary-plan",
		MenuAliases:   []string{"5", "trip", "plan a trip"},
		DetectSignals: []string{"trip", "travel", "paris", "hotel", "flight", "itinerary"},
		PrimaryViewPath:  "itinerary.md",
		PrimaryViewLabel: "Itinerary",
		WorkingWith:      "Trip notes, dates, and planning details",
		Done:             []string{"Itinerary drafted", "Budget sketch created"},
		NextStep:         "Open Itinerary",
		PrioritizedViewIDs: []string{"itinerary", "budget-sketch", "travel-logistics", "still-to-book"},
		TasksView: &starterTasksViewProfile{
			ID:      "still-to-book",
			Label:   "Still To Book",
			Aliases: []string{"tasks", "task list", "booking"},
		},
		SmartLinks: []smartLink{
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
		},
		ProviderArtifactGuidance: strings.Join([]string{
			"Return a trip-planning artifact, not a generic travel essay.",
			"Use these sections exactly:",
			"- `## Day by day`",
			"- `## Budget`",
			"- `## Travel logistics`",
			"- `## Still to book`",
			"- `## Still to confirm`",
			"When a key destination, museum, or landmark is clearly part of the plan, add one smart Markdown link on first mention.",
			"Prefer canonical destination links and avoid turning every bullet into a link list.",
		}, "\n"),
		DraftQualityProfile: draftQualityProfile{
			RequiredHeadings:       []string{"## Day by day", "## Still to confirm"},
			PreferredHeadings:      []string{"## Budget", "## Travel logistics", "## Still to book"},
			RequiredHeadingWeight:  7,
			PreferredHeadingWeight: 3,
			UncertaintyWeight:      4,
		},
		MissingBuilder: func(state string, details []catalogItem, source string) []string {
			if state != "decided" && state != "in_make" {
				return nil
			}
			ctx := parseTravelStarterContext(source)
			if len(ctx.Missing) == 0 {
				return nil
			}
			return travelStillToConfirm(ctx)
		},
		UncertainBuilder: func(source string, missing []string) []string {
			if len(missing) == 0 {
				return nil
			}
			ctx := parseTravelStarterContext(source)
			if len(ctx.MustDos) > 0 {
				return []string{fmt.Sprintf("Which of %s should be time-locked first", strings.ToLower(ctx.MustDos[0]))}
			}
			return []string{"Which one or two anchor experiences should be locked first"}
		},
		AskBuilder: func(summary *workSummary, source string) *threadAsk {
			if len(summary.Missing) == 0 {
				return nil
			}
			options := []string{}
			for _, item := range summary.Missing {
				options = append(options, "Add "+strings.ToLower(item))
				if len(options) >= 3 {
					break
				}
			}
			options = append(options, "Skip for now")
			return &threadAsk{
				AskID:   "confirm-trip-basics",
				Prompt:  "Confirm the highest-impact trip details before booking from this draft.",
				Reason:  "These details materially change the itinerary, booking order, and cost guidance.",
				Options: options,
				AssumptionsIfSkipped: []string{
					"Jini will keep the itinerary as a draft and leave booking decisions visibly open.",
				},
				Blocking: true,
			}
		},
		Writer: writeTravelStarterWork,
	},
	"general-work": {
		PackID:        "general-work",
		ChoiceLabel:   "Something else",
		DefaultName:   "Working Draft",
		State:         "decided",
		WorkClass:     "general",
		RequestCohort: "general-pass",
		ArtifactFamily:"general-pass",
		MenuAliases:   []string{"6", "something else", "something"},
		DetectSignals: nil,
		PrimaryViewPath:  "first-useful-pass.md",
		PrimaryViewLabel: "Working Draft",
		WorkingWith:      "The files and notes in this work",
		NextStep:         "Review what is ready",
		PrioritizedViewIDs: []string{"first-useful-pass", "next-actions"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "first-useful-pass", FileStem: "first-useful-pass", Label: "Working Draft", Aliases: []string{"working draft", "first pass", "useful pass", "summary", "draft"}},
		},
		TasksView: &starterTasksViewProfile{
			ID:      "next-actions",
			Label:   "Next Actions",
			Aliases: []string{"tasks", "task list", "actions"},
		},
		Writer: func(workDir, title, source, detail string) error {
			return writeFirstUsefulPassStarterWork(workDir, title, source)
		},
	},
}

var starterScopePlannerProfiles = []scopePlannerProfile{
	{
		RequestCohorts: []string{"trip-itinerary"},
		Intro:          "Before I draft it, help me narrow the scope in one line:",
		SkipHint:       "Type `skip` if you want a generic first draft.",
		Example:        "Example: early October, mixed pace, central hotel area, one museum and one day trip are must-dos",
		MinimumMissing: 2,
		Dimensions:     travelScopeDimensions,
	},
}

var travelScopeDimensions = []scopePlannerDimension{
	{
		Label:             "travelers",
		NormalizedSignals: []string{"solo", "couple", "friends", "family", "kids", "children", "parents", "honeymoon", "wife", "husband", "partner"},
	},
	{
		Label:             "budget range",
		NormalizedSignals: []string{"cheap", "luxury", "midrange", "2500", "3000", "2000", "1500", "4000"},
		RawSignals:        []string{"$", "budget"},
	},
	{
		Label:             "dates or season",
		NormalizedSignals: []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december", "spring", "summer", "fall", "autumn", "winter", "weekend", "weekday", "christmas", "new year"},
	},
	{
		Label:             "pace or style",
		NormalizedSignals: []string{"food", "museum", "romantic", "nightlife", "shopping", "family friendly", "mixed", "slow pace", "fast pace", "walking", "architecture", "relaxed", "packed", "kid friendly", "honeymoon", "adventure"},
	},
	{
		Label:             "base area, or whether you want help choosing one",
		NormalizedSignals: []string{"hotel", "stay", "marais", "latin quarter", "montmartre", "central", "area", "neighborhood", "neighbourhood", "arrondissement", "left bank", "right bank"},
	},
	{
		Label:             "must-do anchors, or whether you want help choosing them",
		NormalizedSignals: []string{"louvre", "versailles", "eiffel", "orsay", "montmartre", "notre dame", "latin quarter", "marais", "disneyland", "seine cruise", "must do", "must see"},
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

func starterProfile(packID string) starterPackProfile {
	if profile, ok := starterProfileForPack(packID); ok {
		return profile
	}
	return starterPackProfiles["general-work"]
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

func scopePlannerForProfile(profile starterPackProfile) (scopePlannerProfile, bool) {
	for _, planner := range starterScopePlannerProfiles {
		for _, cohort := range planner.RequestCohorts {
			if cohort == profile.RequestCohort {
				return planner, true
			}
		}
	}
	return scopePlannerProfile{}, false
}

func clarificationPromptForProfile(profile starterPackProfile, source string) (string, bool) {
	planner, ok := scopePlannerForProfile(profile)
	if !ok {
		return "", false
	}
	missing := missingScopeDimensions(source, planner.Dimensions)
	if len(missing) < maxInt(1, planner.MinimumMissing) {
		return "", false
	}
	lines := []string{planner.Intro}
	for _, item := range missing {
		lines = append(lines, "- "+item)
	}
	if strings.TrimSpace(planner.SkipHint) != "" {
		lines = append(lines, planner.SkipHint)
	}
	if strings.TrimSpace(planner.Example) != "" {
		lines = append(lines, planner.Example)
	}
	return strings.Join(lines, "\n"), true
}

func missingScopeDimensions(source string, dimensions []scopePlannerDimension) []string {
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func starterPrimaryView(packID string) (string, string) {
	profile := starterProfile(packID)
	return firstNonEmpty(profile.PrimaryViewPath, "first-useful-pass.md"), firstNonEmpty(profile.PrimaryViewLabel, "Working Draft")
}

func starterWorkingWith(packID string) string {
	return firstNonEmpty(strings.TrimSpace(starterProfile(packID).WorkingWith), "The files and notes in this work")
}

func starterDone(packID string, views []catalogItem) []string {
	if done := starterProfile(packID).Done; len(done) > 0 {
		return append([]string{}, done...)
	}
	if len(views) > 0 {
		return []string{views[0].Label + " drafted"}
	}
	return []string{"First useful pass created"}
}

func starterNextStep(packID string, views []catalogItem) string {
	if next := strings.TrimSpace(starterProfile(packID).NextStep); next != "" {
		return next
	}
	if len(views) > 0 {
		return "Open " + views[0].Label
	}
	return "Review what is ready"
}

func starterDoing(packID, state string) string {
	profile := starterProfile(packID)
	if override := strings.TrimSpace(profile.DoingByState[state]); override != "" {
		return override
	}
	switch state {
	case "awaiting_verification":
		return "Checking assumptions and approval gaps"
	case "decided":
		return "Turning decisions into a usable draft"
	case "in_make":
		return "Drafting the next usable version"
	case "operational":
		return "Keeping the work current and verified"
	case "incident":
		return "Checking recovery work and missing proof"
	default:
		return "Turning the work into something usable"
	}
}

func starterMissing(packID, state string, details []catalogItem, source string) []string {
	if builder := starterProfile(packID).MissingBuilder; builder != nil {
		return builder(state, details, source)
	}
	return nil
}

func starterUncertain(packID string, missing []string, source string) []string {
	if len(missing) == 0 {
		return nil
	}
	if builder := starterProfile(packID).UncertainBuilder; builder != nil {
		return builder(source, missing)
	}
	return []string{"Whether the missing items already exist outside this work record"}
}

func starterAsk(summary *workSummary, source string) *threadAsk {
	if builder := starterProfile(summary.PackID).AskBuilder; builder != nil {
		return builder(summary, source)
	}
	if len(summary.Missing) == 0 {
		return nil
	}
	return &threadAsk{
		AskID:   "confirm-blocking-detail",
		Prompt:  inferNeed(summary.Missing),
		Reason:  "This is the highest-impact missing detail before Jini can strengthen the next draft.",
		Options: []string{"Answer now", "Skip for now"},
		AssumptionsIfSkipped: []string{
			"Jini will keep the missing detail visible and avoid pretending the work is complete.",
		},
		Blocking: true,
	}
}

func starterPrioritizeViews(packID string, items []catalogItem) []catalogItem {
	order := starterProfile(packID).PrioritizedViewIDs
	if len(order) == 0 || len(items) == 0 {
		return items
	}
	rank := map[string]int{}
	for i, id := range order {
		rank[id] = i
	}
	sort.SliceStable(items, func(i, j int) bool {
		left, leftOK := rank[items[i].ID]
		right, rightOK := rank[items[j].ID]
		switch {
		case leftOK && rightOK:
			return left < right
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return items[i].Label < items[j].Label
		}
	})
	return items
}

func starterSynthesizedViews(root, packID string) []catalogItem {
	profile := starterProfile(packID)
	items := []catalogItem{}
	for _, spec := range profile.SynthesizedViews {
		if spec.FileStem == "" {
			continue
		}
		path := filepath.Join(root, "views", spec.FileStem+".md")
		if !fileExists(path) {
			continue
		}
		items = append(items, catalogItem{
			ID:      spec.ID,
			Label:   spec.Label,
			Path:    path,
			Aliases: append([]string{}, spec.Aliases...),
		})
	}
	return items
}

func starterTasksView(packID, dir, path string) (catalogItem, bool) {
	profile := starterProfile(packID)
	if profile.TasksView == nil {
		return catalogItem{}, false
	}
	taskView := profile.TasksView
	if companion := taskView.CompanionPresentView; companion != nil && taskView.CompanionFileStem != "" {
		if fileExists(filepath.Join(dir, taskView.CompanionFileStem+".md")) {
			return catalogItem{
				ID:      companion.ID,
				Label:   companion.Label,
				Path:    path,
				Aliases: append([]string{}, companion.Aliases...),
			}, true
		}
	}
	return catalogItem{
		ID:      taskView.ID,
		Label:   taskView.Label,
		Path:    path,
		Aliases: append([]string{}, taskView.Aliases...),
	}, true
}

func starterViewForStem(packID, stem, path string) (catalogItem, bool) {
	if stem == "tasks" {
		return starterTasksView(packID, filepath.Dir(path), path)
	}
	for _, spec := range starterProfile(packID).SynthesizedViews {
		if spec.FileStem != stem {
			continue
		}
		return catalogItem{
			ID:      spec.ID,
			Label:   spec.Label,
			Path:    path,
			Aliases: append([]string{}, spec.Aliases...),
		}, true
	}
	return catalogItem{}, false
}

func starterRequestCohort(packID string) string {
	if packID == "" || packID == "general-work" {
		return ""
	}
	profile, ok := starterProfileForPack(packID)
	if !ok {
		return ""
	}
	return strings.TrimSpace(profile.RequestCohort)
}

func starterArtifactFamily(packID string) string {
	if packID == "" || packID == "general-work" {
		return ""
	}
	profile, ok := starterProfileForPack(packID)
	if !ok {
		return ""
	}
	return strings.TrimSpace(profile.ArtifactFamily)
}

func starterWorkClass(packID string) string {
	if packID == "" || packID == "general-work" {
		return ""
	}
	profile, ok := starterProfileForPack(packID)
	if !ok {
		return ""
	}
	return strings.TrimSpace(profile.WorkClass)
}

func starterPackSmartLinks(packID string) []smartLink {
	links := starterProfile(packID).SmartLinks
	if len(links) == 0 {
		return nil
	}
	out := make([]smartLink, len(links))
	copy(out, links)
	return out
}

func starterProviderArtifactGuidance(packID string) string {
	return strings.TrimSpace(starterProfile(packID).ProviderArtifactGuidance)
}

func starterDraftQualityProfile(packID string) draftQualityProfile {
	return starterProfile(packID).DraftQualityProfile
}
