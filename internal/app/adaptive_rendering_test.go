package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestProjectSemanticEnvelopeForMeetingFollowupKeepsProfileTruth(t *testing.T) {
	summary := meetingSemanticTestSummary(t.TempDir())
	inputs := []inputItem{{
		InputID:   "request",
		Kind:      "text",
		Title:     "Your request",
		Status:    "processed",
		Preview:   "Notes from weekly product review",
		OriginRef: "Metrics need owner. Legal review date is unclear.",
	}}
	state := savedThreadState{
		CurrentTurn: threadTurnRecord{
			JustFinished:     []string{"Sendable follow-up drafted"},
			DoingNow:         "Turning notes into owners and next steps",
			UpNext:           "Open Sendable Follow-up",
			ArtifactsCreated: []string{"Sendable Follow-up", "Owners and Due Points"},
		},
	}

	envelope := projectSemanticEnvelope(summary, inputs, state)

	if envelope.SchemaVersion != "0.1.0" {
		t.Fatalf("expected semantic schema version, got %q", envelope.SchemaVersion)
	}
	if envelope.IntentKind != "create_artifact" {
		t.Fatalf("expected create_artifact intent, got %q", envelope.IntentKind)
	}
	if envelope.WorkClass != "planning" {
		t.Fatalf("expected planning work class, got %q", envelope.WorkClass)
	}
	if envelope.ArtifactFamily != "narrative-draft" {
		t.Fatalf("expected narrative-draft artifact family, got %q", envelope.ArtifactFamily)
	}
	if !slices.Contains(envelope.Missing, "Metric and legal-review decision") {
		t.Fatalf("expected meeting blocker in semantic envelope, got %#v", envelope.Missing)
	}
	if got := envelope.TurnDelta.ArtifactsCreated; !slices.Contains(got, "Sendable Follow-up") {
		t.Fatalf("expected turn delta to preserve created artifacts, got %#v", got)
	}
	if len(envelope.Artifacts) == 0 || envelope.Artifacts[0].Title != "Sendable Follow-up" {
		t.Fatalf("expected first artifact envelope to be Sendable Follow-up, got %#v", envelope.Artifacts)
	}
}

func TestRenderPolicyForMeetingFirstResultUsesCompactCLIWithoutChangingTruth(t *testing.T) {
	summary := meetingSemanticTestSummary(t.TempDir())
	envelope := projectSemanticEnvelope(summary, nil, synthesizeThreadState(summary))

	request := selectRenderRequest(envelope, "cli", "new")

	if request.Surface != "cli" {
		t.Fatalf("expected cli surface, got %q", request.Surface)
	}
	if request.Mode != "first_result" {
		t.Fatalf("expected first_result mode, got %q", request.Mode)
	}
	if request.Density != "compact" {
		t.Fatalf("expected compact density, got %q", request.Density)
	}
	for _, want := range []string{"Keep going", "Make it fuller", "Show what is missing", "Start new work"} {
		if !slices.Contains(request.AvailableActions, want) {
			t.Fatalf("expected render action %q in %#v", want, request.AvailableActions)
		}
	}
	if request.RouteVisibility != "compact" {
		t.Fatalf("expected compact route visibility, got %q", request.RouteVisibility)
	}
	if envelope.Artifacts[0].Title != "Sendable Follow-up" {
		t.Fatalf("render policy changed artifact truth: %#v", envelope.Artifacts)
	}
}

func TestBootstrapStarterWorkPersistsMeetingSemanticEnvelope(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-preview")

	summary, err := bootstrapStarterWork(
		starterChoice{PackID: "meeting-followup", DefaultName: "Meeting Follow-up", State: "decided"},
		"Weekly product review notes. Metrics need owner. Legal review date is unclear.",
		"quick",
		[]inputItem{{
			InputID:   "request",
			Kind:      "text",
			Title:     "Your request",
			Status:    "processed",
			Preview:   "Weekly product review notes",
			OriginRef: "Weekly product review notes. Metrics need owner. Legal review date is unclear.",
		}},
	)
	if err != nil {
		t.Fatalf("bootstrap starter work: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(summary.Dir, "semantic-envelope.json"))
	if err != nil {
		t.Fatalf("expected semantic envelope to be persisted: %v", err)
	}
	var envelope semanticEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode semantic envelope: %v", err)
	}

	if envelope.ContextType != "JiniSemanticEnvelope" {
		t.Fatalf("expected semantic envelope context, got %q", envelope.ContextType)
	}
	if envelope.WorkClass != "planning" || envelope.ArtifactFamily != "narrative-draft" {
		t.Fatalf("expected meeting semantic metadata, got work=%q family=%q", envelope.WorkClass, envelope.ArtifactFamily)
	}
	if len(envelope.RenderHints) == 0 || envelope.RenderHints[0].Surface != "cli" {
		t.Fatalf("expected persisted CLI render hint, got %#v", envelope.RenderHints)
	}
}

func meetingSemanticTestSummary(dir string) *workSummary {
	return &workSummary{
		Dir:        dir,
		PackID:     "meeting-followup",
		WorkUnitID: "weekly-product-review",
		Title:      "Weekly Product Review",
		State:      "decided",
		Views: []catalogItem{
			{ID: "sendable-follow-up", Label: "Sendable Follow-up", Path: filepath.Join(dir, "views", "followup.md")},
			{ID: "owners-and-due-points", Label: "Owners and Due Points", Path: filepath.Join(dir, "views", "owners-and-due-points.md")},
		},
		Missing:     []string{"Metric and legal-review decision"},
		Uncertain:   []string{"Whether the metric decision also needs legal review"},
		Doing:       "Turning notes into owners and next steps",
		NextStep:    "Open Sendable Follow-up",
		SafeToDo:    "Nothing has been sent yet. You can review before sharing.",
		WorkingWith: "Local preview (chosen automatically)",
	}
}
