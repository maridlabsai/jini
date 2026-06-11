package app_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maridlabsai/jini/internal/app"
)

func TestDirectTaskArgumentsStartNativeIntake(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		workingOn    string
		startWith    string
		expectedPack string
	}{
		{
			name:         "fix tests",
			args:         []string{"fix", "failing", "tests"},
			workingOn:    "Working on: fix failing tests",
			startWith:    "Start with `make test`.",
			expectedPack: "general-work",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.Run(tc.args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			out := stdout.String()
			for _, want := range []string{
				tc.workingOn,
				tc.startWith,
				"Saved. Use `jini status` for full status or `jini open` for the draft.",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{
				"Jini",
				"Paste what you want finished.",
				"Your first draft is ready.",
				"Working Draft",
				"Task Snapshot",
				"Actions",
				">",
			} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected direct task output to avoid %q, got:\n%s", unwanted, out)
				}
			}
			if got := nonEmptyLineCount(out); got > 4 {
				t.Fatalf("expected direct task output to stay compact, got %d non-empty lines:\n%s", got, out)
			}

			current := readCurrentWork(t, stateDir)
			if current["pack_id"] != tc.expectedPack {
				t.Fatalf("expected direct task to preserve saved work in %s, got %#v", tc.expectedPack, current)
			}
		})
	}
}

func TestDirectRepoReviewPrintsModelFreeSnapshotWithoutSavedWorkflow(t *testing.T) {
	stateDir := t.TempDir()
	repoDir := t.TempDir()
	t.Chdir(repoDir)
	runGitForTest(t, repoDir, "init")
	runGitForTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForTest(t, repoDir, "config", "user.name", "Test User")
	writeFile(t, filepath.Join(repoDir, "go.mod"), "module example.com/reviewme\n")
	writeFile(t, filepath.Join(repoDir, "main.go"), "package main\n\nfunc main() {}\n")
	runGitForTest(t, repoDir, "add", ".")
	runGitForTest(t, repoDir, "commit", "-m", "initial")
	writeFile(t, filepath.Join(repoDir, "main.go"), "package main\n\nfunc main() { println(\"changed\") }\n")
	writeFile(t, filepath.Join(repoDir, "README.md"), "# Review Me\n")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"review", "this", "repo"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Repo review",
		"- Workspace: " + filepath.Base(repoDir),
		"- Changed files: 1",
		"- Untracked files: 1",
		"Changed:",
		"- main.go",
		"Untracked:",
		"- README.md",
		"Next:",
		"- git diff --stat",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Working on:",
		"Task Snapshot",
		"Artifact created.",
		"Saved.",
		"Open the saved draft",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct repo review to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected direct repo review not to create current work, stat error: %v", err)
	}
}

func TestInteractiveRepoReviewInspectsCurrentGitRepoDirectly(t *testing.T) {
	stateDir := t.TempDir()
	repoDir := t.TempDir()
	t.Chdir(repoDir)
	runGitForTest(t, repoDir, "init")
	runGitForTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForTest(t, repoDir, "config", "user.name", "Test User")
	writeFile(t, filepath.Join(repoDir, "go.mod"), "module example.com/reviewme\n")
	runGitForTest(t, repoDir, "add", ".")
	runGitForTest(t, repoDir, "commit", "-m", "initial")
	writeFile(t, filepath.Join(repoDir, "go.mod"), "module example.com/reviewme\n\ngo 1.26\n")

	out := runInteractiveForTest(t, stateDir, "review this repo for uncommitted changes\n")
	for _, want := range []string{
		"Repo review",
		"- Workspace: " + filepath.Base(repoDir),
		"- Changed files: 1",
		"Changed:",
		"- go.mod",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected interactive repo review to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Task Snapshot", "Artifact created.", "Saved artifact:", "Open the saved draft"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected interactive repo review to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestRepoReviewOutsideGitRepoSaysSoDirectly(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)

	out := runInteractiveForTest(t, stateDir, "review this repo for uncommitted changes\n")
	if !strings.Contains(out, "This folder is not a git repo.") {
		t.Fatalf("expected non-repo prompt to say so directly, got:\n%s", out)
	}
	for _, unwanted := range []string{"Task Snapshot", "Artifact created.", "Saved artifact:"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected non-repo prompt to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)
	target := filepath.Join(workDir, "pear fellow script.txt")
	writeFile(t, target, "intro\n")
	if err := os.Chmod(target, 0o600); err != nil {
		t.Fatalf("chmod target: %v", err)
	}

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("add a line saying \"jini was here\" in the pear fellow script .txt file in this folder\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Updated pear fellow script.txt",
		"- Added line: jini was here",
		"- Location: " + target,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Your first draft is ready.",
		"Working Draft",
		"Name the audience or recipient",
		"Saved. Type status",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected local edit to avoid draft flow %q, got:\n%s", unwanted, out)
		}
	}
	if got := mustReadFile(t, target); got != "intro\njini was here\n" {
		t.Fatalf("expected target file to be updated, got:\n%s", got)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected target mode to be preserved, got %v", got)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected local file edit not to create current work, stat error: %v", err)
	}
}

func TestCurrentWorkLocalTextEditExecutesWithoutStartPrompt(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)
	target := filepath.Join(workDir, "pear vc script.txt")
	writeFile(t, target, "intro\n")
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Add A Line Saying Jini Was Here In The Pear Vc Script Txt File In This Folder", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("Add A Line Saying Jini Was Here In The Pear Vc Script Txt File In This Folder\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Updated pear vc script.txt",
		"- Added line: jini was here",
		"- Location: " + target,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"New work",
		"Actions",
		"- Start",
		"Your first draft is ready.",
		"Working Draft",
		"Saved. Type status",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct edit to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if got := mustReadFile(t, target); got != "intro\njini was here\n" {
		t.Fatalf("expected target file to be updated, got:\n%s", got)
	}
}

func TestCurrentWorkLocalTextEditPreservesUnquotedLineContainingIn(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)
	target := filepath.Join(workDir, "notes.txt")
	writeFile(t, target, "intro\n")
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("add a line saying check in at 9 in notes.txt\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if got := mustReadFile(t, target); got != "intro\ncheck in at 9\n" {
		t.Fatalf("expected unquoted line with in to be preserved, got:\n%s", got)
	}
	if strings.Contains(stdout.String(), "Working Draft") {
		t.Fatalf("expected direct edit to avoid draft flow, got:\n%s", stdout.String())
	}
}

func TestInteractiveFollowupEmailReturnsSendableEmailWithoutArtifactShell(t *testing.T) {
	stateDir := t.TempDir()

	out := runInteractiveForTest(t, stateDir, "write a follow-up email summarizing today standup: shipped login fix, analytics blocked, Sara owns QA by Friday\n")
	for _, want := range []string{
		"Subject: Standup follow-up",
		"Hi team,",
		"Shipped login fix.",
		"Analytics is blocked.",
		"Sara owns QA by Friday.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected sendable email to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Artifact created.",
		"Task Snapshot",
		"Sendable Follow-up",
		"Saved artifact:",
		"Next: `jini",
		"Also ready:",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct email to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected direct email not to create current work, stat error: %v", err)
	}
}

func TestDirectArgumentFollowupEmailWithHyphenIsTreatedAsPrompt(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"write a follow-up email summarizing today standup: shipped login fix"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected single-argument prompt to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Subject: Standup follow-up") || !strings.Contains(out, "Shipped login fix.") {
		t.Fatalf("expected single-argument prompt to render email, got:\n%s", out)
	}
	if strings.Contains(out, "Unknown command") || strings.Contains(out, "Run `jini commands`") {
		t.Fatalf("expected single-argument prompt not to be rejected as command, got:\n%s", out)
	}
}

func TestDirectArgsLocalTextEditAppendsQuotedLine(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)
	target := filepath.Join(workDir, "pear fellow script.txt")
	writeFile(t, target, "intro")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"add", "a", "line", "saying", "\"jini was here\"", "in", "the", "pear", "fellow", "script", ".txt", "file", "in", "this", "folder"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Updated pear fellow script.txt") {
		t.Fatalf("expected direct args edit confirmation, got:\n%s", stdout.String())
	}
	if got := mustReadFile(t, target); got != "intro\njini was here\n" {
		t.Fatalf("expected target file to be updated with newline preservation, got:\n%s", got)
	}
}

func TestLocalTextEditDoesNotGuessAmongAmbiguousTextFiles(t *testing.T) {
	stateDir := t.TempDir()
	workDir := t.TempDir()
	t.Chdir(workDir)
	first := filepath.Join(workDir, "pear first.txt")
	second := filepath.Join(workDir, "pear second.txt")
	writeFile(t, first, "first\n")
	writeFile(t, second, "second\n")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("add a line saying \"jini was here\" in the pear .txt file in this folder\n"), &stdout, &stdout)
	if exitCode == 0 {
		t.Fatalf("expected ambiguous local edit to fail safely, got output:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Multiple .txt files match. Include the exact filename.") {
		t.Fatalf("expected ambiguity message, got:\n%s", stdout.String())
	}
	for _, want := range []string{
		"- pear first.txt",
		"- pear second.txt",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("expected ambiguity message to list candidate %q, got:\n%s", want, stdout.String())
		}
	}
	if got := mustReadFile(t, first); got != "first\n" {
		t.Fatalf("expected first file to remain unchanged, got:\n%s", got)
	}
	if got := mustReadFile(t, second); got != "second\n" {
		t.Fatalf("expected second file to remain unchanged, got:\n%s", got)
	}
}

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

func TestStatusShowsPrivacyPreservingCLIHandoffReceipt(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")
	writeCurrentWorkRoute(t, packDir, map[string]any{
		"schema_version":       "0.1.0",
		"context_type":         "JiniWorkRoute",
		"tool_mode":            "claude-code",
		"tool_label":           "Claude Code CLI handoff",
		"route_policy":         "CLI handoff",
		"model_label":          "Claude Code CLI handoff",
		"provider_label":       "Claude Code CLI handoff",
		"chosen_automatically": false,
		"reason":               "Jini handed this request to the installed Claude Code CLI handoff subprocess.",
		"cli_handoff_receipt": map[string]any{
			"context_type":  "JiniCLIHandoffReceipt",
			"mode":          "claude-code",
			"label":         "Claude Code CLI handoff",
			"executable":    "/tmp/fake-claude",
			"args_template": []string{"--print", "{{prompt}}"},
			"cwd":           "/tmp/jini-work",
			"exit_status":   0,
			"duration_ms":   42,
			"prompt_chars":  79,
			"stdout_chars":  18,
			"stderr_chars":  0,
			"completed_at":  "2026-06-08T22:11:00Z",
		},
	})
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"status"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected status to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Last CLI handoff",
		"Claude Code CLI handoff via /tmp/fake-claude",
		"Args: --print {{prompt}}",
		"Exit 0 in 42ms; prompt 79 chars, stdout 18 chars, stderr 0 chars.",
		"Completed: 2026-06-08T22:11:00Z",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected status to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Summarize the private launch notes",
		"fake claude draft",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("status must not expose prompt or CLI output body %q, got:\n%s", unwanted, out)
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

func TestProviderDoctorJSONDetectsAzureOpenAIWithoutLeakingSecrets(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")
	t.Setenv("AZURE_OPENAI_API_VERSION", "2024-10-21")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"doctor", "--format", "json"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	var report struct {
		ResultType string `json:"result_type"`
		ProviderID string `json:"provider_id"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode doctor json: %v\n%s", err, stdout.String())
	}
	if report.ResultType != "JiniProviderDoctor" || report.ProviderID != "azure-openai" || report.Status != "ok" {
		t.Fatalf("unexpected provider doctor json: %#v", report)
	}
	rendered := stdout.String()
	if !strings.Contains(rendered, "AZURE_OPENAI_API_KEY") {
		t.Fatalf("expected secret presence marker in JSON, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "super-secret-key") {
		t.Fatalf("provider doctor leaked secret value:\n%s", rendered)
	}
}

func TestPublishReadinessJSONIsNativeGoAndReportsMigrationComplete(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"publish-readiness", "--format", "json"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	var report struct {
		ResultType string `json:"result_type"`
		Status     string `json:"status"`
		Runtime    struct {
			Language       string `json:"language"`
			LegacyFallback bool   `json:"legacy_fallback"`
		} `json:"runtime"`
		Sections []struct {
			ID     string `json:"id"`
			Checks []struct {
				Path   string `json:"path"`
				Exists bool   `json:"exists"`
				Status string `json:"status"`
			} `json:"checks"`
		} `json:"sections"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish readiness json: %v\n%s", err, stdout.String())
	}
	if report.ResultType != "JiniPublishReadiness" || report.Status != "ok" {
		t.Fatalf("unexpected publish readiness header: %#v", report)
	}
	if report.Runtime.Language != "go" || report.Runtime.LegacyFallback {
		t.Fatalf("expected native Go runtime with no legacy fallback, got: %#v", report.Runtime)
	}
	foundAppShippingGate := false
	for _, section := range report.Sections {
		if section.ID != "docs" {
			continue
		}
		for _, check := range section.Checks {
			if check.Path == "specs/app-platform-shipping-playbook.md" {
				foundAppShippingGate = true
				if !check.Exists || check.Status != "ok" {
					t.Fatalf("expected app platform shipping playbook gate to pass, got: %#v", check)
				}
			}
		}
	}
	if !foundAppShippingGate {
		t.Fatalf("publish readiness docs section did not include app platform shipping playbook gate: %#v", report.Sections)
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
		"Describe the task.",
		"Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.",
		"Add a line to the matching .txt file in this folder",
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

func TestTopLevelHelpFlagShowsPublicInventory(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"--help"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Usage:",
		"Essential commands:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestTopLevelShortHelpFlagShowsPublicInventory(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"-h"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Usage:",
		"Essential commands:",
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
		"Jini",
		"Usage:",
		"Examples:",
		"Essential commands:",
		"Setup:",
		"review this repo",
		"jini status",
		"jini continue",
		"jini open",
		"jini route",
		"jini doctor",
		"jini admin help",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"jini check", "jini provider doctor", "source runtime", "Python", "fallback"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("did not expect output to contain %q, got:\n%s", unwanted, out)
		}
	}
}

func TestPublicHelpReadsLikeShippedProduct(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"commands"}, {"--help"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range []string{
				"Jini",
				"Usage:",
				"Examples:",
				"Essential commands:",
				"Setup:",
				"jini route help",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected polished help to contain %q, got:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{
				"Public command inventory",
				"FREE-TIER VALUE",
				"START WITH JINI",
				"TRY FIRST IN A REPO",
				"native Go preview",
				"preview",
			} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected polished help not to contain %q, got:\n%s", unwanted, out)
				}
			}
		})
	}
}

func TestCommercialOnlyCommandFailsClosedInFreeCLI(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"skills"}, &stdout, &stdout)
	if exitCode == 0 {
		t.Fatalf("expected commercial-only command to fail closed, got output:\n%s", stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Not available in this CLI.",
		"Skills OS requires commercial entitlement",
		"jini route help",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected commercial-only command not to create current work, stat error: %v", err)
	}
}

func TestCommercialOnlyInteractiveInputDoesNotCreateWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("delegate\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected interactive commercial-only input to fail closed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Managed delegation requires commercial entitlement") {
		t.Fatalf("expected entitlement boundary output, got:\n%s", out)
	}
	for _, unwanted := range []string{"Task Snapshot", "Working Draft", "Saved artifact:"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected entitlement boundary to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected commercial-only input not to create current work, stat error: %v", err)
	}
}

func TestTopLevelHelpFlagsShowPublicCommandInventory(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range []string{
				"Jini",
				"Usage:",
				"Essential commands:",
				"Setup:",
				"jini admin help",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected top-level help flag output to contain %q, got:\n%s", want, out)
				}
			}
			if strings.Contains(out, "Current work") {
				t.Fatalf("top-level help flags should not show contextual current-work help, got:\n%s", out)
			}
		})
	}
}

func TestTopLevelHelpShowsPublicCommandInventory(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"help"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Jini",
		"Usage:",
		"Essential commands:",
		"jini admin help",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected top-level help output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Current work") {
		t.Fatalf("top-level help should not show contextual current-work help, got:\n%s", out)
	}
}

func TestHelpAllShowsPublicCommandInventory(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := app.Run([]string{"help", "--all"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Essential commands:") {
		t.Fatalf("expected public help, got:\n%s", stdout.String())
	}
}

func TestHelpCommandAliasesShowPublicCommandInventory(t *testing.T) {
	for _, args := range [][]string{{"help", "all"}, {"help", "commands"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			if !strings.Contains(stdout.String(), "Essential commands:") {
				t.Fatalf("expected public help, got:\n%s", stdout.String())
			}
		})
	}
}

func TestHelpAdminAliasShowsAdminInventory(t *testing.T) {
	for _, args := range [][]string{{"help", "admin"}, {"help", "--admin"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			if !strings.Contains(stdout.String(), "Admin and developer command inventory") {
				t.Fatalf("expected admin inventory, got:\n%s", stdout.String())
			}
		})
	}
}

func TestAdminHelpAliasShowsAdminInventory(t *testing.T) {
	for _, args := range [][]string{{"admin", "help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			for _, want := range []string{
				"Admin and developer command inventory",
				"jini provider doctor",
				"jini observe status",
				"jini check ship",
				"jini open <artifact>",
				"Admin commands stay intentionally narrow.",
				"jini publish-readiness",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, out)
				}
			}
			for _, unwanted := range []string{"source runtime", "Python", "fallback"} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("did not expect output to contain %q, got:\n%s", unwanted, out)
				}
			}
		})
	}
}

func TestAdminHelpFlagAliasesShowAdminInventory(t *testing.T) {
	for _, args := range [][]string{{"admin", "--help"}, {"admin", "-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			if !strings.Contains(stdout.String(), "Admin and developer command inventory") {
				t.Fatalf("expected admin inventory, got:\n%s", stdout.String())
			}
		})
	}
}

func TestProviderHelpAliasesShowAdminInventory(t *testing.T) {
	for _, args := range [][]string{{"provider", "help"}, {"provider", "--help"}, {"provider", "-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			exitCode := app.Run(args, &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}
			if !strings.Contains(stdout.String(), "Admin and developer command inventory") {
				t.Fatalf("expected admin inventory, got:\n%s", stdout.String())
			}
		})
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

func TestProviderDoctorJSONSubcommandMatchesTopLevelDoctor(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	var topLevel bytes.Buffer
	topExit := app.Run([]string{"doctor", "--format", "json"}, &topLevel, &topLevel)
	if topExit != 0 {
		t.Fatalf("expected top-level doctor json to succeed, got %d with output:\n%s", topExit, topLevel.String())
	}

	var subcommand bytes.Buffer
	subExit := app.Run([]string{"provider", "doctor", "--format", "json"}, &subcommand, &subcommand)
	if subExit != 0 {
		t.Fatalf("expected provider doctor json to succeed, got %d with output:\n%s", subExit, subcommand.String())
	}

	if topLevel.String() != subcommand.String() {
		t.Fatalf("expected provider doctor json to match top-level doctor json.\nTOP LEVEL:\n%s\nSUBCOMMAND:\n%s", topLevel.String(), subcommand.String())
	}
}

func TestProviderDoctorTextSubcommandMatchesTopLevelDoctor(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	var topLevel bytes.Buffer
	topExit := app.Run([]string{"doctor", "--format", "text"}, &topLevel, &topLevel)
	if topExit != 0 {
		t.Fatalf("expected top-level doctor text to succeed, got %d with output:\n%s", topExit, topLevel.String())
	}

	var subcommand bytes.Buffer
	subExit := app.Run([]string{"provider", "doctor", "--format", "text"}, &subcommand, &subcommand)
	if subExit != 0 {
		t.Fatalf("expected provider doctor text to succeed, got %d with output:\n%s", subExit, subcommand.String())
	}

	if topLevel.String() != subcommand.String() {
		t.Fatalf("expected provider doctor text to match top-level doctor text.\nTOP LEVEL:\n%s\nSUBCOMMAND:\n%s", topLevel.String(), subcommand.String())
	}
}

func TestProviderDoctorInlineFormatMatchesSpacedFormat(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "azure-openai")
	t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
	t.Setenv("AZURE_OPENAI_API_KEY", "super-secret-key")
	t.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4o-prod")

	for _, args := range []struct {
		name   string
		spaced []string
		inline []string
	}{
		{
			name:   "doctor json",
			spaced: []string{"doctor", "--format", "json"},
			inline: []string{"doctor", "--format=json"},
		},
		{
			name:   "doctor text",
			spaced: []string{"doctor", "--format", "text"},
			inline: []string{"doctor", "--format=text"},
		},
		{
			name:   "provider json",
			spaced: []string{"provider", "--format", "json"},
			inline: []string{"provider", "--format=json"},
		},
		{
			name:   "provider text",
			spaced: []string{"provider", "--format", "text"},
			inline: []string{"provider", "--format=text"},
		},
		{
			name:   "provider doctor json",
			spaced: []string{"provider", "doctor", "--format", "json"},
			inline: []string{"provider", "doctor", "--format=json"},
		},
		{
			name:   "provider doctor text",
			spaced: []string{"provider", "doctor", "--format", "text"},
			inline: []string{"provider", "doctor", "--format=text"},
		},
	} {
		t.Run(args.name, func(t *testing.T) {
			var spaced bytes.Buffer
			spacedExit := app.Run(args.spaced, &spaced, &spaced)
			if spacedExit != 0 {
				t.Fatalf("expected spaced format to succeed, got %d with output:\n%s", spacedExit, spaced.String())
			}

			var inline bytes.Buffer
			inlineExit := app.Run(args.inline, &inline, &inline)
			if inlineExit != 0 {
				t.Fatalf("expected inline format to succeed, got %d with output:\n%s", inlineExit, inline.String())
			}

			if spaced.String() != inline.String() {
				t.Fatalf("expected inline format to match spaced format.\nSPACED (%v):\n%s\nINLINE (%v):\n%s", args.spaced, spaced.String(), args.inline, inline.String())
			}
		})
	}
}

func TestProviderDoctorJSONDoesNotEscapeProviderArrows(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "auto")
	t.Setenv("JINI_MODEL", "sonnet-4.6")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "SECRETEXAMPLE")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"doctor", "--format", "json"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if strings.Contains(out, `\u003e`) {
		t.Fatalf("expected provider doctor json to preserve arrows, got:\n%s", out)
	}
	if !strings.Contains(out, `"presence": "auto -> Amazon Bedrock / Claude Sonnet 4.6"`) {
		t.Fatalf("expected readable auto-provider arrow in json, got:\n%s", out)
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
		"Setup saved. Working with Claude API route via Claude API / Claude Sonnet 4.",
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

func TestInteractiveSetupCanSaveClaudeAPIRouteInsideJini(t *testing.T) {
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
		"Setup saved. Working with Claude API route via Claude API / Claude Sonnet 4.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "claude-api"`) {
		t.Fatalf("expected saved Claude API route, got:\n%s", string(routerSaved))
	}
}

func TestRouteCommandListsAvailableRoutes(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	fakeBin := t.TempDir()
	fakeCodex := writeFakeExecutable(t, fakeBin, "codex", "printf 'fake codex\\n'\n")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "list"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route list to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Routes",
		"Current: auto",
		"Available",
		"auto",
		"- codex: Codex CLI handoff (cli, external, ok)",
		"- claude-code: Claude Code CLI handoff (cli, external, needs setup: missing executable)",
		"- gemini-cli: Gemini CLI handoff (cli, external, needs setup: missing executable)",
		"- aider: Aider CLI handoff (cli, external, needs setup: missing executable)",
		"- opencode: OpenCode CLI handoff (cli, external, needs setup: missing executable)",
		"claude-api",
		"- azure-code: Azure code route (remote, standard, needs setup)",
		"- local-preview: Local preview (local, free, ok)",
		"Use `jini route set codex` or `jini route set azure-code` to lock a configured route.",
		"CLI handoffs invoke installed CLIs or fail closed; they are not provider aliases.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route list to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, fakeCodex) {
		t.Fatalf("route list must not leak executable path %q, got:\n%s", fakeCodex, out)
	}
}

func TestRouteCommandShowsSetupHelp(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	for _, args := range [][]string{
		{"route", "help"},
		{"route help"},
	} {
		var stdout bytes.Buffer
		exitCode := app.Run(args, &stdout, &stdout)
		if exitCode != 0 {
			t.Fatalf("expected route help %v to succeed, got %d with output:\n%s", args, exitCode, stdout.String())
		}

		out := stdout.String()
		for _, want := range []string{
			"Route setup",
			"1. Run `jini route list`.",
			"2. Run `jini doctor`.",
			"3. Run `jini route set codex` or `jini route set claude-code`.",
			"4. Run `jini route dogfood` before release validation.",
			"Azure OpenAI API: set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_API_KEY`, and `AZURE_OPENAI_DEPLOYMENT`; then `jini route set azure-openai`.",
			"Bedrock API: set `AWS_REGION` plus `AWS_PROFILE` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`; then `jini route set bedrock-sonnet`.",
			"Local/offline model: run a local OpenAI-compatible server, then `jini route set local-slm` or `jini route auto`.",
			"Local preview: run `jini route set local-preview` when you want deterministic no-model fallback.",
			"Use env overrides only when auto-detect fails:",
			"- `AZURE_OPENAI_ENDPOINT=https://...openai.azure.com`",
			"- `AZURE_OPENAI_DEPLOYMENT=your-deployment`",
			"- `AWS_PROFILE=your-profile` or `AWS_ACCESS_KEY_ID=...`",
			"- Wave 1 CLI overrides: `JINI_CODEX_CLI`, `JINI_CLAUDE_CODE_CLI`, `JINI_GEMINI_CLI`, `JINI_AIDER_CLI`, `JINI_OPENCODE_CLI`.",
			"- `JINI_LOCAL_SLM_ENDPOINT=http://127.0.0.1:11434/v1`",
			"- `JINI_LOCAL_SLM_MODEL=qwen3:8b`",
			"Gemini API and Vertex AI provider routes stay planned until `gemini-cli` dogfood evidence is complete.",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("expected route help %v to contain %q, got:\n%s", args, want, out)
			}
		}
		for _, unwanted := range []string{
			"Available",
			"Common commands",
			"Provider and local routes",
			"Task Snapshot",
			"Working Draft",
			"API key:",
			"Secret:",
		} {
			if strings.Contains(out, unwanted) {
				t.Fatalf("expected route help %v to avoid %q, got:\n%s", args, unwanted, out)
			}
		}
		if got := nonEmptyLineCount(out); got > 18 {
			t.Fatalf("expected route help %v to stay compact, got %d non-empty lines:\n%s", args, got, out)
		}
	}
}

func TestRouteDogfoodShowsWave1ValidationGuide(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	fakeBin := t.TempDir()
	fakeCodex := writeFakeExecutable(t, fakeBin, "codex", "printf 'fake codex\\n'\n")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "dogfood"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route dogfood guide to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"CLI dogfood",
		"Evidence file:",
		"Required checks: auth, approvals, output shape, route receipt privacy",
		"- codex: ready; missing auth, approvals, output shape, route receipt privacy",
		"- claude-code: setup blocked (missing executable)",
		"- gemini-cli: setup blocked (missing executable)",
		"Setup fixes:",
		"- claude-code: install Claude Code CLI or set JINI_CLAUDE_CODE_CLI, then rerun `jini route dogfood`.",
		"- gemini-cli: install Gemini CLI or set JINI_GEMINI_CLI, then rerun `jini route dogfood`.",
		"Release claim policy:",
		"- Installed CLI routes and routes named in JINI_CLI_RELEASE_ROUTES must be trusted and dogfooded before release claims.",
		"- Missing optional CLI executables are setup backlog until the release claim names them.",
		"Validation steps:",
		"- For each ready route, select that route and run a harmless prompt through Jini using the real installed CLI.",
		"- Confirm downstream auth, approval behavior, output shape, and route receipt privacy before editing evidence.",
		"Evidence rules:",
		"- Do not use fake CLIs, provider API aliases, skipped trust checks, or stale evidence from an older CLI version.",
		"Template:",
		`"context_type": "JiniCLIHandoffDogfoodEvidence"`,
		`"codex"`,
		`"validated_at": "YYYY-MM-DDTHH:MM:SSZ"`,
		`"checks": ["auth", "approvals", "output shape", "route receipt privacy"]`,
		"Do not mark a route validated until you used the real installed CLI.",
		"Then run `jini check ship --format json`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route dogfood output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		fakeCodex,
		"Task Snapshot",
		"Working Draft",
		"Start/Keep",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected route dogfood output not to contain %q, got:\n%s", unwanted, out)
		}
	}

	stdout.Reset()
	exitCode = app.Run([]string{"route", "dogfood", "--format", "json"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route dogfood JSON guide to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected route dogfood JSON to decode: %v\n%s", err, stdout.String())
	}
	if got := payload["result_type"]; got != "JiniRouteDogfoodGuide" {
		t.Fatalf("expected route dogfood result type, got %#v in %s", got, stdout.String())
	}
	for _, key := range []string{"release_claim_policy", "validation_steps", "evidence_rules"} {
		values, ok := payload[key].([]any)
		if !ok || len(values) == 0 {
			t.Fatalf("expected route dogfood JSON %s to be a non-empty list, got %#v in %s", key, payload[key], stdout.String())
		}
	}
	if strings.Contains(stdout.String(), fakeCodex) || strings.Contains(stdout.String(), `"executable"`) || strings.Contains(stdout.String(), "args_template") {
		t.Fatalf("route dogfood JSON must not leak executable paths or command templates, got:\n%s", stdout.String())
	}
}

func TestRouteCommandShowsReadinessAndTokenPosture(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-preview")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route status to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Route and cost",
		"Current route: Local preview (chosen automatically). Readiness: ok.",
		"Token posture: compact context first",
		"Continuity: offline and online work stitch into the same session.",
		"Least-expense capable route is the default",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route output to contain %q, got:\n%s", want, out)
		}
	}
	if got := nonEmptyLineCount(out); got > 6 {
		t.Fatalf("expected route status to stay compact, got %d non-empty lines:\n%s", got, out)
	}
}

func TestRouteCommandShowsSelectedCLIWithoutExecutingIt(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "claude-code")
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	markerPath := filepath.Join(t.TempDir(), "executed.txt")
	t.Setenv("JINI_TEST_CLI_MARKER", markerPath)
	fakeCLI := writeFakeExecutable(t, binDir, "claude", "printf 'executed\\n' > \"$JINI_TEST_CLI_MARKER\"\n")
	t.Setenv("PATH", binDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route status to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Route and cost",
		"Current route: Claude Code CLI handoff. Readiness: ok.",
		"CLI handoff: " + fakeCLI,
		"Args: --print {{prompt}}",
		"Provider alias: disabled; Jini invokes the CLI subprocess only when running work.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route output to contain %q, got:\n%s", want, out)
		}
	}
	if _, err := os.Stat(markerPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("route status must not execute the configured CLI; marker stat err: %v", err)
	}
	if got := nonEmptyLineCount(out); got > 9 {
		t.Fatalf("expected CLI route status to stay compact, got %d non-empty lines:\n%s", got, out)
	}
}

func TestRouteCommandShowsMissingCLIAsHandoffSetupProblem(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "aider")
	t.Setenv("PATH", t.TempDir())

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route status to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Current route: Aider CLI handoff. Readiness: needs setup.",
		"CLI handoff: aider",
		"Args: --message {{prompt}}",
		"Setup: Aider CLI handoff requires an installed CLI executable: aider.",
		"Provider alias: disabled; Jini invokes the CLI subprocess only when running work.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "API route") {
		t.Fatalf("expected missing CLI handoff to avoid provider API framing, got:\n%s", out)
	}
}

func TestRouteCommandCanSetRouteWithoutSetupWizard(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "set", "azure-code"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route set to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Route saved.",
		"Current: Azure code route",
		"Mode: azure-code",
		"Use `jini route auto` to restore automatic routing.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route set output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "azure-code"`) {
		t.Fatalf("expected saved Azure code route, got:\n%s", string(routerSaved))
	}
}

func TestRouteCommandCanSetWaveOneCLIWhenExecutableIsPresent(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	binDir := t.TempDir()
	writeFakeExecutable(t, binDir, "codex", "printf 'fake codex\\n'\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "set", "codex"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected codex route set to succeed with fake CLI, got %d:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Route saved.",
		"Current: Codex CLI handoff",
		"Mode: codex",
		"CLI handoff: Codex CLI handoff",
		"Use `jini route auto` to restore automatic routing.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected codex route set output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "codex"`) {
		t.Fatalf("expected saved Codex route, got:\n%s", string(routerSaved))
	}
}

func TestRouteCommandCanSetLocalSLMAliasToEligibleProfile(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_DEVICE_CLASS_OVERRIDE", "mobile-small")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "gemma3n:e2b-it")

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "set", "local-slm"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected local-slm route set to succeed, got %d:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Route saved.",
		"Current: Local SLM fast",
		"Mode: local-slm",
		"Use `jini route auto` to restore automatic routing.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected local-slm route set output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "local-slm"`) {
		t.Fatalf("expected saved Local SLM alias route, got:\n%s", string(routerSaved))
	}
}

func TestRouteCommandFailsClosedForMissingWaveOneCLI(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", t.TempDir())

	for _, mode := range []string{"gemini-cli", "aider", "opencode"} {
		var stdout bytes.Buffer
		exitCode := app.Run([]string{"route", "set", mode}, &stdout, &stdout)
		if exitCode == 0 {
			t.Fatalf("expected %s route set to fail closed, got output:\n%s", mode, stdout.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"CLI handoff is not ready.",
			"requires an installed CLI executable",
			"will not use a provider API alias",
			"jini doctor",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("expected %s reserved route output to contain %q, got:\n%s", mode, want, out)
			}
		}
	}
}

func TestProviderDoctorFailsClosedForReservedCLIHandoffToolMode(t *testing.T) {
	t.Setenv("JINI_TOOL", "codex")
	t.Setenv("PATH", t.TempDir())

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"doctor"}, &stdout, &stdout)
	if exitCode == 0 {
		t.Fatalf("expected reserved CLI handoff doctor to fail closed, got output:\n%s", stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Codex CLI handoff",
		"ROUTE_POLICY: CLI handoff required",
		"CLI_HANDOFF_CONTRACT: cwd, prompt, stdout, stderr, exit status, route receipt",
		"AUTO_MODEL: Jini cannot claim this CLI route until it can hand off to the installed CLI.",
		"will not use a provider API alias",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected reserved CLI doctor output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "ROUTE_POLICY: Locked by you") {
		t.Fatalf("expected reserved CLI doctor to avoid generic locked-route policy, got:\n%s", out)
	}
}

func writeFakeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\nset -eu\n" + body
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake executable %s: %v", name, err)
	}
	return path
}

func TestRouteCommandCanRestoreAutoMode(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.Run([]string{"route", "set", "azure-code"}, io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected route set to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "auto"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route auto to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Auto route restored.",
		"Current: auto",
		"Jini will choose the least-expense capable route per request.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected route auto output to contain %q, got:\n%s", want, out)
		}
	}

	routerSaved, err := os.ReadFile(filepath.Join(stateDir, "router.json"))
	if err != nil {
		t.Fatalf("expected router settings file: %v", err)
	}
	if !strings.Contains(string(routerSaved), `"tool_mode": "auto"`) {
		t.Fatalf("expected saved auto route, got:\n%s", string(routerSaved))
	}
}

func TestRouteCommandRejectsUnknownRoute(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "set", "banana"}, &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected unknown route to fail, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Unknown route \"banana\".",
		"Run `jini route list` to see available routes.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected unknown route output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRouteCommandRejectsUnknownSubcommandWithRouteGuidance(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"route", "banana"}, &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected unknown route subcommand to fail, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Unknown route command \"banana\".",
		"Run `jini route list` to see available routes.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected unknown route command output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Run `jini commands`") {
		t.Fatalf("expected route parser to avoid generic command guidance, got:\n%s", out)
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
		"JINI_TOOL: auto -> Claude API route",
		"ROUTE_POLICY: Automatic",
		"JINI_MODEL_DECISION: Claude Sonnet 4",
		"AUTO_MODEL: Jini uses Claude Sonnet 4 by default on the Claude API route.",
		"AUTO_ROUTE: Auto mode chose Claude API route because this looks like general work, the request does not ask for deep review, so Jini favored the cheapest suitable route.",
		"JINI_EFFORT: auto -> dynamic per request",
		"AUTO_EFFORT: Jini judges effort separately for each request instead of pinning one level globally.",
		"AUTO_MODE: frameworks=auto; models=auto; speed=auto; approvals=approval-gated",
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

func TestProviderDoctorJSONUsesCachedDeviceProfileEvidence(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_PROVIDER", "local-slm")
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	t.Setenv("JINI_LOCAL_SLM_FAST_MODEL", "phi4-mini")
	t.Setenv("JINI_LOCAL_SLM_API_KEY", "top-secret-local-key")

	deviceProfilePayload := `{
  "schema_version": "0.2.0",
  "context_type": "JiniDeviceProfile",
  "captured_at": "2026-05-30T14:00:00Z",
  "os": "darwin",
  "os_version": "15.5",
  "arch": "arm64",
  "accelerator_class": "apple-gpu",
  "local_runtime_class": "local-ollama",
  "device_class": "laptop-strong",
  "local_profile_states": {
    "local-fast": "ready",
    "local-workhorse": "ready",
    "local-deep": "degraded",
    "local-multimodal": "ready"
  }
}`
	if err := os.WriteFile(filepath.Join(stateDir, "device-profile.json"), []byte(deviceProfilePayload), 0o644); err != nil {
		t.Fatalf("write device profile fixture: %v", err)
	}

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"doctor", "--format", "json"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	var report struct {
		ProviderID string `json:"provider_id"`
		Settings   []struct {
			Name     string `json:"name"`
			Presence string `json:"presence"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode doctor json: %v\n%s", err, stdout.String())
	}
	if report.ProviderID != "local-slm" {
		t.Fatalf("unexpected provider id: %#v", report)
	}
	settings := map[string]string{}
	for _, item := range report.Settings {
		settings[item.Name] = item.Presence
	}
	if settings["DEVICE_CLASS"] != "laptop-strong" {
		t.Fatalf("expected cached device class, got %#v", settings)
	}
	if settings["DEVICE_OS"] != "darwin 15.5" {
		t.Fatalf("expected cached device OS, got %#v", settings)
	}
	if settings["LOCAL_ACCELERATOR"] != "apple-gpu" {
		t.Fatalf("expected cached accelerator, got %#v", settings)
	}
	if settings["LOCAL_RUNTIME_CLASS"] != "local-ollama" {
		t.Fatalf("expected cached runtime class, got %#v", settings)
	}
	if settings["JINI_LOCAL_SLM_FAST_MODEL"] != "set (ready on this device)" {
		t.Fatalf("expected cached local-fast state, got %#v", settings)
	}
	if settings["JINI_LOCAL_SLM_WORKHORSE_MODEL"] != "missing (ready on this device)" {
		t.Fatalf("expected cached local-workhorse state, got %#v", settings)
	}
	if settings["JINI_LOCAL_SLM_DEEP_MODEL"] != "missing (degraded on this device)" {
		t.Fatalf("expected cached local-deep state, got %#v", settings)
	}
	if strings.Contains(stdout.String(), "top-secret-local-key") {
		t.Fatalf("provider doctor leaked local SLM API key:\n%s", stdout.String())
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
		"Describe the task, or run `jini commands`.",
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
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Current work",
		"Weekly Product Review",
		"Resume",
		"Other active work",
		"Switch",
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
		"Paste a new request to start a different task.",
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
		"tool_mode":            "azure-code",
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
		"tool_mode":            "chatgpt",
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
		"Describe the task in one sentence.",
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
		"Describe the task.",
		"Turn meeting notes into something I can send",
		"Check whether a plan is ready to hand off",
		"Plan a 7 day Paris trip for two adults in October",
		"Compare these vendors and recommend one",
		"Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.",
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

func TestShellOutputUsesPreciseProfessionalLanguage(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var helpOut bytes.Buffer
	if exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &helpOut, &helpOut); exitCode != 0 {
		t.Fatalf("expected help to succeed, got %d with output:\n%s", exitCode, helpOut.String())
	}
	for _, want := range []string{
		"Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.",
		"Jini asks before sending, booking, committing, or running destructive changes.",
	} {
		if !strings.Contains(helpOut.String(), want) {
			t.Fatalf("expected precise help output %q, got:\n%s", want, helpOut.String())
		}
	}

	var artifactOut bytes.Buffer
	if exitCode := app.RunInteractive(nil, strings.NewReader("Weekly product review for pricing launch. Need owners, due dates, and open questions.\n"), &artifactOut, &artifactOut); exitCode != 0 {
		t.Fatalf("expected artifact creation to succeed, got %d with output:\n%s", exitCode, artifactOut.String())
	}
	for _, want := range []string{
		"Artifact created.",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
	} {
		if !strings.Contains(artifactOut.String(), want) {
			t.Fatalf("expected precise artifact output %q, got:\n%s", want, artifactOut.String())
		}
	}

	freshStateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", freshStateDir)
	var bareEntityOut bytes.Buffer
	if exitCode := app.RunInteractive(nil, strings.NewReader("Paris\n"), &bareEntityOut, &bareEntityOut); exitCode != 0 {
		t.Fatalf("expected bare entity clarification to succeed, got %d with output:\n%s", exitCode, bareEntityOut.String())
	}
	if !strings.Contains(bareEntityOut.String(), "Specify what to do with Paris.") {
		t.Fatalf("expected precise bare-entity clarification, got:\n%s", bareEntityOut.String())
	}

	for _, output := range []string{helpOut.String(), artifactOut.String(), bareEntityOut.String()} {
		for _, forbidden := range []string{
			"Result ready.",
			"What would you like me to do",
			"Rough notes are fine",
			"act when safe",
		} {
			if strings.Contains(output, forbidden) {
				t.Fatalf("shell output must avoid imprecise phrase %q, got:\n%s", forbidden, output)
			}
		}
	}
}

func TestShellOutputRejectsStaleWorkflowVocabulary(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("3\nI have a messy note about renewal risks and next steps\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected generic fallback to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Artifact created.",
		"Request Brief",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected stale-free shell output to contain %q, got:\n%s", want, out)
		}
	}
	assertNoStaleShellVocabulary(t, out)
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
		"Artifact created.",
		"Sendable Follow-up",
		"## Send this",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
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
		"Artifact created.",
		"Build-Readiness Check",
		"## What looks ready now",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
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
		"Describe the task in one sentence.",
		"Jini can answer, edit files, route to a configured tool, or ask one clarification.",
		"Nothing will be sent, booked, committed, or changed without a visible step.",
		"Artifact created.",
		"Request Brief",
		"Request",
		"Current read",
		"Next options",
		"Safety",
		"Nothing has been sent",
		"Describe a new task when you want to move on.",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Type `Start`") {
		t.Fatalf("expected fallback result to avoid Start/Keep workflow language, got:\n%s", out)
	}
	if strings.Contains(out, "Short version or full version") {
		t.Fatalf("expected no first-run output-size prompt, got:\n%s", out)
	}
	if strings.Contains(out, "Goal") && strings.Index(out, "Request Brief") > strings.Index(out, "Goal") {
		t.Fatalf("expected first useful result before work summary, got:\n%s", out)
	}
	assertNoStaleShellVocabulary(t, out)
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
		"Describe the task in one sentence.",
		"Jini can answer, edit files, route to a configured tool, or ask one clarification.",
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
		"Jini",
		"Hi.",
		"Describe the task when you're ready.",
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

func TestInteractiveLauncherSupportsSlashCommandAliasesWithoutCreatingWork(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "help",
			line: "/help\n",
			want: []string{"Examples:", "Describe the task. Jini can answer"},
		},
		{
			name: "status",
			line: "/status\n",
			want: []string{"No current work yet."},
		},
		{
			name: "doctor",
			line: "/doctor\n",
			want: []string{"Provider", "Status"},
		},
		{
			name: "init",
			line: "/init\n",
			want: []string{"No init step is required before first value."},
		},
		{
			name: "memory",
			line: "/memory\n",
			want: []string{"Memory", "No current work is saved yet."},
		},
		{
			name: "permissions",
			line: "/permissions\n",
			want: []string{"Permissions", "Nothing has been sent, published, booked, or changed."},
		},
		{
			name: "clear",
			line: "/clear\n",
			want: []string{"Nothing to clear yet.", "Describe the task when you're ready."},
		},
		{
			name: "route",
			line: "/route\n",
			want: []string{"Route and cost", "Token posture: compact context first"},
		},
		{
			name: "route help",
			line: "/route help\n",
			want: []string{"Route setup", "jini route list", "Wave 1 CLI overrides", "JINI_GEMINI_CLI", "Gemini API and Vertex AI provider routes stay planned"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
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
			for _, unwanted := range []string{"Your first draft is ready.", "Working Draft"} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected slash command not to create work %q, got:\n%s", unwanted, out)
				}
			}
			if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected no current work file after slash command, got err=%v", err)
			}
		})
	}
}

func TestInteractiveLauncherRejectsUnknownSlashCommand(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("/banana\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Unknown command \"/banana\".") {
		t.Fatalf("expected unknown slash command rejection, got:\n%s", out)
	}
	if strings.Contains(out, "Working Draft") {
		t.Fatalf("expected unknown slash command not to create work, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no current work file after unknown slash command, got err=%v", err)
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
				"Type `help` for examples and commands.",
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
		"Artifact created.",
		"Sendable Follow-up",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
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

func TestNewCommandWithInputMatchesRunNew(t *testing.T) {
	input := "Turn meeting notes into something I can send.\n"

	runCase := func(args []string) (string, map[string]any) {
		t.Helper()
		stateDir := t.TempDir()
		t.Setenv("JINI_STATE_DIR", stateDir)

		var stdout bytes.Buffer
		exitCode := app.RunInteractive(args, strings.NewReader(input), &stdout, &stdout)
		if exitCode != 0 {
			t.Fatalf("expected exit code 0 for %v, got %d with output:\n%s", args, exitCode, stdout.String())
		}
		return stdout.String(), readCurrentWork(t, stateDir)
	}

	newOut, newCurrent := runCase([]string{"new"})
	runNewOut, runNewCurrent := runCase([]string{"run", "new"})

	for _, out := range []string{newOut, runNewOut} {
		for _, want := range []string{
			"Artifact created.",
			"Sendable Follow-up",
			"Saved artifact:",
			"Next commands: `jini continue`, `jini open`, or `jini status`.",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("expected output to contain %q, got:\n%s", want, out)
			}
		}
	}

	if normalizeSavedPathsForTest(newOut) != normalizeSavedPathsForTest(runNewOut) {
		t.Fatalf("expected `new` with stdin to match `run new`.\nNEW:\n%s\nRUN NEW:\n%s", newOut, runNewOut)
	}
	if newCurrent["pack_id"] != "meeting-followup" || runNewCurrent["pack_id"] != "meeting-followup" {
		t.Fatalf("expected both commands to create meeting-followup work, got new=%#v run_new=%#v", newCurrent, runNewCurrent)
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
				"Next commands: `jini continue`, `jini open`, or `jini status`.",
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
				"Next commands: `jini continue`, `jini open`, or `jini status`.",
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
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Current work",
		"Weekly Product Review",
		"Resume",
		"Other active work",
		"Switch",
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

func TestCurrentWorkPromptHidesResumeHintAfterFocusChanges(t *testing.T) {
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
	for _, unwanted := range []string{
		"Resume",
		"Owners and Due Points",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected prompt to hide %q after focus change, got:\n%s", unwanted, out)
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
			name: "slash status",
			line: "/status\n",
			want: []string{"Goal", "Weekly Product Review", "Ready now", "Sendable Follow-up"},
		},
		{
			name: "memory",
			line: "memory\n",
			want: []string{"Memory", "Current work is saved: Weekly Product Review."},
		},
		{
			name: "slash memory",
			line: "/memory\n",
			want: []string{"Memory", "Current work is saved: Weekly Product Review."},
		},
		{
			name: "route",
			line: "route\n",
			want: []string{
				"Route and cost",
				"Token posture: compact context first",
				"Continuity: offline and online work stitch into the same session.",
				"Route inputs: device",
				"CLI throttle levels affect switching",
				"Least-expense capable route",
			},
		},
		{
			name: "slash route",
			line: "/route\n",
			want: []string{
				"Route and cost",
				"Token posture: compact context first",
				"Continuity: offline and online work stitch into the same session.",
				"Route inputs: device",
				"CLI throttle levels affect switching",
				"Least-expense capable route",
			},
		},
		{
			name: "route help",
			line: "route help\n",
			want: []string{"Route setup", "jini route list", "Wave 1 CLI overrides", "JINI_GEMINI_CLI", "Gemini API and Vertex AI provider routes stay planned"},
		},
		{
			name: "slash route help",
			line: "/route help\n",
			want: []string{"Route setup", "jini route list", "Wave 1 CLI overrides", "JINI_GEMINI_CLI", "Gemini API and Vertex AI provider routes stay planned"},
		},
		{
			name: "permissions",
			line: "permissions\n",
			want: []string{"Permissions", "Nothing has been sent, published, booked, or changed."},
		},
		{
			name: "slash permissions",
			line: "/permissions\n",
			want: []string{"Permissions", "Nothing has been sent, published, booked, or changed."},
		},
		{
			name: "clear",
			line: "clear\n",
			want: []string{"Nothing was deleted.", "Paste a new request when ready."},
		},
		{
			name: "slash clear",
			line: "/clear\n",
			want: []string{"Nothing was deleted.", "Paste a new request when ready."},
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

func TestCurrentWorkInteractiveRejectsUnknownSlashCommand(t *testing.T) {
	stateDir := t.TempDir()
	packDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, packDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("/banana\n"), &stdout, &stdout)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Unknown command \"/banana\".") {
		t.Fatalf("expected unknown slash command rejection, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "Working Draft") {
		t.Fatalf("expected unknown slash command not to start literal work, got:\n%s", stdout.String())
	}
	current := readCurrentWork(t, stateDir)
	if current["title"] != "Weekly Product Review" {
		t.Fatalf("expected current work to remain unchanged, got %#v", current)
	}
}

func TestPostResultStatusCommandShowsFullState(t *testing.T) {
	for _, command := range []string{"status", "/status"} {
		t.Run(command, func(t *testing.T) {
			source := "Weekly product review. Need owners, due dates, and open questions."
			out := runInteractiveForTest(t, t.TempDir(), source+"\n"+command+"\n")

			for _, want := range []string{
				"Artifact created.",
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
		})
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
		{name: "route", arg: "route", maxNonEmpty: 6},
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
		{name: "direct travel draft", input: "plan me a 7 day paris trip\n", maxLines: 55},
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

func TestInteractiveTravelPromptDraftsFirstForClearDestinationAndDuration(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("plan a 3 day trip to Paris for a first time visitor\n")
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, stdin, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"3 day Paris itinerary",
		"Day 1: arrive, settle in, and take an easy first walk in Paris.",
		"Day 2: pick one neighborhood or museum anchor, then leave room for food and wandering.",
		"Day 3: keep this as a flexible final day with one favorite stop and departure buffer.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected travel prompt to draft with %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Before I draft it",
		"Type `skip`",
		"Task Snapshot",
		"Artifact created.",
		"Saved artifact:",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected clear travel prompt to draft first without %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected direct travel prompt not to create current work, stat error: %v", err)
	}
}

func TestInteractiveTravelPromptDraftsForGenericDestination(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	out := runInteractiveForTest(t, stateDir, "plan a 2 day trip to Kyoto for a first time visitor\n")
	for _, want := range []string{
		"2 day Kyoto itinerary",
		"Day 1: arrive, settle in, and take an easy first walk in Kyoto.",
		"Day 2: keep this as a flexible final day with one favorite stop and departure buffer.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected generic destination travel draft to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Before I draft it", "Task Snapshot", "Saved artifact:"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected generic destination travel draft to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveTravelPromptAsksWhenDestinationMissing(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	out := runInteractiveForTest(t, stateDir, "plan a 3 day trip for a first time visitor\n")
	for _, want := range []string{
		"Before I draft it",
		"- base area, or whether you want help choosing one",
		"Type `skip` if you want a generic first draft.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected destination-missing travel prompt to ask with %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"3 day the destination itinerary", "Artifact created.", "Saved artifact:"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected destination-missing travel prompt to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveLauncherAsksForTravelClarificationWhenDestinationOnly(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	stdin := strings.NewReader("Paris trip\ncouple, around $2500, early October, mixed pace, central hotel area, Versailles optional\n")
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
		"Paris Trip",
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

	stdin := strings.NewReader("Paris trip for a couple with a $2500 budget\nearly October, mixed pace, central hotel area, Louvre and Versailles are must-dos\n")
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

	stdin := strings.NewReader("5 day Rome trip for a couple, around $2500, early October, mixed pace, central stay, Colosseum is a must-do\n")
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
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Current work",
		"Weekly Product Review",
		"Other active work",
		"7-Day Paris Trip",
		"Switch",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected default launcher to hide saved work %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkHelpGroupsDuplicateOtherActiveWorkTitles(t *testing.T) {
	stateDir := t.TempDir()
	travelDir := copyWorkDir(t, filepath.Join(stateDir, "work", "travel-plan-paris"), seedTravelWork(t))
	copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	duplicateDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review-2"), seedMeetingWork(t))
	writeFile(t, filepath.Join(duplicateDir, "work-unit.yaml"), strings.TrimSpace(`
work_unit_id: example-meeting-followup-2
title: weekly product review
current_state: decided
`)+"\n")
	writeCurrentWork(t, stateDir, travelDir, "travel-plan", "paris-7d", "7-Day Paris Trip", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "- Weekly Product Review (2 saved)") {
		t.Fatalf("expected duplicate active work to be grouped, got:\n%s", out)
	}
	if count := strings.Count(out, "- Weekly Product Review"); count != 1 {
		t.Fatalf("expected one visible Weekly Product Review row, got %d:\n%s", count, out)
	}
}

func TestCurrentWorkHelpDisambiguatesOtherActiveWorkMatchingCurrentTitle(t *testing.T) {
	stateDir := t.TempDir()
	currentDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review-2"), seedMeetingWork(t))
	writeCurrentWork(t, stateDir, currentDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("help\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Goal\nWeekly Product Review",
		"Other active work\n- Weekly Product Review (another saved)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Switch") {
		t.Fatalf("expected help to avoid visible Switch vocabulary, got:\n%s", out)
	}
}

func TestInteractiveLauncherCanResumeNamedActiveProject(t *testing.T) {
	stateDir := t.TempDir()
	travelDir := copyWorkDir(t, filepath.Join(stateDir, "work", "travel-plan-paris"), seedTravelWork(t))
	meetingDir := copyWorkDir(t, filepath.Join(stateDir, "work", "meeting-followup-weekly-review"), seedMeetingWork(t))
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("7-Day Paris Trip\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Resumed",
		"7-Day Paris Trip",
		"Ready now",
		"Itinerary",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"Switch",
		"Type a number",
		"Other active work",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct saved-work resume to avoid %q, got:\n%s", unwanted, out)
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

func TestCurrentWorkFreeformInputStartsNewWorkDirectly(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Meeting Notes With Owners And Due Dates For Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("plan me a 7 day paris trip for a couple, around $2500, early October, mixed pace, central hotel area\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"Artifact created.",
		"7 Day Paris Trip",
		"Itinerary",
		"Budget Sketch",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{
		"New work",
		"Current:",
		"\n- Start\n",
		"\n- Keep\n",
		"Choose `Start`",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected direct freeform flow to avoid %q, got:\n%s", unwanted, out)
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

func TestCurrentWorkFreeformInputDoesNotTreatKeepAsModalChoice(t *testing.T) {
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
	if !strings.Contains(out, "Artifact created.") || !strings.Contains(out, "7 Day Paris Trip") {
		t.Fatalf("expected freeform request to start new work directly, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"New work",
		"Keeping current work.",
		"\n- Start\n",
		"\n- Keep\n",
		"Choose `Start`",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected no modal keep path %q, got:\n%s", unwanted, out)
		}
	}
	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected current work to become travel-plan, got %#v", current)
	}
}

func TestCurrentWorkSimpleFactualQuestionAnswersDirectly(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Add A Line Saying Jini Was Here In The Pear Vc Script Txt File In This Folder", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what is the capital of france\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Paris.") {
		t.Fatalf("expected direct answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nJust finished\n",
		"\nDoing now\n",
		"\nReady now\n",
		"New work",
		"Working Draft",
		"Your first draft is ready.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected simple question to avoid status/draft flow %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkCapitalQuestionAnswersFromSmallLookup(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what is the capital of japan\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Tokyo.") {
		t.Fatalf("expected direct capital answer, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "New work") || strings.Contains(stdout.String(), "Working Draft") {
		t.Fatalf("expected direct answer to avoid work flow, got:\n%s", stdout.String())
	}
}

func TestCurrentWorkCapitalQuestionAcceptsNaturalPhrasing(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what's the capital city of Germany?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Berlin.") {
		t.Fatalf("expected direct capital answer, got:\n%s", stdout.String())
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"New work",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(stdout.String(), unwanted) {
			t.Fatalf("expected natural capital question to avoid %q, got:\n%s", unwanted, stdout.String())
		}
	}
}

func TestDirectArgsSimpleFactualQuestionAnswersDirectly(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.Run([]string{"what's", "the", "capital", "of", "the", "United", "States?"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Washington, DC.") {
		t.Fatalf("expected direct capital answer, got:\n%s", stdout.String())
	}
	for _, unwanted := range []string{
		"Working on:",
		"Saved.",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(stdout.String(), unwanted) {
			t.Fatalf("expected direct question args to avoid %q, got:\n%s", unwanted, stdout.String())
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected direct question not to create current work, stat error: %v", err)
	}
}

func TestInteractiveSimpleFactualQuestionAnswersDirectlyWithoutCurrentWork(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what's the capital city of Germany?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Berlin.") {
		t.Fatalf("expected direct capital answer, got:\n%s", stdout.String())
	}
	for _, unwanted := range []string{
		"Artifact created.",
		"Task Snapshot",
		"Saved artifact:",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(stdout.String(), unwanted) {
			t.Fatalf("expected standalone question to avoid %q, got:\n%s", unwanted, stdout.String())
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected standalone question not to create current work, stat error: %v", err)
	}
}

func TestInteractiveSimpleArithmeticQuestionAnswersDirectlyWithoutArtifactShell(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("whats 3+7\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "10.") {
		t.Fatalf("expected direct arithmetic answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Artifact created.",
		"Task Snapshot",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected arithmetic question to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected arithmetic question not to create current work, stat error: %v", err)
	}
}

func TestCurrentWorkSimpleArithmeticQuestionAnswersDirectly(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what is 12 / 3?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "4.") {
		t.Fatalf("expected direct arithmetic answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nJust finished\n",
		"New work",
		"Task Snapshot",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected arithmetic question to avoid current-work flow %q, got:\n%s", unwanted, out)
		}
	}
}

func TestInteractiveTypoCapitalQuestionAnswersDirectlyWithoutArtifactShell(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("whats teh capital of france\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Paris.") {
		t.Fatalf("expected direct capital answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Artifact created.",
		"Task Snapshot",
		"Saved artifact:",
		"Next commands: `jini continue`, `jini open`, or `jini status`.",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
		"Itinerary",
		"Before I draft it",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected typo capital question to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected typo capital question not to create current work, stat error: %v", err)
	}
}

func TestInteractiveMalformedCapitalQuestionCorrectsWithoutTravelFlow(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("whats the capital of paris\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Paris is a city, not a country. Paris is the capital of France.") {
		t.Fatalf("expected compact correction, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Before I draft it",
		"travelers",
		"Artifact created.",
		"Itinerary",
		"Trip at a glance",
		"Saved artifact:",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected malformed capital question to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected malformed capital question not to create current work, stat error: %v", err)
	}
}

func TestInteractiveBareEntityAsksForIntentWithoutCreatingWork(t *testing.T) {
	for _, input := range []string{"Paris\n", "France\n", "Acme Corp\n", "OpenAI\n"} {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			stateDir := t.TempDir()
			t.Setenv("JINI_STATE_DIR", stateDir)

			var stdout bytes.Buffer
			exitCode := app.RunInteractive(nil, strings.NewReader(input), &stdout, &stdout)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
			}

			out := stdout.String()
			if !strings.Contains(out, "Specify what to do with "+strings.TrimSpace(input)+".") {
				t.Fatalf("expected intent clarification, got:\n%s", out)
			}
			for _, unwanted := range []string{
				"Before I draft it",
				"Artifact created.",
				"Itinerary",
				"Task Snapshot",
				"Saved artifact:",
				"Working Draft",
				"Your first draft is ready.",
			} {
				if strings.Contains(out, unwanted) {
					t.Fatalf("expected bare entity to avoid %q, got:\n%s", unwanted, out)
				}
			}
			if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected bare entity not to create current work, stat error: %v", err)
			}
		})
	}
}

func TestInteractiveExplicitTripChoiceCanUseBareDestination(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("trip\nParis\nskip\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "# Itinerary: Paris") || !strings.Contains(out, "Itinerary") {
		t.Fatalf("expected explicit trip choice to create travel work, got:\n%s", out)
	}
	if strings.Contains(out, "Specify what to do with Paris.") {
		t.Fatalf("expected explicit trip choice not to trigger bare-entity clarification, got:\n%s", out)
	}
	current := readCurrentWork(t, stateDir)
	if current["pack_id"] != "travel-plan" {
		t.Fatalf("expected explicit trip choice to create travel work, got %#v", current)
	}
}

func TestCurrentWorkTypoCapitalQuestionAnswersDirectly(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("whats teh capital of france\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Paris.") {
		t.Fatalf("expected direct capital answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nJust finished\n",
		"Artifact created.",
		"Task Snapshot",
		"Saved artifact:",
		"New work",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected current-work typo capital question to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkMalformedCapitalQuestionCorrectsDirectly(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("whats the capital of paris\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Paris is a city, not a country. Paris is the capital of France.") {
		t.Fatalf("expected compact correction, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nJust finished\n",
		"Before I draft it",
		"Artifact created.",
		"Itinerary",
		"New work",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected current-work malformed capital question to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestCurrentWorkUnknownStandaloneQuestionStaysCompact(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("who is the president of atlantis\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "No configured route can answer this locally.") {
		t.Fatalf("expected compact setup-needed fallback, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nGoal\n",
		"\nAI route\n",
		"\nJust finished\n",
		"New work",
		"Task Snapshot",
		"Working Draft",
		"Your first draft is ready.",
		"- Start",
		"- Keep",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected unknown question to avoid work flow %q, got:\n%s", unwanted, out)
		}
	}
}

func TestStandaloneQuestionUsesConfiguredCLIRouteWithoutCreatingWork(t *testing.T) {
	stateDir := t.TempDir()
	fakeBin := t.TempDir()
	fakeCodex := writeFakeExecutable(t, fakeBin, "codex", "printf 'Tim Cook.\\n'\n")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "codex")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("who is the CEO of Apple?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Tim Cook.") {
		t.Fatalf("expected compact routed answer, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Artifact created.",
		"Task Snapshot",
		"Saved artifact:",
		"Working Draft",
		"Your first draft is ready.",
		"I don't know locally.",
		"No configured route",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected routed standalone question to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected routed standalone question not to create current work, stat error: %v", err)
	}
}

func TestStandaloneQuestionTimeoutStaysCompact(t *testing.T) {
	stateDir := t.TempDir()
	fakeBin := t.TempDir()
	fakeCodex := writeFakeExecutable(t, fakeBin, "codex", "sleep 1\n")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "codex")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_STANDALONE_QUESTION_TIMEOUT", "20ms")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("who is the CEO of Apple?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Route timed out.") {
		t.Fatalf("expected compact timeout fallback, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"Task Snapshot",
		"Saved artifact:",
		"Working Draft",
		"stdout",
		"stderr",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected timeout fallback to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestStandaloneQuestionFailedCLIRouteStaysCompact(t *testing.T) {
	stateDir := t.TempDir()
	fakeBin := t.TempDir()
	fakeCodex := writeFakeExecutable(t, fakeBin, "codex", "printf 'long internal stdout that should not appear\\n'; printf 'secret stderr detail\\n' >&2; exit 42\n")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_TOOL", "codex")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("who is the CEO of Apple?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Route unavailable.") {
		t.Fatalf("expected compact route unavailable fallback, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"long internal stdout",
		"secret stderr detail",
		"stdout",
		"stderr",
		"Task Snapshot",
		"Saved artifact:",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected compact failed CLI fallback to avoid %q, got:\n%s", unwanted, out)
		}
	}
}

func TestTaskShapedQuestionDoesNotUseStandaloneQuestionFallback(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("can you fix the failing tests?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected task-shaped question to start normal work, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if strings.Contains(out, "No configured route can answer this locally.") {
		t.Fatalf("expected task-shaped question not to use standalone fallback, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "current-work.json")); err != nil {
		t.Fatalf("expected task-shaped question to create normal work, stat error: %v", err)
	}
}

func TestCurrentWorkQuestionStillCanShowWorkState(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what is blocked?\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Blocked") || !strings.Contains(stdout.String(), "- Metric and legal-review decision") {
		t.Fatalf("expected compact current-work blocked answer, got:\n%s", stdout.String())
	}
	for _, unwanted := range []string{
		"New work",
		"\n- Start\n",
		"\n- Keep\n",
		"I don't know locally.",
	} {
		if strings.Contains(stdout.String(), unwanted) {
			t.Fatalf("expected current-work question not to use %q, got:\n%s", unwanted, stdout.String())
		}
	}
}

func TestCurrentWorkQuestionClassifierDoesNotHijackStandaloneNextQuestion(t *testing.T) {
	stateDir := t.TempDir()
	meetingDir := seedMeetingWork(t)
	writeCurrentWork(t, stateDir, meetingDir, "meeting-followup", "example-meeting-followup", "Weekly Product Review", "decided", "ready-to-make")

	t.Setenv("JINI_STATE_DIR", stateDir)
	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader("what is the next version of gemma\n"), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "No configured route can answer this locally.") {
		t.Fatalf("expected standalone fallback, got:\n%s", out)
	}
	for _, unwanted := range []string{
		"\nNext\n",
		"ready-to-make",
		"New work",
		"Task Snapshot",
		"\n- Start\n",
		"\n- Keep\n",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected standalone next question not to use current-work state %q, got:\n%s", unwanted, out)
		}
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
	out := runInteractiveForTest(t, stateDir, "7 day paris trip for a couple, around $2500, early October, mixed pace, central hotel area, Versailles optional\ncontext\n")

	for _, want := range []string{
		"Context",
		"From you",
		"Your request: 7 day paris trip for a couple, around $2500, early October, mixed pace, central hotel area, Versailles optional",
		"Kept visible",
		"Route and continuity",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPromptHidesContextResumeHintAfterOpeningContextSurface(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)

	if exitCode := app.RunInteractive(nil, strings.NewReader("7 day paris trip for a couple, around $2500, early October, mixed pace, central hotel area, Versailles optional\ncontext\n"), io.Discard, io.Discard); exitCode != 0 {
		t.Fatalf("expected context action to succeed, got %d", exitCode)
	}

	var stdout bytes.Buffer
	exitCode := app.RunInteractive(nil, strings.NewReader(""), &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected prompt after context focus to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}
	out := stdout.String()
	for _, unwanted := range []string{
		"Context",
		"What Jini used",
		"Resume",
		"Current work",
	} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected task-first prompt not to show resume hint %q, got:\n%s", unwanted, out)
		}
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

func normalizeSavedPathsForTest(out string) string {
	lines := strings.Split(out, "\n")
	for index, line := range lines {
		if strings.HasPrefix(line, "Saved artifact: ") {
			lines[index] = "Saved artifact: <path>"
		}
	}
	return strings.Join(lines, "\n")
}

func assertNoStaleShellVocabulary(t *testing.T, out string) {
	t.Helper()
	for _, forbidden := range []string{
		"Task Snapshot",
		"Working Draft",
		"Your first draft is ready",
		"Result ready",
		"Start/Keep",
		"Switch to change focus",
		"Also ready:",
		"Next Actions",
		"Safe right now",
		"Useful starting point",
		"Best next inputs",
		"What this looks like",
		"Paste what you want finished",
		"Rough notes are fine",
		"act when safe",
	} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("shell output must not include stale vocabulary %q, got:\n%s", forbidden, out)
		}
	}
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

func BenchmarkInteractiveCurrentWorkDirectFreeform(b *testing.B) {
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

func runGitForTest(t testing.TB, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}
