package app_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		"You're working on",
		"Jini Research To PRD",
		"Working with",
		"Local preview",
		"Jini is using",
		"Latest PRD draft and review comments",
		"Ready now",
		"Handoff Brief",
		"Build-Readiness Check",
		"Still missing",
		"Approval",
		"Next step",
		"Open Build-Readiness Check",
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
		"Want a connected provider instead?",
		"Type `Use Claude`",
		"Type `Use Bedrock`",
		"Type `Use Azure`",
		"Type `Use Auto`",
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
		"You're working on",
		"Weekly Product Review",
		"Ready now",
		"Sendable Follow-up",
		"Next step",
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
		"Still missing",
		"Metric and legal-review decision",
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
		"Still missing",
		"Dates, budget, and hotel area",
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
		"Jini shell",
		"Paste messy notes, or type the outcome you want.",
		"Turn meeting notes into something I can send",
		"Check whether a plan is ready to hand off",
		"I am not sure",
		"Plan this first",
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
		"Jini shell",
		"Sendable Follow-up",
		"## Send this",
		"Keep going",
		"Make it fuller",
		"See what is still missing",
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
	if strings.Contains(out, "You're working on") && strings.Index(out, "## Send this") > strings.Index(out, "You're working on") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "meeting-followup" {
		t.Fatalf("expected meeting-followup current work, got %#v", current)
	}
	followupPath := filepath.Join(current["pack_dir"].(string), "views", "followup.md")
	content := mustReadFile(t, followupPath)
	if !strings.Contains(content, "pricing launch") {
		t.Fatalf("expected followup to carry source context, got:\n%s", content)
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
		"Jini shell",
		"Build-Readiness Check",
		"## Build-readiness check",
		"Keep going",
		"Make it fuller",
		"See what is still missing",
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
	if strings.Contains(out, "You're working on") && strings.Index(out, "## Build-readiness check") > strings.Index(out, "You're working on") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}

	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "research-prd" {
		t.Fatalf("expected research-prd current work, got %#v", current)
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
	if strings.Contains(out, "You're working on") && strings.Index(out, "First Useful Pass") > strings.Index(out, "You're working on") {
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
			action:          "See what is still missing",
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
				"See what is still missing",
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
			action:          "See what is still missing",
			wantAfterAction: "Approval",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := "Notifications PRD needs a build-readiness check and handoff call."
			baseline := runInteractiveForTest(t, t.TempDir(), source+"\n")

			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+tc.action+"\n")
			for _, want := range []string{
				"## Build-readiness check",
				"Keep going",
				"See what is still missing",
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
	exitCode := app.RunInteractive(nil, strings.NewReader("Open ready work\n"), &stdout, &stdout)
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
	exitCode := app.RunInteractive(nil, strings.NewReader("Plan this first\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Plan this first",
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
		"You're working on",
		"7 Day Paris Trip For A Couple With A 2500 Budget",
		"Ready now",
		"Itinerary",
		"Budget Sketch",
		"Still missing",
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
