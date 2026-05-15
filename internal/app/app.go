package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type currentWork struct {
	PackDir    string `json:"pack_dir"`
	PackID     string `json:"pack_id"`
	WorkUnitID string `json:"work_unit_id"`
	Title      string `json:"title"`
	State      string `json:"state"`
	Health     string `json:"health"`
}

type workSummary struct {
	Dir        string
	PackID     string
	WorkUnitID string
	Title      string
	State      string
	Views      []catalogItem
	Exports    []catalogItem
	Details    []catalogItem
	Missing    []string
	Uncertain  []string
	Using      string
	Doing      string
	Progress   string
	NextStep   string
	SafeToDo   string
}

type catalogItem struct {
	ID      string
	Label   string
	Path    string
	Aliases []string
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunInteractive(args, nil, stdout, stderr)
}

func RunInteractive(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	if len(args) == 0 {
		return runLauncher(stdin, stdout, stderr)
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:], stdout, stderr)
	case "open":
		return runOpen(args[1:], stdout, stderr)
	case "run":
		if len(args) > 1 && (args[1] == "--new" || args[1] == "new") {
			if stdin != nil {
				return runNewWorkIntake(stdin, stdout, stderr)
			}
			renderNewWorkLauncher(stdout)
			return 0
		}
		return runLauncher(stdin, stdout, stderr)
	case "provider":
		return runProvider(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
		fmt.Fprintln(stderr, "Try `jini`, `jini provider doctor`, or a scriptable command such as `jini check`.")
		return 1
	}
}

func runLauncher(stdin io.Reader, stdout, stderr io.Writer) int {
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		if stdin != nil {
			return runNewWorkIntake(stdin, stdout, stderr)
		}
		renderNewWorkLauncher(stdout)
		return 0
	}

	summary, err := loadWorkSummary(current.PackDir, current)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = clearCurrentWork()
			fmt.Fprintln(stdout, "Remembered work is no longer available.")
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			return 0
		}
		fmt.Fprintln(stderr, "Could not load current work. Run `jini` to start again or pass a valid work directory.")
		renderNewWorkLauncher(stdout)
		return 0
	}

	renderCurrentWorkLauncher(stdout, summary, stdin != nil)
	if stdin == nil {
		return 0
	}

	session := bufio.NewScanner(stdin)
	action, ok := readOptionalInputLine(session, stdout)
	if !ok || strings.TrimSpace(action) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	return handleCurrentWorkAction(action, summary, stdin, stdout, stderr)
}

func handleCurrentWorkAction(action string, summary *workSummary, stdin io.Reader, stdout, stderr io.Writer) int {
	switch normalizeName(action) {
	case "1", "continue", "continue current work", "keep going":
		renderCheck(stdout, summary)
	case "2", "open", "open ready work", "open ready", "open whats ready", "open what's ready", "open what is ready":
		renderOpenShelf(stdout, summary)
	case "plan this first", "plan first", "plan", "requirements", "design":
		renderPlanFirst(stdout, summary)
	case "3", "new", "start new", "start something new":
		if stdin == nil {
			renderNewWorkLauncher(stdout)
			return 0
		}
		return runNewWorkIntake(stdin, stdout, stderr)
	default:
		renderCheck(stdout, summary)
	}
	return 0
}

type starterChoice struct {
	PackID      string
	ChoiceLabel string
	DefaultName string
	State       string
}

type providerConfig struct {
	ID       string
	Label    string
	Status   string
	Missing  []string
	Settings []string
	Secrets  []string
}

func runProvider(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] != "doctor" {
		fmt.Fprintf(stderr, "Unknown provider command %q.\n", args[0])
		fmt.Fprintln(stderr, "Try `jini provider doctor`.")
		return 1
	}
	provider := detectProvider()
	renderProviderDoctor(stdout, provider)
	if provider.Status == "ok" {
		return 0
	}
	return 1
}

func detectProvider() providerConfig {
	switch configuredProviderMode() {
	case "local-preview":
		return detectLocalPreviewProvider()
	case "azure-openai":
		return detectAzureOpenAIProvider()
	case "bedrock":
		return detectBedrockProvider()
	case "anthropic":
		return detectAnthropicProvider()
	case "auto":
		return detectAutoProvider()
	default:
		providerID := normalizeName(configValue("JINI_PROVIDER"))
		return providerConfig{
			ID:      providerID,
			Label:   titleCase(providerID),
			Status:  "needs setup",
			Missing: []string{"Supported JINI_PROVIDER value: auto, claude, azure-openai, bedrock, or local-preview"},
		}
	}
}

func detectLocalPreviewProvider() providerConfig {
	return providerConfig{
		ID:     "local-preview",
		Label:  "Local preview",
		Status: "ok",
		Settings: []string{
			"JINI_PROVIDER: " + firstNonEmpty(configValue("JINI_PROVIDER"), "auto") + " -> Local preview",
		},
	}
}

func detectAzureOpenAIProvider() providerConfig {
	missing := missingEnvVars([]string{
		"AZURE_OPENAI_ENDPOINT",
		"AZURE_OPENAI_API_KEY",
		"AZURE_OPENAI_DEPLOYMENT",
	})
	label := "Azure OpenAI"
	if deployment := configValue("AZURE_OPENAI_DEPLOYMENT"); deployment != "" {
		label += " / " + deployment
	}
	settings := []string{
		providerSettingLine("azure-openai"),
		"AZURE_OPENAI_ENDPOINT: " + presentOrMissing("AZURE_OPENAI_ENDPOINT"),
		"AZURE_OPENAI_DEPLOYMENT: " + presentOrMissing("AZURE_OPENAI_DEPLOYMENT"),
		"AZURE_OPENAI_API_VERSION: " + valueOrDefault("AZURE_OPENAI_API_VERSION", "2024-10-21"),
	}
	if modelLine := azureModelSettingLine(); modelLine != "" {
		settings = append(settings, modelLine)
	}
	return providerConfig{
		ID:       "azure-openai",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"AZURE_OPENAI_API_KEY: " + presentOrMissing("AZURE_OPENAI_API_KEY")},
	}
}

func detectBedrockProvider() providerConfig {
	missing := []string{}
	if strings.TrimSpace(resolveAWSRegion()) == "" {
		missing = append(missing, "AWS_REGION or AWS_DEFAULT_REGION")
	}
	if configValue("AWS_PROFILE") == "" && (configValue("AWS_ACCESS_KEY_ID") == "" || configValue("AWS_SECRET_ACCESS_KEY") == "") {
		missing = append(missing, "AWS_PROFILE or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY")
	}
	modelID, modelLabel := resolveBedrockModel()
	label := "Amazon Bedrock"
	if modelLabel != "" {
		label += " / " + modelLabel
	}
	settings := []string{
		providerSettingLine("bedrock"),
		"AWS_REGION or profile region: " + presentOrMissingValue(resolveAWSRegion()),
		bedrockModelSettingLine(modelID, modelLabel),
	}
	return providerConfig{
		ID:       "bedrock",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets: []string{
			"AWS_PROFILE: " + presentOrMissing("AWS_PROFILE"),
			"AWS_ACCESS_KEY_ID: " + presentOrMissing("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY: " + presentOrMissing("AWS_SECRET_ACCESS_KEY"),
		},
	}
}

func detectAnthropicProvider() providerConfig {
	missing := []string{}
	if configValue("ANTHROPIC_API_KEY") == "" {
		missing = append(missing, "ANTHROPIC_API_KEY")
	}
	modelID, modelLabel, modelIssue := resolveAnthropicModel()
	if modelIssue != "" {
		missing = append(missing, modelIssue)
	}
	label := "Claude API"
	if modelLabel != "" {
		label += " / " + modelLabel
	}
	settings := []string{
		providerSettingLine("anthropic"),
		anthropicModelSettingLine(modelID, modelLabel),
		"ANTHROPIC_BASE_URL: " + presentOrMissing("ANTHROPIC_BASE_URL") + " (default https://api.anthropic.com)",
	}
	return providerConfig{
		ID:       "anthropic",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"ANTHROPIC_API_KEY: " + presentOrMissing("ANTHROPIC_API_KEY")},
	}
}

func missingEnvVars(names []string) []string {
	missing := []string{}
	for _, name := range names {
		if configValue(name) == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func statusFromMissing(missing []string) string {
	if len(missing) == 0 {
		return "ok"
	}
	return "needs setup"
}

func presentOrMissing(name string) string {
	if configValue(name) == "" {
		return "missing"
	}
	return "set"
}

func presentOrMissingEither(names ...string) string {
	for _, name := range names {
		if configValue(name) != "" {
			return "set"
		}
	}
	return "missing"
}

func presentOrMissingValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "missing"
	}
	return "set"
}

func valueOrDefault(name, fallback string) string {
	value := configValue(name)
	if value == "" {
		return fallback
	}
	return value
}

func renderProviderDoctor(w io.Writer, provider providerConfig) {
	fmt.Fprintln(w, "Provider")
	fmt.Fprintln(w, provider.Label)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Status")
	fmt.Fprintln(w, provider.Status)
	if len(provider.Settings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Settings")
		for _, item := range provider.Settings {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(provider.Secrets) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Secrets")
		for _, item := range provider.Secrets {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(provider.Missing) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Missing")
		for _, item := range provider.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
}

func runNewWorkIntake(stdin io.Reader, stdout, stderr io.Writer) int {
	renderNewWorkLauncher(stdout)
	session := bufio.NewScanner(stdin)

	for {
		firstRaw, ok := readInputLine(session, stdout)
		if !ok {
			return 0
		}
		if handled, exitCode := maybeHandleProviderSetupIntent(firstRaw, session, stdout, stderr); handled {
			if exitCode != 0 {
				fmt.Fprintln(stdout)
				renderNewWorkLauncher(stdout)
				continue
			}
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			continue
		}

		var source string
		choice, err := resolveStarterChoice(firstRaw)
		if err != nil {
			source = strings.TrimSpace(firstRaw)
			choice = classifyStarterChoice(source)
		} else {
			source, ok = readPromptLine(session, stdout, sourcePromptForChoice(choice))
			if !ok || strings.TrimSpace(source) == "" {
				fmt.Fprintln(stderr, "I need one line of source context to start this work.")
				return 1
			}
			if choice.PackID == "auto" {
				choice = classifyStarterChoice(source)
			}
		}
		if strings.TrimSpace(source) == "" {
			fmt.Fprintln(stderr, "I need one line of source context to start this work.")
			return 1
		}

		summary, err := bootstrapStarterWork(choice, source, "quick")
		if err != nil {
			fmt.Fprintf(stderr, "Could not start this work: %v\n", err)
			return 1
		}

		renderFirstRunResult(stdout, summary)
		action, ok := readOptionalInputLine(session, stdout)
		if !ok || strings.TrimSpace(action) == "" {
			fmt.Fprintln(stdout)
			renderCheck(stdout, summary)
			return 0
		}
		fmt.Fprintln(stdout)
		return handlePostResultAction(action, summary, stdout, stderr)
	}
}

func readInputLine(scanner *bufio.Scanner, stdout io.Writer) (string, bool) {
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func readOptionalInputLine(scanner *bufio.Scanner, stdout io.Writer) (string, bool) {
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func readPromptLine(scanner *bufio.Scanner, stdout io.Writer, prompt string) (string, bool) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, prompt)
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func resolveStarterChoice(raw string) (starterChoice, error) {
	choice := normalizeName(raw)
	switch choice {
	case "1", "meeting", "turn meeting notes into something i can send", "follow up after a meeting", "follow up", "meeting follow up", "meeting followup":
		return starterChoice{PackID: "meeting-followup", ChoiceLabel: "Turn meeting notes into something I can send", DefaultName: "Meeting Follow-up", State: "decided"}, nil
	case "2", "plan", "check whether a plan is ready to hand off", "check if a plan is ready", "spec", "spec readiness":
		return starterChoice{PackID: "research-prd", ChoiceLabel: "Check whether a plan is ready to hand off", DefaultName: "Plan Readiness", State: "awaiting_verification"}, nil
	case "3", "i am not sure", "i'm not sure", "i’m not sure", "im not sure", "i m not sure", "not sure", "unsure":
		return starterChoice{PackID: "auto", ChoiceLabel: "I am not sure", DefaultName: "First Useful Pass", State: "decided"}, nil
	case "plan this first", "plan first":
		return starterChoice{PackID: "auto", ChoiceLabel: "Plan this first", DefaultName: "Plan First", State: "modeled"}, nil
	case "compare options", "compare options and choose one", "vendor":
		return starterChoice{PackID: "vendor-selection", ChoiceLabel: "Compare options and choose one", DefaultName: "Option Review", State: "decided"}, nil
	case "4", "incident", "clean up an incident":
		return starterChoice{PackID: "incident-response", ChoiceLabel: "Clean up an incident", DefaultName: "Incident Cleanup", State: "incident"}, nil
	case "5", "trip", "plan a trip":
		return starterChoice{PackID: "travel-plan", ChoiceLabel: "Plan a trip", DefaultName: "Trip Plan", State: "decided"}, nil
	case "6", "something else", "something":
		return starterChoice{PackID: "general-work", ChoiceLabel: "Something else", DefaultName: "General Work", State: "decided"}, nil
	default:
		return starterChoice{}, fmt.Errorf("I couldn't match %q to a starter flow yet.", raw)
	}
}

func sourcePromptForChoice(choice starterChoice) string {
	if choice.PackID == "auto" {
		return strings.Join([]string{
			"Paste what you have. A rough version is fine.",
			"I will help figure out whether this is follow-up, a plan check, or something else.",
			"Nothing will be sent yet.",
		}, "\n")
	}
	return "Paste what you have. A rough version is fine."
}

func classifyStarterChoice(source string) starterChoice {
	normalized := normalizeName(source)
	switch {
	case containsAny(normalized, []string{"meeting", "follow up", "followup", "action items", "owners", "due dates", "open questions"}):
		return starterChoice{PackID: "meeting-followup", ChoiceLabel: "Turn meeting notes into something I can send", DefaultName: "Meeting Follow-up", State: "decided"}
	case containsAny(normalized, []string{"prd", "spec", "build readiness", "ready to hand off", "handoff", "hand off", "rollback", "implementation slice"}):
		return starterChoice{PackID: "research-prd", ChoiceLabel: "Check whether a plan is ready to hand off", DefaultName: "Plan Readiness", State: "awaiting_verification"}
	case containsAny(normalized, []string{"trip", "travel", "paris", "hotel", "flight", "itinerary"}):
		return starterChoice{PackID: "travel-plan", ChoiceLabel: "Plan a trip", DefaultName: "Trip Plan", State: "decided"}
	default:
		return starterChoice{PackID: "general-work", ChoiceLabel: "First Useful Pass", DefaultName: "First Useful Pass", State: "decided"}
	}
}

func bootstrapStarterWork(choice starterChoice, source, detail string) (*workSummary, error) {
	stateRoot := sessionStateRoot()
	workRoot := filepath.Join(stateRoot, "work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, err
	}

	title := deriveStarterTitle(choice.DefaultName, source)
	workDir, err := uniqueWorkDir(workRoot, choice.PackID, title)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(workDir, "artifacts"), 0o755); err != nil {
		return nil, err
	}

	if err := writeStarterWork(choice, workDir, title, source, detail); err != nil {
		return nil, err
	}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, title, source); err != nil {
		return nil, err
	}

	current := &currentWork{
		PackDir:    workDir,
		PackID:     choice.PackID,
		WorkUnitID: slugify(title),
		Title:      title,
		State:      choice.State,
		Health:     inferHealthFromState(choice.State),
	}
	if err := saveCurrentWork(current); err != nil {
		return nil, err
	}
	return loadWorkSummary(workDir, current)
}

func saveCurrentWork(current *currentWork) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(map[string]any{
		"schema_version": "0.1.0",
		"context_type":   "JiniCurrentWork",
		"pack_dir":       current.PackDir,
		"pack_id":        current.PackID,
		"work_unit_id":   current.WorkUnitID,
		"title":          current.Title,
		"state":          current.State,
		"health":         current.Health,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionStateRoot(), "current-work.json"), append(data, '\n'), 0o644)
}

func writeStarterWork(choice starterChoice, workDir, title, source, detail string) error {
	workUnit := strings.Join([]string{
		fmt.Sprintf("work_unit_id: %s", slugify(title)),
		fmt.Sprintf("title: %s", title),
		fmt.Sprintf("current_state: %s", choice.State),
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "work-unit.yaml"), []byte(workUnit), 0o644); err != nil {
		return err
	}

	switch choice.PackID {
	case "meeting-followup":
		return writeMeetingStarterWork(workDir, title, source, detail)
	case "research-prd":
		return writeResearchStarterWork(workDir, title, source, detail)
	case "travel-plan":
		return writeTravelStarterWork(workDir, title, source, detail)
	case "vendor-selection":
		return writeSimpleStarterWork(workDir, title, "Recommendation Memo", source, []string{
			"Top option",
			"Tradeoffs still to review",
			"Budget or approval boundary",
		})
	case "incident-response":
		return writeSimpleStarterWork(workDir, title, "Closure Checklist", source, []string{
			"Recovery proof",
			"Open follow-up owners",
			"Customer or leadership update status",
		})
	default:
		return writeFirstUsefulPassStarterWork(workDir, title, source)
	}
}

func writeMeetingStarterWork(workDir, title, source, detail string) error {
	followup := strings.Join([]string{
		fmt.Sprintf("# Sendable Follow-Up: %s", title),
		"",
		"## Send this",
		fmt.Sprintf("Here is the clean follow-up from **%s**.", title),
		"",
		fmt.Sprintf("Source context: %s.", source),
		"",
		"## What we agreed",
		"- Pricing draft moves to Sarah by Thursday so the launch page is no longer blocked on copy.",
		"- Landing page review stays with Amir and comments are due by Wednesday for design cleanup.",
		"- Analytics event coverage still needs Priya's metric decision before implementation starts.",
		"",
		"## Owners and due points",
		"- Sarah: draft the pricing update by Thursday.",
		"- Amir: land the pricing page review comments by Wednesday.",
		"- Priya: confirm the launch metric decision by Friday.",
		"",
		"## Open questions",
		"- Should launch success be measured by sign-ups or paid conversion?",
		"- Do we need legal review before publishing pricing changes?",
		"",
		"## What happens next",
		"- Send this note today so everyone works from the same decisions and due points.",
		"- Ask each owner to confirm date risk or dependency risk before the next standup.",
		"- Close the metric and legal-review questions before implementation starts.",
		"",
	}, "\n")
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(detail)), "f") {
		followup += strings.Join([]string{
			"## Why this note exists",
			"- Decisions, owners, and due dates must stay explicit before the meeting is considered closed.",
			"- Open questions should stay visible instead of getting buried in notes or chat.",
			"",
		}, "\n")
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "followup.md"), []byte(followup), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Task List\n\n- Sarah: draft the pricing update by Thursday.\n- Amir: land the pricing page review comments by Wednesday.\n- Priya: confirm the launch metric decision by Friday.\n"), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Tasks"})
}

func writeResearchStarterWork(workDir, title, source, detail string) error {
	prd := strings.Join([]string{
		fmt.Sprintf("# Build-Readiness Check: %s", title),
		"",
		"## Build-readiness check",
		"- Safe to start: scope is visible enough to plan the first implementation pass.",
		"- Still missing: approval, rollback note, and one clear decision on the highest-risk assumption.",
		"",
		"## Source summary",
		source,
		"",
		"## What to build first",
		"- Capture the first thin slice that proves the user value.",
		"- Keep the approval owner attached before implementation starts.",
		"- Record the rollback note in the same surface as the plan.",
		"",
	}, "\n")
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(detail)), "f") {
		prd += strings.Join([]string{
			"## Risks to clear",
			"- Product approval may still be implicit rather than recorded.",
			"- Rollback behavior is still under-specified.",
			"- The highest-risk assumption needs one owner before build starts.",
			"",
		}, "\n")
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "prd.md"), []byte(prd), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Missing Pieces Before Build\n\n- Confirm approval owner.\n- Add rollback note.\n- Lock the first implementation slice.\n"), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Plan", "Tasks", "Evidence"})
}

func writeTravelStarterWork(workDir, title, source, detail string) error {
	lines := []string{
		fmt.Sprintf("# Itinerary: %s", title),
		"",
		"## Trip at a glance",
		source,
		"",
		"This draft gives you a usable week shape first, then shows the booking, budget, and contingency items to lock next.",
		"",
		"## Day-by-day draft",
	}
	for _, item := range starterTripDays(title) {
		lines = append(lines, item)
	}
	lines = append(lines,
		"",
		"## Budget sketch",
	)
	for _, item := range starterTripBudget(title) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## Logistics to lock",
	)
	for _, item := range starterTripLogistics(title) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## If something changes",
	)
	for _, item := range starterTripContingencies(title) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## Still to confirm",
		"- Dates, budget, and hotel area",
		"- Whether Versailles is a must-do day trip or an optional slot",
		"",
	)
	if err := os.WriteFile(filepath.Join(workDir, "views", "itinerary.md"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "budget-sketch.md"), []byte("# Budget Sketch\n\n"+bulletLines(starterTripBudget(title))), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "travel-logistics.md"), []byte("# Travel Logistics\n\n"+bulletLines(starterTripLogistics(title))), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Task List\n\n- Confirm dates and budget.\n- Choose the hotel area.\n- Decide whether Versailles is locked or optional.\n"), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Tasks"})
}

func writeFirstUsefulPassStarterWork(workDir, title, source string) error {
	pass := strings.Join([]string{
		fmt.Sprintf("# First Useful Pass: %s", title),
		"",
		"## What this seems to be",
		fmt.Sprintf("- %s", source),
		"",
		"## What can be used now",
		"- A named work record has been started so the context is not lost.",
		"- The next pass can turn this into follow-up, a plan check, a decision memo, or another concrete output.",
		"",
		"## What I need next",
		"- The audience or recipient.",
		"- The outcome you want after someone reads or uses this.",
		"- Any deadline, owner, blocker, or decision that should not be guessed.",
		"",
		"## Safe right now",
		"- Nothing has been sent, changed, booked, or committed.",
		"- You can review this pass before sharing or turning it into a fuller artifact.",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "views", "first-useful-pass.md"), []byte(pass), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# What I Need Next\n\n- Name the audience or recipient.\n- Confirm the desired outcome.\n- Add any deadline, owner, blocker, or decision.\n"), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Tasks"})
}

func writeSimpleStarterWork(workDir, title, viewLabel, source string, bullets []string) error {
	var lines []string
	lines = append(lines, fmt.Sprintf("# %s: %s", viewLabel, title), "", source, "")
	for _, item := range bullets {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "")
	if err := os.WriteFile(filepath.Join(workDir, "views", normalizeFilename(viewLabel)+".md"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Task List\n\n- Review what is ready.\n- Confirm what is still missing.\n"), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Tasks"})
}

func writeStarterArtifacts(workDir string, artifactTypes []string) error {
	for index, artifactType := range artifactTypes {
		filename := fmt.Sprintf("%02d-%s.yaml", index+1, normalizeFilename(artifactType))
		content := fmt.Sprintf("artifact_type: %s\nstatus: ready\n", artifactType)
		if err := os.WriteFile(filepath.Join(workDir, "artifacts", filename), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func starterTripDays(title string) []string {
	if strings.Contains(strings.ToLower(title), "paris") {
		return []string{
			"### Day 1: Arrive and settle into Paris",
			"- Keep arrival day light: hotel check-in, neighborhood walk, easy dinner, early night.",
			"### Day 2: Louvre, Tuileries, and the Seine",
			"- Anchor the day around one major museum, then keep the evening for a Seine walk or river cruise.",
			"### Day 3: Ile de la Cite and the Latin Quarter",
			"- Pair Sainte-Chapelle or Notre-Dame area time with a slower Left Bank afternoon and cafe stop.",
			"### Day 4: Montmartre and Sacre-Coeur",
			"- Use the morning for Montmartre before crowds build, then keep the afternoon flexible for shopping or rest.",
			"### Day 5: Versailles or a second museum day",
			"- If energy is high, use this as the day trip. If not, keep it in Paris with Musee d'Orsay and the Left Bank.",
			"### Day 6: Le Marais and flexible favorites",
			"- Revisit the neighborhood you liked most, leave space for food, markets, or anything skipped earlier.",
			"### Day 7: Buffer and departure",
			"- Keep the final day intentionally light so checkout, bags, and airport transfer do not turn into stress.",
		}
	}
	return []string{
		"### Day 1: Arrive and settle in",
		"- Keep the first day light and focus on arrival, check-in, and one easy local activity.",
		"### Day 2: First major anchor",
		"- Use one major sight or neighborhood as the headline and keep the evening unstacked.",
		"### Day 3: Local exploration",
		"- Build around a second neighborhood or cultural stop and protect time for rest or weather shifts.",
	}
}

func starterTripBudget(title string) []string {
	if strings.Contains(strings.ToLower(title), "paris") {
		return []string{
			"Lodging: prioritize location over room size; central neighborhoods usually save time and transit cost.",
			"Food: plan one stronger meal per day and keep breakfast/lunch simple to protect the total budget.",
			"Transit: budget for airport transfer plus a metro pass; Paris is easier when local transit is decided early.",
			"Tickets: reserve room for at least two paid anchors such as Louvre, Musee d'Orsay, or Versailles.",
		}
	}
	return []string{
		"Lodging: choose the base first because it shapes daily transit and fatigue.",
		"Food: separate special meals from routine meals so the budget stays honest.",
		"Transit: include both arrival/departure transfer and local movement.",
	}
}

func starterTripLogistics(title string) []string {
	if strings.Contains(strings.ToLower(title), "paris") {
		return []string{
			"Choose the hotel area before booking tickets; central Paris usually beats a cheaper far-out stay.",
			"Lock airport transfer logic early: RER, taxi, or pre-booked car depending on arrival time and luggage.",
			"Reserve high-demand museum or Versailles slots before filling the rest of the week.",
		}
	}
	return []string{
		"Pick the base area before overcommitting the daily plan.",
		"Lock arrival and departure transfer details before adding optional activities.",
	}
}

func starterTripContingencies(title string) []string {
	if strings.Contains(strings.ToLower(title), "paris") {
		return []string{
			"Swap outdoor time for museums, passages, or food halls if weather turns.",
			"If a headline sight is sold out, use the day for a neighborhood loop instead of forcing a bad backup.",
			"Protect one slower day in the middle of the trip so fatigue does not wreck the last half.",
		}
	}
	return []string{
		"Have one indoor backup for every outdoor-heavy day.",
		"Leave one uncommitted slot so a sold-out booking does not ruin the week.",
	}
}

func bulletLines(items []string) string {
	lines := make([]string, 0, len(items)+1)
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func deriveStarterTitle(defaultName, source string) string {
	cleaned := titleCase(cleanForTitle(source))
	if strings.TrimSpace(cleaned) == "" {
		return defaultName
	}
	return cleaned
}

func cleanForTitle(value string) string {
	var builder strings.Builder
	spacePending := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if spacePending && builder.Len() > 0 {
				builder.WriteRune(' ')
			}
			builder.WriteRune(r)
			spacePending = false
		default:
			spacePending = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func normalizeFilename(value string) string {
	return strings.ReplaceAll(normalizeName(value), " ", "-")
}

func slugify(value string) string {
	slug := normalizeFilename(cleanForTitle(value))
	if slug == "" {
		return "jini-work"
	}
	return slug
}

func uniqueWorkDir(root, packID, title string) (string, error) {
	base := filepath.Join(root, fmt.Sprintf("%s-%s", packID, slugify(title)))
	candidate := base
	for index := 1; ; index++ {
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, index)
		if index > 1000 {
			return "", fmt.Errorf("could not allocate work directory under %s", root)
		}
	}
}

func inferHealthFromState(state string) string {
	switch state {
	case "awaiting_verification":
		return "ready-to-verify"
	case "incident":
		return "active"
	default:
		return "ready-to-make"
	}
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	summary, err := resolveSummary(args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	renderCheck(stdout, summary)
	return 0
}

func runOpen(args []string, stdout, stderr io.Writer) int {
	summary, err := resolveSummary(nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if len(args) == 0 {
		renderOpenShelf(stdout, summary)
		return 0
	}

	item, err := resolveOpenItem(summary, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	content, err := os.ReadFile(item.Path)
	if err != nil {
		fmt.Fprintf(stderr, "Could not read %q: %v\n", item.Label, err)
		return 1
	}
	fmt.Fprint(stdout, string(content))
	if len(content) == 0 || content[len(content)-1] != '\n' {
		fmt.Fprintln(stdout)
	}
	return 0
}

func resolveSummary(args []string) (*workSummary, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return loadWorkSummary(args[0], nil)
	}
	current, err := loadCurrentWork()
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("Nothing is in progress yet. Run `jini` to start something.")
	}
	summary, err := loadWorkSummary(current.PackDir, current)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = clearCurrentWork()
			return nil, errors.New("Remembered work is no longer available. Run `jini` to start something.")
		}
		return nil, errors.New("Could not load current work. Run `jini` to start again or pass a valid work directory.")
	}
	return summary, nil
}

func loadCurrentWork() (*currentWork, error) {
	path := filepath.Join(sessionStateRoot(), "current-work.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var payload currentWork
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.PackDir) == "" {
		return nil, nil
	}
	return &payload, nil
}

func sessionStateRoot() string {
	if override := strings.TrimSpace(os.Getenv("JINI_STATE_DIR")); override != "" {
		return override
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ".jini"
	}
	return filepath.Join(cwd, ".jini")
}

func clearCurrentWork() error {
	err := os.Remove(filepath.Join(sessionStateRoot(), "current-work.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func loadWorkSummary(dir string, current *currentWork) (*workSummary, error) {
	resolved, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a work directory", resolved)
	}

	workUnitPath := filepath.Join(resolved, "work-unit.yaml")
	workUnit, _ := parseSimpleYAML(workUnitPath)

	packID := strings.TrimSpace(workUnit["pack_id"])
	if packID == "" && current != nil {
		packID = current.PackID
	}
	if packID == "" {
		packID = inferPackID(resolved)
	}

	title := strings.TrimSpace(workUnit["title"])
	if title == "" && current != nil {
		title = current.Title
	}
	if title == "" {
		title = filepath.Base(resolved)
	}

	state := strings.TrimSpace(workUnit["current_state"])
	if state == "" && current != nil {
		state = current.State
	}

	views := collectViews(resolved, packID)
	exports := collectExports(resolved)
	details := collectDetails(resolved)
	missing := inferMissing(state, details, packID)
	uncertain := inferUncertain(packID, missing)

	summary := &workSummary{
		Dir:    resolved,
		PackID: packID,
		WorkUnitID: firstNonEmpty(strings.TrimSpace(workUnit["work_unit_id"]), func() string {
			if current != nil {
				return current.WorkUnitID
			}
			return ""
		}()),
		Title:     title,
		State:     state,
		Views:     views,
		Exports:   exports,
		Details:   details,
		Missing:   missing,
		Uncertain: uncertain,
		Using:     inferUsing(packID),
		Doing:     inferDoing(packID, state),
		Progress:  inferProgress(state),
		SafeToDo:  "Nothing has been sent yet. You can review before sharing.",
	}
	summary.NextStep = inferNextStep(packID, summary.Views)
	return summary, nil
}

func parseSimpleYAML(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}
	out := map[string]string{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		out[key] = value
	}
	return out, nil
}

func inferPackID(dir string) string {
	base := filepath.Base(dir)
	base = strings.TrimPrefix(base, "jini-sim-")
	switch {
	case strings.Contains(base, "prd"):
		return "research-prd"
	case strings.Contains(base, "meeting"):
		return "meeting-followup"
	case strings.Contains(base, "vendor"):
		return "vendor-selection"
	case strings.Contains(base, "incident"):
		return "incident-response"
	case strings.Contains(base, "trip") || strings.Contains(base, "travel") || strings.Contains(base, "paris"):
		return "travel-plan"
	default:
		return base
	}
}

func collectViews(root, packID string) []catalogItem {
	viewDir := filepath.Join(root, "views")
	entries, err := os.ReadDir(viewDir)
	if err != nil {
		return synthesizeViews(root, packID)
	}

	items := []catalogItem{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		stem := strings.TrimSuffix(entry.Name(), ".md")
		item := viewCatalogItem(packID, stem, filepath.Join(viewDir, entry.Name()))
		items = append(items, item)
	}
	items = append(items, synthesizedPackViews(root, packID)...)
	return prioritizeViews(packID, dedupeItems(items))
}

func synthesizeViews(root, packID string) []catalogItem {
	return prioritizeViews(packID, synthesizedPackViews(root, packID))
}

func synthesizedPackViews(root, packID string) []catalogItem {
	switch packID {
	case "research-prd":
		prdPath := filepath.Join(root, "views", "prd.md")
		if fileExists(prdPath) {
			return []catalogItem{
				{
					ID:      "build-readiness-check",
					Label:   "Build-Readiness Check",
					Path:    prdPath,
					Aliases: []string{"readiness", "build readiness check", "check"},
				},
			}
		}
	case "meeting-followup":
		followupPath := filepath.Join(root, "views", "followup.md")
		if fileExists(followupPath) {
			return []catalogItem{
				{
					ID:      "sendable-follow-up",
					Label:   "Sendable Follow-up",
					Path:    followupPath,
					Aliases: []string{"follow-up", "followup", "summary"},
				},
			}
		}
	}
	return nil
}

func viewCatalogItem(packID, stem, path string) catalogItem {
	switch stem {
	case "prd":
		return catalogItem{ID: "handoff-brief", Label: "Handoff Brief", Path: path, Aliases: []string{"prd", "summary", "brief", "handoff"}}
	case "followup":
		return catalogItem{ID: "sendable-follow-up", Label: "Sendable Follow-up", Path: path, Aliases: []string{"follow-up", "followup", "summary"}}
	case "tasks":
		switch packID {
		case "meeting-followup":
			return catalogItem{ID: "owners-and-due-points", Label: "Owners and Due Points", Path: path, Aliases: []string{"tasks", "task list", "owners"}}
		case "research-prd":
			return catalogItem{ID: "missing-pieces-before-build", Label: "Missing Pieces Before Build", Path: path, Aliases: []string{"tasks", "task list", "missing"}}
		case "travel-plan":
			return catalogItem{ID: "still-to-book", Label: "Still To Book", Path: path, Aliases: []string{"tasks", "task list", "booking"}}
		default:
			return catalogItem{ID: "next-actions", Label: "Next Actions", Path: path, Aliases: []string{"tasks", "task list", "actions"}}
		}
	case "selection":
		return catalogItem{ID: "recommendation-memo", Label: "Recommendation Memo", Path: path, Aliases: []string{"selection", "memo"}}
	case "response":
		return catalogItem{ID: "closure-checklist", Label: "Closure Checklist", Path: path, Aliases: []string{"response", "checklist"}}
	case "first-useful-pass":
		return catalogItem{ID: "first-useful-pass", Label: "First Useful Pass", Path: path, Aliases: []string{"first pass", "useful pass", "summary"}}
	default:
		return catalogItem{
			ID:      normalizeName(stem),
			Label:   titleCase(stem),
			Path:    path,
			Aliases: []string{stem},
		}
	}
}

func collectExports(root string) []catalogItem {
	candidates := []catalogItem{
		{ID: "markdown-wiki", Label: "Markdown Wiki", Path: filepath.Join(root, "exports", "wiki", "markdown", "overview.md"), Aliases: []string{"markdown", "wiki"}},
		{ID: "github-issues", Label: "GitHub Issues", Path: filepath.Join(root, "exports", "issues", "github", "README.md"), Aliases: []string{"github"}},
		{ID: "jira-issues", Label: "Jira Issues", Path: filepath.Join(root, "exports", "issues", "jira", "README.md"), Aliases: []string{"jira"}},
	}
	items := []catalogItem{}
	for _, item := range candidates {
		if fileExists(item.Path) {
			items = append(items, item)
		}
	}
	return items
}

func collectDetails(root string) []catalogItem {
	artifactDir := filepath.Join(root, "artifacts")
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return nil
	}

	items := []catalogItem{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(artifactDir, entry.Name())
		doc, _ := parseSimpleYAML(path)
		artifactType := firstNonEmpty(doc["artifact_type"], titleCase(strings.TrimSuffix(entry.Name(), ".yaml")))
		label := artifactType
		id := normalizeName(artifactType)
		if artifactType == "Tasks" {
			label = "Tasks Record"
			id = "tasks-record"
		}
		items = append(items, catalogItem{
			ID:      id,
			Label:   label,
			Path:    path,
			Aliases: []string{id, strings.TrimSuffix(entry.Name(), ".yaml")},
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

func inferMissing(state string, details []catalogItem, packID string) []string {
	available := map[string]bool{}
	for _, item := range details {
		available[item.Label] = true
	}
	missing := []string{}
	switch state {
	case "awaiting_verification", "operational", "incident":
		if !available["Approval"] {
			missing = append(missing, "Approval")
		}
		if state != "awaiting_verification" && !available["Evidence"] {
			missing = append(missing, "Evidence")
		}
	case "decided", "in_make":
		if packID == "meeting-followup" {
			missing = append(missing, "Metric and legal-review decision")
		}
		if packID == "travel-plan" {
			missing = append(missing, "Dates, budget, and hotel area")
		}
	}
	return missing
}

func inferUncertain(packID string, missing []string) []string {
	if len(missing) == 0 {
		return nil
	}
	switch packID {
	case "research-prd":
		return []string{"Whether approval was already granted in the review thread"}
	case "meeting-followup":
		return []string{"Whether the metric decision also needs legal review"}
	case "travel-plan":
		return []string{"Whether Versailles is a must-do day trip or an optional slot"}
	default:
		return []string{"Whether the missing items already exist outside this work record"}
	}
}

func inferUsing(packID string) string {
	switch packID {
	case "research-prd":
		return "Latest PRD draft and review comments"
	case "meeting-followup":
		return "Meeting notes and follow-up tasks"
	case "vendor-selection":
		return "Vendor notes, tradeoffs, and decision criteria"
	case "incident-response":
		return "Incident notes, timeline, and follow-up tasks"
	case "travel-plan":
		return "Trip notes, dates, and planning details"
	default:
		return "The files and notes in this work"
	}
}

func inferDoing(packID, state string) string {
	switch state {
	case "awaiting_verification":
		return "Checking assumptions and approval gaps"
	case "decided":
		if packID == "meeting-followup" {
			return "Turning notes into owners and next steps"
		}
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

func inferProgress(state string) string {
	switch state {
	case "awaiting_verification", "incident":
		return "3 of 4 steps done"
	case "decided", "in_make":
		return "2 of 4 steps done"
	case "modeled", "probed":
		return "1 of 4 steps done"
	case "operational", "retired":
		return "4 of 4 steps done"
	default:
		return "1 of 4 steps done"
	}
}

func inferNextStep(packID string, views []catalogItem) string {
	switch packID {
	case "research-prd":
		return "Open Build-Readiness Check"
	case "meeting-followup":
		return "Open Sendable Follow-up"
	case "travel-plan":
		return "Open Itinerary"
	default:
		if len(views) > 0 {
			return "Open " + views[0].Label
		}
		return "Review what is ready"
	}
}

func prioritizeViews(packID string, items []catalogItem) []catalogItem {
	if len(items) == 0 {
		return items
	}
	order := map[string]int{}
	switch packID {
	case "research-prd":
		order = map[string]int{
			"build-readiness-check":       0,
			"handoff-brief":               1,
			"missing-pieces-before-build": 2,
		}
	case "meeting-followup":
		order = map[string]int{
			"sendable-follow-up":    0,
			"owners-and-due-points": 1,
		}
	case "travel-plan":
		order = map[string]int{
			"itinerary":        0,
			"budget-sketch":    1,
			"travel-logistics": 2,
			"still-to-book":    3,
		}
	default:
		return items
	}

	sort.SliceStable(items, func(i, j int) bool {
		left, leftOK := order[items[i].ID]
		right, rightOK := order[items[j].ID]
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

func resolveOpenItem(summary *workSummary, name string) (*catalogItem, error) {
	normalized := normalizeName(name)
	for _, item := range append(append([]catalogItem{}, summary.Views...), append(summary.Exports, summary.Details...)...) {
		candidates := append([]string{item.ID, item.Label}, item.Aliases...)
		for _, candidate := range candidates {
			if normalizeName(candidate) == normalized {
				return &item, nil
			}
		}
	}
	return nil, fmt.Errorf("I couldn't find %q. Use `jini open` to see what is ready.", name)
}

func renderNewWorkLauncher(w io.Writer) {
	fmt.Fprintln(w, "What do you need help finishing?")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini shell")
	fmt.Fprintln(w, "Paste messy notes, or type the outcome you want.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	provider := detectProvider()
	fmt.Fprintln(w, provider.Label)
	if provider.ID == "local-preview" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Want a connected provider instead?")
		fmt.Fprintln(w, "- Type `Use Claude`")
		fmt.Fprintln(w, "- Type `Use Bedrock`")
		fmt.Fprintln(w, "- Type `Use Azure`")
		fmt.Fprintln(w, "- Type `Use Auto`")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Good inputs:")
	fmt.Fprintln(w, "- Turn meeting notes into something I can send")
	fmt.Fprintln(w, "- Check whether a plan is ready to hand off")
	fmt.Fprintln(w, "- Plan this first")
	fmt.Fprintln(w, "- I am not sure")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Nothing will be sent yet.")
}

func maybeHandleProviderSetupIntent(raw string, scanner *bufio.Scanner, stdout, stderr io.Writer) (bool, int) {
	switch normalizeName(raw) {
	case "use claude", "connect claude", "claude", "use anthropic", "anthropic":
		return true, runProviderSetupWizard("claude", scanner, stdout, stderr)
	case "use bedrock", "connect bedrock", "bedrock", "amazon bedrock":
		return true, runProviderSetupWizard("bedrock", scanner, stdout, stderr)
	case "use azure", "connect azure", "azure", "azure openai", "azure open ai":
		return true, runProviderSetupWizard("azure-openai", scanner, stdout, stderr)
	case "use auto", "auto", "choose automatically":
		return true, runProviderSetupWizard("auto", scanner, stdout, stderr)
	case "use local preview", "local preview", "preview", "work offline":
		return true, runProviderSetupWizard("local-preview", scanner, stdout, stderr)
	default:
		return false, 0
	}
}

func runProviderSetupWizard(mode string, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	fmt.Fprintln(stdout)
	switch mode {
	case "claude":
		fmt.Fprintln(stdout, "Connect Claude")
		fmt.Fprintln(stdout, "Paste your Anthropic API key. Jini will save it only in this repo's .jini folder.")
		key, ok := readPromptLine(scanner, stdout, "Anthropic API key")
		if !ok || strings.TrimSpace(key) == "" {
			fmt.Fprintln(stderr, "Claude setup needs an API key.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet)")
		if !ok {
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":     "claude",
			"ANTHROPIC_API_KEY": key,
			"JINI_MODEL":        firstNonEmpty(model, "sonnet"),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Claude setup: %v\n", err)
			return 1
		}
	case "bedrock":
		fmt.Fprintln(stdout, "Connect Amazon Bedrock")
		region, ok := readPromptLine(scanner, stdout, "AWS region (press Enter for us-east-1)")
		if !ok {
			return 1
		}
		profile, ok := readPromptLine(scanner, stdout, "AWS profile name")
		if !ok || strings.TrimSpace(profile) == "" {
			fmt.Fprintln(stderr, "Bedrock setup needs an AWS profile name.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet-4.6)")
		if !ok {
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "bedrock",
			"AWS_REGION":    firstNonEmpty(region, "us-east-1"),
			"AWS_PROFILE":   profile,
			"JINI_MODEL":    firstNonEmpty(model, "sonnet-4.6"),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock setup: %v\n", err)
			return 1
		}
	case "azure-openai":
		fmt.Fprintln(stdout, "Connect Azure OpenAI")
		endpoint, ok := readPromptLine(scanner, stdout, "Azure endpoint")
		if !ok || strings.TrimSpace(endpoint) == "" {
			fmt.Fprintln(stderr, "Azure setup needs an endpoint.")
			return 1
		}
		apiKey, ok := readPromptLine(scanner, stdout, "Azure API key")
		if !ok || strings.TrimSpace(apiKey) == "" {
			fmt.Fprintln(stderr, "Azure setup needs an API key.")
			return 1
		}
		deployment, ok := readPromptLine(scanner, stdout, "Azure deployment name")
		if !ok || strings.TrimSpace(deployment) == "" {
			fmt.Fprintln(stderr, "Azure setup needs a deployment name.")
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":            "azure-openai",
			"AZURE_OPENAI_ENDPOINT":    endpoint,
			"AZURE_OPENAI_API_KEY":     apiKey,
			"AZURE_OPENAI_DEPLOYMENT":  deployment,
			"AZURE_OPENAI_API_VERSION": "2024-10-21",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Azure setup: %v\n", err)
			return 1
		}
	case "auto":
		fmt.Fprintln(stdout, "Use auto mode")
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "auto",
			"JINI_MODEL":    "auto",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save auto mode: %v\n", err)
			return 1
		}
	case "local-preview":
		fmt.Fprintln(stdout, "Use local preview")
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "local-preview",
			"JINI_MODEL":    "auto",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save local preview mode: %v\n", err)
			return 1
		}
	}

	provider := detectProvider()
	fmt.Fprintln(stdout)
	if provider.Status == "ok" {
		fmt.Fprintf(stdout, "Setup saved. Working with %s.\n", provider.Label)
		return 0
	}
	fmt.Fprintf(stderr, "Setup is still incomplete. Missing: %s\n", strings.Join(provider.Missing, ", "))
	return 1
}

func renderFirstRunResult(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Your first draft is ready.")
	fmt.Fprintln(w)

	item := firstResultItem(summary)
	if item == nil {
		fmt.Fprintln(w, "First Useful Pass")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No result file is ready yet. Jini still created the work record so the source context is not lost.")
		renderPostResultActions(w, summary, nil)
		return
	}

	fmt.Fprintln(w, item.Label)
	fmt.Fprintln(w)
	content, err := os.ReadFile(item.Path)
	if err != nil {
		fmt.Fprintf(w, "Could not read %q yet.\n", item.Label)
	} else {
		fmt.Fprint(w, strings.TrimSpace(string(content)))
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	renderPostResultActions(w, summary, item)
}

func firstResultItem(summary *workSummary) *catalogItem {
	if len(summary.Views) == 0 {
		return nil
	}
	return &summary.Views[0]
}

func renderPostResultActions(w io.Writer, summary *workSummary, item *catalogItem) {
	fmt.Fprintln(w, "What do you want to do next?")
	fmt.Fprintln(w, "- Keep going")
	fmt.Fprintln(w, "- Make it fuller")
	fmt.Fprintln(w, "- See what is still missing")
	fmt.Fprintln(w, "- Plan this first")
	fmt.Fprintln(w, "- Start something new")
}

func handlePostResultAction(action string, summary *workSummary, stdout, stderr io.Writer) int {
	switch normalizeName(action) {
	case "keep going", "1", "continue":
		item := nextUsefulItem(summary)
		if item == nil {
			renderCheck(stdout, summary)
			return 0
		}
		renderItem(stdout, item)
	case "make it fuller", "full", "full version", "2":
		renderFullerPrompt(stdout, summary)
	case "see what is still missing", "missing", "check", "3":
		renderMissingOnly(stdout, summary)
	case "plan this first", "plan first", "plan", "4":
		renderPlanFirst(stdout, summary)
	case "start something new", "new", "5":
		renderNewWorkLauncher(stdout)
	default:
		renderCheck(stdout, summary)
	}
	return 0
}

func nextUsefulItem(summary *workSummary) *catalogItem {
	preferred := []string{"owners-and-due-points", "missing-pieces-before-build", "next-actions", "task-list"}
	for _, id := range preferred {
		if item := findViewByID(summary, id); item != nil {
			return item
		}
	}
	if len(summary.Views) == 0 {
		return nil
	}
	return &summary.Views[0]
}

func findViewByID(summary *workSummary, id string) *catalogItem {
	for index := range summary.Views {
		if summary.Views[index].ID == id {
			return &summary.Views[index]
		}
	}
	return nil
}

func renderItem(w io.Writer, item *catalogItem) {
	fmt.Fprintln(w, item.Label)
	fmt.Fprintln(w)
	content, err := os.ReadFile(item.Path)
	if err != nil {
		fmt.Fprintf(w, "Could not read %q yet.\n", item.Label)
		return
	}
	fmt.Fprint(w, strings.TrimSpace(string(content)))
	fmt.Fprintln(w)
}

func renderFullerPrompt(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Make it fuller")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Add any missing audience, owner, deadline, budget, approval, or blocker detail. Then Jini can turn this into a stronger next draft.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Views {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
}

func renderMissingOnly(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Still missing")
	if len(summary.Missing) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Uncertain) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Uncertain {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next step")
	fmt.Fprintln(w, summary.NextStep)
}

func renderPlanFirst(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Plan this first")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Goal")
	fmt.Fprintf(w, "- Finish %s without hiding what is missing.\n", summary.Title)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Requirements")
	fmt.Fprintln(w, "- Keep the useful output visible.")
	fmt.Fprintln(w, "- Keep missing proof, approval, or owner gaps visible.")
	fmt.Fprintln(w, "- Do not send, change, book, or commit anything without review.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Design")
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- No ready surface exists yet; create the first useful pass before deeper execution.")
	} else {
		for _, item := range summary.Views {
			fmt.Fprintf(w, "- Use %s as an openable work surface.\n", item.Label)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Steps")
	fmt.Fprintln(w, "- Review what is ready.")
	fmt.Fprintln(w, "- Clear the highest-risk missing item.")
	fmt.Fprintln(w, "- Then continue only if the next action is safe.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run")
	fmt.Fprintln(w, "- Keep going")
	fmt.Fprintln(w, "- Open ready work")
	fmt.Fprintln(w, "- See what is still missing")
}

func renderCurrentWorkLauncher(w io.Writer, summary *workSummary, interactive bool) {
	fmt.Fprintln(w, "You're working on")
	fmt.Fprintln(w, summary.Title)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	fmt.Fprintln(w, detectProvider().Label)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Views {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Still missing")
	if len(summary.Missing) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next step")
	fmt.Fprintln(w, summary.NextStep)
	fmt.Fprintln(w)
	if !interactive {
		return
	}
	fmt.Fprintln(w, "Jini shell")
	fmt.Fprintln(w, "What do you want to do?")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "- Keep going")
	fmt.Fprintln(w, "- Open ready work")
	fmt.Fprintln(w, "- See what is still missing")
	fmt.Fprintln(w, "- Plan this first")
	fmt.Fprintln(w, "- Start something else")
}

func renderCheck(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "You're working on")
	fmt.Fprintln(w, summary.Title)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	fmt.Fprintln(w, detectProvider().Label)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini is using")
	fmt.Fprintln(w, summary.Using)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini is doing")
	fmt.Fprintln(w, summary.Doing)
	fmt.Fprintln(w, summary.Progress)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Views {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Still missing")
	if len(summary.Missing) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	if len(summary.Uncertain) > 0 {
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Uncertain {
			fmt.Fprintf(w, "- %s\n", item)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Next step")
	fmt.Fprintln(w, summary.NextStep)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Safe to do")
	fmt.Fprintln(w, summary.SafeToDo)
}

func renderOpenShelf(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Open something ready")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Views {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Send / share")
	if len(summary.Exports) == 0 {
		fmt.Fprintln(w, "- Nothing to share yet")
	} else {
		for _, item := range summary.Exports {
			fmt.Fprintf(w, "- %s\n", item.Label)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Details")
	if len(summary.Details) == 0 {
		fmt.Fprintln(w, "- No extra details yet")
	} else {
		for _, item := range summary.Details {
			fmt.Fprintf(w, "- %s\n", item.Label)
		}
	}
}

func dedupeItems(items []catalogItem) []catalogItem {
	seen := map[string]bool{}
	out := make([]catalogItem, 0, len(items))
	for _, item := range items {
		key := normalizeName(item.Label + ":" + item.Path)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func normalizeName(value string) string {
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ")
	value = replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
	return strings.Join(strings.Fields(value), " ")
}

func containsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func titleCase(value string) string {
	parts := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(value))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
