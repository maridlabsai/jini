package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type starterPackProfile struct {
	PackID             string
	ChoiceLabel        string
	DefaultName        string
	State              string
	WorkClass          string
	RequestCohort      string
	ArtifactFamily     string
	MenuAliases        []string
	DetectSignals      []string
	PrimaryViewPath    string
	PrimaryViewLabel   string
	WorkingWith        string
	Done               []string
	NextStep           string
	PrioritizedViewIDs []string
	SynthesizedViews   []starterCatalogSpec
	TasksView          *starterTasksViewProfile
	DoingByState       map[string]string
	SmartLinks         []smartLink
	MissingBuilder     func(state string, details []catalogItem, source string) []string
	UncertainBuilder   func(source string, missing []string) []string
	AskBuilder         func(summary *workSummary, source string) *threadAsk
	Writer             func(workDir, title, source, detail string) error
}

type scopePlannerProfile struct {
	RequestCohorts          []string
	Intro                   string
	SkipHint                string
	Example                 string
	MinimumMissing          int
	Dimensions              []scopePlannerDimension
	PreferDraftSignalGroups [][]string
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
		PackID:             "meeting-followup",
		ChoiceLabel:        "Turn meeting notes into something I can send",
		DefaultName:        "Meeting Follow-up",
		State:              "decided",
		WorkClass:          "planning",
		RequestCohort:      "sendable-followup",
		ArtifactFamily:     "narrative-draft",
		MenuAliases:        []string{"1", "meeting", "turn meeting notes into something i can send", "follow up after a meeting", "follow up", "meeting follow up", "meeting followup"},
		DetectSignals:      []string{"meeting", "follow up", "followup", "action items", "owners", "due dates", "open questions"},
		PrimaryViewPath:    "followup.md",
		PrimaryViewLabel:   "Sendable Follow-Up",
		WorkingWith:        "Meeting notes and follow-up tasks",
		Done:               []string{"Sendable follow-up drafted", "Owners and due points pulled out"},
		NextStep:           "Open Sendable Follow-up",
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
				ID:       "task-list",
				FileStem: "tasks",
				Label:    "Task List",
				Aliases:  []string{"tasks", "task list"},
			},
		},
		DoingByState: map[string]string{
			"decided": "Turning notes into owners and next steps",
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
				Options: []string{"Add missing owner", "Add due date", "Skip"},
				AssumptionsIfSkipped: []string{
					"Jini will keep the follow-up in draft form and leave missing owner or date gaps visible.",
				},
				Blocking: true,
			}
		},
		Writer: writeMeetingStarterWork,
	},
	"research-prd": {
		PackID:             "research-prd",
		ChoiceLabel:        "Check whether a plan is ready to hand off",
		DefaultName:        "Plan Readiness",
		State:              "awaiting_verification",
		WorkClass:          "code",
		RequestCohort:      "build-readiness",
		ArtifactFamily:     "structured-check",
		MenuAliases:        []string{"2", "plan", "check whether a plan is ready to hand off", "check if a plan is ready", "spec", "spec readiness"},
		DetectSignals:      []string{"prd", "spec", "build readiness", "ready to hand off", "handoff", "hand off", "rollback", "implementation slice"},
		PrimaryViewPath:    "prd.md",
		PrimaryViewLabel:   "Build-Readiness Check",
		WorkingWith:        "Latest PRD draft and review comments",
		Done:               []string{"Build-readiness draft created", "Missing build blockers identified"},
		NextStep:           "Open Build-Readiness Check",
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
				ID:       "task-list",
				FileStem: "tasks",
				Label:    "Task List",
				Aliases:  []string{"tasks", "task list"},
			},
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
				Options: []string{"Set approval owner", "Set first slice", "Skip"},
				AssumptionsIfSkipped: []string{
					"Jini will keep approval and first-slice gaps visible instead of treating the plan as build-ready.",
				},
				Blocking: true,
			}
		},
		Writer: writeResearchStarterWork,
	},
	"vendor-selection": {
		PackID:             "vendor-selection",
		ChoiceLabel:        "Compare options and choose one",
		DefaultName:        "Option Review",
		State:              "decided",
		WorkClass:          "planning",
		RequestCohort:      "option-compare",
		ArtifactFamily:     "comparison-matrix",
		MenuAliases:        []string{"compare options", "compare options and choose one", "vendor"},
		DetectSignals:      []string{"vendor", "compare options", "choose one", "recommendation memo"},
		PrimaryViewPath:    "recommendation-memo.md",
		PrimaryViewLabel:   "Recommendation Memo",
		WorkingWith:        "Vendor notes, tradeoffs, and decision criteria",
		Done:               []string{"Recommendation memo drafted", "Tradeoffs laid out"},
		NextStep:           "Open Recommendation Memo",
		PrioritizedViewIDs: []string{"recommendation-memo"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "recommendation-memo", FileStem: "selection", Label: "Recommendation Memo", Aliases: []string{"selection", "memo"}},
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
		PackID:             "incident-response",
		ChoiceLabel:        "Clean up an incident",
		DefaultName:        "Incident Cleanup",
		State:              "incident",
		WorkClass:          "code",
		RequestCohort:      "incident-cleanup",
		ArtifactFamily:     "step-plan",
		MenuAliases:        []string{"4", "incident", "clean up an incident"},
		DetectSignals:      []string{"incident", "outage", "customer impact", "root cause", "recovery"},
		PrimaryViewPath:    "closure-checklist.md",
		PrimaryViewLabel:   "Closure Checklist",
		WorkingWith:        "Incident notes, timeline, and follow-up tasks",
		Done:               []string{"Closure checklist drafted", "Recovery follow-ups pulled out"},
		NextStep:           "Open Closure Checklist",
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
		PackID:             "travel-plan",
		ChoiceLabel:        "Plan a trip",
		DefaultName:        "Trip Plan",
		State:              "decided",
		WorkClass:          "planning",
		RequestCohort:      "trip-itinerary",
		ArtifactFamily:     "itinerary-plan",
		MenuAliases:        []string{"5", "trip", "plan a trip"},
		DetectSignals:      []string{"trip", "travel", "hotel", "flight", "itinerary"},
		PrimaryViewPath:    "itinerary.md",
		PrimaryViewLabel:   "Itinerary",
		WorkingWith:        "Trip notes, dates, and planning details",
		Done:               []string{"Itinerary drafted", "Budget sketch created"},
		NextStep:           "Open Itinerary",
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
			options = append(options, "Skip")
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
		PackID:             "general-work",
		ChoiceLabel:        "Something else",
		DefaultName:        "Request Brief",
		State:              "decided",
		WorkClass:          "general",
		RequestCohort:      "general-pass",
		ArtifactFamily:     "general-pass",
		MenuAliases:        []string{"6", "something else", "something"},
		DetectSignals:      nil,
		PrimaryViewPath:    "first-useful-pass.md",
		PrimaryViewLabel:   "Request Brief",
		WorkingWith:        "The files and notes in this work",
		NextStep:           "Review what is ready",
		PrioritizedViewIDs: []string{"repo-review", "first-useful-pass", "next-actions"},
		SynthesizedViews: []starterCatalogSpec{
			{ID: "repo-review", FileStem: "repo-review", Label: "Repo Review", Aliases: []string{"repo review", "review", "repository review"}},
			{ID: "first-useful-pass", FileStem: "first-useful-pass", Label: "Request Brief", Aliases: []string{"request brief", "task snapshot", "working draft", "first pass", "useful pass", "summary", "draft"}},
		},
		TasksView: &starterTasksViewProfile{
			ID:      "next-actions",
			Label:   "Next Steps",
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
		Intro:          "Before I draft it, help me narrow the highest-impact details in one line:",
		SkipHint:       "Type `skip` if you want a generic first draft.",
		Example:        "Example: early October, mixed pace, central hotel area, one museum and one day trip are must-dos",
		MinimumMissing: 2,
		Dimensions: []scopePlannerDimension{
			scopeDimensionTravelers(),
			scopeDimensionBudgetRange(),
			scopeDimensionDatesOrSeason(),
			scopeDimensionPaceOrStyle(),
			scopeDimensionBaseArea(),
			scopeDimensionMustDoAnchors(),
		},
	},
	{
		RequestCohorts: []string{"build-readiness"},
		Intro:          "Before I draft it, help me narrow the highest-impact details in one line:",
		SkipHint:       "Type `skip` if you want a first pass with the gaps called out.",
		Example:        "Example: notifications PRD, first slice is digest emails, rollback is still open, approval owner is Priya",
		MinimumMissing: 2,
		Dimensions: []scopePlannerDimension{
			scopeDimensionPlanUnderReview(),
			scopeDimensionFirstSlice(),
			scopeDimensionKnownBlockers(),
			scopeDimensionApprovalOrOwner(),
		},
		PreferDraftSignalGroups: [][]string{
			{"prd", "handoff"},
			{"prd", "build readiness"},
			{"spec", "handoff"},
			{"spec", "build readiness"},
		},
	},
}

var travelScopeDimensions = []scopePlannerDimension{
	scopeDimensionTravelers(),
	scopeDimensionBudgetRange(),
	scopeDimensionDatesOrSeason(),
	scopeDimensionPaceOrStyle(),
	scopeDimensionBaseArea(),
	scopeDimensionMustDoAnchors(),
}

func scopeDimensionTravelers() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "travelers",
		NormalizedSignals: []string{"solo", "couple", "friends", "family", "kids", "children", "parents", "honeymoon", "wife", "husband", "partner"},
	}
}

func scopeDimensionBudgetRange() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "budget range",
		NormalizedSignals: []string{"cheap", "luxury", "midrange", "2500", "3000", "2000", "1500", "4000", "5000", "6000"},
		RawSignals:        []string{"$", "budget"},
	}
}

func scopeDimensionDatesOrSeason() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "dates or season",
		NormalizedSignals: []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december", "spring", "summer", "fall", "autumn", "winter", "weekend", "weekday", "christmas", "new year"},
	}
}

func scopeDimensionPaceOrStyle() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "pace or style",
		NormalizedSignals: []string{"food", "museum", "romantic", "nightlife", "shopping", "family friendly", "mixed", "slow pace", "fast pace", "walking", "architecture", "relaxed", "packed", "kid friendly", "honeymoon", "adventure"},
	}
}

func scopeDimensionBaseArea() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "base area, or whether you want help choosing one",
		NormalizedSignals: []string{"hotel", "stay", "marais", "latin quarter", "montmartre", "central", "area", "neighborhood", "neighbourhood", "arrondissement", "left bank", "right bank", "base"},
	}
}

func scopeDimensionMustDoAnchors() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "must-do anchors, or whether you want help choosing them",
		NormalizedSignals: []string{"louvre", "versailles", "eiffel", "orsay", "montmartre", "notre dame", "latin quarter", "marais", "disneyland", "seine cruise", "must do", "must see", "anchor"},
	}
}

func scopeDimensionPlanUnderReview() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "which plan or feature this is for",
		NormalizedSignals: []string{"notifications", "billing", "pricing", "onboarding", "checkout", "auth", "authentication", "search", "dashboard", "email", "mobile", "api", "portal", "integration", "migration", "feature", "prd", "spec"},
	}
}

func scopeDimensionFirstSlice() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "the first slice or decision this handoff should cover",
		NormalizedSignals: []string{"first slice", "first pass", "phase 1", "mvp", "rollout", "digest", "v1", "launch", "implementation slice", "cut this", "scope"},
		RawSignals:        []string{"slice", "phase"},
	}
}

func scopeDimensionKnownBlockers() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "known blockers, risks, or open gaps",
		NormalizedSignals: []string{"blocker", "risk", "rollback", "dependency", "open question", "missing", "approval gap", "legal", "compliance", "owner missing", "unclear"},
	}
}

func scopeDimensionApprovalOrOwner() scopePlannerDimension {
	return scopePlannerDimension{
		Label:             "approval owner or review owner",
		NormalizedSignals: []string{"owner", "approver", "approval", "reviewer", "pm", "eng", "design", "legal", "priya", "alex", "jordan"},
	}
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
	return scopePlannerForCohort(profile.RequestCohort)
}

func scopePlannerForCohort(requestCohort string) (scopePlannerProfile, bool) {
	for _, planner := range starterScopePlannerProfiles {
		for _, cohort := range planner.RequestCohorts {
			if cohort == requestCohort {
				return planner, true
			}
		}
	}
	return scopePlannerProfile{}, false
}

func clarificationPromptForProfile(profile starterPackProfile, source string) (string, bool) {
	return clarificationPromptForCohort(profile.RequestCohort, source)
}

func clarificationPromptForCohort(requestCohort, source string) (string, bool) {
	planner, ok := scopePlannerForCohort(requestCohort)
	if !ok {
		return "", false
	}
	missing := missingScopeDimensions(source, planner.Dimensions)
	if shouldPreferDraftForPlanner(source, planner, missing) {
		return "", false
	}
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

func shouldPreferDraftForPlanner(source string, planner scopePlannerProfile, missing []string) bool {
	if len(missing) <= 1 {
		return true
	}
	normalized := normalizeName(source)
	if plannerHandlesCohort(planner, "trip-itinerary") && clearTravelDraftPrompt(source) {
		return true
	}
	for _, group := range planner.PreferDraftSignalGroups {
		if len(group) == 0 {
			continue
		}
		allPresent := true
		for _, signal := range group {
			if !strings.Contains(normalized, normalizeName(signal)) {
				allPresent = false
				break
			}
		}
		if allPresent {
			return true
		}
	}
	return false
}

func plannerHandlesCohort(planner scopePlannerProfile, cohort string) bool {
	for _, candidate := range planner.RequestCohorts {
		if candidate == cohort {
			return true
		}
	}
	return false
}

func clearTravelDraftPrompt(source string) bool {
	normalized := normalizeName(source)
	if !containsAny(normalized, []string{"trip", "travel", "itinerary"}) {
		return false
	}
	padded := " " + normalized + " "
	hasDuration := extractTravelDayCount(source) > 0 ||
		strings.Contains(padded, " day ") ||
		strings.Contains(padded, " days ") ||
		strings.Contains(padded, " weekend ") ||
		strings.Contains(padded, " week ")
	return hasDuration && strings.TrimSpace(extractTravelDestination(source)) != ""
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
	return firstNonEmpty(profile.PrimaryViewPath, "first-useful-pass.md"), firstNonEmpty(profile.PrimaryViewLabel, "Request Brief")
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
		Options: []string{"Answer", "Skip"},
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
