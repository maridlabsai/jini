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

type awsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func maybeWriteProviderFirstDraft(ctx context.Context, choice starterChoice, workDir, title, source string) error {
	text, used, err := generateWithConfiguredProvider(ctx, providerGenerationRequest{
		Choice: choice,
		Title:  title,
		Source: source,
	})
	if err != nil {
		return err
	}
	if !used {
		return nil
	}

	path, label := providerPrimaryView(choice.PackID)
	if path == "" {
		return nil
	}
	content := normalizeProviderMarkdown(label, title, text)
	return os.WriteFile(filepath.Join(workDir, "views", path), []byte(content), 0o644)
}

func generateWithConfiguredProvider(ctx context.Context, request providerGenerationRequest) (string, bool, error) {
	provider := detectProvider()
	if provider.ID == "local-preview" {
		return "", false, nil
	}
	if provider.Status != "ok" {
		return "", true, providerSetupError(provider)
	}

	switch provider.ID {
	case "azure-openai":
		text, err := generateWithAzureOpenAI(ctx, request)
		return text, true, err
	case "bedrock":
		text, err := generateWithBedrock(ctx, request)
		return text, true, err
	case "anthropic":
		text, err := generateWithAnthropic(ctx, request)
		return text, true, err
	default:
		return "", true, providerSetupError(provider)
	}
}

func generateWithAnthropic(ctx context.Context, request providerGenerationRequest) (string, error) {
	modelID, _, modelIssue := resolveAnthropicModel()
	if modelIssue != "" {
		return "", providerSetupError(providerConfig{Missing: []string{modelIssue}})
	}
	baseURL := strings.TrimRight(strings.TrimSpace(firstNonEmpty(os.Getenv("ANTHROPIC_BASE_URL"), "https://api.anthropic.com")), "/")
	target := baseURL + "/v1/messages"
	payload := map[string]any{
		"model":      modelID,
		"system":     providerSystemPrompt(),
		"max_tokens": 1600,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "text", "text": providerUserPrompt(request)},
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
	req.Header.Set("x-api-key", strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")))
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

func generateWithAzureOpenAI(ctx context.Context, request providerGenerationRequest) (string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(os.Getenv("AZURE_OPENAI_ENDPOINT")), "/")
	deployment := strings.TrimSpace(os.Getenv("AZURE_OPENAI_DEPLOYMENT"))
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
			{"role": "system", "content": providerSystemPrompt()},
			{"role": "user", "content": providerUserPrompt(request)},
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
	req.Header.Set("api-key", strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_KEY")))

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

func generateWithBedrock(ctx context.Context, request providerGenerationRequest) (string, error) {
	region := strings.TrimSpace(resolveAWSRegion())
	if region == "" {
		return "", providerSetupError(providerConfig{Missing: []string{"AWS_REGION or AWS_DEFAULT_REGION"}})
	}
	modelID, _ := resolveBedrockModel()
	credentials, err := loadAWSCredentials()
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(strings.TrimSpace(os.Getenv("JINI_BEDROCK_ENDPOINT")), "/")
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}
	target := endpoint + "/model/" + url.PathEscape(modelID) + "/converse"
	payload := map[string]any{
		"system": []map[string]string{
			{"text": providerSystemPrompt()},
		},
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]string{
					{"text": providerUserPrompt(request)},
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

func loadAWSCredentials() (awsCredentials, error) {
	if strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID")) != "" && strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY")) != "" {
		return awsCredentials{
			AccessKeyID:     strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID")),
			SecretAccessKey: strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY")),
			SessionToken:    strings.TrimSpace(os.Getenv("AWS_SESSION_TOKEN")),
		}, nil
	}
	profile := strings.TrimSpace(os.Getenv("AWS_PROFILE"))
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
	if region := strings.TrimSpace(firstNonEmpty(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION"))); region != "" {
		return region
	}
	profile := strings.TrimSpace(os.Getenv("AWS_PROFILE"))
	if profile == "" {
		return ""
	}
	return strings.TrimSpace(loadAWSProfileValues(profile)["region"])
}

func configuredProviderMode() string {
	raw := normalizeName(firstNonEmpty(strings.TrimSpace(os.Getenv("JINI_PROVIDER")), "auto"))
	switch raw {
	case "", "auto":
		return "auto"
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
	for _, mode := range []string{"anthropic", "azure-openai", "bedrock"} {
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
	case "local-preview":
		return detectLocalPreviewProvider()
	default:
		return providerConfig{
			ID:      mode,
			Label:   titleCase(mode),
			Status:  "needs setup",
			Missing: []string{"Supported JINI_PROVIDER value: auto, claude, azure-openai, bedrock, or local-preview"},
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
	return strings.TrimSpace(firstNonEmpty(os.Getenv("JINI_MODEL"), "auto"))
}

func azureModelSettingLine() string {
	raw := configuredModelInput()
	if normalizeName(raw) == "auto" {
		return ""
	}
	return "JINI_MODEL: " + raw + " -> deployment decides the actual Azure model"
}

func bedrockModelSettingLine(modelID, modelLabel string) string {
	raw := strings.TrimSpace(os.Getenv("BEDROCK_MODEL_ID"))
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
	raw := strings.TrimSpace(firstNonEmpty(os.Getenv("ANTHROPIC_MODEL"), os.Getenv("JINI_MODEL")))
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
	rawID := strings.TrimSpace(os.Getenv("BEDROCK_MODEL_ID"))
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
	rawModel := strings.TrimSpace(firstNonEmpty(os.Getenv("ANTHROPIC_MODEL"), os.Getenv("JINI_MODEL")))
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
	switch packID {
	case "travel-plan":
		return "itinerary.md", "Itinerary"
	case "meeting-followup":
		return "followup.md", "Sendable Follow-Up"
	case "research-prd":
		return "prd.md", "Build-Readiness Check"
	case "vendor-selection":
		return "recommendation-memo.md", "Recommendation Memo"
	case "incident-response":
		return "closure-checklist.md", "Closure Checklist"
	default:
		return "first-useful-pass.md", "First Useful Pass"
	}
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
	return strings.Join([]string{
		"Outcome needed:",
		request.Title,
		"",
		"Work type:",
		request.Choice.PackID,
		"",
		"User source:",
		request.Source,
		"",
		"Produce the first useful artifact for this work.",
		"Use headings, short bullets, and a clear `Still to confirm` section.",
	}, "\n")
}

func homeDir() string {
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return home
	}
	return "."
}
