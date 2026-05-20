package app

import (
	"strings"
	"testing"
)

func TestClassifyWorkEnvelopeUsesSharedTravelMetadata(t *testing.T) {
	envelope := classifyWorkEnvelope(starterChoice{}, "7 day Paris trip for a couple with a $2500 budget")
	if envelope.Choice.PackID != "travel-plan" {
		t.Fatalf("expected travel-plan choice, got %#v", envelope)
	}
	if envelope.WorkClass != "planning" {
		t.Fatalf("expected planning work class, got %#v", envelope)
	}
	if envelope.RequestCohort != "trip-itinerary" {
		t.Fatalf("expected trip-itinerary cohort, got %#v", envelope)
	}
	if envelope.ArtifactFamily != "itinerary-plan" {
		t.Fatalf("expected itinerary-plan artifact family, got %#v", envelope)
	}
}

func TestClassifyWorkEnvelopePreservesExplicitChoiceOverDetectedSignals(t *testing.T) {
	envelope := classifyWorkEnvelope(
		starterChoice{PackID: "meeting-followup", ChoiceLabel: "Meeting", DefaultName: "Meeting Follow-up", State: "decided"},
		"Plan a 7 day Paris trip",
	)
	if envelope.Choice.PackID != "meeting-followup" {
		t.Fatalf("expected explicit meeting choice to win, got %#v", envelope)
	}
	if envelope.RequestCohort != "sendable-followup" {
		t.Fatalf("expected meeting cohort metadata, got %#v", envelope)
	}
}

func TestClarificationPromptForEnvelopeUsesSharedCohortPlanner(t *testing.T) {
	envelope := workEnvelope{
		Choice:        starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"},
		Source:        "7 day Paris trip",
		RequestCohort: "trip-itinerary",
	}
	prompt, ok := clarificationPromptForEnvelope(envelope)
	if !ok {
		t.Fatalf("expected clarification prompt for underspecified travel envelope")
	}
	for _, want := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- travelers",
		"- budget range",
		"- must-do anchors, or whether you want help choosing them",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestClarificationPromptForEnvelopeSupportsBuildReadinessCohort(t *testing.T) {
	envelope := workEnvelope{
		Choice:        starterChoice{PackID: "research-prd", DefaultName: "Plan Readiness", State: "awaiting_verification"},
		Source:        "check whether this plan is ready to hand off",
		RequestCohort: "build-readiness",
	}
	prompt, ok := clarificationPromptForEnvelope(envelope)
	if !ok {
		t.Fatalf("expected clarification prompt for underspecified build-readiness envelope")
	}
	for _, want := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- which plan or feature this is for",
		"- the first slice or decision this handoff should cover",
		"- known blockers, risks, or open gaps",
		"- approval owner or review owner",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestClarificationPromptForEnvelopePrefersDraftForSubstantiveBuildReadinessAsk(t *testing.T) {
	envelope := workEnvelope{
		Choice:        starterChoice{PackID: "research-prd", DefaultName: "Plan Readiness", State: "awaiting_verification"},
		Source:        "notifications PRD needs a build-readiness check and handoff call",
		RequestCohort: "build-readiness",
	}
	if prompt, ok := clarificationPromptForEnvelope(envelope); ok {
		t.Fatalf("expected substantive build-readiness ask to draft directly, got prompt:\n%s", prompt)
	}
}
