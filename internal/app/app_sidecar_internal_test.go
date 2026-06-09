package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHiddenAppSidecarServeCommandRequiresExactShape(t *testing.T) {
	if !isHiddenAppSidecarServeCommand([]string{"app", "serve", "--stdio", "--surface", "macos"}) {
		t.Fatalf("expected exact macOS sidecar command to match")
	}

	for _, args := range [][]string{
		{"app"},
		{"app", "serve"},
		{"app", "serve", "--stdio"},
		{"app", "serve", "--surface", "macos", "--stdio"},
		{"app", "serve", "--stdio", "--surface", "ios"},
		{"app", "serve", "--stdio", "--surface", "macos", "--verbose"},
		{"app", "status", "--stdio", "--surface", "macos"},
	} {
		if isHiddenAppSidecarServeCommand(args) {
			t.Fatalf("expected non-exact sidecar command not to match: %#v", args)
		}
	}
}

func TestAppSidecarCommandStaysOutOfPublicCommandCanonicalization(t *testing.T) {
	if got := canonicalTopLevelCommand("app"); got != "" {
		t.Fatalf("expected app not to be a public top-level command, got %q", got)
	}
	if got := canonicalHelpTopic("app"); got != "" {
		t.Fatalf("expected app not to be a public help topic, got %q", got)
	}
}

func TestDiagnosticsTextRedactionRemovesStateHomeAndSecretValues(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", "/tmp/jini-state")
	t.Setenv("HOME", "/Users/example")
	t.Setenv("OPENAI_API_KEY", "sk-test-secret")

	redacted := redactDiagnosticsText("state=/tmp/jini-state home=/Users/example key=sk-test-secret")
	for _, forbidden := range []string{"/tmp/jini-state", "/Users/example", "sk-test-secret"} {
		if strings.Contains(redacted, forbidden) {
			t.Fatalf("expected diagnostics text to redact %q, got %q", forbidden, redacted)
		}
	}
	for _, want := range []string{"$JINI_STATE_DIR", "~", "[redacted]"} {
		if !strings.Contains(redacted, want) {
			t.Fatalf("expected diagnostics redaction marker %q, got %q", want, redacted)
		}
	}
}

func TestMacOSAppSnapshotUsesSharedOfflineRouteState(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_CONNECTIVITY_OVERRIDE", "offline")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	snapshot := newMacOSAppSidecar().snapshot()
	if snapshot.OnlineState != "offline" {
		t.Fatalf("expected snapshot online_state to reflect offline connectivity, got %#v", snapshot)
	}
	if snapshot.OfflineState != "local-preview" {
		t.Fatalf("expected snapshot offline_state to reflect fallback route, got %#v", snapshot)
	}
	if snapshot.RouteSummary.RouteID != "local-preview" {
		t.Fatalf("expected app route status to use shared offline route decision, got %#v", snapshot.RouteSummary)
	}
	if snapshot.RouteSummary.OfflineState != "active" {
		t.Fatalf("expected route offline_state to be active, got %#v", snapshot.RouteSummary)
	}
}

func TestMacOSAppTurnSubmitUsesSharedProviderRouteEngine(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_CONNECTIVITY_OVERRIDE", "offline")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "phi4-mini")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "127.0.0.1:11434" {
			t.Fatalf("expected app turn submit to use local route, got %s", req.URL.String())
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"Local app answer."}}]}`), nil
	})

	params, err := json.Marshal(turnSubmitParams{Text: "summarize this note: Jini should route offline"})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	response := newMacOSAppSidecar().dispatch(appRPCRequest{
		ProtocolVersion: macOSAppProtocolVersion,
		ID:              "req_turn",
		IdempotencyKey:  "idem_turn",
		Method:          "turn.submit",
		Surface:         "macos",
		Params:          params,
	})
	if !response.OK {
		t.Fatalf("expected routed app turn to succeed, got %#v", response)
	}
	result, ok := response.Result.(transientResponseVM)
	if !ok {
		t.Fatalf("expected transient response result, got %#v", response.Result)
	}
	if result.AssistantText != "Local app answer." {
		t.Fatalf("expected local provider answer, got %#v", result)
	}
	if !result.RouteVisible {
		t.Fatalf("expected routed app answer to expose route visibility")
	}
}
