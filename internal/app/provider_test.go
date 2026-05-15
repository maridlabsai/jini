package app

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func withProviderHTTPClient(t *testing.T, fn roundTripFunc) {
	t.Helper()
	previous := providerHTTPClient
	providerHTTPClient = &http.Client{Transport: fn}
	t.Cleanup(func() { providerHTTPClient = previous })
}

func TestGenerateWithConfiguredProviderCallsAzureOpenAI(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")
	t.Setenv("AZURE_OPENAI_API_VERSION", "2024-10-21")

	requestSeen := false
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		requestSeen = true
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "https://example.openai.azure.com/openai/deployments/gpt-4o-prod/chat/completions?api-version=2024-10-21" {
			t.Fatalf("unexpected Azure URL: %s", req.URL.String())
		}
		if got := req.Header.Get("api-key"); got != "super-secret-key" {
			t.Fatalf("unexpected Azure api-key header: %q", got)
		}
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, "Plan 7 day Paris trip") {
			t.Fatalf("expected source in Azure request, got:\n%s", body)
		}
		if strings.Contains(body, "super-secret-key") {
			t.Fatalf("Azure request body leaked API key:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Plan 7 day Paris trip",
	})
	if err != nil {
		t.Fatalf("generate with Azure: %v", err)
	}
	if !used || !requestSeen {
		t.Fatalf("expected Azure provider to be used")
	}
	if !strings.Contains(result, "Provider Paris") {
		t.Fatalf("expected provider content, got:\n%s", result)
	}
}

func TestGenerateWithConfiguredProviderCallsAnthropicMessages(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "super-secret-key")
	t.Setenv("JINI_MODEL", "sonnet")

	requestSeen := false
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		requestSeen = true
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "https://api.anthropic.com/v1/messages" {
			t.Fatalf("unexpected Anthropic URL: %s", req.URL.String())
		}
		if got := req.Header.Get("x-api-key"); got != "super-secret-key" {
			t.Fatalf("unexpected Anthropic api key header: %q", got)
		}
		if got := req.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("unexpected anthropic version: %q", got)
		}
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, "claude-sonnet-4-20250514") {
			t.Fatalf("expected resolved Anthropic model in body, got:\n%s", body)
		}
		if strings.Contains(body, "super-secret-key") {
			t.Fatalf("Anthropic request body leaked API key:\n%s", body)
		}
		return jsonResponse(200, `{"content":[{"type":"text","text":"# First Useful Pass: Claude Draft\n\nAnthropic draft."}]}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Claude Draft",
		Source: "Turn this into a useful first pass.",
	})
	if err != nil {
		t.Fatalf("generate with Anthropic: %v", err)
	}
	if !used || !requestSeen {
		t.Fatalf("expected Anthropic provider to be used")
	}
	if !strings.Contains(result, "Anthropic draft") {
		t.Fatalf("expected Anthropic content, got:\n%s", result)
	}
}

func TestGenerateWithConfiguredProviderCallsBedrockConverse(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "bedrock")
	t.Setenv("JINI_BEDROCK_ENDPOINT", "https://bedrock-runtime.us-east-1.amazonaws.com")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "SECRETEXAMPLE")
	t.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-3-5-sonnet-20240620-v1:0")

	requestSeen := false
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		requestSeen = true
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-3-5-sonnet-20240620-v1:0/converse" {
			t.Fatalf("unexpected Bedrock URL: %s", req.URL.String())
		}
		auth := req.Header.Get("Authorization")
		for _, want := range []string{"AWS4-HMAC-SHA256", "Credential=AKIAEXAMPLE/", "/us-east-1/bedrock/aws4_request", "SignedHeaders="} {
			if !strings.Contains(auth, want) {
				t.Fatalf("expected Authorization to contain %q, got:\n%s", want, auth)
			}
		}
		for _, header := range []string{"X-Amz-Date", "X-Amz-Content-Sha256"} {
			if req.Header.Get(header) == "" {
				t.Fatalf("expected Bedrock header %s", header)
			}
		}
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, "Weekly product review") {
			t.Fatalf("expected source in Bedrock request, got:\n%s", body)
		}
		if strings.Contains(body, "SECRETEXAMPLE") {
			t.Fatalf("Bedrock request body leaked AWS secret:\n%s", body)
		}
		return jsonResponse(200, `{"output":{"message":{"role":"assistant","content":[{"text":"# Sendable Follow-Up: Provider Meeting\n\nBedrock draft."}]}}}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Weekly Product Review",
		Source: "Weekly product review. Need owners and due dates.",
	})
	if err != nil {
		t.Fatalf("generate with Bedrock: %v", err)
	}
	if !used || !requestSeen {
		t.Fatalf("expected Bedrock provider to be used")
	}
	if !strings.Contains(result, "Bedrock draft") {
		t.Fatalf("expected Bedrock content, got:\n%s", result)
	}
}

func TestGenerateWithConfiguredProviderAutoPrefersBedrockForSonnet46Alias(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "sonnet-4.6")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "SECRETEXAMPLE")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.String(), "bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-sonnet-4-6/converse") {
			t.Fatalf("expected auto mode to choose Bedrock Sonnet 4.6, got %s", req.URL.String())
		}
		return jsonResponse(200, `{"output":{"message":{"content":[{"text":"Bedrock auto draft."}]}}}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Auto Mode",
		Source: "Use the best configured provider automatically.",
	})
	if err != nil {
		t.Fatalf("generate with auto provider: %v", err)
	}
	if !used || !strings.Contains(result, "Bedrock auto draft") {
		t.Fatalf("expected Bedrock auto draft, used=%v result=%q", used, result)
	}
}

func TestGenerateWithConfiguredProviderUsesAWSProfileCredentialsAndRegion(t *testing.T) {
	awsDir := t.TempDir()
	credentialsPath := filepath.Join(awsDir, "credentials")
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(credentialsPath, []byte("[work]\naws_access_key_id = PROFILEKEY\naws_secret_access_key = PROFILESECRET\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("[profile work]\nregion = us-west-2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("JINI_PROVIDER", "bedrock")
	t.Setenv("JINI_BEDROCK_ENDPOINT", "https://bedrock-runtime.us-west-2.amazonaws.com")
	t.Setenv("AWS_PROFILE", "work")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsPath)
	t.Setenv("AWS_CONFIG_FILE", configPath)
	t.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-3-5-sonnet-20240620-v1:0")

	provider := detectProvider()
	if provider.Status != "ok" {
		t.Fatalf("expected profile-backed provider to be ok, got %#v", provider)
	}

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "bedrock-runtime.us-west-2.amazonaws.com" {
			t.Fatalf("expected profile region in host, got %s", req.URL.Host)
		}
		auth := req.Header.Get("Authorization")
		for _, want := range []string{"Credential=PROFILEKEY/", "/us-west-2/bedrock/aws4_request"} {
			if !strings.Contains(auth, want) {
				t.Fatalf("expected Authorization to contain %q, got:\n%s", want, auth)
			}
		}
		if strings.Contains(mustReadAll(t, req.Body), "PROFILESECRET") {
			t.Fatalf("request body leaked profile secret")
		}
		return jsonResponse(200, `{"output":{"message":{"content":[{"text":"Profile-backed Bedrock draft."}]}}}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Profile Test",
		Source: "Use profile credentials.",
	})
	if err != nil {
		t.Fatalf("generate with profile: %v", err)
	}
	if !used || !strings.Contains(result, "Profile-backed Bedrock draft") {
		t.Fatalf("expected profile-backed Bedrock draft, used=%v result=%q", used, result)
	}
}

func TestMaybeWriteProviderFirstDraftOverwritesPrimaryView(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "artifacts"), 0o755); err != nil {
		t.Fatal(err)
	}
	choice := starterChoice{PackID: "travel-plan"}
	if err := writeTravelStarterWork(workDir, "7-Day Paris Trip", "Plan 7 day Paris trip", "quick"); err != nil {
		t.Fatalf("write local starter: %v", err)
	}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, "7-Day Paris Trip", "Plan 7 day Paris trip"); err != nil {
		t.Fatalf("write provider draft: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "views", "itinerary.md"))
	if err != nil {
		t.Fatalf("read itinerary: %v", err)
	}
	if !strings.Contains(string(content), "Provider day one") {
		t.Fatalf("expected primary view to use provider draft, got:\n%s", string(content))
	}
}

func TestBootstrapStarterWorkUsesConfiguredProviderDraft(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	summary, err := bootstrapStarterWork(starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"}, "Plan 7 day Paris trip", "quick")
	if err != nil {
		t.Fatalf("bootstrap with provider: %v", err)
	}
	if summary.Title == "" || len(summary.Views) == 0 {
		t.Fatalf("expected provider-backed summary, got %#v", summary)
	}
	content, err := os.ReadFile(filepath.Join(summary.Dir, "views", "itinerary.md"))
	if err != nil {
		t.Fatalf("read itinerary: %v", err)
	}
	if !strings.Contains(string(content), "Provider day one") {
		t.Fatalf("expected bootstrap to save provider draft, got:\n%s", string(content))
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func mustReadAll(t *testing.T, reader io.Reader) string {
	t.Helper()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(data)
}
