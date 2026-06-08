package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestGenerateWithConfiguredProviderHandsOffToConfiguredCLI(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	captureDir := t.TempDir()
	argsPath := filepath.Join(captureDir, "args.txt")
	stdinPath := filepath.Join(captureDir, "stdin.txt")
	t.Setenv("JINI_TEST_CLI_ARGS_PATH", argsPath)
	t.Setenv("JINI_TEST_CLI_STDIN_PATH", stdinPath)
	writeProviderFakeExecutable(t, binDir, "claude", strings.Join([]string{
		"printf '%s\\n' \"$@\" > \"$JINI_TEST_CLI_ARGS_PATH\"",
		"cat > \"$JINI_TEST_CLI_STDIN_PATH\"",
		"printf 'fake claude handled request\\n'",
	}, "\n"))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result, used, actualDecision, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Fix failing tests",
		Source: "fix failing tests",
	}, detectRouteForToolMode("claude-code", false))
	if err != nil {
		t.Fatalf("generate with CLI handoff: %v", err)
	}
	if !used {
		t.Fatalf("expected CLI handoff to be used")
	}
	if strings.TrimSpace(result) != "fake claude handled request" {
		t.Fatalf("expected fake CLI output, got:\n%s", result)
	}
	if actualDecision.ToolMode != "claude-code" || actualDecision.RoutePolicy != "CLI handoff" {
		t.Fatalf("expected actual CLI handoff decision, got %#v", actualDecision)
	}
	args := mustReadFile(t, argsPath)
	if !strings.Contains(args, "--print") || !strings.Contains(args, "fix failing tests") {
		t.Fatalf("expected Claude CLI handoff args to include --print and source prompt, got:\n%s", args)
	}
	stdin := mustReadFile(t, stdinPath)
	if strings.TrimSpace(stdin) != "" {
		t.Fatalf("expected prompt to be passed as args, not stdin, got:\n%s", stdin)
	}
}

func TestGenerateWithConfiguredProviderPreservesQuotedCustomCLIArgs(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CLAUDE_CODE_ARGS", `--model "Claude Sonnet" --permission-mode acceptEdits {{prompt}}`)
	binDir := t.TempDir()
	captureDir := t.TempDir()
	argsPath := filepath.Join(captureDir, "args.txt")
	t.Setenv("JINI_TEST_CLI_ARGS_PATH", argsPath)
	writeProviderFakeExecutable(t, binDir, "claude", strings.Join([]string{
		"printf '%s\\n' \"$@\" > \"$JINI_TEST_CLI_ARGS_PATH\"",
		"printf 'fake claude handled quoted args\\n'",
	}, "\n"))
	t.Setenv("PATH", binDir)

	_, used, actualDecision, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Fix failing tests",
		Source: "fix failing tests",
	}, detectRouteForToolMode("claude-code", false))
	if err != nil {
		t.Fatalf("generate with quoted CLI handoff args: %v", err)
	}
	if !used {
		t.Fatalf("expected CLI handoff to be used")
	}
	if actualDecision.CLIHandoffReceipt == nil {
		t.Fatalf("expected CLI handoff receipt")
	}
	args := strings.Split(strings.TrimSpace(mustReadFile(t, argsPath)), "\n")
	want := []string{"--model", "Claude Sonnet", "--permission-mode", "acceptEdits", "fix failing tests"}
	if strings.Join(args, "\n") != strings.Join(want, "\n") {
		t.Fatalf("expected quoted args to stay grouped as %#v, got %#v", want, args)
	}
	if got := actualDecision.CLIHandoffReceipt.ArgsTemplate; strings.Join(got, "\n") != strings.Join([]string{"--model", "Claude Sonnet", "--permission-mode", "acceptEdits", "{{prompt}}"}, "\n") {
		t.Fatalf("expected receipt args template to preserve parsed quoted arg, got %#v", got)
	}
	if got := formatCLIHandoffArgs(actualDecision.CLIHandoffReceipt.ArgsTemplate); got != `--model "Claude Sonnet" --permission-mode acceptEdits {{prompt}}` {
		t.Fatalf("expected inspectable quoted args display, got %q", got)
	}
}

func TestGenerateWithConfiguredProviderFailsClosedForMalformedCustomCLIArgs(t *testing.T) {
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CLAUDE_CODE_ARGS", `--model "Claude Sonnet`)
	binDir := t.TempDir()
	writeProviderFakeExecutable(t, binDir, "claude", "printf 'should not run\\n'")
	t.Setenv("PATH", binDir)

	_, used, _, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Fix failing tests",
		Source: "fix failing tests",
	}, detectRouteForToolMode("claude-code", false))
	if err == nil {
		t.Fatalf("expected malformed CLI args to fail closed")
	}
	if !used {
		t.Fatalf("expected CLI handoff route to be selected")
	}
	for _, want := range []string{
		"CLI handoff needs setup",
		"JINI_CLAUDE_CODE_ARGS has invalid quoting",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected setup error to contain %q, got %v", want, err)
		}
	}
}

func TestMaybeWriteProviderFirstDraftPersistsPrivacyPreservingCLIHandoffReceipt(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	fakeCLI := writeProviderFakeExecutable(t, binDir, "claude", "printf 'fake claude draft\\n'")
	t.Setenv("PATH", binDir)

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	source := "Summarize the private launch notes without leaking the prompt into route state."
	if err := maybeWriteProviderFirstDraft(context.Background(), starterChoice{PackID: "general-work"}, workDir, "Launch Notes", source); err != nil {
		t.Fatalf("write CLI-backed provider draft: %v", err)
	}

	routeSaved := mustReadFile(t, filepath.Join(workDir, "route.json"))
	for _, want := range []string{
		`"cli_handoff_receipt": {`,
		`"context_type": "JiniCLIHandoffReceipt"`,
		`"mode": "claude-code"`,
		`"label": "Claude Code CLI handoff"`,
		`"executable": "` + fakeCLI + `"`,
		`"args_template": [`,
		`"--print"`,
		`"{{prompt}}"`,
		`"exit_status": 0`,
		`"cwd": "`,
		`"duration_ms":`,
		`"prompt_chars": 79`,
		`"stdout_chars": 18`,
		`"stderr_chars": 0`,
	} {
		if !strings.Contains(routeSaved, want) {
			t.Fatalf("expected route receipt to contain %q, got:\n%s", want, routeSaved)
		}
	}
	for _, unwanted := range []string{
		source,
		"fake claude draft",
	} {
		if strings.Contains(routeSaved, unwanted) {
			t.Fatalf("route receipt must not persist prompt or CLI output body %q, got:\n%s", unwanted, routeSaved)
		}
	}
}

func TestRunInteractiveKeepsCurrentWorkAfterFailedCLIHandoff(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	writeProviderFakeExecutable(t, binDir, "claude", strings.Join([]string{
		"printf 'partial stdout\\n'",
		"printf 'downstream secret stderr\\n' >&2",
		"exit 7",
	}, "\n"))
	t.Setenv("PATH", binDir)

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected failed downstream CLI to keep local work available, got %d:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Result ready.") {
		t.Fatalf("expected local starter result despite failed handoff, got:\n%s", stdout.String())
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work to remain available after failed handoff: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	routeSaved := mustReadFile(t, filepath.Join(packDir, "route.json"))
	for _, want := range []string{
		`"cli_handoff_receipt": {`,
		`"exit_status": 7`,
		`"stdout_chars": 15`,
		`"stderr_chars": 25`,
	} {
		if !strings.Contains(routeSaved, want) {
			t.Fatalf("expected failed handoff route receipt to contain %q, got:\n%s", want, routeSaved)
		}
	}
	for _, unwanted := range []string{
		"partial stdout",
		"downstream secret stderr",
	} {
		if strings.Contains(routeSaved, unwanted) {
			t.Fatalf("route receipt must not persist CLI output body %q, got:\n%s", unwanted, routeSaved)
		}
	}

	var status bytes.Buffer
	exitCode = Run([]string{"status"}, &status, &status)
	if exitCode != 0 {
		t.Fatalf("expected status after failed handoff, got %d:\n%s", exitCode, status.String())
	}
	statusOut := status.String()
	for _, want := range []string{
		"Last CLI handoff",
		"Status: failed",
		"Exit 7 in",
		"stdout 15 chars",
		"stderr 25 chars",
	} {
		if !strings.Contains(statusOut, want) {
			t.Fatalf("expected status to contain %q, got:\n%s", want, statusOut)
		}
	}
	for _, unwanted := range []string{
		"partial stdout",
		"downstream secret stderr",
	} {
		if strings.Contains(statusOut, unwanted) {
			t.Fatalf("status must not expose CLI output body %q, got:\n%s", unwanted, statusOut)
		}
	}
}

func TestGenerateWithConfiguredProviderSanitizesFailedCLIHandoffError(t *testing.T) {
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	writeProviderFakeExecutable(t, binDir, "claude", strings.Join([]string{
		"printf 'partial stdout\\n'",
		"printf 'downstream secret stderr\\n' >&2",
		"exit 9",
	}, "\n"))
	t.Setenv("PATH", binDir)

	_, used, actualDecision, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Failing CLI",
		Source: "Use a downstream CLI that fails.",
	}, detectRouteForToolMode("claude-code", false))
	if err == nil {
		t.Fatalf("expected failed CLI handoff error")
	}
	if !used {
		t.Fatalf("expected CLI handoff route to be selected")
	}
	if actualDecision.CLIHandoffReceipt == nil || actualDecision.CLIHandoffReceipt.ExitStatus != 9 {
		t.Fatalf("expected failed handoff receipt on actual decision, got %#v", actualDecision.CLIHandoffReceipt)
	}
	errorText := err.Error()
	for _, want := range []string{
		"Claude Code CLI handoff failed: exit status 9",
		"stdout 15 chars",
		"stderr 25 chars",
		"output omitted",
	} {
		if !strings.Contains(errorText, want) {
			t.Fatalf("expected sanitized error to contain %q, got %q", want, errorText)
		}
	}
	for _, unwanted := range []string{
		"partial stdout",
		"downstream secret stderr",
	} {
		if strings.Contains(errorText, unwanted) {
			t.Fatalf("sanitized error must not expose CLI output body %q, got %q", unwanted, errorText)
		}
	}
}

func TestSaveWorkRouteClearsCLIHandoffReceiptWhenRouteSwitchesAwayFromCLI(t *testing.T) {
	workDir := t.TempDir()
	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Switch Route",
		Source: "Switch this work back to local preview.",
	}
	if err := saveWorkRoute(workDir, request, routeDecision{
		Active:      true,
		ToolMode:    "claude-code",
		ToolLabel:   "Claude Code CLI handoff",
		RoutePolicy: "CLI handoff",
		Provider:    detectCLIHandoffProvider("claude-code"),
		CLIHandoffReceipt: &cliHandoffReceipt{
			ContextType: "JiniCLIHandoffReceipt",
			Mode:        "claude-code",
			Label:       "Claude Code CLI handoff",
			Executable:  "/tmp/fake-claude",
			ExitStatus:  0,
		},
	}); err != nil {
		t.Fatalf("save CLI route: %v", err)
	}

	if err := saveWorkRoute(workDir, request, routeDecision{
		Active:      true,
		ToolMode:    "local-preview",
		ToolLabel:   "Local preview",
		RoutePolicy: "Local fallback",
		Provider:    detectLocalPreviewProvider(),
	}); err != nil {
		t.Fatalf("save local route: %v", err)
	}

	routeSaved := mustReadFile(t, filepath.Join(workDir, "route.json"))
	if strings.Contains(routeSaved, "cli_handoff_receipt") || strings.Contains(routeSaved, "fake-claude") {
		t.Fatalf("expected route switch away from CLI to clear stale handoff receipt, got:\n%s", routeSaved)
	}
	if !strings.Contains(routeSaved, `"previous_tool_mode": "claude-code"`) {
		t.Fatalf("expected route switch metadata to remain, got:\n%s", routeSaved)
	}
}

func TestGenerateWithConfiguredProviderFailsClosedForMissingCLI(t *testing.T) {
	t.Setenv("JINI_TOOL", "aider")
	t.Setenv("PATH", t.TempDir())

	_, used, _, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Fix failing tests",
		Source: "fix failing tests",
	}, detectRouteForToolMode("aider", false))
	if err == nil {
		t.Fatalf("expected missing CLI handoff to fail closed")
	}
	if !used {
		t.Fatalf("expected missing CLI handoff route to be selected and fail closed")
	}
	if !strings.Contains(err.Error(), "CLI handoff needs setup") {
		t.Fatalf("expected CLI handoff setup error, got %v", err)
	}
	if strings.Contains(err.Error(), "Provider needs setup") {
		t.Fatalf("missing CLI handoff must not be reported as provider setup: %v", err)
	}
}

func TestGenerateWithConfiguredProviderFailsClosedWhenCLIHandoffTrustCheckRejectsExecutable(t *testing.T) {
	t.Setenv("JINI_TOOL", "claude-code")
	binDir := t.TempDir()
	markerPath := filepath.Join(t.TempDir(), "executed.txt")
	t.Setenv("JINI_TEST_CLI_MARKER", markerPath)
	fakeCLI := writeProviderFakeExecutable(t, binDir, "claude", "printf 'executed\\n' > \"$JINI_TEST_CLI_MARKER\"\n")
	t.Setenv("PATH", binDir)

	previousTrustCheck := cliHandoffTrustIssueForPath
	cliHandoffTrustIssueForPath = func(path string) string {
		if path != fakeCLI {
			t.Fatalf("expected trust check to inspect resolved fake CLI %q, got %q", fakeCLI, path)
		}
		return "macOS Gatekeeper rejected CLI executable: " + path + "."
	}
	t.Cleanup(func() { cliHandoffTrustIssueForPath = previousTrustCheck })

	_, used, actualDecision, err := generateWithConfiguredProviderDecision(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Fix failing tests",
		Source: "fix failing tests",
	}, detectRouteForToolMode("claude-code", false))
	if err == nil {
		t.Fatalf("expected rejected CLI executable to fail closed")
	}
	if !used {
		t.Fatalf("expected CLI handoff route to be selected")
	}
	if actualDecision.CLIHandoffReceipt != nil {
		t.Fatalf("rejected CLI must not execute or create a handoff receipt, got %#v", actualDecision.CLIHandoffReceipt)
	}
	for _, want := range []string{
		"CLI handoff needs setup",
		"macOS Gatekeeper rejected CLI executable: " + fakeCLI + ".",
		"until the executable passes local OS trust checks",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected trust setup error to contain %q, got %v", want, err)
		}
	}
	if _, statErr := os.Stat(markerPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("rejected CLI must not execute; marker stat err: %v", statErr)
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
		return jsonResponse(200, `{"content":[{"type":"text","text":"# Working Draft: Claude Draft\n\nAnthropic draft."}]}`), nil
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

func TestGenerateWithConfiguredProviderCallsLocalSLMOpenAICompatible(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_API_KEY", "local-secret")

	requestSeen := false
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		requestSeen = true
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "http://127.0.0.1:11434/v1/chat/completions" {
			t.Fatalf("unexpected Local SLM URL: %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer local-secret" {
			t.Fatalf("unexpected Local SLM auth header: %q", got)
		}
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, `"model":"qwen3:8b"`) {
			t.Fatalf("expected local model in request body, got:\n%s", body)
		}
		if !strings.Contains(body, "Turn meeting notes into something I can send") {
			t.Fatalf("expected source in local request, got:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Weekly Product Review",
		Source: "Turn meeting notes into something I can send.",
	})
	if err != nil {
		t.Fatalf("generate with Local SLM: %v", err)
	}
	if !used || !requestSeen {
		t.Fatalf("expected Local SLM provider to be used")
	}
	if !strings.Contains(result, "Local draft") {
		t.Fatalf("expected Local SLM content, got:\n%s", result)
	}
	report := loadLocalRuntimeCapabilities()
	history := report.CohortHistory["local-workhorse"]["sendable-followup"]
	if len(history) != 1 {
		t.Fatalf("expected Local SLM request to record sendable-followup cohort history, got %#v", history)
	}
	if history[0].QualityClass != "strong" || history[0].StructuredReliability != "strong" || history[0].Status != "ok" {
		t.Fatalf("expected strong sendable-followup cohort history, got %#v", history[0])
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

func TestGenerateWithConfiguredProviderRecordsSubtypeSpecificMultimodalCohortHistory(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, `"model":"gemma3:12b"`) {
			t.Fatalf("expected multimodal model in request body, got:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Extracted evidence\n- The screenshot shows an error banner.\n\n## What is visible\n- The save button is disabled.\n\n## Still unclear\n- The hidden validation rule is not shown.\n\n## Recommended next move\n- Reproduce with the field expanded."}}]}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "UI Screenshot Review",
		Source: "Review this screenshot and extract the evidence from the visible UI states.",
	})
	if err != nil {
		t.Fatalf("generate with Local SLM multimodal: %v", err)
	}
	if !used || !strings.Contains(result, "## What is visible") {
		t.Fatalf("expected multimodal Local SLM content, used=%v result=%q", used, result)
	}
	report := loadLocalRuntimeCapabilities()
	history := report.CohortHistory["local-multimodal"]["multimodal-image-screenshot"]
	if len(history) != 1 {
		t.Fatalf("expected subtype-specific multimodal cohort history, got %#v", history)
	}
	if history[0].QualityClass != "strong" || history[0].StructuredReliability != "strong" || history[0].Status != "ok" {
		t.Fatalf("expected strong subtype-specific multimodal cohort history, got %#v", history[0])
	}
}

func TestDetectRouteForRequestAutoPrefersLocalWorkhorseForTravelPlanningWhenReady(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Plan 7 day Paris trip with a clear itinerary and budget.",
	})
	if decision.ToolMode != "local-workhorse" {
		t.Fatalf("expected Local SLM workhorse for travel planning, got %#v", decision)
	}
	if decision.ToolLabel != "Local SLM workhorse" {
		t.Fatalf("expected Local SLM workhorse label, got %#v", decision)
	}
	if decision.ModelLabel != "qwen3:8b-instruct" {
		t.Fatalf("expected Local SLM workhorse model label, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "planning work") || !strings.Contains(decision.Reason, "cheapest suitable") {
		t.Fatalf("expected local planning reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoKeepsCurrentCodingRouteWhenGapIsNotMaterial(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "test-claude-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	workDir := t.TempDir()
	writeProviderTestCurrentWork(t, workDir, "research-prd", "existing-coding-work", "Large Coding Project")
	if err := saveWorkRoute(workDir, providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Large Coding Project",
		Source: "Continue implementing and refactoring the codebase.",
	}, routeDecision{
		Active:              true,
		ToolMode:            "claude-api",
		ToolLabel:           "Claude API route",
		ChosenAutomatically: true,
		Provider:            detectProviderForMode("anthropic"),
		RoutePolicy:         "Automatic",
		ModelLabel:          "Claude Sonnet 4",
		Reason:              "Previous coding route.",
	}); err != nil {
		t.Fatalf("save work route: %v", err)
	}

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Large Coding Project",
		Source: "Continue refactoring this code and clean up the implementation.",
	})
	if decision.ToolMode != "claude-api" {
		t.Fatalf("expected continuity to keep claude-api, got %#v", decision)
	}
	if !strings.Contains(decision.ContinuityReason, "preserve context continuity") {
		t.Fatalf("expected continuity explanation field, got %#v", decision)
	}
	if strings.Contains(decision.Reason, "preserve context continuity") {
		t.Fatalf("expected generic reason to stay separate from continuity explanation, got %#v", decision)
	}
}

func TestRenderRouteDecisionCardShowsContinuityReason(t *testing.T) {
	var stdout bytes.Buffer
	renderRouteDecisionCard(&stdout, providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Large Coding Project",
		Source: "Continue implementing and refactoring the codebase.",
	}, routeDecision{
		Active:              true,
		ToolMode:            "claude-api",
		ToolLabel:           "Claude API route",
		ChosenAutomatically: true,
		Provider:            detectProviderForMode("anthropic"),
		RoutePolicy:         "Automatic",
		ModelLabel:          "Claude Sonnet 4",
		Reason:              "Auto mode picked the cheapest suitable coding route.",
		ContinuityReason:    "Kept the current coding route to preserve context continuity because the quality gap was not material.",
	})

	out := stdout.String()
	for _, want := range []string{
		"Continuity",
		"Kept the current coding route to preserve context continuity because the quality gap was not material.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route decision card to contain %q, got:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "Why this route") {
		t.Fatalf("expected route decision card to keep the generic route explanation, got:\n%s", out)
	}
}

func TestDetectRouteForRequestAutoSwitchesCodingRouteWhenGapIsMaterial(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "phi4-mini")

	workDir := t.TempDir()
	writeProviderTestCurrentWork(t, workDir, "research-prd", "existing-coding-work", "Large Coding Project")
	if err := saveWorkRoute(workDir, providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Large Coding Project",
		Source: "Continue implementing and refactoring the codebase.",
	}, routeDecision{
		Active:              true,
		ToolMode:            "local-fast",
		ToolLabel:           "Local SLM fast",
		ChosenAutomatically: true,
		Provider:            detectProviderForMode("local-slm"),
		RoutePolicy:         "Automatic",
		ModelLabel:          "phi4-mini",
		Reason:              "Previous local route.",
	}); err != nil {
		t.Fatalf("save work route: %v", err)
	}

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Architecture Review",
		Source: "Do a deep architecture review of this codebase and produce a rigorous implementation plan.",
	})
	if decision.ToolMode == "local-fast" {
		t.Fatalf("expected material gap to force a stronger route, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoUsesManualCodingRoutePreference(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "test-claude-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Large Coding Project",
		Source: "Implement the code changes and fix the failing tests in this repository.",
	}
	stats := localRouteFeedbackStats{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniRouteFeedback",
		Routes:        map[string]routeFeedbackRow{},
		Cohorts:       map[string]map[string]localCohortFeedbackRow{},
		ManualOverrides: map[string]map[string]int{
			"claude-api": {
				"build-readiness": 4,
			},
		},
	}
	if err := saveLocalRouteFeedbackStats(stats); err != nil {
		t.Fatalf("save route feedback stats: %v", err)
	}

	decision := detectRouteForRequest(request)
	if decision.ToolMode != "claude-api" {
		t.Fatalf("expected manual coding preference to bias claude-api, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "Past route choices on similar coding work") {
		t.Fatalf("expected manual preference explanation, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersLocalFastOnTinyDeviceForQuickPass(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "tiny")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Quick follow-up",
		Source: "Quick one-line follow-up from this meeting.",
	})
	if decision.ToolMode != "local-fast" {
		t.Fatalf("expected Local SLM fast for a tiny device quick pass, got %#v", decision)
	}
	if decision.ModelLabel != "phi4-mini" {
		t.Fatalf("expected fast local model label, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersLocalMultimodalForScannedPDFWhenReady(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Invoice Review",
		Source: "Review this scanned PDF invoice and extract the evidence.",
	})
	if decision.ToolMode != "local-multimodal" {
		t.Fatalf("expected Local SLM multimodal for scanned PDF work, got %#v", decision)
	}
	if decision.ModelLabel != "gemma3:12b" {
		t.Fatalf("expected Local SLM multimodal model label, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "scanned document or PDF evidence") {
		t.Fatalf("expected PDF-aware route reason, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "learns scanned PDF and document work separately") {
		t.Fatalf("expected PDF subtype learning note in route reason, got %#v", decision)
	}
	if !strings.Contains(decision.ModelReason, "learns scanned PDF and document work separately") {
		t.Fatalf("expected PDF subtype learning note in model reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersLocalMultimodalForScreenshotWhenReady(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "UI Screenshot Review",
		Source: "Review this screenshot and extract the evidence from the visible UI states.",
	})
	if decision.ToolMode != "local-multimodal" {
		t.Fatalf("expected Local SLM multimodal for screenshot work, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "image or screenshot evidence") {
		t.Fatalf("expected screenshot-aware route reason, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "learns screenshot work separately") {
		t.Fatalf("expected screenshot subtype learning note in route reason, got %#v", decision)
	}
	if !strings.Contains(decision.ModelReason, "learns screenshot work separately") {
		t.Fatalf("expected screenshot subtype learning note in model reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersLocalWorkhorseForAudioTranscriptWhenReady(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Call Review",
		Source: "Review this audio recording and transcript and extract the evidence.",
	})
	if decision.ToolMode != "local-workhorse" {
		t.Fatalf("expected Local SLM workhorse for audio transcript work, got %#v", decision)
	}
	if decision.ModelLabel != "qwen3:8b-instruct" {
		t.Fatalf("expected Local SLM workhorse model label, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "audio or transcript evidence") {
		t.Fatalf("expected audio-aware route reason, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "learns audio/transcript work separately") {
		t.Fatalf("expected audio subtype learning note in route reason, got %#v", decision)
	}
	if !strings.Contains(decision.ModelReason, "learns audio/transcript work separately") {
		t.Fatalf("expected audio subtype learning note in model reason, got %#v", decision)
	}
}

func TestClassifyModelChoiceIncludesDeviceClassForLocalRoute(t *testing.T) {
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "workstation")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")

	label, reason := classifyModelChoice(providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "Trip",
		Source: "Plan a 7 day Paris trip",
	}, "local-workhorse")
	if label != "qwen3:8b-instruct" {
		t.Fatalf("expected workhorse local model label, got %q", label)
	}
	if !strings.Contains(reason, "workstation") {
		t.Fatalf("expected local model reason to mention device class, got %q", reason)
	}
}

func TestRouteFeedbackBiasCanImproveLocalRouteScore(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		PrefersCheapest: true,
	}
	base := scoreRouteMode("local-workhorse", features)
	for i := 0; i < 3; i++ {
		if err := recordRouteFeedback("local-workhorse", "upvoted"); err != nil {
			t.Fatalf("record local feedback: %v", err)
		}
	}
	boosted := scoreRouteMode("local-workhorse", features)
	if boosted <= base {
		t.Fatalf("expected route feedback to boost local route score, base=%d boosted=%d", base, boosted)
	}
}

func TestRouteFeedbackBiasIsScopedByDeviceAndModelFingerprint(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	if err := recordRouteFeedback("local-workhorse", "upvoted"); err != nil {
		t.Fatalf("record feedback: %v", err)
	}
	if got := routeFeedbackBias("local-workhorse"); got <= 0 {
		t.Fatalf("expected matching fingerprint to receive positive bias, got %d", got)
	}

	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "tiny")
	if got := routeFeedbackBias("local-workhorse"); got != 0 {
		t.Fatalf("expected different device fingerprint to avoid inherited bias, got %d", got)
	}
}

func TestBenchmarkLocalRuntimeCapabilitiesRecordsMeasuredResults(t *testing.T) {
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_LOCAL_SLM_DEEP_MODEL", "qwen3:14b")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		body := mustReadAll(t, req.Body)
		switch {
		case strings.Contains(body, `"model":"phi4-mini"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"READY FAST ALPHA BETA GAMMA DELTA ETA THETA IOTA"}}],"usage":{"completion_tokens":18}}`), nil
		case strings.Contains(body, `"model":"qwen3:8b-instruct"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"{\"status\":\"ready\",\"items\":[\"one\",\"two\",\"three\",\"four\",\"five\",\"six\"]}"}}],"usage":{"completion_tokens":28}}`), nil
		case strings.Contains(body, `"model":"qwen3:14b"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"1. first step now\n2. second step next\n3. third step here\n4. fourth step later\n5. fifth step done"}}],"usage":{"completion_tokens":24}}`), nil
		case strings.Contains(body, `"model":"gemma3:12b"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"READY MULTIMODAL VISION AUDIO DOCUMENT IMAGE PARSE"}}],"usage":{"completion_tokens":16}}`), nil
		default:
			t.Fatalf("unexpected benchmark body:\n%s", body)
			return nil, nil
		}
	})

	report := benchmarkLocalRuntimeCapabilities(context.Background())
	for _, mode := range []string{"local-fast", "local-workhorse", "local-deep", "local-multimodal"} {
		row, ok := report.Adapters[mode]
		if !ok {
			t.Fatalf("expected adapter row for %s", mode)
		}
		if row.Status != "ok" {
			t.Fatalf("expected ok benchmark for %s, got %#v", mode, row)
		}
		if row.QualityClass != "strong" {
			t.Fatalf("expected strong benchmark quality for %s, got %#v", mode, row)
		}
		if row.StructuredReliability != "strong" {
			t.Fatalf("expected strong benchmark reliability for %s, got %#v", mode, row)
		}
		if row.OutputTokens <= 0 {
			t.Fatalf("expected output tokens for %s, got %#v", mode, row)
		}
		if row.TokensPerSecond <= 0 {
			t.Fatalf("expected tokens/sec for %s, got %#v", mode, row)
		}
	}
}

func TestLocalBenchmarkBiasUsesSavedMeasuredResults(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             900,
				ColdStartCostMS:       220,
				OutputTokens:          28,
				TokensPerSecond:       31.1,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save capabilities: %v", err)
	}
	if got := localBenchmarkBias("local-workhorse"); got <= 0 {
		t.Fatalf("expected positive measured benchmark bias, got %d", got)
	}
}

func TestSaveLocalRuntimeCapabilitiesAppendsRollingHistory(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	reportOne := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             900,
				TokensPerSecond:       24,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339),
			},
		},
	}
	reportTwo := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             780,
				TokensPerSecond:       28,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(reportOne); err != nil {
		t.Fatalf("save report one: %v", err)
	}
	if err := saveLocalRuntimeCapabilities(reportTwo); err != nil {
		t.Fatalf("save report two: %v", err)
	}
	loaded := loadLocalRuntimeCapabilities()
	history := loaded.History["local-workhorse"]
	if len(history) != 2 {
		t.Fatalf("expected two history samples, got %#v", history)
	}
	if history[1].LatencyMS != 780 {
		t.Fatalf("expected latest history sample to be preserved, got %#v", history[1])
	}
}

func TestLocalBenchmarkBiasPenalizesRegressionTrend(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             3600,
				ColdStartCostMS:       2200,
				TokensPerSecond:       3.4,
				QualityClass:          "usable",
				StructuredReliability: "weak",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		History: map[string][]localAdapterSample{
			"local-workhorse": {
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 850, TokensPerSecond: 25, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Add(-3 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 820, TokensPerSecond: 26, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Add(-2 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 780, TokensPerSecond: 27, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Add(-1 * time.Minute).Format(time.RFC3339)},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}
	if got := localBenchmarkBias("local-workhorse"); got >= 0 {
		t.Fatalf("expected negative bias for regression trend, got %d", got)
	}
}

func TestLocalBenchmarkHistoryBiasKeepsSingleRegressionConservative(t *testing.T) {
	row := localAdapterCapability{
		AdapterID:             "local-workhorse",
		ModelID:               "qwen3:8b",
		Status:                "ok",
		LatencyMS:             2600,
		TokensPerSecond:       8.5,
		QualityClass:          "usable",
		StructuredReliability: "usable",
		BenchmarkedAt:         time.Now().UTC().Format(time.RFC3339),
	}
	history := []localAdapterSample{
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 820, TokensPerSecond: 26, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: time.Now().UTC().Add(-3 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 790, TokensPerSecond: 27, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: row.LatencyMS, TokensPerSecond: row.TokensPerSecond, QualityClass: row.QualityClass, StructuredReliability: row.StructuredReliability, BenchmarkedAt: row.BenchmarkedAt},
	}
	got := localBenchmarkHistoryBias(row, history)
	if got >= 0 || got < -8 {
		t.Fatalf("expected conservative single-regression penalty, got %d", got)
	}
}

func TestLocalBenchmarkHistoryBiasRaisesConfidenceForRepeatedRegression(t *testing.T) {
	row := localAdapterCapability{
		AdapterID:             "local-workhorse",
		ModelID:               "qwen3:8b",
		Status:                "ok",
		LatencyMS:             3400,
		TokensPerSecond:       4.2,
		QualityClass:          "usable",
		StructuredReliability: "weak",
		BenchmarkedAt:         time.Now().UTC().Format(time.RFC3339),
	}
	history := []localAdapterSample{
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 880, TokensPerSecond: 25, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: time.Now().UTC().Add(-4 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2400, TokensPerSecond: 9.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: time.Now().UTC().Add(-3 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2900, TokensPerSecond: 6.0, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: row.LatencyMS, TokensPerSecond: row.TokensPerSecond, QualityClass: row.QualityClass, StructuredReliability: row.StructuredReliability, BenchmarkedAt: row.BenchmarkedAt},
	}
	got := localBenchmarkHistoryBias(row, history)
	if got > -10 {
		t.Fatalf("expected stronger repeated-regression penalty, got %d", got)
	}
}

func TestLocalBenchmarkHistoryBiasDecaysStaleRegression(t *testing.T) {
	now := time.Now().UTC()
	row := localAdapterCapability{
		AdapterID:             "local-workhorse",
		ModelID:               "qwen3:8b",
		Status:                "ok",
		LatencyMS:             3200,
		TokensPerSecond:       4.8,
		QualityClass:          "usable",
		StructuredReliability: "weak",
		BenchmarkedAt:         now.Format(time.RFC3339),
	}
	freshHistory := []localAdapterSample{
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2500, TokensPerSecond: 9.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-45 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2800, TokensPerSecond: 6.2, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-20 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: row.LatencyMS, TokensPerSecond: row.TokensPerSecond, QualityClass: row.QualityClass, StructuredReliability: row.StructuredReliability, BenchmarkedAt: row.BenchmarkedAt},
	}
	staleHistory := []localAdapterSample{
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2500, TokensPerSecond: 9.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-23 * time.Hour).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2800, TokensPerSecond: 6.2, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-20 * time.Hour).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: row.LatencyMS, TokensPerSecond: row.TokensPerSecond, QualityClass: row.QualityClass, StructuredReliability: row.StructuredReliability, BenchmarkedAt: row.BenchmarkedAt},
	}
	freshPenalty := localBenchmarkHistoryBias(row, freshHistory)
	stalePenalty := localBenchmarkHistoryBias(row, staleHistory)
	if freshPenalty >= stalePenalty {
		t.Fatalf("expected stale regression to count less than fresh regression, got fresh=%d stale=%d", freshPenalty, stalePenalty)
	}
	if stalePenalty >= 0 {
		t.Fatalf("expected stale regression to still carry some penalty, got %d", stalePenalty)
	}
}

func TestLocalBenchmarkHistoryBiasRewardsRecoveryAfterRegression(t *testing.T) {
	now := time.Now().UTC()
	row := localAdapterCapability{
		AdapterID:             "local-workhorse",
		ModelID:               "qwen3:8b",
		Status:                "ok",
		LatencyMS:             820,
		TokensPerSecond:       26.5,
		QualityClass:          "strong",
		StructuredReliability: "strong",
		BenchmarkedAt:         now.Format(time.RFC3339),
	}
	history := []localAdapterSample{
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2600, TokensPerSecond: 8.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-40 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2400, TokensPerSecond: 9.4, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-25 * time.Minute).Format(time.RFC3339)},
		{ModelID: "qwen3:8b", Status: "ok", LatencyMS: row.LatencyMS, TokensPerSecond: row.TokensPerSecond, QualityClass: row.QualityClass, StructuredReliability: row.StructuredReliability, BenchmarkedAt: row.BenchmarkedAt},
	}
	got := localBenchmarkHistoryBias(row, history)
	if got <= 0 {
		t.Fatalf("expected positive recovery bias, got %d", got)
	}
	if trend := localBenchmarkTrend(row, history); trend != "recovered" {
		t.Fatalf("expected recovered trend label, got %q", trend)
	}
}

func TestLocalBenchmarkRecoveryBiasIsScopedByWorkClass(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             820,
				TokensPerSecond:       26.5,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		History: map[string][]localAdapterSample{
			"local-workhorse": {
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2600, TokensPerSecond: 8.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-40 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2400, TokensPerSecond: 9.4, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-25 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 820, TokensPerSecond: 26.5, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Format(time.RFC3339)},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}

	planningBias := localBenchmarkBiasForFeatures("local-workhorse", routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		EffortClass:     "medium",
		PrefersCheapest: true,
	})
	codeBias := localBenchmarkBiasForFeatures("local-workhorse", routeFeatures{
		WorkClass:       "code",
		DepthClass:      "standard",
		ModalityClass:   "text",
		EffortClass:     "medium",
		PrefersCheapest: true,
	})
	if planningBias <= codeBias {
		t.Fatalf("expected planning work to receive more recovered workhorse bias than code work, planning=%d code=%d", planningBias, codeBias)
	}
}

func TestLocalBenchmarkRecoveryBiasIsScopedByPlanningSubtype(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             820,
				TokensPerSecond:       26.5,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		History: map[string][]localAdapterSample{
			"local-workhorse": {
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2600, TokensPerSecond: 8.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-40 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2400, TokensPerSecond: 9.4, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-25 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 820, TokensPerSecond: 26.5, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Format(time.RFC3339)},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}

	readinessBias := localBenchmarkBiasForFeatures("local-workhorse", routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "build-readiness",
		ArtifactFamily:  "structured-check",
		EffortClass:     "medium",
		PrefersCheapest: true,
	})
	travelBias := localBenchmarkBiasForFeatures("local-workhorse", routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "trip-itinerary",
		ArtifactFamily:  "itinerary-plan",
		EffortClass:     "medium",
		PrefersCheapest: true,
	})
	if readinessBias <= travelBias {
		t.Fatalf("expected recovered workhorse benchmark to help build-readiness more than trip-itinerary, readiness=%d travel=%d", readinessBias, travelBias)
	}
}

func TestLocalBenchmarkBiasUsesDirectCohortEvidenceForTripItinerary(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             820,
				TokensPerSecond:       26.5,
				QualityClass:          "strong",
				StructuredReliability: "strong",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		History: map[string][]localAdapterSample{
			"local-workhorse": {
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2600, TokensPerSecond: 8.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-40 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2400, TokensPerSecond: 9.4, QualityClass: "usable", StructuredReliability: "weak", BenchmarkedAt: now.Add(-25 * time.Minute).Format(time.RFC3339)},
				{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 820, TokensPerSecond: 26.5, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Format(time.RFC3339)},
			},
		},
		CohortHistory: map[string]map[string][]localAdapterSample{
			"local-workhorse": {
				"trip-itinerary": {
					{ModelID: "qwen3:8b", Status: "degraded", LatencyMS: 2200, TokensPerSecond: 9.0, QualityClass: "usable", StructuredReliability: "usable", BenchmarkedAt: now.Add(-35 * time.Minute).Format(time.RFC3339)},
					{ModelID: "qwen3:8b", Status: "ok", LatencyMS: 900, TokensPerSecond: 22.0, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Format(time.RFC3339)},
				},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}

	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "trip-itinerary",
		ArtifactFamily:  "itinerary-plan",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	transferOnly := int(float64(localBenchmarkBias("local-workhorse")) * localBenchmarkScopeWeight("local-workhorse", features))
	got := localBenchmarkBiasForFeatures("local-workhorse", features)
	if got <= transferOnly {
		t.Fatalf("expected direct trip-itinerary cohort evidence to improve bias beyond transferred adapter evidence, got=%d transferOnly=%d", got, transferOnly)
	}
}

func TestLocalBenchmarkBiasKeepsMultimodalSubtypeEvidenceScoped(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-multimodal": {
				AdapterID:             "local-multimodal",
				ModelID:               "gemma3:12b",
				Status:                "ok",
				LatencyMS:             1400,
				TokensPerSecond:       12.0,
				QualityClass:          "usable",
				StructuredReliability: "usable",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		CohortHistory: map[string]map[string][]localAdapterSample{
			"local-multimodal": {
				"multimodal-image-screenshot": {
					{ModelID: "gemma3:12b", Status: "ok", LatencyMS: 1200, TokensPerSecond: 14.0, QualityClass: "strong", StructuredReliability: "strong", BenchmarkedAt: now.Format(time.RFC3339)},
				},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}

	screenshotFeatures := routeFeatures{
		WorkClass:       "general",
		DepthClass:      "standard",
		ModalityClass:   "multimodal",
		ModalitySubtype: "image-screenshot",
		RequestCohort:   "multimodal-extract",
		ArtifactFamily:  "multimodal-extract",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	pdfFeatures := routeFeatures{
		WorkClass:       "general",
		DepthClass:      "standard",
		ModalityClass:   "multimodal",
		ModalitySubtype: "pdf-scan",
		RequestCohort:   "multimodal-extract",
		ArtifactFamily:  "multimodal-extract",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}

	screenshotBias := localBenchmarkBiasForFeatures("local-multimodal", screenshotFeatures)
	pdfBias := localBenchmarkBiasForFeatures("local-multimodal", pdfFeatures)
	if screenshotBias <= pdfBias {
		t.Fatalf("expected screenshot subtype evidence to stay scoped and beat unrelated pdf transfer, screenshot=%d pdf=%d", screenshotBias, pdfBias)
	}
}

func TestLocalBenchmarkBiasUsesCohortFeedbackForTripItinerary(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             1200,
				TokensPerSecond:       14.0,
				QualityClass:          "usable",
				StructuredReliability: "usable",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"trip-itinerary": {Upvotes: 2},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}
	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "trip-itinerary",
		ArtifactFamily:  "itinerary-plan",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	base := localBenchmarkSampleBias(report.Adapters["local-workhorse"]) + localBenchmarkHistoryBias(report.Adapters["local-workhorse"], nil)
	got := localBenchmarkBiasForFeatures("local-workhorse", features)
	if got <= base {
		t.Fatalf("expected cohort feedback to improve trip-itinerary bias, got=%d base=%d", got, base)
	}
}

func TestLocalBenchmarkBiasUsesGradedCohortFeedbackForTripItinerary(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             1200,
				TokensPerSecond:       14.0,
				QualityClass:          "usable",
				StructuredReliability: "usable",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"trip-itinerary": {AcceptedAsIs: 1, NeededLightEdits: 1},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}
	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "trip-itinerary",
		ArtifactFamily:  "itinerary-plan",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	withoutFeedback := report
	withoutFeedback.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	if err := saveLocalRuntimeCapabilities(withoutFeedback); err != nil {
		t.Fatalf("save report without feedback: %v", err)
	}
	base := localBenchmarkBiasForFeatures("local-workhorse", features)
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("restore report with feedback: %v", err)
	}
	got := localBenchmarkBiasForFeatures("local-workhorse", features)
	if got <= base {
		t.Fatalf("expected graded cohort feedback to improve trip-itinerary bias, got=%d base=%d", got, base)
	}
}

func TestLocalBenchmarkBiasPenalizesPassiveHeavyEditsForTripItinerary(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             1200,
				TokensPerSecond:       14.0,
				QualityClass:          "usable",
				StructuredReliability: "usable",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"trip-itinerary": {PassiveNeededHeavyEdits: 1},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}
	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "trip-itinerary",
		ArtifactFamily:  "itinerary-plan",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	withPassive := localBenchmarkBiasForFeatures("local-workhorse", features)
	report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save clean report: %v", err)
	}
	withoutPassive := localBenchmarkBiasForFeatures("local-workhorse", features)
	if withPassive >= withoutPassive {
		t.Fatalf("expected passive heavy edits to reduce trip-itinerary bias, with=%d without=%d", withPassive, withoutPassive)
	}
}

func TestLocalBenchmarkBiasUsesOutcomeAdoptionForSendableFollowup(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	now := time.Now().UTC()
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                now.Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters: map[string]localAdapterCapability{
			"local-workhorse": {
				AdapterID:             "local-workhorse",
				ModelID:               "qwen3:8b",
				Status:                "ok",
				LatencyMS:             1200,
				TokensPerSecond:       14.0,
				QualityClass:          "usable",
				StructuredReliability: "usable",
				BenchmarkedAt:         now.Format(time.RFC3339),
			},
		},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"sendable-followup": {OutcomeUsed: 1, OutcomeShared: 1},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save report: %v", err)
	}
	features := routeFeatures{
		WorkClass:       "planning",
		DepthClass:      "standard",
		ModalityClass:   "text",
		RequestCohort:   "sendable-followup",
		ArtifactFamily:  "narrative-draft",
		EffortClass:     "medium",
		PrefersCheapest: true,
	}
	withOutcome := localBenchmarkBiasForFeatures("local-workhorse", features)
	report.CohortFeedback = map[string]map[string]localCohortFeedbackRow{}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save clean report: %v", err)
	}
	withoutOutcome := localBenchmarkBiasForFeatures("local-workhorse", features)
	if withOutcome <= withoutOutcome {
		t.Fatalf("expected adoption outcome to improve sendable-followup bias, with=%d without=%d", withOutcome, withoutOutcome)
	}
}

func TestClassifyArtifactEditSignalTreatsTitleOnlyChangeAsHeaderOnly(t *testing.T) {
	original := "# Sendable Follow-Up: Weekly Review\n\n## Send this note\nDraft body.\n"
	current := "# Sendable Follow-Up: Weekly Product Review\n\n## Send this note\nDraft body.\n"
	editClass, editScope, semanticClass := classifyArtifactEditSignal(original, current)
	if editClass != "light" || editScope != "header-only" {
		t.Fatalf("expected light/header-only edit signal, got class=%q scope=%q", editClass, editScope)
	}
	if semanticClass != "header-only" {
		t.Fatalf("expected header-only semantic class, got %q", semanticClass)
	}
}

func TestClassifyArtifactEditSignalTreatsCoreRewriteAsCoreSections(t *testing.T) {
	original := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- Intent is clear.\n\n## Must clear before build\n- Rollback owner.\n"
	current := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- This is now a different direction entirely with multiple changed assumptions and constraints.\n\n## Must clear before build\n- Rollback owner.\n"
	editClass, editScope, semanticClass := classifyArtifactEditSignal(original, current)
	if editScope != "core-sections" {
		t.Fatalf("expected core-sections scope, got %q", editScope)
	}
	if editClass == "none" {
		t.Fatalf("expected non-empty edit class for core rewrite")
	}
	if semanticClass != "core-decision-change" {
		t.Fatalf("expected core-decision-change semantic class, got %q", semanticClass)
	}
}

func TestClassifyArtifactEditSignalTreatsCoreWordingEditSeparately(t *testing.T) {
	original := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- Intent is clear and scoped.\n- Rollback path is listed.\n"
	current := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- Intent is clear, scoped, and ready for implementation.\n- Rollback path is listed for the team.\n"
	editClass, editScope, semanticClass := classifyArtifactEditSignal(original, current)
	if editScope != "core-sections" {
		t.Fatalf("expected core-sections scope, got %q", editScope)
	}
	if editClass == "none" {
		t.Fatalf("expected a non-empty edit class for core wording change")
	}
	if semanticClass != "core-wording" {
		t.Fatalf("expected core-wording semantic class, got %q", semanticClass)
	}
}

func TestMaybeWarmLocalRuntimeCapabilitiesAsyncWritesFreshReport(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_LOCAL_SLM_DEEP_MODEL", "qwen3:14b")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		body := mustReadAll(t, req.Body)
		switch {
		case strings.Contains(body, `"model":"phi4-mini"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"READY FAST ALPHA BETA GAMMA DELTA ETA THETA IOTA"}}],"usage":{"completion_tokens":18}}`), nil
		case strings.Contains(body, `"model":"qwen3:8b-instruct"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"{\"status\":\"ready\",\"items\":[\"one\",\"two\",\"three\",\"four\",\"five\",\"six\"]}"}}],"usage":{"completion_tokens":28}}`), nil
		case strings.Contains(body, `"model":"qwen3:14b"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"1. first step now\n2. second step next\n3. third step here\n4. fourth step later\n5. fifth step done"}}],"usage":{"completion_tokens":24}}`), nil
		case strings.Contains(body, `"model":"gemma3:12b"`):
			return jsonResponse(200, `{"choices":[{"message":{"content":"READY MULTIMODAL VISION AUDIO DOCUMENT IMAGE PARSE"}}],"usage":{"completion_tokens":16}}`), nil
		default:
			t.Fatalf("unexpected warm benchmark body:\n%s", body)
			return nil, nil
		}
	})

	if scheduled := maybeWarmLocalRuntimeCapabilitiesAsync(); !scheduled {
		t.Fatalf("expected async warm to schedule on first local availability")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		report := loadLocalRuntimeCapabilities()
		if localRuntimeCapabilitiesAreFresh(report) && report.Adapters["local-workhorse"].Status == "ok" {
			if scheduled := maybeWarmLocalRuntimeCapabilitiesAsync(); scheduled {
				t.Fatalf("expected no second async warm once report is fresh")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected async warm to write a fresh local runtime report")
}

func TestDetectRouteForRequestAutoPrefersAzureWritingRouteForTravelPlanning(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Plan 7 day Paris trip with a clear itinerary and budget.",
	})
	if decision.ToolMode != "chatgpt" {
		t.Fatalf("expected Azure writing route for travel planning, got %#v", decision)
	}
	if decision.ToolLabel != "Azure writing route" {
		t.Fatalf("expected Azure writing route label for travel planning, got %#v", decision)
	}
	if decision.ModelLabel != "gpt-4o-prod" {
		t.Fatalf("expected Azure deployment model label for travel planning, got %#v", decision)
	}
	if decision.EffortLevel != "medium" {
		t.Fatalf("expected medium effort for normal travel planning, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "planning work") || !strings.Contains(decision.Reason, "cheapest suitable") {
		t.Fatalf("expected planning reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersAzureCodeRouteForCodeWork(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "CLI test failure",
		Source: "Fix failing CLI tests in this repo and update the Go code path.",
	})
	if decision.ToolMode != "azure-code" {
		t.Fatalf("expected Azure code route for code work, got %#v", decision)
	}
	if decision.ToolLabel != "Azure code route" {
		t.Fatalf("expected Azure code route label for code work, got %#v", decision)
	}
	if decision.ModelLabel != "gpt-4o-prod" {
		t.Fatalf("expected Azure deployment model label for code work, got %#v", decision)
	}
	if decision.EffortLevel != "medium" {
		t.Fatalf("expected medium effort for normal code work, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "code-heavy work") || !strings.Contains(decision.Reason, "cheapest suitable") {
		t.Fatalf("expected code reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersBestPlanningToolForDeepWork(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "SECRETEXAMPLE")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Do a deep, rigorous Paris trip plan with benchmarks and comprehensive tradeoffs.",
	})
	if decision.ToolMode != "bedrock-sonnet" {
		t.Fatalf("expected Bedrock Sonnet route for deep planning work, got %#v", decision)
	}
	if decision.ModelLabel != "Claude Sonnet 4.6" {
		t.Fatalf("expected Bedrock model label for deep planning work, got %#v", decision)
	}
	if decision.EffortLevel != "extra high" {
		t.Fatalf("expected extra high effort for benchmark-heavy planning work, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "planning work") || !strings.Contains(decision.Reason, "strongest suitable route") {
		t.Fatalf("expected deep planning reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersBestCodeToolForDeepWork(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "CLI architecture review",
		Source: "Do deep work on the CLI architecture with rigorous critique and comprehensive code review.",
	})
	if decision.ToolMode != "claude-api" {
		t.Fatalf("expected Claude API route for deep code work, got %#v", decision)
	}
	if decision.ModelLabel != "Claude Sonnet 4" {
		t.Fatalf("expected Claude model label for deep code work, got %#v", decision)
	}
	if decision.EffortLevel != "extra high" {
		t.Fatalf("expected extra high effort for architecture review, got %#v", decision)
	}
	if !strings.Contains(decision.Reason, "code-heavy work") || !strings.Contains(decision.Reason, "strongest suitable route") {
		t.Fatalf("expected deep code reason, got %#v", decision)
	}
}

func TestDetectRouteForRequestAutoPrefersLowEffortForQuickPass(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Quick follow-up",
		Source: "Quick one-line follow-up from this meeting.",
	})
	if decision.EffortLevel != "low" {
		t.Fatalf("expected low effort for quick pass, got %#v", decision)
	}
	if decision.ModelLabel != "gpt-4o-prod" {
		t.Fatalf("expected Azure deployment model label for quick pass, got %#v", decision)
	}
	if !strings.Contains(decision.EffortReason, "low effort") {
		t.Fatalf("expected low effort reason, got %#v", decision)
	}
}

func TestProviderUserPromptShapesMeetingFollowupArtifact(t *testing.T) {
	prompt := providerUserPrompt(providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Weekly Product Review",
		Source: "Weekly product review for pricing launch. Need owners, due dates, and open questions.",
	})

	for _, want := range []string{
		"Return a sendable follow-up note first.",
		"`## Send this note`",
		"`## Decisions captured from the notes`",
		"`## Owners and due dates to confirm`",
		"`## Open questions to close`",
		"`## Recommended next move`",
		"Do not invent names, dates, or commitments",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected meeting prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestProviderUserPromptShapesBuildReadinessArtifact(t *testing.T) {
	prompt := providerUserPrompt(providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Notifications PRD",
		Source: "Notifications PRD needs a build-readiness check and handoff call.",
	})

	for _, want := range []string{
		"Return a build-readiness artifact, not a vague summary.",
		"`## What looks ready now`",
		"`## Must clear before build`",
		"`## Recommended first slice`",
		"`## Who needs to answer what`",
		"`## Still to confirm`",
		"Do not reduce the answer to a binary verdict",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected research prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestProviderUserPromptAddsHiddenPlanGuidanceForMultistepWork(t *testing.T) {
	prompt := providerUserPrompt(providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Plan 7 day Paris trip with a clear itinerary and budget.",
	})
	for _, want := range []string{
		"Internal planning guidance:",
		"make a short hidden plan for yourself",
		"Do not print the hidden plan unless the user explicitly asks for it.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected hidden planning prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestGenerateWithConfiguredProviderRunsSingleSelectiveRefineForWeakPlanningRoute(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")

	callCount := 0
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		body := mustReadAll(t, req.Body)
		if callCount == 1 {
			if !strings.Contains(body, "Internal planning guidance:") {
				t.Fatalf("expected first pass to include hidden planning guidance, got:\n%s", body)
			}
			return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Paris\n\n## Day by day\n- Draft stop one.\n\n## Still to confirm\n- Dates."}}]}`), nil
		}
		if !strings.Contains(body, "Refinement trigger:") || !strings.Contains(body, "Current draft:") {
			t.Fatalf("expected refine prompt on second pass, got:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Paris\n\n## Day by day\n- Draft stop one.\n- Draft stop two.\n\n## Still to confirm\n- Dates.\n- Hotel area."}}]}`), nil
	})

	result, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "travel-plan"},
		Title:  "7-Day Paris Trip",
		Source: "Plan 7 day Paris trip with a clear itinerary and budget.",
	})
	if err != nil {
		t.Fatalf("generate with selective refine: %v", err)
	}
	if !used {
		t.Fatalf("expected provider to be used")
	}
	if callCount != 2 {
		t.Fatalf("expected exactly two provider calls, got %d", callCount)
	}
	if !strings.Contains(result, "Draft stop two.") || !strings.Contains(result, "Hotel area.") {
		t.Fatalf("expected refined artifact to be returned, got:\n%s", result)
	}
}

func TestGenerateWithConfiguredProviderSkipsSelectiveRefineForLowEffortQuickPass(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "tiny")

	callCount := 0
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		body := mustReadAll(t, req.Body)
		if strings.Contains(body, "Refinement trigger:") {
			t.Fatalf("did not expect refine prompt for low-effort pass, got:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nQuick draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	_, used, err := generateWithConfiguredProvider(context.Background(), providerGenerationRequest{
		Choice: starterChoice{PackID: "meeting-followup"},
		Title:  "Quick follow-up",
		Source: "Quick one-line follow-up from this meeting.",
	})
	if err != nil {
		t.Fatalf("generate low effort pass: %v", err)
	}
	if !used {
		t.Fatalf("expected provider to be used")
	}
	if callCount != 1 {
		t.Fatalf("expected single provider call for low-effort pass, got %d", callCount)
	}
}

func TestGenerateWithConfiguredProviderRunsConsistencyCheckForExtraHighWork(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	callCount := 0
	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		body := mustReadAll(t, req.Body)
		if !strings.Contains(body, "architecture") {
			t.Fatalf("expected architecture request in body, got:\n%s", body)
		}
		if callCount == 1 {
			if strings.Contains(body, "Consistency-check trigger:") {
				t.Fatalf("did not expect first pass to be the consistency prompt, got:\n%s", body)
			}
			return jsonResponse(200, `{"choices":[{"message":{"content":"## What looks ready now\n- Core plan exists.\n\n## Still to confirm\n- Ownership."}}]}`), nil
		}
		if !strings.Contains(body, "Consistency-check trigger:") {
			t.Fatalf("expected consistency prompt on second pass, got:\n%s", body)
		}
		return jsonResponse(200, `{"choices":[{"message":{"content":"## What looks ready now\n- Core plan exists.\n- Architecture constraints are listed.\n\n## Must clear before build\n- Confirm ownership.\n- Confirm rollback path.\n\n## Recommended first slice\n- Validate the riskiest path first.\n\n## Who needs to answer what\n- Engineering manager: rollback owner.\n\n## Still to confirm\n- Approval path."}}]}`), nil
	})

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Architecture Review",
		Source: "Do a benchmark and architecture review for this production ready plan.",
	}
	decision := detectRouteForRequest(request)
	result, used, actualDecision, err := generateWithConfiguredProviderDecision(context.Background(), request, decision)
	if err != nil {
		t.Fatalf("generate with consistency check: %v", err)
	}
	if !used {
		t.Fatalf("expected provider to be used")
	}
	if callCount != 2 {
		t.Fatalf("expected exactly two provider calls for consistency check, got %d", callCount)
	}
	if actualDecision.VerificationLevel != "Consistency check" {
		t.Fatalf("expected consistency verification level, got %#v", actualDecision)
	}
	if !strings.Contains(result, "## Must clear before build") || !strings.Contains(result, "## Who needs to answer what") {
		t.Fatalf("expected stronger consistency winner to be returned, got:\n%s", result)
	}
}

func TestSelectConsistencyWinnerUsesCohortLearningForBuildReadiness(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters:                  map[string]localAdapterCapability{},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"build-readiness": {
					OutcomeReplaced:        3,
					PassiveDecisionChanges: 4,
				},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save local runtime capabilities: %v", err)
	}

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Architecture Review",
		Source: "Do a benchmark and architecture review for this production ready plan.",
	}
	plain := "## What looks ready now\n- Core plan exists.\n\n## Must clear before build\n- Confirm ownership."
	grounded := "## What looks ready now\n- Core plan exists.\n\n## Must clear before build\n- Confirm ownership.\n\n## Recommended first slice\n- Validate the riskiest path first.\n\n## Still to confirm\n- Approval path."
	winner := selectConsistencyWinner(request, routeDecision{}, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected cohort-aware scorer to prefer the grounded draft, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerPrefersRouteSpecificCohortMemory(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	decision := detectRouteForRequest(providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Architecture Review",
		Source: "Do a benchmark and architecture review for this production ready plan.",
	})
	stats := localRouteFeedbackStats{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniRouteFeedback",
		Routes:        map[string]routeFeedbackRow{},
		Cohorts: map[string]map[string]localCohortFeedbackRow{
			routeFeedbackKeyForDecision(decision): {
				"build-readiness": {
					OutcomeReplaced:        3,
					PassiveDecisionChanges: 4,
				},
			},
		},
	}
	if err := saveLocalRouteFeedbackStats(stats); err != nil {
		t.Fatalf("save route feedback stats: %v", err)
	}
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters:                  map[string]localAdapterCapability{},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"build-readiness": {
					AcceptedAsIs:  4,
					OutcomeShared: 2,
				},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save local runtime capabilities: %v", err)
	}

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "research-prd"},
		Title:  "Architecture Review",
		Source: "Do a benchmark and architecture review for this production ready plan.",
	}
	plain := "## What looks ready now\n- Core plan exists.\n\n## Must clear before build\n- Confirm ownership."
	grounded := "## What looks ready now\n- Core plan exists.\n\n## Must clear before build\n- Confirm ownership.\n\n## Recommended first slice\n- Validate the riskiest path first.\n\n## Still to confirm\n- Approval path."
	winner := selectConsistencyWinner(request, decision, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected route-specific cohort memory to override generic fallback, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerUsesMultimodalEvidenceSections(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "PDF Extract",
		Source: "Review this PDF and screenshot and extract the evidence.",
	}
	plain := "## Summary\n- Main point.\n"
	grounded := "## Extracted evidence\n- The PDF shows the signed approval block.\n\n## What the source shows\n- The screenshot confirms the same owner and date.\n\n## Still unclear\n- OCR confidence is weak on the final line.\n\n## Recommended next move\n- Verify the last label manually."
	winner := selectConsistencyWinner(request, routeDecision{}, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected multimodal-aware scorer to prefer the evidence-grounded draft, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerUsesRouteSpecificMultimodalCohortMemory(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "PDF Extract",
		Source: "Review this PDF and screenshot and extract the evidence.",
	}
	decision := detectRouteForRequest(request)
	stats := localRouteFeedbackStats{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniRouteFeedback",
		Routes:        map[string]routeFeedbackRow{},
		Cohorts: map[string]map[string]localCohortFeedbackRow{
			routeFeedbackKeyForDecision(decision): {
				"multimodal-pdf-scan": {
					OutcomeReplaced:        3,
					PassiveDecisionChanges: 4,
				},
			},
		},
	}
	if err := saveLocalRouteFeedbackStats(stats); err != nil {
		t.Fatalf("save route feedback stats: %v", err)
	}
	report := localRuntimeCapabilities{
		SchemaVersion:             localRuntimeCapabilitiesSchemaVersion,
		ContextType:               "JiniLocalRuntimeCapabilities",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		DeviceProbeFingerprint:    currentProbeFingerprint(),
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		Adapters:                  map[string]localAdapterCapability{},
		CohortFeedback: map[string]map[string]localCohortFeedbackRow{
			"local-workhorse": {
				"multimodal-pdf-scan": {
					AcceptedAsIs:  4,
					OutcomeShared: 2,
				},
			},
		},
	}
	if err := saveLocalRuntimeCapabilities(report); err != nil {
		t.Fatalf("save local runtime capabilities: %v", err)
	}

	plain := "## Summary\n- The source looks fine."
	grounded := "## Extracted evidence\n- The PDF shows the signed approval block.\n\n## What the source shows\n- The screenshot confirms the same owner and date.\n\n## Still unclear\n- OCR confidence is weak on the final line.\n\n## Recommended next move\n- Verify the last label manually."
	winner := selectConsistencyWinner(request, decision, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected route-specific multimodal cohort memory to prefer the evidence-grounded draft, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerDoesNotBorrowPDFRouteMemoryForScreenshot(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "UI Screenshot Review",
		Source: "Review this screenshot and extract the evidence from the UI.",
	}
	decision := detectRouteForRequest(request)
	stats := localRouteFeedbackStats{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniRouteFeedback",
		Routes:        map[string]routeFeedbackRow{},
		Cohorts: map[string]map[string]localCohortFeedbackRow{
			routeFeedbackKeyForDecision(decision): {
				"multimodal-pdf-scan": {
					OutcomeReplaced:        4,
					PassiveDecisionChanges: 5,
				},
			},
		},
	}
	if err := saveLocalRouteFeedbackStats(stats); err != nil {
		t.Fatalf("save route feedback stats: %v", err)
	}

	plain := "## Summary\n- The screen looks fine."
	grounded := "## Extracted evidence\n- The screenshot shows the warning banner and disabled publish button.\n\n## What is visible\n- The right panel still lists an unresolved approval.\n\n## Still unclear\n- The screenshot does not show who owns the approval.\n\n## Recommended next move\n- Confirm the approval owner before publishing."
	winner := selectConsistencyWinner(request, decision, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected screenshot evidence rubric to ignore PDF route memory bleed, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerUsesPDFSpecificEvidenceSections(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Invoice Review",
		Source: "Review this scanned PDF invoice and extract the evidence.",
	}
	plain := "## Extracted evidence\n- The document appears to be an invoice.\n\n## What the source shows\n- Some fields are visible."
	grounded := "## Extracted evidence\n- The PDF shows invoice number INV-44 and a vendor signature.\n\n## What the document shows\n- Page 1 lists the total and due date in the summary table.\n\n## Still unclear\n- OCR confidence is weak on the tax line.\n\n## Recommended next move\n- Verify the tax field manually.\n\n## OCR or confidence notes\n- The lower-right stamp is partially obscured."
	winner := selectConsistencyWinner(request, routeDecision{}, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected pdf-aware scorer to prefer the document-grounded draft, got:\n%s", winner)
	}
}

func TestSelectConsistencyWinnerUsesAudioSpecificEvidenceSections(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())

	request := providerGenerationRequest{
		Choice: starterChoice{PackID: "general-work"},
		Title:  "Call Notes",
		Source: "Review this audio recording and transcript and extract the evidence.",
	}
	plain := "## Extracted evidence\n- There was a discussion.\n\n## What the source shows\n- The recording is about a project."
	grounded := "## Extracted evidence\n- The transcript shows the PM asked for a Friday cutoff.\n\n## What the recording says\n- The speaker said the approval still depends on legal review.\n\n## Still unclear\n- One name is muffled near the end of the recording.\n\n## Recommended next move\n- Confirm the missing speaker name before sending notes.\n\n## Confidence notes\n- The final 20 seconds are noisy."
	winner := selectConsistencyWinner(request, routeDecision{}, plain, grounded)
	if winner != grounded {
		t.Fatalf("expected audio-aware scorer to prefer the recording-grounded draft, got:\n%s", winner)
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

func TestBootstrapStarterWorkAddsSmartDestinationLinks(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-preview")

	summary, err := bootstrapStarterWork(
		starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"},
		"Plan 7 day Paris trip. Louvre is a must-do; Montmartre is a must-do; Versailles is a must-do.",
		"quick",
		[]inputItem{{InputID: "request", Kind: "text", Title: "Your request", Status: "processed", Preview: "Plan 7 day Paris trip. Louvre is a must-do; Montmartre is a must-do; Versailles is a must-do.", OriginRef: "Plan 7 day Paris trip. Louvre is a must-do; Montmartre is a must-do; Versailles is a must-do."}},
	)
	if err != nil {
		t.Fatalf("bootstrap local starter: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(summary.Dir, "views", "itinerary.md"))
	if err != nil {
		t.Fatalf("read itinerary: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"[Louvre](https://www.louvre.fr/en)",
		"[Montmartre](https://parisjetaime.com/eng/article/montmartre-a043)",
		"[Versailles](https://en.chateauversailles.fr/)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected itinerary to contain %q, got:\n%s", want, text)
		}
	}
}

func TestMaybeWriteProviderFirstDraftAddsSmartDestinationLinks(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Paris\n\n## Day by day\n- Louvre in the morning, then Montmartre.\n- Versailles on day five.\n\n## Still to confirm\n- Dates."}}]}`), nil
	})

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "artifacts"), 0o755); err != nil {
		t.Fatal(err)
	}
	choice := starterChoice{PackID: "travel-plan"}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, "7-Day Paris Trip", "Plan 7 day Paris trip"); err != nil {
		t.Fatalf("write provider draft: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "views", "itinerary.md"))
	if err != nil {
		t.Fatalf("read itinerary: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"[Louvre](https://www.louvre.fr/en)",
		"[Montmartre](https://parisjetaime.com/eng/article/montmartre-a043)",
		"[Versailles](https://en.chateauversailles.fr/)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected itinerary to contain %q, got:\n%s", want, text)
		}
	}
}

func TestMaybeWriteProviderFirstDraftPreservesSourceReferenceLinks(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## What looks ready now\n- The Project brief is clear enough to start.\n\n## Must clear before build\n- Approval path.\n\n## Recommended first slice\n- Ship the riskiest path first.\n\n## Who needs to answer what\n- PM: acceptance criteria.\n\n## Still to confirm\n- Final owner."}}]}`), nil
	})

	workDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "artifacts"), 0o755); err != nil {
		t.Fatal(err)
	}
	source := "Use the [Project brief](https://example.com/brief) before building."
	choice := starterChoice{PackID: "research-prd"}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, "Weekly product work", source); err != nil {
		t.Fatalf("write provider draft: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "views", "prd.md"))
	if err != nil {
		t.Fatalf("read prd: %v", err)
	}
	if !strings.Contains(string(content), "[Project brief](https://example.com/brief)") {
		t.Fatalf("expected source reference link to be preserved, got:\n%s", string(content))
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

	summary, err := bootstrapStarterWork(starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"}, "Plan 7 day Paris trip", "quick", []inputItem{{InputID: "request", Kind: "text", Title: "Your request", Status: "processed", Preview: "Plan 7 day Paris trip"}})
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

func TestBootstrapStarterWorkPersistsChosenRouteForCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	summary, err := bootstrapStarterWork(starterChoice{PackID: "travel-plan", DefaultName: "Trip Plan", State: "decided"}, "Plan 7 day Paris trip", "quick", []inputItem{{InputID: "request", Kind: "text", Title: "Your request", Status: "processed", Preview: "Plan 7 day Paris trip"}})
	if err != nil {
		t.Fatalf("bootstrap with auto route: %v", err)
	}
	if !strings.Contains(summary.WorkingWith, "Azure writing route via Azure OpenAI / gpt-4o-prod") {
		t.Fatalf("expected saved route label in summary, got %#v", summary)
	}
	if summary.ModelLabel != "gpt-4o-prod" {
		t.Fatalf("expected saved model label in summary, got %#v", summary)
	}
	if summary.EffortLevel != "medium" {
		t.Fatalf("expected saved effort level in summary, got %#v", summary)
	}
	if summary.VerificationLevel != "Single pass" {
		t.Fatalf("expected saved verification level in summary, got %#v", summary)
	}

	routeSaved, err := os.ReadFile(filepath.Join(summary.Dir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	for _, want := range []string{
		`"auto_mode": {`,
		`"framework_switching": "auto"`,
		`"model_switching": "auto"`,
		`"speed_switching": "auto"`,
		`"user_approval_mode": "approval-gated"`,
	} {
		if !strings.Contains(string(routeSaved), want) {
			t.Fatalf("expected route auto mode receipt to contain %q, got:\n%s", want, string(routeSaved))
		}
	}
}

func TestInteractiveLauncherShowsDecisionCardBeforeFirstDraft(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("Plan 7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area.\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Result ready.",
		"Itinerary",
		"Provider Paris",
		"Saved:",
		"Next: `jini continue`, `jini open`, or `jini status`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Jini will start with",
		"Tool",
		"How chosen",
		"Want a different route?",
		"Claude",
		"Bedrock",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected first-run output not to expose early route card %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCheckShowsSavedDecisionExplanation(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Plan 7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected setup run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := Run([]string{"status"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Working with",
		"Azure writing route via Azure OpenAI / gpt-4o-prod (chosen automatically)",
		"Model",
		"gpt-4o-prod",
		"Why this model",
		"The deployment decides the actual Azure model.",
		"Effort level",
		"Medium",
		"Verification",
		"Single pass",
		"Why this verification",
		"How chosen",
		"Automatic",
		"Auto mode",
		"Frameworks: auto",
		"Models: auto",
		"Speed: auto",
		"Approvals: approval-gated",
		"Why this route",
		"planning work",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCurrentWorkCanSaveModelFeedback(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"# Itinerary: Provider Paris\n\n- Provider day one."}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Plan 7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected setup run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("upvote\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected model feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Saved model feedback: upvoted.") {
		t.Fatalf("expected model feedback confirmation, got:\n%s", stdout.String())
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	if strings.TrimSpace(packDir) == "" {
		t.Fatalf("expected current work pack_dir, got:\n%s", string(current))
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"model_feedback": "upvoted"`) {
		t.Fatalf("expected saved model feedback, got:\n%s", string(routeSaved))
	}
	feedbackSaved, err := os.ReadFile(localFeedbackPath())
	if err != nil {
		t.Fatalf("expected route feedback file: %v", err)
	}
	if !strings.Contains(string(feedbackSaved), `"local-workhorse"`) && !strings.Contains(string(feedbackSaved), `"chatgpt|`) {
		t.Fatalf("expected saved route feedback stats, got:\n%s", string(feedbackSaved))
	}
}

func TestCurrentWorkLocalModelFeedbackRecordsCohortFeedback(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("upvote\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local model feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.Upvotes != 1 || row.Downvotes != 0 {
		t.Fatalf("expected local cohort feedback to record upvote, got %#v", row)
	}
}

func TestCurrentWorkLocalModelFeedbackRecordsSubtypeSpecificRouteCohortFeedback(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Extracted evidence\n- The screenshot shows a failed approval state.\n\n## What is visible\n- The publish button is disabled and an approval warning banner is open.\n\n## Still unclear\n- The owner is not visible.\n\n## Recommended next move\n- Confirm the approval owner before retrying."}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Review this screenshot and extract the evidence from the UI.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local screenshot run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("upvote\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local model feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	if strings.TrimSpace(packDir) == "" {
		t.Fatalf("expected current work pack_dir, got:\n%s", string(current))
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	var routeState map[string]any
	if err := json.Unmarshal(routeSaved, &routeState); err != nil {
		t.Fatalf("expected route json: %v", err)
	}
	toolMode, _ := routeState["tool_mode"].(string)
	feedbackKey := routeFeedbackKeyForCurrentMode(toolMode)
	if strings.TrimSpace(feedbackKey) == "" {
		t.Fatalf("expected derived route feedback key for tool mode %q, got route:\n%s", toolMode, string(routeSaved))
	}

	stats := loadLocalRouteFeedbackStats()
	foundScreenshot := false
	for key, cohorts := range stats.Cohorts {
		row := cohorts["multimodal-image-screenshot"]
		if row.Upvotes == 1 && row.Downvotes == 0 {
			foundScreenshot = true
		}
		if pdfRow, ok := cohorts["multimodal-pdf-scan"]; ok && (pdfRow.Upvotes != 0 || pdfRow.Downvotes != 0 || pdfRow.AcceptedAsIs != 0 || pdfRow.OutcomeShared != 0) {
			t.Fatalf("expected no PDF route cohort feedback bleed, found %#v under %q", pdfRow, key)
		}
	}
	if !foundScreenshot {
		t.Fatalf("expected screenshot-scoped route cohort feedback to record upvote, got %#v", stats.Cohorts)
	}
}

func TestCurrentWorkLocalArtifactAcceptanceRecordsGradedCohortFeedback(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	if strings.TrimSpace(packDir) == "" {
		t.Fatalf("expected current work pack_dir, got:\n%s", string(current))
	}
	followupPath := filepath.Join(packDir, "views", "followup.md")
	original, err := os.ReadFile(followupPath)
	if err != nil {
		t.Fatalf("expected followup view: %v", err)
	}
	edited := string(original) + "\n\n- Confirm the ETA before sending.\n"
	if err := os.WriteFile(followupPath, []byte(edited), 0o644); err != nil {
		t.Fatalf("write edited followup: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("accept\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local artifact feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Saved artifact feedback: accept.") {
		t.Fatalf("expected artifact feedback confirmation, got:\n%s", stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.AcceptedAsIs != 1 || row.NeededLightEdits != 0 || row.NotUseful != 0 || row.PassiveNeededLightEdits != 1 || row.PassiveAcceptedAsIs != 0 || row.PassiveNeededHeavyEdits != 0 || row.PassiveHeaderOnlyEdits != 0 || row.PassiveCoreSectionEdits != 0 {
		t.Fatalf("expected local graded cohort feedback to record acceptance, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"artifact_edit_class": "light"`) || !strings.Contains(string(routeSaved), `"artifact_edit_scope": "supporting-sections"`) || !strings.Contains(string(routeSaved), `"artifact_feedback_reason": "accepted-after-light-edits"`) {
		t.Fatalf("expected route to record passive artifact feedback reason, got:\n%s", string(routeSaved))
	}
}

func TestCurrentWorkLocalArtifactAcceptanceTreatsHeaderEditAsCosmetic(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	if strings.TrimSpace(packDir) == "" {
		t.Fatalf("expected current work pack_dir, got:\n%s", string(current))
	}
	followupPath := filepath.Join(packDir, "views", "followup.md")
	controlled := "# Sendable Follow-Up: Meeting Follow-up\n\n## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next\n"
	if err := os.WriteFile(followupPath, []byte(controlled), 0o644); err != nil {
		t.Fatalf("write controlled followup: %v", err)
	}
	if err := saveArtifactFeedbackBaseline(packDir); err != nil {
		t.Fatalf("save baseline: %v", err)
	}
	lines := strings.Split(controlled, "\n")
	headingIndex := -1
	for index, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			headingIndex = index
			break
		}
	}
	if headingIndex < 0 {
		t.Fatalf("expected a markdown heading in followup view")
	}
	lines[headingIndex] = lines[headingIndex] + " Review"
	edited := strings.Join(lines, "\n")
	if err := os.WriteFile(followupPath, []byte(edited), 0o644); err != nil {
		t.Fatalf("write edited followup: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("accept\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local artifact feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveHeaderOnlyEdits != 1 || row.PassiveCoreSectionEdits != 0 || row.PassiveAcceptedAsIs != 1 {
		t.Fatalf("expected cosmetic header edit signal, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"artifact_edit_scope": "header-only"`) || !strings.Contains(string(routeSaved), `"artifact_feedback_reason": "accepted-without-edits"`) {
		t.Fatalf("expected route to record cosmetic edit scope, got:\n%s", string(routeSaved))
	}
}

func TestCurrentWorkLocalArtifactAcceptanceTracksCoreDecisionChange(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## What looks ready now\n- Intent is clear and scoped.\n\n## Must clear before build\n- Rollback owner.\n"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Notifications PRD needs a build-readiness check and handoff call.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	if strings.TrimSpace(packDir) == "" {
		t.Fatalf("expected current work pack_dir, got:\n%s", string(current))
	}
	prdPath := filepath.Join(packDir, "views", "prd.md")
	controlled := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- Intent is clear and scoped.\n\n## Must clear before build\n- Rollback owner.\n"
	if err := os.WriteFile(prdPath, []byte(controlled), 0o644); err != nil {
		t.Fatalf("write controlled prd: %v", err)
	}
	if err := saveArtifactFeedbackBaseline(packDir); err != nil {
		t.Fatalf("save baseline: %v", err)
	}
	rewritten := "# Build-Readiness Check: Notifications\n\n## What looks ready now\n- The team should pause implementation and switch to a phased rollout with a new approval gate.\n\n## Must clear before build\n- New rollback owner and product approval.\n"
	if err := os.WriteFile(prdPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("write rewritten prd: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("accept\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local artifact feedback run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["build-readiness"]
	if row.PassiveDecisionChanges != 1 || row.PassiveCoreWordingEdits != 0 {
		t.Fatalf("expected core decision change signal, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"artifact_semantic_class": "core-decision-change"`) || !strings.Contains(string(routeSaved), `"artifact_feedback_reason": "accepted-after-decision-change"`) {
		t.Fatalf("expected route to record decision-change semantic class, got:\n%s", string(routeSaved))
	}
}

func TestCurrentWorkLocalArtifactOutcomeRecordsSharedUse(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("share\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local artifact outcome run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Saved artifact outcome: share.") {
		t.Fatalf("expected artifact outcome confirmation, got:\n%s", stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.OutcomeShared != 1 || row.OutcomeUsed != 0 || row.OutcomeReplaced != 0 {
		t.Fatalf("expected shared artifact outcome signal, got %#v", row)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"artifact_outcome_signal": "shared-this"`) || !strings.Contains(string(routeSaved), `"artifact_outcome_reason": "shared-or-handed-off"`) {
		t.Fatalf("expected route to record artifact outcome, got:\n%s", string(routeSaved))
	}
}

func TestCurrentWorkLocalArtifactOutcomeRecordsSubtypeSpecificRouteCohortOutcome(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Extracted evidence\n- The screenshot shows the publish button disabled.\n\n## What is visible\n- The approval warning banner is open in the right panel.\n\n## Still unclear\n- The owner is not visible.\n\n## Recommended next move\n- Confirm the owner before retrying."}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Review this screenshot and extract the evidence from the UI.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local screenshot run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := RunInteractive(nil, strings.NewReader("share\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected screenshot artifact outcome run to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	stats := loadLocalRouteFeedbackStats()
	foundScreenshot := false
	for key, cohorts := range stats.Cohorts {
		row := cohorts["multimodal-image-screenshot"]
		if row.OutcomeShared == 1 && row.OutcomeUsed == 0 && row.OutcomeReplaced == 0 {
			foundScreenshot = true
		}
		if pdfRow, ok := cohorts["multimodal-pdf-scan"]; ok && (pdfRow.OutcomeShared != 0 || pdfRow.OutcomeUsed != 0 || pdfRow.OutcomeReplaced != 0) {
			t.Fatalf("expected no PDF route cohort outcome bleed, found %#v under %q", pdfRow, key)
		}
	}
	if !foundScreenshot {
		t.Fatalf("expected screenshot-scoped route cohort outcome to record shared artifact, got %#v", stats.Cohorts)
	}
}

func TestRunOpenPassiveLocalArtifactExportSignalsSharedIntent(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	exportPath := filepath.Join(packDir, "exports", "wiki", "markdown", "overview.md")
	if err := os.MkdirAll(filepath.Dir(exportPath), 0o755); err != nil {
		t.Fatalf("create export dir: %v", err)
	}
	if err := os.WriteFile(exportPath, []byte("# Shared follow-up\n"), 0o644); err != nil {
		t.Fatalf("write export: %v", err)
	}

	var stdout bytes.Buffer
	if exitCode := Run([]string{"open", "markdown"}, &stdout, &stdout); exitCode != 0 {
		t.Fatalf("expected open markdown export to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveExportOpened != 1 {
		t.Fatalf("expected passive export-open signal, got %#v", row)
	}

	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"passive_artifact_outcome_signal": "shared-this"`) || !strings.Contains(string(routeSaved), `"passive_artifact_outcome_reason": "connector-markdown-opened-export-artifact"`) {
		t.Fatalf("expected passive route outcome to record export open, got:\n%s", string(routeSaved))
	}
}

func TestRunOpenPassiveLocalArtifactReopenSignalsUsefulness(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	for i := 0; i < 2; i++ {
		var stdout bytes.Buffer
		if exitCode := Run([]string{"open", "Sendable Follow-Up"}, &stdout, &stdout); exitCode != 0 {
			t.Fatalf("expected open sendable follow-up to succeed, got %d with output:\n%s", exitCode, stdout.String())
		}
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveReopened != 1 {
		t.Fatalf("expected passive reopen signal, got %#v", row)
	}
}

func TestRunOpenPassiveLocalArtifactDetectsSubstantiveReplacement(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	viewPath := filepath.Join(packDir, "views", "followup.md")
	rewritten := "## Send this note\nDiscard the original follow-up and draft a launch hold notice for legal review only.\n\n## Decisions captured from the notes\n- Freeze the rollout.\n- Cancel the customer handoff.\n\n## Owners and due dates to confirm\n- Legal lead by Tuesday.\n- Ops approver before any resend.\n\n## Open questions to close\n- Whether the incident summary replaces the meeting recap.\n- Whether support needs a separate notice.\n\n## Recommended next move\n- Replace the original artifact with a hold notice before anyone sends it."
	if err := os.WriteFile(viewPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("rewrite view: %v", err)
	}

	var stdout bytes.Buffer
	if exitCode := Run([]string{"open", "Sendable Follow-Up"}, &stdout, &stdout); exitCode != 0 {
		t.Fatalf("expected open sendable follow-up to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveReplacedLater != 1 {
		t.Fatalf("expected passive replacement signal, got %#v", row)
	}
}

func TestObserveAddPassiveExternalTargetSignalsSharedIntent(t *testing.T) {
	stateDir := t.TempDir()
	externalDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	sourcePath := filepath.Join(packDir, "views", "followup.md")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source artifact: %v", err)
	}
	externalPath := filepath.Join(externalDir, "handoff-followup.md")
	if err := os.WriteFile(externalPath, source, 0o644); err != nil {
		t.Fatalf("write external copy: %v", err)
	}

	var stdout bytes.Buffer
	if exitCode := Run([]string{"observe", "add", "--connector", "github", externalPath}, &stdout, &stdout); exitCode != 0 {
		t.Fatalf("expected observe add to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveExportOpened != 1 {
		t.Fatalf("expected passive external shared signal, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"passive_artifact_outcome_signal": "shared-this"`) || !strings.Contains(string(routeSaved), `"passive_artifact_outcome_reason": "connector-github-external-target-present"`) {
		t.Fatalf("expected passive route outcome to record external target presence, got:\n%s", string(routeSaved))
	}
	obsSaved, err := os.ReadFile(filepath.Join(packDir, "external-observation.json"))
	if err != nil {
		t.Fatalf("expected external observation registry: %v", err)
	}
	if !strings.Contains(string(obsSaved), `"connector_id": "github"`) || !strings.Contains(string(obsSaved), `"connector_label": "GitHub"`) {
		t.Fatalf("expected external observation registry to record connector metadata, got:\n%s", string(obsSaved))
	}
}

func TestObserveScanPassiveExternalEditSignalsUsefulness(t *testing.T) {
	stateDir := t.TempDir()
	externalDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	sourcePath := filepath.Join(packDir, "views", "followup.md")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source artifact: %v", err)
	}
	externalPath := filepath.Join(externalDir, "handoff-followup.md")
	if err := os.WriteFile(externalPath, source, 0o644); err != nil {
		t.Fatalf("write external copy: %v", err)
	}
	if exitCode := Run([]string{"observe", "add", "--connector", "github", externalPath}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe add to succeed, got %d", exitCode)
	}
	edited := strings.ReplaceAll(string(source), "Local draft.", "Local draft with recipient cleanup.")
	if err := os.WriteFile(externalPath, []byte(edited), 0o644); err != nil {
		t.Fatalf("rewrite external copy: %v", err)
	}

	if exitCode := Run([]string{"observe", "scan"}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe scan to succeed, got %d", exitCode)
	}
	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveReopened != 1 {
		t.Fatalf("expected passive external used signal, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"passive_artifact_outcome_reason": "connector-github-external-target-edited"`) {
		t.Fatalf("expected passive route outcome to record connector-aware edit, got:\n%s", string(routeSaved))
	}
}

func TestObserveScanPassiveExternalRewriteSignalsReplacement(t *testing.T) {
	stateDir := t.TempDir()
	externalDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Send this note\nLocal draft.\n\n## Decisions captured from the notes\n- One\n\n## Owners and due dates to confirm\n- Owner\n\n## Open questions to close\n- Question\n\n## Recommended next move\n- Next"}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local setup run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	sourcePath := filepath.Join(packDir, "views", "followup.md")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source artifact: %v", err)
	}
	externalPath := filepath.Join(externalDir, "handoff-followup.md")
	if err := os.WriteFile(externalPath, source, 0o644); err != nil {
		t.Fatalf("write external copy: %v", err)
	}
	if exitCode := Run([]string{"observe", "add", "--connector", "github", externalPath}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe add to succeed, got %d", exitCode)
	}
	rewritten := "## Send this note\nDiscard the original follow-up and draft a launch hold notice for legal review only.\n\n## Decisions captured from the notes\n- Freeze the rollout.\n- Cancel the customer handoff.\n\n## Owners and due dates to confirm\n- Legal lead by Tuesday.\n- Ops approver before any resend.\n\n## Open questions to close\n- Whether the incident summary replaces the meeting recap.\n- Whether support needs a separate notice.\n\n## Recommended next move\n- Replace the original artifact with a hold notice before anyone sends it."
	if err := os.WriteFile(externalPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("rewrite external copy: %v", err)
	}

	if exitCode := Run([]string{"observe", "scan"}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe scan to succeed, got %d", exitCode)
	}
	report := loadLocalRuntimeCapabilities()
	row := report.CohortFeedback["local-workhorse"]["sendable-followup"]
	if row.PassiveReplacedLater != 1 {
		t.Fatalf("expected passive external replacement signal, got %#v", row)
	}
	routeSaved, err := os.ReadFile(filepath.Join(packDir, "route.json"))
	if err != nil {
		t.Fatalf("expected saved route file: %v", err)
	}
	if !strings.Contains(string(routeSaved), `"passive_artifact_outcome_signal": "replaced-this"`) || !strings.Contains(string(routeSaved), `"passive_artifact_outcome_reason": "connector-github-external-target-decision-change"`) {
		t.Fatalf("expected passive route outcome to record external rewrite, got:\n%s", string(routeSaved))
	}
}

func TestObserveScanPassiveExternalRewriteRecordsSubtypeSpecificRouteCohortOutcome(t *testing.T) {
	stateDir := t.TempDir()
	externalDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	withProviderHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"choices":[{"message":{"content":"## Extracted evidence\n- The screenshot shows the publish button disabled.\n\n## What is visible\n- The approval warning banner is open in the right panel.\n\n## Still unclear\n- The owner is not visible.\n\n## Recommended next move\n- Confirm the owner before retrying."}}]}`), nil
	})

	if exitCode := RunInteractive(nil, strings.NewReader("Review this screenshot and extract the evidence from the UI.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected local screenshot run to succeed, got %d", exitCode)
	}

	current, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("expected current work state: %v", err)
	}
	var currentState map[string]any
	if err := json.Unmarshal(current, &currentState); err != nil {
		t.Fatalf("expected current work json: %v", err)
	}
	packDir, _ := currentState["pack_dir"].(string)
	viewFile, _ := providerPrimaryView("general-work")
	sourcePath := filepath.Join(packDir, "views", viewFile)
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source artifact: %v", err)
	}
	externalPath := filepath.Join(externalDir, "ui-review.md")
	if err := os.WriteFile(externalPath, source, 0o644); err != nil {
		t.Fatalf("write external copy: %v", err)
	}
	if exitCode := Run([]string{"observe", "add", "--connector", "github", externalPath}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe add to succeed, got %d", exitCode)
	}
	rewritten := "## Extracted evidence\n- Discard the original screenshot review and treat this as a legal hold notice only.\n\n## What is visible\n- The launch is paused, customer publishing must stop, and the right panel lists a hold notice instead of a retry path.\n\n## Still unclear\n- The screenshot does not show the legal reviewer or release approver.\n\n## Recommended next move\n- Replace the original UI guidance with a legal-hold escalation and do not retry publishing."
	if err := os.WriteFile(externalPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("rewrite external copy: %v", err)
	}

	if exitCode := Run([]string{"observe", "scan"}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected observe scan to succeed, got %d", exitCode)
	}

	stats := loadLocalRouteFeedbackStats()
	foundScreenshot := false
	for key, cohorts := range stats.Cohorts {
		row := cohorts["multimodal-image-screenshot"]
		if row.PassiveReplacedLater == 1 {
			foundScreenshot = true
		}
		if pdfRow, ok := cohorts["multimodal-pdf-scan"]; ok && (pdfRow.PassiveReplacedLater != 0 || pdfRow.PassiveExportOpened != 0 || pdfRow.PassiveReopened != 0) {
			t.Fatalf("expected no PDF passive route cohort bleed, found %#v under %q", pdfRow, key)
		}
	}
	if !foundScreenshot {
		t.Fatalf("expected screenshot-scoped passive route cohort outcome to record replacement, got %#v", stats.Cohorts)
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func writeProviderTestCurrentWork(t *testing.T, packDir, packID, workUnitID, title string) {
	t.Helper()
	payload := currentWork{
		PackDir:    packDir,
		PackID:     packID,
		WorkUnitID: workUnitID,
		Title:      title,
		State:      "in_progress",
		Health:     "working",
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal current work: %v", err)
	}
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		t.Fatalf("mkdir state root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionStateRoot(), "current-work.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write current work: %v", err)
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

func writeProviderFakeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake executable %s: %v", name, err)
	}
	return path
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
