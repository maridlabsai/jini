package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// BYO credential validation — the PRD requires validating a bring-your-own
// provider key with ONE live call and a typed error taxonomy, and that "a shape
// without a passing fixture is not claimed" (number-one-platform-prd.md
// §Routing And Resource Policy). Each shape names the credential env var and
// builds one cheap, read-only preflight request (a models-list GET where the
// API offers one); the response status is classified through
// classifyProviderHTTPError so the result reuses the shared taxonomy.
//
// Keys still come from configValue (env or ~/.jini/provider.json at 0600). OS
// keychain storage is a tracked follow-up; validation does not depend on it.

// byoShape describes one bring-your-own provider credential shape. The same
// registry drives both validation (buildRequest) and routing (chatPath/modelEnv)
// so a shape is defined in exactly one place — one paste, it validates and it
// routes.
type byoShape struct {
	ID     string
	Label  string
	KeyEnv string // primary credential env var / config key
	// baseEnv optionally overrides the API base URL (gateways, self-host).
	baseEnv     string
	defaultBase string
	// buildRequest constructs the read-only preflight request for a given key.
	buildRequest func(ctx context.Context, base, key string) (*http.Request, error)

	// Routing (OpenAI-compatible chat shapes only). When chatPath is set, the
	// shape is routable through generateWithOpenAICompatible: a pasted key works
	// with no further config because modelEnv has a sensible defaultModel.
	chatPath     string // e.g. "/v1/chat/completions"
	modelEnv     string // e.g. "XAI_MODEL"
	defaultModel string // used when modelEnv is unset — keeps setup to just a key
}

// routable reports whether the shape can serve generations (not validate-only).
func (s byoShape) routable() bool { return strings.TrimSpace(s.chatPath) != "" }

// bearerModelsRequest builds `GET {base}{path}` with a Bearer token — the
// OpenAI-compatible preflight shared by OpenAI, xAI (Grok), Groq, DeepSeek, and
// Mistral.
func bearerModelsRequest(path string) func(context.Context, string, string) (*http.Request, error) {
	return func(ctx context.Context, base, key string) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		return req, nil
	}
}

// byoShapes is the release-gated compatibility matrix. Every entry here is
// covered by a fixture in byo_validation_test.go.
var byoShapes = map[string]byoShape{
	"anthropic": {
		ID: "anthropic", Label: "Anthropic (Claude)", KeyEnv: "ANTHROPIC_API_KEY",
		baseEnv: "ANTHROPIC_BASE_URL", defaultBase: "https://api.anthropic.com",
		buildRequest: func(ctx context.Context, base, key string) (*http.Request, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/v1/models", nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("x-api-key", key)
			req.Header.Set("anthropic-version", "2023-06-01")
			return req, nil
		},
	},
	"openai": {
		ID: "openai", Label: "OpenAI", KeyEnv: "OPENAI_API_KEY",
		baseEnv: "OPENAI_BASE_URL", defaultBase: "https://api.openai.com",
		buildRequest: bearerModelsRequest("/v1/models"),
		chatPath:     "/v1/chat/completions", modelEnv: "OPENAI_MODEL", defaultModel: "gpt-4o",
	},
	"xai": {
		ID: "xai", Label: "xAI (Grok)", KeyEnv: "XAI_API_KEY",
		baseEnv: "XAI_BASE_URL", defaultBase: "https://api.x.ai",
		buildRequest: bearerModelsRequest("/v1/models"),
		chatPath:     "/v1/chat/completions", modelEnv: "XAI_MODEL", defaultModel: "grok-2-latest",
	},
	"groq": {
		ID: "groq", Label: "Groq", KeyEnv: "GROQ_API_KEY",
		baseEnv: "GROQ_BASE_URL", defaultBase: "https://api.groq.com",
		buildRequest: bearerModelsRequest("/openai/v1/models"),
		chatPath:     "/openai/v1/chat/completions", modelEnv: "GROQ_MODEL", defaultModel: "llama-3.3-70b-versatile",
	},
	"cerebras": {
		ID: "cerebras", Label: "Cerebras", KeyEnv: "CEREBRAS_API_KEY",
		baseEnv: "CEREBRAS_BASE_URL", defaultBase: "https://api.cerebras.ai",
		buildRequest: bearerModelsRequest("/v1/models"),
		chatPath:     "/v1/chat/completions", modelEnv: "CEREBRAS_MODEL", defaultModel: "llama-3.3-70b",
	},
	"deepseek": {
		ID: "deepseek", Label: "DeepSeek", KeyEnv: "DEEPSEEK_API_KEY",
		baseEnv: "DEEPSEEK_BASE_URL", defaultBase: "https://api.deepseek.com",
		buildRequest: bearerModelsRequest("/models"),
		chatPath:     "/chat/completions", modelEnv: "DEEPSEEK_MODEL", defaultModel: "deepseek-chat",
	},
	"mistral": {
		ID: "mistral", Label: "Mistral", KeyEnv: "MISTRAL_API_KEY",
		baseEnv: "MISTRAL_BASE_URL", defaultBase: "https://api.mistral.ai",
		buildRequest: bearerModelsRequest("/v1/models"),
		chatPath:     "/v1/chat/completions", modelEnv: "MISTRAL_MODEL", defaultModel: "mistral-large-latest",
	},
}

// byoShapeAliases maps user-facing names to a canonical shape id.
var byoShapeAliases = map[string]string{
	"anthropic": "anthropic", "claude": "anthropic",
	"openai": "openai", "gpt": "openai",
	"xai": "xai", "grok": "xai", "x ai": "xai", // normalizeName turns "x.ai"/"x-ai" into "x ai"
	"groq":     "groq",
	"cerebras": "cerebras",
	"deepseek": "deepseek",
	"mistral":  "mistral",
}

func resolveBYOShape(name string) (byoShape, bool) {
	id, ok := byoShapeAliases[normalizeName(name)]
	if !ok {
		return byoShape{}, false
	}
	shape, ok := byoShapes[id]
	return shape, ok
}

func sortedBYOShapeIDs() []string {
	ids := make([]string, 0, len(byoShapes))
	for id := range byoShapes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

type byoValidationResult struct {
	Shape      string
	Label      string
	KeyEnv     string
	Configured bool
	OK         bool
	Kind       providerCredentialErrorKind
	Message    string
}

// validateBYOCredential makes one live, read-only call and classifies the
// result. It never sends the key when none is configured, and it drains/closes
// the response body without reading it (status is the only signal needed).
func validateBYOCredential(ctx context.Context, shape byoShape, client *http.Client) byoValidationResult {
	result := byoValidationResult{Shape: shape.ID, Label: shape.Label, KeyEnv: shape.KeyEnv}
	key := configValue(shape.KeyEnv)
	if strings.TrimSpace(key) == "" {
		result.Message = fmt.Sprintf("%s: no credential found. Set %s (env or `jini provider`).", shape.Label, shape.KeyEnv)
		return result
	}
	result.Configured = true
	base := firstNonEmpty(configValue(shape.baseEnv), shape.defaultBase)
	req, err := shape.buildRequest(ctx, base, key)
	if err != nil {
		result.Message = fmt.Sprintf("%s: could not build validation request: %v", shape.Label, err)
		return result
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Message = fmt.Sprintf("%s: could not reach the API (%v). Check network access.", shape.Label, err)
		return result
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<12))
	resp.Body.Close()

	if classifyErr := classifyProviderHTTPError(shape.Label, shape.KeyEnv, resp.StatusCode); classifyErr != nil {
		var ce *providerCredentialError
		if errors.As(classifyErr, &ce) {
			result.Kind = ce.Kind
		}
		result.Message = classifyErr.Error()
		return result
	}
	result.OK = true
	result.Message = fmt.Sprintf("%s: credential accepted (HTTP %d).", shape.Label, resp.StatusCode)
	return result
}

func byoValidationClient() *http.Client {
	return &http.Client{Timeout: 12 * time.Second}
}

// detectOpenAICompatibleProvider reports readiness for a routable BYO chat
// shape. Setup is just the key: the model defaults so a pasted credential works.
func detectOpenAICompatibleProvider(shape byoShape) providerConfig {
	missing := []string{}
	if strings.TrimSpace(configValue(shape.KeyEnv)) == "" {
		missing = append(missing, shape.KeyEnv)
	}
	model := firstNonEmpty(configValue(shape.modelEnv), shape.defaultModel)
	base := firstNonEmpty(configValue(shape.baseEnv), shape.defaultBase)
	label := shape.Label
	if model != "" {
		label += " / " + model
	}
	return providerConfig{
		ID:      shape.ID,
		Label:   label,
		Status:  statusFromMissing(missing),
		Missing: missing,
		Settings: []string{
			"JINI_PROVIDER: " + shape.ID + " -> " + shape.Label,
			shape.modelEnv + ": " + presentOrMissing(shape.modelEnv) + " (default " + shape.defaultModel + ")",
			shape.baseEnv + ": " + base,
		},
		Secrets: []string{shape.KeyEnv + ": " + presentOrMissing(shape.KeyEnv)},
	}
}
