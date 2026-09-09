package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Data-driven provider catalog — onboard a NEW OpenAI-compatible provider (or fix
// a stale field on a built-in one, like a deprecated default model) WITHOUT a
// Jini release. The built-in byoShapes are the floor; a catalog file overlays
// them. A catalog entry with a new id adds a provider; one matching a built-in
// id overrides only the fields it sets (so `{"id":"groq","default_model":"…"}`
// fixes a stale default without redefining the whole provider).
//
// The catalog lives at <JINI_HOME>/.jini/providers.json (same home as mode/trust/
// ledger, so JINI_HOME sandboxes it). A signed REMOTE catalog Jini fetches
// periodically is the next layer; this local file is the mechanism it will reuse.

type catalogProviderEntry struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	KeyEnv       string `json:"key_env"`
	BaseEnv      string `json:"base_env"`
	DefaultBase  string `json:"default_base"`
	ModelsPath   string `json:"models_path"` // default "/models"
	ChatPath     string `json:"chat_path"`
	ModelEnv     string `json:"model_env"`
	DefaultModel string `json:"default_model"`
	PrivacyNote  string `json:"privacy_note"`
}

type providerCatalogFile struct {
	SchemaVersion string                 `json:"schema_version"`
	Providers     []catalogProviderEntry `json:"providers"`
}

// providerCatalogLoader is overridable in tests so the suite never reads a real
// on-disk catalog (which would make the provider set non-hermetic).
var providerCatalogLoader = loadProviderCatalogFromDisk

func providerCatalogPath() string {
	home, err := executionModeHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".jini", "providers.json")
}

func loadProviderCatalogFromDisk() []catalogProviderEntry {
	path := providerCatalogPath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // no catalog is the normal case
	}
	var file providerCatalogFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil // a malformed catalog must never break the built-in providers
	}
	return file.Providers
}

// overlay applies the entry's non-empty fields onto a shape (built-in or new).
func (e catalogProviderEntry) overlay(shape byoShape) byoShape {
	if shape.ID == "" {
		shape.ID = e.ID
	}
	shape.Label = firstNonEmpty(e.Label, shape.Label)
	shape.KeyEnv = firstNonEmpty(e.KeyEnv, shape.KeyEnv)
	shape.baseEnv = firstNonEmpty(e.BaseEnv, shape.baseEnv)
	shape.defaultBase = firstNonEmpty(e.DefaultBase, shape.defaultBase)
	shape.chatPath = firstNonEmpty(e.ChatPath, shape.chatPath)
	shape.modelEnv = firstNonEmpty(e.ModelEnv, shape.modelEnv)
	shape.defaultModel = firstNonEmpty(e.DefaultModel, shape.defaultModel)
	shape.privacyNote = firstNonEmpty(e.PrivacyNote, shape.privacyNote)
	if e.ModelsPath != "" {
		shape.buildRequest = bearerModelsRequest(e.ModelsPath)
	} else if shape.buildRequest == nil {
		shape.buildRequest = bearerModelsRequest("/models")
	}
	return shape
}

// byoShapeRegistry returns the effective provider set: built-in shapes overlaid
// with any catalog entries. Rebuilt per call (cheap; not a hot path) so a
// dropped-in catalog takes effect without a restart.
func byoShapeRegistry() map[string]byoShape {
	reg := make(map[string]byoShape, len(byoShapes)+4)
	for id, shape := range byoShapes {
		reg[id] = shape
	}
	for _, entry := range providerCatalogLoader() {
		id := strings.TrimSpace(entry.ID)
		if id == "" {
			continue
		}
		merged := entry.overlay(reg[id])
		// A provider Jini can't authenticate is useless; require a key env.
		if strings.TrimSpace(merged.KeyEnv) == "" {
			continue
		}
		reg[id] = merged
	}
	return reg
}

// byoAliasRegistry merges the built-in aliases with an identity alias for every
// catalog-added id (so `resolveBYOShape("<new>")` works).
func byoAliasRegistry() map[string]string {
	al := make(map[string]string, len(byoShapeAliases)+4)
	for k, v := range byoShapeAliases {
		al[k] = v
	}
	for id := range byoShapeRegistry() {
		if _, ok := byoShapes[id]; !ok {
			al[normalizeName(id)] = id
		}
	}
	return al
}

// byoShapeByID looks up a shape in the effective registry.
func byoShapeByID(id string) (byoShape, bool) {
	shape, ok := byoShapeRegistry()[id]
	return shape, ok
}
