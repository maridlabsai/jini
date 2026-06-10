package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type routeTarget struct {
	ID           string
	Label        string
	ProviderMode string
}

type routeFeatures struct {
	WorkClass       string
	DepthClass      string
	ModalityClass   string
	ModalitySubtype string
	RequestCohort   string
	ArtifactFamily  string
	EffortClass     string
	DeviceClass     string
	PrefersCheapest bool
}

type routeCandidateScore struct {
	Mode  string
	Score int
}

type localSLMProfileSlot struct {
	ID    string
	Label string
}

type routeDecision struct {
	Active              bool
	ToolMode            string
	ToolLabel           string
	RoutePolicy         string
	FeedbackKey         string
	AutoMode            autoModePolicy
	ModelLabel          string
	ModelReason         string
	Provider            providerConfig
	ChosenAutomatically bool
	Reason              string
	ContinuityReason    string
	EffortLevel         string
	EffortReason        string
	VerificationLevel   string
	VerificationReason  string
	CLIHandoffReceipt   *cliHandoffReceipt
}

type autoModePolicy struct {
	FrameworkSwitching string `json:"framework_switching,omitempty"`
	ModelSwitching     string `json:"model_switching,omitempty"`
	SpeedSwitching     string `json:"speed_switching,omitempty"`
	UserApprovalMode   string `json:"user_approval_mode,omitempty"`
}

func routeTargetForMode(mode string) (routeTarget, bool) {
	descriptor, ok := adapterDescriptorForMode(mode)
	if !ok {
		return routeTarget{}, false
	}
	return routeTarget{ID: descriptor.ID, Label: descriptor.Label, ProviderMode: descriptor.ProviderMode}, true
}

func detectRoute() routeDecision {
	return detectRouteForRequest(providerGenerationRequest{})
}

func detectRouteForRequest(request providerGenerationRequest) routeDecision {
	toolMode := configuredToolMode()
	if toolMode == "" {
		return routeDecision{}
	}
	if toolMode == "auto" {
		return detectAutoRouteForRequest(request)
	}
	if toolMode == "local-slm" {
		return detectLocalSLMAliasRouteForRequest(request, false)
	}
	return enrichRouteDecisionForRequest(request, detectRouteForToolMode(toolMode, false))
}

func detectAutoRoute() routeDecision {
	return detectAutoRouteForRequest(providerGenerationRequest{})
}

func detectAutoRouteForRequest(request providerGenerationRequest) routeDecision {
	features := classifyRouteFeatures(request)
	availability := detectRuntimeAvailability(request)
	scored := routeCandidatesForRequest(request)
	if preserved, ok := preserveCurrentCodingRoute(request, features, scored); ok {
		preserved.Reason = appendRuntimeAvailabilityReason(explainAutoRouteChoice(features, preserved, false), availability)
		preserved.ContinuityReason = "Kept the current coding route to preserve context continuity because the quality gap was not material."
		return enrichRouteDecisionForRequest(request, preserved)
	}
	candidates := candidateModes(scored)
	switch configuredProviderMode() {
	case "anthropic":
		if !shouldSkipPinnedRemoteProvider(availability) {
			candidates = []string{"claude-api"}
		}
	case "bedrock":
		if !shouldSkipPinnedRemoteProvider(availability) {
			candidates = []string{"bedrock-sonnet"}
		}
	case "azure-openai":
		if !shouldSkipPinnedRemoteProvider(availability) {
			candidates = []string{"chatgpt", "azure-code", "azure-openai"}
		}
	case "local-slm":
		candidates = localSLMCandidateModesForScores(scored)
	case "local-preview":
		candidates = []string{"local-preview"}
	}
	if len(candidates) == 0 {
		candidates = []string{"local-preview"}
	}

	first := routeDecision{}
	for index, mode := range candidates {
		decision := detectRouteForToolMode(mode, true)
		decision.Reason = appendRuntimeAvailabilityReason(explainAutoRouteChoice(features, decision, index > 0), availability)
		decision = enrichRouteDecisionForRequest(request, decision)
		if index == 0 {
			first = decision
		}
		if decision.Provider.Status == "ok" {
			return decision
		}
	}
	return first
}

func shouldSkipPinnedRemoteProvider(availability runtimeAvailability) bool {
	return availability.OfflineMode && availability.ConnectivityState == "offline"
}

func enrichRouteDecisionForRequest(request providerGenerationRequest, decision routeDecision) routeDecision {
	if !decision.Active {
		return decision
	}
	if cliHandoffMode(decision.ToolMode) {
		decision.RoutePolicy = routePolicyForDecision(decision)
		decision.ModelLabel = cliHandoffLabel(decision.ToolMode)
		if decision.Provider.Status == "ok" {
			decision.ModelReason = "Jini hands this request to the installed CLI subprocess and records the route receipt."
		} else {
			decision.ModelReason = "Jini cannot claim this CLI route until it can hand off to the installed CLI."
		}
		decision.Provider = withRoutePolicy(decision.Provider, decision.RoutePolicy)
		decision.Provider = withRouteReason(decision.Provider, decision.Reason)
		decision.Provider = withModelDecision(decision.Provider, decision.ModelLabel, decision.ModelReason)
		return decision
	}
	decision.ModelLabel, decision.ModelReason = classifyModelChoice(request, decision.ToolMode)
	decision.EffortLevel, decision.EffortReason = classifyRequestEffort(request)
	decision.VerificationLevel, decision.VerificationReason = classifyVerificationDecision(request, decision)
	decision.RoutePolicy = routePolicyForDecision(decision)
	decision.AutoMode = autoModePolicyForDecision(decision)
	decision.FeedbackKey = routeFeedbackKeyForDecision(decision)
	decision.Provider = withRoutePolicy(decision.Provider, decision.RoutePolicy)
	decision.Provider = withRouteReason(decision.Provider, decision.Reason)
	decision.Provider = withModelDecision(decision.Provider, decision.ModelLabel, decision.ModelReason)
	decision.Provider = withEffortSetting(decision.Provider, decision.EffortLevel, decision.EffortReason)
	decision.Provider = withAutoModeSetting(decision.Provider, decision.AutoMode)
	return decision
}

func routeCandidatesForRequest(request providerGenerationRequest) []routeCandidateScore {
	features := classifyRouteFeatures(request)
	availability := detectRuntimeAvailability(request)
	scores := []routeCandidateScore{}
	for _, mode := range autoRouteCandidateModes() {
		if !routeModeAutoEligible(mode) {
			continue
		}
		scores = append(scores, routeCandidateScore{
			Mode:  mode,
			Score: scoreRouteMode(mode, features) + runtimeAvailabilityRouteBias(mode, availability),
		})
	}
	if !routeCandidateScoresContain(scores, "local-preview") {
		scores = append(scores, routeCandidateScore{
			Mode:  "local-preview",
			Score: scoreRouteMode("local-preview", features) + runtimeAvailabilityRouteBias("local-preview", availability),
		})
	}
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	return scores
}

func autoRouteCandidateModes() []string {
	modes := []string{}
	for _, target := range defaultSavedRouteTargets() {
		if !target.Enabled {
			continue
		}
		if descriptor, ok := adapterDescriptorForMode(target.ID); ok && descriptor.SupportsAutoRoute {
			modes = append(modes, descriptor.ID)
		}
	}
	return modes
}

func routeCandidateScoresContain(scores []routeCandidateScore, mode string) bool {
	for _, score := range scores {
		if score.Mode == mode {
			return true
		}
	}
	return false
}

func routeModeAutoEligible(mode string) bool {
	if mode == "local-preview" {
		return true
	}
	if cliHandoffMode(mode) {
		return detectCLIHandoffProvider(mode).Status == "ok"
	}
	descriptor, ok := adapterDescriptorForMode(mode)
	if !ok {
		return false
	}
	switch descriptor.ProviderMode {
	case "local-slm":
		return localSLMRouteModeEligible(mode)
	case "local-preview":
		return true
	default:
		return detectProviderForMode(descriptor.ProviderMode).Status == "ok"
	}
}

func anyCLIHandoffReady() bool {
	for _, target := range defaultSavedRouteTargets() {
		if !target.Enabled || !cliHandoffMode(target.ID) {
			continue
		}
		if detectCLIHandoffProvider(target.ID).Status == "ok" {
			return true
		}
	}
	return false
}

func candidateModes(scored []routeCandidateScore) []string {
	modes := make([]string, 0, len(scored))
	for _, candidate := range scored {
		modes = append(modes, candidate.Mode)
	}
	return modes
}

func routeCandidateLeadMargin(request providerGenerationRequest, selectedMode string) int {
	selectedMode = strings.TrimSpace(selectedMode)
	if selectedMode == "" {
		return -1
	}
	scored := routeCandidatesForRequest(request)
	selectedScore := -1
	bestOtherScore := -1
	for _, candidate := range scored {
		if candidate.Mode == selectedMode && selectedScore < 0 {
			selectedScore = candidate.Score
			continue
		}
		if candidate.Score > bestOtherScore {
			bestOtherScore = candidate.Score
		}
	}
	if selectedScore < 0 {
		return -1
	}
	if bestOtherScore < 0 {
		return selectedScore
	}
	return selectedScore - bestOtherScore
}

func classifyRouteFeatures(request providerGenerationRequest) routeFeatures {
	depthClass := classifyRouteDepth(request)
	effortLevel, _ := classifyRequestEffort(request)
	device := currentDeviceProfile()
	return routeFeatures{
		WorkClass:       classifyRouteWork(request),
		DepthClass:      depthClass,
		ModalityClass:   classifyRouteModality(request),
		ModalitySubtype: classifyRouteModalitySubtype(request),
		RequestCohort:   classifyRequestCohort(request),
		ArtifactFamily:  classifyArtifactFamily(request),
		EffortClass:     effortLevel,
		DeviceClass:     device.DeviceClass,
		PrefersCheapest: depthClass != "deep",
	}
}

func scoreRouteMode(mode string, features routeFeatures) int {
	score := 0
	switch mode {
	case "codex":
		score = 66
	case "claude-code":
		score = 65
	case "gemini-cli":
		score = 58
	case "aider":
		score = 62
	case "opencode":
		score = 60
	case "claude-api":
		score = 62
	case "bedrock-sonnet":
		score = 60
	case "chatgpt":
		score = 58
	case "azure-code":
		score = 58
	case "azure-openai":
		score = 52
	case "local-preview":
		score = 5
	case "local-fast":
		score = 60
	case "local-workhorse":
		score = 66
	case "local-deep":
		score = 67
	case "local-multimodal":
		score = 64
	}

	switch features.WorkClass {
	case "planning":
		switch mode {
		case "chatgpt":
			score += 14
		case "azure-openai":
			score += 10
		case "bedrock-sonnet":
			score += 13
		case "claude-api":
			score += 8
		case "azure-code":
			score -= 8
		case "local-workhorse":
			score += 28
		case "local-fast":
			score += 12
		case "local-deep":
			score += 8
		}
	case "code":
		switch mode {
		case "codex":
			score += 28
		case "claude-code":
			score += 26
		case "aider":
			score += 24
		case "opencode":
			score += 20
		case "gemini-cli":
			score += 14
		case "azure-code":
			score += 24
		case "claude-api":
			score += 18
		case "azure-openai":
			score += 9
		case "bedrock-sonnet":
			score += 7
		case "chatgpt":
			score -= 6
		case "local-workhorse":
			score += 14
		case "local-deep":
			score += 12
		case "local-fast":
			score += 6
		}
	default:
		switch mode {
		case "chatgpt", "claude-api", "claude-code", "codex":
			score += 10
		case "azure-code", "bedrock-sonnet", "azure-openai", "gemini-cli", "aider", "opencode":
			score += 7
		case "local-fast":
			score += 18
		case "local-workhorse":
			if features.ModalityClass == "multimodal" {
				score += 4
			} else {
				score += 14
			}
		case "local-deep":
			score += 6
		case "local-multimodal":
			if features.ModalityClass == "multimodal" {
				score += 18
			}
		}
	}

	if features.DepthClass == "deep" {
		switch mode {
		case "claude-api", "claude-code", "codex":
			score += 24
		case "bedrock-sonnet", "gemini-cli":
			score += 23
		case "azure-code", "aider", "opencode":
			score += 9
		case "chatgpt":
			score += 7
		case "azure-openai":
			score += 5
		case "local-deep":
			score += 22
		case "local-workhorse":
			score += 6
		}
	} else if features.PrefersCheapest {
		switch mode {
		case "chatgpt", "azure-code", "codex", "aider", "opencode":
			score += 10
		case "azure-openai", "gemini-cli":
			score += 8
		case "local-fast":
			score += 20
		case "local-workhorse":
			score += 18
		case "local-deep":
			score += 4
		case "claude-api":
			score += 5
		case "bedrock-sonnet":
			score += 3
		}
	}

	switch features.EffortClass {
	case "low":
		switch mode {
		case "local-fast":
			score += 18
		case "local-workhorse":
			score += 8
		}
	case "medium":
		switch mode {
		case "local-workhorse":
			score += 10
		case "local-fast":
			score += 6
		}
	case "high", "extra high":
		switch mode {
		case "local-deep":
			score += 12
		case "local-workhorse":
			score += 5
		}
	}

	if features.ModalityClass == "multimodal" {
		switch mode {
		case "local-multimodal":
			score += 18
		case "azure-openai", "bedrock-sonnet":
			score += 8
		case "claude-api", "chatgpt":
			score += 4
		case "local-workhorse":
			score += 2
		case "local-fast":
			score -= 18
		}
		switch features.ModalitySubtype {
		case "pdf-scan":
			switch mode {
			case "local-multimodal":
				score += 34
			case "azure-openai":
				score += 8
			case "bedrock-sonnet":
				score += 6
			case "local-workhorse":
				score -= 2
			case "local-fast":
				score -= 10
			}
		case "image-screenshot":
			switch mode {
			case "local-multimodal":
				score += 38
			case "azure-openai", "bedrock-sonnet":
				score += 5
			case "local-workhorse":
				score -= 4
			case "local-fast":
				score -= 12
			}
		case "audio-transcript":
			switch mode {
			case "local-workhorse":
				score += 12
			case "local-deep":
				score += 6
			case "azure-openai", "chatgpt":
				score += 8
			case "bedrock-sonnet":
				score += 5
			case "local-multimodal":
				score += 4
			case "local-fast":
				score -= 8
			}
		}
	}
	score += localDeviceCapabilityBias(mode, currentDeviceProfile())
	score += localBenchmarkBiasForFeatures(mode, features)
	score += codingQuotaHeadroomBias(mode, features)
	score += codingManualOverrideBias(mode, features)
	return score + routeFeedbackBias(mode)
}

func codingQuotaHeadroomBias(mode string, features routeFeatures) int {
	if strings.TrimSpace(features.WorkClass) != "code" {
		return 0
	}
	switch mode {
	case "local-workhorse":
		return 12
	case "local-deep":
		return 8
	case "azure-code":
		return 8
	case "azure-openai":
		return 6
	case "chatgpt":
		return 4
	case "claude-api":
		return 2
	case "bedrock-sonnet":
		return 1
	case "local-fast":
		return 1
	default:
		return 0
	}
}

func codingManualOverrideBias(mode string, features routeFeatures) int {
	if strings.TrimSpace(features.WorkClass) != "code" {
		return 0
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForFeatures(features))
	if cohort == "" {
		return 0
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.ManualOverrides == nil {
		return 0
	}
	bias := stats.ManualOverrides[strings.TrimSpace(mode)][cohort] * 6
	if bias > 24 {
		return 24
	}
	return bias
}

func codingManualOverrideNote(mode string, features routeFeatures) string {
	if strings.TrimSpace(features.WorkClass) != "code" {
		return ""
	}
	cohort := strings.TrimSpace(localBenchmarkCohortKeyForFeatures(features))
	if cohort == "" {
		return ""
	}
	stats := loadLocalRouteFeedbackStats()
	if stats.ManualOverrides == nil {
		return ""
	}
	if stats.ManualOverrides[strings.TrimSpace(mode)][cohort] <= 0 {
		return ""
	}
	return " Past route choices on similar coding work also favor this route."
}

func preserveCurrentCodingRoute(request providerGenerationRequest, features routeFeatures, scored []routeCandidateScore) (routeDecision, bool) {
	if strings.TrimSpace(features.WorkClass) != "code" {
		return routeDecision{}, false
	}
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		return routeDecision{}, false
	}
	if strings.TrimSpace(current.PackDir) == "" || strings.TrimSpace(current.PackID) == "" {
		return routeDecision{}, false
	}
	if request.Choice.PackID != "" && strings.TrimSpace(request.Choice.PackID) != strings.TrimSpace(current.PackID) {
		return routeDecision{}, false
	}
	route := loadWorkRoute(current.PackDir)
	mode := strings.TrimSpace(route.ToolMode)
	if mode == "" || mode == "local-preview" {
		return routeDecision{}, false
	}
	bestScore := -999
	currentScore := -999
	for _, candidate := range scored {
		if candidate.Score > bestScore {
			bestScore = candidate.Score
		}
		if candidate.Mode == mode {
			currentScore = candidate.Score
		}
	}
	if currentScore <= -999 {
		return routeDecision{}, false
	}
	allowedGap := 14
	if strings.TrimSpace(features.DepthClass) == "deep" || strings.TrimSpace(features.EffortClass) == "extra high" {
		allowedGap = 4
	}
	if bestScore-currentScore > allowedGap {
		return routeDecision{}, false
	}
	decision := detectRouteForToolMode(mode, true)
	if decision.Provider.Status != "ok" {
		return routeDecision{}, false
	}
	return decision, true
}

func explainAutoRouteChoice(features routeFeatures, decision routeDecision, readinessFallback bool) string {
	parts := []string{}
	switch features.WorkClass {
	case "planning":
		parts = append(parts, "this looks like planning work")
	case "code":
		parts = append(parts, "this looks like code-heavy work")
	default:
		parts = append(parts, "this looks like general work")
	}
	switch features.ModalitySubtype {
	case "pdf-scan":
		parts = append(parts, "the request depends on scanned document or PDF evidence")
	case "image-screenshot":
		parts = append(parts, "the request depends on image or screenshot evidence")
	case "audio-transcript":
		parts = append(parts, "the request depends on audio or transcript evidence")
	}
	if features.DepthClass == "deep" {
		parts = append(parts, "the request asks for deeper or more rigorous work")
		reason := "Auto mode chose " + decision.ToolLabel + " because " + strings.Join(parts, ", ") + ", so Jini favored the strongest suitable route."
		if strings.HasPrefix(decision.ToolMode, "local-") {
			reason += multimodalLearningSeparationNote(features)
		}
		reason += codingManualOverrideNote(decision.ToolMode, features)
		if readinessFallback {
			reason += " It was the first ready route in this environment."
		}
		return reason
	}
	parts = append(parts, "the request does not ask for deep review")
	reason := "Auto mode chose " + decision.ToolLabel + " because " + strings.Join(parts, ", ") + ", so Jini favored the cheapest suitable route."
	if strings.HasPrefix(decision.ToolMode, "local-") {
		reason += multimodalLearningSeparationNote(features)
	}
	reason += codingManualOverrideNote(decision.ToolMode, features)
	if readinessFallback {
		reason += " It was the first ready route in this environment."
	}
	return reason
}

func classifyRouteModality(request providerGenerationRequest) string {
	if classifyRouteModalitySubtype(request) != "" {
		return "multimodal"
	}
	return "text"
}

func classifyRouteModalitySubtype(request providerGenerationRequest) string {
	source := normalizeName(strings.Join([]string{
		request.Choice.PackID,
		request.Title,
		request.Source,
	}, " "))
	switch {
	case containsAny(source, []string{
		"audio", "voice note", "recording", "transcript", "call recording",
		"meeting recording", "voice memo", "podcast",
	}):
		return "audio-transcript"
	case containsAny(source, []string{
		"pdf", "scan", "scanned", "document", "invoice", "report", "form",
		"ocr", "page",
	}):
		return "pdf-scan"
	case containsAny(source, []string{
		"image", "images", "screenshot", "screenshots", "photo", "picture",
		"diagram", "ui capture",
	}):
		return "image-screenshot"
	default:
		return ""
	}
}

func localSLMProfileSlots() []localSLMProfileSlot {
	slots := []localSLMProfileSlot{}
	for _, mode := range localBenchmarkableAdapterModes() {
		descriptor, ok := adapterDescriptorForMode(mode)
		if !ok {
			continue
		}
		slots = append(slots, localSLMProfileSlot{ID: descriptor.ID, Label: descriptor.Label})
	}
	return slots
}

func localSLMRuntimeReady() bool {
	return detectProviderForMode("local-slm").Status == "ok"
}

func localSLMRouteModeEligible(mode string) bool {
	if !localSLMRuntimeReady() {
		return false
	}
	descriptor, ok := adapterDescriptorForMode(mode)
	if !ok || descriptor.ProviderMode != "local-slm" {
		return false
	}
	return strings.TrimSpace(currentDeviceProfile().LocalProfileStates[mode]) != "unavailable"
}

func detectLocalSLMAliasRouteForRequest(request providerGenerationRequest, auto bool) routeDecision {
	mode := bestLocalSLMRouteModeForRequest(request)
	if mode == "" || mode == "local-preview" {
		provider := detectLocalSLMProvider()
		return routeDecision{
			Active:              true,
			ToolMode:            "local-slm",
			ToolLabel:           "Local SLM",
			RoutePolicy:         "Local SLM required",
			FeedbackKey:         routeFeedbackKeyForCurrentMode("local-slm"),
			ModelLabel:          "Local SLM",
			ModelReason:         "Jini cannot use Local SLM until an eligible local profile is available.",
			ChosenAutomatically: auto,
			Reason:              "Local SLM is selected, but no eligible local profile is ready for this device and request.",
			Provider:            withRoutePolicy(provider, "Local SLM required"),
		}
	}
	return enrichRouteDecisionForRequest(request, detectRouteForToolMode(mode, auto))
}

func localDeviceCapabilityBias(mode string, profile deviceProfile) int {
	state := strings.TrimSpace(profile.LocalProfileStates[mode])
	power := currentPowerProfile()
	powerBias := 0
	if power.LowBattery {
		switch mode {
		case "local-fast":
			powerBias = 24
		case "local-workhorse":
			powerBias = -6
		case "local-deep":
			powerBias = -30
		case "local-multimodal":
			powerBias = -24
		}
	}
	switch state {
	case "unavailable":
		return -100
	case "limited":
		switch mode {
		case "local-fast":
			return 4 + powerBias
		case "local-workhorse":
			return -4 + powerBias
		case "local-deep":
			return -18 + powerBias
		case "local-multimodal":
			return -22 + powerBias
		}
	case "available":
		switch deviceClassForPolicy(profile.DeviceClass) {
		case "mobile-small":
			switch mode {
			case "local-fast":
				return 22 + powerBias
			case "local-workhorse", "local-deep", "local-multimodal":
				return -40 + powerBias
			}
		case "tiny":
			switch mode {
			case "local-fast":
				return 16 + powerBias
			case "local-workhorse":
				return -8 + powerBias
			case "local-deep", "local-multimodal":
				return -24 + powerBias
			}
		case "laptop-light":
			switch mode {
			case "local-fast":
				return 12 + powerBias
			case "local-workhorse":
				return 6 + powerBias
			case "local-deep":
				return -8 + powerBias
			case "local-multimodal":
				return -14 + powerBias
			}
		case "laptop-pro":
			switch mode {
			case "local-fast":
				return 6 + powerBias
			case "local-workhorse":
				return 10 + powerBias
			case "local-deep":
				return 2 + powerBias
			case "local-multimodal":
				return -4 + powerBias
			}
		case "workstation":
			switch mode {
			case "local-fast":
				return 2 + powerBias
			case "local-workhorse":
				return 10 + powerBias
			case "local-deep":
				return 10 + powerBias
			case "local-multimodal":
				return 8 + powerBias
			}
		case "gpu-heavy":
			switch mode {
			case "local-fast":
				return 2 + powerBias
			case "local-workhorse":
				return 8 + powerBias
			case "local-deep":
				return 14 + powerBias
			case "local-multimodal":
				return 16 + powerBias
			}
		}
	}
	return powerBias
}

func localSLMCandidateModes() []string {
	modes := make([]string, 0, len(localSLMProfileSlots()))
	for _, slot := range localSLMProfileSlots() {
		if !localSLMRouteModeEligible(slot.ID) {
			continue
		}
		modes = append(modes, slot.ID)
	}
	return modes
}

func localSLMCandidateModesForScores(scored []routeCandidateScore) []string {
	allowed := map[string]bool{}
	for _, mode := range localSLMCandidateModes() {
		allowed[mode] = true
	}
	modes := []string{}
	for _, candidate := range scored {
		if allowed[candidate.Mode] {
			modes = append(modes, candidate.Mode)
		}
	}
	return modes
}

func classifyRouteWork(request providerGenerationRequest) string {
	source := normalizeName(strings.Join([]string{
		request.Choice.PackID,
		request.Title,
		request.Source,
	}, " "))
	if workClass := starterWorkClass(request.Choice.PackID); workClass != "" {
		return workClass
	}
	if containsAny(source, []string{
		"fix", "bug", "failing", "test", "repo", "repository", "code", "implement",
		"refactor", "compile", "ci", "build", "stack trace", "cli", "function", "package",
	}) {
		return "code"
	}
	if containsAny(source, []string{
		"trip", "travel", "itinerary", "meeting", "follow up", "followup", "plan",
		"compare", "options", "decision", "budget", "hotel", "flight", "memo",
	}) {
		return "planning"
	}
	return "general"
}

func classifyRequestCohort(request providerGenerationRequest) string {
	source := normalizeName(strings.Join([]string{
		request.Choice.PackID,
		request.Title,
		request.Source,
	}, " "))
	if cohort := starterRequestCohort(request.Choice.PackID); cohort != "" {
		return cohort
	}
	switch {
	case containsAny(source, []string{"meeting", "follow up", "followup", "sendable", "owners", "due date"}):
		return "sendable-followup"
	case containsAny(source, []string{"readiness", "ready to hand off", "handoff", "spec", "build", "missing pieces"}):
		return "build-readiness"
	case containsAny(source, []string{"trip", "travel", "itinerary", "hotel", "flight"}):
		return "trip-itinerary"
	case containsAny(source, []string{"vendor", "option", "compare", "shortlist", "choose one"}):
		return "option-compare"
	case containsAny(source, []string{"incident", "outage", "cleanup", "postmortem", "root cause"}):
		return "incident-cleanup"
	case containsAny(source, []string{"image", "audio", "pdf", "screenshot", "photo"}):
		return "multimodal-extract"
	case containsAny(source, []string{"fix", "bug", "repo", "code", "refactor", "test", "implement"}):
		return "code-change"
	default:
		return "general-pass"
	}
}

func classifyArtifactFamily(request providerGenerationRequest) string {
	if family := starterArtifactFamily(request.Choice.PackID); family != "" {
		return family
	}
	switch classifyRequestCohort(request) {
	case "sendable-followup":
		return "narrative-draft"
	case "build-readiness":
		return "structured-check"
	case "trip-itinerary":
		return "itinerary-plan"
	case "option-compare":
		return "comparison-matrix"
	case "incident-cleanup":
		return "step-plan"
	case "multimodal-extract":
		return "multimodal-extract"
	case "code-change":
		return "code-change"
	default:
		return "general-pass"
	}
}

func classifyRouteDepth(request providerGenerationRequest) string {
	source := normalizeName(strings.Join([]string{
		request.Choice.ChoiceLabel,
		request.Title,
		request.Source,
	}, " "))
	if containsAny(source, []string{
		"deep work", "deep", "thorough", "careful", "rigorous", "high rigor",
		"benchmark", "root cause", "root-cause", "architecture", "architectural",
		"hard criticism", "stress test", "critique", "compare deeply", "production ready",
		"production-ready", "end to end", "end-to-end", "comprehensive", "exhaustive",
	}) {
		return "deep"
	}
	return "standard"
}

func detectRouteForToolMode(mode string, auto bool) routeDecision {
	if cliHandoffMode(mode) {
		label := cliHandoffLabel(mode)
		provider := detectCLIHandoffProvider(mode)
		routePolicy := "CLI handoff"
		reason := "Jini will hand this request to the installed " + label + " subprocess."
		modelReason := "Jini hands this request to the installed CLI subprocess and records the route receipt."
		if provider.Status != "ok" {
			routePolicy = "CLI handoff required"
			reason = "Jini will not treat " + label + " as a provider API route."
			modelReason = "Jini cannot claim this CLI route until it can hand off to the installed CLI."
		}
		provider = withRoutePolicy(provider, routePolicy)
		provider = withRouteReason(provider, reason)
		provider = withModelDecision(provider, label, modelReason)
		return routeDecision{
			Active:              true,
			ToolMode:            mode,
			ToolLabel:           label,
			RoutePolicy:         routePolicy,
			FeedbackKey:         routeFeedbackKeyForCurrentMode(mode),
			ModelLabel:          label,
			ModelReason:         modelReason,
			ChosenAutomatically: auto,
			Reason:              reason,
			Provider:            provider,
		}
	}
	target, ok := routeTargetForMode(mode)
	if !ok {
		return routeDecision{
			Active:              true,
			ToolMode:            mode,
			ToolLabel:           titleCase(mode),
			ChosenAutomatically: auto,
			Provider: providerConfig{
				ID:      mode,
				Label:   titleCase(mode),
				Status:  "needs setup",
				Missing: []string{"Supported JINI_TOOL value: auto, codex, claude-code, gemini-cli, aider, opencode, claude-api, bedrock-sonnet, chatgpt, azure-code, azure-openai, local-fast, local-workhorse, local-deep, local-multimodal, or local-preview."},
			},
		}
	}

	provider := detectProviderForMode(target.ProviderMode)
	if configuredProviderMode() == "auto" {
		provider = withAutoProviderSetting(provider)
	}
	provider = withToolSetting(provider, mode, target.Label, auto)
	routePolicy := routePolicyForDecision(routeDecision{Active: true, ToolMode: mode, ToolLabel: target.Label, ChosenAutomatically: auto, Provider: provider})
	provider = withRoutePolicy(provider, routePolicy)
	modelLabel := modelLabelForToolMode(mode)
	modelReason := modelReasonForToolMode(mode)
	provider = withModelDecision(provider, modelLabel, modelReason)
	reason := routeReasonForToolMode(mode, target.Label, auto)
	provider = withRouteReason(provider, reason)

	return routeDecision{
		Active:              true,
		ToolMode:            mode,
		ToolLabel:           target.Label,
		RoutePolicy:         routePolicy,
		FeedbackKey:         routeFeedbackKeyForCurrentMode(mode),
		ModelLabel:          modelLabel,
		ModelReason:         modelReason,
		ChosenAutomatically: auto,
		Reason:              reason,
		Provider:            provider,
	}
}

func routeReasonForToolMode(mode, label string, auto bool) string {
	if auto {
		return ""
	}
	if strings.TrimSpace(label) == "" {
		label = titleCase(mode)
	}
	if strings.TrimSpace(label) == "" {
		return ""
	}
	return "Jini kept the route you chose for this repo: " + label + "."
}

func routePolicyForDecision(decision routeDecision) string {
	if !decision.Active {
		return ""
	}
	if cliHandoffMode(decision.ToolMode) {
		if decision.Provider.Status == "ok" {
			return "CLI handoff"
		}
		return "CLI handoff required"
	}
	if decision.ChosenAutomatically {
		if decision.Provider.ID == "local-preview" {
			return "Fallback"
		}
		return "Automatic"
	}
	return "Locked by you"
}

func autoModePolicyForDecision(decision routeDecision) autoModePolicy {
	if !decision.Active || !decision.ChosenAutomatically {
		return autoModePolicy{}
	}
	return autoModePolicy{
		FrameworkSwitching: "auto",
		ModelSwitching:     "auto",
		SpeedSwitching:     "auto",
		UserApprovalMode:   "approval-gated",
	}
}

func autoModePolicyEnabled(policy autoModePolicy) bool {
	return strings.TrimSpace(policy.FrameworkSwitching) != "" ||
		strings.TrimSpace(policy.ModelSwitching) != "" ||
		strings.TrimSpace(policy.SpeedSwitching) != "" ||
		strings.TrimSpace(policy.UserApprovalMode) != ""
}

func autoModePolicyPointer(policy autoModePolicy) *autoModePolicy {
	if !autoModePolicyEnabled(policy) {
		return nil
	}
	return &policy
}

func autoModePolicyValue(policy *autoModePolicy) autoModePolicy {
	if policy == nil {
		return autoModePolicy{}
	}
	return *policy
}

func withToolSetting(provider providerConfig, toolMode, toolLabel string, auto bool) providerConfig {
	line := ""
	switch {
	case toolMode == "":
		return provider
	case auto:
		line = "JINI_TOOL: auto -> " + toolLabel
	default:
		line = "JINI_TOOL: " + toolMode + " -> " + toolLabel
	}

	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "JINI_TOOL:") {
			return provider
		}
	}
	provider.Settings = append([]string{line}, provider.Settings...)
	return provider
}

func withAutoModeSetting(provider providerConfig, policy autoModePolicy) providerConfig {
	if !autoModePolicyEnabled(policy) {
		return provider
	}
	line := "AUTO_MODE: frameworks=" + policy.FrameworkSwitching + "; models=" + policy.ModelSwitching + "; speed=" + policy.SpeedSwitching + "; approvals=" + policy.UserApprovalMode
	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "AUTO_MODE:") {
			return provider
		}
	}
	provider.Settings = append(provider.Settings, line)
	return provider
}

func withRouteReason(provider providerConfig, reason string) providerConfig {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return provider
	}
	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "AUTO_ROUTE:") {
			return provider
		}
	}
	provider.Settings = append(provider.Settings, "AUTO_ROUTE: "+reason)
	return provider
}

func withRoutePolicy(provider providerConfig, policy string) providerConfig {
	policy = strings.TrimSpace(policy)
	if policy == "" {
		return provider
	}
	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "ROUTE_POLICY:") {
			return provider
		}
	}
	provider.Settings = append(provider.Settings, "ROUTE_POLICY: "+policy)
	return provider
}

func withModelDecision(provider providerConfig, label, reason string) providerConfig {
	label = strings.TrimSpace(label)
	if label != "" {
		already := false
		for _, existing := range provider.Settings {
			if strings.HasPrefix(existing, "JINI_MODEL_DECISION:") {
				already = true
				break
			}
		}
		if !already {
			provider.Settings = append(provider.Settings, "JINI_MODEL_DECISION: "+label)
		}
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return provider
	}
	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "AUTO_MODEL:") {
			return provider
		}
	}
	provider.Settings = append(provider.Settings, "AUTO_MODEL: "+reason)
	return provider
}

func withEffortSetting(provider providerConfig, level, reason string) providerConfig {
	level = strings.TrimSpace(level)
	if level == "" {
		return provider
	}
	for _, existing := range provider.Settings {
		if strings.HasPrefix(existing, "JINI_EFFORT:") {
			return provider
		}
	}
	provider.Settings = append(provider.Settings, "JINI_EFFORT: auto -> "+level)
	if strings.TrimSpace(reason) != "" {
		provider.Settings = append(provider.Settings, "AUTO_EFFORT: "+reason)
	}
	return provider
}

func classifyRequestEffort(request providerGenerationRequest) (string, string) {
	if strings.TrimSpace(request.Choice.PackID) == "" && strings.TrimSpace(request.Title) == "" && strings.TrimSpace(request.Source) == "" {
		return "dynamic per request", "Jini judges effort separately for each request instead of pinning one level globally."
	}
	source := normalizeName(strings.Join([]string{
		request.Choice.ChoiceLabel,
		request.Title,
		request.Source,
	}, " "))
	switch {
	case containsAny(source, []string{
		"benchmark", "root cause", "root-cause", "production ready", "production-ready",
		"architecture", "architectural", "exhaustive", "hard criticism", "stress test",
		"deeply compare", "deep comparison", "threat model",
	}):
		return "extra high", "Jini judged this request as extra high effort because it asks for rigorous, high-stakes, or exhaustive work."
	case containsAny(source, []string{
		"deep work", "deep", "thorough", "careful", "rigorous", "high rigor",
		"comprehensive", "end to end", "end-to-end", "critique", "full review",
	}):
		return "high", "Jini judged this request as high effort because it asks for deeper or more careful work."
	case containsAny(source, []string{
		"quick", "brief", "simple", "fast", "one line", "one-liner", "lightweight",
		"short draft", "rough draft",
	}):
		return "low", "Jini judged this request as low effort because it asks for a quick or lightweight pass."
	default:
		return "medium", "Jini judged this request as medium effort because the request needs a normal balanced pass."
	}
}

func classifyModelChoice(request providerGenerationRequest, toolMode string) (string, string) {
	if strings.TrimSpace(request.Choice.PackID) == "" && strings.TrimSpace(request.Title) == "" && strings.TrimSpace(request.Source) == "" {
		return modelLabelForToolMode(toolMode), "Jini judges the model together with the tool choice for each request."
	}
	switch {
	case strings.TrimSpace(configValue("BEDROCK_MODEL_ID")) != "" && toolMode == "bedrock-sonnet":
		return friendlyModelLabel("bedrock", configValue("BEDROCK_MODEL_ID")), "Jini kept the exact Bedrock model id you configured."
	case isBedrockOnlyModelMode(normalizeName(configuredModelInput())) && toolMode == "bedrock-sonnet":
		return "Claude Sonnet 4.6", "Jini used the Bedrock Sonnet 4.6 path because the request or config asked for a Bedrock-only model."
	case toolMode == "bedrock-sonnet":
		return "Claude Sonnet 4.6", "Jini uses Claude Sonnet 4.6 by default on the Bedrock Sonnet route."
	case toolMode == "claude-api":
		return "Claude Sonnet 4", "Jini uses Claude Sonnet 4 by default on the Claude API route."
	case toolMode == "chatgpt" || toolMode == "azure-code" || toolMode == "azure-openai":
		deployment := strings.TrimSpace(configValue("AZURE_OPENAI_DEPLOYMENT"))
		if deployment == "" {
			return "Azure deployment", "Jini uses the Azure deployment configured for this repo."
		}
		return deployment, "Jini uses the Azure deployment configured for this repo. The deployment decides the actual Azure model."
	case strings.HasPrefix(toolMode, "local-"):
		modelID, modelLabel := resolveLocalSLMModelForToolMode(toolMode)
		deviceClass := strings.TrimSpace(currentDeviceProfile().DeviceClass)
		suffix := localBenchmarkReasonSuffixForRequest(toolMode, request)
		separationNote := multimodalLearningSeparationNote(classifyRouteFeatures(request))
		powerNote := localSLMModelSelectionReasonSuffix()
		if strings.TrimSpace(modelID) == "" {
			return "Local SLM", "Jini will use the configured Local SLM profile for this route." + suffix + separationNote + powerNote
		}
		if deviceClass == "" {
			return modelLabel, "Jini uses the Local SLM profile mapped to this route." + suffix + separationNote + powerNote
		}
		return modelLabel, "Jini uses the Local SLM profile mapped to this route for a " + deviceClass + " device." + suffix + separationNote + powerNote
	case toolMode == "local-preview":
		return "Local preview", "Jini stays local when no cloud route is selected or ready."
	default:
		return modelLabelForToolMode(toolMode), "Jini chose the default model for the selected tool route."
	}
}

func modelLabelForToolMode(toolMode string) string {
	switch toolMode {
	case "bedrock-sonnet":
		return "Claude Sonnet 4.6"
	case "claude-api":
		return "Claude Sonnet 4"
	case "chatgpt", "azure-code", "azure-openai":
		if deployment := strings.TrimSpace(configValue("AZURE_OPENAI_DEPLOYMENT")); deployment != "" {
			return deployment
		}
		return "Azure deployment"
	case "local-fast", "local-workhorse", "local-deep", "local-multimodal":
		_, label := resolveLocalSLMModelForToolMode(toolMode)
		return firstNonEmpty(label, "Local SLM")
	case "local-preview":
		return "Local preview"
	default:
		return ""
	}
}

func modelReasonForToolMode(toolMode string) string {
	switch toolMode {
	case "bedrock-sonnet":
		return "Jini uses Claude Sonnet 4.6 by default on the Bedrock Sonnet route."
	case "claude-api":
		return "Jini uses Claude Sonnet 4 by default on the Claude API route."
	case "chatgpt", "azure-code", "azure-openai":
		return "Jini uses the Azure deployment configured for this repo. The deployment decides the actual Azure model."
	case "local-fast":
		return "Jini uses the Local SLM fast profile for lightweight first passes."
	case "local-workhorse":
		return "Jini uses the Local SLM workhorse profile for normal drafting and planning work."
	case "local-deep":
		return "Jini uses the Local SLM deep profile for stronger local reasoning before escalating to paid routes."
	case "local-multimodal":
		return "Jini uses the Local SLM multimodal profile when the request depends on images, PDFs, or audio."
	case "local-preview":
		return "Jini stays local when no cloud route is selected or ready."
	default:
		return ""
	}
}

type savedWorkRoute struct {
	SchemaVersion                string             `json:"schema_version"`
	ContextType                  string             `json:"context_type"`
	ToolMode                     string             `json:"tool_mode"`
	ToolLabel                    string             `json:"tool_label"`
	RoutePolicy                  string             `json:"route_policy"`
	AutoMode                     *autoModePolicy    `json:"auto_mode,omitempty"`
	FeedbackKey                  string             `json:"feedback_key,omitempty"`
	ModelLabel                   string             `json:"model_label"`
	ModelReason                  string             `json:"model_reason"`
	ProviderLabel                string             `json:"provider_label"`
	ChosenAutomatically          bool               `json:"chosen_automatically"`
	Reason                       string             `json:"reason"`
	ContinuityReason             string             `json:"continuity_reason,omitempty"`
	EffortLevel                  string             `json:"effort_level"`
	EffortReason                 string             `json:"effort_reason"`
	VerificationLevel            string             `json:"verification_level,omitempty"`
	VerificationReason           string             `json:"verification_reason,omitempty"`
	ModelFeedback                string             `json:"model_feedback,omitempty"`
	ArtifactFeedbackPath         string             `json:"artifact_feedback_path,omitempty"`
	ArtifactFeedbackReason       string             `json:"artifact_feedback_reason,omitempty"`
	ArtifactEditClass            string             `json:"artifact_edit_class,omitempty"`
	ArtifactEditScope            string             `json:"artifact_edit_scope,omitempty"`
	ArtifactSemanticClass        string             `json:"artifact_semantic_class,omitempty"`
	ArtifactOutcomeSignal        string             `json:"artifact_outcome_signal,omitempty"`
	ArtifactOutcomeReason        string             `json:"artifact_outcome_reason,omitempty"`
	PassiveArtifactOutcomeSignal string             `json:"passive_artifact_outcome_signal,omitempty"`
	PassiveArtifactOutcomeReason string             `json:"passive_artifact_outcome_reason,omitempty"`
	PreviousToolMode             string             `json:"previous_tool_mode,omitempty"`
	RouteSwitchCount             int                `json:"route_switch_count,omitempty"`
	LastRouteSwitchReason        string             `json:"last_route_switch_reason,omitempty"`
	CLIHandoffReceipt            *cliHandoffReceipt `json:"cli_handoff_receipt,omitempty"`
}

type artifactFeedbackBaseline struct {
	SchemaVersion string            `json:"schema_version"`
	ContextType   string            `json:"context_type"`
	Views         map[string]string `json:"views"`
}

type passiveArtifactUsageState struct {
	SchemaVersion string                             `json:"schema_version"`
	ContextType   string                             `json:"context_type"`
	Items         map[string]passiveArtifactUsageRow `json:"items"`
}

type passiveArtifactUsageRow struct {
	OpenCount        int    `json:"open_count,omitempty"`
	ExportOpenCount  int    `json:"export_open_count,omitempty"`
	ReopenRecorded   bool   `json:"reopen_recorded,omitempty"`
	ExportRecorded   bool   `json:"export_recorded,omitempty"`
	ReplacedRecorded bool   `json:"replaced_recorded,omitempty"`
	LastOpenedAt     string `json:"last_opened_at,omitempty"`
}

func workRoutePath(workDir string) string {
	return filepath.Join(workDir, "route.json")
}

func saveWorkRoute(workDir string, request providerGenerationRequest, decision routeDecision) error {
	if workDir == "" || !decision.Active {
		return nil
	}
	existing := loadWorkRoute(workDir)
	previousToolMode := strings.TrimSpace(existing.PreviousToolMode)
	routeSwitchCount := existing.RouteSwitchCount
	lastRouteSwitchReason := strings.TrimSpace(existing.LastRouteSwitchReason)
	if existing.ToolMode != "" && existing.ToolMode != decision.ToolMode {
		previousToolMode = existing.ToolMode
		routeSwitchCount++
		lastRouteSwitchReason = strings.TrimSpace(decision.Reason)
	}
	cliHandoffReceipt := decision.CLIHandoffReceipt
	if cliHandoffReceipt == nil && existing.ToolMode == decision.ToolMode {
		cliHandoffReceipt = existing.CLIHandoffReceipt
	}
	data, err := json.MarshalIndent(savedWorkRoute{
		SchemaVersion:                "0.1.0",
		ContextType:                  "JiniWorkRoute",
		ToolMode:                     decision.ToolMode,
		ToolLabel:                    decision.ToolLabel,
		RoutePolicy:                  decision.RoutePolicy,
		AutoMode:                     autoModePolicyPointer(decision.AutoMode),
		FeedbackKey:                  firstNonEmpty(strings.TrimSpace(decision.FeedbackKey), routeFeedbackKeyForDecision(decision)),
		ModelLabel:                   decision.ModelLabel,
		ModelReason:                  decision.ModelReason,
		ProviderLabel:                decision.Provider.Label,
		ChosenAutomatically:          decision.ChosenAutomatically,
		Reason:                       decision.Reason,
		ContinuityReason:             decision.ContinuityReason,
		EffortLevel:                  decision.EffortLevel,
		EffortReason:                 decision.EffortReason,
		VerificationLevel:            decision.VerificationLevel,
		VerificationReason:           decision.VerificationReason,
		ModelFeedback:                existing.ModelFeedback,
		ArtifactFeedbackPath:         existing.ArtifactFeedbackPath,
		ArtifactFeedbackReason:       existing.ArtifactFeedbackReason,
		ArtifactEditClass:            existing.ArtifactEditClass,
		ArtifactEditScope:            existing.ArtifactEditScope,
		ArtifactSemanticClass:        existing.ArtifactSemanticClass,
		ArtifactOutcomeSignal:        existing.ArtifactOutcomeSignal,
		ArtifactOutcomeReason:        existing.ArtifactOutcomeReason,
		PassiveArtifactOutcomeSignal: existing.PassiveArtifactOutcomeSignal,
		PassiveArtifactOutcomeReason: existing.PassiveArtifactOutcomeReason,
		PreviousToolMode:             previousToolMode,
		RouteSwitchCount:             routeSwitchCount,
		LastRouteSwitchReason:        lastRouteSwitchReason,
		CLIHandoffReceipt:            cliHandoffReceipt,
	}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(workRoutePath(workDir), append(data, '\n'), 0o600); err != nil {
		return err
	}
	if existing.ToolMode == "" {
		if err := recordManualRouteOverride(request, decision); err != nil {
			return err
		}
	}
	return nil
}

func artifactFeedbackBaselinePath(workDir string) string {
	return filepath.Join(workDir, "feedback-baseline.json")
}

func passiveArtifactUsagePath(workDir string) string {
	return filepath.Join(workDir, "artifact-usage.json")
}

func saveArtifactFeedbackBaseline(workDir string) error {
	viewDir := filepath.Join(workDir, "views")
	entries, err := os.ReadDir(viewDir)
	if err != nil {
		return nil
	}
	payload := artifactFeedbackBaseline{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniArtifactFeedbackBaseline",
		Views:         map[string]string{},
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(viewDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(workDir, path)
		if err != nil {
			continue
		}
		payload.Views[filepath.ToSlash(rel)] = string(content)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(artifactFeedbackBaselinePath(workDir), append(data, '\n'), 0o600)
}

func loadArtifactFeedbackBaseline(workDir string) artifactFeedbackBaseline {
	data, err := os.ReadFile(artifactFeedbackBaselinePath(workDir))
	if err != nil {
		return artifactFeedbackBaseline{Views: map[string]string{}}
	}
	var payload artifactFeedbackBaseline
	if err := json.Unmarshal(data, &payload); err != nil {
		return artifactFeedbackBaseline{Views: map[string]string{}}
	}
	if payload.Views == nil {
		payload.Views = map[string]string{}
	}
	return payload
}

func loadPassiveArtifactUsageState(workDir string) passiveArtifactUsageState {
	data, err := os.ReadFile(passiveArtifactUsagePath(workDir))
	if err != nil {
		return passiveArtifactUsageState{Items: map[string]passiveArtifactUsageRow{}}
	}
	var payload passiveArtifactUsageState
	if err := json.Unmarshal(data, &payload); err != nil {
		return passiveArtifactUsageState{Items: map[string]passiveArtifactUsageRow{}}
	}
	if payload.Items == nil {
		payload.Items = map[string]passiveArtifactUsageRow{}
	}
	return payload
}

func savePassiveArtifactUsageState(workDir string, state passiveArtifactUsageState) error {
	if state.SchemaVersion == "" {
		state.SchemaVersion = "0.1.0"
	}
	if state.ContextType == "" {
		state.ContextType = "JiniPassiveArtifactUsage"
	}
	if state.Items == nil {
		state.Items = map[string]passiveArtifactUsageRow{}
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(passiveArtifactUsagePath(workDir), append(data, '\n'), 0o600)
}

func inferArtifactFeedbackSignal(workDir, artifactPath, feedback string) (string, string, string, string) {
	if strings.TrimSpace(workDir) == "" || strings.TrimSpace(artifactPath) == "" {
		return "", "", "", ""
	}
	baseline := loadArtifactFeedbackBaseline(workDir)
	if len(baseline.Views) == 0 {
		return "", "", "", ""
	}
	rel, err := filepath.Rel(workDir, artifactPath)
	if err != nil {
		return "", "", "", ""
	}
	key := filepath.ToSlash(rel)
	original, ok := baseline.Views[key]
	if !ok {
		return "", "", "", ""
	}
	current, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", "", "", ""
	}
	editClass, editScope, semanticClass := classifyArtifactEditSignal(original, string(current))
	if editClass == "" {
		return "", "", "", ""
	}
	switch feedback {
	case "accepted-as-is":
		switch editClass {
		case "none":
			return editClass, editScope, semanticClass, "accepted-without-edits"
		case "light":
			if editScope == "header-only" {
				return editClass, editScope, semanticClass, "accepted-after-cosmetic-edits"
			}
			if semanticClass == "core-wording" {
				return editClass, editScope, semanticClass, "accepted-after-core-wording-edits"
			}
			return editClass, editScope, semanticClass, "accepted-after-light-edits"
		default:
			if semanticClass == "core-decision-change" {
				return editClass, editScope, semanticClass, "accepted-after-decision-change"
			}
			if editScope == "core-sections" {
				return editClass, editScope, semanticClass, "accepted-after-core-rewrite"
			}
			return editClass, editScope, semanticClass, "accepted-after-heavy-edits"
		}
	case "needed-light-edits":
		switch editClass {
		case "none":
			return editClass, editScope, semanticClass, "light-edits-not-detected"
		case "light":
			if semanticClass == "core-wording" {
				return editClass, editScope, semanticClass, "needed-core-wording-edits"
			}
			if editScope == "core-sections" {
				return editClass, editScope, semanticClass, "needed-core-edits"
			}
			return editClass, editScope, semanticClass, "light-edits-confirmed"
		default:
			if semanticClass == "core-decision-change" {
				return editClass, editScope, semanticClass, "needed-decision-rewrite"
			}
			if editScope == "core-sections" {
				return editClass, editScope, semanticClass, "needed-core-rewrite"
			}
			return editClass, editScope, semanticClass, "needed-more-than-light-edits"
		}
	case "not-useful":
		switch editClass {
		case "none":
			return editClass, editScope, semanticClass, "not-useful-without-editing"
		case "light":
			if semanticClass == "core-wording" {
				return editClass, editScope, semanticClass, "core-wording-still-not-useful"
			}
			if editScope == "core-sections" {
				return editClass, editScope, semanticClass, "core-sections-still-not-useful"
			}
			return editClass, editScope, semanticClass, "not-useful-after-light-edits"
		default:
			if editScope == "header-only" {
				return editClass, editScope, semanticClass, "not-useful-even-after-cosmetic-edits"
			}
			if semanticClass == "core-decision-change" {
				return editClass, editScope, semanticClass, "decision-changed-still-not-useful"
			}
			return editClass, editScope, semanticClass, "substantive-rewrite-needed"
		}
	default:
		return editClass, editScope, semanticClass, ""
	}
}

func classifyArtifactEditSignal(original, current string) (string, string, string) {
	a := artifactEditTokens(original)
	b := artifactEditTokens(current)
	if len(a) == 0 && len(b) == 0 {
		return "none", "none", "none"
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return "none", "none", "none"
	}
	distance := tokenLevenshteinDistance(a, b)
	ratio := float64(distance) / float64(maxLen)
	scope := classifyArtifactEditScope(original, current)
	semantic := classifyArtifactSemanticClass(original, current, scope)
	switch {
	case ratio <= 0.03:
		return "none", scope, semantic
	case ratio <= artifactLightEditThreshold(scope):
		return "light", scope, semantic
	default:
		return "heavy", scope, semantic
	}
}

func artifactLightEditThreshold(scope string) float64 {
	switch scope {
	case "header-only":
		return 0.25
	case "supporting-sections":
		return 0.18
	case "core-sections":
		return 0.12
	default:
		return 0.18
	}
}

type artifactSection struct {
	Heading string
	Body    string
}

func classifyArtifactEditScope(original, current string) string {
	originalSections := artifactSections(original)
	currentSections := artifactSections(current)
	if len(originalSections) == 0 && len(currentSections) == 0 {
		return "none"
	}
	changedHeader := false
	changedSupporting := false
	changedCore := false
	originalByHeading := map[string]string{}
	currentByHeading := map[string]string{}
	for _, section := range originalSections {
		originalByHeading[section.Heading] = section.Body
	}
	for _, section := range currentSections {
		currentByHeading[section.Heading] = section.Body
	}
	seen := map[string]bool{}
	for _, section := range originalSections {
		seen[section.Heading] = true
		if currentByHeading[section.Heading] == section.Body {
			continue
		}
		switch artifactSectionRole(section.Heading) {
		case "header":
			changedHeader = true
		case "core":
			changedCore = true
		default:
			changedSupporting = true
		}
	}
	for _, section := range currentSections {
		if seen[section.Heading] {
			continue
		}
		switch artifactSectionRole(section.Heading) {
		case "header":
			changedHeader = true
		case "core":
			changedCore = true
		default:
			changedSupporting = true
		}
		if originalByHeading[section.Heading] != section.Body {
			continue
		}
	}
	switch {
	case changedCore:
		return "core-sections"
	case changedSupporting:
		return "supporting-sections"
	case changedHeader:
		return "header-only"
	default:
		return "none"
	}
}

func artifactSections(content string) []artifactSection {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	sections := []artifactSection{}
	currentHeading := "Document"
	currentBody := []string{}
	flush := func() {
		sections = append(sections, artifactSection{
			Heading: currentHeading,
			Body:    strings.TrimSpace(strings.Join(currentBody, "\n")),
		})
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if len(sections) > 0 || len(currentBody) > 0 || currentHeading != "Document" {
				flush()
			}
			currentHeading = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			currentBody = nil
			continue
		}
		currentBody = append(currentBody, line)
	}
	flush()
	out := []artifactSection{}
	for _, section := range sections {
		if strings.TrimSpace(section.Heading) == "" && strings.TrimSpace(section.Body) == "" {
			continue
		}
		out = append(out, section)
	}
	return out
}

func artifactSectionRole(heading string) string {
	normalized := normalizeName(heading)
	switch {
	case normalized == "document" || containsAny(normalized, []string{"sendable follow up", "build readiness check", "itinerary", "trip at a glance"}):
		return "header"
	case containsAny(normalized, []string{
		"send this note",
		"decisions captured",
		"what looks ready now",
		"must clear before build",
		"extracted evidence",
		"what the source shows",
		"what the document shows",
		"what is visible",
		"what the recording says",
		"recommended first slice",
		"day by day draft",
		"budget sketch",
		"top option",
		"recovery proof",
		"first useful pass",
	}):
		return "core"
	default:
		return "supporting"
	}
}

func classifyArtifactSemanticClass(original, current, scope string) string {
	if scope != "core-sections" {
		return scope
	}
	originalCore := coreSectionBody(original)
	currentCore := coreSectionBody(current)
	if strings.TrimSpace(originalCore) == "" || strings.TrimSpace(currentCore) == "" {
		return "core-decision-change"
	}
	originalTokens := artifactEditTokens(originalCore)
	currentTokens := artifactEditTokens(currentCore)
	if len(originalTokens) == 0 || len(currentTokens) == 0 {
		return "core-decision-change"
	}
	overlap := tokenOverlapRatio(originalTokens, currentTokens)
	originalBullets := bulletCount(originalCore)
	currentBullets := bulletCount(currentCore)
	bulletDelta := absInt(originalBullets - currentBullets)
	if overlap >= 0.45 && bulletDelta <= 1 {
		return "core-wording"
	}
	if overlap >= 0.35 && bulletDelta == 0 {
		return "core-wording"
	}
	return "core-decision-change"
}

func coreSectionBody(content string) string {
	sections := artifactSections(content)
	parts := []string{}
	for _, section := range sections {
		if artifactSectionRole(section.Heading) == "core" {
			parts = append(parts, section.Body)
		}
	}
	return strings.Join(parts, "\n")
}

func tokenOverlapRatio(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	counts := map[string]int{}
	for _, token := range a {
		counts[token]++
	}
	shared := 0
	for _, token := range b {
		if counts[token] > 0 {
			shared++
			counts[token]--
		}
	}
	denom := len(a)
	if len(b) > denom {
		denom = len(b)
	}
	if denom == 0 {
		return 0
	}
	return float64(shared) / float64(denom)
}

func bulletCount(content string) int {
	count := 0
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			count++
		}
	}
	return count
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func artifactEditTokens(content string) []string {
	content = strings.ToLower(strings.TrimSpace(content))
	if content == "" {
		return nil
	}
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"\t", " ",
		",", " ",
		".", " ",
		":", " ",
		";", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"-", " ",
		"_", " ",
		"/", " ",
		"\\", " ",
		"#", " ",
		"*", " ",
		"`", " ",
	)
	return strings.Fields(replacer.Replace(content))
}

func tokenLevenshteinDistance(a, b []string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			insert := curr[j-1] + 1
			delete := prev[j] + 1
			replace := prev[j-1] + cost
			curr[j] = insert
			if delete < curr[j] {
				curr[j] = delete
			}
			if replace < curr[j] {
				curr[j] = replace
			}
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func loadWorkRoute(workDir string) savedWorkRoute {
	data, err := os.ReadFile(workRoutePath(workDir))
	if err != nil {
		return savedWorkRoute{}
	}
	var payload savedWorkRoute
	if err := json.Unmarshal(data, &payload); err != nil {
		return savedWorkRoute{}
	}
	return payload
}

func providerRequestForWorkDir(workDir string) providerGenerationRequest {
	packID := inferPackID(workDir)
	inputs := loadInputItems(workDir, packID)
	return providerRequestForInputs(packID, inferUsing(packID), inputs)
}

func providerRequestForInputs(packID, title string, inputs []inputItem) providerGenerationRequest {
	request := providerGenerationRequest{
		Choice: starterChoice{PackID: packID},
		Title:  title,
	}
	if len(inputs) == 0 {
		return request
	}
	parts := make([]string, 0, len(inputs))
	for _, item := range inputs {
		switch item.Kind {
		case "text", "clarification", "derived":
			text := strings.TrimSpace(item.OriginRef)
			if text == "" {
				text = strings.TrimSpace(item.Preview)
			}
			if text != "" {
				parts = append(parts, text)
			}
		case "image":
			label := firstNonEmpty(strings.TrimSpace(item.Title), strings.TrimSpace(item.Preview), "image")
			parts = append(parts, "Image attachment: "+label)
		case "audio":
			label := firstNonEmpty(strings.TrimSpace(item.Title), strings.TrimSpace(item.Preview), "audio")
			parts = append(parts, "Audio attachment: "+label)
		case "file":
			label := firstNonEmpty(strings.TrimSpace(item.Title), strings.TrimSpace(item.Preview), "file")
			parts = append(parts, "File attachment: "+label)
		default:
			text := firstNonEmpty(strings.TrimSpace(item.OriginRef), strings.TrimSpace(item.Preview), strings.TrimSpace(item.Title))
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	request.Source = strings.Join(parts, "\n")
	return request
}

func saveModelFeedback(workDir, feedback, artifactPath string) error {
	route := loadWorkRoute(workDir)
	request := providerRequestForWorkDir(workDir)
	if strings.TrimSpace(route.ToolMode) == "" && configuredProviderMode() == "local-slm" {
		route.ToolMode = resolveLocalSLMToolModeForRequest(request)
		route.ProviderLabel = "Local SLM"
	}
	if strings.TrimSpace(route.ToolMode) == "" {
		return nil
	}
	route.ModelFeedback = strings.TrimSpace(feedback)
	editClass, editScope, semanticClass, reason := inferArtifactFeedbackSignal(workDir, artifactPath, route.ModelFeedback)
	if strings.TrimSpace(artifactPath) != "" {
		route.ArtifactFeedbackPath = artifactPath
	}
	route.ArtifactEditClass = editClass
	route.ArtifactEditScope = editScope
	route.ArtifactSemanticClass = semanticClass
	route.ArtifactFeedbackReason = reason
	data, err := json.MarshalIndent(route, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(workRoutePath(workDir), append(data, '\n'), 0o600); err != nil {
		return err
	}
	feedbackKey := strings.TrimSpace(route.FeedbackKey)
	if feedbackKey == "" {
		feedbackKey = routeFeedbackKeyForSavedRoute(route)
	}
	if err := recordRouteFeedbackForKey(feedbackKey, route.ModelFeedback); err != nil {
		return err
	}
	if err := recordRouteCohortFeedback(feedbackKey, request, route.ModelFeedback, route.ArtifactEditClass, route.ArtifactEditScope, route.ArtifactSemanticClass, route.ArtifactFeedbackReason); err != nil {
		return err
	}
	return recordLocalCohortFeedback(route.ToolMode, request, route.ModelFeedback, route.ArtifactEditClass, route.ArtifactEditScope, route.ArtifactSemanticClass, route.ArtifactFeedbackReason)
}

func inferArtifactOutcomeReason(workDir, artifactPath, outcome string) string {
	_ = workDir
	_ = artifactPath
	switch strings.TrimSpace(outcome) {
	case "used-this":
		return "used-in-real-work"
	case "shared-this":
		return "shared-or-handed-off"
	case "replaced-this":
		return "replaced-after-review"
	default:
		return ""
	}
}

func saveArtifactOutcome(workDir, outcome, artifactPath string) error {
	route := loadWorkRoute(workDir)
	request := providerRequestForWorkDir(workDir)
	if strings.TrimSpace(route.ToolMode) == "" && configuredProviderMode() == "local-slm" {
		route.ToolMode = resolveLocalSLMToolModeForRequest(request)
		route.ProviderLabel = "Local SLM"
	}
	if strings.TrimSpace(route.ToolMode) == "" {
		return nil
	}
	route.ArtifactOutcomeSignal = strings.TrimSpace(outcome)
	route.ArtifactOutcomeReason = inferArtifactOutcomeReason(workDir, artifactPath, route.ArtifactOutcomeSignal)
	if strings.TrimSpace(artifactPath) != "" {
		route.ArtifactFeedbackPath = artifactPath
	}
	data, err := json.MarshalIndent(route, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(workRoutePath(workDir), append(data, '\n'), 0o600); err != nil {
		return err
	}
	feedbackKey := strings.TrimSpace(route.FeedbackKey)
	if feedbackKey == "" {
		feedbackKey = routeFeedbackKeyForSavedRoute(route)
	}
	if err := recordRouteCohortOutcome(feedbackKey, request, route.ArtifactOutcomeSignal); err != nil {
		return err
	}
	return recordLocalCohortOutcome(route.ToolMode, request, route.ArtifactOutcomeSignal, route.ArtifactOutcomeReason)
}

func savePassiveArtifactOutcome(workDir, outcome, artifactPath, reason string) error {
	route := loadWorkRoute(workDir)
	request := providerRequestForWorkDir(workDir)
	if strings.TrimSpace(route.ToolMode) == "" && configuredProviderMode() == "local-slm" {
		route.ToolMode = resolveLocalSLMToolModeForRequest(request)
		route.ProviderLabel = "Local SLM"
	}
	if strings.TrimSpace(route.ToolMode) == "" {
		return nil
	}
	route.PassiveArtifactOutcomeSignal = strings.TrimSpace(outcome)
	route.PassiveArtifactOutcomeReason = strings.TrimSpace(reason)
	if strings.TrimSpace(artifactPath) != "" {
		route.ArtifactFeedbackPath = artifactPath
	}
	data, err := json.MarshalIndent(route, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(workRoutePath(workDir), append(data, '\n'), 0o600); err != nil {
		return err
	}
	feedbackKey := strings.TrimSpace(route.FeedbackKey)
	if feedbackKey == "" {
		feedbackKey = routeFeedbackKeyForSavedRoute(route)
	}
	if err := recordPassiveRouteCohortOutcome(feedbackKey, request, route.PassiveArtifactOutcomeSignal); err != nil {
		return err
	}
	return recordPassiveLocalCohortOutcome(route.ToolMode, request, route.PassiveArtifactOutcomeSignal, route.PassiveArtifactOutcomeReason)
}

func inferPassiveArtifactReplacementSignal(workDir, artifactPath string) (string, string) {
	if strings.TrimSpace(workDir) == "" || strings.TrimSpace(artifactPath) == "" {
		return "", ""
	}
	baseline := loadArtifactFeedbackBaseline(workDir)
	if len(baseline.Views) == 0 {
		return "", ""
	}
	rel, err := filepath.Rel(workDir, artifactPath)
	if err != nil {
		return "", ""
	}
	original, ok := baseline.Views[filepath.ToSlash(rel)]
	if !ok {
		return "", ""
	}
	current, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", ""
	}
	editClass, editScope, semanticClass := classifyArtifactEditSignal(original, string(current))
	if semanticClass == "core-decision-change" && editClass != "none" {
		return "replaced-this", "reopened-after-decision-change"
	}
	if editClass != "heavy" {
		return "", ""
	}
	if editScope == "core-sections" {
		return "replaced-this", "reopened-after-substantive-rewrite"
	}
	return "", ""
}

func recordPassiveArtifactObservation(workDir string, item catalogItem) error {
	if strings.TrimSpace(workDir) == "" || strings.TrimSpace(item.Path) == "" {
		return nil
	}
	rel, err := filepath.Rel(workDir, item.Path)
	if err != nil {
		return nil
	}
	key := filepath.ToSlash(rel)
	state := loadPassiveArtifactUsageState(workDir)
	row := state.Items[key]
	row.LastOpenedAt = time.Now().UTC().Format(time.RFC3339)
	connectorID := inferConnectorForCatalogItem(item)
	if strings.Contains(key, "/exports/") || strings.HasPrefix(key, "exports/") {
		row.ExportOpenCount++
		if !row.ExportRecorded {
			if err := savePassiveArtifactOutcome(workDir, "shared-this", item.Path, connectorAwareOutcomeReason(connectorID, "opened-export-artifact")); err != nil {
				return err
			}
			row.ExportRecorded = true
		}
		state.Items[key] = row
		return savePassiveArtifactUsageState(workDir, state)
	}
	row.OpenCount++
	if !row.ReplacedRecorded {
		if outcome, reason := inferPassiveArtifactReplacementSignal(workDir, item.Path); outcome != "" {
			if err := savePassiveArtifactOutcome(workDir, outcome, item.Path, reason); err != nil {
				return err
			}
			row.ReplacedRecorded = true
		}
	}
	if row.OpenCount >= 2 && !row.ReopenRecorded {
		if err := savePassiveArtifactOutcome(workDir, "used-this", item.Path, "reopened-repeatedly"); err != nil {
			return err
		}
		row.ReopenRecorded = true
	}
	state.Items[key] = row
	return savePassiveArtifactUsageState(workDir, state)
}

func workingWithLabelForSavedRoute(route savedWorkRoute) string {
	label := strings.TrimSpace(route.ToolLabel)
	if label == "" {
		return ""
	}
	providerLabel := strings.TrimSpace(route.ProviderLabel)
	if providerLabel != "" && providerLabel != label && route.ToolMode != "local-preview" {
		label += " via " + providerLabel
	}
	if route.ChosenAutomatically {
		label += " (chosen automatically)"
	}
	return label
}

func routeFeedbackKeyForDecision(decision routeDecision) string {
	if strings.TrimSpace(decision.ToolMode) == "" {
		return ""
	}
	profile := currentDeviceProfile()
	parts := []string{
		strings.TrimSpace(decision.ToolMode),
		strings.TrimSpace(firstNonEmpty(decision.ModelLabel, modelLabelForToolMode(decision.ToolMode))),
	}
	if strings.HasPrefix(decision.ToolMode, "local-") {
		parts = append(parts,
			strings.TrimSpace(profile.DeviceClass),
			strings.TrimSpace(profile.LocalRuntimeClass),
			strings.TrimSpace(profile.LocalEndpointSignature),
		)
	}
	return strings.Join(parts, "|")
}

func routeFeedbackKeyForSavedRoute(route savedWorkRoute) string {
	parts := []string{
		strings.TrimSpace(route.ToolMode),
		strings.TrimSpace(firstNonEmpty(route.ModelLabel, modelLabelForToolMode(route.ToolMode))),
	}
	return strings.Join(parts, "|")
}

func resolveLocalSLMDefaultModel() (string, string) {
	modelID := strings.TrimSpace(configFirstNonEmpty(
		"JINI_LOCAL_SLM_MODEL",
		"JINI_LOCAL_SLM_WORKHORSE_MODEL",
		"JINI_LOCAL_SLM_FAST_MODEL",
		"JINI_LOCAL_SLM_DEEP_MODEL",
		"JINI_LOCAL_SLM_MULTIMODAL_MODEL",
	))
	if modelID == "" {
		return autoLocalSLMModelForToolMode("local-workhorse")
	}
	return modelID, modelID
}

func resolveLocalSLMModelForToolMode(toolMode string) (string, string) {
	switch toolMode {
	case "local-fast":
		modelID := strings.TrimSpace(configFirstNonEmpty("JINI_LOCAL_SLM_FAST_MODEL", "JINI_LOCAL_SLM_MODEL"))
		if modelID == "" {
			return autoLocalSLMModelForToolMode(toolMode)
		}
		return modelID, firstNonEmpty(modelID, "Local SLM fast")
	case "local-workhorse":
		modelID := strings.TrimSpace(configFirstNonEmpty("JINI_LOCAL_SLM_WORKHORSE_MODEL", "JINI_LOCAL_SLM_MODEL", "JINI_LOCAL_SLM_FAST_MODEL"))
		if modelID == "" {
			return autoLocalSLMModelForToolMode(toolMode)
		}
		return modelID, firstNonEmpty(modelID, "Local SLM workhorse")
	case "local-deep":
		modelID := strings.TrimSpace(configFirstNonEmpty("JINI_LOCAL_SLM_DEEP_MODEL", "JINI_LOCAL_SLM_WORKHORSE_MODEL", "JINI_LOCAL_SLM_MODEL"))
		if modelID == "" {
			return autoLocalSLMModelForToolMode(toolMode)
		}
		return modelID, firstNonEmpty(modelID, "Local SLM deep")
	case "local-multimodal":
		modelID := strings.TrimSpace(configFirstNonEmpty("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "JINI_LOCAL_SLM_MODEL"))
		if modelID == "" {
			return autoLocalSLMModelForToolMode(toolMode)
		}
		return modelID, firstNonEmpty(modelID, "Local SLM multimodal")
	default:
		return resolveLocalSLMDefaultModel()
	}
}

func resolveLocalSLMModelForRequest(request providerGenerationRequest) (string, string) {
	if toolMode := resolveLocalSLMToolModeForRequest(request); strings.HasPrefix(toolMode, "local-") {
		return resolveLocalSLMModelForToolMode(toolMode)
	}
	return resolveLocalSLMDefaultModel()
}

func resolveLocalSLMToolModeForRequest(request providerGenerationRequest) string {
	if route := detectRouteForRequest(request); route.Active && strings.HasPrefix(route.ToolMode, "local-") {
		return route.ToolMode
	}

	features := classifyRouteFeatures(request)
	switch {
	case features.ModalitySubtype == "pdf-scan", features.ModalitySubtype == "image-screenshot":
		return "local-multimodal"
	case features.ModalitySubtype == "audio-transcript":
		if features.DepthClass == "deep" || features.EffortClass == "high" || features.EffortClass == "extra high" {
			return "local-deep"
		}
		return "local-workhorse"
	case features.DepthClass == "deep" || features.EffortClass == "high" || features.EffortClass == "extra high":
		return "local-deep"
	case features.WorkClass == "planning" || features.WorkClass == "code":
		return "local-workhorse"
	case features.EffortClass == "low":
		return "local-fast"
	default:
		return "local-fast"
	}
}
