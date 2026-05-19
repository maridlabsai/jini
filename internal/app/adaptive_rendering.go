package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type semanticEnvelope struct {
	SchemaVersion         string               `json:"schema_version"`
	ContextType           string               `json:"context_type"`
	ThreadID              string               `json:"thread_id"`
	TurnID                string               `json:"turn_id"`
	IntentKind            string               `json:"intent_kind"`
	WorkClass             string               `json:"work_class"`
	ArtifactFamily        string               `json:"artifact_family"`
	Stage                 string               `json:"stage"`
	Complexity            string               `json:"complexity"`
	InputQuality          string               `json:"input_quality"`
	SurfaceRecommendation string               `json:"surface_recommendation"`
	RouteSummary          semanticRouteSummary `json:"route_summary"`
	Artifacts             []semanticArtifact   `json:"artifacts"`
	Decisions             []string             `json:"decisions"`
	Missing               []string             `json:"missing"`
	Uncertainty           []string             `json:"uncertainty"`
	NextAction            string               `json:"next_action"`
	TrustState            string               `json:"trust_state"`
	ConfirmationRequired  bool                 `json:"confirmation_required"`
	ActiveAsk             *threadAsk           `json:"active_ask,omitempty"`
	TurnDelta             threadTurnRecord     `json:"turn_delta"`
	RenderHints           []renderRequest      `json:"render_hints"`
}

type semanticRouteSummary struct {
	Label             string `json:"label"`
	Reason            string `json:"reason,omitempty"`
	Policy            string `json:"policy,omitempty"`
	ContinuityReason  string `json:"continuity_reason,omitempty"`
	ModelLabel        string `json:"model_label,omitempty"`
	EffortLevel       string `json:"effort_level,omitempty"`
	VerificationLevel string `json:"verification_level,omitempty"`
}

type semanticArtifact struct {
	ArtifactID    string   `json:"artifact_id"`
	Family        string   `json:"family"`
	Title         string   `json:"title"`
	Purpose       string   `json:"purpose"`
	Status        string   `json:"status"`
	SourceRefs    []string `json:"source_refs"`
	OpenDecisions []string `json:"open_decisions"`
	MissingInputs []string `json:"missing_inputs"`
	TrustState    string   `json:"trust_state"`
	RenderHints   []string `json:"render_hints"`
}

type renderRequest struct {
	Surface          string   `json:"surface"`
	Mode             string   `json:"mode"`
	Density          string   `json:"density"`
	UserFamiliarity  string   `json:"user_familiarity"`
	RiskLevel        string   `json:"risk_level"`
	RouteVisibility  string   `json:"route_visibility"`
	AvailableActions []string `json:"available_actions"`
}

type threadProjector struct{}

func (threadProjector) Project(summary *workSummary, inputs []inputItem, state savedThreadState) semanticEnvelope {
	return projectSemanticEnvelope(summary, inputs, state)
}

type renderPolicy struct{}

func (renderPolicy) Select(envelope semanticEnvelope, surface, userFamiliarity string) renderRequest {
	return selectRenderRequest(envelope, surface, userFamiliarity)
}

func semanticEnvelopePath(workDir string) string {
	return filepath.Join(workDir, "semantic-envelope.json")
}

func saveSemanticEnvelopeForSummary(summary *workSummary) error {
	if summary == nil || strings.TrimSpace(summary.Dir) == "" {
		return nil
	}
	inputs := summary.Thread.InputItems
	if len(inputs) == 0 {
		inputs = loadInputItems(summary.Dir, summary.PackID)
	}
	state := loadThreadState(summary.Dir, summary)
	projector := threadProjector{}
	policy := renderPolicy{}
	envelope := projector.Project(summary, inputs, state)
	envelope.RenderHints = []renderRequest{policy.Select(envelope, "cli", "new")}
	return saveSemanticEnvelope(summary.Dir, envelope)
}

func refreshSemanticEnvelopeForSummary(summary *workSummary) {
	_ = saveSemanticEnvelopeForSummary(summary)
}

func saveSemanticEnvelope(workDir string, envelope semanticEnvelope) error {
	if strings.TrimSpace(workDir) == "" {
		return nil
	}
	envelope = normalizeSemanticEnvelope(envelope)
	envelope.SchemaVersion = firstNonEmpty(envelope.SchemaVersion, "0.1.0")
	envelope.ContextType = firstNonEmpty(envelope.ContextType, "JiniSemanticEnvelope")
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	path := semanticEnvelopePath(workDir)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func projectSemanticEnvelope(summary *workSummary, inputs []inputItem, state savedThreadState) semanticEnvelope {
	if summary == nil {
		return semanticEnvelope{
			SchemaVersion: "0.1.0",
			ContextType:   "JiniSemanticEnvelope",
			IntentKind:    "answer",
			Stage:         "recovery",
			Complexity:    "low",
			InputQuality:  "unknown",
			TrustState:    "unknown",
		}
	}

	profile := starterProfile(summary.PackID)
	turn := state.CurrentTurn
	if len(turn.JustFinished) == 0 && turn.DoingNow == "" && turn.UpNext == "" {
		turn = synthesizeTurnRecord(summary)
	}
	routeLabel := strings.TrimSpace(summary.WorkingWith)
	if routeLabel == "" {
		routeLabel = workingWithLabel(detectProvider())
	}
	missing := append([]string{}, summary.Missing...)
	envelope := semanticEnvelope{
		SchemaVersion:         "0.1.0",
		ContextType:           "JiniSemanticEnvelope",
		ThreadID:              firstNonEmpty(summary.WorkUnitID, slugify(summary.Title)),
		TurnID:                semanticTurnID(summary, turn),
		IntentKind:            semanticIntentKind(summary),
		WorkClass:             firstNonEmpty(profile.WorkClass, summary.PackID, "general"),
		ArtifactFamily:        firstNonEmpty(profile.ArtifactFamily, "general-pass"),
		Stage:                 semanticStage(summary, state),
		Complexity:            semanticComplexity(summary, inputs),
		InputQuality:          semanticInputQuality(inputs),
		SurfaceRecommendation: "artifact_first",
		RouteSummary: semanticRouteSummary{
			Label:             routeLabel,
			Reason:            summary.RouteReason,
			Policy:            summary.RoutePolicy,
			ContinuityReason:  summary.ContinuityReason,
			ModelLabel:        summary.ModelLabel,
			EffortLevel:       summary.EffortLevel,
			VerificationLevel: summary.VerificationLevel,
		},
		Artifacts:            semanticArtifacts(summary, profile, inputs),
		Decisions:            []string{},
		Missing:              missing,
		Uncertainty:          append([]string{}, summary.Uncertain...),
		NextAction:           summary.NextStep,
		TrustState:           semanticTrustState(summary),
		ConfirmationRequired: len(missing) > 0 || (state.ActiveAsk != nil && state.ActiveAsk.Blocking),
		ActiveAsk:            state.ActiveAsk,
		TurnDelta:            turn,
	}
	return envelope
}

func selectRenderRequest(envelope semanticEnvelope, surface, userFamiliarity string) renderRequest {
	surface = firstNonEmpty(strings.TrimSpace(surface), "cli")
	userFamiliarity = firstNonEmpty(strings.TrimSpace(userFamiliarity), "new")
	mode := envelope.Stage
	if mode == "" {
		mode = "first_result"
	}
	density := "standard"
	if surface == "cli" || surface == "mobile" {
		density = "compact"
	}
	return renderRequest{
		Surface:          surface,
		Mode:             mode,
		Density:          density,
		UserFamiliarity:  userFamiliarity,
		RiskLevel:        renderRiskLevel(envelope),
		RouteVisibility:  routeVisibility(envelope),
		AvailableActions: renderActionsForMode(mode),
	}
}

func semanticIntentKind(summary *workSummary) string {
	if summary != nil && len(summary.Views) > 0 {
		return "create_artifact"
	}
	return "answer"
}

func semanticStage(summary *workSummary, state savedThreadState) string {
	if state.ActiveAsk != nil && state.ActiveAsk.Blocking && len(summary.Views) == 0 {
		return "ask"
	}
	if summary != nil && len(summary.Views) > 0 {
		return "first_result"
	}
	return "preflight"
}

func semanticComplexity(summary *workSummary, inputs []inputItem) string {
	if summary != nil && (len(summary.Missing) > 0 || len(summary.Uncertain) > 0) {
		return "medium"
	}
	if len(sourceFromInputItems(inputs)) > 240 {
		return "medium"
	}
	return "low"
}

func semanticInputQuality(inputs []inputItem) string {
	if len(inputs) > 0 {
		allDerived := true
		for _, item := range inputs {
			if item.Kind != "derived" {
				allDerived = false
				break
			}
		}
		if allDerived {
			return "derived"
		}
	}
	source := strings.TrimSpace(sourceFromInputItems(inputs))
	switch {
	case source == "":
		return "derived"
	case len(source) > 160:
		return "scoped"
	default:
		return "minimal"
	}
}

func semanticArtifacts(summary *workSummary, profile starterPackProfile, inputs []inputItem) []semanticArtifact {
	if summary == nil {
		return nil
	}
	sourceRefs := inputSourceRefs(inputs)
	artifacts := make([]semanticArtifact, 0, len(summary.Views))
	for index, item := range summary.Views {
		purpose := "Supporting result"
		if index == 0 || item.Label == profile.PrimaryViewLabel {
			purpose = "Primary user-facing result"
		}
		artifacts = append(artifacts, semanticArtifact{
			ArtifactID:    semanticArtifactID(item),
			Family:        firstNonEmpty(profile.ArtifactFamily, "general-pass"),
			Title:         item.Label,
			Purpose:       purpose,
			Status:        semanticArtifactStatus(summary),
			SourceRefs:    sourceRefs,
			OpenDecisions: append([]string{}, summary.Uncertain...),
			MissingInputs: append([]string{}, summary.Missing...),
			TrustState:    semanticTrustState(summary),
			RenderHints:   []string{"openable", "artifact_shelf"},
		})
	}
	return artifacts
}

func semanticArtifactID(item catalogItem) string {
	id := normalizeFilename(item.ID)
	if id == "" {
		id = normalizeFilename(item.Label)
	}
	return id
}

func semanticTurnID(summary *workSummary, turn threadTurnRecord) string {
	parts := []string{}
	if summary != nil {
		parts = append(parts, summary.WorkUnitID, summary.Title)
	}
	parts = append(parts, turn.DoingNow, turn.UpNext)
	parts = append(parts, turn.ArtifactsCreated...)
	parts = append(parts, turn.ArtifactsUpdated...)
	id := slugify(strings.Join(parts, " "))
	if id == "" {
		return "current-turn"
	}
	if len(id) > 80 {
		id = strings.Trim(id[:80], "-")
	}
	return "turn-" + id
}

func normalizeSemanticEnvelope(envelope semanticEnvelope) semanticEnvelope {
	if envelope.Artifacts == nil {
		envelope.Artifacts = []semanticArtifact{}
	}
	if envelope.Decisions == nil {
		envelope.Decisions = []string{}
	}
	if envelope.Missing == nil {
		envelope.Missing = []string{}
	}
	if envelope.Uncertainty == nil {
		envelope.Uncertainty = []string{}
	}
	if envelope.RenderHints == nil {
		envelope.RenderHints = []renderRequest{}
	}
	envelope.TurnDelta.JustFinished = stringSliceOrEmpty(envelope.TurnDelta.JustFinished)
	envelope.TurnDelta.ArtifactsCreated = stringSliceOrEmpty(envelope.TurnDelta.ArtifactsCreated)
	envelope.TurnDelta.ArtifactsUpdated = stringSliceOrEmpty(envelope.TurnDelta.ArtifactsUpdated)
	for index := range envelope.Artifacts {
		envelope.Artifacts[index].SourceRefs = stringSliceOrEmpty(envelope.Artifacts[index].SourceRefs)
		envelope.Artifacts[index].OpenDecisions = stringSliceOrEmpty(envelope.Artifacts[index].OpenDecisions)
		envelope.Artifacts[index].MissingInputs = stringSliceOrEmpty(envelope.Artifacts[index].MissingInputs)
		envelope.Artifacts[index].RenderHints = stringSliceOrEmpty(envelope.Artifacts[index].RenderHints)
	}
	return envelope
}

func stringSliceOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func inputSourceRefs(inputs []inputItem) []string {
	refs := []string{}
	for _, item := range inputs {
		ref := strings.TrimSpace(item.InputID)
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

func semanticArtifactStatus(summary *workSummary) string {
	if summary != nil && len(summary.Missing) > 0 {
		return "draft"
	}
	return "ready"
}

func semanticTrustState(summary *workSummary) string {
	if summary == nil {
		return "unknown"
	}
	if len(summary.Missing) > 0 || len(summary.Uncertain) > 0 {
		return "draft_reviewable"
	}
	return "ready_reviewable"
}

func renderRiskLevel(envelope semanticEnvelope) string {
	if envelope.ConfirmationRequired || len(envelope.Missing) > 0 {
		return "medium"
	}
	return "low"
}

func routeVisibility(envelope semanticEnvelope) string {
	label := strings.ToLower(envelope.RouteSummary.Label + " " + envelope.RouteSummary.Reason)
	if strings.Contains(label, "fallback") || strings.Contains(label, "degraded") {
		return "expanded"
	}
	return "compact"
}

func renderActionsForMode(mode string) []string {
	switch mode {
	case "first_result":
		return []string{"Keep going", "Make it fuller", "Show what is missing", "Start new work"}
	case "work_summary", "multi_thread_home":
		return []string{"Continue current work", "Open what is ready", "Start new work"}
	case "ask":
		return []string{"Answer ask", "Skip for now", "Show what is ready"}
	case "recovery":
		return []string{"Show what is ready", "Start new work"}
	default:
		return []string{"Keep going", "Show what is missing", "Start new work"}
	}
}
