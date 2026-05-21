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

	envelope := threadProjector{}.Project(summary, inputs, state)

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
	envelope := threadProjector{}.Project(summary, nil, synthesizeThreadState(summary))

	request := renderPolicy{}.Select(envelope, "cli", "new")

	if request.Surface != "cli" {
		t.Fatalf("expected cli surface, got %q", request.Surface)
	}
	if request.Mode != "first_result" {
		t.Fatalf("expected first_result mode, got %q", request.Mode)
	}
	if request.Density != "compact" {
		t.Fatalf("expected compact density, got %q", request.Density)
	}
	for _, want := range []string{"Continue", "Missing", "Plan", "Start"} {
		if !slices.Contains(request.AvailableActions, want) {
			t.Fatalf("expected render action %q in %#v", want, request.AvailableActions)
		}
	}
	if slices.Contains(request.AvailableActions, "Expand") {
		t.Fatalf("expected placeholder action to be removed from first-result actions: %#v", request.AvailableActions)
	}
	if request.RouteVisibility != "compact" {
		t.Fatalf("expected compact route visibility, got %q", request.RouteVisibility)
	}
	if envelope.Artifacts[0].Title != "Sendable Follow-up" {
		t.Fatalf("render policy changed artifact truth: %#v", envelope.Artifacts)
	}
}

func TestSemanticEnvelopeNormalizesArtifactIDsForSurfaceContracts(t *testing.T) {
	dir := t.TempDir()
	summary := &workSummary{
		Dir:        dir,
		PackID:     "travel-plan",
		WorkUnitID: "7-day-paris-trip",
		Title:      "7 Day Paris Trip",
		State:      "decided",
		Views: []catalogItem{
			{ID: "itinerary", Label: "Itinerary", Path: filepath.Join(dir, "views", "itinerary.md")},
			{ID: "budget sketch", Label: "Budget Sketch", Path: filepath.Join(dir, "views", "budget-sketch.md")},
			{ID: "travel logistics", Label: "Travel Logistics", Path: filepath.Join(dir, "views", "travel-logistics.md")},
		},
		Missing:     []string{"Must do sights"},
		Uncertain:   []string{"Which anchor experience should be locked first"},
		NextStep:    "Open Itinerary",
		WorkingWith: "Local preview",
	}

	envelope := threadProjector{}.Project(summary, nil, synthesizeThreadState(summary))

	got := []string{}
	for _, artifact := range envelope.Artifacts {
		got = append(got, artifact.ArtifactID)
	}
	for _, want := range []string{"itinerary", "budget-sketch", "travel-logistics"} {
		if !slices.Contains(got, want) {
			t.Fatalf("expected normalized artifact id %q in %#v", want, got)
		}
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
	if envelope.Decisions == nil || envelope.Artifacts[0].SourceRefs == nil {
		t.Fatalf("expected stable empty arrays in semantic envelope, got decisions=%#v source_refs=%#v", envelope.Decisions, envelope.Artifacts[0].SourceRefs)
	}
	if envelope.ActiveAsk == nil || !envelope.ActiveAsk.Blocking {
		t.Fatalf("expected blocking ask to be preserved in semantic envelope, got %#v", envelope.ActiveAsk)
	}
	if envelope.TurnID == "current-turn" || envelope.TurnID == "" {
		t.Fatalf("expected turn id to be derived from turn state, got %q", envelope.TurnID)
	}
}

func TestLoadWorkSummaryRefreshesSemanticEnvelopeSidecar(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-preview")

	summary, err := bootstrapStarterWork(
		starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"},
		"7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area",
		"quick",
		[]inputItem{{
			InputID:   "request",
			Kind:      "text",
			Title:     "Your request",
			Status:    "processed",
			Preview:   "7 day Paris trip",
			OriginRef: "7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area",
		}},
	)
	if err != nil {
		t.Fatalf("bootstrap starter work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(summary.Dir, "semantic-envelope.json"), []byte("{\"schema_version\":\"stale\"}\n"), 0o600); err != nil {
		t.Fatalf("write stale envelope: %v", err)
	}

	reloaded, err := loadWorkSummary(summary.Dir, nil)
	if err != nil {
		t.Fatalf("load work summary: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(reloaded.Dir, "semantic-envelope.json"))
	if err != nil {
		t.Fatalf("read refreshed envelope: %v", err)
	}
	var envelope semanticEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode refreshed envelope: %v", err)
	}

	if envelope.ContextType != "JiniSemanticEnvelope" || envelope.ThreadID != "7-day-paris-trip" {
		t.Fatalf("expected refreshed semantic envelope, got context=%q thread=%q", envelope.ContextType, envelope.ThreadID)
	}
	artifactIDs := []string{}
	for _, artifact := range envelope.Artifacts {
		artifactIDs = append(artifactIDs, artifact.ArtifactID)
	}
	for _, want := range []string{"itinerary", "budget-sketch", "travel-logistics", "still-to-book"} {
		if !slices.Contains(artifactIDs, want) {
			t.Fatalf("expected refreshed envelope to preserve normalized view id %q, got %#v", want, artifactIDs)
		}
	}
}

func TestLoadWorkSummaryIgnoresSemanticEnvelopeWriteFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "views"), 0o755); err != nil {
		t.Fatalf("mkdir views: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "semantic-envelope.json"), 0o755); err != nil {
		t.Fatalf("mkdir semantic envelope conflict: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "work-unit.yaml"), []byte("work_unit_id: write-failure\npack_id: meeting-followup\ntitle: Write Failure\ncurrent_state: decided\n"), 0o644); err != nil {
		t.Fatalf("write work unit: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "views", "followup.md"), []byte("# Sendable Follow-up\n\nDraft.\n"), 0o644); err != nil {
		t.Fatalf("write view: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "artifacts", "01-brief.yaml"), []byte("artifact_type: Brief\nstatus: ready\n"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	summary, err := loadWorkSummary(root, nil)
	if err != nil {
		t.Fatalf("semantic envelope write failure should not block summary load: %v", err)
	}
	if summary.Title != "Write Failure" {
		t.Fatalf("expected summary to load despite envelope write failure, got %q", summary.Title)
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
