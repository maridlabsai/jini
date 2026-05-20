package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type starterArtifactPlan struct {
	Docs          []starterArtifactDoc
	ArtifactTypes []string
}

type starterArtifactDoc struct {
	Path     string
	Title    string
	Sections []starterArtifactSection
}

type starterArtifactSection struct {
	Heading    string
	Paragraphs []string
	Bullets    []string
	RawLines   []string
}

func writeStarterArtifactPlan(workDir string, plan starterArtifactPlan) error {
	for _, doc := range plan.Docs {
		if err := os.WriteFile(filepath.Join(workDir, "views", doc.Path), []byte(renderStarterArtifactDoc(doc)), 0o644); err != nil {
			return err
		}
	}
	return writeStarterArtifacts(workDir, plan.ArtifactTypes)
}

func renderStarterArtifactDoc(doc starterArtifactDoc) string {
	lines := []string{fmt.Sprintf("# %s", strings.TrimSpace(doc.Title)), ""}
	for _, section := range doc.Sections {
		if strings.TrimSpace(section.Heading) != "" {
			lines = append(lines, "## "+strings.TrimSpace(section.Heading))
		}
		for _, paragraph := range section.Paragraphs {
			if strings.TrimSpace(paragraph) == "" {
				continue
			}
			lines = append(lines, strings.TrimSpace(paragraph))
			lines = append(lines, "")
		}
		for _, bullet := range section.Bullets {
			if strings.TrimSpace(bullet) == "" {
				continue
			}
			lines = append(lines, "- "+strings.TrimSpace(bullet))
		}
		if len(section.RawLines) > 0 {
			lines = append(lines, section.RawLines...)
		}
		lines = append(lines, "")
	}
	return strings.Join(trimTrailingBlankLines(lines), "\n") + "\n"
}

func trimTrailingBlankLines(lines []string) []string {
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return append([]string{}, lines[:end]...)
}

func buildMeetingArtifactPlan(title, source, detail string) starterArtifactPlan {
	decisions := starterSourceBullets(source, 4, []string{
		"Summarize the main meeting decisions in one short list before sending anything.",
	})
	ownersToConfirm := starterMeetingOwnersToConfirm(source, decisions)
	openQuestions := starterMeetingOpenQuestions(source)
	nextMoves := starterMeetingNextMoves(source)
	followupSections := []starterArtifactSection{
		{
			Heading: "Send this note",
			Paragraphs: []string{
				fmt.Sprintf("Team, here is the clean follow-up from **%s**.", title),
				fmt.Sprintf("I pulled this from the current notes: %s", source),
				"Please reply if any owner, due date, dependency, or open question below needs correction before this is treated as final.",
			},
		},
		{Heading: "Decisions captured from the notes", Bullets: decisions},
		{Heading: "Owners and due dates to confirm", Bullets: ownersToConfirm},
		{Heading: "Open questions to close", Bullets: openQuestions},
		{Heading: "Recommended next move", Bullets: nextMoves},
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(detail)), "f") {
		followupSections = append(followupSections, starterArtifactSection{
			Heading: "Why this note exists",
			Bullets: []string{
				"Decisions, owners, and due dates must stay explicit before the meeting is considered closed.",
				"Open questions should stay visible instead of getting buried in notes or chat.",
			},
		})
	}
	return starterArtifactPlan{
		Docs: []starterArtifactDoc{
			{
				Path:     "followup.md",
				Title:    fmt.Sprintf("Sendable Follow-Up: %s", title),
				Sections: followupSections,
			},
			{
				Path:  "owners-and-due-points.md",
				Title: "Owners and Due Points",
				Sections: []starterArtifactSection{
					{Heading: "Confirmed from the notes", Bullets: decisions},
					{Heading: "Still missing owner or date", Bullets: ownersToConfirm},
					{Heading: "Follow-up questions", Bullets: openQuestions},
				},
			},
			{
				Path:  "tasks.md",
				Title: "Task List",
				Sections: []starterArtifactSection{
					{Bullets: ownersToConfirm},
				},
			},
		},
		ArtifactTypes: []string{"Brief", "Tasks"},
	}
}

func buildResearchArtifactPlan(title, source, detail string) starterArtifactPlan {
	readyNow := starterSourceBullets(source, 3, []string{
		"User intent is clear enough to shape the first build slice.",
	})
	mustClear := starterResearchMustClear(source)
	firstSlice := starterResearchFirstSlice(source)
	whoNeeds := starterResearchWhoNeeds(source)
	stillToConfirm := starterResearchStillToConfirm(source)
	prdSections := []starterArtifactSection{
		{Heading: "What looks ready now", Bullets: readyNow},
		{Heading: "Must clear before build", Bullets: mustClear},
		{Heading: "Recommended first slice", Bullets: firstSlice},
		{Heading: "Who needs to answer what", Bullets: whoNeeds},
		{Heading: "Still to confirm", Bullets: stillToConfirm},
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(detail)), "f") {
		prdSections = append(prdSections, starterArtifactSection{
			Heading: "Risks to clear",
			Bullets: []string{
				"Product approval may still be implicit rather than recorded.",
				"Rollback behavior is still under-specified.",
				"The highest-risk assumption needs one owner before build starts.",
			},
		})
	}
	return starterArtifactPlan{
		Docs: []starterArtifactDoc{
			{
				Path:     "prd.md",
				Title:    fmt.Sprintf("Build-Readiness Check: %s", title),
				Sections: prdSections,
			},
			{
				Path:  "missing-pieces-before-build.md",
				Title: "Missing Pieces Before Build",
				Sections: []starterArtifactSection{
					{Heading: "Must clear before build", Bullets: mustClear},
					{Heading: "Who needs to answer what", Bullets: whoNeeds},
					{Heading: "Still to confirm", Bullets: stillToConfirm},
				},
			},
			{
				Path:  "tasks.md",
				Title: "Missing Pieces Before Build",
				Sections: []starterArtifactSection{
					{Bullets: mustClear},
				},
			},
		},
		ArtifactTypes: []string{"Brief", "Plan", "Tasks", "Evidence"},
	}
}

func buildTravelArtifactPlan(title, source string) starterArtifactPlan {
	ctx := parseTravelStarterContext(source)
	return starterArtifactPlan{
		Docs: []starterArtifactDoc{
			{
				Path:  "itinerary.md",
				Title: fmt.Sprintf("Itinerary: %s", title),
				Sections: []starterArtifactSection{
					{
						Heading:    "Trip at a glance",
						Paragraphs: []string{source, "This draft gives you a usable week shape first, then shows the booking, budget, and contingency items to lock next."},
					},
					{
						Heading:  "Day-by-day draft",
						RawLines: starterTripDays(ctx),
					},
					{Heading: "Budget sketch", Bullets: starterTripBudget(ctx)},
					{Heading: "Logistics to lock", Bullets: starterTripLogistics(ctx)},
					{Heading: "If something changes", Bullets: starterTripContingencies(ctx)},
					{Heading: "Still to confirm", Bullets: travelStillToConfirm(ctx)},
				},
			},
			{
				Path:  "budget-sketch.md",
				Title: "Budget Sketch",
				Sections: []starterArtifactSection{
					{Bullets: starterTripBudget(ctx)},
				},
			},
			{
				Path:  "travel-logistics.md",
				Title: "Travel Logistics",
				Sections: []starterArtifactSection{
					{Bullets: starterTripLogistics(ctx)},
				},
			},
			{
				Path:  "tasks.md",
				Title: "Task List",
				Sections: []starterArtifactSection{
					{Bullets: travelTaskList(ctx)},
				},
			},
		},
		ArtifactTypes: []string{"Brief", "Tasks"},
	}
}

func buildFirstUsefulPassArtifactPlan(title, source string) starterArtifactPlan {
	return starterArtifactPlan{
		Docs: []starterArtifactDoc{
			{
				Path:  "first-useful-pass.md",
				Title: fmt.Sprintf("Working Draft: %s", title),
				Sections: []starterArtifactSection{
					{Heading: "What this looks like", Bullets: []string{source}},
					{
						Heading: "Useful starting point",
						Bullets: []string{
							"This is enough to begin shaping a real output without guessing hidden details.",
							"The next pass can turn this into a follow-up, plan check, memo, checklist, or another concrete artifact.",
						},
					},
					{
						Heading: "Best next inputs",
						Bullets: []string{
							"The audience or recipient.",
							"The outcome you want after someone reads or uses this.",
							"Any deadline, owner, blocker, or decision that should not be guessed.",
						},
					},
					{
						Heading: "Safe right now",
						Bullets: []string{
							"Nothing has been sent, changed, booked, or committed.",
							"You can review this pass before sharing or turning it into a fuller artifact.",
						},
					},
				},
			},
			{
				Path:  "tasks.md",
				Title: "What I Need Next",
				Sections: []starterArtifactSection{
					{Bullets: []string{"Name the audience or recipient.", "Confirm the desired outcome.", "Add any deadline, owner, blocker, or decision."}},
				},
			},
		},
		ArtifactTypes: []string{"Brief", "Tasks"},
	}
}

func buildSimpleArtifactPlan(title, viewLabel, source string, bullets []string) starterArtifactPlan {
	return starterArtifactPlan{
		Docs: []starterArtifactDoc{
			{
				Path:  normalizeFilename(viewLabel) + ".md",
				Title: fmt.Sprintf("%s: %s", viewLabel, title),
				Sections: []starterArtifactSection{
					{Paragraphs: []string{source}, Bullets: bullets},
				},
			},
			{
				Path:  "tasks.md",
				Title: "Task List",
				Sections: []starterArtifactSection{
					{Bullets: []string{"Review what is ready.", "Confirm what is still missing."}},
				},
			},
		},
		ArtifactTypes: []string{"Brief", "Tasks"},
	}
}
