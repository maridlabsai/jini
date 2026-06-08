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
				Title: fmt.Sprintf("Task Snapshot: %s", title),
				Sections: []starterArtifactSection{
					{Heading: "Request", Bullets: []string{source}},
					{
						Heading: "Current read",
						Bullets: []string{
							"Jini saved the task because it did not match a safe direct action yet.",
							"The next step should either route to a configured CLI, use a local model, or ask for one missing detail.",
						},
					},
					{
						Heading: "Next options",
						Bullets: []string{
							"Type `Continue` to refine this task.",
							"Type `Open` to inspect saved artifacts.",
							"Describe a new task when you want to move on.",
						},
					},
					{
						Heading: "Safety",
						Bullets: []string{
							"Nothing has been sent, changed, booked, or committed.",
							"Jini should prefer a visible route or explicit confirmation before side effects.",
						},
					},
				},
			},
			{
				Path:  "tasks.md",
				Title: "Next Actions",
				Sections: []starterArtifactSection{
					{Bullets: []string{"Resolve the best route for this task.", "Use a configured downstream CLI when available.", "Use local/offline execution only when the task can be handled safely."}},
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
