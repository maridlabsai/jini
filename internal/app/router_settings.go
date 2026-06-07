package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type savedRouteTarget struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	ProviderMode string `json:"provider_mode"`
	ModelHint    string `json:"model_hint"`
	Enabled      bool   `json:"enabled"`
}

type savedRouterSettings struct {
	SchemaVersion string             `json:"schema_version"`
	ContextType   string             `json:"context_type"`
	ToolMode      string             `json:"tool_mode"`
	Targets       []savedRouteTarget `json:"targets"`
}

func routerSettingsPath() string {
	return filepath.Join(sessionStateRoot(), "router.json")
}

func defaultSavedRouteTargets() []savedRouteTarget {
	return []savedRouteTarget{
		{ID: "claude-code", Label: "Claude Code", ProviderMode: "anthropic", ModelHint: "sonnet", Enabled: true},
		{ID: "bedrock-sonnet", Label: "Bedrock Sonnet", ProviderMode: "bedrock", ModelHint: "sonnet-4.6", Enabled: true},
		{ID: "chatgpt", Label: "Azure writing route", ProviderMode: "azure-openai", ModelHint: "auto", Enabled: true},
		{ID: "codex", Label: "Azure code route", ProviderMode: "azure-openai", ModelHint: "auto", Enabled: true},
		{ID: "local-fast", Label: "Local SLM fast", ProviderMode: "local-slm", ModelHint: "fast", Enabled: true},
		{ID: "local-workhorse", Label: "Local SLM workhorse", ProviderMode: "local-slm", ModelHint: "workhorse", Enabled: true},
		{ID: "local-deep", Label: "Local SLM deep", ProviderMode: "local-slm", ModelHint: "deep", Enabled: true},
		{ID: "local-multimodal", Label: "Local SLM multimodal", ProviderMode: "local-slm", ModelHint: "multimodal", Enabled: true},
		{ID: "local-preview", Label: "Local preview", ProviderMode: "local-preview", ModelHint: "auto", Enabled: true},
	}
}

func loadSavedRouterSettings() savedRouterSettings {
	data, err := os.ReadFile(routerSettingsPath())
	if err != nil {
		return savedRouterSettings{}
	}
	var payload savedRouterSettings
	if err := json.Unmarshal(data, &payload); err != nil {
		return savedRouterSettings{}
	}
	payload.ToolMode = strings.TrimSpace(payload.ToolMode)
	if len(payload.Targets) == 0 {
		payload.Targets = defaultSavedRouteTargets()
	}
	return payload
}

func saveRouterSettings(toolMode string) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	payload := savedRouterSettings{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniRouterSettings",
		ToolMode:      strings.TrimSpace(toolMode),
		Targets:       defaultSavedRouteTargets(),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(routerSettingsPath(), append(data, '\n'), 0o600)
}

func clearRouterSettings() error {
	err := os.Remove(routerSettingsPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func configuredToolMode() string {
	raw := strings.TrimSpace(os.Getenv("JINI_TOOL"))
	if raw == "" {
		raw = loadSavedRouterSettings().ToolMode
	}
	if raw == "" {
		return ""
	}
	switch normalizeName(raw) {
	case "auto", "choose automatically":
		return "auto"
	case "claude", "claude code", "claudecode", "anthropic":
		return "claude-code"
	case "bedrock", "bedrock sonnet", "bedrocksonnet", "amazon bedrock":
		return "bedrock-sonnet"
	case "azure", "azure openai", "azure open ai", "azureopenai":
		return "azure-openai"
	case "azure writing route", "azure writing", "writing route":
		return "chatgpt"
	case "azure code route", "azure code", "code route":
		return "codex"
	case "chatgpt", "chat gpt":
		return "chatgpt"
	case "codex":
		return "codex"
	case "local slm", "localslm", "local model":
		return "local-slm"
	case "local fast", "localfast":
		return "local-fast"
	case "local workhorse", "localworkhorse":
		return "local-workhorse"
	case "local deep", "localdeep":
		return "local-deep"
	case "local multimodal", "localmultimodal":
		return "local-multimodal"
	case "local", "local preview", "localpreview":
		return "local-preview"
	default:
		return normalizeName(raw)
	}
}
