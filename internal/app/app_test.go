package app_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maridlabsai/jini/internal/app"
)

func TestCheckRendersPlainLanguageCurrentWorkScreen(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"check"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal",
		"Jini Research To PRD",
		"Working with",
		"Latest PRD draft and review comments",
		"Just finished",
		"Doing now",
		"Up next",
		"AI route",
		"Local preview",
		"Now",
		"Checking assumptions and approval gaps",
		"Done",
		"Build-readiness draft created",
		"Need",
		"Name the approval owner and confirm the first implementation slice.",
		"Why this matters",
		"Options",
		"Set approval owner",
		"Next",
		"Open Build-Readiness Check",
		"Ready now",
		"Handoff Brief",
		"Build-Readiness Check",
		"Blocked",
		"Safe to do",
		"Nothing has been sent yet",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestProviderDoctorDetectsAzureOpenAIWithoutLeakingSecrets(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")
	t.Setenv("AZURE_OPENAI_API_VERSION", "2024-10-21")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Provider",
		"Azure OpenAI",
		"Status",
		"ok",
		"AZURE_OPENAI_ENDPOINT",
		"AZURE_OPENAI_DEPLOYMENT",
		"AZURE_OPENAI_API_VERSION",
		"AZURE_OPENAI_API_KEY: set",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "super-secret-key") {
		t.Fatalf("provider doctor leaked secret value:\n%s", out)
	}
}

func TestProviderDoctorDetectsBedrockWithoutLeakingCredentials(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "bedrock")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_PROFILE", "work-profile")
	t.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-3-5-sonnet-20240620-v1:0")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Provider",
		"Amazon Bedrock",
		"Status",
		"ok",
		"AWS_REGION",
		"BEDROCK_MODEL_ID",
		"AWS_PROFILE: set",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "work-profile") || strings.Contains(out, "anthropic.claude-3-5-sonnet-20240620-v1:0") {
		t.Fatalf("provider doctor should not expose profile or model values in safe mode:\n%s", out)
	}
}

func TestProviderDoctorReportsMissingRequiredSettings(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Provider",
		"Azure OpenAI",
		"Status",
		"needs setup",
		"Missing",
		"AZURE_OPENAI_ENDPOINT",
		"AZURE_OPENAI_API_KEY",
		"AZURE_OPENAI_DEPLOYMENT",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestProviderDoctorDetectsAnthropicDirectSetup(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "super-secret-key")
	t.Setenv("JINI_MODEL", "sonnet")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Provider",
		"Claude API",
		"Status",
		"ok",
		"JINI_MODEL",
		"Claude Sonnet 4",
		"ANTHROPIC_BASE_URL: missing (default https://api.anthropic.com)",
		"ANTHROPIC_API_KEY: set",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "super-secret-key") {
		t.Fatalf("provider doctor leaked Anthropic secret:\n%s", out)
	}
}

func TestProviderDoctorAutoChoosesBedrockForSonnet46Alias(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "sonnet-4.6")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "SECRETEXAMPLE")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Amazon Bedrock",
		"Claude Sonnet 4.6",
		"JINI_PROVIDER: auto -> Amazon Bedrock / Claude Sonnet 4.6",
		"JINI_MODEL: sonnet-4.6",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Count(out, "JINI_PROVIDER:") != 1 {
		t.Fatalf("expected one provider setting line, got:\n%s", out)
	}
}

func TestProviderDoctorRejectsSonnet46ShortcutForDirectClaude(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "super-secret-key")
	t.Setenv("JINI_MODEL", "sonnet-4.6")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Claude API",
		"needs setup",
		"Sonnet 4.6 shortcut is supported only on Bedrock",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestLauncherShowsInlineProviderChoicesWhenUsingLocalPreview(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(""), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Working with",
		"Local preview",
		"Paste notes or type what you want finished.",
		"Need setup help? Type `Use Auto` and Jini will help you connect the best available option.",
		"Not sure? Type `help me finish this`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Choose how Jini should work",
		"Type `Connect Claude`",
		"Type `Connect Bedrock`",
		"Type `Connect Azure OpenAI`",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected output not to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestLauncherShowsAutoModeStateWhenConfigured(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Working with",
		"Local preview (chosen automatically)",
		"Auto mode is on. No cloud provider is ready yet.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveSetupCanSaveClaudeProfileInsideJini(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Use Claude\nsk-test-key\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Connect Claude",
		"Setup saved. Working with Claude API / Claude Sonnet 4.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	saved, err := os.ReadFile(filepath.Join(stateDir, "provider.json"))
	if err != nil {
		t.Fatalf("expected provider settings file: %v", err)
	}
	if !strings.Contains(string(saved), `"JINI_PROVIDER": "claude"`) || !strings.Contains(string(saved), `"JINI_MODEL": "sonnet"`) {
		t.Fatalf("expected saved Claude settings, got:\n%s", string(saved))
	}

	stdout.Reset()
	exitCode = app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected saved profile to drive provider doctor, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Claude API / Claude Sonnet 4") {
		t.Fatalf("expected provider doctor to use saved Claude profile, got:\n%s", stdout.String())
	}
}

func TestInteractiveSetupCanSaveClaudeCodeRouteInsideJini(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Use Claude Code\nsk-test-key\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Connect Claude",
		"Setup saved. Working with Claude Code via Claude API / Claude Sonnet 4.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "claude-code"`) {
		t.Fatalf("expected saved Claude Code route, got:\n%s", string(routerSaved))
	}
}

func TestInteractiveSetupCanSaveLocalSLMInsideJini(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Connect Local SLM\nhttp://127.0.0.1:11434/v1\nqwen3:8b\n\n\n\n\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Connect Local SLM",
		"Setup saved. Working with Local SLM / qwen3:8b.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	saved, err := os.ReadFile(filepath.Join(stateDir, "provider.json"))
	if err != nil {
		t.Fatalf("expected provider settings file: %v", err)
	}
	for _, want := range []string{
		`"JINI_PROVIDER": "local-slm"`,
		`"JINI_LOCAL_SLM_ENDPOINT": "http://127.0.0.1:11434/v1"`,
		`"JINI_LOCAL_SLM_MODEL": "qwen3:8b"`,
	} {
		if !strings.Contains(string(saved), want) {
			t.Fatalf("expected saved local SLM settings to contain %q, got:\n%s", want, string(saved))
		}
	}
}

func TestProviderDoctorShowsAutoToolDecision(t *testing.T) {
	t.Setenv("JINI_TOOL", "auto")
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")
	t.Setenv("ANTHROPIC_API_KEY", "super-secret-key")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Claude API / Claude Sonnet 4",
		"JINI_TOOL: auto -> Claude Code",
		"ROUTE_POLICY: Automatic",
		"JINI_MODEL_DECISION: Claude Sonnet 4",
		"AUTO_MODEL: Jini uses Claude Sonnet 4 by default on the Claude Code route.",
		"AUTO_ROUTE: Auto mode chose Claude Code because this looks like general work, the request does not ask for deep review, so Jini favored the cheapest suitable route. It was the first ready route in this environment.",
		"JINI_EFFORT: auto -> dynamic per request",
		"AUTO_EFFORT: Jini judges effort separately for each request instead of pinning one level globally.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestProviderDoctorDetectsLocalSLMWithoutLeakingOptionalKey(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_API_KEY", "local-secret")
	t.Setenv("JINI_LOCAL_SLM_WORKHORSE_MODEL", "qwen3:8b-instruct")
	t.Setenv("JINI_SKIP_LOCAL_BENCHMARK", "1")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Local SLM",
		"JINI_PROVIDER: local-slm",
		"JINI_LOCAL_SLM_ENDPOINT: set",
		"DEVICE_CLASS: laptop-strong",
		"LOCAL_RUNTIME_CLASS: ollama-openai-compatible",
		"JINI_LOCAL_SLM_MODEL: set -> qwen3:8b",
		"JINI_LOCAL_SLM_WORKHORSE_MODEL: set",
		"JINI_LOCAL_SLM_API_KEY: set",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "local-secret") {
		t.Fatalf("provider doctor leaked local SLM API key:\n%s", out)
	}
}

func TestProviderDoctorShowsSubtypeScopedMultimodalLearning(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "laptop-strong")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_MULTIMODAL_MODEL", "gemma3:12b")
	t.Setenv("JINI_SKIP_LOCAL_BENCHMARK", "1")

	writeLocalRuntimeCapabilitiesFixture(t, stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"provider", "doctor"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"MULTIMODAL_LEARNING_SCREENSHOT: local-multimodal samples=1 signals=1",
		"MULTIMODAL_LEARNING_PDF_SCAN: local-multimodal samples=1 signals=1",
		"MULTIMODAL_LEARNING_AUDIO_TRANSCRIPT: local-workhorse samples=1 signals=1",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveLauncherShowsMultimodalLearningBuckets(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	writeLocalRuntimeCapabilitiesFixture(t, stateDir)

	imagePath := filepath.Join(t.TempDir(), "dashboard-screenshot.png")
	if err := os.WriteFile(imagePath, []byte("not-a-real-image"), 0o644); err != nil {
		t.Fatalf("write screenshot fixture: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(imagePath+"\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Multimodal learning",
		"Screenshot work: local-multimodal (samples 1, signals 1)",
		"Scanned PDF work: local-multimodal (samples 1, signals 1)",
		"Audio/transcript work: local-workhorse (samples 1, signals 1)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCheckShowsMultimodalLearningBucketsForCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")
	writeLocalRuntimeCapabilitiesFixture(t, stateDir)

	inputs := map[string]any{
		"schema_version": "0.1.0",
		"context_type":   "JiniInputItems",
		"items": []map[string]any{{
			"input_id": "attachment-1",
			"kind":     "image",
			"title":    "dashboard-screenshot.png",
			"status":   "received",
		}},
	}
	data, err := json.MarshalIndent(inputs, "", "  ")
	if err != nil {
		t.Fatalf("marshal input items: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "inputs.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write input items: %v", err)
	}

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"check"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Multimodal learning",
		"Screenshot work: local-multimodal (samples 1, signals 1)",
		"Scanned PDF work: local-multimodal (samples 1, signals 1)",
		"Audio/transcript work: local-workhorse (samples 1, signals 1)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveLauncherReportsProviderSetupBeforeCreatingWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "azure-openai")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Plan 7 day Paris trip\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Could not start this work",
		"Provider needs setup.",
		"Run `jini provider doctor`.",
		"AZURE_OPENAI_ENDPOINT",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no current work after provider setup failure, got err=%v", err)
	}
}

func TestCheckHandlesStaleCurrentWorkWithoutLeakingPath(t *testing.T) {
	stateDir := t.TempDir()
	missingPackDir := filepath.Join(t.TempDir(), "deleted-work")
	writeCurrentWork(t, stateDir, missingPackDir, "research-prd", "stale-work", "Stale Work", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"check"}, &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Remembered work is no longer available.",
		"Run `jini` to start something.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, leak := range []string{missingPackDir, "stat ", "no such file or directory"} {
		if strings.Contains(out, leak) {
			t.Fatalf("stale work error leaked filesystem detail %q:\n%s", leak, out)
		}
	}
}

func TestLauncherRecoversFromStaleCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	missingPackDir := filepath.Join(t.TempDir(), "deleted-work")
	writeCurrentWork(t, stateDir, missingPackDir, "research-prd", "stale-work", "Stale Work", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Remembered work is no longer available.",
		"What do you need help finishing?",
		"Working with",
		"Local preview",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, leak := range []string{missingPackDir, "stat ", "no such file or directory"} {
		if strings.Contains(out, leak) {
			t.Fatalf("launcher leaked stale filesystem detail %q:\n%s", leak, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected stale current-work file to be cleared, got err=%v", err)
	}
}

func TestOpenListsHumanReadableOutputs(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"open"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Ready now",
		"Handoff Brief",
		"Missing Pieces Before Build",
		"Send / share",
		"Markdown Wiki",
		"Details",
		"Brief",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestOpenPrintsNamedView(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"open", "prd"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	if got := stdout.String(); !strings.Contains(got, "# Build-Readiness Check") {
		t.Fatalf("expected build-readiness content, got:\n%s", got)
	}
}

func TestLauncherShowsCurrentWorkRecap(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal",
		"Weekly Product Review",
		"Working with",
		"Meeting notes and follow-up tasks",
		"Just finished",
		"Doing now",
		"Up next",
		"Now",
		"Turning notes into owners and next steps",
		"Ready now",
		"Sendable Follow-up",
		"Next",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Continue current work",
		"Open what's ready",
		"Start something new",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected noninteractive launcher to hide fake choice %q, got:\n%s", unwanted, out)
		}
	}
}

func TestLauncherShowsCurrentWorkContinuityReason(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	writeCurrentWorkRoute(t, packDir, map[string]any{
		"schema_version":       "0.1.0",
		"context_type":         "JiniWorkRoute",
		"tool_mode":            "codex",
		"tool_label":           "Azure code route",
		"route_policy":         "Automatic",
		"model_label":          "gpt-4.1",
		"provider_label":       "Azure OpenAI / gpt-4.1",
		"chosen_automatically": true,
		"reason":               "Auto mode picked the cheapest suitable coding route.",
		"continuity_reason":    "Kept the current coding route to preserve context continuity because the quality gap was not material.",
	})

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Continuity",
		"Kept the current coding route to preserve context continuity because the quality gap was not material.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCheckHighlightsSpecificMeetingGaps(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"check"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Blocked",
		"Metric and legal-review decision",
		"Confirm any missing owner or due date before sending this follow-up.",
		"Why this matters",
		"If you skip this",
		"Not sure about",
		"Whether the metric decision also needs legal review",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCheckHighlightsSpecificTravelGaps(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedTravelWork(t)
	writeCurrentWork(t, stateDir, packDir, "travel-plan", "paris-7d", "7-Day Paris Trip", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"check"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Blocked",
		"Dates, budget, and hotel area",
		"Confirm dates, budget, and hotel area before booking from this draft.",
		"Options",
		"Add dates",
		"Not sure about",
		"Whether Versailles is a must-do day trip or an optional slot",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestLauncherShowsStartChoicesWithoutCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"What do you need help finishing?",
		"Paste notes or type what you want finished.",
		"Not sure? Type `help me finish this`.",
		"Turn meeting notes into something I can send",
		"Check whether a plan is ready to hand off",
		"I am not sure",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"choose a common job below",
		"1. Turn meeting notes",
		"2. Check whether",
		"3. I am not sure",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected shell-first launcher not to expose menu %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveLauncherCreatesMeetingWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("Weekly product review for pricing launch. Need owners, due dates, and open questions.\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"What do you need help finishing?",
		"Sendable Follow-up",
		"## Send this",
		"Keep going",
		"Make it fuller",
		"Show what is missing",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Paste notes, paste a doc, or describe it in one line",
		"Short version or full version",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected output not to contain %q, got:\n%s", unwanted, out)
		}
	}
	if strings.Contains(out, "Goal") && strings.Index(out, "## Send this") > strings.Index(out, "Goal") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "meeting-followup" {
		t.Fatalf("expected meeting-followup current work, got %#v", current)
	}
	followupPath := filepath.Join(current["pack_dir"].(string), "views", "followup.md")
	content := mustReadFile(t, followupPath)
	for _, want := range []string{
		"pricing launch",
		"## Decisions captured from the notes",
		"## Owners and due dates to confirm",
		"## Open questions to close",
		"## Recommended next move",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected followup to contain %q, got:\n%s", want, content)
		}
	}
	ownersPath := filepath.Join(current["pack_dir"].(string), "views", "owners-and-due-points.md")
	owners := mustReadFile(t, ownersPath)
	for _, want := range []string{
		"# Owners and Due Points",
		"## Confirmed from the notes",
		"## Still missing owner or date",
		"## Follow-up questions",
	} {
		if !strings.Contains(owners, want) {
			t.Fatalf("expected owners view to contain %q, got:\n%s", want, owners)
		}
	}
	threadStatePath := filepath.Join(current["pack_dir"].(string), "thread-state.json")
	threadState := mustReadFile(t, threadStatePath)
	for _, want := range []string{
		"\"current_turn\"",
		"\"active_ask\"",
		"confirm-owners-and-dates",
	} {
		if !strings.Contains(threadState, want) {
			t.Fatalf("expected thread state to contain %q, got:\n%s", want, threadState)
		}
	}
}

func TestInteractiveLauncherCreatesSpecReadinessWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("Notifications PRD needs a build-readiness check and handoff call.\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"What do you need help finishing?",
		"Build-Readiness Check",
		"## What looks ready now",
		"Keep going",
		"Make it fuller",
		"Show what is missing",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Short version or full version") {
		t.Fatalf("expected no first-run output-size prompt, got:\n%s", out)
	}
	if strings.Contains(out, "Paste notes, paste a doc, or describe it in one line") {
		t.Fatalf("expected no legacy source-context prompt, got:\n%s", out)
	}
	if strings.Contains(out, "Goal") && strings.Index(out, "## What looks ready now") > strings.Index(out, "Goal") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "research-prd" {
		t.Fatalf("expected research-prd current work, got %#v", current)
	}
	prdPath := filepath.Join(current["pack_dir"].(string), "views", "prd.md")
	prd := mustReadFile(t, prdPath)
	for _, want := range []string{
		"Notifications PRD",
		"## What looks ready now",
		"## Must clear before build",
		"## Recommended first slice",
		"## Who needs to answer what",
		"## Still to confirm",
	} {
		if !strings.Contains(prd, want) {
			t.Fatalf("expected readiness draft to contain %q, got:\n%s", want, prd)
		}
	}
	missingPath := filepath.Join(current["pack_dir"].(string), "views", "missing-pieces-before-build.md")
	missing := mustReadFile(t, missingPath)
	for _, want := range []string{
		"# Missing Pieces Before Build",
		"approval",
		"rollback",
		"first implementation slice",
	} {
		if !strings.Contains(strings.ToLower(missing), strings.ToLower(want)) {
			t.Fatalf("expected missing-pieces view to contain %q, got:\n%s", want, missing)
		}
	}
}

func TestInteractiveLauncherHandlesUnsureInputWithUsefulPass(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("3\nI have a messy note about renewal risks and next steps\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"I am not sure",
		"Paste what you have. A rough version is fine.",
		"I will help figure out whether this is follow-up, a plan check, or something else.",
		"Nothing will be sent yet.",
		"First Useful Pass",
		"What this seems to be",
		"What can be used now",
		"What I need next",
		"Safe right now",
		"Nothing has been sent",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Short version or full version") {
		t.Fatalf("expected no first-run output-size prompt, got:\n%s", out)
	}
	if strings.Contains(out, "Goal") && strings.Index(out, "First Useful Pass") > strings.Index(out, "Goal") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "general-work" {
		t.Fatalf("expected general-work current work, got %#v", current)
	}
}

func TestInteractiveLauncherRunsMeetingPostResultActions(t *testing.T) {
	cases := []struct {
		name            string
		action          string
		wantAfterAction string
	}{
		{
			name:            "keep going opens the next useful meeting surface",
			action:          "Keep going",
			wantAfterAction: "Owners and Due Points",
		},
		{
			name:            "see missing opens the meeting gap summary",
			action:          "Show what is missing",
			wantAfterAction: "Metric and legal-review decision",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Weekly product review. Need owners, due dates, and open questions."
			baseline := runInteractiveForTest(t, t.TempDir(), source+"\n")

			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+tc.action+"\n")
			for _, want := range []string{
				"## Send this",
				"Keep going",
				"Show what is missing",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			if out == baseline {
				t.Fatalf("expected %q to be handled as a real action, but output matched the no-action run:\n%s", tc.action, out)
			}
			requireStringAfter(t, out, tc.action, tc.wantAfterAction)
			for _, unwanted := range []string{"coming soon", "not implemented"} {
				if strings.Contains(strings.ToLower(out), unwanted) {
					t.Fatalf("expected real action, got placeholder %q in:\n%s", unwanted, out)
				}
			}
		})
	}
}

func TestInteractiveLauncherRunsSpecPostResultActions(t *testing.T) {
	cases := []struct {
		name            string
		action          string
		wantAfterAction string
	}{
		{
			name:            "keep going opens the next useful spec surface",
			action:          "Keep going",
			wantAfterAction: "Missing Pieces Before Build",
		},
		{
			name:            "see missing opens the spec gap summary",
			action:          "Show what is missing",
			wantAfterAction: "Approval",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Notifications PRD needs a build-readiness check and handoff call."
			baseline := runInteractiveForTest(t, t.TempDir(), source+"\n")

			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+tc.action+"\n")
			for _, want := range []string{
				"## What looks ready now",
				"Keep going",
				"Show what is missing",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			if out == baseline {
				t.Fatalf("expected %q to be handled as a real action, but output matched the no-action run:\n%s", tc.action, out)
			}
			requireStringAfter(t, out, tc.action, tc.wantAfterAction)
			for _, unwanted := range []string{"coming soon", "not implemented"} {
				if strings.Contains(strings.ToLower(out), unwanted) {
					t.Fatalf("expected real action, got placeholder %q in:\n%s", unwanted, out)
				}
			}
		})
	}
}

func TestCurrentWorkInteractiveChoicesAreReal(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Show what's ready\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Open something ready") {
		t.Fatalf("expected open shelf after choosing current work option 2, got:\n%s", out)
	}
	if !strings.Contains(out, "Sendable Follow-up") {
		t.Fatalf("expected sendable follow-up in open shelf, got:\n%s", out)
	}
}

func TestCurrentWorkCanEnterPlanFirstMode(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Help me plan this\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Help me plan this",
		"Goal",
		"Requirements",
		"Design",
		"Steps",
		"Run",
		"Build-Readiness Check",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveLauncherCreatesTravelWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("7 day Paris trip for a couple with a $2500 budget\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal",
		"7 Day Paris Trip For A Couple With A 2500 Budget",
		"Working with",
		"Your request: 7 day Paris trip for a couple with a $2500 budget",
		"Ready now",
		"Itinerary",
		"Budget Sketch",
		"Blocked",
		"Dates, budget, and hotel area",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected travel-plan current work, got %#v", current)
	}
}

func TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist(t *testing.T) {
	stateDir := t.TempDir()
	copyWorkDir(t, filepath.Join(stateDir, "work", "travel-plan-paris"), seedTravelWork(t))
	meetingDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal",
		"Weekly Product Review",
		"Other active work",
		"7-Day Paris Trip",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveLauncherCanSwitchBetweenActiveProjects(t *testing.T) {
	stateDir := t.TempDir()
	travelDir := copyWorkDir(t, filepath.Join(stateDir, "work", "travel-plan-paris"), seedTravelWork(t))
	meetingDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Switch project\n1\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Switch project",
		"1. 7-Day Paris Trip",
		"Switched to",
		"7-Day Paris Trip",
		"Ready now",
		"Itinerary",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected current work to switch to travel-plan, got %#v", current)
	}
	if current["pack_dir"] != travelDir {
		t.Fatalf("expected current work to switch to copied travel dir %q, got %#v", travelDir, current)
	}
}

func TestCurrentWorkFreeformInputStartsNewWork(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Your first draft is ready.",
		"Goal",
		"Plan Me A 7 Day Paris Trip",
		"Ready now",
		"Itinerary",
		"Budget Sketch",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Weekly Product Review Follow-up") && !strings.Contains(out, "Plan Me A 7 Day Paris Trip") {
		t.Fatalf("expected new work to replace the old current-work view, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected current work to switch to travel-plan, got %#v", current)
	}
}

func TestInteractiveLauncherShowsAttachmentInputChipForTextFile(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	attachment := filepath.Join(t.TempDir(), "meeting-notes.txt")
	writeFile(t, attachment, "Weekly product review. Need owners and due dates.")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(attachment+"\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Working with",
		"meeting-notes.txt (processed)",
		"Goal",
		"Need",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func writeCurrentWork(t *testing.T, stateDir, packDir, packID, workUnitID, title, state, health string) {
	t.Helper()

	payload := map[string]any{
		"schema_version": "0.1.0",
		"context_type":   "JiniCurrentWork",
		"pack_dir":       packDir,
		"pack_id":        packID,
		"work_unit_id":   workUnitID,
		"title":          title,
		"state":          state,
		"health":         health,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal current work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "current-work.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write current work: %v", err)
	}
}

func writeCurrentWorkRoute(t *testing.T, packDir string, payload map[string]any) {
	t.Helper()
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal current work route: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packDir, "route.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write current work route: %v", err)
	}
}

func readCurrentWork(t *testing.T, stateDir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(stateDir, "current-work.json"))
	if err != nil {
		t.Fatalf("read current work: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode current work: %v", err)
	}
	return payload
}

func copyWorkDir(t *testing.T, destination, source string) string {
	t.Helper()
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", destination, err)
	}
	if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	}); err != nil {
		t.Fatalf("copy work dir: %v", err)
	}
	return destination
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func runInteractiveForTest(t *testing.T, stateDir, input string) string {
	t.Helper()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(input), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	return stdout.String()
}

func requireStringAfter(t *testing.T, out, anchor, want string) {
	t.Helper()
	anchorIndex := strings.Index(out, anchor)
	if anchorIndex < 0 {
		t.Fatalf("expected output to contain action %q, got:\n%s", anchor, out)
	}
	if !strings.Contains(out[anchorIndex+len(anchor):], want) {
		t.Fatalf("expected output after %q to contain %q, got:\n%s", anchor, want, out)
	}
}

func seedResearchPRDWork(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "views"))
	mkdirAll(t, filepath.Join(root, "artifacts"))
	mkdirAll(t, filepath.Join(root, "exports", "wiki", "markdown"))

	writeFile(t, filepath.Join(root, "work-unit.yaml"), strings.TrimSpace(`
work_unit_id: example-research-prd
title: Jini Research To PRD
current_state: awaiting_verification
`)+"\n")
	writeFile(t, filepath.Join(root, "views", "prd.md"), "# Build-Readiness Check\n\nBuild-ready draft.\n")
	writeFile(t, filepath.Join(root, "views", "tasks.md"), "# Missing Pieces Before Build\n\n- Verify approval\n")
	writeFile(t, filepath.Join(root, "artifacts", "01-brief.yaml"), "artifact_type: Brief\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "artifacts", "08-plan.yaml"), "artifact_type: Plan\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "artifacts", "09-tasks.yaml"), "artifact_type: Tasks\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "artifacts", "10-evidence.yaml"), "artifact_type: Evidence\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "exports", "wiki", "markdown", "overview.md"), "# Markdown Wiki\n")

	return root
}

func writeLocalRuntimeCapabilitiesFixture(t *testing.T, stateDir string) {
	t.Helper()
	report := map[string]any{
		"schema_version": "0.4.0",
		"context_type":   "JiniLocalRuntimeCapabilities",
		"captured_at":    time.Now().UTC().Format(time.RFC3339),
		"adapters": map[string]any{
			"local-multimodal": map[string]any{"adapter_id": "local-multimodal", "model_id": "gemma3:12b", "status": "ok"},
		},
		"cohort_history": map[string]any{
			"local-multimodal": map[string]any{
				"multimodal-image-screenshot": []map[string]any{{"model_id": "gemma3:12b", "status": "ok", "benchmarked_at": time.Now().UTC().Format(time.RFC3339)}},
				"multimodal-pdf-scan":         []map[string]any{{"model_id": "gemma3:12b", "status": "ok", "benchmarked_at": time.Now().UTC().Format(time.RFC3339)}},
			},
			"local-workhorse": map[string]any{
				"multimodal-audio-transcript": []map[string]any{{"model_id": "qwen3:8b", "status": "ok", "benchmarked_at": time.Now().UTC().Format(time.RFC3339)}},
			},
		},
		"cohort_feedback": map[string]any{
			"local-multimodal": map[string]any{
				"multimodal-image-screenshot": map[string]any{"outcome_shared": 1},
				"multimodal-pdf-scan":         map[string]any{"accepted_as_is": 1},
			},
			"local-workhorse": map[string]any{
				"multimodal-audio-transcript": map[string]any{"needed_light_edits": 1},
			},
		},
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "local-runtime-capabilities.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write local runtime capabilities: %v", err)
	}
}

func seedMeetingWork(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "views"))
	mkdirAll(t, filepath.Join(root, "artifacts"))

	writeFile(t, filepath.Join(root, "work-unit.yaml"), strings.TrimSpace(`
work_unit_id: example-meeting-followup
title: Weekly Product Review
current_state: decided
`)+"\n")
	writeFile(t, filepath.Join(root, "views", "followup.md"), "# Sendable Follow-up\n\nHere is what we agreed.\n")
	writeFile(t, filepath.Join(root, "views", "tasks.md"), "# Task List\n\n- Confirm owners\n")
	writeFile(t, filepath.Join(root, "artifacts", "01-brief.yaml"), "artifact_type: Brief\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "artifacts", "06-tasks.yaml"), "artifact_type: Tasks\nstatus: ready\n")

	return root
}

func seedTravelWork(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "views"))
	mkdirAll(t, filepath.Join(root, "artifacts"))

	writeFile(t, filepath.Join(root, "work-unit.yaml"), strings.TrimSpace(`
work_unit_id: paris-7d
title: 7-Day Paris Trip
current_state: decided
`)+"\n")
	writeFile(t, filepath.Join(root, "views", "itinerary.md"), "# Itinerary\n\nDay-by-day draft.\n")
	writeFile(t, filepath.Join(root, "views", "budget-sketch.md"), "# Budget Sketch\n\n- Lodging\n")
	writeFile(t, filepath.Join(root, "views", "travel-logistics.md"), "# Travel Logistics\n\n- Hotel area\n")
	writeFile(t, filepath.Join(root, "views", "tasks.md"), "# Task List\n\n- Lock dates and budget\n")
	writeFile(t, filepath.Join(root, "artifacts", "01-brief.yaml"), "artifact_type: Brief\nstatus: ready\n")
	writeFile(t, filepath.Join(root, "artifacts", "06-tasks.yaml"), "artifact_type: Tasks\nstatus: ready\n")

	return root
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
