package app_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maridlabsai/jini/internal/app"
)

func TestStatusRendersPlainLanguageCurrentWorkScreen(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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

func TestLauncherHelpHidesProviderStateWhenUsingLocalPreview(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Paste what you want finished.",
		"If you want help shaping a messy ask, type `I'm not sure`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Working with",
		"Local preview",
		"Need setup help?",
		"Choose how Jini should work",
		"Type `Claude`",
		"Type `Bedrock`",
		"Type `Azure`",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected output not to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestLauncherHelpHidesAutoModeStateWhenConfigured(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "auto")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, unwanted := range []string{
		"Working with",
		"Local preview (chosen automatically)",
		"Auto mode is on.",
		"Need setup help?",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected output not to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestTopLevelHelpFlagShowsLauncherHelp(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"--help"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Paste what you want finished.",
		"If you need commands, type `help`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCommandsAliasShowsPublicCommandInventory(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"commands"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Public command inventory",
		"START WITH JINI",
		"SUPPORT THE CURRENT WORK",
		"jini status",
		"jini continue",
		"jini open",
		"jini doctor",
		"jini admin help",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"jini check", "jini provider doctor", "jini help"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("did not expect output to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestHelpAllShowsPublicCommandInventory(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"help", "--all"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Public command inventory") {
		t.Fatalf("expected public command inventory, got:\n%s", stdout.String())
	}
}

func TestAdminHelpAliasShowsAdminInventory(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"admin", "help"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Admin and developer command inventory",
		"jini provider doctor",
		"jini observe status",
		"jini open <artifact>",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestProviderDoctorSubcommandMatchesTopLevelDoctor(t *testing.T) {
	var topLevel bytes.Buffer
	topExit := app.Run([]string{"doctor"}, &topLevel, &topLevel)
	if topExit != 0 {
		t.Fatalf("expected top-level doctor to succeed, got %d with output:\n%s", topExit, topLevel.String())
	}

	var subcommand bytes.Buffer
	subExit := app.Run([]string{"provider", "doctor"}, &subcommand, &subcommand)
	if subExit != 0 {
		t.Fatalf("expected provider doctor to succeed, got %d with output:\n%s", subExit, subcommand.String())
	}

	if topLevel.String() != subcommand.String() {
		t.Fatalf("expected provider doctor to match top-level doctor.\nTOP LEVEL:\n%s\nSUBCOMMAND:\n%s", topLevel.String(), subcommand.String())
	}
}

func TestCheckAliasRendersCurrentWorkScreen(t *testing.T) {
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
		"Ready now",
		"Build-Readiness Check",
		"Blocked",
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
exitCode := app.RunInteractive(nil, strings.NewReader("Claude\nsk-test-key\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Claude",
		"Setup saved. Working with Claude Code via Claude API / Claude Sonnet 4.",
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
exitCode = app.Run([]string{"doctor"}, &stdout, &stdout)
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
	exitCode := app.RunInteractive(nil, strings.NewReader("Claude\nsk-test-key\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Claude",
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
	exitCode := app.RunInteractive(nil, strings.NewReader("Local\nhttp://127.0.0.1:11434/v1\nqwen3:8b\n\n\n\n\n\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Local",
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
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
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
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
	exitCode := app.RunInteractive(nil, strings.NewReader("Plan 7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Could not start this work",
		"Provider needs setup.",
		"Run `jini doctor`.",
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

func TestStatusHandlesStaleCurrentWorkWithoutLeakingPath(t *testing.T) {
	stateDir := t.TempDir()
	missingPackDir := filepath.Join(t.TempDir(), "deleted-work")
	writeCurrentWork(t, stateDir, missingPackDir, "research-prd", "stale-work", "Stale Work", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Remembered work is no longer available.",
		"No current work yet.",
		"Paste what you want finished.",
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
		"Jini",
		"Paste what you want finished.",
		"Type `help` if you want examples or commands.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Working with",
		"Examples:",
		"Need setup help?",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected compact stale-work recovery not to contain %q, got:\n%s", unwanted, out)
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

func TestContinuePrintsNextUsefulView(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"continue"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") {
		t.Fatalf("expected continue to open the next useful artifact, got:\n%s", out)
	}
	if strings.Contains(out, "Working Draft") {
		t.Fatalf("expected continue not to fall back to a working draft, got:\n%s", out)
	}
}

func TestLauncherStartsAsCompactShellWhenCurrentWorkExists(t *testing.T) {
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
		"Jini",
		"Current work",
		"Weekly Product Review",
		"Paste a new request, or type `help` to inspect current work.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Goal",
		"Working with",
		"Just finished",
		"Doing now",
		"Need",
		"Why this matters",
		"Continue",
		"Open",
		"Start",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected bare launcher to hide default work detail %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkHelpShowsCurrentWorkRecap(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal",
		"Weekly Product Review",
		"Working with",
		"Meeting notes and follow-up tasks",
		"Ready now",
		"Sendable Follow-up",
		"Blocked",
		"Metric and legal-review decision",
		"Actions",
		"Continue",
		"Start",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Just finished",
		"Doing now",
		"Up next",
		"Now",
		"AI route",
		"Continuity",
		"Why this model",
		"Why this route",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected help recap to hide %q, got:\n%s", unwanted, out)
		}
	}
}

func TestHelpHidesCurrentWorkContinuityReason(t *testing.T) {
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
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	for _, unwanted := range []string{
		"Continuity",
		"Kept the current coding route to preserve context continuity because the quality gap was not material.",
		"AI route",
		"gpt-4.1",
		"Azure OpenAI / gpt-4.1",
	} {
		if strings.Contains(stdout.String(), unwanted) {
			t.Fatalf("expected help surface to hide %q, got:\n%s", unwanted, stdout.String())
		}
	}
}

func TestHelpShowsPlainLanguageFeedbackShelfWhenModelContextExists(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	writeCurrentWorkRoute(t, packDir, map[string]any{
		"schema_version":       "0.1.0",
		"context_type":         "JiniWorkRoute",
		"tool_mode":            "codex",
		"tool_label":           "Azure writing route",
		"route_policy":         "Automatic",
		"model_label":          "gpt-4o-prod",
		"provider_label":       "Azure OpenAI / gpt-4o-prod",
		"chosen_automatically": true,
	})

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected help surface to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Feedback") {
		t.Fatalf("expected help surface to show feedback shelf, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "Tell Jini how this draft went") {
		t.Fatalf("expected help surface to avoid product-shaped feedback heading, got:\n%s", stdout.String())
	}
}

func TestStatusHighlightsSpecificMeetingGaps(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
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

func TestStatusHighlightsSpecificTravelGaps(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected setup run to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Blocked",
		"Must Do Anchors, Or Whether You Want Help Choosing Them",
		"Confirm the highest-impact trip details before booking from this draft.",
		"Options",
		"Add must do anchors, or whether you want help choosing them",
		"Not sure about",
		"Which one or two anchor experiences should be locked first",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestLauncherStartsAsCompactShellWithoutCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run(nil, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Paste what you want finished.",
		"Type `help` if you want examples or commands.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Examples:",
		"Working with",
		"Need setup help?",
		"choose a common job below",
		"1. Turn meeting notes",
		"2. Check whether",
		"If you want help shaping a messy ask, type `I'm not sure`.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected shell-first launcher not to expose menu %q, got:\n%s", unwanted, out)
		}
	}
}

func TestLauncherHelpShowsStartChoicesWithoutCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Paste what you want finished.",
		"Turn meeting notes into something I can send",
		"Check whether a plan is ready to hand off",
		"Plan a 7 day Paris trip for two adults in October",
		"Compare these vendors and recommend one",
		"If you want help shaping a messy ask, type `I'm not sure`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Working with",
		"Need setup help?",
		"Auto mode is on.",
		"Claude",
		"OpenAI",
		"Azure OpenAI",
		"Bedrock",
		"Local preview",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected help launcher not to leak provider/setup state %q, got:\n%s", unwanted, out)
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
		"Jini",
		"Paste what you want finished.",
		"Sendable Follow-up",
		"## Send this",
		"Continue",
		"Open",
		"Missing",
		"Expand",
		"Plan",
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
	assertNoFirstRunStatusDump(t, out)

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
		"Jini",
		"Paste what you want finished.",
		"Build-Readiness Check",
		"## What looks ready now",
		"Continue",
		"Open",
		"Missing",
		"Expand",
		"Plan",
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
	assertNoFirstRunStatusDump(t, out)

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
		"Jini",
		"Paste what you want finished.",
		"Paste what you want finished. Rough notes are fine.",
		"I will turn it into a useful draft or ask one short follow-up if something important is missing.",
		"Nothing will be sent yet.",
		"Working Draft",
		"What this looks like",
		"Useful starting point",
		"Best next inputs",
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
	if strings.Contains(out, "Goal") && strings.Index(out, "Working Draft") > strings.Index(out, "Goal") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}
	assertNoFirstRunStatusDump(t, out)

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "general-work" {
		t.Fatalf("expected general-work current work, got %#v", current)
	}
}

func TestInteractiveLauncherHelpMeFinishThisAsksForRoughContext(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("help me finish this\nWeekly product review for pricing launch. Need owners, due dates, and open questions.\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Paste what you want finished. Rough notes are fine.",
		"I will turn it into a useful draft or ask one short follow-up if something important is missing.",
		"Sendable Follow-up",
		"## Send this",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Working Draft: Help Me Finish This",
		"What this looks like\n- help me finish this",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected help-me-finish path not to create literal work %q, got:\n%s", unwanted, out)
		}
	}
	assertNoFirstRunStatusDump(t, out)

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "meeting-followup" {
		t.Fatalf("expected help-me-finish context to classify as meeting follow-up, got %#v", current)
	}
}

func TestInteractiveLauncherGreetingDoesNotCreateWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("hello\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Paste what you want finished.",
		"Type `help` if you want examples or commands.",
		"Hi.",
		"Tell me what you want finished, or paste notes when you're ready.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Your first draft is ready.", "Working Draft", "Goal"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("did not expect output to contain %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no current work file after greeting, got err=%v", err)
	}
}

func TestInteractiveLauncherSocialAckDoesNotCreateWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("thanks\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Nothing to do yet.",
		"Paste the work when you're ready.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Your first draft is ready.", "Working Draft", "Goal"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("did not expect output to contain %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no current work file after social ack, got err=%v", err)
	}
}

func TestInteractiveLauncherRejectsSlashCommands(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "status",
			line: "/status\n",
			want: []string{"Unknown command \"/status\"."},
		},
		{
			name: "doctor",
			line: "/doctor\n",
			want: []string{"Unknown command \"/doctor\"."},
		},
		{
			name: "init",
			line: "/init\n",
			want: []string{"Unknown command \"/init\"."},
		},
		{
			name: "memory",
			line: "/memory\n",
			want: []string{"Unknown command \"/memory\"."},
		},
		{
			name: "permissions",
			line: "/permissions\n",
			want: []string{"Unknown command \"/permissions\"."},
		},
		{
			name: "clear",
			line: "/clear\n",
			want: []string{"Unknown command \"/clear\"."},
		},
		{
			name: "route",
			line: "/route\n",
			want: []string{"Unknown command \"/route\"."},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(tc.line), &stdout, &stdout)
			if exitCode != 1 {
				t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{"Your first draft is ready.", "Working Draft"} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected slash command not to create work %q, got:\n%s", unwanted, out)
				}
			}
			if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected no current work file after rejected slash command, got err=%v", err)
			}
		})
	}
}

func TestInteractiveLauncherHelpWithPunctuationDoesNotCreateWork(t *testing.T) {
	for _, line := range []string{"what can you do?\n", "help!\n", "?\n"} {
		t.Run(strings.TrimSpace(line), func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(line), &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range []string{
				"Jini",
				"Examples:",
				"If you need commands, type `help`.",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{"Your first draft is ready.", "Working Draft"} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected punctuated help not to create work %q, got:\n%s", unwanted, out)
				}
			}
			if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected no current work file after help input, got err=%v", err)
			}
		})
	}
}

func TestInteractiveLauncherMenuPhraseWithSentencePunctuationStartsWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Turn meeting notes into something I can send.\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Your first draft is ready.",
		"Sendable Follow-up",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "meeting-followup" {
		t.Fatalf("expected meeting-followup current work, got %#v", current)
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
			action:          "Continue",
			wantAfterAction: "Owners and Due Points",
		},
		{
			name:            "continue opens the next useful meeting surface",
			action:          "Continue",
			wantAfterAction: "Owners and Due Points",
		},
		{
			name:            "see missing opens the meeting decision surface",
			action:          "Missing",
			wantAfterAction: "Pending decision",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Weekly product review. Need owners, due dates, and open questions."
			baseline := runInteractiveForTest(t, t.TempDir(), source+"\n")

			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+tc.action+"\n")
			for _, want := range []string{
				"## Send this",
				"Continue",
				"Missing",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			if out == baseline {
				t.Fatalf("expected %q to be handled as a real action, but output matched the no-action run:\n%s", tc.action, out)
			}
			if strings.Contains(out, tc.action) {
				requireStringAfter(t, out, tc.action, tc.wantAfterAction)
			} else if !strings.Contains(out, tc.wantAfterAction) {
				t.Fatalf("expected output to contain %q for action %q, got:\n%s", tc.wantAfterAction, tc.action, out)
			}
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
			action:          "Continue",
			wantAfterAction: "Missing Pieces Before Build",
		},
		{
			name:            "see missing opens the spec decision surface",
			action:          "Missing",
			wantAfterAction: "Pending decision",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Notifications PRD needs a build-readiness check and handoff call."
			baseline := runInteractiveForTest(t, t.TempDir(), source+"\n")

			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+tc.action+"\n")
			for _, want := range []string{
				"## What looks ready now",
				"Continue",
				"Missing",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			if out == baseline {
				t.Fatalf("expected %q to be handled as a real action, but output matched the no-action run:\n%s", tc.action, out)
			}
			if strings.Contains(out, tc.action) {
				requireStringAfter(t, out, tc.action, tc.wantAfterAction)
			} else if !strings.Contains(out, tc.wantAfterAction) {
				t.Fatalf("expected output to contain %q for action %q, got:\n%s", tc.wantAfterAction, tc.action, out)
			}
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
	exitCode := app.RunInteractive(nil, strings.NewReader("open\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Open") {
		t.Fatalf("expected open shelf after choosing current work option 2, got:\n%s", out)
	}
	if !strings.Contains(out, "Sendable Follow-up") {
		t.Fatalf("expected sendable follow-up in open shelf, got:\n%s", out)
	}
}

func TestCurrentWorkInteractiveOpenCommandOpensShelf(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("open\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{"Open", "1. Sendable Follow-up", "Type a number or name to open one"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Working Draft") {
		t.Fatalf("expected open command not to start work, got:\n%s", out)
	}
}

func TestCurrentWorkInteractiveResumeCommandsOpenFocusedSurface(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{line: "continue\n", want: "Owners and Due Points"},
		{line: "resume\n", want: "Sendable Follow-up"},
	}
	for _, tc := range cases {
		t.Run(strings.TrimSpace(tc.line), func(t *testing.T) {
			stateDir := t.TempDir()
			packDir := seedMeetingWork(t)
			writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(tc.line), &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			if !strings.Contains(out, tc.want) {
				t.Fatalf("expected resume command to reopen focused surface, got:\n%s", out)
			}
			if strings.Contains(out, "Working Draft") {
				t.Fatalf("expected continuation command not to start new work, got:\n%s", out)
			}
		})
	}
}

func TestCurrentWorkResumeSetsArtifactFocusWhenNoFocusExists(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedTravelWork(t)
	writeCurrentWork(t, stateDir, packDir, "travel-plan", "example-travel-plan", "7-Day Paris Trip", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("resume\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Itinerary") {
		t.Fatalf("expected resume to open the primary artifact, got:\n%s", stdout.String())
	}

	current := readCurrentWork(t, stateDir)
	threadState := mustReadFile(t, filepath.Join(current["pack_dir"].(string), "thread-state.json"))
	for _, want := range []string{`"kind": "artifact"`, `"artifact_label": "Itinerary"`} {
		if !strings.Contains(threadState, want) {
			t.Fatalf("expected resume to persist artifact focus %q, got:\n%s", want, threadState)
		}
	}
}

func TestCurrentWorkInteractiveContinueOpensNextUsefulSurface(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("continue\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") {
		t.Fatalf("expected continue to open next useful surface, got:\n%s", out)
	}
}

func TestCurrentWorkInteractiveLauncherIsCompactByDefault(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(""), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Current work",
		"Weekly Product Review",
		"Paste a new request, or type `help` to inspect current work.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Goal",
		"Working with",
		"AI route",
		"Up next",
		"Ready now",
		"Actions",
		"Need",
		"Why this matters",
		"Options",
		"If you skip this",
		"Just finished",
		"Doing now",
		"Safe to do",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected compact launcher not to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkPromptShowsResumeHintAfterFocusChanges(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected selection to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(""), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Resume",
		"Owners and Due Points",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected prompt to contain %q after focus change, got:\n%s", want, out)
		}
	}
}

func TestCurrentWorkInteractiveMissingChoiceShowsGapSummary(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("skip for now\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected blocking ask to be skippable, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("missing\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Still missing",
		"Metric and legal-review decision",
		"Not sure about",
		"Next step",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCurrentWorkInteractiveContinueOpensNextUsefulMeetingSurface(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Continue\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") {
		t.Fatalf("expected continue to open next useful surface, got:\n%s", out)
	}
}

func TestCurrentWorkInteractiveExpandOpensRicherSurface(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("expand\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") {
		t.Fatalf("expected fuller action to open richer surface, got:\n%s", out)
	}
}

func TestCurrentWorkReadyShelfSelectionOpensArtifactByNumber(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Open",
		"2. Owners and Due Points",
		"Owners and Due Points",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCurrentWorkReadyShelfSelectionOpensArtifactByName(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("open\nOwners and Due Points\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Open",
		"Owners and Due Points",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCurrentWorkResumeReopensSelectedReadyArtifact(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected selection to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("resume\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") {
		t.Fatalf("expected resume to reopen selected artifact, got:\n%s", out)
	}
}

func TestOpenCommandUpdatesFocusedArtifactForResume(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var openOut bytes.Buffer
	exitCode := app.RunInteractive([]string{"open", "Owners and Due Points"}, nil, &openOut, &openOut)
	if exitCode != 0 {
		t.Fatalf("expected open command to succeed, got %d with output:\n%s", exitCode, openOut.String())
	}
	if !strings.Contains(openOut.String(), "Owners and Due Points") {
		t.Fatalf("expected open command to render named artifact surface, got:\n%s", openOut.String())
	}

	var stdout bytes.Buffer
	exitCode = app.RunInteractive(nil, strings.NewReader("resume\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected resume after open command to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Owners and Due Points") {
		t.Fatalf("expected resume to reopen artifact selected by open command, got:\n%s", stdout.String())
	}
}

func TestCurrentWorkResumeReopensMissingSurface(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("skip for now\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected blocking ask to be skippable, got %d", exitCode)
	}
	if exitCode := app.RunInteractive(nil, strings.NewReader("missing\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected missing view to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("resume\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Still missing") {
		t.Fatalf("expected resume to reopen missing surface, got:\n%s", out)
	}
}

func TestCurrentWorkShowMissingWithActiveAskOpensDecisionSurface(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("Weekly product review for pricing launch. Need owners, due dates, and open questions.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected starter work to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("missing\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected active ask surface to open, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{"Pending decision", "Options", "If you skip this"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected active ask surface to contain %q, got:\n%s", want, out)
		}
	}

	current := readCurrentWork(t, stateDir)
	threadState := mustReadFile(t, filepath.Join(current["pack_dir"].(string), "thread-state.json"))
	if !strings.Contains(threadState, "\"kind\": \"ask\"") {
		t.Fatalf("expected current focus to move to ask surface, got:\n%s", threadState)
	}
}

func TestCurrentWorkResumeReopensDecisionSurface(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("Notifications PRD needs a build-readiness check and handoff call.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected starter work to succeed, got %d", exitCode)
	}
	if exitCode := app.RunInteractive(nil, strings.NewReader("missing\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected active ask surface to open, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("resume\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected resume to reopen active ask surface, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Pending decision") {
		t.Fatalf("expected resume to reopen pending decision, got:\n%s", out)
	}
}

func TestCurrentWorkTransformUsesFocusedArtifact(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected selection to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("shorter\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected focused transform to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Owners and Due Points") || !strings.Contains(out, "## Short version") {
		t.Fatalf("expected focused transform to target selected artifact, got:\n%s", out)
	}

	followup := mustReadFile(t, filepath.Join(packDir, "views", "followup.md"))
	if strings.Contains(followup, "## Short version") {
		t.Fatalf("expected primary follow-up artifact to remain unchanged, got:\n%s", followup)
	}
	owners := mustReadFile(t, filepath.Join(packDir, "views", "tasks.md"))
	if !strings.Contains(owners, "## Short version") {
		t.Fatalf("expected selected artifact to be rewritten, got:\n%s", owners)
	}
}

func TestCurrentWorkFocusedOutcomeUsesSelectedArtifact(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	writeCurrentWorkRoute(t, packDir, map[string]any{
		"tool_mode":       "local-workhorse",
		"provider_label":  "Local SLM",
		"feedback_key":    "local-workhorse",
		"route_policy":    "auto",
		"provider_mode":   "local-slm",
		"artifact_family": "narrative-draft",
	})

	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected selection to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("share\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected focused outcome to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	routeSaved := mustReadFile(t, filepath.Join(packDir, "route.json"))
	if !strings.Contains(routeSaved, "tasks.md") {
		t.Fatalf("expected route feedback path to point at focused artifact, got:\n%s", routeSaved)
	}
	threadState := mustReadFile(t, filepath.Join(packDir, "thread-state.json"))
	if !strings.Contains(threadState, "Share") {
		t.Fatalf("expected thread state to record shared decision, got:\n%s", threadState)
	}
}

func TestCurrentWorkCanResolveBlockingAskBySkipping(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("Weekly product review for pricing launch. Need owners, due dates, and open questions.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected starter work to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("skip for now\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected ask skip to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Recorded decision: Skipped for now.") {
		t.Fatalf("expected skip confirmation, got:\n%s", stdout.String())
	}

	current := readCurrentWork(t, stateDir)
	threadState := mustReadFile(t, filepath.Join(current["pack_dir"].(string), "thread-state.json"))
	if strings.Contains(threadState, "\"active_ask\"") {
		t.Fatalf("expected active ask to be cleared, got:\n%s", threadState)
	}
	if !strings.Contains(threadState, "Skipped for now") {
		t.Fatalf("expected thread state to record skip decision, got:\n%s", threadState)
	}
}

func TestCurrentWorkCanResolveBlockingAskByApproving(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("Notifications PRD needs a build-readiness check and handoff call.\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected starter work to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("approved\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected approval decision to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Recorded decision: Approved.") {
		t.Fatalf("expected approval confirmation, got:\n%s", stdout.String())
	}

	current := readCurrentWork(t, stateDir)
	threadState := mustReadFile(t, filepath.Join(current["pack_dir"].(string), "thread-state.json"))
	if strings.Contains(threadState, "\"active_ask\"") {
		t.Fatalf("expected active ask to be cleared, got:\n%s", threadState)
	}
	if !strings.Contains(threadState, "Approved") {
		t.Fatalf("expected thread state to record approval, got:\n%s", threadState)
	}
}

func TestCurrentWorkInteractiveSocialAckDoesNotStartNewWork(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("thanks\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Nothing changed.",
		"Use `Continue`, `Open`, or paste a new request.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Working Draft: Thanks") {
		t.Fatalf("expected thanks not to start literal work, got:\n%s", out)
	}
	current := readCurrentWork(t, stateDir)
	if current["title"] != "Weekly Product Review" {
		t.Fatalf("expected current work to remain unchanged, got %#v", current)
	}
}

func TestCurrentWorkInteractiveDoctorDoesNotStartNewWork(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("doctor\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{"Provider", "Status"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Working Draft: Doctor") {
		t.Fatalf("expected doctor not to start literal work, got:\n%s", out)
	}
	current := readCurrentWork(t, stateDir)
	if current["title"] != "Weekly Product Review" {
		t.Fatalf("expected current work to remain unchanged, got %#v", current)
	}
}

func TestCurrentWorkInteractiveTacticalCommandsDoNotStartNewWork(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "status",
			line: "status\n",
			want: []string{"Goal", "Weekly Product Review", "Ready now", "Sendable Follow-up"},
		},
		{
			name: "memory",
			line: "memory\n",
			want: []string{"Memory", "Current work is saved: Weekly Product Review."},
		},
		{
			name: "route",
			line: "route\n",
			want: []string{"Route and cost", "Least-expense capable route"},
		},
		{
			name: "permissions",
			line: "permissions\n",
			want: []string{"Permissions", "Nothing has been sent, published, booked, or changed."},
		},
		{
			name: "clear",
			line: "clear\n",
			want: []string{"Nothing was deleted.", "Type `Start` to switch focus without removing this work."},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			packDir := seedMeetingWork(t)
			writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(tc.line), &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			if strings.Contains(out, "Working Draft") {
				t.Fatalf("expected tactical command not to start literal work, got:\n%s", out)
			}
			current := readCurrentWork(t, stateDir)
			if current["title"] != "Weekly Product Review" {
				t.Fatalf("expected current work to remain unchanged, got %#v", current)
			}
		})
	}
}

func TestPostResultStatusCommandShowsFullState(t *testing.T) {
	source := "Weekly product review. Need owners, due dates, and open questions."
	out := runInteractiveForTest(t, t.TempDir(), source+"\nstatus\n")

	for _, want := range []string{
		"Your first draft is ready.",
		"Goal",
		"Weekly Product Review Need Owners",
		"Ready now",
		"Sendable Follow-up",
		"Safe to do",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Working Draft: Status") {
		t.Fatalf("expected status not to start literal work, got:\n%s", out)
	}
}

func TestInteractiveTacticalSurfacesStayCompact(t *testing.T) {
	cases := []struct {
		name        string
		arg         string
		maxNonEmpty int
	}{
		{name: "status", arg: "status", maxNonEmpty: 3},
		{name: "memory", arg: "memory", maxNonEmpty: 3},
		{name: "permissions", arg: "permissions", maxNonEmpty: 3},
		{name: "route", arg: "route", maxNonEmpty: 3},
		{name: "help", arg: "help", maxNonEmpty: 16},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive([]string{tc.arg}, nil, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			out := stdout.String()
			if got := nonEmptyLineCount(out); got > tc.maxNonEmpty {
				t.Fatalf("expected %s to stay within %d non-empty lines, got %d:\n%s", tc.name, tc.maxNonEmpty, got, out)
			}
		})
	}
}

func TestCurrentWorkTacticalSurfacesStayCompact(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		maxLines int
	}{
		{name: "fuller", input: "Make it fuller\n", maxLines: 14},
		{name: "ready shelf", input: "open\n", maxLines: 16},
		{name: "interrupt prompt", input: "plan me a 7 day paris trip\n", maxLines: 14},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			packDir := seedMeetingWork(t)
			writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(tc.input), &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			if got := nonEmptyLineCount(stdout.String()); got > tc.maxLines {
				t.Fatalf("expected %s to stay within %d non-empty lines, got %d:\n%s", tc.name, tc.maxLines, got, stdout.String())
			}
		})
	}
}

func TestInteractiveTacticalSurfacesStayWithinLatencySmokeBudget(t *testing.T) {
	budget := 500 * time.Millisecond
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "top-level status",
			run: func(t *testing.T) {
				t.Helper()
				stateDir := t.TempDir()
				t.Setenv("JINI_STATE_DIR", stateDir)

				var stdout bytes.Buffer
				if exitCode := app.RunInteractive([]string{"status"}, nil, &stdout, &stdout); exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
				}
			},
		},
		{
			name: "top-level memory",
			run: func(t *testing.T) {
				t.Helper()
				stateDir := t.TempDir()
				t.Setenv("JINI_STATE_DIR", stateDir)

				var stdout bytes.Buffer
				if exitCode := app.RunInteractive([]string{"memory"}, nil, &stdout, &stdout); exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
				}
			},
		},
		{
			name: "current work ready shelf",
			run: func(t *testing.T) {
				t.Helper()
				stateDir := t.TempDir()
				packDir := seedMeetingWork(t)
				writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
				t.Setenv("JINI_STATE_DIR", stateDir)

				var stdout bytes.Buffer
				if exitCode := app.RunInteractive(nil, strings.NewReader("open\n"), &stdout, &stdout); exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
				}
			},
		},
		{
			name: "current work fuller surface",
			run: func(t *testing.T) {
				t.Helper()
				stateDir := t.TempDir()
				packDir := seedMeetingWork(t)
				writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
				t.Setenv("JINI_STATE_DIR", stateDir)

				var stdout bytes.Buffer
				if exitCode := app.RunInteractive(nil, strings.NewReader("expand\n"), &stdout, &stdout); exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
				}
			},
		},
		{
			name: "current work interruption prompt",
			run: func(t *testing.T) {
				t.Helper()
				stateDir := t.TempDir()
				packDir := seedMeetingWork(t)
				writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
				t.Setenv("JINI_STATE_DIR", stateDir)

				var stdout bytes.Buffer
				if exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\n"), &stdout, &stdout); exitCode != 0 {
					t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			tc.run(t)
			if elapsed := time.Since(start); elapsed > budget {
				t.Fatalf("expected %s to finish within %s, took %s", tc.name, budget, elapsed)
			}
		})
	}
}

func TestTopLevelSlashAliasesAreRejected(t *testing.T) {
	cases := [][]string{
		{"/help"},
		{"/doctor"},
		{"/status"},
		{"/init"},
		{"/memory"},
		{"/permissions"},
		{"/route"},
		{"/new"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(args, nil, &stdout, &stdout)
			if exitCode != 1 {
				t.Fatalf("expected exit code 1 for %v, got %d with output:\n%s", args, exitCode, stdout.String())
			}
			if !strings.Contains(stdout.String(), "Unknown command") {
				t.Fatalf("expected slash alias to be rejected, got:\n%s", stdout.String())
			}
		})
	}
}

func TestCurrentWorkCanEnterPlanFirstMode(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedResearchPRDWork(t)
	writeCurrentWork(t, stateDir, packDir, "research-prd", "example-research-prd", "Jini Research To PRD", "awaiting_verification", "ready-to-verify")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Plan\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Plan",
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

	stdin := strings.NewReader("7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"7 Day Paris Trip",
		"Itinerary",
		"# Itinerary: 7 Day Paris Trip",
		"Budget Sketch",
		"## Still to confirm",
		"Must Do Anchors, Or Whether You Want Help Choosing Them",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	assertNoFirstRunStatusDump(t, out)

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected travel-plan current work, got %#v", current)
	}
}

func TestInteractiveLauncherAsksForTravelClarificationWhenUnderspecified(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("7 day paris trip\ncouple, around $2500, early October, mixed pace, central hotel area, Versailles optional\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- travelers",
		"- budget range",
		"- dates or season",
		"- pace or style",
		"- base area, or whether you want help choosing one",
		"- must-do anchors, or whether you want help choosing them",
		"Type `skip` if you want a generic first draft.",
		"Clarified scope",
		"7 Day Paris Trip",
		"Itinerary",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "one museum and one day trip are must-dos") {
		t.Fatalf("expected shared scope-planner example, got:\n%s", out)
	}
	assertNoFirstRunStatusDump(t, out)
}

func TestInteractiveLauncherSkipsTravelClarificationWhenRequestAlreadyScoped(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("7 day Paris trip for a couple with a $2500 budget in early October, mixed pace, central hotel area\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if strings.Contains(out, "Before I draft it, help me narrow the highest-impact details in one line:") {
		t.Fatalf("expected already-scoped travel request to skip clarification, got:\n%s", out)
	}
}

func TestInteractiveLauncherAsksOnlyForMissingTravelDimensions(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("7 day Paris trip for a couple with a $2500 budget\nearly October, mixed pace, central hotel area, Louvre and Versailles are must-dos\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- dates or season",
		"- pace or style",
		"- base area, or whether you want help choosing one",
		"- must-do anchors, or whether you want help choosing them",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected targeted clarification to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"- travelers",
		"- budget range",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected targeted clarification to omit %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveLauncherCreatesNonParisTravelWithoutParisFallbacks(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("5 day Rome trip\ncouple, around $2500, early October, mixed pace, central stay, Colosseum is a must-do\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"5 Day Rome Trip",
		"Colosseum",
		"### Day 5: Buffer and departure",
		"- Nothing right now",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Whether Versailles") || strings.Contains(out, "Louvre and Versailles are must-dos") {
		t.Fatalf("expected non-Paris trip to avoid Paris-specific fallback, got:\n%s", out)
	}
	assertNoFirstRunStatusDump(t, out)
}

func TestInteractiveLauncherCreatesLongerScopedTravelWithExactDayCount(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("12 day Spain trip for a family with kids, $5000 budget, early June, mixed pace, central stays, Barcelona and Granada are must-dos\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"12 Day Spain Trip",
		"### Day 12: Buffer and departure",
		"Barcelona and Granada",
		"- Nothing right now",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Whether Versailles") || strings.Contains(out, "Louvre and Versailles are must-dos") {
		t.Fatalf("expected Spain trip to avoid Paris-specific fallback, got:\n%s", out)
	}
	if strings.Contains(out, "Before I draft it, help me narrow the highest-impact details in one line:") {
		t.Fatalf("expected fully scoped travel request to skip clarification, got:\n%s", out)
	}
	assertNoFirstRunStatusDump(t, out)
}

func TestInteractiveLauncherAsksForBuildReadinessClarificationWhenUnderspecified(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("check whether this plan is ready to hand off\nnotifications PRD, rollback is still open, approval owner is Priya\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- which plan or feature this is for",
		"- the first slice or decision this handoff should cover",
		"- known blockers, risks, or open gaps",
		"- approval owner or review owner",
		"Type `skip` if you want a first pass with the gaps called out.",
		"Build-Readiness Check",
		"Clarified scope",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestInteractiveLauncherDraftsBuildReadinessWhenOnlyOneDimensionIsMissing(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("notifications PRD needs a build-readiness check, rollback is still open, approval owner is Priya\nfirst slice is digest emails\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, unwanted := range []string{
		"Before I draft it, help me narrow the highest-impact details in one line:",
		"- the first slice or decision this handoff should cover",
		"Clarified scope",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected single-gap build-readiness request to draft immediately without %q, got:\n%s", unwanted, out)
		}
	}
	if !strings.Contains(out, "Build-Readiness Check") {
		t.Fatalf("expected build-readiness draft, got:\n%s", out)
	}
}

func TestInteractiveLauncherSkipsBuildReadinessClarificationWhenRequestAlreadyScoped(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("Notifications PRD needs a build-readiness check for digest emails, rollback is still open, approval owner is Priya\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if strings.Contains(out, "Before I draft it, help me narrow the highest-impact details in one line:") {
		t.Fatalf("expected already-scoped build-readiness request to skip clarification, got:\n%s", out)
	}
	if !strings.Contains(out, "Build-Readiness Check") {
		t.Fatalf("expected build-readiness draft, got:\n%s", out)
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
		"Jini",
		"Current work",
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
	exitCode := app.RunInteractive(nil, strings.NewReader("switch\n1\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Switch",
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

func TestCurrentWorkFreeformInputConfirmsBeforeStartingNewWork(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\nstart\ncouple, around $2500, early October, mixed pace, central hotel area\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"New work",
		"Current:",
		"Start",
		"Your first draft is ready.",
		"7 Day Paris Trip",
		"Itinerary",
		"Budget Sketch",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Weekly Product Review Follow-up") && !strings.Contains(out, "7 Day Paris Trip") {
		t.Fatalf("expected new work to replace the old current-work view, got:\n%s", out)
	}
	assertNoFirstRunStatusDump(t, out)

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected current work to switch to travel-plan, got %#v", current)
	}
}

func TestCurrentWorkFreeformInputCanKeepCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\nkeep\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"New work",
		"Keeping current work.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "7 Day Paris Trip") {
		t.Fatalf("expected keep path not to start new work, got:\n%s", out)
	}
	if strings.Contains(out, "Switch") {
		t.Fatalf("expected interrupt prompt to hide switch when no other work exists, got:\n%s", out)
	}
	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "meeting-followup" {
		t.Fatalf("expected current work to remain meeting-followup, got %#v", current)
	}
}

func TestCurrentWorkInterruptPromptErrorMatchesAvailableChoices(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\nmaybe later\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected invalid interruption choice to return exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Choose `Start` or `Keep`.") {
		t.Fatalf("expected guidance to match shown choices, got:\n%s", out)
	}
	if strings.Contains(out, "Switch project") || strings.Contains(out, "Switch") {
		t.Fatalf("expected invalid-choice guidance not to mention hidden switch-project option, got:\n%s", out)
	}
}

func TestCurrentWorkFreeformInputCanSwitchProjectFromInterruptPrompt(t *testing.T) {
	stateDir := t.TempDir()
	travelDir := copyWorkDir(t, filepath.Join(stateDir, "work", "travel-plan-paris"), seedTravelWork(t))
	meetingDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\nswitch\n1\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"New work",
		"Switch",
		"Switched to",
		"7-Day Paris Trip",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" || current["pack_dir"] != travelDir {
		t.Fatalf("expected current work to switch to travel plan, got %#v", current)
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
		"meeting-notes.txt (processed)",
		"Sendable Follow-up",
		"## Send this",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	assertNoFirstRunStatusDump(t, out)
}

func TestPostResultCanShowWhatJiniUsed(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "7 day paris trip\ncouple, around $2500, early October, mixed pace, central hotel area, Versailles optional\ncontext\n")

	for _, want := range []string{
		"Context",
		"From you",
		"Your request: 7 day paris trip",
		"Clarified scope: couple, around $2500, early October, mixed pace, central hotel area, Versailles optional",
		"Kept visible",
		"Route and continuity",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestResumeHintUsesContextLabelAfterOpeningContextSurface(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("7 day paris trip\ncouple, around $2500, early October, mixed pace, central hotel area, Versailles optional\ncontext\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected context action to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(""), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected prompt after context focus to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Context") {
		t.Fatalf("expected prompt resume hint to use Context label, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "What Jini used") {
		t.Fatalf("expected prompt not to use older context label, got:\n%s", stdout.String())
	}
}

func TestPostResultMenuUsesShowContextLabel(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "7 day paris trip\ncouple, around $2500, early October, mixed pace, central hotel area, Versailles optional\n")

	if !strings.Contains(out, "Context") {
		t.Fatalf("expected post-result menu to use Context label, got:\n%s", out)
	}
	if strings.Contains(out, "Show what Jini used") {
		t.Fatalf("expected post-result menu to avoid older product-teaching label, got:\n%s", out)
	}
}

func TestPostResultCanMakeItShorterAndUndoLastChange(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "Weekly product review for pricing launch. Need owners, due dates, and open questions.\nshorter\n")

	for _, want := range []string{
		"Sendable Follow-up",
		"## Short version",
		"## Still to confirm",
		"## Next move",
		"Saved a restorable version.",
		"Versions",
		"Undo",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected shorter transform output to contain %q, got:\n%s", want, out)
		}
	}

	current := readCurrentWork(t, stateDir)
	followupPath := filepath.Join(current["pack_dir"].(string), "views", "followup.md")
	shortened := mustReadFile(t, followupPath)
	if !strings.Contains(shortened, "## Short version") {
		t.Fatalf("expected transformed artifact to be saved, got:\n%s", shortened)
	}

	versionsOut := runInteractiveForTest(t, stateDir, "versions\n")
	for _, want := range []string{
		"Versions",
		"Shorter",
		"Undo",
	} {
		if !strings.Contains(versionsOut, want) {
			t.Fatalf("expected versions output to contain %q, got:\n%s", want, versionsOut)
		}
	}

	undoOut := runInteractiveForTest(t, stateDir, "undo\n")
	for _, want := range []string{
		"Restored the previous version.",
		"## Send this note",
	} {
		if !strings.Contains(undoOut, want) {
			t.Fatalf("expected undo output to contain %q, got:\n%s", want, undoOut)
		}
	}

	restored := mustReadFile(t, followupPath)
	if strings.Contains(restored, "## Short version") {
		t.Fatalf("expected undo to restore original artifact, got:\n%s", restored)
	}
	if !strings.Contains(restored, "## Send this note") {
		t.Fatalf("expected original artifact content after undo, got:\n%s", restored)
	}
}

func TestPostResultCanMakeItExecutive(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "Notifications PRD needs a build-readiness check and handoff call.\nexecutive\n")

	for _, want := range []string{
		"Build-Readiness Check",
		"## Executive summary",
		"## Risks or gaps",
		"## Next move",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected executive transform output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPostResultUsesSharedStructuredCheckTransformProfileForShorter(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "Notifications PRD needs a build-readiness check and handoff call.\nshorter\n")

	for _, want := range []string{
		"Build-Readiness Check",
		"## What looks ready now",
		"## Must clear before build",
		"## Recommended first slice",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected structured-check shorter output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPostResultCanTurnArtifactIntoChecklist(t *testing.T) {
	stateDir := t.TempDir()
	out := runInteractiveForTest(t, stateDir, "Weekly product review for pricing launch. Need owners, due dates, and open questions.\nchecklist\n")

	for _, want := range []string{
		"Sendable Follow-up",
		"## Do now",
		"- [ ]",
		"## Confirm",
		"## Watch",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected checklist transform output to contain %q, got:\n%s", want, out)
		}
	}
}

func assertNoFirstRunStatusDump(t *testing.T, out string) {
	t.Helper()
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nSafe to do\n",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected first-run output not to dump full status frame %q, got:\n%s", unwanted, out)
		}
	}
}

func writeCurrentWork(t testing.TB, stateDir, packDir, packID, workUnitID, title, state, health string) {
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

func nonEmptyLineCount(out string) int {
	count := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func BenchmarkInteractiveNoWorkHelp(b *testing.B) {
	stateDir := b.TempDir()
	b.Setenv("JINI_STATE_DIR", stateDir)
	b.ReportAllocs()

	for b.Loop() {
		var stdout bytes.Buffer
		if exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, io.Discard); exitCode != 0 {
			b.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
		}
	}
}

func BenchmarkInteractiveCurrentWorkReadyShelf(b *testing.B) {
	stateDir := b.TempDir()
	packDir := seedMeetingWork(b)
	writeCurrentWork(b, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	b.Setenv("JINI_STATE_DIR", stateDir)
	b.ReportAllocs()

	for b.Loop() {
		var stdout bytes.Buffer
		if exitCode := app.RunInteractive(nil, strings.NewReader("open\n"), &stdout, io.Discard); exitCode != 0 {
			b.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
		}
	}
}

func BenchmarkInteractiveCurrentWorkOpenShelfSelection(b *testing.B) {
	stateDir := b.TempDir()
	packDir := seedMeetingWork(b)
	writeCurrentWork(b, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	b.Setenv("JINI_STATE_DIR", stateDir)
	b.ReportAllocs()

	for b.Loop() {
		var stdout bytes.Buffer
		if exitCode := app.RunInteractive(nil, strings.NewReader("open\n2\n"), &stdout, io.Discard); exitCode != 0 {
			b.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
		}
	}
}

func BenchmarkInteractiveCurrentWorkFuller(b *testing.B) {
	stateDir := b.TempDir()
	packDir := seedMeetingWork(b)
	writeCurrentWork(b, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	b.Setenv("JINI_STATE_DIR", stateDir)
	b.ReportAllocs()

	for b.Loop() {
		var stdout bytes.Buffer
		if exitCode := app.RunInteractive(nil, strings.NewReader("expand\n"), &stdout, io.Discard); exitCode != 0 {
			b.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
		}
	}
}

func BenchmarkInteractiveCurrentWorkInterruptionPrompt(b *testing.B) {
	stateDir := b.TempDir()
	packDir := seedMeetingWork(b)
	writeCurrentWork(b, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	b.Setenv("JINI_STATE_DIR", stateDir)
	b.ReportAllocs()

	for b.Loop() {
		var stdout bytes.Buffer
		if exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip\n"), &stdout, io.Discard); exitCode != 0 {
			b.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
		}
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

func seedMeetingWork(t testing.TB) string {
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

func seedTravelWork(t testing.TB) string {
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

func mkdirAll(t testing.TB, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func writeFile(t testing.TB, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
