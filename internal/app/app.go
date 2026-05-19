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
	"regexp"
	"sort"
	"strconv"
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
	Dir                string
	PackID             string
	WorkUnitID         string
	Title              string
	WorkingWith        string
	ModelLabel         string
	ModelReason        string
	ModelFeedback      string
	EffortLevel        string
	VerificationLevel  string
	VerificationReason string
	RoutePolicy        string
	RouteReason        string
	ContinuityReason   string
	State              string
	Views              []catalogItem
	Exports            []catalogItem
	Details            []catalogItem
	Missing            []string
	Uncertain          []string
	Using              string
	Doing              string
	Progress           string
	NextStep           string
	SafeToDo           string
	Thread             workThread
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

	return safelyRunInteractive(stderr, func() int {
		if len(args) == 0 {
			return runLauncher(stdin, stdout, stderr)
		}

		switch args[0] {
		case "check":
			return runCheck(args[1:], stdout, stderr)
		case "observe":
			return runObserve(args[1:], stdout, stderr)
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
	})
}

func safelyRunInteractive(stderr io.Writer, fn func() int) (exitCode int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintf(stderr, "Jini hit an unexpected internal error and stopped safely.\n")
			exitCode = 1
		}
	}()
	return fn()
}

func runLauncher(stdin io.Reader, stdout, stderr io.Writer) int {
	if stdin != nil {
		_ = maybeWarmLocalRuntimeCapabilitiesAsync()
	}
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		active, activeErr := listActiveWorkSummaries(nil)
		if activeErr == nil && len(active) > 0 {
			return runActiveWorkLauncher(active, stdin, stdout, stderr)
		}
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
	return handleCurrentWorkAction(action, summary, session, stdout, stderr)
}

func handleCurrentWorkAction(action string, summary *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	switch normalizeName(action) {
	case "1", "continue", "continue current work", "keep going":
		item := nextUsefulItem(summary)
		if item == nil {
			renderCheck(stdout, summary)
			return 0
		}
		renderItem(stdout, item)
	case "2", "open", "open ready work", "open ready", "open whats ready", "open what's ready", "open what is ready", "show whats ready", "show what's ready", "show what is ready":
		renderOpenShelf(stdout, summary)
	case "see what is still missing", "show what is missing", "missing", "check":
		renderMissingOnly(stdout, summary)
	case "model upvote", "upvote model", "model was right":
		if err := saveModelFeedback(summary.Dir, "upvoted", ""); err != nil {
			fmt.Fprintf(stderr, "Could not save model feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved model feedback: upvoted.")
	case "accepted as is", "accept as is", "artifact accepted", "accepted":
		if err := saveModelFeedback(summary.Dir, "accepted-as-is", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact feedback: accepted-as-is.")
	case "needed light edits", "needs light edits", "light edits":
		if err := saveModelFeedback(summary.Dir, "needed-light-edits", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact feedback: needed-light-edits.")
	case "used this", "used it", "kept this":
		if err := saveArtifactOutcome(summary.Dir, "used-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact outcome: used-this.")
	case "shared this", "sent this", "forwarded this", "handed this off", "used this to hand off":
		if err := saveArtifactOutcome(summary.Dir, "shared-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact outcome: shared-this.")
	case "model downvote", "downvote model", "model was wrong":
		if err := saveModelFeedback(summary.Dir, "downvoted", ""); err != nil {
			fmt.Fprintf(stderr, "Could not save model feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved model feedback: downvoted.")
	case "not useful", "artifact was not useful", "not good enough":
		if err := saveModelFeedback(summary.Dir, "not-useful", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact feedback: not-useful.")
	case "replaced this", "rewrote this", "made a new one", "did not use this":
		if err := saveArtifactOutcome(summary.Dir, "replaced-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved artifact outcome: replaced-this.")
	case "switch work", "switch", "show active work", "active work", "switch project":
		return runSwitchWorkPicker(summary, scanner, stdout, stderr)
	case "plan this first", "plan first", "plan", "requirements", "design", "help me plan this":
		renderPlanFirst(stdout, summary)
	case "3", "new", "start new", "start something new", "start something else", "start new work":
		if scanner == nil {
			renderNewWorkLauncher(stdout)
			return 0
		}
		return runNewWorkIntakeWithScanner(scanner, stdout, stderr)
	default:
		if scanner != nil && strings.TrimSpace(action) != "" {
			return startNewWorkFromRawInput(action, scanner, stdout, stderr)
		}
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
	if shouldRefreshLocalBenchmarkForDoctor() {
		_ = currentLocalRuntimeCapabilities(context.Background())
	}
	provider := detectProvider()
	renderProviderDoctor(stdout, provider)
	if provider.Status == "ok" {
		return 0
	}
	return 1
}

func detectProvider() providerConfig {
	if route := detectRoute(); route.Active {
		return route.Provider
	}
	return detectLegacyProvider()
}

func detectLegacyProvider() providerConfig {
	switch configuredProviderMode() {
	case "local-preview":
		return detectLocalPreviewProvider()
	case "local-slm":
		return detectLocalSLMProvider()
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
			Missing: []string{"Supported JINI_PROVIDER value: auto, claude, azure-openai, bedrock, local-slm, or local-preview"},
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

func detectLocalSLMProvider() providerConfig {
	defaultModelID, defaultModelLabel := resolveLocalSLMDefaultModel()
	device := currentDeviceProfile()
	missing := []string{}
	if strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")) == "" {
		missing = append(missing, "JINI_LOCAL_SLM_ENDPOINT")
	}
	if strings.TrimSpace(defaultModelID) == "" {
		missing = append(missing, "JINI_LOCAL_SLM_MODEL")
	}
	settings := []string{
		providerSettingLine("local-slm"),
		"JINI_LOCAL_SLM_ENDPOINT: " + presentOrMissing("JINI_LOCAL_SLM_ENDPOINT"),
		"DEVICE_CLASS: " + firstNonEmpty(device.DeviceClass, "unknown"),
		"DEVICE_OS: " + strings.TrimSpace(firstNonEmpty(device.OS, "unknown")) + " " + strings.TrimSpace(firstNonEmpty(device.OSVersion, "unknown")),
		"LOCAL_ACCELERATOR: " + firstNonEmpty(device.AcceleratorClass, "unknown"),
		"LOCAL_RUNTIME_CLASS: " + firstNonEmpty(device.LocalRuntimeClass, "unknown"),
	}
	for _, slot := range []struct {
		env   string
		label string
	}{
		{"JINI_LOCAL_SLM_MODEL", "Default local model"},
		{"JINI_LOCAL_SLM_FAST_MODEL", "Fast profile"},
		{"JINI_LOCAL_SLM_WORKHORSE_MODEL", "Workhorse profile"},
		{"JINI_LOCAL_SLM_DEEP_MODEL", "Deep profile"},
		{"JINI_LOCAL_SLM_MULTIMODAL_MODEL", "Multimodal profile"},
	} {
		if slot.env == "JINI_LOCAL_SLM_MODEL" {
			settings = append(settings, slot.env+": "+presentOrMissing(slot.env)+" -> "+firstNonEmpty(defaultModelLabel, "missing"))
			continue
		}
		state := strings.TrimSpace(device.LocalProfileStates[toolModeForLocalSetting(slot.env)])
		if state != "" {
			settings = append(settings, slot.env+": "+presentOrMissing(slot.env)+" ("+state+" on this device)")
			continue
		}
		settings = append(settings, slot.env+": "+presentOrMissing(slot.env))
	}
	if len(missing) == 0 {
		settings = append(settings, freshLocalBenchmarkSummaryLines()...)
		settings = append(settings, freshLocalMultimodalLearningSummaryLines()...)
	}
	return providerConfig{
		ID:       "local-slm",
		Label:    "Local SLM",
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"JINI_LOCAL_SLM_API_KEY: " + presentOrMissing("JINI_LOCAL_SLM_API_KEY")},
	}
}

func shouldRefreshLocalBenchmarkForDoctor() bool {
	if configuredProviderMode() == "local-slm" {
		return true
	}
	switch configuredToolMode() {
	case "local-fast", "local-workhorse", "local-deep", "local-multimodal":
		return true
	default:
		return false
	}
}

func toolModeForLocalSetting(envName string) string {
	switch envName {
	case "JINI_LOCAL_SLM_FAST_MODEL":
		return "local-fast"
	case "JINI_LOCAL_SLM_WORKHORSE_MODEL":
		return "local-workhorse"
	case "JINI_LOCAL_SLM_DEEP_MODEL":
		return "local-deep"
	case "JINI_LOCAL_SLM_MULTIMODAL_MODEL":
		return "local-multimodal"
	default:
		return ""
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

func workingWithLabel(provider providerConfig) string {
	if route := detectRoute(); route.Active {
		label := route.ToolLabel
		if provider.Label != "" && provider.Label != route.ToolLabel && provider.ID != "local-preview" {
			label += " via " + provider.Label
		}
		if route.ChosenAutomatically {
			return label + " (chosen automatically)"
		}
		return label
	}
	if configuredProviderMode() == "auto" {
		return provider.Label + " (chosen automatically)"
	}
	return provider.Label
}

func runNewWorkIntake(stdin io.Reader, stdout, stderr io.Writer) int {
	session := bufio.NewScanner(stdin)
	return runNewWorkIntakeWithScanner(session, stdout, stderr)
}

func runNewWorkIntakeWithScanner(session *bufio.Scanner, stdout, stderr io.Writer) int {
	_ = maybeWarmLocalRuntimeCapabilitiesAsync()
	renderNewWorkLauncher(stdout)

	for {
		firstRaw, ok := readInputLine(session, stdout)
		if !ok {
			return 0
		}
		if isGreetingOnly(firstRaw) {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Hi.")
			fmt.Fprintln(stdout, "Tell me what you want finished, or paste notes when you're ready.")
			fmt.Fprintln(stdout)
			continue
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
		return startNewWorkFromRawInput(firstRaw, session, stdout, stderr)
	}
}

func startNewWorkFromRawInput(firstRaw string, session *bufio.Scanner, stdout, stderr io.Writer) int {
	var source string
	choice, err := resolveStarterChoice(firstRaw)
	if err != nil {
		source = strings.TrimSpace(firstRaw)
		choice = classifyStarterChoice(source)
	} else {
		var ok bool
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
	inputItems, normalizedSource := inputItemsForSource(source)
	if strings.TrimSpace(normalizedSource) != "" {
		source = normalizedSource
	}
	clarifiedSource, clarificationItem, ok := maybeClarifyStarterSource(choice, source, session, stdout)
	if !ok {
		return 0
	}
	source = clarifiedSource
	if clarificationItem.InputID != "" {
		inputItems = append(inputItems, clarificationItem)
	}

	request := providerGenerationRequest{
		Choice: choice,
		Title:  deriveStarterTitle(choice.DefaultName, source, choice.PackID),
		Source: source,
	}
	decision := detectRouteForRequest(request)
	renderRouteDecisionCard(stdout, request, decision)

	summary, err := bootstrapStarterWork(choice, source, "quick", inputItems)
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
	case "3", "i am not sure", "i'm not sure", "i’m not sure", "im not sure", "i m not sure", "not sure", "unsure":
		return starterChoice{PackID: "auto", ChoiceLabel: "I am not sure", DefaultName: "First Useful Pass", State: "decided"}, nil
	case "plan this first", "plan first":
		return starterChoice{PackID: "auto", ChoiceLabel: "Plan this first", DefaultName: "Plan First", State: "modeled"}, nil
	}
	for _, packID := range starterPackMenuOrder {
		profile, ok := starterProfileForPack(packID)
		if !ok {
			continue
		}
		for _, alias := range profile.MenuAliases {
			if choice == alias {
				resolved, ok := starterChoiceForPack(profile.PackID)
				if ok {
					return resolved, nil
				}
			}
		}
	}
	return starterChoice{}, fmt.Errorf("I couldn't match %q to a starter flow yet.", raw)
}

func isGreetingOnly(raw string) bool {
	switch normalizeName(raw) {
	case "hello", "hi", "hey", "hey there", "hello there", "good morning", "good afternoon", "good evening", "morning", "evening":
		return true
	default:
		return false
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
	packID := detectStarterPackFromSource(source)
	choice, ok := starterChoiceForPack(packID)
	if !ok {
		return starterChoice{PackID: "general-work", ChoiceLabel: "First Useful Pass", DefaultName: "First Useful Pass", State: "decided"}
	}
	return choice
}

func maybeClarifyStarterSource(choice starterChoice, source string, scanner *bufio.Scanner, stdout io.Writer) (string, inputItem, bool) {
	if scanner == nil {
		return source, inputItem{}, true
	}
	prompt, ok := clarificationPromptForStarter(choice, source)
	if !ok {
		return source, inputItem{}, true
	}
	answer, answered := readPromptLine(scanner, stdout, prompt)
	if !answered {
		return "", inputItem{}, false
	}
	answer = strings.TrimSpace(answer)
	if answer == "" || normalizeName(answer) == "skip" {
		return source, inputItem{}, true
	}
	return mergeClarifiedSource(source, answer), inputItem{
		InputID:   "clarified-scope",
		Kind:      "clarification",
		Title:     "Clarified scope",
		Status:    "processed",
		Preview:   compactPreview(answer, 120),
		OriginRef: answer,
	}, true
}

func clarificationPromptForStarter(choice starterChoice, source string) (string, bool) {
	profile, ok := starterProfileForPack(choice.PackID)
	if !ok {
		return "", false
	}
	return clarificationPromptForProfile(profile, source)
}

func mergeClarifiedSource(source, answer string) string {
	base := strings.TrimSpace(source)
	scope := strings.TrimSpace(answer)
	if base == "" {
		return scope
	}
	if scope == "" {
		return base
	}
	if strings.HasSuffix(base, ".") || strings.HasSuffix(base, "!") || strings.HasSuffix(base, "?") {
		return base + " Scope: " + scope
	}
	return base + ". Scope: " + scope
}

func bootstrapStarterWork(choice starterChoice, source, detail string, inputItems []inputItem) (*workSummary, error) {
	if len(inputItems) == 0 {
		inputItems, _ = inputItemsForSource(source)
	}
	stateRoot := sessionStateRoot()
	workRoot := filepath.Join(stateRoot, "work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, err
	}

	title := deriveStarterTitle(choice.DefaultName, source, choice.PackID)
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
	if err := enrichSmartHyperlinksInViews(workDir, providerGenerationRequest{
		Choice: choice,
		Title:  title,
		Source: source,
	}); err != nil {
		return nil, err
	}
	if err := saveInputItems(workDir, inputItems); err != nil {
		return nil, err
	}
	initialSummary := &workSummary{
		Dir:        workDir,
		PackID:     choice.PackID,
		WorkUnitID: slugify(title),
		Title:      title,
		State:      choice.State,
		Views:      collectViews(workDir, choice.PackID),
		Missing:    inferMissing(choice.State, collectDetails(workDir), choice.PackID, source),
		Uncertain:  inferUncertain(choice.PackID, inferMissing(choice.State, collectDetails(workDir), choice.PackID, source), source),
		Doing:      inferDoing(choice.PackID, choice.State),
		NextStep:   inferNextStep(choice.PackID, collectViews(workDir, choice.PackID)),
		SafeToDo:   "Nothing has been sent yet. You can review before sharing.",
	}
	if err := saveThreadState(workDir, synthesizeThreadState(initialSummary)); err != nil {
		return nil, err
	}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, title, source); err != nil {
		return nil, err
	}
	if err := saveArtifactFeedbackBaseline(workDir); err != nil {
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
	summary, err := loadWorkSummary(workDir, current)
	if err != nil {
		return nil, err
	}
	if err := saveSemanticEnvelopeForSummary(summary); err != nil {
		return nil, err
	}
	return summary, nil
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

	profile, ok := starterProfileForPack(choice.PackID)
	if ok && profile.Writer != nil {
		return profile.Writer(workDir, title, source, detail)
	}
	return writeFirstUsefulPassStarterWork(workDir, title, source)
}

func writeMeetingStarterWork(workDir, title, source, detail string) error {
	decisions := starterSourceBullets(source, 4, []string{
		"Summarize the main meeting decisions in one short list before sending anything.",
	})
	ownersToConfirm := starterMeetingOwnersToConfirm(source, decisions)
	openQuestions := starterMeetingOpenQuestions(source)
	nextMoves := starterMeetingNextMoves(source)
	followup := strings.Join([]string{
		fmt.Sprintf("# Sendable Follow-Up: %s", title),
		"",
		"## Send this note",
		fmt.Sprintf("Team, here is the clean follow-up from **%s**.", title),
		"",
		fmt.Sprintf("I pulled this from the current notes: %s", source),
		"",
		"Please reply if any owner, due date, dependency, or open question below needs correction before this is treated as final.",
		"",
		"## Decisions captured from the notes",
		bulletLines(decisions),
		"## Owners and due dates to confirm",
		bulletLines(ownersToConfirm),
		"## Open questions to close",
		bulletLines(openQuestions),
		"## Recommended next move",
		bulletLines(nextMoves),
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
	ownersView := strings.Join([]string{
		"# Owners and Due Points",
		"",
		"## Confirmed from the notes",
		bulletLines(decisions),
		"## Still missing owner or date",
		bulletLines(ownersToConfirm),
		"## Follow-up questions",
		bulletLines(openQuestions),
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "views", "owners-and-due-points.md"), []byte(ownersView), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Task List\n\n"+bulletLines(ownersToConfirm)), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Tasks"})
}

func writeResearchStarterWork(workDir, title, source, detail string) error {
	readyNow := starterSourceBullets(source, 3, []string{
		"User intent is clear enough to shape the first build slice.",
	})
	mustClear := starterResearchMustClear(source)
	firstSlice := starterResearchFirstSlice(source)
	whoNeeds := starterResearchWhoNeeds(source)
	stillToConfirm := starterResearchStillToConfirm(source)
	prd := strings.Join([]string{
		fmt.Sprintf("# Build-Readiness Check: %s", title),
		"",
		"## What looks ready now",
		bulletLines(readyNow),
		"## Must clear before build",
		bulletLines(mustClear),
		"## Recommended first slice",
		bulletLines(firstSlice),
		"## Who needs to answer what",
		bulletLines(whoNeeds),
		"## Still to confirm",
		bulletLines(stillToConfirm),
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
	missingView := strings.Join([]string{
		"# Missing Pieces Before Build",
		"",
		"## Must clear before build",
		bulletLines(mustClear),
		"## Who needs to answer what",
		bulletLines(whoNeeds),
		"## Still to confirm",
		bulletLines(stillToConfirm),
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "views", "missing-pieces-before-build.md"), []byte(missingView), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Missing Pieces Before Build\n\n"+bulletLines(mustClear)), 0o644); err != nil {
		return err
	}
	return writeStarterArtifacts(workDir, []string{"Brief", "Plan", "Tasks", "Evidence"})
}

func writeTravelStarterWork(workDir, title, source, detail string) error {
	ctx := parseTravelStarterContext(source)
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
	for _, item := range starterTripDays(ctx) {
		lines = append(lines, item)
	}
	lines = append(lines,
		"",
		"## Budget sketch",
	)
	for _, item := range starterTripBudget(ctx) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## Logistics to lock",
	)
	for _, item := range starterTripLogistics(ctx) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## If something changes",
	)
	for _, item := range starterTripContingencies(ctx) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines,
		"",
		"## Still to confirm",
	)
	for _, item := range travelStillToConfirm(ctx) {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "")
	if err := os.WriteFile(filepath.Join(workDir, "views", "itinerary.md"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "budget-sketch.md"), []byte("# Budget Sketch\n\n"+bulletLines(starterTripBudget(ctx))), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "travel-logistics.md"), []byte("# Travel Logistics\n\n"+bulletLines(starterTripLogistics(ctx))), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(workDir, "views", "tasks.md"), []byte("# Task List\n\n"+bulletLines(travelTaskList(ctx))), 0o644); err != nil {
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

func starterTripDays(ctx travelStarterContext) []string {
	dayCount := ctx.DayCount
	if dayCount <= 0 {
		dayCount = 7
	}
	if dayCount > 14 {
		dayCount = 14
	}
	destination := ctx.Destination
	if destination == "" {
		destination = "the destination"
	}
	themes := []string{
		"First neighborhood anchor",
		"Headliner day",
		"Food and wandering day",
		"Flexible favorite day",
		"Day trip or second anchor",
		"Slow day and catch-up",
		"Last highlights and packing buffer",
		"Second-city or deeper exploration",
		"Open favorites day",
		"Departure buffer",
	}
	lines := []string{}
	for day := 1; day <= dayCount; day++ {
		switch {
		case day == 1:
			lines = append(lines,
				fmt.Sprintf("### Day %d: Arrive and settle into %s", day, destination),
				fmt.Sprintf("- Keep the first day light so arrival, check-in, and your first look at %s do not turn into a forced march.", destination),
			)
		case day == dayCount:
			lines = append(lines,
				fmt.Sprintf("### Day %d: Buffer and departure", day),
				"- Keep the final day intentionally light so checkout, bags, airport or rail transfer, and one last meal fit without stress.",
			)
		default:
			index := day - 2
			if index < len(ctx.MustDos) {
				mustDo := ctx.MustDos[index]
				lines = append(lines,
					fmt.Sprintf("### Day %d: %s", day, titleCase(cleanForTitle(mustDo))),
					fmt.Sprintf("- Build the day around %s, then leave enough room for meals, transit, queues, and one lighter backup if timing slips.", strings.ToLower(strings.TrimSpace(mustDo))),
				)
				continue
			}
			theme := themes[min(index, len(themes)-1)]
			lines = append(lines,
				fmt.Sprintf("### Day %d: %s", day, theme),
				fmt.Sprintf("- Use this day to balance one anchor in %s with slower time for food, neighborhoods, or recovery so the trip stays usable end to end.", destination),
			)
		}
	}
	return lines
}

func starterTripBudget(ctx travelStarterContext) []string {
	destination := firstNonEmpty(ctx.Destination, "the destination")
	items := []string{
		fmt.Sprintf("Lodging: choose the base first because where you stay in %s will shape daily transit, fatigue, and how much you can fit into each day.", destination),
		"Food: separate anchor meals from routine meals so the budget stays honest instead of getting eaten by convenience spending.",
		"Transit: include both arrival or departure transfer and the local movement needed to connect the trip's main anchors.",
	}
	if len(ctx.MustDos) > 0 {
		items = append(items, fmt.Sprintf("Tickets: reserve room for at least one paid anchor such as %s before filling the rest of the week.", strings.ToLower(ctx.MustDos[0])))
	}
	return items
}

func starterTripLogistics(ctx travelStarterContext) []string {
	items := []string{
		"Pick the base area before overcommitting the daily plan so each day does not inherit hidden transit cost.",
		"Lock arrival and departure transfer details before filling optional activities around them.",
	}
	if len(ctx.MustDos) > 0 {
		items = append(items, fmt.Sprintf("Reserve any timed anchors like %s before treating the rest of the itinerary as fixed.", strings.ToLower(ctx.MustDos[0])))
	} else {
		items = append(items, "Reserve the first timed anchor before treating the rest of the itinerary as fixed.")
	}
	return items
}

func starterTripContingencies(ctx travelStarterContext) []string {
	return []string{
		"Have one indoor or low-effort backup for every outdoor-heavy or reservation-heavy day.",
		"Leave one uncommitted slot so a sold-out booking or bad weather does not wreck the rest of the plan.",
		"Protect one slower day in the middle of longer trips so fatigue does not spill into the final third.",
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

func starterSourceBullets(source string, limit int, fallback []string) []string {
	parts := splitSourceFragments(source)
	out := make([]string, 0, min(limit, len(parts)))
	seen := map[string]bool{}
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		if !strings.HasSuffix(clean, ".") && !strings.HasSuffix(clean, "!") && !strings.HasSuffix(clean, "?") {
			clean += "."
		}
		key := normalizeName(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, strings.ToUpper(clean[:1])+clean[1:])
		if len(out) >= limit {
			break
		}
	}
	if len(out) == 0 {
		return append([]string{}, fallback...)
	}
	return out
}

func splitSourceFragments(source string) []string {
	fields := strings.FieldsFunc(source, func(r rune) bool {
		switch r {
		case '\n', '\r', '.', '!', '?', ';':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		clean := strings.TrimSpace(field)
		if clean == "" {
			continue
		}
		out = append(out, clean)
	}
	return out
}

func starterMeetingOwnersToConfirm(source string, decisions []string) []string {
	items := []string{}
	for _, decision := range decisions {
		trimmed := strings.TrimSpace(strings.TrimSuffix(decision, "."))
		if trimmed == "" {
			continue
		}
		items = append(items, "Confirm owner and due date for: "+strings.ToLower(trimmed)+".")
		if len(items) >= 3 {
			break
		}
	}
	if containsAny(normalizeName(source), []string{"owner", "due date", "deadline"}) {
		items = append(items, "Reply with any missing owner, due date, or dependency before this note is treated as final.")
	}
	if len(items) == 0 {
		items = []string{
			"Confirm the owner for the highest-priority follow-up item.",
			"Confirm the due date or next checkpoint for each promised deliverable.",
		}
	}
	return dedupeStrings(items)
}

func starterMeetingOpenQuestions(source string) []string {
	items := []string{}
	normalized := normalizeName(source)
	switch {
	case containsAny(normalized, []string{"pricing", "launch"}):
		items = append(items, "Which launch or pricing detail is still blocked on a separate decision?")
	case containsAny(normalized, []string{"renewal", "risk"}):
		items = append(items, "Which renewal risk changes the immediate action plan if it stays unresolved?")
	}
	if containsAny(normalized, []string{"open question", "question", "unknown"}) {
		items = append(items, "Which open question needs a named owner before the next check-in?")
	}
	items = append(items, "Does any dependency or approval need to be called out before work starts?")
	return dedupeStrings(items)
}

func starterMeetingNextMoves(source string) []string {
	items := []string{
		"Send this note today so everyone is working from the same commitments.",
		"Ask each owner to reply with any date or dependency risk before the next check-in.",
	}
	if containsAny(normalizeName(source), []string{"launch", "release", "ship"}) {
		items = append(items, "Close any launch-blocking question before treating the plan as locked.")
	} else {
		items = append(items, "Close the highest-risk open question before the work is considered settled.")
	}
	return items
}

func starterResearchMustClear(source string) []string {
	items := []string{
		"Name the approval owner and the exact decision needed before build starts.",
		"Add a rollback note for the first released slice so failure handling is not implicit.",
		"Lock the first implementation slice to one user-facing outcome instead of a broad theme.",
	}
	normalized := normalizeName(source)
	if containsAny(normalized, []string{"handoff", "hand off"}) {
		items = append(items, "Confirm what the receiving team is expected to build first after handoff.")
	}
	return dedupeStrings(items)
}

func starterResearchFirstSlice(source string) []string {
	items := starterSourceBullets(source, 1, []string{
		"Start with the thinnest slice that proves the user-facing value before wider expansion.",
	})
	first := strings.TrimSpace(strings.TrimSuffix(items[0], "."))
	return []string{
		"Turn " + strings.ToLower(first) + " into the first thin slice that can be built and checked quickly.",
		"Keep the first slice small enough that approval, rollback, and success checks stay obvious.",
	}
}

func starterResearchWhoNeeds(source string) []string {
	items := []string{
		"Product owner: confirm approval and the success boundary for the first slice.",
		"Engineering owner: record rollback behavior and implementation scope before build starts.",
	}
	if containsAny(normalizeName(source), []string{"handoff", "hand off"}) {
		items = append(items, "Receiving team lead: confirm the handoff is specific enough to execute without guesswork.")
	}
	return items
}

func starterResearchStillToConfirm(source string) []string {
	items := []string{
		"Whether approval is already explicit or still only implied in discussion.",
		"Whether the rollback note belongs in the same artifact or a linked operational surface.",
	}
	if containsAny(normalizeName(source), []string{"notification", "notifications"}) {
		items = append(items, "Which notification path or user case should be the very first implementation slice.")
	}
	return items
}

func dedupeStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		key := normalizeName(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, clean)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func deriveStarterTitle(defaultName, source, packID string) string {
	if packID == "travel-plan" {
		ctx := parseTravelStarterContext(source)
		if ctx.DayCount > 0 && strings.TrimSpace(ctx.Destination) != "" {
			return fmt.Sprintf("%d Day %s Trip", ctx.DayCount, ctx.Destination)
		}
		if strings.TrimSpace(ctx.Destination) != "" {
			return fmt.Sprintf("Trip To %s", ctx.Destination)
		}
		source = ctx.BaseSource
	} else {
		source = primaryStarterTitleSource(source)
	}
	cleaned := titleCase(cleanForTitle(source))
	if strings.TrimSpace(cleaned) == "" {
		return defaultName
	}
	return cleaned
}

func primaryStarterTitleSource(source string) string {
	trimmed := strings.TrimSpace(source)
	lower := strings.ToLower(trimmed)
	if idx := strings.Index(lower, " scope: "); idx >= 0 {
		return strings.TrimSpace(strings.TrimRight(trimmed[:idx], ".!? "))
	}
	return trimmed
}

type travelStarterContext struct {
	BaseSource  string
	Destination string
	DayCount    int
	MustDos     []string
	Missing     []string
}

func parseTravelStarterContext(source string) travelStarterContext {
	ctx := travelStarterContext{
		BaseSource: primaryStarterTitleSource(source),
	}
	ctx.DayCount = extractTravelDayCount(ctx.BaseSource)
	ctx.Destination = extractTravelDestination(ctx.BaseSource)
	ctx.MustDos = extractTravelMustDos(source)
	ctx.Missing = travelMissingDimensions(source)
	return ctx
}

func travelMissingDimensions(source string) []string {
	return missingScopeDimensions(source, travelScopeDimensions)
}

func extractTravelDayCount(source string) int {
	re := regexp.MustCompile(`(?i)\b(\d+)\s*[- ]day\b`)
	match := re.FindStringSubmatch(source)
	if len(match) < 2 {
		return 0
	}
	count, err := strconv.Atoi(match[1])
	if err != nil || count <= 0 {
		return 0
	}
	return count
}

func extractTravelDestination(source string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b\d+\s*[- ]day\b\s+(.+?)\s+trip\b`),
		regexp.MustCompile(`(?i)\btrip\s+to\s+(.+?)(?:\s+for\b|\s+with\b|[,.]|$)`),
	}
	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(source)
		if len(match) < 2 {
			continue
		}
		destination := titleCase(cleanForTitle(match[1]))
		if strings.TrimSpace(destination) != "" {
			return destination
		}
	}
	return ""
}

func extractTravelMustDos(source string) []string {
	mustDos := []string{}
	seen := map[string]bool{}
	pattern := regexp.MustCompile(`(?i)^(.*?)(?:\s+(?:is|are))?\s+(?:a\s+)?must[- ]?(?:do|see)s?$`)
	fragments := strings.FieldsFunc(source, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ';'
	})
	for _, fragment := range fragments {
		trimmed := strings.TrimSpace(fragment)
		normalized := normalizeName(trimmed)
		if !containsAny(normalized, []string{"must do", "must see", "must dos", "must sees"}) {
			continue
		}
		match := pattern.FindStringSubmatch(trimmed)
		if len(match) < 2 {
			continue
		}
		item := strings.TrimSpace(match[1])
		item = strings.Trim(item, " ,")
		item = strings.TrimPrefix(strings.ToLower(item), "that ")
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := normalizeName(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		mustDos = append(mustDos, item)
	}
	return mustDos
}

func travelStillToConfirm(ctx travelStarterContext) []string {
	if len(ctx.Missing) > 0 {
		items := make([]string, 0, len(ctx.Missing))
		for _, missing := range ctx.Missing {
			items = append(items, titleCase(missing))
		}
		return items
	}
	return []string{"Nothing right now"}
}

func travelTaskList(ctx travelStarterContext) []string {
	if len(ctx.Missing) > 0 {
		items := make([]string, 0, len(ctx.Missing))
		for _, missing := range ctx.Missing {
			items = append(items, "Confirm "+strings.ToLower(missing)+".")
		}
		return items
	}
	return []string{"Nothing right now"}
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

func runObserve(args []string, stdout, stderr io.Writer) int {
	summary, err := resolveSummary(nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if len(args) == 0 || normalizeName(args[0]) == "status" {
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		renderExternalObservationStatus(stdout, summary.Dir)
		return 0
	}
	switch normalizeName(args[0]) {
	case "add":
		targetPath, connectorID, parseErr := parseObserveAddArgs(args[1:])
		if parseErr != nil {
			fmt.Fprintf(stderr, "%v\n", parseErr)
			fmt.Fprintln(stderr, "Usage: jini observe add [--connector github|jira|confluence|markdown] <external-file>")
			return 1
		}
		artifactPath := currentFeedbackArtifactPath(summary)
		item, err := addExternalObservation(summary.Dir, artifactPath, targetPath, connectorID)
		if err != nil {
			fmt.Fprintf(stderr, "Could not add observed file: %v\n", err)
			return 1
		}
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Observing external file")
		fmt.Fprintln(stdout, item.TargetPath)
		return 0
	case "scan":
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		renderExternalObservationStatus(stdout, summary.Dir)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown observe command %q.\n", args[0])
		fmt.Fprintln(stderr, "Try `jini observe`, `jini observe add [--connector github|jira|confluence|markdown] <external-file>`, or `jini observe scan`.")
		return 1
	}
}

func parseObserveAddArgs(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("missing external file to observe")
	}
	var targetPath string
	var connectorID string
	for index := 0; index < len(args); index++ {
		part := strings.TrimSpace(args[index])
		if part == "" {
			continue
		}
		if part == "--connector" {
			if index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" {
				return "", "", fmt.Errorf("missing connector value after --connector")
			}
			connectorID = strings.TrimSpace(args[index+1])
			index++
			continue
		}
		if strings.HasPrefix(part, "--connector=") {
			connectorID = strings.TrimSpace(strings.TrimPrefix(part, "--connector="))
			continue
		}
		if targetPath == "" {
			targetPath = part
			continue
		}
		return "", "", fmt.Errorf("too many arguments for observe add")
	}
	if strings.TrimSpace(targetPath) == "" {
		return "", "", fmt.Errorf("missing external file to observe")
	}
	if connectorID != "" && normalizeConnectorID(connectorID) == "" {
		return "", "", fmt.Errorf("unsupported connector %q", connectorID)
	}
	return targetPath, normalizeConnectorID(connectorID), nil
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
	if err := recordPassiveArtifactObservation(summary.Dir, *item); err != nil {
		fmt.Fprintf(stderr, "Could not record artifact open: %v\n", err)
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
	inputs := loadInputItems(resolved, packID)
	source := sourceFromInputItems(inputs)
	missing := inferMissing(state, details, packID, source)
	uncertain := inferUncertain(packID, missing, source)
	route := loadWorkRoute(resolved)
	_ = scanExternalObservations(resolved)
	route = loadWorkRoute(resolved)

	summary := &workSummary{
		Dir:    resolved,
		PackID: packID,
		WorkUnitID: firstNonEmpty(strings.TrimSpace(workUnit["work_unit_id"]), func() string {
			if current != nil {
				return current.WorkUnitID
			}
			return ""
		}()),
		Title:              title,
		WorkingWith:        workingWithLabelForSavedRoute(route),
		ModelLabel:         strings.TrimSpace(route.ModelLabel),
		ModelReason:        strings.TrimSpace(route.ModelReason),
		ModelFeedback:      strings.TrimSpace(route.ModelFeedback),
		EffortLevel:        strings.TrimSpace(route.EffortLevel),
		VerificationLevel:  strings.TrimSpace(route.VerificationLevel),
		VerificationReason: strings.TrimSpace(route.VerificationReason),
		RoutePolicy:        strings.TrimSpace(route.RoutePolicy),
		RouteReason:        strings.TrimSpace(route.Reason),
		ContinuityReason:   strings.TrimSpace(route.ContinuityReason),
		State:              state,
		Views:              views,
		Exports:            exports,
		Details:            details,
		Missing:            missing,
		Uncertain:          uncertain,
		Using:              inferUsing(packID),
		Doing:              inferDoing(packID, state),
		Progress:           inferProgress(state),
		SafeToDo:           "Nothing has been sent yet. You can review before sharing.",
	}
	summary.NextStep = inferNextStep(packID, summary.Views)
	summary.Thread = buildWorkThread(summary, inputs, loadThreadState(resolved, summary))
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
	return starterSynthesizedViews(root, packID)
}

func viewCatalogItem(packID, stem, path string) catalogItem {
	if item, ok := starterViewForStem(packID, stem, path); ok {
		return item
	}
	return catalogItem{
		ID:      normalizeName(stem),
		Label:   titleCase(stem),
		Path:    path,
		Aliases: []string{stem},
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

func sourceFromInputItems(items []inputItem) string {
	parts := []string{}
	for _, item := range items {
		switch item.Kind {
		case "text", "clarification", "derived":
			text := strings.TrimSpace(item.OriginRef)
			if text == "" {
				text = strings.TrimSpace(item.Preview)
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, ". ")
}

func inferMissing(state string, details []catalogItem, packID, source string) []string {
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
		missing = append(missing, starterMissing(packID, state, details, source)...)
	}
	return missing
}

func inferUncertain(packID string, missing []string, source string) []string {
	return starterUncertain(packID, missing, source)
}

func inferUsing(packID string) string {
	return starterWorkingWith(packID)
}

func inferDoing(packID, state string) string {
	return starterDoing(packID, state)
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
	return starterNextStep(packID, views)
}

func prioritizeViews(packID string, items []catalogItem) []catalogItem {
	return starterPrioritizeViews(packID, items)
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
	fmt.Fprintln(w, "Paste notes or type what you want finished.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	provider := detectProvider()
	fmt.Fprintln(w, workingWithLabel(provider))
	if provider.ID == "local-preview" {
		fmt.Fprintln(w)
		if configuredToolMode() == "auto" {
			fmt.Fprintln(w, "Auto mode is on. No cloud tool is ready yet.")
		} else if configuredProviderMode() == "auto" {
			fmt.Fprintln(w, "Auto mode is on. No cloud provider is ready yet.")
		}
		fmt.Fprintln(w, "Need setup help? Type `Use Auto` and Jini will help you connect the best available option.")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Not sure? Type `help me finish this`.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "- Turn meeting notes into something I can send")
	fmt.Fprintln(w, "- Check whether a plan is ready to hand off")
	fmt.Fprintln(w, "- Help me plan this")
	fmt.Fprintln(w, "- I am not sure")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Nothing will be sent yet.")
}

func renderRouteDecisionCard(w io.Writer, request providerGenerationRequest, decision routeDecision) {
	if !decision.Active {
		return
	}
	features := classifyRouteFeatures(request)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini will start with")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Tool")
	fmt.Fprintln(w, firstNonEmpty(strings.TrimSpace(decision.ToolLabel), "Not decided yet"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Provider")
	fmt.Fprintln(w, firstNonEmpty(strings.TrimSpace(decision.Provider.Label), "Local preview"))
	if strings.TrimSpace(decision.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, decision.RoutePolicy)
	}
	if strings.TrimSpace(decision.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, decision.ModelLabel)
	}
	if strings.TrimSpace(decision.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, decision.ModelReason)
	}
	if strings.TrimSpace(decision.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(decision.EffortLevel))
	}
	if strings.TrimSpace(decision.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, decision.VerificationLevel)
	}
	if strings.TrimSpace(decision.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, decision.VerificationReason)
	}
	if strings.TrimSpace(decision.Reason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, decision.Reason)
	}
	if strings.TrimSpace(decision.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, decision.ContinuityReason)
	}
	if lines := freshLocalMultimodalLearningViewLines(features); len(lines) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range lines {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Want a different route?")
	fmt.Fprintln(w, "Type `Connect Claude`, `Connect Bedrock`, `Connect Azure OpenAI`, `Connect Local SLM`, or `Use Auto`.")
	fmt.Fprintln(w)
}

func maybeHandleProviderSetupIntent(raw string, scanner *bufio.Scanner, stdout, stderr io.Writer) (bool, int) {
	switch normalizeName(raw) {
	case "use claude code", "connect claude code", "claude code":
		return true, runProviderSetupWizard("claude-code", scanner, stdout, stderr)
	case "use bedrock sonnet", "connect bedrock sonnet", "bedrock sonnet":
		return true, runProviderSetupWizard("bedrock-sonnet", scanner, stdout, stderr)
	case "use chatgpt", "connect chatgpt", "chatgpt":
		return true, runProviderSetupWizard("chatgpt", scanner, stdout, stderr)
	case "use codex", "connect codex", "codex":
		return true, runProviderSetupWizard("codex", scanner, stdout, stderr)
	case "use azure openai", "connect azure openai":
		return true, runProviderSetupWizard("azure-openai", scanner, stdout, stderr)
	case "use claude", "connect claude", "claude", "use anthropic", "anthropic":
		return true, runProviderSetupWizard("claude", scanner, stdout, stderr)
	case "use bedrock", "connect bedrock", "bedrock", "amazon bedrock":
		return true, runProviderSetupWizard("bedrock", scanner, stdout, stderr)
	case "use azure", "connect azure", "azure", "azure openai", "azure open ai":
		return true, runProviderSetupWizard("azure-openai", scanner, stdout, stderr)
	case "use local slm", "connect local slm", "local slm", "local model":
		return true, runProviderSetupWizard("local-slm", scanner, stdout, stderr)
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
	case "claude-code":
		fmt.Fprintln(stdout, "Connect Claude")
		fmt.Fprintln(stdout, "Paste your Anthropic API key. Jini will route this repo through Claude for code work.")
		key, ok := readPromptLine(scanner, stdout, "Anthropic API key")
		if !ok || strings.TrimSpace(key) == "" {
			fmt.Fprintln(stderr, "Claude Code setup needs an API key.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet)")
		if !ok {
			return 1
		}
		if err := saveRouterSettings("claude-code"); err != nil {
			fmt.Fprintf(stderr, "Could not save Claude Code route: %v\n", err)
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
	case "bedrock-sonnet":
		fmt.Fprintln(stdout, "Connect Bedrock")
		region, ok := readPromptLine(scanner, stdout, "AWS region (press Enter for us-east-1)")
		if !ok {
			return 1
		}
		profile, ok := readPromptLine(scanner, stdout, "AWS profile name")
		if !ok || strings.TrimSpace(profile) == "" {
			fmt.Fprintln(stderr, "Bedrock Sonnet setup needs an AWS profile name.")
			return 1
		}
		if err := saveRouterSettings("bedrock-sonnet"); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock Sonnet route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "bedrock",
			"AWS_REGION":    firstNonEmpty(region, "us-east-1"),
			"AWS_PROFILE":   profile,
			"JINI_MODEL":    "sonnet-4.6",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock setup: %v\n", err)
			return 1
		}
	case "chatgpt", "codex", "azure-openai":
		targetLabel := map[string]string{
			"chatgpt":      "Azure OpenAI",
			"codex":        "Azure OpenAI",
			"azure-openai": "Azure OpenAI",
		}[mode]
		fmt.Fprintf(stdout, "Connect %s\n", targetLabel)
		endpoint, ok := readPromptLine(scanner, stdout, "Azure endpoint")
		if !ok || strings.TrimSpace(endpoint) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs an endpoint.")
			return 1
		}
		apiKey, ok := readPromptLine(scanner, stdout, "Azure API key")
		if !ok || strings.TrimSpace(apiKey) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs an API key.")
			return 1
		}
		deployment, ok := readPromptLine(scanner, stdout, "Azure deployment name")
		if !ok || strings.TrimSpace(deployment) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs a deployment name.")
			return 1
		}
		if err := saveRouterSettings(mode); err != nil {
			fmt.Fprintf(stderr, "Could not save %s route: %v\n", targetLabel, err)
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
		if err := clearRouterSettings(); err != nil {
			fmt.Fprintf(stderr, "Could not clear saved tool route: %v\n", err)
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
		if err := clearRouterSettings(); err != nil {
			fmt.Fprintf(stderr, "Could not clear saved tool route: %v\n", err)
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
	case "auto":
		fmt.Fprintln(stdout, "Use auto mode")
		if err := saveRouterSettings("auto"); err != nil {
			fmt.Fprintf(stderr, "Could not save auto tool mode: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "auto",
			"JINI_MODEL":    "auto",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save auto mode: %v\n", err)
			return 1
		}
	case "local-slm":
		fmt.Fprintln(stdout, "Connect Local SLM")
		fmt.Fprintln(stdout, "Point Jini at an OpenAI-compatible local model server. Jini will keep the exact endpoint in this repo's .jini folder.")
		endpoint, ok := readPromptLine(scanner, stdout, "Local SLM endpoint")
		if !ok || strings.TrimSpace(endpoint) == "" {
			fmt.Fprintln(stderr, "Local SLM setup needs an endpoint.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Default local model")
		if !ok || strings.TrimSpace(model) == "" {
			fmt.Fprintln(stderr, "Local SLM setup needs a default model.")
			return 1
		}
		apiKey, ok := readPromptLine(scanner, stdout, "Local SLM API key (press Enter if not needed)")
		if !ok {
			return 1
		}
		fastModel, ok := readPromptLine(scanner, stdout, "Fast profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		workhorseModel, ok := readPromptLine(scanner, stdout, "Workhorse profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		deepModel, ok := readPromptLine(scanner, stdout, "Deep profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		multimodalModel, ok := readPromptLine(scanner, stdout, "Multimodal profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		if err := saveRouterSettings("auto"); err != nil {
			fmt.Fprintf(stderr, "Could not save Local SLM route mode: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":                   "local-slm",
			"JINI_MODEL":                      "auto",
			"JINI_LOCAL_SLM_ENDPOINT":         endpoint,
			"JINI_LOCAL_SLM_MODEL":            model,
			"JINI_LOCAL_SLM_API_KEY":          apiKey,
			"JINI_LOCAL_SLM_FAST_MODEL":       firstNonEmpty(strings.TrimSpace(fastModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_WORKHORSE_MODEL":  firstNonEmpty(strings.TrimSpace(workhorseModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_DEEP_MODEL":       firstNonEmpty(strings.TrimSpace(deepModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_MULTIMODAL_MODEL": firstNonEmpty(strings.TrimSpace(multimodalModel), strings.TrimSpace(model)),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Local SLM setup: %v\n", err)
			return 1
		}
		_ = saveLocalRuntimeCapabilities(benchmarkLocalRuntimeCapabilities(context.Background()))
	case "local-preview":
		fmt.Fprintln(stdout, "Use local preview")
		if err := saveRouterSettings("local-preview"); err != nil {
			fmt.Fprintf(stderr, "Could not save local preview route: %v\n", err)
			return 1
		}
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
		if mode == "local-slm" {
			fmt.Fprintf(stdout, "Setup saved. Working with %s / %s.\n", provider.Label, firstNonEmpty(strings.TrimSpace(configValue("JINI_LOCAL_SLM_MODEL")), "default local model"))
			return 0
		}
		fmt.Fprintf(stdout, "Setup saved. Working with %s.\n", workingWithLabel(provider))
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

func currentFeedbackArtifactPath(summary *workSummary) string {
	if summary == nil {
		return ""
	}
	if item := firstResultItem(summary); item != nil {
		return item.Path
	}
	if len(summary.Views) > 0 {
		return summary.Views[0].Path
	}
	return ""
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
	fmt.Fprintln(w, "- Show what is missing")
	fmt.Fprintln(w, "- Help me plan this")
	fmt.Fprintln(w, "- Start new work")
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
	case "see what is still missing", "show what is missing", "missing", "check", "3":
		renderMissingOnly(stdout, summary)
	case "plan this first", "plan first", "plan", "help me plan this", "4":
		renderPlanFirst(stdout, summary)
	case "start something new", "start new work", "new", "5":
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
	fmt.Fprintln(w, "Help me plan this")
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
	fmt.Fprintln(w, "- Show what's ready")
	fmt.Fprintln(w, "- Show what is missing")
}

func renderCurrentWorkLauncher(w io.Writer, summary *workSummary, interactive bool) {
	fmt.Fprintln(w, "Goal")
	fmt.Fprintln(w, summary.Thread.Goal)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	for _, item := range summary.Thread.WorkingWith {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(summary.Thread.WorkingWith) == 0 {
		fmt.Fprintln(w, "- Nothing attached yet")
	}
	if strings.TrimSpace(summary.Thread.CurrentRoute) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "AI route")
		fmt.Fprintln(w, summary.Thread.CurrentRoute)
	}
	if interactive {
		if strings.TrimSpace(summary.Thread.UpNext) != "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Up next")
			fmt.Fprintln(w, summary.Thread.UpNext)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Ready now")
		if len(summary.Thread.ReadyNow) == 0 {
			fmt.Fprintln(w, "- Nothing is ready yet")
		} else {
			for _, item := range summary.Thread.ReadyNow {
				fmt.Fprintf(w, "- %s\n", item.Label)
			}
		}
		if len(summary.Thread.Blocked) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Blocked")
			for _, item := range summary.Thread.Blocked {
				fmt.Fprintf(w, "- %s\n", item)
			}
		}
		other := otherActiveWorkSummaries(summary)
		if len(other) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Other active work")
			for _, item := range other {
				fmt.Fprintf(w, "- %s\n", item.Title)
			}
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Choose one")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "- Keep going")
		fmt.Fprintln(w, "- Show what's ready")
		fmt.Fprintln(w, "- Show what is missing")
		fmt.Fprintln(w, "- Help me plan this")
		if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Tell Jini how this draft went")
			fmt.Fprintln(w, "- Accepted as is")
			fmt.Fprintln(w, "- Needed light edits")
			fmt.Fprintln(w, "- Not useful")
			fmt.Fprintln(w, "- Shared this")
			fmt.Fprintln(w, "- Replaced this")
			fmt.Fprintln(w, "Advanced: `Used this`, `Model upvote`, or `Model downvote`.")
		}
		if len(other) > 0 {
			fmt.Fprintln(w, "- Switch project")
		}
		fmt.Fprintln(w, "- Start new work")
		return
	}
	if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, summary.Thread.ModelLabel)
	}
	if strings.TrimSpace(summary.Thread.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, summary.Thread.ModelReason)
	}
	if strings.TrimSpace(summary.Thread.ModelFeedback) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model feedback")
		fmt.Fprintln(w, titleCase(summary.Thread.ModelFeedback))
	}
	if strings.TrimSpace(summary.Thread.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, summary.Thread.RoutePolicy)
	}
	if strings.TrimSpace(summary.Thread.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(summary.Thread.EffortLevel))
	}
	if strings.TrimSpace(summary.Thread.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, summary.Thread.VerificationLevel)
	}
	if strings.TrimSpace(summary.Thread.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, summary.Thread.VerificationReason)
	}
	if strings.TrimSpace(summary.Thread.RouteReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, summary.Thread.RouteReason)
	}
	if strings.TrimSpace(summary.Thread.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, summary.Thread.ContinuityReason)
	}
	if len(summary.Thread.MultimodalLearning) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range summary.Thread.MultimodalLearning {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Just finished")
	for _, item := range summary.Thread.JustFinished {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Doing now")
	fmt.Fprintln(w, summary.Thread.DoingNow)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Up next")
	fmt.Fprintln(w, summary.Thread.UpNext)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Now")
	fmt.Fprintln(w, summary.Thread.Now)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Done")
	for _, item := range summary.Thread.Done {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Need")
	fmt.Fprintln(w, summary.Thread.Need)
	if strings.TrimSpace(summary.Thread.NeedReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this matters")
		fmt.Fprintln(w, summary.Thread.NeedReason)
	}
	if len(summary.Thread.NeedOptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options")
		for _, item := range summary.Thread.NeedOptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.Assumptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "If you skip this")
		for _, item := range summary.Thread.Assumptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next")
	fmt.Fprintln(w, summary.Thread.Next)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Thread.ReadyNow {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Thread.ReadyNow) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked")
	if len(summary.Thread.Blocked) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.NotSureAbout) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Thread.NotSureAbout {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	other := otherActiveWorkSummaries(summary)
	if len(other) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Other active work")
		for _, item := range other {
			fmt.Fprintf(w, "- %s\n", item.Title)
		}
	}
	fmt.Fprintln(w)
}

func runActiveWorkLauncher(active []*workSummary, stdin io.Reader, stdout, stderr io.Writer) int {
	renderActiveWorkLauncher(stdout, active)
	if stdin == nil {
		return 0
	}
	session := bufio.NewScanner(stdin)
	action, ok := readOptionalInputLine(session, stdout)
	if !ok || strings.TrimSpace(action) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	return handleActiveWorkSelection(action, active, session, stdout, stderr)
}

func renderActiveWorkLauncher(w io.Writer, active []*workSummary) {
	fmt.Fprintln(w, "Active work")
	fmt.Fprintln(w)
	for index, item := range active {
		fmt.Fprintf(w, "%d. %s\n", index+1, item.Title)
		fmt.Fprintf(w, "   Next: %s\n", item.NextStep)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Type a number to open one, or type `Start something new`.")
}

func handleActiveWorkSelection(action string, active []*workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	selection, err := resolveActiveWorkSelection(action, active)
	if err != nil {
		switch normalizeName(action) {
		case "start something new", "start something else", "start new work", "new", "start new":
			if scanner == nil {
				renderNewWorkLauncher(stdout)
				return 0
			}
			return runNewWorkIntakeWithScanner(scanner, stdout, stderr)
		}
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if err := saveSummaryAsCurrent(selection); err != nil {
		fmt.Fprintf(stderr, "Could not switch work: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Switched to")
	fmt.Fprintln(stdout, selection.Title)
	fmt.Fprintln(stdout)
	renderCheck(stdout, selection)
	return 0
}

func runSwitchWorkPicker(current *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	other := otherActiveWorkSummaries(current)
	if len(other) == 0 {
		fmt.Fprintln(stdout, "No other active work right now.")
		return 0
	}
	fmt.Fprintln(stdout, "Switch project")
	fmt.Fprintln(stdout)
	for index, item := range other {
		fmt.Fprintf(stdout, "%d. %s\n", index+1, item.Title)
		fmt.Fprintf(stdout, "   Next: %s\n", item.NextStep)
	}
	if scanner == nil {
		return 0
	}
	choice, ok := readOptionalInputLine(scanner, stdout)
	if !ok || strings.TrimSpace(choice) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	selection, err := resolveActiveWorkSelection(choice, other)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if err := saveSummaryAsCurrent(selection); err != nil {
		fmt.Fprintf(stderr, "Could not switch work: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Switched to")
	fmt.Fprintln(stdout, selection.Title)
	fmt.Fprintln(stdout)
	renderCheck(stdout, selection)
	return 0
}

func listActiveWorkSummaries(current *workSummary) ([]*workSummary, error) {
	root := filepath.Join(sessionStateRoot(), "work")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	active := []*workSummary{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		summary, loadErr := loadWorkSummary(filepath.Join(root, entry.Name()), nil)
		if loadErr != nil {
			continue
		}
		active = append(active, summary)
	}
	sort.SliceStable(active, func(i, j int) bool {
		return active[i].Title < active[j].Title
	})
	if current == nil {
		return active, nil
	}
	currentPath := ""
	if resolved, err := filepath.Abs(current.Dir); err == nil {
		currentPath = resolved
	}
	out := make([]*workSummary, 0, len(active))
	for _, item := range active {
		itemPath := item.Dir
		if resolved, err := filepath.Abs(item.Dir); err == nil {
			itemPath = resolved
		}
		if currentPath != "" && itemPath == currentPath {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func otherActiveWorkSummaries(current *workSummary) []*workSummary {
	active, err := listActiveWorkSummaries(current)
	if err != nil {
		return nil
	}
	return active
}

func resolveActiveWorkSelection(action string, active []*workSummary) (*workSummary, error) {
	normalized := normalizeName(action)
	for index, item := range active {
		if normalized == normalizeName(fmt.Sprintf("%d", index+1)) || normalized == normalizeName(item.Title) {
			return item, nil
		}
	}
	return nil, fmt.Errorf("I couldn't find that work. Pick a shown number or title.")
}

func saveSummaryAsCurrent(summary *workSummary) error {
	current := &currentWork{
		PackDir:    summary.Dir,
		PackID:     summary.PackID,
		WorkUnitID: summary.WorkUnitID,
		Title:      summary.Title,
		State:      summary.State,
		Health:     inferHealthFromState(summary.State),
	}
	return saveCurrentWork(current)
}

func renderCheck(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Goal")
	fmt.Fprintln(w, summary.Thread.Goal)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	for _, item := range summary.Thread.WorkingWith {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(summary.Thread.WorkingWith) == 0 {
		fmt.Fprintln(w, "- Nothing attached yet")
	}
	if strings.TrimSpace(summary.Thread.CurrentRoute) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "AI route")
		fmt.Fprintln(w, summary.Thread.CurrentRoute)
	}
	if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, summary.Thread.ModelLabel)
	}
	if strings.TrimSpace(summary.Thread.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, summary.Thread.ModelReason)
	}
	if strings.TrimSpace(summary.Thread.ModelFeedback) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model feedback")
		fmt.Fprintln(w, titleCase(summary.Thread.ModelFeedback))
	}
	if strings.TrimSpace(summary.Thread.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, summary.Thread.RoutePolicy)
	}
	if strings.TrimSpace(summary.Thread.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(summary.Thread.EffortLevel))
	}
	if strings.TrimSpace(summary.Thread.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, summary.Thread.VerificationLevel)
	}
	if strings.TrimSpace(summary.Thread.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, summary.Thread.VerificationReason)
	}
	if strings.TrimSpace(summary.Thread.RouteReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, summary.Thread.RouteReason)
	}
	if strings.TrimSpace(summary.Thread.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, summary.Thread.ContinuityReason)
	}
	if len(summary.Thread.MultimodalLearning) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range summary.Thread.MultimodalLearning {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Just finished")
	for _, item := range summary.Thread.JustFinished {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Doing now")
	fmt.Fprintln(w, summary.Thread.DoingNow)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Up next")
	fmt.Fprintln(w, summary.Thread.UpNext)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Now")
	fmt.Fprintln(w, summary.Thread.Now)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Done")
	for _, item := range summary.Thread.Done {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Need")
	fmt.Fprintln(w, summary.Thread.Need)
	if strings.TrimSpace(summary.Thread.NeedReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this matters")
		fmt.Fprintln(w, summary.Thread.NeedReason)
	}
	if len(summary.Thread.NeedOptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options")
		for _, item := range summary.Thread.NeedOptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.Assumptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "If you skip this")
		for _, item := range summary.Thread.Assumptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next")
	fmt.Fprintln(w, summary.Thread.Next)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Thread.ReadyNow {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Thread.ReadyNow) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked")
	if len(summary.Thread.Blocked) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	if len(summary.Thread.NotSureAbout) > 0 {
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Thread.NotSureAbout {
			fmt.Fprintf(w, "- %s\n", item)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Safe to do")
	fmt.Fprintln(w, summary.Thread.SafeToDo)
}

func summaryWorkingWith(summary *workSummary) string {
	if summary != nil && strings.TrimSpace(summary.WorkingWith) != "" {
		return summary.WorkingWith
	}
	return workingWithLabel(detectProvider())
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
