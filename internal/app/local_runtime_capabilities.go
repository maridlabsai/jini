package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	localRuntimeCapabilitiesSchemaVersion = "0.4.0"
	localRuntimeCapabilitiesTTL           = 24 * time.Hour
	localRuntimeHistoryLimit              = 6
)

type localRuntimeCapabilities struct {
	SchemaVersion             string                                       `json:"schema_version"`
	ContextType               string                                       `json:"context_type"`
	CapturedAt                string                                       `json:"captured_at"`
	JiniVersion               string                                       `json:"jini_version"`
	CapabilityRegistryVersion string                                       `json:"capability_registry_version"`
	DeviceProbeFingerprint    string                                       `json:"device_probe_fingerprint"`
	LocalEndpointSignature    string                                       `json:"local_endpoint_signature"`
	LocalRuntimeClass         string                                       `json:"local_runtime_class"`
	Adapters                  map[string]localAdapterCapability            `json:"adapters"`
	History                   map[string][]localAdapterSample              `json:"history,omitempty"`
	CohortHistory             map[string]map[string][]localAdapterSample   `json:"cohort_history,omitempty"`
	CohortFeedback            map[string]map[string]localCohortFeedbackRow `json:"cohort_feedback,omitempty"`
}

type localCohortFeedbackRow struct {
	Upvotes                 int `json:"upvotes"`
	Downvotes               int `json:"downvotes"`
	AcceptedAsIs            int `json:"accepted_as_is,omitempty"`
	NeededLightEdits        int `json:"needed_light_edits,omitempty"`
	NotUseful               int `json:"not_useful,omitempty"`
	OutcomeUsed             int `json:"outcome_used,omitempty"`
	OutcomeShared           int `json:"outcome_shared,omitempty"`
	OutcomeReplaced         int `json:"outcome_replaced,omitempty"`
	PassiveReopened         int `json:"passive_reopened,omitempty"`
	PassiveExportOpened     int `json:"passive_export_opened,omitempty"`
	PassiveReplacedLater    int `json:"passive_replaced_later,omitempty"`
	PassiveAcceptedAsIs     int `json:"passive_accepted_as_is,omitempty"`
	PassiveNeededLightEdits int `json:"passive_needed_light_edits,omitempty"`
	PassiveNeededHeavyEdits int `json:"passive_needed_heavy_edits,omitempty"`
	PassiveHeaderOnlyEdits  int `json:"passive_header_only_edits,omitempty"`
	PassiveCoreSectionEdits int `json:"passive_core_section_edits,omitempty"`
	PassiveCoreWordingEdits int `json:"passive_core_wording_edits,omitempty"`
	PassiveDecisionChanges  int `json:"passive_decision_changes,omitempty"`
}

type localAdapterCapability struct {
	AdapterID             string  `json:"adapter_id"`
	ModelID               string  `json:"model_id"`
	Status                string  `json:"status"`
	LatencyMS             int     `json:"latency_ms"`
	WarmLatencyMS         int     `json:"warm_latency_ms,omitempty"`
	ColdStartCostMS       int     `json:"cold_start_cost_ms,omitempty"`
	OutputChars           int     `json:"output_chars"`
	OutputTokens          int     `json:"output_tokens,omitempty"`
	TokensPerSecond       float64 `json:"tokens_per_second,omitempty"`
	QualityClass          string  `json:"quality_class"`
	StructuredReliability string  `json:"structured_reliability,omitempty"`
	Error                 string  `json:"error,omitempty"`
	BenchmarkedAt         string  `json:"benchmarked_at"`
}

type localAdapterSample struct {
	ModelID               string  `json:"model_id"`
	Status                string  `json:"status"`
	LatencyMS             int     `json:"latency_ms"`
	ColdStartCostMS       int     `json:"cold_start_cost_ms,omitempty"`
	TokensPerSecond       float64 `json:"tokens_per_second,omitempty"`
	QualityClass          string  `json:"quality_class"`
	StructuredReliability string  `json:"structured_reliability,omitempty"`
	BenchmarkedAt         string  `json:"benchmarked_at"`
}

type localBenchmarkProbe struct {
	LatencyMS    int
	Output       string
	OutputChars  int
	OutputTokens int
	StatusCode   int
	Error        string
}

var (
	localBenchmarkWarmMu       sync.Mutex
	localBenchmarkWarmInFlight bool
)

func localRuntimeCapabilitiesPath() string {
	return filepath.Join(sessionStateRoot(), "local-runtime-capabilities.json")
}

func loadLocalRuntimeCapabilities() localRuntimeCapabilities {
	data, err := os.ReadFile(localRuntimeCapabilitiesPath())
	if err != nil {
		return localRuntimeCapabilities{Adapters: map[string]localAdapterCapability{}}
	}
	var payload localRuntimeCapabilities
	if err := json.Unmarshal(data, &payload); err != nil {
		return localRuntimeCapabilities{Adapters: map[string]localAdapterCapability{}}
	}
	if payload.Adapters == nil {
		payload.Adapters = map[string]localAdapterCapability{}
	}
	if payload.History == nil {
		payload.History = map[string][]localAdapterSample{}
	}
	if payload.CohortHistory == nil {
		payload.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if payload.CohortFeedback == nil {
		payload.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	return payload
}

func saveLocalRuntimeCapabilities(report localRuntimeCapabilities) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	if report.SchemaVersion == "" {
		report.SchemaVersion = localRuntimeCapabilitiesSchemaVersion
	}
	if report.ContextType == "" {
		report.ContextType = "JiniLocalRuntimeCapabilities"
	}
	if report.CapturedAt == "" {
		report.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if report.JiniVersion == "" {
		report.JiniVersion = currentJiniVersion()
	}
	if report.CapabilityRegistryVersion == "" {
		report.CapabilityRegistryVersion = deviceCapabilityRegistryVersion
	}
	if report.DeviceProbeFingerprint == "" {
		report.DeviceProbeFingerprint = currentProbeFingerprint()
	}
	if report.LocalEndpointSignature == "" {
		report.LocalEndpointSignature = normalizedLocalEndpointSignature()
	}
	if report.LocalRuntimeClass == "" {
		report.LocalRuntimeClass = detectLocalRuntimeClass()
	}
	if report.Adapters == nil {
		report.Adapters = map[string]localAdapterCapability{}
	}
	if report.History == nil {
		report.History = map[string][]localAdapterSample{}
	}
	if report.CohortHistory == nil {
		report.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if report.CohortFeedback == nil {
		report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	report.History = mergedLocalRuntimeHistory(loadLocalRuntimeCapabilities(), report)
	report.CohortHistory = mergedLocalRuntimeCohortHistory(loadLocalRuntimeCapabilities(), report)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(localRuntimeCapabilitiesPath(), append(data, '\n'), 0o600)
}

func localRuntimeCapabilitiesAreFresh(report localRuntimeCapabilities) bool {
	if report.SchemaVersion != localRuntimeCapabilitiesSchemaVersion {
		return false
	}
	if report.ContextType != "JiniLocalRuntimeCapabilities" {
		return false
	}
	if report.JiniVersion != currentJiniVersion() {
		return false
	}
	if report.CapabilityRegistryVersion != deviceCapabilityRegistryVersion {
		return false
	}
	if report.DeviceProbeFingerprint != currentProbeFingerprint() {
		return false
	}
	if report.LocalEndpointSignature != normalizedLocalEndpointSignature() {
		return false
	}
	if report.LocalRuntimeClass != detectLocalRuntimeClass() {
		return false
	}
	if report.History == nil {
		return false
	}
	stamp, err := time.Parse(time.RFC3339, strings.TrimSpace(report.CapturedAt))
	if err != nil {
		return false
	}
	return time.Since(stamp) <= localRuntimeCapabilitiesTTL
}

func currentLocalRuntimeCapabilities(ctx context.Context) localRuntimeCapabilities {
	if strings.TrimSpace(configValue("JINI_SKIP_LOCAL_BENCHMARK")) != "" {
		return loadLocalRuntimeCapabilities()
	}
	report := loadLocalRuntimeCapabilities()
	if localRuntimeCapabilitiesAreFresh(report) {
		return report
	}
	if !isActuallyLocalRuntimeClass(detectLocalRuntimeClass()) {
		return localRuntimeCapabilities{
			Adapters:       map[string]localAdapterCapability{},
			CohortHistory:  map[string]map[string][]localAdapterSample{},
			CohortFeedback: map[string]map[string]localCohortFeedbackRow{},
		}
	}
	fresh := benchmarkLocalRuntimeCapabilities(ctx)
	_ = saveLocalRuntimeCapabilities(fresh)
	return fresh
}

func maybeWarmLocalRuntimeCapabilitiesAsync() bool {
	if strings.TrimSpace(configValue("JINI_SKIP_LOCAL_BENCHMARK")) != "" {
		return false
	}
	if !localSLMRuntimeReady() {
		return false
	}
	report := loadLocalRuntimeCapabilities()
	if localRuntimeCapabilitiesAreFresh(report) {
		return false
	}

	localBenchmarkWarmMu.Lock()
	if localBenchmarkWarmInFlight {
		localBenchmarkWarmMu.Unlock()
		return false
	}
	localBenchmarkWarmInFlight = true
	localBenchmarkWarmMu.Unlock()

	go func() {
		defer func() {
			localBenchmarkWarmMu.Lock()
			localBenchmarkWarmInFlight = false
			localBenchmarkWarmMu.Unlock()
		}()
		report := loadLocalRuntimeCapabilities()
		if localRuntimeCapabilitiesAreFresh(report) {
			return
		}
		if !localSLMRuntimeReady() {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		fresh := benchmarkLocalRuntimeCapabilities(ctx)
		_ = saveLocalRuntimeCapabilities(fresh)
	}()
	return true
}

func benchmarkLocalRuntimeCapabilities(ctx context.Context) localRuntimeCapabilities {
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters:                  map[string]localAdapterCapability{},
		History:                   map[string][]localAdapterSample{},
		CohortHistory:             map[string]map[string][]localAdapterSample{},
		CohortFeedback:            map[string]map[string]localCohortFeedbackRow{},
	}
	if !isActuallyLocalRuntimeClass(report.LocalRuntimeClass) {
		return report
	}
	for _, mode := range localBenchmarkableAdapterModes() {
		descriptor, ok := adapterDescriptorForMode(mode)
		if !ok {
			continue
		}
		report.Adapters[mode] = benchmarkLocalAdapter(ctx, descriptor)
	}
	return report
}

func benchmarkLocalAdapter(ctx context.Context, descriptor adapterDescriptor) localAdapterCapability {
	modelID, _ := resolveLocalSLMModelForToolMode(descriptor.ID)
	now := time.Now().UTC().Format(time.RFC3339)
	base := localAdapterCapability{
		AdapterID:             descriptor.ID,
		ModelID:               strings.TrimSpace(modelID),
		BenchmarkedAt:         now,
		StructuredReliability: "unknown",
	}
	if strings.TrimSpace(modelID) == "" {
		base.Status = "not-configured"
		base.QualityClass = "unknown"
		return base
	}
	endpoint := strings.TrimRight(strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")), "/")
	if endpoint == "" {
		base.Status = "not-configured"
		base.QualityClass = "unknown"
		return base
	}
	prompt := localBenchmarkPromptForMode(descriptor.ID)
	coldProbe := benchmarkLocalAdapterProbe(ctx, endpoint, modelID, prompt)
	if coldProbe.Error != "" {
		base.Status = "failed"
		base.QualityClass = "failed"
		base.StructuredReliability = "failed"
		base.Error = coldProbe.Error
		base.LatencyMS = coldProbe.LatencyMS
		return base
	}
	warmProbe := benchmarkLocalAdapterProbe(ctx, endpoint, modelID, prompt)
	if warmProbe.Error != "" {
		base.Status = "degraded"
		base.QualityClass = classifyBenchmarkQuality(descriptor.ID, coldProbe.Output)
		base.StructuredReliability = "weak"
		base.Error = warmProbe.Error
		base.LatencyMS = coldProbe.LatencyMS
		base.OutputChars = coldProbe.OutputChars
		base.OutputTokens = coldProbe.OutputTokens
		base.TokensPerSecond = tokensPerSecond(base.OutputTokens, base.LatencyMS)
		return base
	}

	coldQuality := classifyBenchmarkQuality(descriptor.ID, coldProbe.Output)
	warmQuality := classifyBenchmarkQuality(descriptor.ID, warmProbe.Output)
	base.LatencyMS = warmProbe.LatencyMS
	base.WarmLatencyMS = warmProbe.LatencyMS
	base.ColdStartCostMS = coldStartCost(coldProbe.LatencyMS, warmProbe.LatencyMS)
	base.OutputChars = warmProbe.OutputChars
	base.OutputTokens = warmProbe.OutputTokens
	base.TokensPerSecond = tokensPerSecond(base.OutputTokens, base.LatencyMS)
	base.QualityClass = aggregateBenchmarkQuality(coldQuality, warmQuality)
	base.StructuredReliability = classifyStructuredReliability(coldQuality, warmQuality)
	switch base.StructuredReliability {
	case "strong", "usable":
		switch base.QualityClass {
		case "strong", "usable":
			base.Status = "ok"
		case "weak":
			base.Status = "degraded"
		default:
			base.Status = "failed"
		}
	case "weak":
		base.Status = "degraded"
	default:
		base.Status = "failed"
	}
	if base.Status == "failed" && base.Error == "" {
		base.Error = "empty or malformed output"
	}
	return base
}

func benchmarkLocalAdapterProbe(ctx context.Context, endpoint, modelID, prompt string) localBenchmarkProbe {
	payload := map[string]any{
		"model": modelID,
		"messages": []map[string]string{
			{"role": "system", "content": "Reply briefly and follow the requested format exactly."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.0,
		"max_tokens":  140,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return localBenchmarkProbe{Error: "marshal failed"}
	}
	requestCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return localBenchmarkProbe{Error: "request build failed"}
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey := strings.TrimSpace(configValue("JINI_LOCAL_SLM_API_KEY")); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	start := time.Now()
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return localBenchmarkProbe{LatencyMS: int(time.Since(start).Milliseconds()), Error: "request failed"}
	}
	defer resp.Body.Close()
	latencyMS := int(time.Since(start).Milliseconds())
	if latencyMS <= 0 {
		latencyMS = 1
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return localBenchmarkProbe{LatencyMS: latencyMS, StatusCode: resp.StatusCode, Error: fmt.Sprintf("http %d", resp.StatusCode)}
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return localBenchmarkProbe{LatencyMS: latencyMS, Error: "decode failed"}
	}
	content := ""
	if len(decoded.Choices) > 0 {
		content = strings.TrimSpace(decoded.Choices[0].Message.Content)
	}
	tokens := decoded.Usage.CompletionTokens
	if tokens <= 0 {
		tokens = estimateTokenCount(content)
	}
	return localBenchmarkProbe{
		LatencyMS:    latencyMS,
		Output:       content,
		OutputChars:  len(content),
		OutputTokens: tokens,
		StatusCode:   resp.StatusCode,
	}
}

func localBenchmarkPromptForMode(mode string) string {
	switch mode {
	case "local-fast":
		return "Reply with exactly this line: READY FAST ALPHA BETA GAMMA DELTA ETA THETA IOTA"
	case "local-workhorse":
		return "Reply as valid JSON exactly like this shape: {\"status\":\"ready\",\"items\":[\"one\",\"two\",\"three\",\"four\",\"five\",\"six\"]}"
	case "local-deep":
		return "Reply with exactly five numbered lines, each 3-6 words long."
	case "local-multimodal":
		return "Reply with exactly this line: READY MULTIMODAL VISION AUDIO DOCUMENT IMAGE PARSE"
	default:
		return "Reply with exactly this line: READY FAST ALPHA BETA GAMMA DELTA ETA THETA IOTA"
	}
}

func classifyBenchmarkQuality(mode, content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "failed"
	}
	switch mode {
	case "local-fast":
		if content == "READY FAST ALPHA BETA GAMMA DELTA ETA THETA IOTA" {
			return "strong"
		}
		if strings.Contains(content, "READY") {
			return "usable"
		}
	case "local-workhorse":
		var payload struct {
			Status string   `json:"status"`
			Items  []string `json:"items"`
		}
		if err := json.Unmarshal([]byte(content), &payload); err == nil && payload.Status == "ready" && len(payload.Items) >= 6 {
			return "strong"
		}
		if strings.Contains(strings.ToLower(content), "ready") {
			return "usable"
		}
	case "local-deep":
		lines := strings.Split(content, "\n")
		count := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") || strings.HasPrefix(line, "5.") {
				count++
			}
		}
		if count >= 5 {
			return "strong"
		}
		if len(lines) >= 3 {
			return "usable"
		}
	case "local-multimodal":
		if content == "READY MULTIMODAL VISION AUDIO DOCUMENT IMAGE PARSE" {
			return "strong"
		}
		if strings.Contains(strings.ToLower(content), "ready") {
			return "usable"
		}
	}
	if len(content) >= 8 {
		return "weak"
	}
	return "failed"
}

func aggregateBenchmarkQuality(coldQuality, warmQuality string) string {
	if coldQuality == "strong" && warmQuality == "strong" {
		return "strong"
	}
	if qualityRank(coldQuality) >= qualityRank("usable") && qualityRank(warmQuality) >= qualityRank("usable") {
		return "usable"
	}
	if qualityRank(coldQuality) >= qualityRank("weak") || qualityRank(warmQuality) >= qualityRank("weak") {
		return "weak"
	}
	return "failed"
}

func classifyStructuredReliability(coldQuality, warmQuality string) string {
	switch {
	case coldQuality == "strong" && warmQuality == "strong":
		return "strong"
	case qualityRank(coldQuality) >= qualityRank("usable") && qualityRank(warmQuality) >= qualityRank("usable"):
		return "usable"
	case qualityRank(coldQuality) >= qualityRank("weak") || qualityRank(warmQuality) >= qualityRank("weak"):
		return "weak"
	default:
		return "failed"
	}
}

func qualityRank(value string) int {
	switch value {
	case "strong":
		return 4
	case "usable":
		return 3
	case "weak":
		return 2
	case "failed":
		return 1
	default:
		return 0
	}
}

func estimateTokenCount(content string) int {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0
	}
	runes := len([]rune(content)) / 4
	words := len(strings.Fields(content))
	if runes > words {
		return runes
	}
	if words > 0 {
		return words
	}
	return 1
}

func tokensPerSecond(tokens, latencyMS int) float64 {
	if tokens <= 0 || latencyMS <= 0 {
		return 0
	}
	return float64(tokens) / (float64(latencyMS) / 1000.0)
}

func coldStartCost(coldLatencyMS, warmLatencyMS int) int {
	if coldLatencyMS <= 0 || warmLatencyMS <= 0 || coldLatencyMS <= warmLatencyMS {
		return 0
	}
	return coldLatencyMS - warmLatencyMS
}

func localBenchmarkSampleBias(row localAdapterCapability) int {
	switch row.Status {
	case "failed":
		return -70
	case "degraded":
		return -18
	case "not-configured":
		return -40
	}
	bias := 0
	switch row.QualityClass {
	case "strong":
		bias += 10
	case "usable":
		bias += 5
	case "weak":
		bias -= 8
	}
	switch row.StructuredReliability {
	case "strong":
		bias += 10
	case "usable":
		bias += 4
	case "weak":
		bias -= 12
	case "failed":
		bias -= 20
	}
	switch {
	case row.LatencyMS <= 0:
	case row.LatencyMS <= 900:
		bias += 10
	case row.LatencyMS <= 1800:
		bias += 6
	case row.LatencyMS <= 3200:
		bias += 1
	default:
		bias -= 10
	}
	switch {
	case row.TokensPerSecond >= 24:
		bias += 10
	case row.TokensPerSecond >= 14:
		bias += 6
	case row.TokensPerSecond >= 8:
		bias += 2
	case row.TokensPerSecond > 0 && row.TokensPerSecond < 4:
		bias -= 8
	}
	switch {
	case row.ColdStartCostMS <= 0:
	case row.ColdStartCostMS <= 600:
		bias += 2
	case row.ColdStartCostMS <= 1800:
		bias -= 1
	default:
		bias -= 8
	}
	return bias
}

func localBenchmarkBias(mode string) int {
	report := loadLocalRuntimeCapabilities()
	if !localRuntimeCapabilitiesAreFresh(report) {
		return 0
	}
	row, ok := report.Adapters[mode]
	if !ok {
		return 0
	}
	bias := localBenchmarkSampleBias(row)
	bias += localBenchmarkHistoryBias(row, report.History[mode])
	return bias
}

func localBenchmarkBiasForFeatures(mode string, features routeFeatures) int {
	report := loadLocalRuntimeCapabilities()
	if !localRuntimeCapabilitiesAreFresh(report) {
		return 0
	}
	row, ok := report.Adapters[mode]
	if !ok {
		return 0
	}
	baseBias := localBenchmarkSampleBias(row) + localBenchmarkHistoryBias(row, report.History[mode])
	weight := localBenchmarkScopeWeight(mode, features)
	if weight <= 0 {
		return 0
	}
	scopedTransfer := baseBias
	if weight < 0.999 {
		scopedTransfer = int(math.Round(float64(baseBias) * weight))
	}
	if cohortBias, ok := localCohortBenchmarkBias(mode, features, report); ok {
		residualTransfer := int(math.Round(float64(scopedTransfer) * 0.25))
		return cohortBias + residualTransfer + localCohortFeedbackBias(mode, features, report)
	}
	return scopedTransfer + localCohortFeedbackBias(mode, features, report)
}

func localBenchmarkSummaryLines(ctx context.Context) []string {
	report := currentLocalRuntimeCapabilities(ctx)
	return localBenchmarkSummaryLinesFromReport(report)
}

func freshLocalBenchmarkSummaryLines() []string {
	report := loadLocalRuntimeCapabilities()
	if !localRuntimeCapabilitiesAreFresh(report) {
		return nil
	}
	return localBenchmarkSummaryLinesFromReport(report)
}

func localBenchmarkSummaryLinesFromReport(report localRuntimeCapabilities) []string {
	lines := []string{}
	keys := []string{}
	for key := range report.Adapters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		row := report.Adapters[key]
		summary := fmt.Sprintf("BENCH_%s: %s", strings.ToUpper(strings.ReplaceAll(key, "-", "_")), row.Status)
		if row.LatencyMS > 0 {
			summary += fmt.Sprintf(" %dms", row.LatencyMS)
		}
		if strings.TrimSpace(row.QualityClass) != "" {
			summary += " " + row.QualityClass
		}
		if strings.TrimSpace(row.StructuredReliability) != "" && row.StructuredReliability != "unknown" {
			summary += " " + row.StructuredReliability
		}
		if row.TokensPerSecond > 0 {
			summary += fmt.Sprintf(" %.1ftok/s", row.TokensPerSecond)
		}
		if row.ColdStartCostMS > 0 {
			summary += fmt.Sprintf(" cold+%dms", row.ColdStartCostMS)
		}
		if trend := localBenchmarkTrend(row, report.History[key]); trend != "" {
			summary += " " + trend
		}
		if strings.TrimSpace(row.ModelID) != "" {
			summary += " -> " + row.ModelID
		}
		lines = append(lines, summary)
	}
	return lines
}

func localMultimodalLearningSummaryLinesFromReport(report localRuntimeCapabilities) []string {
	type subtypeSpec struct {
		key   string
		label string
	}
	subtypes := []subtypeSpec{
		{key: "multimodal-image-screenshot", label: "screenshot"},
		{key: "multimodal-pdf-scan", label: "pdf-scan"},
		{key: "multimodal-audio-transcript", label: "audio-transcript"},
	}
	lines := []string{}
	for _, subtype := range subtypes {
		bestMode, bestSamples, bestSignals := bestMultimodalLearningBucket(report, subtype.key)
		if bestMode == "" {
			lines = append(lines, fmt.Sprintf("MULTIMODAL_LEARNING_%s: none yet", strings.ToUpper(strings.ReplaceAll(subtype.label, "-", "_"))))
			continue
		}
		line := fmt.Sprintf(
			"MULTIMODAL_LEARNING_%s: %s samples=%d signals=%d",
			strings.ToUpper(strings.ReplaceAll(subtype.label, "-", "_")),
			bestMode,
			bestSamples,
			bestSignals,
		)
		lines = append(lines, line)
	}
	return lines
}

func localMultimodalLearningViewLinesFromReport(report localRuntimeCapabilities) []string {
	type subtypeSpec struct {
		key   string
		label string
	}
	subtypes := []subtypeSpec{
		{key: "multimodal-image-screenshot", label: "Screenshot work"},
		{key: "multimodal-pdf-scan", label: "Scanned PDF work"},
		{key: "multimodal-audio-transcript", label: "Audio/transcript work"},
	}
	lines := []string{}
	for _, subtype := range subtypes {
		bestMode, bestSamples, bestSignals := bestMultimodalLearningBucket(report, subtype.key)
		if bestMode == "" {
			lines = append(lines, fmt.Sprintf("%s: none yet", subtype.label))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s (samples %d, signals %d)", subtype.label, bestMode, bestSamples, bestSignals))
	}
	return lines
}

func bestMultimodalLearningBucket(report localRuntimeCapabilities, subtypeKey string) (string, int, int) {
	bestMode := ""
	bestSamples := 0
	bestSignals := 0
	for mode, cohorts := range report.CohortHistory {
		samples := len(cohorts[subtypeKey])
		signals := localCohortFeedbackSignalCount(report.CohortFeedback[mode][subtypeKey])
		if samples == 0 && signals == 0 {
			continue
		}
		if samples > bestSamples || (samples == bestSamples && signals > bestSignals) || (samples == bestSamples && signals == bestSignals && (bestMode == "" || mode < bestMode)) {
			bestMode = mode
			bestSamples = samples
			bestSignals = signals
		}
	}
	return bestMode, bestSamples, bestSignals
}

func freshLocalMultimodalLearningSummaryLines() []string {
	report := loadLocalRuntimeCapabilities()
	if report.ContextType != "JiniLocalRuntimeCapabilities" {
		return nil
	}
	return localMultimodalLearningSummaryLinesFromReport(report)
}

func freshLocalMultimodalLearningViewLines(features routeFeatures) []string {
	if strings.TrimSpace(features.ModalityClass) != "multimodal" {
		return nil
	}
	report := loadLocalRuntimeCapabilities()
	if report.ContextType != "JiniLocalRuntimeCapabilities" {
		return nil
	}
	return localMultimodalLearningViewLinesFromReport(report)
}

func localCohortFeedbackSignalCount(row localCohortFeedbackRow) int {
	return row.Upvotes +
		row.Downvotes +
		row.AcceptedAsIs +
		row.NeededLightEdits +
		row.NotUseful +
		row.OutcomeUsed +
		row.OutcomeShared +
		row.OutcomeReplaced +
		row.PassiveReopened +
		row.PassiveExportOpened +
		row.PassiveReplacedLater +
		row.PassiveAcceptedAsIs +
		row.PassiveNeededLightEdits +
		row.PassiveNeededHeavyEdits +
		row.PassiveHeaderOnlyEdits +
		row.PassiveCoreSectionEdits +
		row.PassiveCoreWordingEdits +
		row.PassiveDecisionChanges
}

func localBenchmarkReasonSuffix(mode string) string {
	report := loadLocalRuntimeCapabilities()
	if !localRuntimeCapabilitiesAreFresh(report) {
		return ""
	}
	row, ok := report.Adapters[mode]
	if !ok {
		return ""
	}
	parts := []string{}
	if row.Status != "" {
		parts = append(parts, row.Status)
	}
	if row.LatencyMS > 0 {
		parts = append(parts, fmt.Sprintf("%dms", row.LatencyMS))
	}
	if row.QualityClass != "" {
		parts = append(parts, row.QualityClass)
	}
	if row.StructuredReliability != "" && row.StructuredReliability != "unknown" {
		parts = append(parts, "reliability "+row.StructuredReliability)
	}
	if row.TokensPerSecond > 0 {
		parts = append(parts, fmt.Sprintf("%.1f tok/s", row.TokensPerSecond))
	}
	if row.ColdStartCostMS > 0 {
		parts = append(parts, fmt.Sprintf("cold+%dms", row.ColdStartCostMS))
	}
	if trend := localBenchmarkTrend(row, report.History[mode]); trend != "" {
		parts = append(parts, "trend "+trend)
	}
	if len(parts) == 0 {
		return ""
	}
	return " Last local benchmark: " + strings.Join(parts, ", ") + "."
}

func localBenchmarkReasonSuffixForRequest(mode string, request providerGenerationRequest) string {
	suffix := localBenchmarkReasonSuffix(mode)
	if suffix == "" {
		return ""
	}
	features := classifyRouteFeatures(request)
	scope := localBenchmarkScopeLabel(mode)
	current := localRequestScopeLabel(features)
	report := loadLocalRuntimeCapabilities()
	if localRuntimeCapabilitiesAreFresh(report) {
		if _, ok := localCohortBenchmarkBias(mode, features, report); ok {
			if current == "" {
				return suffix + " Jini also has direct local evidence for this request cohort."
			}
			return suffix + " Jini also has direct local evidence for " + current + " work."
		}
	}
	if scope == "" {
		return suffix
	}
	weight := localBenchmarkScopeWeight(mode, features)
	switch {
	case weight >= 0.95:
		if current == "" {
			return suffix + " This benchmark matches the current " + scope + " work closely."
		}
		return suffix + " This benchmark matches the current " + current + " work closely."
	case weight >= 0.75:
		if current == "" {
			return suffix + " Jini applies this benchmark with a partial match for the current " + scope + " work."
		}
		return suffix + " Jini applies this benchmark with a partial match from " + scope + " evidence to the current " + current + " work."
	default:
		if current == "" {
			return suffix + " Jini discounts this benchmark for the current request because it was measured on a different local work shape."
		}
		return suffix + " Jini discounts this benchmark for the current " + current + " request because it was measured on a different " + scope + " work shape."
	}
}

func mergedLocalRuntimeHistory(existing, report localRuntimeCapabilities) map[string][]localAdapterSample {
	history := map[string][]localAdapterSample{}
	sameContext := existing.ContextType == report.ContextType &&
		existing.JiniVersion == report.JiniVersion &&
		existing.CapabilityRegistryVersion == report.CapabilityRegistryVersion &&
		existing.DeviceProbeFingerprint == report.DeviceProbeFingerprint &&
		existing.LocalEndpointSignature == report.LocalEndpointSignature &&
		existing.LocalRuntimeClass == report.LocalRuntimeClass
	if sameContext {
		for mode, samples := range existing.History {
			history[mode] = append([]localAdapterSample{}, samples...)
		}
	}
	for mode, row := range report.Adapters {
		sample := localAdapterSample{
			ModelID:               row.ModelID,
			Status:                row.Status,
			LatencyMS:             row.LatencyMS,
			ColdStartCostMS:       row.ColdStartCostMS,
			TokensPerSecond:       row.TokensPerSecond,
			QualityClass:          row.QualityClass,
			StructuredReliability: row.StructuredReliability,
			BenchmarkedAt:         row.BenchmarkedAt,
		}
		series := history[mode]
		if len(series) > 0 && series[len(series)-1].BenchmarkedAt == sample.BenchmarkedAt {
			history[mode] = trimLocalRuntimeHistory(series)
			continue
		}
		history[mode] = trimLocalRuntimeHistory(append(series, sample))
	}
	return history
}

func mergedLocalRuntimeCohortHistory(existing, report localRuntimeCapabilities) map[string]map[string][]localAdapterSample {
	cohortHistory := map[string]map[string][]localAdapterSample{}
	sameContext := existing.ContextType == report.ContextType &&
		existing.JiniVersion == report.JiniVersion &&
		existing.CapabilityRegistryVersion == report.CapabilityRegistryVersion &&
		existing.DeviceProbeFingerprint == report.DeviceProbeFingerprint &&
		existing.LocalEndpointSignature == report.LocalEndpointSignature &&
		existing.LocalRuntimeClass == report.LocalRuntimeClass
	if sameContext {
		for mode, cohorts := range existing.CohortHistory {
			if cohortHistory[mode] == nil {
				cohortHistory[mode] = map[string][]localAdapterSample{}
			}
			for cohort, samples := range cohorts {
				cohortHistory[mode][cohort] = append([]localAdapterSample{}, samples...)
			}
		}
	}
	for mode, cohorts := range report.CohortHistory {
		if cohortHistory[mode] == nil {
			cohortHistory[mode] = map[string][]localAdapterSample{}
		}
		for cohort, samples := range cohorts {
			series := cohortHistory[mode][cohort]
			for _, sample := range samples {
				if len(series) > 0 && series[len(series)-1].BenchmarkedAt == sample.BenchmarkedAt {
					continue
				}
				series = trimLocalRuntimeHistory(append(series, sample))
			}
			cohortHistory[mode][cohort] = series
		}
	}
	return cohortHistory
}

func trimLocalRuntimeHistory(samples []localAdapterSample) []localAdapterSample {
	if len(samples) <= localRuntimeHistoryLimit {
		return samples
	}
	return append([]localAdapterSample{}, samples[len(samples)-localRuntimeHistoryLimit:]...)
}

func localBenchmarkHistoryBias(row localAdapterCapability, history []localAdapterSample) int {
	recent := recentComparableHistory(row, history)
	if len(recent) < 2 {
		return 0
	}
	previous := recent[:len(recent)-1]
	latest := recent[len(recent)-1]
	if latest.BenchmarkedAt != row.BenchmarkedAt {
		return 0
	}
	bonus := 0
	penalty := 0
	failures := 0
	var latencyTotal, tpsTotal float64
	stableCount := 0
	recentIssues := 0
	stabilityWeight := 0.0
	for _, sample := range previous {
		if sample.Status == "failed" || sample.Status == "degraded" {
			failures++
		}
		if sample.Status == "failed" || sample.Status == "degraded" || reliabilityRank(sample.StructuredReliability) < reliabilityRank("usable") || qualityRank(sample.QualityClass) < qualityRank("usable") {
			recentIssues++
		}
		if sample.LatencyMS > 0 {
			latencyTotal += float64(sample.LatencyMS)
		}
		if sample.TokensPerSecond > 0 {
			tpsTotal += sample.TokensPerSecond
		}
		if sample.Status == "ok" && sample.QualityClass == "strong" && sample.StructuredReliability == "strong" {
			stableCount++
			stabilityWeight += historySampleDecayWeight(sample.BenchmarkedAt, latest.BenchmarkedAt)
		}
	}
	if failures >= 2 {
		penalty -= 8
	}
	prevCount := float64(len(previous))
	if prevCount == 0 {
		return penalty
	}
	confidence := localBenchmarkHistoryConfidence(previous, latest, stableCount, recentIssues)
	confidence *= historyIssueConfidenceWeight(previous, latest)
	avgLatency := latencyTotal / prevCount
	avgTPS := tpsTotal / prevCount
	if stableCount == len(previous) && latest.Status == "ok" && latest.QualityClass == "strong" && latest.StructuredReliability == "strong" {
		bonus += scaleBonusByConfidence(stableHistoryBonus(len(previous)), stableHistoryConfidenceWeight(stabilityWeight, len(previous)))
	}
	if recoveryBonus, ok := localBenchmarkRecoveryBonus(previous, latest); ok {
		bonus += recoveryBonus
	}
	if avgLatency > 0 {
		switch {
		case float64(latest.LatencyMS) >= avgLatency*1.7:
			penalty -= 8
		case float64(latest.LatencyMS) >= avgLatency*1.35:
			penalty -= 4
		case latest.LatencyMS > 0 && float64(latest.LatencyMS) <= avgLatency*0.8:
			bonus += 2
		}
	}
	if avgTPS > 0 {
		switch {
		case latest.TokensPerSecond > 0 && latest.TokensPerSecond <= avgTPS*0.55:
			penalty -= 8
		case latest.TokensPerSecond > 0 && latest.TokensPerSecond <= avgTPS*0.75:
			penalty -= 4
		case latest.TokensPerSecond >= avgTPS*1.2:
			bonus += 2
		}
	}
	if reliabilityRank(latest.StructuredReliability) < reliabilityRank("usable") {
		penalty -= 6
	}
	return bonus + scalePenaltyByConfidence(penalty, confidence)
}

func localBenchmarkTrend(row localAdapterCapability, history []localAdapterSample) string {
	recent := recentComparableHistory(row, history)
	if len(recent) < 2 {
		return ""
	}
	previous := recent[:len(recent)-1]
	latest := recent[len(recent)-1]
	if latest.BenchmarkedAt != row.BenchmarkedAt {
		return ""
	}
	var latencyTotal, tpsTotal float64
	stableCount := 0
	recentIssues := 0
	for _, sample := range previous {
		latencyTotal += float64(sample.LatencyMS)
		tpsTotal += sample.TokensPerSecond
		if sample.Status == "ok" && sample.QualityClass == "strong" && sample.StructuredReliability == "strong" {
			stableCount++
		}
		if sample.Status == "failed" || sample.Status == "degraded" || reliabilityRank(sample.StructuredReliability) < reliabilityRank("usable") || qualityRank(sample.QualityClass) < qualityRank("usable") {
			recentIssues++
		}
	}
	avgLatency := latencyTotal / float64(len(previous))
	avgTPS := tpsTotal / float64(len(previous))
	confidence := localBenchmarkHistoryConfidence(previous, latest, stableCount, recentIssues) * historyIssueConfidenceWeight(previous, latest)
	switch {
	case avgLatency > 0 && float64(latest.LatencyMS) >= avgLatency*1.7:
		if confidence < 0.45 {
			return "slower-watch"
		}
		return "slower"
	case avgTPS > 0 && latest.TokensPerSecond > 0 && latest.TokensPerSecond <= avgTPS*0.55:
		if confidence < 0.45 {
			return "throughput-watch"
		}
		return "throughput-down"
	case localBenchmarkRecovered(previous, latest):
		return "recovered"
	case latest.Status == "ok" && latest.QualityClass == "strong" && latest.StructuredReliability == "strong":
		return "stable"
	default:
		return ""
	}
}

func localCohortBenchmarkBias(mode string, features routeFeatures, report localRuntimeCapabilities) (int, bool) {
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForFeatures(features))
	if cohort == "" || report.CohortHistory == nil {
		return 0, false
	}
	modeHistory := report.CohortHistory[mode]
	if len(modeHistory) == 0 {
		return 0, false
	}
	history := modeHistory[cohort]
	if len(history) == 0 {
		return 0, false
	}
	latest := history[len(history)-1]
	row := localAdapterCapabilityFromSample(mode, latest)
	bias := localBenchmarkSampleBias(row)
	bias += localBenchmarkHistoryBias(row, history)
	return bias, true
}

func localCohortFeedbackBias(mode string, features routeFeatures, report localRuntimeCapabilities) int {
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForFeatures(features))
	if cohort == "" || report.CohortFeedback == nil {
		return 0
	}
	modeFeedback := report.CohortFeedback[mode]
	if len(modeFeedback) == 0 {
		return 0
	}
	row, ok := modeFeedback[cohort]
	if !ok {
		return 0
	}
	bias := row.AcceptedAsIs*6 + row.NeededLightEdits*2 + row.Upvotes*5
	bias -= row.NotUseful*6 + row.Downvotes*5
	bias += row.OutcomeUsed*7 + row.OutcomeShared*9
	bias -= row.OutcomeReplaced * 10
	bias += row.PassiveReopened * 2
	bias += row.PassiveExportOpened * 5
	bias -= row.PassiveReplacedLater * 7
	bias += row.PassiveAcceptedAsIs*3 + row.PassiveNeededLightEdits
	bias -= row.PassiveNeededHeavyEdits * 4
	bias += row.PassiveHeaderOnlyEdits * 2
	bias -= row.PassiveCoreSectionEdits * 3
	bias -= row.PassiveCoreWordingEdits
	bias -= row.PassiveDecisionChanges * 5
	if bias > 24 {
		return 24
	}
	if bias < -24 {
		return -24
	}
	return bias
}

func localBenchmarkScopeWeight(mode string, features routeFeatures) float64 {
	switch mode {
	case "local-fast":
		switch {
		case features.ModalityClass == "multimodal":
			return 0.2
		case features.ArtifactFamily == "narrative-draft" || features.ArtifactFamily == "general-pass":
			return 1.0
		case features.ArtifactFamily == "structured-check" || features.ArtifactFamily == "comparison-matrix":
			return 0.85
		case features.ArtifactFamily == "step-plan":
			return 0.75
		case features.ArtifactFamily == "itinerary-plan":
			return 0.65
		case features.ArtifactFamily == "code-change":
			return 0.55
		case features.DepthClass == "deep":
			return 0.55
		case features.WorkClass == "planning" || features.WorkClass == "general":
			return 1.0
		case features.WorkClass == "code":
			return 0.7
		default:
			return 0.85
		}
	case "local-workhorse":
		switch {
		case features.ModalityClass == "multimodal":
			return 0.25
		case features.RequestCohort == "build-readiness":
			return 1.0
		case features.RequestCohort == "option-compare":
			return 0.95
		case features.RequestCohort == "sendable-followup":
			return 0.8
		case features.RequestCohort == "incident-cleanup":
			return 0.78
		case features.RequestCohort == "trip-itinerary":
			return 0.6
		case features.WorkClass == "general":
			return 0.95
		case features.WorkClass == "code" && features.DepthClass == "deep":
			return 0.55
		case features.WorkClass == "code":
			return 0.6
		case features.DepthClass == "deep":
			return 0.8
		default:
			return 0.85
		}
	case "local-deep":
		switch {
		case features.ModalityClass == "multimodal":
			return 0.25
		case features.RequestCohort == "incident-cleanup" || features.RequestCohort == "build-readiness":
			return 0.95
		case features.DepthClass == "deep" && features.WorkClass == "code":
			return 1.0
		case features.DepthClass == "deep":
			return 0.95
		case features.EffortClass == "high" || features.EffortClass == "extra high":
			return 0.8
		default:
			return 0.6
		}
	case "local-multimodal":
		if features.ModalityClass == "multimodal" {
			return 1.0
		}
		return 0.2
	default:
		return 1.0
	}
}

func localBenchmarkScopeLabel(mode string) string {
	switch mode {
	case "local-fast":
		return "quick text"
	case "local-workhorse":
		return "structured checklist and drafting"
	case "local-deep":
		return "deep analysis"
	case "local-multimodal":
		return "multimodal"
	default:
		return ""
	}
}

func localRequestScopeLabel(features routeFeatures) string {
	switch features.ModalitySubtype {
	case "pdf-scan":
		return "pdf or scanned document extraction"
	case "image-screenshot":
		return "image or screenshot extraction"
	case "audio-transcript":
		return "audio or transcript extraction"
	}
	switch features.RequestCohort {
	case "sendable-followup":
		return "sendable follow-up"
	case "build-readiness":
		return "build-readiness"
	case "trip-itinerary":
		return "trip itinerary"
	case "option-compare":
		return "option comparison"
	case "incident-cleanup":
		return "incident cleanup"
	case "multimodal-extract":
		return "multimodal extraction"
	case "code-change":
		return "code change"
	default:
		switch features.ArtifactFamily {
		case "narrative-draft":
			return "narrative draft"
		case "structured-check":
			return "structured check"
		case "itinerary-plan":
			return "itinerary plan"
		case "comparison-matrix":
			return "comparison matrix"
		case "step-plan":
			return "step plan"
		case "multimodal-extract":
			return "multimodal extraction"
		case "code-change":
			return "code change"
		default:
			return ""
		}
	}
}

func multimodalLearningSeparationNote(features routeFeatures) string {
	switch features.ModalitySubtype {
	case "pdf-scan":
		return " Jini learns scanned PDF and document work separately from screenshot and audio/transcript work."
	case "image-screenshot":
		return " Jini learns screenshot work separately from scanned PDF and audio/transcript work."
	case "audio-transcript":
		return " Jini learns audio/transcript work separately from screenshot and scanned PDF work."
	default:
		return ""
	}
}

func localBenchmarkCohortKeyForRequest(request providerGenerationRequest) string {
	features := classifyRouteFeatures(request)
	return localBenchmarkCohortKeyForFeatures(features)
}

func localBenchmarkCohortKeyForFeatures(features routeFeatures) string {
	cohort := strings.TrimSpace(features.RequestCohort)
	if cohort != "multimodal-extract" {
		return cohort
	}
	switch strings.TrimSpace(features.ModalitySubtype) {
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

func localAdapterCapabilityFromSample(mode string, sample localAdapterSample) localAdapterCapability {
	return localAdapterCapability{
		AdapterID:             mode,
		ModelID:               sample.ModelID,
		Status:                sample.Status,
		LatencyMS:             sample.LatencyMS,
		ColdStartCostMS:       sample.ColdStartCostMS,
		OutputTokens:          0,
		TokensPerSecond:       sample.TokensPerSecond,
		QualityClass:          sample.QualityClass,
		StructuredReliability: sample.StructuredReliability,
		BenchmarkedAt:         sample.BenchmarkedAt,
	}
}

func recordLocalCohortSample(mode string, request providerGenerationRequest, latencyMS int, content string) error {
	if !strings.HasPrefix(mode, "local-") {
		return nil
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForRequest(request))
	if cohort == "" {
		return nil
	}
	report := loadLocalRuntimeCapabilities()
	if report.Adapters == nil {
		report.Adapters = map[string]localAdapterCapability{}
	}
	if report.History == nil {
		report.History = map[string][]localAdapterSample{}
	}
	if report.CohortHistory == nil {
		report.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if report.CohortFeedback == nil {
		report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	if report.CohortHistory[mode] == nil {
		report.CohortHistory[mode] = map[string][]localAdapterSample{}
	}
	modelID, _ := resolveLocalSLMModelForToolMode(mode)
	quality := classifyGeneratedCohortQuality(request, content)
	reliability := classifyGeneratedCohortReliability(quality)
	status := cohortSampleStatus(quality, reliability)
	tokens := estimateTokenCount(content)
	now := time.Now().UTC().Format(time.RFC3339)
	sample := localAdapterSample{
		ModelID:               strings.TrimSpace(modelID),
		Status:                status,
		LatencyMS:             latencyMS,
		TokensPerSecond:       tokensPerSecond(tokens, latencyMS),
		QualityClass:          quality,
		StructuredReliability: reliability,
		BenchmarkedAt:         now,
	}
	report.CohortHistory[mode][cohort] = trimLocalRuntimeHistory(append(report.CohortHistory[mode][cohort], sample))
	return saveLocalRuntimeCapabilities(report)
}

func recordLocalCohortFeedback(mode string, request providerGenerationRequest, feedback, editClass, editScope, semanticClass, reason string) error {
	if !strings.HasPrefix(mode, "local-") {
		return nil
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForRequest(request))
	if cohort == "" {
		return nil
	}
	report := loadLocalRuntimeCapabilities()
	if report.Adapters == nil {
		report.Adapters = map[string]localAdapterCapability{}
	}
	if report.History == nil {
		report.History = map[string][]localAdapterSample{}
	}
	if report.CohortHistory == nil {
		report.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if report.CohortFeedback == nil {
		report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	if report.CohortFeedback[mode] == nil {
		report.CohortFeedback[mode] = map[string]localCohortFeedbackRow{}
	}
	row := report.CohortFeedback[mode][cohort]
	switch strings.TrimSpace(feedback) {
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
	switch strings.TrimSpace(editClass) {
	case "none":
		row.PassiveAcceptedAsIs++
	case "light":
		row.PassiveNeededLightEdits++
	case "heavy":
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
	_ = reason
	report.CohortFeedback[mode][cohort] = row
	return saveLocalRuntimeCapabilities(report)
}

func recordLocalCohortOutcome(mode string, request providerGenerationRequest, outcome, reason string) error {
	if !strings.HasPrefix(mode, "local-") {
		return nil
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForRequest(request))
	if cohort == "" {
		return nil
	}
	report := loadLocalRuntimeCapabilities()
	if report.Adapters == nil {
		report.Adapters = map[string]localAdapterCapability{}
	}
	if report.History == nil {
		report.History = map[string][]localAdapterSample{}
	}
	if report.CohortHistory == nil {
		report.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if report.CohortFeedback == nil {
		report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	if report.CohortFeedback[mode] == nil {
		report.CohortFeedback[mode] = map[string]localCohortFeedbackRow{}
	}
	row := report.CohortFeedback[mode][cohort]
	switch strings.TrimSpace(outcome) {
	case "used-this":
		row.OutcomeUsed++
	case "shared-this":
		row.OutcomeShared++
	case "replaced-this":
		row.OutcomeReplaced++
	default:
		return nil
	}
	_ = reason
	report.CohortFeedback[mode][cohort] = row
	return saveLocalRuntimeCapabilities(report)
}

func recordPassiveLocalCohortOutcome(mode string, request providerGenerationRequest, outcome, reason string) error {
	if !strings.HasPrefix(mode, "local-") {
		return nil
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForRequest(request))
	if cohort == "" {
		return nil
	}
	report := loadLocalRuntimeCapabilities()
	if report.Adapters == nil {
		report.Adapters = map[string]localAdapterCapability{}
	}
	if report.History == nil {
		report.History = map[string][]localAdapterSample{}
	}
	if report.CohortHistory == nil {
		report.CohortHistory = map[string]map[string][]localAdapterSample{}
	}
	if report.CohortFeedback == nil {
		report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	}
	if report.CohortFeedback[mode] == nil {
		report.CohortFeedback[mode] = map[string]localCohortFeedbackRow{}
	}
	row := report.CohortFeedback[mode][cohort]
	switch strings.TrimSpace(outcome) {
	case "used-this":
		row.PassiveReopened++
	case "shared-this":
		row.PassiveExportOpened++
	case "replaced-this":
		row.PassiveReplacedLater++
	default:
		return nil
	}
	_ = reason
	report.CohortFeedback[mode][cohort] = row
	return saveLocalRuntimeCapabilities(report)
}

func classifyGeneratedCohortQuality(request providerGenerationRequest, content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "failed"
	}
	switch localBenchmarkCohortKeyForRequest(request) {
	case "sendable-followup":
		required := []string{
			"## Send this note",
			"## Decisions captured from the notes",
			"## Owners and due dates to confirm",
			"## Open questions to close",
			"## Recommended next move",
		}
		if containsAllCaseInsensitive(content, required) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"send this note", "recommended next move", "open questions"}) {
			return "usable"
		}
	case "build-readiness":
		required := []string{
			"## What looks ready now",
			"## Must clear before build",
			"## Recommended first slice",
			"## Who needs to answer what",
			"## Still to confirm",
		}
		if containsAllCaseInsensitive(content, required) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"must clear before build", "still to confirm", "recommended first slice"}) {
			return "usable"
		}
	case "trip-itinerary":
		days := countTripDayMentions(content)
		if days >= 4 && containsAnyCaseInsensitive(content, []string{"still to confirm", "budget", "day 1"}) {
			return "strong"
		}
		if days >= 2 || containsAnyCaseInsensitive(content, []string{"itinerary", "budget", "travel logistics"}) {
			return "usable"
		}
	case "option-compare":
		if countCaseInsensitive(content, "option") >= 2 && containsAnyCaseInsensitive(content, []string{"recommended", "still to confirm"}) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"option", "recommended"}) {
			return "usable"
		}
	case "incident-cleanup":
		if containsAllCaseInsensitive(content, []string{"immediate", "next", "verify"}) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"immediate", "next step", "verify"}) {
			return "usable"
		}
	case "multimodal-pdf-scan":
		if containsAllCaseInsensitive(content, []string{"extracted evidence", "what the document shows", "still unclear"}) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"pdf", "document", "ocr", "scan", "what the document shows"}) {
			return "usable"
		}
	case "multimodal-image-screenshot":
		if containsAllCaseInsensitive(content, []string{"extracted evidence", "what is visible", "still unclear"}) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"image", "screenshot", "visible", "screen", "what is visible"}) {
			return "usable"
		}
	case "multimodal-audio-transcript":
		if containsAllCaseInsensitive(content, []string{"extracted evidence", "what the recording says", "still unclear"}) {
			return "strong"
		}
		if containsAnyCaseInsensitive(content, []string{"audio", "recording", "transcript", "speaker", "what the recording says"}) {
			return "usable"
		}
	case "code-change":
		if containsAnyCaseInsensitive(content, []string{"```", "diff", "tests", "fix"}) {
			return "usable"
		}
	}
	if len(content) >= 120 {
		return "weak"
	}
	return "failed"
}

func classifyGeneratedCohortReliability(quality string) string {
	switch quality {
	case "strong":
		return "strong"
	case "usable":
		return "usable"
	case "weak":
		return "weak"
	default:
		return "failed"
	}
}

func cohortSampleStatus(quality, reliability string) string {
	switch reliability {
	case "strong", "usable":
		switch quality {
		case "strong", "usable":
			return "ok"
		case "weak":
			return "degraded"
		default:
			return "failed"
		}
	case "weak":
		return "degraded"
	default:
		return "failed"
	}
}

func containsAllCaseInsensitive(content string, required []string) bool {
	lower := strings.ToLower(content)
	for _, item := range required {
		if !strings.Contains(lower, strings.ToLower(item)) {
			return false
		}
	}
	return true
}

func containsAnyCaseInsensitive(content string, required []string) bool {
	lower := strings.ToLower(content)
	for _, item := range required {
		if strings.Contains(lower, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func countCaseInsensitive(content, needle string) int {
	if strings.TrimSpace(needle) == "" {
		return 0
	}
	return strings.Count(strings.ToLower(content), strings.ToLower(needle))
}

func countTripDayMentions(content string) int {
	count := 0
	lower := strings.ToLower(content)
	for day := 1; day <= 7; day++ {
		if strings.Contains(lower, fmt.Sprintf("day %d", day)) {
			count++
		}
	}
	return count
}

func recentComparableHistory(row localAdapterCapability, history []localAdapterSample) []localAdapterSample {
	filtered := []localAdapterSample{}
	for _, sample := range history {
		if strings.TrimSpace(sample.ModelID) != strings.TrimSpace(row.ModelID) {
			continue
		}
		filtered = append(filtered, sample)
	}
	if len(filtered) > 4 {
		return filtered[len(filtered)-4:]
	}
	return filtered
}

func reliabilityRank(value string) int {
	switch value {
	case "strong":
		return 4
	case "usable":
		return 3
	case "weak":
		return 2
	case "failed":
		return 1
	default:
		return 0
	}
}

func localBenchmarkHistoryConfidence(previous []localAdapterSample, latest localAdapterSample, stableCount, recentIssues int) float64 {
	confidence := 0.25
	switch len(previous) {
	case 1:
		confidence = 0.25
	case 2:
		confidence = 0.35
	case 3:
		confidence = 0.5
	default:
		confidence = 0.65
	}
	if recentIssues >= 2 {
		confidence += 0.25
	} else if recentIssues == 1 {
		confidence += 0.12
	}
	if stableCount == len(previous) && latest.Status == "ok" && (latest.QualityClass != "strong" || latest.StructuredReliability != "strong") {
		confidence -= 0.1
	}
	if confidence < 0.2 {
		confidence = 0.2
	}
	if confidence > 0.95 {
		confidence = 0.95
	}
	return confidence
}

func stableHistoryBonus(count int) int {
	switch {
	case count >= 4:
		return 4
	case count == 3:
		return 3
	case count == 2:
		return 2
	default:
		return 1
	}
}

func scalePenaltyByConfidence(penalty int, confidence float64) int {
	if penalty >= 0 {
		return penalty
	}
	return int(math.Round(float64(penalty) * confidence))
}

func scaleBonusByConfidence(bonus int, confidence float64) int {
	if bonus <= 0 {
		return bonus
	}
	return int(math.Round(float64(bonus) * confidence))
}

func historyIssueConfidenceWeight(previous []localAdapterSample, latest localAdapterSample) float64 {
	weights := []float64{}
	for _, sample := range previous {
		if sample.Status == "failed" || sample.Status == "degraded" || reliabilityRank(sample.StructuredReliability) < reliabilityRank("usable") || qualityRank(sample.QualityClass) < qualityRank("usable") {
			weights = append(weights, historySampleDecayWeight(sample.BenchmarkedAt, latest.BenchmarkedAt))
		}
	}
	if len(weights) == 0 {
		return 1.0
	}
	total := 0.0
	for _, weight := range weights {
		total += weight
	}
	avg := total / float64(len(weights))
	if avg < 0.35 {
		return 0.35
	}
	if avg > 1.0 {
		return 1.0
	}
	return avg
}

func stableHistoryConfidenceWeight(stabilityWeight float64, stableCount int) float64 {
	if stableCount <= 0 {
		return 0
	}
	avg := stabilityWeight / float64(stableCount)
	if avg < 0.3 {
		return 0.3
	}
	if avg > 1.0 {
		return 1.0
	}
	return avg
}

func historySampleDecayWeight(sampleAt, latestAt string) float64 {
	sampleTime, err := time.Parse(time.RFC3339, strings.TrimSpace(sampleAt))
	if err != nil {
		return 1.0
	}
	latestTime, err := time.Parse(time.RFC3339, strings.TrimSpace(latestAt))
	if err != nil {
		return 1.0
	}
	age := latestTime.Sub(sampleTime)
	switch {
	case age <= 2*time.Hour:
		return 1.0
	case age <= 6*time.Hour:
		return 0.85
	case age <= 12*time.Hour:
		return 0.65
	case age <= 18*time.Hour:
		return 0.5
	case age <= 24*time.Hour:
		return 0.35
	default:
		return 0.2
	}
}

func localBenchmarkRecovered(previous []localAdapterSample, latest localAdapterSample) bool {
	if latest.Status != "ok" || latest.QualityClass != "strong" || latest.StructuredReliability != "strong" {
		return false
	}
	recentIssueWeight := 0.0
	for _, sample := range previous {
		if sample.Status == "failed" || sample.Status == "degraded" || reliabilityRank(sample.StructuredReliability) < reliabilityRank("usable") || qualityRank(sample.QualityClass) < qualityRank("usable") {
			recentIssueWeight += historySampleDecayWeight(sample.BenchmarkedAt, latest.BenchmarkedAt)
		}
	}
	return recentIssueWeight >= 0.6
}

func localBenchmarkRecoveryBonus(previous []localAdapterSample, latest localAdapterSample) (int, bool) {
	if !localBenchmarkRecovered(previous, latest) {
		return 0, false
	}
	recoveryWeight := 0.0
	for _, sample := range previous {
		if sample.Status == "failed" || sample.Status == "degraded" || reliabilityRank(sample.StructuredReliability) < reliabilityRank("usable") || qualityRank(sample.QualityClass) < qualityRank("usable") {
			recoveryWeight += historySampleDecayWeight(sample.BenchmarkedAt, latest.BenchmarkedAt)
		}
	}
	switch {
	case recoveryWeight >= 1.6:
		return 8, true
	case recoveryWeight >= 1.0:
		return 6, true
	default:
		return 4, true
	}
}
