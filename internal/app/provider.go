package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var providerHTTPClient = &http.Client{Timeout: 30 * time.Second}

const (
	defaultAnthropicModel = "claude-sonnet-4-20250514"
	defaultBedrockModel   = "anthropic.claude-sonnet-4-6"
)

type providerGenerationRequest struct {
	Choice starterChoice
	Title  string
	Source string
}

type localRouteFeedbackStats struct {
	SchemaVersion   string                                       `json:"schema_version"`
	ContextType     string                                       `json:"context_type"`
	Routes          map[string]routeFeedbackRow                  `json:"routes"`
	Cohorts         map[string]map[string]localCohortFeedbackRow `json:"cohorts,omitempty"`
	ManualOverrides map[string]map[string]int                    `json:"manual_overrides,omitempty"`
}

type routeFeedbackRow struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
}

type awsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func maybeWriteProviderFirstDraft(ctx context.Context, choice starterChoice, workDir, title, source string) error {
	request := providerGenerationRequest{
		Choice: choice,
		Title:  title,
		Source: source,
	}
	decision := detectRouteForRequest(request)
	if err := saveWorkRoute(workDir, request, decision); err != nil {
		return err
	}
	text, used, actualDecision, err := generateWithConfiguredProviderDecision(ctx, request, decision)
	if err != nil {
		return err
	}
	if !used {
		return nil
	}
	if err := saveWorkRoute(workDir, request, actualDecision); err != nil {
		return err
	}

	path, label := providerPrimaryView(choice.PackID)
	if path == "" {
		return nil
	}
	content := normalizeProviderMarkdown(label, title, text)
	if err := os.WriteFile(filepath.Join(workDir, "views", path), []byte(content), 0o644); err != nil {
		return err
	}
	return enrichSmartHyperlinksInViews(workDir, request)
}

func generateWithConfiguredProvider(ctx context.Context, request providerGenerationRequest) (string, bool, error) {
	text, used, _, err := generateWithConfiguredProviderDecision(ctx, request, detectRouteForRequest(request))
	return text, used, err
}

func generateWithConfiguredProviderDecision(ctx context.Context, request providerGenerationRequest, decision routeDecision) (string, bool, routeDecision, error) {
	provider := providerForDecision(request, decision)
	if provider.ID == "local-preview" {
		return "", false, decision, nil
	}
	if provider.Status != "ok" {
		return "", true, decision, providerSetupError(provider)
	}

	systemPrompt := providerSystemPrompt()
	userPrompt := providerUserPrompt(request)
	text, err := generateProviderText(ctx, provider, request, systemPrompt, userPrompt)
	if err != nil {
		return "", true, decision, err
	}
	consistencyUsed := false
	if shouldCheck, reason := shouldRunSelectiveConsistencyCheck(request, decision); shouldCheck {
		if alternate, consistencyErr := generateConsistencyDraft(ctx, provider, request, reason); consistencyErr == nil && strings.TrimSpace(alternate) != "" {
			text = selectConsistencyWinner(request, decision, text, alternate)
			consistencyUsed = true
		}
	}
	refinedUsed := false
	if shouldRefine, reason := shouldRunSelectiveRefine(request, decision); shouldRefine {
		if refined, refineErr := generateRefinedDraft(ctx, provider, request, text, reason); refineErr == nil && strings.TrimSpace(refined) != "" {
			text = refined
			refinedUsed = true
		}
	}
	decision = actualizeVerificationDecision(request, decision, consistencyUsed, refinedUsed)
	return text, true, decision, nil
}

func generateProviderText(ctx context.Context, provider providerConfig, request providerGenerationRequest, systemPrompt, userPrompt string) (string, error) {
	switch provider.ID {
	case "azure-openai":
		return generateWithAzureOpenAI(ctx, request, systemPrompt, userPrompt)
	case "bedrock":
		return generateWithBedrock(ctx, request, systemPrompt, userPrompt)
	case "anthropic":
		return generateWithAnthropic(ctx, request, systemPrompt, userPrompt)
	case "local-slm":
		return generateWithLocalSLM(ctx, request, systemPrompt, userPrompt)
	default:
		return "", providerSetupError(provider)
	}
}

func providerForRequest(request providerGenerationRequest) providerConfig {
	if route := detectRouteForRequest(request); route.Active {
		return route.Provider
	}
	return detectLegacyProvider()
}

func providerForDecision(request providerGenerationRequest, decision routeDecision) providerConfig {
	if decision.Active {
		return decision.Provider
	}
	return providerForRequest(request)
}

func generateWithAnthropic(ctx context.Context, request providerGenerationRequest, systemPrompt, userPrompt string) (string, error) {
	modelID, _, modelIssue := resolveAnthropicModel()
	if modelIssue != "" {
		return "", providerSetupError(providerConfig{Missing: []string{modelIssue}})
	}
	baseURL := strings.TrimRight(strings.TrimSpace(firstNonEmpty(configValue("ANTHROPIC_BASE_URL"), "https://api.anthropic.com")), "/")
	target := baseURL + "/v1/messages"
	payload := map[string]any{
		"model":      modelID,
		"system":     systemPrompt,
		"max_tokens": 1600,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "text", "text": userPrompt},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configValue("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Claude API request failed. Check network access and API key")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("Claude API request failed with HTTP %d. Run `jini provider doctor` and check the model choice", resp.StatusCode)
	}

	var decoded struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("Claude API returned a response Jini could not read")
	}
	parts := []string{}
	for _, item := range decoded.Content {
		if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("Claude API returned an empty draft")
	}
	return strings.Join(parts, "\n\n"), nil
}

func generateWithAzureOpenAI(ctx context.Context, request providerGenerationRequest, systemPrompt, userPrompt string) (string, error) {
	endpoint := strings.TrimRight(configValue("AZURE_OPENAI_ENDPOINT"), "/")
	deployment := configValue("AZURE_OPENAI_DEPLOYMENT")
	apiVersion := valueOrDefault("AZURE_OPENAI_API_VERSION", "2024-10-21")

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("Azure OpenAI endpoint is not a valid URL. Run `jini provider doctor`.")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/openai/deployments/" + url.PathEscape(deployment) + "/chat/completions"
	query := parsed.Query()
	query.Set("api-version", apiVersion)
	parsed.RawQuery = query.Encode()

	payload := map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
		"max_tokens":  1600,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", configValue("AZURE_OPENAI_API_KEY"))

	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Azure OpenAI request failed. Check network access, endpoint, and deployment")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("Azure OpenAI request failed with HTTP %d. Run `jini provider doctor` and check the deployment", resp.StatusCode)
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("Azure OpenAI returned a response Jini could not read")
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("Azure OpenAI returned an empty draft")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

func generateWithBedrock(ctx context.Context, request providerGenerationRequest, systemPrompt, userPrompt string) (string, error) {
	region := strings.TrimSpace(resolveAWSRegion())
	if region == "" {
		return "", providerSetupError(providerConfig{Missing: []string{"AWS_REGION or AWS_DEFAULT_REGION"}})
	}
	modelID, _ := resolveBedrockModel()
	credentials, err := loadAWSCredentials()
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(configValue("JINI_BEDROCK_ENDPOINT"), "/")
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}
	target := endpoint + "/model/" + url.PathEscape(modelID) + "/converse"
	payload := map[string]any{
		"system": []map[string]string{
			{"text": systemPrompt},
		},
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]string{
					{"text": userPrompt},
				},
			},
		},
		"inferenceConfig": map[string]any{
			"maxTokens":   1600,
			"temperature": 0.2,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	signAWSV4(req, body, credentials, region, "bedrock", time.Now().UTC())

	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Amazon Bedrock request failed. Check network access, region, and model access")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("Amazon Bedrock request failed with HTTP %d. Run `jini provider doctor` and check model access", resp.StatusCode)
	}

	var decoded struct {
		Output struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("Amazon Bedrock returned a response Jini could not read")
	}
	parts := []string{}
	for _, item := range decoded.Output.Message.Content {
		if strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("Amazon Bedrock returned an empty draft")
	}
	return strings.Join(parts, "\n\n"), nil
}

func generateWithLocalSLM(ctx context.Context, request providerGenerationRequest, systemPrompt, userPrompt string) (string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")), "/")
	modelID, modelLabel := resolveLocalSLMModelForRequest(request)
	if endpoint == "" || strings.TrimSpace(modelID) == "" {
		return "", providerSetupError(providerConfig{Missing: []string{"JINI_LOCAL_SLM_ENDPOINT", "JINI_LOCAL_SLM_MODEL"}})
	}

	target := endpoint
	if !strings.HasSuffix(strings.ToLower(target), "/chat/completions") {
		target += "/chat/completions"
	}
	payload := map[string]any{
		"model": modelID,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
		"max_tokens":  1600,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey := strings.TrimSpace(configValue("JINI_LOCAL_SLM_API_KEY")); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	start := time.Now()
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Local SLM request failed. Check the endpoint and local model server")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("Local SLM request failed with HTTP %d. Run `jini provider doctor` and check the local model endpoint", resp.StatusCode)
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("Local SLM returned a response Jini could not read")
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("Local SLM returned an empty draft for %s", firstNonEmpty(modelLabel, "the configured model"))
	}
	text := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if toolMode := resolveLocalSLMToolModeForRequest(request); strings.HasPrefix(toolMode, "local-") {
		_ = recordLocalCohortSample(toolMode, request, int(time.Since(start).Milliseconds()), text)
	}
	return text, nil
}

func loadAWSCredentials() (awsCredentials, error) {
	if configValue("AWS_ACCESS_KEY_ID") != "" && configValue("AWS_SECRET_ACCESS_KEY") != "" {
		return awsCredentials{
			AccessKeyID:     configValue("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: configValue("AWS_SECRET_ACCESS_KEY"),
			SessionToken:    configValue("AWS_SESSION_TOKEN"),
		}, nil
	}
	profile := configValue("AWS_PROFILE")
	if profile == "" {
		return awsCredentials{}, errors.New("Amazon Bedrock credentials are missing. Set AWS_PROFILE or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY.")
	}
	values := loadAWSProfileValues(profile)
	if strings.TrimSpace(values["aws_access_key_id"]) == "" || strings.TrimSpace(values["aws_secret_access_key"]) == "" {
		return awsCredentials{}, errors.New("Amazon Bedrock credentials are not available for AWS_PROFILE. Check your AWS credentials file.")
	}
	return awsCredentials{
		AccessKeyID:     strings.TrimSpace(values["aws_access_key_id"]),
		SecretAccessKey: strings.TrimSpace(values["aws_secret_access_key"]),
		SessionToken:    strings.TrimSpace(values["aws_session_token"]),
	}, nil
}

func resolveAWSRegion() string {
	if region := strings.TrimSpace(firstNonEmpty(configValue("AWS_REGION"), configValue("AWS_DEFAULT_REGION"))); region != "" {
		return region
	}
	profile := configValue("AWS_PROFILE")
	if profile == "" {
		return ""
	}
	return strings.TrimSpace(loadAWSProfileValues(profile)["region"])
}

func configuredProviderMode() string {
	raw := normalizeName(firstNonEmpty(configValue("JINI_PROVIDER"), "auto"))
	switch raw {
	case "", "auto":
		return "auto"
	case "localslm", "local slm", "local model":
		return "local-slm"
	case "local", "local preview", "localpreview":
		return "local-preview"
	case "azure", "azure openai", "azure open ai":
		return "azure-openai"
	case "bedrock", "amazon bedrock", "aws bedrock":
		return "bedrock"
	case "claude", "claude api", "anthropic", "anthropic api":
		return "anthropic"
	default:
		return raw
	}
}

func detectAutoProvider() providerConfig {
	if forced := forcedAutoProviderMode(); forced != "" {
		return withAutoProviderSetting(detectProviderForMode(forced))
	}
	for _, mode := range []string{"local-slm", "anthropic", "azure-openai", "bedrock"} {
		candidate := detectProviderForMode(mode)
		if candidate.Status == "ok" {
			return withAutoProviderSetting(candidate)
		}
	}
	return withAutoProviderSetting(detectLocalPreviewProvider())
}

func detectProviderForMode(mode string) providerConfig {
	switch mode {
	case "anthropic":
		return detectAnthropicProvider()
	case "azure-openai":
		return detectAzureOpenAIProvider()
	case "bedrock":
		return detectBedrockProvider()
	case "local-slm":
		return detectLocalSLMProvider()
	case "local-preview":
		return detectLocalPreviewProvider()
	default:
		return providerConfig{
			ID:      mode,
			Label:   titleCase(mode),
			Status:  "needs setup",
			Missing: []string{"Supported JINI_PROVIDER value: auto, claude, azure-openai, bedrock, local-slm, or local-preview"},
		}
	}
}

func withAutoProviderSetting(provider providerConfig) providerConfig {
	setting := "JINI_PROVIDER: auto -> " + provider.Label
	if len(provider.Settings) > 0 && strings.HasPrefix(provider.Settings[0], "JINI_PROVIDER:") {
		provider.Settings[0] = setting
		return provider
	}
	provider.Settings = append([]string{setting}, provider.Settings...)
	return provider
}

func forcedAutoProviderMode() string {
	modelMode := normalizeName(configuredModelInput())
	switch {
	case isBedrockOnlyModelMode(modelMode):
		return "bedrock"
	case strings.HasPrefix(configuredModelInput(), "claude-"), strings.HasPrefix(modelMode, "claude sonnet"), strings.HasPrefix(modelMode, "claude opus"), strings.HasPrefix(modelMode, "claude haiku"):
		return "anthropic"
	default:
		return ""
	}
}

func configuredModelInput() string {
	return strings.TrimSpace(firstNonEmpty(configValue("JINI_MODEL"), "auto"))
}

func azureModelSettingLine() string {
	raw := configuredModelInput()
	if normalizeName(raw) == "auto" {
		return ""
	}
	return "JINI_MODEL: " + raw + " -> deployment decides the actual Azure model"
}

func bedrockModelSettingLine(modelID, modelLabel string) string {
	raw := configValue("BEDROCK_MODEL_ID")
	modelInput := configuredModelInput()
	switch {
	case strings.TrimSpace(raw) != "":
		return "BEDROCK_MODEL_ID: set -> " + modelLabel
	case normalizeName(modelInput) == "auto":
		return "JINI_MODEL: auto -> " + modelLabel
	default:
		return "JINI_MODEL: " + modelInput + " -> " + modelLabel
	}
}

func anthropicModelSettingLine(modelID, modelLabel string) string {
	raw := strings.TrimSpace(firstNonEmpty(configValue("ANTHROPIC_MODEL"), configValue("JINI_MODEL")))
	if normalizeName(raw) == "auto" || raw == "" {
		return "JINI_MODEL: auto -> " + modelLabel
	}
	return "JINI_MODEL: " + raw + " -> " + modelLabel
}

func providerSettingLine(mode string) string {
	configured := configuredProviderMode()
	if configured == "auto" {
		return "JINI_PROVIDER: auto"
	}
	return "JINI_PROVIDER: " + mode
}

func resolveBedrockModel() (string, string) {
	rawID := configValue("BEDROCK_MODEL_ID")
	if rawID != "" {
		return rawID, friendlyModelLabel("bedrock", rawID)
	}
	rawModel := strings.TrimSpace(configuredModelInput())
	modelMode := normalizeName(rawModel)
	compact := compactModelMode(modelMode)
	switch {
	case modelMode == "", modelMode == "auto", modelMode == "sonnet", modelMode == "claude sonnet", compact == "sonnet46", compact == "claudesonnet46":
		return defaultBedrockModel, "Claude Sonnet 4.6"
	case strings.HasPrefix(rawModel, "anthropic."):
		return rawModel, friendlyModelLabel("bedrock", rawModel)
	default:
		return rawModel, friendlyModelLabel("bedrock", rawModel)
	}
}

func resolveAnthropicModel() (string, string, string) {
	rawModel := strings.TrimSpace(firstNonEmpty(configValue("ANTHROPIC_MODEL"), configValue("JINI_MODEL")))
	modelMode := normalizeName(rawModel)
	compact := compactModelMode(modelMode)
	switch modelMode {
	case "", "auto", "sonnet", "claude sonnet", "sonnet 4", "claude sonnet 4":
		return defaultAnthropicModel, "Claude Sonnet 4", ""
	}
	switch compact {
	case "sonnet46", "claudesonnet46":
		return "", "", "Sonnet 4.6 shortcut is supported only on Bedrock. For direct Claude API, set a full Anthropic model name like claude-sonnet-4-20250514."
	default:
		return rawModel, friendlyModelLabel("anthropic", rawModel), ""
	}
}

func friendlyModelLabel(provider, model string) string {
	mode := normalizeName(model)
	switch {
	case strings.Contains(mode, "sonnet 4 6"), strings.Contains(mode, "sonnet46"):
		return "Claude Sonnet 4.6"
	case strings.Contains(mode, "claude sonnet 4 20250514"), strings.Contains(mode, "claude sonnet 4"), strings.Contains(mode, "sonnet 4"):
		return "Claude Sonnet 4"
	case strings.Contains(mode, "3 5 sonnet"), strings.Contains(mode, "sonnet 3 5"):
		return "Claude Sonnet 3.5"
	case model == "":
		if provider == "bedrock" {
			return "Claude Sonnet 4.6"
		}
		if provider == "anthropic" {
			return "Claude Sonnet 4"
		}
		return ""
	default:
		return model
	}
}

func isBedrockOnlyModelMode(modelMode string) bool {
	compact := compactModelMode(modelMode)
	return compact == "sonnet46" || compact == "claudesonnet46" || strings.HasPrefix(configuredModelInput(), "anthropic.")
}

func compactModelMode(modelMode string) string {
	replacer := strings.NewReplacer(" ", "", ".", "", "-", "", "_", "", "/", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(modelMode)))
}

func localFeedbackPath() string {
	return filepath.Join(sessionStateRoot(), "route-feedback.json")
}

func loadLocalRouteFeedbackStats() localRouteFeedbackStats {
	data, err := os.ReadFile(localFeedbackPath())
	if err != nil {
		return localRouteFeedbackStats{Routes: map[string]routeFeedbackRow{}}
	}
	var payload localRouteFeedbackStats
	if err := json.Unmarshal(data, &payload); err != nil {
		return localRouteFeedbackStats{Routes: map[string]routeFeedbackRow{}}
	}
	if payload.Routes == nil {
		payload.Routes = map[string]routeFeedbackRow{}
	}
	if payload.Cohorts == nil {
		payload.Cohorts = map[string]map[string]localCohortFeedbackRow{}
	}
	if payload.ManualOverrides == nil {
		payload.ManualOverrides = map[string]map[string]int{}
	}
	return payload
}

func saveLocalRouteFeedbackStats(stats localRouteFeedbackStats) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	stats.SchemaVersion = "0.1.0"
	stats.ContextType = "JiniRouteFeedback"
	if stats.Routes == nil {
		stats.Routes = map[string]routeFeedbackRow{}
	}
	if stats.Cohorts == nil {
		stats.Cohorts = map[string]map[string]localCohortFeedbackRow{}
	}
	if stats.ManualOverrides == nil {
		stats.ManualOverrides = map[string]map[string]int{}
	}
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(localFeedbackPath(), append(data, '\n'), 0o600)
}

func recordManualRouteOverride(request providerGenerationRequest, decision routeDecision) error {
	if !decision.Active || decision.ChosenAutomatically {
		return nil
	}
	features := classifyRouteFeatures(request)
	if strings.TrimSpace(features.WorkClass) != "code" {
		return nil
	}
	cohort := routeFeedbackCohortKeyForRequest(request)
	if cohort == "" || strings.TrimSpace(decision.ToolMode) == "" {
		return nil
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.ManualOverrides == nil {
		stats.ManualOverrides = map[string]map[string]int{}
	}
	if stats.ManualOverrides[decision.ToolMode] == nil {
		stats.ManualOverrides[decision.ToolMode] = map[string]int{}
	}
	stats.ManualOverrides[decision.ToolMode][cohort]++
	return saveLocalRouteFeedbackStats(stats)
}

func recordRouteFeedback(toolMode, feedback string) error {
	return recordRouteFeedbackForKey(routeFeedbackKeyForCurrentMode(toolMode), feedback)
}

func recordRouteFeedbackForKey(feedbackKey, feedback string) error {
	feedbackKey = strings.TrimSpace(feedbackKey)
	feedback = strings.TrimSpace(feedback)
	if feedbackKey == "" || feedback == "" {
		return nil
	}
	stats := loadLocalRouteFeedbackStats()
	row := stats.Routes[feedbackKey]
	switch feedback {
	case "upvoted":
		row.Upvotes++
	case "downvoted":
		row.Downvotes++
	default:
		return nil
	}
	stats.Routes[feedbackKey] = row
	return saveLocalRouteFeedbackStats(stats)
}

func routeFeedbackBias(toolMode string) int {
	row := loadLocalRouteFeedbackStats().Routes[routeFeedbackKeyForCurrentMode(toolMode)]
	bias := (row.Upvotes - row.Downvotes) * 4
	if bias > 16 {
		return 16
	}
	if bias < -16 {
		return -16
	}
	return bias
}

func routeFeedbackCohortKeyForRequest(request providerGenerationRequest) string {
	features := classifyRouteFeatures(request)
	return strings.TrimSpace(localBenchmarkCohortKeyForFeatures(features))
}

func routeFeedbackCohortKeyForProfile(profile draftQualityProfile) string {
	cohort := strings.TrimSpace(profile.Cohort)
	if cohort == "" {
		return ""
	}
	if cohort != "multimodal-extract" {
		return cohort
	}
	switch strings.TrimSpace(profile.ModalitySubtype) {
	case "pdf-scan":
		return "multimodal-pdf-scan"
	case "image-screenshot":
		return "multimodal-image-screenshot"
	case "audio-transcript":
		return "multimodal-audio-transcript"
	default:
		return cohort
	}
}

func recordRouteCohortFeedback(feedbackKey string, request providerGenerationRequest, feedback, editClass, editScope, semanticClass, reason string) error {
	feedbackKey = strings.TrimSpace(feedbackKey)
	cohort := routeFeedbackCohortKeyForRequest(request)
	feedback = strings.TrimSpace(feedback)
	if feedbackKey == "" || cohort == "" || feedback == "" {
		return nil
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.Cohorts == nil {
		stats.Cohorts = map[string]map[string]localCohortFeedbackRow{}
	}
	if stats.Cohorts[feedbackKey] == nil {
		stats.Cohorts[feedbackKey] = map[string]localCohortFeedbackRow{}
	}
	row := stats.Cohorts[feedbackKey][cohort]
	switch feedback {
	case "upvoted":
		row.Upvotes++
	case "accepted-as-is":
		row.AcceptedAsIs++
	case "needed-light-edits":
		row.NeededLightEdits++
	case "downvoted":
		row.Downvotes++
	case "not-useful":
		row.NotUseful++
	default:
		return nil
	}
	switch strings.TrimSpace(reason) {
	case "accepted-without-edits":
		row.PassiveAcceptedAsIs++
	case "accepted-after-cosmetic-edits", "accepted-after-light-edits", "needed-core-wording-edits", "light-edits-confirmed":
		row.PassiveNeededLightEdits++
	case "accepted-after-heavy-edits", "accepted-after-core-rewrite", "needed-core-rewrite", "needed-heavy-edits":
		row.PassiveNeededHeavyEdits++
	}
	switch strings.TrimSpace(editScope) {
	case "header-only":
		row.PassiveHeaderOnlyEdits++
	case "core-sections":
		row.PassiveCoreSectionEdits++
	}
	switch strings.TrimSpace(semanticClass) {
	case "core-wording":
		row.PassiveCoreWordingEdits++
	case "core-decision-change":
		row.PassiveDecisionChanges++
	}
	stats.Cohorts[feedbackKey][cohort] = row
	return saveLocalRouteFeedbackStats(stats)
}

func recordRouteCohortOutcome(feedbackKey string, request providerGenerationRequest, outcome string) error {
	feedbackKey = strings.TrimSpace(feedbackKey)
	cohort := routeFeedbackCohortKeyForRequest(request)
	outcome = strings.TrimSpace(outcome)
	if feedbackKey == "" || cohort == "" || outcome == "" {
		return nil
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.Cohorts == nil {
		stats.Cohorts = map[string]map[string]localCohortFeedbackRow{}
	}
	if stats.Cohorts[feedbackKey] == nil {
		stats.Cohorts[feedbackKey] = map[string]localCohortFeedbackRow{}
	}
	row := stats.Cohorts[feedbackKey][cohort]
	switch outcome {
	case "used-this":
		row.OutcomeUsed++
	case "shared-this":
		row.OutcomeShared++
	case "replaced-this":
		row.OutcomeReplaced++
	default:
		return nil
	}
	stats.Cohorts[feedbackKey][cohort] = row
	return saveLocalRouteFeedbackStats(stats)
}

func recordPassiveRouteCohortOutcome(feedbackKey string, request providerGenerationRequest, outcome string) error {
	feedbackKey = strings.TrimSpace(feedbackKey)
	cohort := routeFeedbackCohortKeyForRequest(request)
	outcome = strings.TrimSpace(outcome)
	if feedbackKey == "" || cohort == "" || outcome == "" {
		return nil
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.Cohorts == nil {
		stats.Cohorts = map[string]map[string]localCohortFeedbackRow{}
	}
	if stats.Cohorts[feedbackKey] == nil {
		stats.Cohorts[feedbackKey] = map[string]localCohortFeedbackRow{}
	}
	row := stats.Cohorts[feedbackKey][cohort]
	switch outcome {
	case "used-this":
		row.PassiveReopened++
	case "shared-this":
		row.PassiveExportOpened++
	case "replaced-this":
		row.PassiveReplacedLater++
	default:
		return nil
	}
	stats.Cohorts[feedbackKey][cohort] = row
	return saveLocalRouteFeedbackStats(stats)
}

func routeFeedbackKeyForCurrentMode(toolMode string) string {
	toolMode = strings.TrimSpace(toolMode)
	if toolMode == "" {
		return ""
	}
	profile := currentDeviceProfile()
	modelLabel := modelLabelForToolMode(toolMode)
	parts := []string{
		toolMode,
		strings.TrimSpace(modelLabel),
	}
	if strings.HasPrefix(toolMode, "local-") {
		parts = append(parts,
			strings.TrimSpace(profile.DeviceClass),
			strings.TrimSpace(profile.LocalRuntimeClass),
			strings.TrimSpace(profile.LocalEndpointSignature),
		)
	}
	return strings.Join(parts, "|")
}

func loadAWSProfileValues(profile string) map[string]string {
	values := map[string]string{}
	for _, candidate := range []string{
		firstNonEmpty(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"), filepath.Join(homeDir(), ".aws", "credentials")),
		firstNonEmpty(os.Getenv("AWS_CONFIG_FILE"), filepath.Join(homeDir(), ".aws", "config")),
	} {
		for key, value := range parseAWSProfileFile(candidate, profile) {
			values[key] = value
		}
	}
	return values
}

func parseAWSProfileFile(path, profile string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	wanted := map[string]bool{
		profile:              true,
		"profile " + profile: true,
	}
	active := false
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			active = wanted[section]
			continue
		}
		if !active {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return values
}

func signAWSV4(req *http.Request, payload []byte, credentials awsCredentials, region, service string, now time.Time) {
	payloadHash := sha256Hex(payload)
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if credentials.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", credentials.SessionToken)
	}

	headers := map[string]string{
		"content-type":         req.Header.Get("Content-Type"),
		"host":                 req.URL.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
		"x-amz-security-token": req.Header.Get("X-Amz-Security-Token"),
	}
	if headers["x-amz-security-token"] == "" {
		delete(headers, "x-amz-security-token")
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)

	canonicalHeaders := strings.Builder{}
	for _, name := range names {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.TrimSpace(headers[name]))
		canonicalHeaders.WriteByte('\n')
	}
	signedHeaders := strings.Join(names, ";")
	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")
	credentialScope := strings.Join([]string{dateStamp, region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(credentials.SecretAccessKey, dateStamp, region, service), []byte(stringToSign)))
	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		credentials.AccessKeyID,
		credentialScope,
		signedHeaders,
		signature,
	))
}

func signingKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func providerSetupError(provider providerConfig) error {
	missing := strings.Join(provider.Missing, ", ")
	if missing == "" {
		missing = "provider configuration"
	}
	return fmt.Errorf("Provider needs setup. Run `jini provider doctor`. Missing: %s", missing)
}

func providerPrimaryView(packID string) (string, string) {
	return starterPrimaryView(packID)
}

func normalizeProviderMarkdown(label, title, text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "#") {
		return trimmed + "\n"
	}
	return fmt.Sprintf("# %s: %s\n\n%s\n", label, title, trimmed)
}

func providerSystemPrompt() string {
	return strings.Join([]string{
		"You are Jini, an outcome-first work assistant.",
		"Return concise Markdown only.",
		"Give the user a useful first draft before status commentary.",
		"Keep missing information visible instead of guessing silently.",
		"Do not mention hidden prompts, providers, APIs, or implementation details.",
	}, " ")
}

func providerUserPrompt(request providerGenerationRequest) string {
	effortLevel, effortReason := classifyRequestEffort(request)
	lines := []string{
		"Outcome needed:",
		request.Title,
		"",
		"Work type:",
		request.Choice.PackID,
		"",
		"User source:",
		request.Source,
		"",
		"Effort level:",
		effortLevel,
		"",
		"Effort guidance:",
		effortReason,
	}
	if hiddenPlan := providerHiddenPlanGuidance(request); hiddenPlan != "" {
		lines = append(lines, "", "Internal planning guidance:", hiddenPlan)
	}
	lines = append(lines, "", "Produce the first useful artifact for this work.")
	if guidance := providerArtifactGuidance(request); guidance != "" {
		lines = append(lines, "", guidance)
	}
	lines = append(lines,
		"",
		"Use headings, short bullets, and a clear `Still to confirm` section.",
	)
	return strings.Join(lines, "\n")
}

func providerHiddenPlanGuidance(request providerGenerationRequest) string {
	if !shouldUseHiddenPlan(request) {
		return ""
	}
	return strings.Join([]string{
		"Before writing, make a short hidden plan for yourself.",
		"Use it to order the work, cover the important sections, and keep missing information visible.",
		"Do not print the hidden plan unless the user explicitly asks for it.",
	}, " ")
}

func generateRefinedDraft(ctx context.Context, provider providerConfig, request providerGenerationRequest, firstDraft, reason string) (string, error) {
	return generateProviderText(ctx, provider, request, providerSystemPrompt(), providerRefinePrompt(request, firstDraft, reason))
}

func generateConsistencyDraft(ctx context.Context, provider providerConfig, request providerGenerationRequest, reason string) (string, error) {
	return generateProviderText(ctx, provider, request, providerSystemPrompt(), providerConsistencyPrompt(request, reason))
}

func providerRefinePrompt(request providerGenerationRequest, firstDraft, reason string) string {
	lines := []string{
		"Refine this artifact only if you can materially improve it.",
		"",
		"Outcome needed:",
		request.Title,
		"",
		"Refinement trigger:",
		firstNonEmpty(strings.TrimSpace(reason), "Improve clarity, completeness, and missing-information visibility."),
		"",
		"Original user source:",
		request.Source,
		"",
		"Current draft:",
		strings.TrimSpace(firstDraft),
		"",
		"Rules:",
		"- Keep the artifact concise and useful.",
		"- Preserve section headings when they already match the work type.",
		"- Improve structure, completeness, and faithfulness to the source.",
		"- Keep missing information visible instead of guessing.",
		"- Return Markdown only, with no commentary about the refinement process.",
	}
	return strings.Join(lines, "\n")
}

func providerConsistencyPrompt(request providerGenerationRequest, reason string) string {
	lines := []string{
		"Create an independent second draft for consistency checking.",
		"Do not reference any earlier draft or compare alternatives in the output.",
		"",
		"Outcome needed:",
		request.Title,
		"",
		"Consistency-check trigger:",
		firstNonEmpty(strings.TrimSpace(reason), "This request is high-stakes enough to justify a second independent pass."),
		"",
		"User source:",
		request.Source,
		"",
		"Rules:",
		"- Return the first useful artifact for this work type.",
		"- Keep missing information visible instead of guessing.",
		"- Return Markdown only.",
	}
	if guidance := providerArtifactGuidance(request); guidance != "" {
		lines = append(lines, "", guidance)
	}
	lines = append(lines, "", "Use headings, short bullets, and a clear `Still to confirm` section.")
	return strings.Join(lines, "\n")
}

func shouldUseHiddenPlan(request providerGenerationRequest) bool {
	features := classifyRouteFeatures(request)
	if features.DepthClass == "deep" {
		return true
	}
	switch features.RequestCohort {
	case "trip-itinerary", "build-readiness", "vendor-selection", "incident-closure":
		return true
	}
	return features.WorkClass == "planning" && features.EffortClass != "low"
}

func shouldRunSelectiveRefine(request providerGenerationRequest, decision routeDecision) (bool, string) {
	features := classifyRouteFeatures(request)
	if features.EffortClass == "low" {
		return false, ""
	}
	if lead := routeCandidateLeadMargin(request, decision.ToolMode); lead >= 0 && lead < 6 {
		return true, "The route scorer had low separation between the top choices, so Jini should use one focused revision pass."
	}
	if strings.HasPrefix(decision.ToolMode, "local-") && localBenchmarkBiasForFeatures(decision.ToolMode, features) < 4 {
		return true, "Local cohort confidence is still weak for this kind of work, so Jini should use one focused revision pass."
	}
	if shouldUseHiddenPlan(request) && features.EffortClass == "high" {
		return true, "This is multi-step work with higher effort, so Jini should tighten the first draft once before returning it."
	}
	return false, ""
}

func shouldRunSelectiveConsistencyCheck(request providerGenerationRequest, decision routeDecision) (bool, string) {
	features := classifyRouteFeatures(request)
	if features.EffortClass != "extra high" {
		return false, ""
	}
	return true, "This request is extra high effort, so Jini should verify it with a second independent draft before returning the artifact."
}

func classifyVerificationDecision(request providerGenerationRequest, decision routeDecision) (string, string) {
	features := classifyRouteFeatures(request)
	if features.EffortClass == "dynamic per request" {
		return "Dynamic per request", "Jini adjusts verification depth separately for each request."
	}
	if shouldCheck, reason := shouldRunSelectiveConsistencyCheck(request, decision); shouldCheck {
		return "Consistency check", reason
	}
	if shouldRefine, reason := shouldRunSelectiveRefine(request, decision); shouldRefine {
		return "Single pass + refine", firstNonEmpty(strings.TrimSpace(reason), "Jini adds one focused revision pass when confidence is weak or the work is higher-risk.")
	}
	if features.DepthClass == "deep" || features.EffortClass == "extra high" {
		return "Stronger route", "Jini increased verification depth by choosing a stronger route for deeper or higher-stakes work."
	}
	return "Single pass", "Jini used a single pass because the request did not justify extra verification cost."
}

func actualizeVerificationDecision(request providerGenerationRequest, decision routeDecision, consistencyUsed, refinedUsed bool) routeDecision {
	if !decision.Active {
		return decision
	}
	if consistencyUsed {
		decision.VerificationLevel = "Consistency check"
		if _, reason := shouldRunSelectiveConsistencyCheck(request, decision); strings.TrimSpace(reason) != "" {
			decision.VerificationReason = reason
		} else {
			decision.VerificationReason = "Jini compared a second independent draft before returning the artifact."
		}
		return decision
	}
	if refinedUsed {
		decision.VerificationLevel = "Single pass + refine"
		if _, reason := shouldRunSelectiveRefine(request, decision); strings.TrimSpace(reason) != "" {
			decision.VerificationReason = reason
		} else {
			decision.VerificationReason = "Jini added one focused revision pass before returning the artifact."
		}
		return decision
	}
	features := classifyRouteFeatures(request)
	if features.DepthClass == "deep" || features.EffortClass == "extra high" {
		decision.VerificationLevel = "Stronger route"
		decision.VerificationReason = "Jini increased verification depth by choosing a stronger route for deeper or higher-stakes work."
		return decision
	}
	decision.VerificationLevel = "Single pass"
	decision.VerificationReason = "Jini used a single pass because the request did not justify extra verification cost."
	return decision
}

func selectConsistencyWinner(request providerGenerationRequest, decision routeDecision, primary, alternate string) string {
	primaryScore := providerDraftQualityScore(request, decision, primary)
	alternateScore := providerDraftQualityScore(request, decision, alternate)
	switch {
	case alternateScore > primaryScore:
		return alternate
	case primaryScore > alternateScore:
		return primary
	case len(strings.TrimSpace(alternate)) > len(strings.TrimSpace(primary)):
		return alternate
	default:
		return primary
	}
}

func providerDraftQualityScore(request providerGenerationRequest, decision routeDecision, draft string) int {
	text := strings.TrimSpace(draft)
	if text == "" {
		return 0
	}
	profile := draftQualityProfileForRequest(request)
	lower := strings.ToLower(text)
	score := 0
	for _, heading := range profile.RequiredHeadings {
		if strings.Contains(text, heading) {
			score += profile.RequiredHeadingWeight
		}
	}
	for _, heading := range profile.PreferredHeadings {
		if strings.Contains(text, heading) {
			score += profile.PreferredHeadingWeight
		}
	}
	for _, signal := range profile.EvidenceSignals {
		if strings.Contains(lower, signal) {
			score += profile.EvidenceWeight
		}
	}
	score += strings.Count(text, "\n- ")
	score += strings.Count(text, "\n## ")
	if strings.Contains(lower, "still to confirm") || strings.Contains(lower, "open questions") || strings.Contains(lower, "still unclear") {
		score += profile.UncertaintyWeight
	}
	if len(text) > 400 {
		score += 2
	}
	score += providerDraftCohortLearningBias(profile, decision, text)
	return score
}

type draftQualityProfile struct {
	Cohort                 string
	ArtifactFamily         string
	ModalitySubtype        string
	RequiredHeadings       []string
	PreferredHeadings      []string
	EvidenceSignals        []string
	RequiredHeadingWeight  int
	PreferredHeadingWeight int
	EvidenceWeight         int
	UncertaintyWeight      int
}

func draftQualityProfileForRequest(request providerGenerationRequest) draftQualityProfile {
	features := classifyRouteFeatures(request)
	cohort := strings.TrimSpace(classifyRequestCohort(request))
	family := strings.TrimSpace(classifyArtifactFamily(request))
	profile := draftQualityProfile{
		Cohort:                 cohort,
		ArtifactFamily:         family,
		ModalitySubtype:        strings.TrimSpace(features.ModalitySubtype),
		RequiredHeadingWeight:  6,
		PreferredHeadingWeight: 3,
		EvidenceWeight:         0,
		UncertaintyWeight:      4,
		RequiredHeadings:       []string{"## Still to confirm"},
	}
	if base := starterDraftQualityProfile(request.Choice.PackID); len(base.RequiredHeadings) > 0 || len(base.PreferredHeadings) > 0 {
		profile.RequiredHeadings = append([]string{}, base.RequiredHeadings...)
		profile.PreferredHeadings = append([]string{}, base.PreferredHeadings...)
		if base.RequiredHeadingWeight != 0 {
			profile.RequiredHeadingWeight = base.RequiredHeadingWeight
		}
		if base.PreferredHeadingWeight != 0 {
			profile.PreferredHeadingWeight = base.PreferredHeadingWeight
		}
		if base.EvidenceWeight != 0 {
			profile.EvidenceWeight = base.EvidenceWeight
		}
		if base.UncertaintyWeight != 0 {
			profile.UncertaintyWeight = base.UncertaintyWeight
		}
		if len(base.EvidenceSignals) > 0 {
			profile.EvidenceSignals = append([]string{}, base.EvidenceSignals...)
		}
	}
	if profile.Cohort == "multimodal-extract" || profile.ArtifactFamily == "multimodal-extract" || features.ModalityClass == "multimodal" {
		profile.RequiredHeadings = []string{
			"## Extracted evidence",
			"## What the source shows",
			"## Still unclear",
		}
		profile.PreferredHeadings = []string{
			"## Recommended next move",
			"## Confidence notes",
		}
		profile.EvidenceSignals = []string{
			"image", "screenshot", "pdf", "audio", "recording", "document",
			"source", "evidence", "shows", "visible", "transcript", "scan",
		}
		profile.RequiredHeadingWeight = 8
		profile.PreferredHeadingWeight = 4
		profile.EvidenceWeight = 5
		profile.UncertaintyWeight = 5
		switch profile.ModalitySubtype {
		case "pdf-scan":
			profile.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What the document shows",
				"## Still unclear",
			}
			profile.PreferredHeadings = []string{
				"## Recommended next move",
				"## OCR or confidence notes",
			}
			profile.EvidenceSignals = []string{
				"pdf", "document", "page", "scan", "ocr", "text", "field",
				"label", "signature", "table", "section",
			}
		case "image-screenshot":
			profile.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What is visible",
				"## Still unclear",
			}
			profile.PreferredHeadings = []string{
				"## Recommended next move",
				"## Confidence notes",
			}
			profile.EvidenceSignals = []string{
				"image", "screenshot", "visible", "screen", "button", "label",
				"panel", "photo", "diagram", "highlight",
			}
		case "audio-transcript":
			profile.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What the recording says",
				"## Still unclear",
			}
			profile.PreferredHeadings = []string{
				"## Recommended next move",
				"## Confidence notes",
			}
			profile.EvidenceSignals = []string{
				"audio", "recording", "voice", "transcript", "speaker",
				"said", "heard", "timecode", "quote",
			}
		}
	}
	return profile
}

func providerDraftCohortLearningBias(profile draftQualityProfile, decision routeDecision, text string) int {
	cohort := routeFeedbackCohortKeyForProfile(profile)
	if cohort == "" {
		return 0
	}
	routeSpecific, routeFound := routeSpecificDraftCohortRow(decision, cohort)
	if routeFound {
		return providerDraftCohortRowBias(profile, routeSpecific, text)
	}
	return providerDraftCohortRowBias(profile, aggregateDraftCohortRow(cohort), text)
}

func routeSpecificDraftCohortRow(decision routeDecision, cohort string) (localCohortFeedbackRow, bool) {
	feedbackKey := strings.TrimSpace(decision.FeedbackKey)
	if feedbackKey == "" {
		feedbackKey = routeFeedbackKeyForDecision(decision)
	}
	if feedbackKey == "" {
		return localCohortFeedbackRow{}, false
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.Cohorts == nil {
		return localCohortFeedbackRow{}, false
	}
	row, ok := stats.Cohorts[feedbackKey][cohort]
	return row, ok
}

func aggregateDraftCohortRow(cohort string) localCohortFeedbackRow {
	report := loadLocalRuntimeCapabilities()
	total := localCohortFeedbackRow{}
	if len(report.CohortFeedback) == 0 {
		return total
	}
	for _, modeFeedback := range report.CohortFeedback {
		row, ok := modeFeedback[cohort]
		if !ok {
			continue
		}
		total.Upvotes += row.Upvotes
		total.Downvotes += row.Downvotes
		total.AcceptedAsIs += row.AcceptedAsIs
		total.NeededLightEdits += row.NeededLightEdits
		total.NotUseful += row.NotUseful
		total.OutcomeUsed += row.OutcomeUsed
		total.OutcomeShared += row.OutcomeShared
		total.OutcomeReplaced += row.OutcomeReplaced
		total.PassiveReopened += row.PassiveReopened
		total.PassiveExportOpened += row.PassiveExportOpened
		total.PassiveReplacedLater += row.PassiveReplacedLater
		total.PassiveAcceptedAsIs += row.PassiveAcceptedAsIs
		total.PassiveNeededLightEdits += row.PassiveNeededLightEdits
		total.PassiveNeededHeavyEdits += row.PassiveNeededHeavyEdits
		total.PassiveHeaderOnlyEdits += row.PassiveHeaderOnlyEdits
		total.PassiveCoreSectionEdits += row.PassiveCoreSectionEdits
		total.PassiveCoreWordingEdits += row.PassiveCoreWordingEdits
		total.PassiveDecisionChanges += row.PassiveDecisionChanges
	}
	return total
}

func providerDraftCohortRowBias(profile draftQualityProfile, total localCohortFeedbackRow, text string) int {
	bias := 0
	lower := strings.ToLower(text)
	if total.OutcomeReplaced+total.PassiveDecisionChanges > total.AcceptedAsIs+total.OutcomeShared {
		if strings.Contains(lower, "still to confirm") || strings.Contains(lower, "open questions") {
			bias += 4
		}
		if strings.Contains(lower, "recommended next move") || strings.Contains(lower, "next move") || strings.Contains(lower, "recommended first slice") {
			bias += 4
		}
	}
	if total.AcceptedAsIs+total.OutcomeShared > total.OutcomeReplaced+total.PassiveNeededHeavyEdits {
		if len(strings.TrimSpace(text)) > 350 {
			bias += 2
		}
	}
	if total.PassiveDecisionChanges > total.PassiveHeaderOnlyEdits {
		if strings.Contains(lower, "decisions captured") || strings.Contains(lower, "must clear before build") || strings.Contains(lower, "risks") {
			bias += 3
		}
	}
	if profile.Cohort == "multimodal-extract" || profile.ArtifactFamily == "multimodal-extract" {
		if total.OutcomeReplaced+total.PassiveDecisionChanges > total.AcceptedAsIs+total.OutcomeShared {
			if strings.Contains(lower, "extracted evidence") || strings.Contains(lower, "what the source shows") {
				bias += 4
			}
			if strings.Contains(lower, "still unclear") || strings.Contains(lower, "confidence notes") {
				bias += 3
			}
		}
		for _, signal := range []string{"image", "screenshot", "pdf", "audio", "recording", "transcript", "scan"} {
			if strings.Contains(lower, signal) {
				bias++
			}
		}
		switch profile.ModalitySubtype {
		case "pdf-scan":
			if strings.Contains(lower, "what the document shows") || strings.Contains(lower, "ocr or confidence notes") {
				bias += 2
			}
		case "image-screenshot":
			if strings.Contains(lower, "what is visible") {
				bias += 2
			}
		case "audio-transcript":
			if strings.Contains(lower, "what the recording says") {
				bias += 2
			}
		}
	}
	return bias
}

func providerArtifactGuidance(request providerGenerationRequest) string {
	if cohort := strings.TrimSpace(classifyRequestCohort(request)); cohort == "multimodal-extract" {
		switch classifyRouteModalitySubtype(request) {
		case "pdf-scan":
			return strings.Join([]string{
				"Return an evidence-grounded document extraction artifact, not a generic summary.",
				"Use these sections exactly:",
				"- `## Extracted evidence`",
				"- `## What the document shows`",
				"- `## Still unclear`",
				"- `## Recommended next move`",
				"Call out what is visible in the PDF or scan, what OCR may have missed, and what still needs verification.",
			}, "\n")
		case "image-screenshot":
			return strings.Join([]string{
				"Return an evidence-grounded image or screenshot extraction artifact, not a generic summary.",
				"Use these sections exactly:",
				"- `## Extracted evidence`",
				"- `## What is visible`",
				"- `## Still unclear`",
				"- `## Recommended next move`",
				"Call out what is visually present, what may be obscured, and what still needs verification.",
			}, "\n")
		case "audio-transcript":
			return strings.Join([]string{
				"Return an evidence-grounded audio extraction artifact, not a generic summary.",
				"Use these sections exactly:",
				"- `## Extracted evidence`",
				"- `## What the recording says`",
				"- `## Still unclear`",
				"- `## Recommended next move`",
				"Call out what is directly supported by the recording or transcript, what may have been misheard, and what still needs verification.",
			}, "\n")
		}
		return strings.Join([]string{
			"Return an evidence-grounded extraction artifact, not a generic summary.",
			"Use these sections exactly:",
			"- `## Extracted evidence`",
			"- `## What the source shows`",
			"- `## Still unclear`",
			"- `## Recommended next move`",
			"Call out what came from the source, what is ambiguous, and what still needs verification.",
		}, "\n")
	}
	if guidance := starterProviderArtifactGuidance(request.Choice.PackID); guidance != "" {
		return guidance
	}
	return strings.Join([]string{
		"Shape the answer as the first useful artifact for this work type.",
		"When the source already includes a named reference or URL that belongs in the artifact, preserve it as one smart Markdown link on first mention.",
	}, "\n")
}

func homeDir() string {
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return home
	}
	return "."
}
