package app

import "strings"

// Ready throttle-fallback routes — the ranked list of routes a throttle can
// switch to. Unlike detectRuntimeAvailability.OfflineRouteMode (which is
// offline-oriented: a local model or local-preview), this surfaces configured,
// ready ONLINE BYO providers (Groq/DeepSeek/…) that differ from the throttled
// route, then the local model as the never-throttles floor. It powers the free
// nudge's manual-switch guidance AND the paid Autopilot's fallback ladder.
//
// Preference order: a different-provider online BYO route (fast, keeps quality)
// ranks above the local floor. The throttled route is always excluded — never
// suggest switching to the route that is currently limited.

// byoFallbackPreference ranks BYO routes for the ladder (stable, deterministic).
// Cerebras first — fastest free inference (~2,600 tok/s) and a large daily
// ceiling — then Groq, then the rest.
var byoFallbackPreference = []string{"cerebras", "groq", "deepseek", "xai", "mistral", "openai"}

// readyThrottleFallbackModes returns ranked route modes that are ready to answer
// and differ from throttledMode. Cheap enough to call lazily on a throttle only.
func readyThrottleFallbackModes(request providerGenerationRequest, throttledMode string) []string {
	throttled := strings.TrimSpace(throttledMode)
	var modes []string

	for _, id := range byoFallbackPreference {
		if id == throttled {
			continue
		}
		if detectProviderForMode(id).Status == "ok" {
			modes = append(modes, id)
		}
	}

	// Local model as the floor — it physically cannot rate-limit (your hardware).
	availability := detectRuntimeAvailability(request)
	local := strings.TrimSpace(availability.OfflineRouteMode)
	if local != "" && local != "local-preview" && local != throttled {
		modes = append(modes, local)
	}
	return modes
}
