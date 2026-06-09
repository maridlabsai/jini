package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const localSLMAutoDiscoveryTimeout = 220 * time.Millisecond

var (
	localSLMDiscoveryHTTPClient = &http.Client{Timeout: localSLMAutoDiscoveryTimeout}
	powerProfileProbeEnabled    = true
)

type localSLMRuntimeDiscovery struct {
	Endpoint       string
	EndpointSource string
	Models         []string
	RuntimeClass   string
}

type powerProfile struct {
	PowerSource    string
	BatteryPercent int
	LowBattery     bool
}

var (
	localSLMAutoDiscoveryMu     sync.Mutex
	localSLMAutoDiscoveryCached bool
	localSLMAutoDiscoveryKey    string
	localSLMAutoDiscoveryValue  localSLMRuntimeDiscovery
)

func resolvedLocalSLMEndpoint() (string, string) {
	discovery := localSLMRuntimeDiscoveryResult()
	return strings.TrimSpace(discovery.Endpoint), strings.TrimSpace(discovery.EndpointSource)
}

func localSLMRuntimeDiscoveryResult() localSLMRuntimeDiscovery {
	key := strings.Join([]string{
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_MODEL")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_FAST_MODEL")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_WORKHORSE_MODEL")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_DEEP_MODEL")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_MULTIMODAL_MODEL")),
		strings.TrimSpace(configValue("JINI_LOCAL_SLM_AUTO_DISCOVERY")),
		strings.TrimSpace(configuredModelInput()),
		configuredProviderMode(),
		configuredToolMode(),
		strconv.FormatBool(configuredModelAllowsLocalAutoDiscovery()),
	}, "|")

	localSLMAutoDiscoveryMu.Lock()
	if localSLMAutoDiscoveryCached && localSLMAutoDiscoveryKey == key {
		cached := cloneLocalSLMRuntimeDiscovery(localSLMAutoDiscoveryValue)
		localSLMAutoDiscoveryMu.Unlock()
		return cached
	}
	localSLMAutoDiscoveryMu.Unlock()

	discovery := discoverLocalSLMRuntime()

	localSLMAutoDiscoveryMu.Lock()
	localSLMAutoDiscoveryCached = true
	localSLMAutoDiscoveryKey = key
	localSLMAutoDiscoveryValue = cloneLocalSLMRuntimeDiscovery(discovery)
	localSLMAutoDiscoveryMu.Unlock()
	return discovery
}

func cloneLocalSLMRuntimeDiscovery(source localSLMRuntimeDiscovery) localSLMRuntimeDiscovery {
	out := source
	out.Models = append([]string{}, source.Models...)
	return out
}

func discoverLocalSLMRuntime() localSLMRuntimeDiscovery {
	if endpoint := strings.TrimRight(strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")), "/"); endpoint != "" {
		models := []string{}
		if strings.TrimSpace(resolveConfiguredLocalSLMAnyModel()) == "" {
			models = discoverLocalSLMModelsAtEndpoint(endpoint)
		}
		return localSLMRuntimeDiscovery{
			Endpoint:       endpoint,
			EndpointSource: "configured",
			Models:         models,
			RuntimeClass:   runtimeClassForLocalSLMEndpoint(endpoint),
		}
	}
	if !localSLMAutoDiscoveryAllowed() {
		return localSLMRuntimeDiscovery{RuntimeClass: "not-configured"}
	}
	for _, endpoint := range defaultLocalSLMDiscoveryEndpoints() {
		models := discoverLocalSLMModelsAtEndpoint(endpoint)
		if len(models) == 0 {
			continue
		}
		return localSLMRuntimeDiscovery{
			Endpoint:       endpoint,
			EndpointSource: "auto",
			Models:         models,
			RuntimeClass:   runtimeClassForLocalSLMEndpoint(endpoint),
		}
	}
	return localSLMRuntimeDiscovery{RuntimeClass: "not-configured"}
}

func localSLMAutoDiscoveryEnabled() bool {
	switch normalizeName(configValue("JINI_LOCAL_SLM_AUTO_DISCOVERY")) {
	case "0", "false", "off", "disabled", "no":
		return false
	default:
		return true
	}
}

func localSLMAutoDiscoveryAllowed() bool {
	if !localSLMAutoDiscoveryEnabled() {
		return false
	}
	providerMode := configuredProviderMode()
	toolMode := configuredToolMode()
	if providerMode == "local-slm" || strings.HasPrefix(toolMode, "local-") {
		return true
	}
	if providerMode == "auto" && configuredModelAllowsLocalAutoDiscovery() {
		return true
	}
	return false
}

func configuredModelAllowsLocalAutoDiscovery() bool {
	model := normalizeName(configuredModelInput())
	if model == "" || model == "auto" {
		return true
	}
	return containsAny(model, []string{"local", "offline", "slm"})
}

func defaultLocalSLMDiscoveryEndpoints() []string {
	return []string{
		"http://127.0.0.1:11434/v1",
		"http://127.0.0.1:1234/v1",
		"http://127.0.0.1:8080/v1",
		"http://127.0.0.1:8000/v1",
	}
}

func discoverLocalSLMModelsAtEndpoint(endpoint string) []string {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), localSLMAutoDiscoveryTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/models", nil)
	if err != nil {
		return nil
	}
	if apiKey := strings.TrimSpace(configValue("JINI_LOCAL_SLM_API_KEY")); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := localSLMDiscoveryHTTPClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil
	}
	var decoded struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil
	}
	models := make([]string, 0, len(decoded.Data))
	seen := map[string]bool{}
	for _, item := range decoded.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" || seen[id] || isNonChatLocalModel(id) {
			continue
		}
		seen[id] = true
		models = append(models, id)
	}
	sort.SliceStable(models, func(i, j int) bool {
		return strings.ToLower(models[i]) < strings.ToLower(models[j])
	})
	return models
}

func isNonChatLocalModel(modelID string) bool {
	normalized := normalizeName(modelID)
	return containsAny(normalized, []string{
		"embed", "embedding", "nomic", "bge", "e5", "rerank", "whisper",
	})
}

func runtimeClassForLocalSLMEndpoint(endpoint string) string {
	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	switch {
	case endpoint == "":
		return "not-configured"
	case strings.Contains(endpoint, "127.0.0.1") || strings.Contains(endpoint, "localhost"):
		if strings.Contains(endpoint, "11434") {
			return "ollama-openai-compatible"
		}
		return "local-openai-compatible"
	default:
		return "remote-openai-compatible"
	}
}

func autoLocalSLMModelForToolMode(toolMode string) (string, string) {
	discovery := localSLMRuntimeDiscoveryResult()
	if len(discovery.Models) == 0 {
		return "", ""
	}
	profile := autoLocalSLMSelectionDeviceProfile()
	power := currentPowerProfile()
	modelID := chooseLocalSLMModelForToolMode(toolMode, discovery.Models, profile, power)
	return modelID, modelID
}

func chooseLocalSLMModelForToolMode(toolMode string, models []string, profile deviceProfile, power powerProfile) string {
	bestModel := ""
	bestScore := -99999
	for index, modelID := range models {
		score := scoreLocalSLMModelForToolMode(toolMode, modelID, profile, power)
		score -= index
		if score > bestScore {
			bestScore = score
			bestModel = strings.TrimSpace(modelID)
		}
	}
	return bestModel
}

func scoreLocalSLMModelForToolMode(toolMode, modelID string, profile deviceProfile, power powerProfile) int {
	normalized := normalizeName(modelID)
	sizeB := localModelSizeBillions(modelID)
	score := 0
	if isNonChatLocalModel(modelID) {
		return -10000
	}
	if containsAny(normalized, []string{"qwen", "llama", "gemma", "mistral", "deepseek", "phi"}) {
		score += 8
	}
	if containsAny(normalized, []string{"instruct", "chat", "coder"}) {
		score += 6
	}
	if containsAny(normalized, []string{"preview", "experimental"}) {
		score -= 4
	}

	switch toolMode {
	case "local-fast":
		score += scoreFastLocalModel(sizeB, normalized)
	case "local-deep":
		score += scoreDeepLocalModel(sizeB, normalized, profile)
	case "local-multimodal":
		score += scoreMultimodalLocalModel(sizeB, normalized)
	default:
		score += scoreWorkhorseLocalModel(sizeB, normalized)
	}

	switch profile.DeviceClass {
	case "tiny":
		if sizeB > 4 {
			score -= 28
		}
	case "laptop-light":
		if sizeB > 8 {
			score -= 18
		}
	case "laptop-strong":
		if sizeB >= 7 && sizeB <= 14 {
			score += 8
		}
	case "workstation", "gpu-heavy":
		if sizeB >= 14 {
			score += 8
		}
	}
	if power.LowBattery {
		switch {
		case sizeB > 14:
			score -= 36
		case sizeB > 8:
			score -= 18
		case sizeB > 0 && sizeB <= 4:
			score += 18
		}
		if toolMode == "local-fast" {
			score += 20
		}
	}
	return score
}

func scoreFastLocalModel(sizeB float64, normalized string) int {
	score := 0
	if containsAny(normalized, []string{"mini", "small", "lite", "phi"}) {
		score += 34
	}
	switch {
	case sizeB > 0 && sizeB <= 4:
		score += 32
	case sizeB > 4 && sizeB <= 8:
		score += 12
	case sizeB > 8:
		score -= 20
	}
	return score
}

func scoreWorkhorseLocalModel(sizeB float64, normalized string) int {
	score := 0
	if containsAny(normalized, []string{"qwen", "llama", "mistral", "gemma"}) {
		score += 18
	}
	switch {
	case sizeB >= 7 && sizeB <= 12:
		score += 36
	case sizeB > 4 && sizeB < 7:
		score += 18
	case sizeB > 12 && sizeB <= 14:
		score += 8
	case sizeB > 14:
		score -= 14
	case sizeB > 0 && sizeB <= 4:
		score -= 8
	}
	return score
}

func scoreDeepLocalModel(sizeB float64, normalized string, profile deviceProfile) int {
	score := 0
	if containsAny(normalized, []string{"deepseek", "qwen", "llama", "reason"}) {
		score += 16
	}
	switch {
	case sizeB >= 14 && sizeB <= 34:
		score += 38
	case sizeB > 8 && sizeB < 14:
		score += 18
	case sizeB >= 7 && sizeB <= 8:
		score += 8
	case sizeB > 34:
		score -= 18
	case sizeB > 0 && sizeB <= 4:
		score -= 22
	}
	if profile.DeviceClass == "tiny" || profile.DeviceClass == "laptop-light" {
		if sizeB >= 14 {
			score -= 30
		}
	}
	return score
}

func scoreMultimodalLocalModel(sizeB float64, normalized string) int {
	score := 0
	if containsAny(normalized, []string{"vision", "vl", "llava", "moondream", "bakllava", "multimodal"}) {
		score += 44
	}
	if strings.Contains(normalized, "gemma3") || strings.Contains(normalized, "gemma 3") {
		score += 36
	}
	switch {
	case sizeB >= 4 && sizeB <= 14:
		score += 18
	case sizeB > 14:
		score -= 6
	case sizeB > 0 && sizeB < 4:
		score += 8
	}
	return score
}

var localModelSizePattern = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*b`)

func localModelSizeBillions(modelID string) float64 {
	match := localModelSizePattern.FindStringSubmatch(modelID)
	if len(match) < 2 {
		return 0
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0
	}
	return value
}

func autoLocalSLMSelectionDeviceProfile() deviceProfile {
	if override := strings.TrimSpace(configValue("JINI_DEVICE_CLASS_OVERRIDE")); override != "" {
		return deviceProfile{DeviceClass: override}
	}
	goos := runtime.GOOS
	arch := runtime.GOARCH
	memoryBytes := probeTotalMemoryBytes(goos)
	profile := deviceProfile{
		OS:               goos,
		Arch:             arch,
		CPUCount:         runtime.NumCPU(),
		TotalMemoryBytes: memoryBytes,
		TotalMemoryGB:    bytesToRoundedGB(memoryBytes),
		AcceleratorClass: detectAcceleratorClass(goos, arch),
	}
	profile.DeviceClass = classifyDeviceClass(profile)
	return profile
}

func currentPowerProfile() powerProfile {
	if override := strings.TrimSpace(configValue("JINI_POWER_SOURCE_OVERRIDE")); override != "" {
		percent := parseBatteryPercent(configValue("JINI_BATTERY_PERCENT_OVERRIDE"))
		return powerProfile{
			PowerSource:    normalizeName(override),
			BatteryPercent: percent,
			LowBattery:     normalizeName(override) == "battery" && percent > 0 && percent <= 25,
		}
	}
	if !powerProfileProbeEnabled || strings.TrimSpace(configValue("JINI_POWER_PROBE_DISABLE")) != "" {
		return powerProfile{PowerSource: "unknown"}
	}
	switch runtime.GOOS {
	case "darwin":
		return probeDarwinPowerProfile()
	case "linux":
		return probeLinuxPowerProfile()
	default:
		return powerProfile{PowerSource: "unknown"}
	}
}

func probeDarwinPowerProfile() powerProfile {
	ctx, cancel := context.WithTimeout(context.Background(), localSLMAutoDiscoveryTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pmset", "-g", "batt")
	cmd.WaitDelay = 50 * time.Millisecond
	outputBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outputBytes))
	if err != nil || strings.TrimSpace(output) == "" {
		return powerProfile{PowerSource: "unknown"}
	}
	normalized := strings.ToLower(output)
	percent := parseBatteryPercent(output)
	source := "unknown"
	if strings.Contains(normalized, "battery power") {
		source = "battery"
	} else if strings.Contains(normalized, "ac power") {
		source = "ac"
	}
	return powerProfile{
		PowerSource:    source,
		BatteryPercent: percent,
		LowBattery:     source == "battery" && percent > 0 && percent <= 25,
	}
}

func probeLinuxPowerProfile() powerProfile {
	capacityPaths, err := filepath.Glob("/sys/class/power_supply/BAT*/capacity")
	if err != nil || len(capacityPaths) == 0 {
		return powerProfile{PowerSource: "unknown"}
	}
	capacityData, err := os.ReadFile(capacityPaths[0])
	if err != nil {
		return powerProfile{PowerSource: "unknown"}
	}
	percent := parseBatteryPercent(string(capacityData))
	statusPath := filepath.Join(filepath.Dir(capacityPaths[0]), "status")
	statusData, _ := os.ReadFile(statusPath)
	status := normalizeName(string(statusData))
	source := "unknown"
	switch status {
	case "discharging":
		source = "battery"
	case "charging", "full", "not charging":
		source = "ac"
	}
	return powerProfile{
		PowerSource:    source,
		BatteryPercent: percent,
		LowBattery:     source == "battery" && percent > 0 && percent <= 25,
	}
}

func parseBatteryPercent(value string) int {
	match := regexp.MustCompile(`(\d+)\s*%?`).FindStringSubmatch(value)
	if len(match) < 2 {
		return 0
	}
	parsed, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	if parsed < 0 {
		return 0
	}
	if parsed > 100 {
		return 100
	}
	return parsed
}

func localSLMEndpointSettingLine() string {
	endpoint, source := resolvedLocalSLMEndpoint()
	switch {
	case endpoint == "":
		return "JINI_LOCAL_SLM_ENDPOINT: missing"
	case source == "auto":
		return "JINI_LOCAL_SLM_ENDPOINT: auto -> " + endpoint
	default:
		return "JINI_LOCAL_SLM_ENDPOINT: " + presentOrMissing("JINI_LOCAL_SLM_ENDPOINT")
	}
}

func localSLMModelSettingLine(envName string, modelID, modelLabel string) string {
	if strings.TrimSpace(configValue(envName)) != "" {
		if envName == "JINI_LOCAL_SLM_MODEL" {
			return envName + ": set -> " + firstNonEmpty(modelLabel, modelID)
		}
		return envName + ": set"
	}
	if strings.TrimSpace(modelID) != "" {
		return envName + ": auto -> " + firstNonEmpty(modelLabel, modelID)
	}
	return envName + ": missing"
}

func localSLMModelSelectionReasonSuffix() string {
	power := currentPowerProfile()
	if power.LowBattery {
		return " Battery is low, so Jini biases toward smaller local models."
	}
	return ""
}

func localSLMRuntimeSetupIssue() string {
	discovery := localSLMRuntimeDiscoveryResult()
	if strings.TrimSpace(discovery.Endpoint) == "" {
		return "JINI_LOCAL_SLM_ENDPOINT or a running local OpenAI-compatible server on 127.0.0.1:11434, :1234, :8080, or :8000"
	}
	if len(discovery.Models) == 0 && strings.TrimSpace(resolveConfiguredLocalSLMAnyModel()) == "" {
		return "JINI_LOCAL_SLM_MODEL or at least one chat model exposed by the local server"
	}
	return ""
}

func resolveConfiguredLocalSLMAnyModel() string {
	return configFirstNonEmpty(
		"JINI_LOCAL_SLM_MODEL",
		"JINI_LOCAL_SLM_WORKHORSE_MODEL",
		"JINI_LOCAL_SLM_FAST_MODEL",
		"JINI_LOCAL_SLM_DEEP_MODEL",
		"JINI_LOCAL_SLM_MULTIMODAL_MODEL",
	)
}

func formatPowerProfileForRoute() string {
	power := currentPowerProfile()
	if power.PowerSource == "" || power.PowerSource == "unknown" {
		return ""
	}
	if power.BatteryPercent > 0 {
		return fmt.Sprintf("%s %d%%", power.PowerSource, power.BatteryPercent)
	}
	return power.PowerSource
}
