package app

import (
	"net"
	"strings"
	"time"
)

const runtimeConnectivityProbeTimeout = 120 * time.Millisecond

type runtimeConnectivityProbeResult struct {
	State  string
	Reason string
	Known  bool
}

var runtimeConnectivityProbe = probeDefaultRuntimeConnectivity

type runtimeAvailability struct {
	OnlineState       string
	OfflineState      string
	OfflineMode       bool
	OfflineRouteMode  string
	RemoteConfigured  bool
	LocalSLMReady     bool
	LocalPreviewReady bool
	ConnectivityState string
	Reason            string
}

func detectRuntimeAvailability(request providerGenerationRequest) runtimeAvailability {
	localReady := localSLMRuntimeReady()
	offlineRoute := "local-preview"
	if localReady {
		offlineRoute = bestLocalSLMRouteModeForRequest(request)
	}
	remoteConfigured := anyRemoteProviderReady()
	connectivityState, connectivityReason, connectivityForced := configuredConnectivityState()
	if !connectivityForced && remoteConfigured {
		if probed := runtimeConnectivityProbe(); probed.Known {
			connectivityState = probed.State
			connectivityReason = probed.Reason
		}
	}
	offlineMode := false
	reason := ""
	switch {
	case connectivityState == "offline":
		offlineMode = true
		reason = connectivityReason
	case !remoteConfigured:
		offlineMode = true
		connectivityState = "no_remote_config"
		reason = "No remote provider or CLI API is configured, so Jini is using offline-capable routes."
	}
	if connectivityState == "" {
		connectivityState = "online"
	}
	offlineState := "available"
	if offlineMode {
		offlineState = "active"
	}
	return runtimeAvailability{
		OnlineState:       connectivityState,
		OfflineState:      offlineState,
		OfflineMode:       offlineMode,
		OfflineRouteMode:  offlineRoute,
		RemoteConfigured:  remoteConfigured,
		LocalSLMReady:     localReady,
		LocalPreviewReady: true,
		ConnectivityState: connectivityState,
		Reason:            reason,
	}
}

func configuredConnectivityState() (string, string, bool) {
	switch normalizeName(configValue("JINI_CONNECTIVITY_OVERRIDE")) {
	case "offline", "internet offline", "network offline", "no internet", "disconnected":
		return "offline", "Internet connectivity is unavailable, so Jini is using offline-capable routes.", true
	case "online", "internet online", "network online", "connected":
		return "online", "Internet connectivity is available.", true
	default:
		return "", "", false
	}
}

func probeDefaultRuntimeConnectivity() runtimeConnectivityProbeResult {
	switch normalizeName(configValue("JINI_CONNECTIVITY_PROBE")) {
	case "0", "false", "off", "disabled", "no":
		return runtimeConnectivityProbeResult{State: "unknown", Known: false}
	}
	target := strings.TrimSpace(configValue("JINI_CONNECTIVITY_PROBE_TARGET"))
	if target == "" {
		target = "1.1.1.1:443"
	} else if !strings.Contains(target, ":") {
		target += ":443"
	}
	conn, err := net.DialTimeout("udp", target, runtimeConnectivityProbeTimeout)
	if err == nil {
		_ = conn.Close()
		return runtimeConnectivityProbeResult{
			State:  "online",
			Reason: "Internet route is available.",
			Known:  true,
		}
	}
	message := strings.ToLower(err.Error())
	if containsAny(message, []string{
		"network is unreachable",
		"network unreachable",
		"no route to host",
		"can't assign requested address",
	}) {
		return runtimeConnectivityProbeResult{
			State:  "offline",
			Reason: "No usable internet route was detected, so Jini is using offline-capable routes.",
			Known:  true,
		}
	}
	return runtimeConnectivityProbeResult{
		State:  "unknown",
		Reason: "Internet route could not be proven quickly.",
		Known:  false,
	}
}

func anyRemoteProviderReady() bool {
	for _, mode := range []string{"anthropic", "azure-openai", "bedrock"} {
		if detectProviderForMode(mode).Status == "ok" {
			return true
		}
	}
	return false
}

func bestLocalSLMRouteModeForRequest(request providerGenerationRequest) string {
	bestMode := ""
	bestScore := -100000
	features := classifyRouteFeatures(request)
	for _, slot := range localSLMProfileSlots() {
		score := scoreRouteMode(slot.ID, features)
		if score > bestScore {
			bestScore = score
			bestMode = slot.ID
		}
	}
	return firstNonEmpty(bestMode, "local-workhorse")
}

func runtimeAvailabilityRouteBias(mode string, availability runtimeAvailability) int {
	if !availability.OfflineMode {
		return 0
	}
	switch {
	case isRemoteRouteMode(mode):
		return -1000
	case mode == availability.OfflineRouteMode:
		return 400
	case strings.HasPrefix(mode, "local-"):
		return 250
	default:
		return 0
	}
}

func isRemoteRouteMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case "claude-api", "bedrock-sonnet", "chatgpt", "azure-code", "azure-openai":
		return true
	default:
		return false
	}
}

func appendRuntimeAvailabilityReason(reason string, availability runtimeAvailability) string {
	if !availability.OfflineMode || strings.TrimSpace(availability.Reason) == "" {
		return reason
	}
	if strings.TrimSpace(reason) == "" {
		return availability.Reason
	}
	if strings.Contains(strings.ToLower(reason), "offline") {
		return reason
	}
	return strings.TrimSpace(reason) + " " + availability.Reason
}
