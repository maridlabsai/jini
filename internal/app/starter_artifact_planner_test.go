package app

import (
	"strings"
	"testing"
)

func TestBuildMeetingArtifactPlanUsesSharedDocumentStructure(t *testing.T) {
	plan := buildMeetingArtifactPlan("Weekly Product Review", "Weekly product review for pricing launch. Need owners and due dates.", "")
	if len(plan.Docs) != 3 {
		t.Fatalf("expected three docs, got %#v", plan.Docs)
	}
	if plan.Docs[0].Path != "followup.md" || !strings.Contains(plan.Docs[0].Title, "Sendable Follow-Up") {
		t.Fatalf("expected sendable follow-up primary doc, got %#v", plan.Docs[0])
	}
	rendered := renderStarterArtifactDoc(plan.Docs[0])
	for _, want := range []string{
		"# Sendable Follow-Up: Weekly Product Review",
		"## Send this note",
		"## Decisions captured from the notes",
		"## Owners and due dates to confirm",
		"## Open questions to close",
		"## Recommended next move",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rendered meeting artifact to contain %q, got:\n%s", want, rendered)
		}
	}
}

func TestBuildResearchArtifactPlanUsesSharedDocumentStructure(t *testing.T) {
	plan := buildResearchArtifactPlan("Notifications", "Notifications PRD needs a build-readiness check and handoff call.", "")
	if len(plan.Docs) != 3 {
		t.Fatalf("expected three docs, got %#v", plan.Docs)
	}
	if plan.Docs[0].Path != "prd.md" || !strings.Contains(plan.Docs[0].Title, "Build-Readiness Check") {
		t.Fatalf("expected build-readiness primary doc, got %#v", plan.Docs[0])
	}
	rendered := renderStarterArtifactDoc(plan.Docs[0])
	for _, want := range []string{
		"# Build-Readiness Check: Notifications",
		"## What looks ready now",
		"## Must clear before build",
		"## Recommended first slice",
		"## Who needs to answer what",
		"## Still to confirm",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rendered research artifact to contain %q, got:\n%s", want, rendered)
		}
	}
}

func TestBuildTravelArtifactPlanKeepsSharedAndRawStructure(t *testing.T) {
	plan := buildTravelArtifactPlan("7 Day Paris Trip", "7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area")
	if len(plan.Docs) != 4 {
		t.Fatalf("expected four docs, got %#v", plan.Docs)
	}
	rendered := renderStarterArtifactDoc(plan.Docs[0])
	for _, want := range []string{
		"# Itinerary: 7 Day Paris Trip",
		"## Trip at a glance",
		"## Day-by-day draft",
		"### Day 1: Arrive and settle into Paris",
		"## Budget sketch",
		"## Logistics to lock",
		"## Still to confirm",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected rendered travel artifact to contain %q, got:\n%s", want, rendered)
		}
	}
}
