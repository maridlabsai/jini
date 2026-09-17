package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type savedProviderSettings struct {
	SchemaVersion string            `json:"schema_version"`
	ContextType   string            `json:"context_type"`
	Values        map[string]string `json:"values"`
}

func providerSettingsPath() string {
	return filepath.Join(sessionStateRoot(), "provider.json")
}

func loadSavedProviderSettings() map[string]string {
	data, err := os.ReadFile(providerSettingsPath())
	if err != nil {
		return map[string]string{}
	}
	var payload savedProviderSettings
	if err := json.Unmarshal(data, &payload); err != nil {
		return map[string]string{}
	}
	if payload.Values == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(payload.Values))
	for key, value := range payload.Values {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	return out
}

func saveProviderSettings(values map[string]string) error {
	current := loadSavedProviderSettings()
	for key, value := range values {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		// Secrets go to the OS keychain when one is available; they are never
		// written to (and are scrubbed from) the plaintext dotfile.
		if isSecretConfigKey(key) && providerSecretStore.available() {
			if value == "" {
				_ = providerSecretStore.delete(key)
			} else if err := providerSecretStore.set(key, value); err != nil {
				return err
			}
			delete(current, key)
			continue
		}
		if value == "" {
			delete(current, key)
			continue
		}
		current[key] = value
	}
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(savedProviderSettings{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniProviderSettings",
		Values:        current,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(providerSettingsPath(), append(data, '\n'), 0o600)
}

func configValue(name string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	// Prefer the OS keychain for secrets, then fall back to the dotfile for
	// back-compat. A legacy plaintext secret keeps working and migrates into the
	// keychain (out of plaintext) the next time it is saved — no write side
	// effects on a read.
	if isSecretConfigKey(name) && providerSecretStore.available() {
		if value, ok := providerSecretStore.get(name); ok {
			return strings.TrimSpace(value)
		}
	}
	return strings.TrimSpace(loadSavedProviderSettings()[name])
}

func configFirstNonEmpty(names ...string) string {
	for _, name := range names {
		if value := configValue(name); value != "" {
			return value
		}
	}
	return ""
}
