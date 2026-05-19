package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type threadTurnRecord struct {
	JustFinished     []string `json:"just_finished"`
	DoingNow         string   `json:"doing_now"`
	UpNext           string   `json:"up_next"`
	ArtifactsCreated []string `json:"artifacts_created"`
	ArtifactsUpdated []string `json:"artifacts_updated"`
}

type threadAsk struct {
	AskID                string   `json:"ask_id"`
	Prompt               string   `json:"prompt"`
	Reason               string   `json:"reason"`
	Options              []string `json:"options"`
	AssumptionsIfSkipped []string `json:"assumptions_if_skipped"`
	Blocking             bool     `json:"blocking"`
}

type savedThreadState struct {
	SchemaVersion string           `json:"schema_version"`
	ContextType   string           `json:"context_type"`
	CurrentTurn   threadTurnRecord `json:"current_turn"`
	ActiveAsk     *threadAsk       `json:"active_ask,omitempty"`
}

type inputItem struct {
	InputID   string `json:"input_id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Preview   string `json:"preview"`
	OriginRef string `json:"origin_ref"`
}

type savedInputs struct {
	SchemaVersion string      `json:"schema_version"`
	ContextType   string      `json:"context_type"`
	Items         []inputItem `json:"items"`
}

type workThread struct {
	ThreadID           string
	Goal               string
	WorkingWith        []string
	Now                string
	JustFinished       []string
	DoingNow           string
	UpNext             string
	Done               []string
	Need               string
	NeedReason         string
	NeedOptions        []string
	Assumptions        []string
	Next               string
	ReadyNow           []catalogItem
	Blocked            []string
	NotSureAbout       []string
	SafeToDo           string
	InputItems         []inputItem
	CurrentRoute       string
	ModelLabel         string
	ModelReason        string
	ModelFeedback      string
	RoutePolicy        string
	RouteReason        string
	ContinuityReason   string
	MultimodalLearning []string
	EffortLevel        string
	VerificationLevel  string
	VerificationReason string
}

func threadStatePath(workDir string) string {
	return filepath.Join(workDir, "thread-state.json")
}

func inputItemsPath(workDir string) string {
	return filepath.Join(workDir, "inputs.json")
}

func saveThreadState(workDir string, state savedThreadState) error {
	if workDir == "" {
		return nil
	}
	state.SchemaVersion = "0.1.0"
	state.ContextType = "JiniThreadState"
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(threadStatePath(workDir), append(data, '\n'), 0o600)
}

func loadThreadState(workDir string, summary *workSummary) savedThreadState {
	data, err := os.ReadFile(threadStatePath(workDir))
	if err == nil {
		var payload savedThreadState
		if json.Unmarshal(data, &payload) == nil {
			if payload.CurrentTurn.DoingNow != "" || payload.ActiveAsk != nil {
				return payload
			}
		}
	}
	return synthesizeThreadState(summary)
}

func saveInputItems(workDir string, items []inputItem) error {
	if workDir == "" {
		return nil
	}
	payload := savedInputs{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniInputItems",
		Items:         items,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(inputItemsPath(workDir), append(data, '\n'), 0o600)
}

func loadInputItems(workDir, packID string) []inputItem {
	data, err := os.ReadFile(inputItemsPath(workDir))
	if err != nil {
		return synthesizeInputItems(packID)
	}
	var payload savedInputs
	if err := json.Unmarshal(data, &payload); err != nil || len(payload.Items) == 0 {
		return synthesizeInputItems(packID)
	}
	return payload.Items
}

func synthesizeInputItems(packID string) []inputItem {
	return []inputItem{{
		InputID: "derived-context",
		Kind:    "derived",
		Title:   inferUsing(packID),
		Status:  "processed",
		Preview: inferUsing(packID),
	}}
}

func inputItemsForSource(source string) ([]inputItem, string) {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return []inputItem{{
			InputID: "request",
			Kind:    "text",
			Title:   "Your request",
			Status:  "processed",
		}}, ""
	}
	if item, normalized, ok := inputItemFromPath(trimmed); ok {
		return []inputItem{item}, normalized
	}
	return []inputItem{{
		InputID: "request",
		Kind:    "text",
		Title:   "Your request",
		Status:  "processed",
		Preview: compactPreview(trimmed, 120),
		OriginRef: trimmed,
	}}, trimmed
}

func inputItemFromPath(raw string) (inputItem, string, bool) {
	info, err := os.Stat(raw)
	if err != nil || info.IsDir() {
		return inputItem{}, "", false
	}

	title := filepath.Base(raw)
	kind := classifyInputKind(title)
	item := inputItem{
		InputID:   "attachment-1",
		Kind:      kind,
		Title:     title,
		Status:    "received",
		OriginRef: raw,
	}
	if text, ok := readTextAttachment(raw); ok {
		item.Status = "processed"
		item.Preview = compactPreview(text, 120)
		return item, text, true
	}
	item.Preview = "Attached " + kind
	return item, "Attachment: " + title, true
}

func classifyInputKind(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".heic":
		return "image"
	case ".mp3", ".m4a", ".wav", ".aac", ".ogg":
		return "audio"
	case ".txt", ".md", ".csv", ".json", ".yaml", ".yml":
		return "file"
	default:
		return "file"
	}
}

func readTextAttachment(path string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".csv", ".json", ".yaml", ".yml":
	default:
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", false
	}
	return text, true
}

func compactPreview(value string, limit int) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit < 4 {
		return value[:limit]
	}
	return strings.TrimSpace(value[:limit-3]) + "..."
}

func formatInputItem(item inputItem) string {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		title = "Working input"
	}
	preview := strings.TrimSpace(item.Preview)
	switch item.Kind {
	case "text":
		if preview != "" {
			return fmt.Sprintf("%s: %s", title, preview)
		}
	case "derived":
		return title
	default:
		if item.Status == "processed" {
			return title + " (processed)"
		}
		return title + " (attached)"
	}
	return title
}

func buildWorkThread(summary *workSummary, inputs []inputItem, state savedThreadState) workThread {
	workingWith := make([]string, 0, len(inputs))
	for _, item := range inputs {
		workingWith = append(workingWith, formatInputItem(item))
	}
	currentRoute := strings.TrimSpace(summary.WorkingWith)
	if currentRoute == "" {
		currentRoute = workingWithLabel(detectProvider())
	}
	need := inferNeed(summary.Missing)
	needReason := ""
	needOptions := []string{}
	assumptions := []string{}
	if state.ActiveAsk != nil {
		need = firstNonEmpty(strings.TrimSpace(state.ActiveAsk.Prompt), need)
		needReason = strings.TrimSpace(state.ActiveAsk.Reason)
		needOptions = append([]string{}, state.ActiveAsk.Options...)
		assumptions = append([]string{}, state.ActiveAsk.AssumptionsIfSkipped...)
	}
	justFinished := append([]string{}, state.CurrentTurn.JustFinished...)
	if len(justFinished) == 0 {
		justFinished = inferDone(summary.PackID, summary.Views)
	}
	doingNow := firstNonEmpty(strings.TrimSpace(state.CurrentTurn.DoingNow), summary.Doing)
	upNext := firstNonEmpty(strings.TrimSpace(state.CurrentTurn.UpNext), summary.NextStep)
	request := providerRequestForInputs(summary.PackID, summary.Title, inputs)
	return workThread{
		ThreadID:           firstNonEmpty(summary.WorkUnitID, slugify(summary.Title)),
		Goal:               summary.Title,
		WorkingWith:        workingWith,
		Now:                summary.Doing,
		JustFinished:       justFinished,
		DoingNow:           doingNow,
		UpNext:             upNext,
		Done:               inferDone(summary.PackID, summary.Views),
		Need:               need,
		NeedReason:         needReason,
		NeedOptions:        needOptions,
		Assumptions:        assumptions,
		Next:               summary.NextStep,
		ReadyNow:           summary.Views,
		Blocked:            summary.Missing,
		NotSureAbout:       summary.Uncertain,
		SafeToDo:           summary.SafeToDo,
		InputItems:         inputs,
		CurrentRoute:       currentRoute,
		ModelLabel:         summary.ModelLabel,
		ModelReason:        summary.ModelReason,
		ModelFeedback:      summary.ModelFeedback,
		RoutePolicy:        summary.RoutePolicy,
		RouteReason:        summary.RouteReason,
		ContinuityReason:   summary.ContinuityReason,
		MultimodalLearning: freshLocalMultimodalLearningViewLines(classifyRouteFeatures(request)),
		EffortLevel:        summary.EffortLevel,
		VerificationLevel:  summary.VerificationLevel,
		VerificationReason: summary.VerificationReason,
	}
}

func synthesizeThreadState(summary *workSummary) savedThreadState {
	state := savedThreadState{
		CurrentTurn: synthesizeTurnRecord(summary),
	}
	if ask := synthesizeThreadAsk(summary); ask != nil {
		state.ActiveAsk = ask
	}
	return state
}

func synthesizeTurnRecord(summary *workSummary) threadTurnRecord {
	record := threadTurnRecord{
		JustFinished: inferDone(summary.PackID, summary.Views),
		DoingNow:     summary.Doing,
		UpNext:       summary.NextStep,
	}
	for _, item := range summary.Views {
		record.ArtifactsCreated = append(record.ArtifactsCreated, item.Label)
	}
	return record
}

func synthesizeThreadAsk(summary *workSummary) *threadAsk {
	return starterAsk(summary, sourceFromInputItems(summary.Thread.InputItems))
}

func inferDone(packID string, views []catalogItem) []string {
	return starterDone(packID, views)
}

func inferNeed(missing []string) string {
	if len(missing) == 0 {
		return "Nothing right now"
	}
	return missing[0]
}
