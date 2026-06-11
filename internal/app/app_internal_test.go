package app

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafelyRunInteractiveRecoversPanics(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := safelyRunInteractive(&stderr, func() int {
		panic("boom")
	})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Jini hit an unexpected internal error and stopped safely.") {
		t.Fatalf("expected safe panic message, got:\n%s", stderr.String())
	}
}

func TestOfflineRegressionGuardrailsFailWhenRequiredSpecContentIsMissing(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create specs dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "local-model-support-matrix.md"), strings.Join([]string{
		"## Registry Contract",
		"`profile_role`",
		"`status`",
	}, "\n"))
	writeTestFile(t, filepath.Join(specDir, "platform-offline-strategy.md"), strings.Join([]string{
		"## Future Update Policy",
		"Future model updates should:",
		"preserve route evidence shape",
		"preserve session and artifact identity",
	}, "\n"))
	writeTestFile(t, filepath.Join(specDir, "adapter-benchmark-gate.md"), strings.Join([]string{
		"### 4. Routing Use",
		"repeated regression across recent samples",
		"strong recovery after degradation",
	}, "\n"))

	section := buildPublishOfflineRegressionSection(root)
	if section.Status != "needs-attention" {
		t.Fatalf("expected missing promotion loop to require attention, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "specs/local-model-support-matrix.md#promotion-loop" && check.Status == "incomplete" {
			for _, want := range []string{"## Promotion Loop", "### 3. Canary", "### 5. Promote", "successor versions"} {
				if !stringSliceContains(check.MissingFragments, want) {
					t.Fatalf("expected missing fragments to include %q, got %#v", want, check.MissingFragments)
				}
			}
			return
		}
	}
	t.Fatalf("expected promotion loop check to be incomplete, got %#v", section.Checks)
}

func TestPublishReadinessTextRendersMissingFragments(t *testing.T) {
	var stdout bytes.Buffer
	renderPublishReadinessText(&stdout, publishReadinessReport{
		Status: "needs-attention",
		Runtime: publishReadinessRuntime{
			Language:       "go",
			LegacyFallback: false,
		},
		Sections: []publishReadinessSection{{
			ID:     "app-platform",
			Status: "needs-attention",
			Checks: []publishReadinessCheck{{
				Path:             "specs/app-platform-shipping-playbook.md#source-backed-inputs",
				Exists:           true,
				Status:           "incomplete",
				MissingFragments: []string{"OpenTelemetry"},
			}},
		}},
	})

	out := stdout.String()
	for _, want := range []string{
		"    INCOMPLETE specs/app-platform-shipping-playbook.md#source-backed-inputs",
		"      MISSING OpenTelemetry",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected text output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPublishReadinessJSONRendersMissingFragments(t *testing.T) {
	encoded, err := json.Marshal(publishReadinessSection{
		ID:     "app-platform",
		Status: "needs-attention",
		Checks: []publishReadinessCheck{{
			Path:             "specs/app-platform-shipping-playbook.md#source-backed-inputs",
			Exists:           true,
			Status:           "incomplete",
			MissingFragments: []string{"OpenTelemetry"},
		}},
	})
	if err != nil {
		t.Fatalf("marshal readiness section: %v", err)
	}

	out := string(encoded)
	for _, want := range []string{
		`"missing_fragments":["OpenTelemetry"]`,
		`"status":"incomplete"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected JSON to contain %q, got:\n%s", want, out)
		}
	}
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestUnknownCommandsDoNotUseFallback(t *testing.T) {
	for _, args := range [][]string{
		{"compile-pack"},
		{"harnesses"},
	} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := Run(args, &stdout, &stderr)
		if exitCode != 1 {
			t.Fatalf("expected exit code 1 for %v, got %d\nstdout:\n%s\nstderr:\n%s", args, exitCode, stdout.String(), stderr.String())
		}
		if strings.Contains(strings.ToLower(stderr.String()), "fallback") {
			t.Fatalf("unexpected fallback reference in stderr for %v:\n%s", args, stderr.String())
		}
		if !strings.Contains(stderr.String(), "Unknown command") {
			t.Fatalf("expected unknown command message for %v, got:\n%s", args, stderr.String())
		}
	}
}

func TestUnsupportedNativeArgumentsFailFast(t *testing.T) {
	cases := [][]string{
		{"help", "admin", "extra"},
		{"status", "extra"},
		{"observe", "status", "extra"},
		{"open", "--print-path"},
		{"publish-readiness", "extra"},
		{"scorecard-gate", "extra"},
	}
	for _, args := range cases {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := Run(args, &stdout, &stderr)
		if exitCode != 1 {
			t.Fatalf("expected %v to fail, got %d\nstdout:\n%s\nstderr:\n%s", args, exitCode, stdout.String(), stderr.String())
		}
		if !strings.Contains(stderr.String(), "Unsupported arguments") {
			t.Fatalf("expected unsupported argument message for %v, got:\n%s", args, stderr.String())
		}
	}
}

func TestCatalogCommandsRejectRequestTailsWithRecoveryPath(t *testing.T) {
	cases := []struct {
		args      []string
		firstLine string
	}{
		{
			args:      []string{"help", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini help` shows the CLI overview; it does not take a request like \"me edit pear fellow script.txt\".",
		},
		{
			args:      []string{"commands", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini commands` shows the public command inventory; it does not take a request like \"me edit pear fellow script.txt\".",
		},
		{
			args:      []string{"--help", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini --help` shows the CLI overview; it does not take a request like \"me edit pear fellow script.txt\".",
		},
		{
			args:      []string{"admin", "help", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini admin help` shows the admin command inventory; it does not take a request like \"me edit pear fellow script.txt\".",
		},
		{
			args:      []string{"provider", "help", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini provider help` shows the admin command inventory; it does not take a request like \"me edit pear fellow script.txt\".",
		},
		{
			args:      []string{"provider", "--help", "me", "edit", "pear", "fellow", "script.txt"},
			firstLine: "ERROR `jini provider --help` shows the admin command inventory; it does not take a request like \"me edit pear fellow script.txt\".",
		},
	}

	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(tc.args, &stdout, &stderr)
			if exitCode != 1 {
				t.Fatalf("expected %v to fail, got %d\nstdout:\n%s\nstderr:\n%s", tc.args, exitCode, stdout.String(), stderr.String())
			}

			out := stderr.String()
			if !strings.HasPrefix(out, tc.firstLine+"\n") {
				t.Fatalf("expected first line %q, got:\n%s", tc.firstLine, out)
			}
			for _, want := range []string{
				"Start with `jini` to resume active work or see the start options.",
				"If you already have current work, use `jini status` once.",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("expected recovery line %q for %v, got:\n%s", want, tc.args, out)
				}
			}
			if strings.Contains(out, "Unsupported arguments") || strings.Contains(out, "Run `jini commands`") {
				t.Fatalf("catalog request tails should avoid generic parser recovery for %v, got:\n%s", tc.args, out)
			}
		})
	}
}

func TestPublishReadinessSupportsTextAndInlineJSONFormat(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"publish-readiness"}, want: "RUNTIME"},
		{args: []string{"publish-readiness", "--format=text"}, want: "LEGACY_FALLBACK false"},
		{args: []string{"publish-readiness", "--format=json"}, want: `"legacy_fallback": false`},
	}
	for _, tc := range cases {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := Run(tc.args, &stdout, &stderr)
		if exitCode != 0 {
			t.Fatalf("expected %v to pass, got %d\nstdout:\n%s\nstderr:\n%s", tc.args, exitCode, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), tc.want) {
			t.Fatalf("expected %v output to contain %q, got:\n%s", tc.args, tc.want, stdout.String())
		}
	}
}

func TestShipCheckReportsRepoValidationEvidenceAsJSON(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommandForInternalTest(t, repoDir, "init")
	runGitCommandForInternalTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitCommandForInternalTest(t, repoDir, "config", "user.name", "Test User")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "# Ship Check\n")
	runGitCommandForInternalTest(t, repoDir, "add", ".")
	runGitCommandForInternalTest(t, repoDir, "commit", "-m", "initial")
	writeTestFile(t, filepath.Join(repoDir, "main.go"), "package main\n")
	fakeBin := t.TempDir()
	fakeCodex := writeProviderFakeExecutable(t, fakeBin, "codex", "printf 'ok\\n'")

	t.Chdir(repoDir)
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"check", "ship", "--format=json"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected dirty repo ship check to block, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report shipCheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode ship check JSON: %v\n%s", err, stdout.String())
	}
	if report.ResultType != "JiniShipCheck" {
		t.Fatalf("expected JiniShipCheck result type, got %#v", report)
	}
	if report.Status != "blocked" {
		t.Fatalf("expected dirty repo to be blocked, got %#v", report)
	}
	if !report.InGitRepo || report.Branch == "" {
		t.Fatalf("expected git repo and branch evidence, got %#v", report)
	}
	if report.DirtyFiles != 1 || !stringSliceContains(report.Blockers, "working tree has uncommitted changes") {
		t.Fatalf("expected dirty file blocker, got %#v", report)
	}
	for _, want := range []string{
		"bash tools/run_required_gates.sh push",
		"git worktree add",
		"write validation report before push",
		"real installed CLI handoff dogfood for Wave 1 routes",
	} {
		if !stringSliceContains(report.RequiredEvidence, want) {
			t.Fatalf("expected ship check evidence %q, got %#v", want, report.RequiredEvidence)
		}
	}
	if !stringSliceContains(report.Next, "Record installed CLI dogfood evidence in .jini/cli-dogfood.json.") {
		t.Fatalf("expected ship check next steps to include dogfood evidence recording, got %#v", report.Next)
	}
	if len(report.CLIHandoffDogfood) != 5 {
		t.Fatalf("expected Wave 1 CLI dogfood matrix for 5 routes, got %#v", report.CLIHandoffDogfood)
	}
	codexDogfood := shipCLIHandoffDogfoodByRoute(report.CLIHandoffDogfood, "codex")
	if codexDogfood == nil {
		t.Fatalf("expected codex dogfood row, got %#v", report.CLIHandoffDogfood)
	}
	if codexDogfood.Status != "ready" || codexDogfood.Executable != fakeCodex {
		t.Fatalf("expected fake codex to be ready, got %#v", codexDogfood)
	}
	if codexDogfood.SetupStatus != "ready" || codexDogfood.DogfoodStatus != "needs-validation" {
		t.Fatalf("expected fake codex to separate setup readiness from dogfood validation, got %#v", codexDogfood)
	}
	for _, want := range []string{"auth", "approvals", "output shape", "route receipt privacy"} {
		if !stringSliceContains(codexDogfood.RequiredChecks, want) {
			t.Fatalf("expected codex dogfood required check %q, got %#v", want, codexDogfood.RequiredChecks)
		}
		if !stringSliceContains(codexDogfood.MissingChecks, want) {
			t.Fatalf("expected codex dogfood missing check %q before validation evidence, got %#v", want, codexDogfood.MissingChecks)
		}
	}
	claudeDogfood := shipCLIHandoffDogfoodByRoute(report.CLIHandoffDogfood, "claude-code")
	if claudeDogfood == nil || claudeDogfood.Status != "needs-setup" || claudeDogfood.SetupStatus != "needs-setup" || claudeDogfood.DogfoodStatus != "setup-blocked" || len(claudeDogfood.Missing) == 0 {
		t.Fatalf("expected missing claude-code dogfood setup row, got %#v", claudeDogfood)
	}
}

func TestShipCheckReadsLocalCLIHandoffDogfoodEvidence(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommandForInternalTest(t, repoDir, "init")
	runGitCommandForInternalTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitCommandForInternalTest(t, repoDir, "config", "user.name", "Test User")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "# Ship Check\n")
	runGitCommandForInternalTest(t, repoDir, "add", ".")
	runGitCommandForInternalTest(t, repoDir, "commit", "-m", "initial")
	fakeBin := t.TempDir()
	fakeCodex := writeProviderFakeExecutable(t, fakeBin, "codex", "printf 'ok\\n'")
	stateDir := t.TempDir()
	writeTestFile(t, filepath.Join(stateDir, "cli-dogfood.json"), `{
  "schema_version": "0.1.0",
  "context_type": "JiniCLIHandoffDogfoodEvidence",
  "routes": {
    "codex": {
      "validated_at": "2026-06-09T00:00:00Z",
      "checks": ["auth", "approvals", "output shape", "route receipt privacy"]
    }
  }
}
`)

	t.Chdir(repoDir)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"check", "ship", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected clean repo ship check to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report shipCheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode ship check JSON: %v\n%s", err, stdout.String())
	}
	codexDogfood := shipCLIHandoffDogfoodByRoute(report.CLIHandoffDogfood, "codex")
	if codexDogfood == nil {
		t.Fatalf("expected codex dogfood row, got %#v", report.CLIHandoffDogfood)
	}
	if codexDogfood.SetupStatus != "ready" || codexDogfood.DogfoodStatus != "validated" || codexDogfood.LastValidatedAt != "2026-06-09T00:00:00Z" {
		t.Fatalf("expected codex dogfood evidence to validate route, got %#v", codexDogfood)
	}
	if len(codexDogfood.ValidatedChecks) != 4 || len(codexDogfood.MissingChecks) != 0 {
		t.Fatalf("expected all codex dogfood checks validated, got %#v", codexDogfood)
	}
	if codexDogfood.EvidencePath != filepath.Join(stateDir, "cli-dogfood.json") {
		t.Fatalf("expected dogfood evidence path, got %#v", codexDogfood)
	}
}

func TestShipCheckBlocksInstalledCLIHandoffSetupFailures(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommandForInternalTest(t, repoDir, "init")
	runGitCommandForInternalTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitCommandForInternalTest(t, repoDir, "config", "user.name", "Test User")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "# Ship Check\n")
	runGitCommandForInternalTest(t, repoDir, "add", ".")
	runGitCommandForInternalTest(t, repoDir, "commit", "-m", "initial")
	fakeBin := t.TempDir()

	t.Chdir(repoDir)
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CODEX_CLI", filepath.Join(fakeBin, "codex"))
	t.Setenv("JINI_CODEX_ARGS", `exec "{{prompt}`)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"check", "ship", "--format=json"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected installed CLI handoff setup failure to block ship check, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report shipCheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode ship check JSON: %v\n%s", err, stdout.String())
	}
	if report.Status != "blocked" {
		t.Fatalf("expected invalid CLI handoff setup to block, got %#v", report)
	}
	if !stringSliceContains(report.Blockers, "CLI handoff setup blocked for installed route: codex (invalid args)") {
		t.Fatalf("expected invalid args blocker, got %#v", report.Blockers)
	}
	for _, blocker := range report.Blockers {
		if strings.Contains(blocker, "claude-code") || strings.Contains(blocker, "gemini-cli") || strings.Contains(blocker, "aider") || strings.Contains(blocker, "opencode") {
			t.Fatalf("missing optional CLI executables should stay warnings, got blockers %#v", report.Blockers)
		}
	}
}

func TestShipCheckTextKeepsSafePushInstructionsCompact(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommandForInternalTest(t, repoDir, "init")
	runGitCommandForInternalTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitCommandForInternalTest(t, repoDir, "config", "user.name", "Test User")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "# Ship Check\n")
	runGitCommandForInternalTest(t, repoDir, "add", ".")
	runGitCommandForInternalTest(t, repoDir, "commit", "-m", "initial")
	fakeBin := t.TempDir()
	fakeCodex := writeProviderFakeExecutable(t, fakeBin, "codex", "printf 'ok\\n'")

	t.Chdir(repoDir)
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK", "1")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"check", "ship"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected clean repo with undogfooded installed CLI to block, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Ship check blocked",
		"Branch:",
		"Blocker: CLI handoff dogfood missing validation for installed route: codex",
		"Run before push: bash tools/run_required_gates.sh push",
		"Safe lane: create an isolated worktree, run gates, then push only after evidence is clean.",
		"CLI handoff setup: 1 executable ready, 4 need setup",
		"CLI handoff dogfood: 0 validated, 1 need validation, 4 setup blocked",
		"CLI handoff routes:",
		"- codex: executable ready, dogfood needs validation",
		"- claude-code: needs setup (missing executable)",
		"- gemini-cli: needs setup (missing executable)",
		"- aider: needs setup (missing executable)",
		"- opencode: needs setup (missing executable)",
		"Dogfood before release: verify auth, approvals, output shape, and route receipt privacy on real installed CLIs.",
		"Evidence file: .jini/cli-dogfood.json",
		"Evidence checks: auth, approvals, output shape, route receipt privacy",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected compact ship check text %q, got:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Working Draft", "Start/Keep", "Actions"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("expected ship check to avoid %q, got:\n%s", unwanted, out)
		}
	}
	if strings.Contains(out, fakeCodex) {
		t.Fatalf("expected compact ship check text to avoid local executable path %q, got:\n%s", fakeCodex, out)
	}
}

func TestRouteDogfoodShowsGatekeeperSetupHint(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("JINI_STATE_DIR", stateDir)
	fakeBin := t.TempDir()
	fakeCodex := writeProviderFakeExecutable(t, fakeBin, "codex", "printf 'fake codex\\n'")
	t.Setenv("JINI_CODEX_CLI", fakeCodex)
	t.Setenv("JINI_CLAUDE_CODE_CLI", filepath.Join(fakeBin, "missing-claude"))
	t.Setenv("JINI_GEMINI_CLI", filepath.Join(fakeBin, "missing-gemini"))
	t.Setenv("JINI_AIDER_CLI", filepath.Join(fakeBin, "missing-aider"))
	t.Setenv("JINI_OPENCODE_CLI", filepath.Join(fakeBin, "missing-opencode"))
	previousTrustCheck := cliHandoffTrustIssueForPath
	cliHandoffTrustIssueForPath = func(path string) string {
		if path == fakeCodex {
			return "macOS Gatekeeper rejected CLI executable: " + path + "."
		}
		return ""
	}
	t.Cleanup(func() { cliHandoffTrustIssueForPath = previousTrustCheck })

	var stdout bytes.Buffer
	exitCode := Run([]string{"route", "dogfood"}, &stdout, &stdout)
	if exitCode != 0 {
		t.Fatalf("expected route dogfood guide to succeed, got %d with output:\n%s", exitCode, stdout.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"- codex: setup blocked (macOS Gatekeeper)",
		"Setup fixes:",
		"- codex: reinstall from a trusted source, open the CLI once, approve it in macOS Privacy & Security if prompted, then rerun `jini route dogfood`.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected Gatekeeper dogfood output to contain %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, fakeCodex) {
		t.Fatalf("route dogfood text must not leak executable path %q, got:\n%s", fakeCodex, out)
	}
}

func shipCLIHandoffDogfoodByRoute(items []shipCLIHandoffDogfood, routeID string) *shipCLIHandoffDogfood {
	for i := range items {
		if items[i].RouteID == routeID {
			return &items[i]
		}
	}
	return nil
}

func runGitCommandForInternalTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func TestScorecardGatePassesAndExposesCompetitorPressure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"scorecard-gate", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected scorecard-gate to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report scorecardGateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode scorecard-gate JSON: %v\n%s", err, stdout.String())
	}
	if report.Status != "ok" {
		t.Fatalf("expected scorecard status ok, got %#v", report)
	}
	if report.PRDImplementation.SourcePath != "specs/prd-implementation-trace.md" {
		t.Fatalf("expected PRD implementation trace source, got %#v", report.PRDImplementation)
	}
	if report.PRDImplementation.TotalRequirements != 13 || report.PRDImplementation.ImplementedRequirements != 13 || report.PRDImplementation.CompletionPercent != 100 {
		t.Fatalf("expected P0 PRD implementation completion to be 13/13 = 100%%, got %#v", report.PRDImplementation)
	}
	if report.PRDImplementation.Status != "ok" {
		t.Fatalf("expected PRD implementation status ok, got %#v", report.PRDImplementation)
	}
	if report.PRDImplementation.ResidualHardeningCount != 1 {
		t.Fatalf("expected residual hardening count to stay visible, got %#v", report.PRDImplementation)
	}
	if len(report.PRDImplementation.ResidualHardening) != 1 {
		t.Fatalf("expected residual hardening details to stay machine-readable, got %#v", report.PRDImplementation)
	}
	for _, want := range []string{
		"Wave 1 command templates use fake downstream CLIs in automated tests and now expose `jini check ship --format json` setup status plus local `.jini/cli-dogfood.json` validation evidence. Real installed CLI dogfood remains required before release claims for auth, approvals, output-shape differences, and route receipt privacy.",
	} {
		if !containsString(report.PRDImplementation.ResidualHardening, want) {
			t.Fatalf("expected residual hardening details to contain %q, got %#v", want, report.PRDImplementation.ResidualHardening)
		}
	}
	requiredCompetitors := map[string]bool{
		"claude-code":                 false,
		"codex":                       false,
		"github-copilot-coding-agent": false,
		"google-jules":                false,
		"cursor":                      false,
		"windsurf":                    false,
		"cline":                       false,
		"aider":                       false,
		"continue":                    false,
		"devin":                       false,
		"replit-agent":                false,
		"goose":                       false,
		"opencode":                    false,
		"sourcegraph-amp":             false,
		"tabnine-agent":               false,
		"qodo-merge":                  false,
		"ellipsis":                    false,
		"langgraph":                   false,
		"openai-agents-sdk":           false,
		"pydantic-ai":                 false,
		"crewai":                      false,
	}
	for _, competitor := range report.RequiredCompetitors {
		if _, ok := requiredCompetitors[competitor.ID]; ok {
			requiredCompetitors[competitor.ID] = competitor.Present
		}
	}
	for id, present := range requiredCompetitors {
		if !present {
			t.Fatalf("expected required competitor %q to be present in scorecard gate report: %#v", id, report.RequiredCompetitors)
		}
	}

	requiredVectors := map[string]bool{
		"async-background-agents":          false,
		"cross-surface-session-continuity": false,
		"offline-online-session-stitching": false,
		"transparent-progress-and-outputs": false,
		"permissioned-sandbox-execution":   false,
		"skills-hooks-and-context-routing": false,
		"local-open-model-optionality":     false,
		"token-frugality-p0":               false,
		"throttle-and-power-aware-routing": false,
		"commit-gated-scorecard-drift":     false,
		"intent-first-cli-parity":          false,
		"open-source-reference-priority":   false,
		"sub-agent-divide-and-conquer":     false,
	}
	for _, vector := range report.PressureVectors {
		if _, ok := requiredVectors[vector.ID]; ok {
			requiredVectors[vector.ID] = vector.Present
		}
	}
	for id, present := range requiredVectors {
		if !present {
			t.Fatalf("expected pressure vector %q to be present in scorecard gate report: %#v", id, report.PressureVectors)
		}
	}

	requiredOutcomeGates := map[string]bool{
		"direct-cwd-file-edit-fixture":         false,
		"simple-question-compact-answer":       false,
		"intent-first-routing-fixture":         false,
		"async-work-receipt-fixture":           false,
		"offline-route-proof-fixture":          false,
		"adversarial-code-review-fixture":      false,
		"competitor-watch-refresh-fixture":     false,
		"commercial-tier-boundary-fixture":     false,
		"cross-surface-continuity-fixture":     false,
		"token-frugality-route-proof-fixture":  false,
		"sub-agent-divide-and-conquer-fixture": false,
		"customer-value-viability-fixture":     false,
	}
	for _, gate := range report.OutcomeGates {
		if _, ok := requiredOutcomeGates[gate.ID]; ok {
			requiredOutcomeGates[gate.ID] = gate.Present
		}
	}
	for id, present := range requiredOutcomeGates {
		if !present {
			t.Fatalf("expected outcome gate %q to be present in scorecard gate report: %#v", id, report.OutcomeGates)
		}
	}
}

func TestScorecardGateBuilderSupportsCustomPolicy(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create specs dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "custom-scorecard.md"), "# Custom Scorecard\n")
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: custom-proof",
		"          kind: named-proof",
		"          ref: specs/custom-scorecard.md#custom-outcome",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      2,
		MinimumWatchlistCompetitors: 2,
		MinimumScenarios:            2,
	}).Build()

	if report.Status != "ok" {
		t.Fatalf("expected custom scorecard policy to pass, got %#v", report)
	}
	if report.ScorecardPath != scorecardPath {
		t.Fatalf("expected scorecard path %q, got %q", scorecardPath, report.ScorecardPath)
	}
	for _, check := range report.Checks {
		if check.Minimum != 1 {
			t.Fatalf("expected YAML minimum to override fallback for %s, got %#v", check.ID, check)
		}
	}
	if len(report.RequiredCompetitors) != 1 || !report.RequiredCompetitors[0].Present {
		t.Fatalf("expected custom competitor to be present, got %#v", report.RequiredCompetitors)
	}
	if len(report.PressureVectors) != 1 || !report.PressureVectors[0].Present {
		t.Fatalf("expected custom pressure vector to be present, got %#v", report.PressureVectors)
	}
	if len(report.OutcomeGates) != 1 || !report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be present, got %#v", report.OutcomeGates)
	}
}

func TestScorecardGateBuilderReportsIncompletePRDTraceCompletion(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create specs dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "custom-scorecard.md"), "# Custom Scorecard\n")
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: custom-proof",
		"          kind: named-proof",
		"          ref: specs/custom-scorecard.md#custom-outcome",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))
	prdTracePath := "specs/custom-prd-trace.md"
	writeTestFile(t, filepath.Join(root, prdTracePath), strings.Join([]string{
		"# PRD Implementation Trace",
		"",
		"| P0 requirement | Runtime surface | Proof |",
		"| --- | --- | --- |",
		"| Implemented requirement | runtime surface | proof gate |",
		"| Missing proof requirement | runtime surface | |",
		"",
		"Residual hardening:",
		"",
		"- Real-world dogfood still required.",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		PRDTracePath:                prdTracePath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected incomplete PRD trace to need attention, got %#v", report)
	}
	summary := report.PRDImplementation
	if summary.TotalRequirements != 2 || summary.ImplementedRequirements != 1 || summary.CompletionPercent != 50 {
		t.Fatalf("expected PRD completion to be 1/2 = 50%%, got %#v", summary)
	}
	if summary.Status != "needs-attention" || summary.ResidualHardeningCount != 1 {
		t.Fatalf("expected incomplete PRD status and residual count, got %#v", summary)
	}
	if len(summary.ResidualHardening) != 1 || summary.ResidualHardening[0] != "Real-world dogfood still required." {
		t.Fatalf("expected residual hardening detail to be preserved, got %#v", summary.ResidualHardening)
	}
	if len(summary.MissingImplementationDetails) != 1 || !strings.Contains(summary.MissingImplementationDetails[0], "Missing proof requirement missing proof") {
		t.Fatalf("expected missing proof detail, got %#v", summary.MissingImplementationDetails)
	}
}

func requireOutcomeGateReason(t *testing.T, report scorecardGateReport, gateID, want string) {
	t.Helper()
	for _, gate := range report.OutcomeGates {
		if gate.ID != gateID {
			continue
		}
		for _, reason := range gate.Reasons {
			if strings.Contains(reason, want) {
				return
			}
		}
		t.Fatalf("expected outcome gate %q to include reason containing %q, got %#v", gateID, want, gate.Reasons)
	}
	t.Fatalf("expected outcome gate %q in report, got %#v", gateID, report.OutcomeGates)
}

func TestScorecardGateBuilderRequiresOutcomeGateEvidenceReference(t *testing.T) {
	root := t.TempDir()
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      gate: Must not pass by id alone.",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected missing outcome evidence to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present || report.OutcomeGates[0].Status != "missing" {
		t.Fatalf("expected custom outcome gate to be missing without evidence, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", "missing proof_references")
}

func TestScorecardGateBuilderRejectsProofReferenceWithoutRef(t *testing.T) {
	root := t.TempDir()
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: custom-proof",
		"          kind: executable",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected proof reference without ref to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be missing without ref, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", "custom-proof is missing ref")
}

func TestScorecardGateBuilderRejectsNamedProofReferenceToMissingFile(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create specs dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "existing-scorecard-proof.md"), "# Existing Proof\n")
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: existing-proof",
		"          kind: named-proof",
		"          ref: specs/existing-scorecard-proof.md#custom-outcome",
		"        - id: missing-proof",
		"          kind: named-proof",
		"          ref: specs/missing-scorecard-proof.md#custom-outcome",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected missing named proof file to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be missing when named proof file is absent, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", "missing-proof references missing file specs/missing-scorecard-proof.md")
}

func TestScorecardGateBuilderRejectsUnsupportedProofReferenceKind(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("create specs dir: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, "custom-scorecard-proof.md"), "# Existing Proof\n")
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: unsupported-proof",
		"          kind: unsupported",
		"          ref: specs/custom-scorecard-proof.md",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected unsupported proof kind to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be missing when proof kind is unsupported, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", `unsupported-proof has unsupported kind "unsupported"`)
}

func TestScorecardGateBuilderRejectsExecutableProofReferenceToMissingTest(t *testing.T) {
	root := t.TempDir()
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: custom-proof",
		"          kind: executable",
		"          ref: \"go test ./internal/app -run TestMissingScorecardProof\"",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected missing executable proof test to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be missing when executable proof test is absent, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", "custom-proof names missing Go test function TestMissingScorecardProof")
}

func TestScorecardGateBuilderRejectsExecutableProofReferenceWithoutGoTestRun(t *testing.T) {
	root := t.TempDir()
	scorecardPath := "custom-scorecard.yaml"
	writeTestFile(t, filepath.Join(root, scorecardPath), strings.Join([]string{
		"scorecard_gates:",
		"  minimum_core_competitors: 1",
		"  minimum_watchlist_competitors: 1",
		"  minimum_scenarios: 1",
		"  required_core_competitors:",
		"    - id: custom-competitor",
		"  required_pressure_vectors:",
		"    - id: custom-vector",
		"  required_outcome_gates:",
		"    - id: custom-outcome",
		"      proof_references:",
		"        - id: custom-proof",
		"          kind: executable",
		"          ref: \"make validate-scorecard\"",
		"core_benchmark_set:",
		"  - id: custom-competitor",
		"watchlist:",
		"  - id: custom-watch",
		"scenarios:",
		"  - id: custom-scenario",
		"    checks:",
		"      - id: custom-vector",
	}, "\n"))

	report := newScorecardGateBuilder(root, scorecardGatePolicy{
		BenchmarkPath:               scorecardPath,
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected executable proof without Go test run to need attention, got %#v", report)
	}
	if len(report.OutcomeGates) != 1 || report.OutcomeGates[0].Present {
		t.Fatalf("expected custom outcome gate to be missing when executable proof is not a Go test run, got %#v", report.OutcomeGates)
	}
	requireOutcomeGateReason(t, report, "custom-outcome", "custom-proof must use go test ./internal/app -run")
}

func TestScorecardGateTextShowsFailureReasons(t *testing.T) {
	report := scorecardGateReport{
		Status: "needs-attention",
		OutcomeGates: []scorecardPresenceCheck{
			{
				ID:      "custom-outcome",
				Present: false,
				Status:  "missing",
				Reasons: []string{
					"named-proof missing-proof references missing file specs/missing-scorecard-proof.md",
				},
			},
		},
	}

	var stdout bytes.Buffer
	renderScorecardGateText(&stdout, report)
	out := stdout.String()
	for _, want := range []string{
		"OUTCOME GATES",
		"  MISSING custom-outcome",
		"    REASON named-proof missing-proof references missing file specs/missing-scorecard-proof.md",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected scorecard-gate text to contain %q, got:\n%s", want, out)
		}
	}
}

func TestScorecardGateBuilderReportsMissingSource(t *testing.T) {
	report := newScorecardGateBuilder(t.TempDir(), scorecardGatePolicy{
		BenchmarkPath:               "missing-scorecard.yaml",
		RequiredCompetitors:         []string{"custom-competitor"},
		RequiredPressureVectors:     []string{"custom-vector"},
		RequiredOutcomeGates:        []string{"custom-outcome"},
		MinimumCoreCompetitors:      1,
		MinimumWatchlistCompetitors: 1,
		MinimumScenarios:            1,
	}).Build()

	if report.Status != "needs-attention" {
		t.Fatalf("expected missing scorecard to need attention, got %#v", report)
	}
	if len(report.Checks) != 1 || report.Checks[0].ID != "source-scorecard-readable" || report.Checks[0].Status != "missing" {
		t.Fatalf("expected missing source check, got %#v", report.Checks)
	}
}

func TestScorecardGateTextShowsCommitGatePressure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"scorecard-gate"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected scorecard-gate text to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"STATUS ok",
		"PRD IMPLEMENTATION",
		"  OK 13/13 P0 requirements implemented (100%)",
		"  SOURCE specs/prd-implementation-trace.md",
		"  RESIDUAL_HARDENING 1",
		"    RESIDUAL Wave 1 command templates use fake downstream CLIs in automated tests and now expose `jini check ship --format json` setup status plus local `.jini/cli-dogfood.json` validation evidence. Real installed CLI dogfood remains required before release claims for auth, approvals, output-shape differences, and route receipt privacy.",
		"COMPETITORS",
		"  OK github-copilot-coding-agent",
		"PRESSURE VECTORS",
		"  OK offline-online-session-stitching",
		"  OK token-frugality-p0",
		"  OK throttle-and-power-aware-routing",
		"  OK commit-gated-scorecard-drift",
		"  OK intent-first-cli-parity",
		"  OK sub-agent-divide-and-conquer",
		"OUTCOME GATES",
		"  OK direct-cwd-file-edit-fixture",
		"  OK simple-question-compact-answer",
		"  OK intent-first-routing-fixture",
		"  OK async-work-receipt-fixture",
		"  OK sub-agent-divide-and-conquer-fixture",
		"  OK customer-value-viability-fixture",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected scorecard-gate text to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPublishReadinessTextIncludesGuardrailCheckDetails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=text"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness text to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"  HONEST-AUDIT ok",
		"    OK specs/honest-system-audit.md#current-implementation-reality",
		"    OK specs/skills-and-delegation-slice.md#tier-boundary",
		"    OK specs/lean-platform-gate.md#command-surface-discipline",
		"    CLAIM P0 competitor watching STATUS guarded RUNTIME false",
		"    CLAIM Configured CLI handoff STATUS implemented RUNTIME true",
		"    CLAIM Native Go CLI STATUS implemented RUNTIME true",
		"  APP-PLATFORM ok",
		"    OK specs/app-platform-shipping-playbook.md#source-backed-inputs",
		"  OFFLINE-REGRESSION ok",
		"    OK specs/local-model-support-matrix.md#promotion-loop",
		"  COMPETITIVE-PRESSURE ok",
		"    OK specs/competitive-release-plan.md#requirement-rejection-filter",
		"    OK specs/number-one-platform-prd.md#market-and-learning-guards",
		"  PRODUCTIVITY-LEARNING ok",
		"    OK specs/number-one-platform-prd.md#market-and-learning-guards",
		"    OK specs/learning-system.md#user-context-productivity-learning",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected publish-readiness text to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPublishReadinessHonestAuditClaimsExposeImplementationTruth(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}

	claims := map[string]publishEvidenceClaim{}
	for _, section := range report.Sections {
		if section.ID != "honest-audit" {
			continue
		}
		for _, claim := range section.Claims {
			claims[claim.Claim] = claim
		}
		if section.Details["guarded_claims"].Value < 1 {
			t.Fatalf("expected guarded claim count in honest audit details, got %#v", section.Details)
		}
	}

	competitor := claims["P0 competitor watching"]
	if competitor.Status != "guarded" || competitor.RuntimeImplemented {
		t.Fatalf("expected competitor watching to be guarded but not runtime implemented, got %#v", competitor)
	}
	if !strings.Contains(competitor.Gap, "No watch packet generator") {
		t.Fatalf("expected competitor watching gap to stay explicit, got %#v", competitor)
	}

	cliHandoff := claims["Configured CLI handoff"]
	if cliHandoff.Status != "implemented" || !cliHandoff.RuntimeImplemented {
		t.Fatalf("expected configured CLI handoff to be implemented with runtime evidence, got %#v", cliHandoff)
	}
	if !strings.Contains(cliHandoff.Evidence, "Wave 0 handoff contract") ||
		!strings.Contains(cliHandoff.Evidence, ".jini/cli-dogfood.json validation evidence") ||
		!strings.Contains(cliHandoff.Gap, "real-world dogfood") ||
		!strings.Contains(cliHandoff.Gap, "route receipt privacy") {
		t.Fatalf("expected configured CLI handoff gap to stay explicit, got %#v", cliHandoff)
	}

	goCLI := claims["Native Go CLI"]
	if goCLI.Status != "implemented" || !goCLI.RuntimeImplemented {
		t.Fatalf("expected native Go CLI to be implemented runtime evidence, got %#v", goCLI)
	}
}

func TestPublishHonestAuditClaimsRemainVisibleWithoutSourceCheckout(t *testing.T) {
	section := buildPublishHonestAuditSection("")
	if section.Status != "ok" {
		t.Fatalf("expected installed-binary honest audit section to be ok, got %#v", section)
	}
	if len(section.Claims) == 0 {
		t.Fatalf("expected installed-binary honest audit section to include embedded claims")
	}
	if section.Details["total_claims"].Value != len(section.Claims) {
		t.Fatalf("expected total claim metric to match claims, got details=%#v claims=%#v", section.Details, section.Claims)
	}
	for _, claim := range section.Claims {
		if claim.Claim == "P0 competitor watching" && claim.RuntimeImplemented {
			t.Fatalf("expected embedded competitor watching claim to remain non-runtime, got %#v", claim)
		}
	}
}

func TestPublishReadinessIncludesHonestAuditGuardrails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}
	required := map[string]bool{
		"specs/honest-system-audit.md#current-implementation-reality": false,
		"specs/honest-system-audit.md#core-feedback-accommodations":   false,
		"specs/skills-and-delegation-slice.md#tier-boundary":          false,
		"specs/lean-platform-gate.md#command-surface-discipline":      false,
	}
	for _, section := range report.Sections {
		if section.ID != "honest-audit" {
			continue
		}
		if section.Status != "ok" {
			t.Fatalf("expected honest-audit section to be ok, got %#v", section)
		}
		for _, check := range section.Checks {
			if _, ok := required[check.Path]; !ok {
				continue
			}
			if !check.Exists || check.Status != "ok" {
				t.Fatalf("expected honest-audit check to be ok, got %#v", check)
			}
			required[check.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			t.Fatalf("missing honest-audit check: %s\nreport: %#v", path, report.Sections)
		}
	}
}

func TestPublishReadinessIncludesOfflineRegressionGuardrails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}
	required := map[string]bool{
		"specs/local-model-support-matrix.md#registry-contract":   false,
		"specs/local-model-support-matrix.md#promotion-loop":      false,
		"specs/platform-offline-strategy.md#future-update-policy": false,
		"specs/adapter-benchmark-gate.md#routing-use":             false,
	}
	for _, section := range report.Sections {
		if section.ID != "offline-regression" {
			continue
		}
		if section.Status != "ok" {
			t.Fatalf("expected offline-regression section to be ok, got %#v", section)
		}
		for _, check := range section.Checks {
			if _, ok := required[check.Path]; !ok {
				continue
			}
			if !check.Exists || check.Status != "ok" {
				t.Fatalf("expected offline-regression check to be ok, got %#v", check)
			}
			required[check.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			t.Fatalf("missing offline-regression check: %s\nreport: %#v", path, report.Sections)
		}
	}
}

func TestPublishReadinessIncludesAppPlatformShippingGuardrails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}
	required := map[string]bool{
		"specs/app-platform-shipping-playbook.md#default-stack-decision":                false,
		"specs/app-platform-shipping-playbook.md#security-baseline":                     false,
		"specs/app-platform-shipping-playbook.md#performance-and-optimization-baseline": false,
		"specs/app-platform-shipping-playbook.md#logging-diagnostics-and-observability": false,
		"specs/app-platform-shipping-playbook.md#update-and-release-policy":             false,
		"specs/app-platform-shipping-playbook.md#app-shipping-gates":                    false,
		"specs/app-platform-shipping-playbook.md#source-backed-inputs":                  false,
	}
	for _, section := range report.Sections {
		if section.ID != "app-platform" {
			continue
		}
		if section.Status != "ok" {
			t.Fatalf("expected app-platform section to be ok, got %#v", section)
		}
		for _, check := range section.Checks {
			if _, ok := required[check.Path]; !ok {
				continue
			}
			if !check.Exists || check.Status != "ok" {
				t.Fatalf("expected app-platform check to be ok, got %#v", check)
			}
			required[check.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			t.Fatalf("missing app-platform check: %s\nreport: %#v", path, report.Sections)
		}
	}
}

func TestPublishReadinessIncludesCompetitivePressureGuardrails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}
	required := map[string]bool{
		"specs/competitive-release-plan.md#competitive-universe":         false,
		"specs/competitive-release-plan.md#requirement-rejection-filter": false,
		"specs/competitive-release-plan.md#p0-feature-selection-loop":    false,
		"specs/number-one-platform-prd.md#market-and-learning-guards":    false,
	}
	for _, section := range report.Sections {
		if section.ID != "competitive-pressure" {
			continue
		}
		if section.Status != "ok" {
			t.Fatalf("expected competitive-pressure section to be ok, got %#v", section)
		}
		for _, check := range section.Checks {
			if _, ok := required[check.Path]; !ok {
				continue
			}
			if !check.Exists || check.Status != "ok" {
				t.Fatalf("expected competitive-pressure check to be ok, got %#v", check)
			}
			required[check.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			t.Fatalf("missing competitive-pressure check: %s\nreport: %#v", path, report.Sections)
		}
	}
}

func TestPublishReadinessIncludesProductivityLearningGuardrails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"publish-readiness", "--format=json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected publish-readiness to pass, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	var report publishReadinessReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode publish-readiness JSON: %v\n%s", err, stdout.String())
	}
	required := map[string]bool{
		"specs/number-one-platform-prd.md#market-and-learning-guards": false,
		"specs/learning-system.md#user-context-productivity-learning": false,
	}
	for _, section := range report.Sections {
		if section.ID != "productivity-learning" {
			continue
		}
		if section.Status != "ok" {
			t.Fatalf("expected productivity-learning section to be ok, got %#v", section)
		}
		for _, check := range section.Checks {
			if _, ok := required[check.Path]; !ok {
				continue
			}
			if !check.Exists || check.Status != "ok" {
				t.Fatalf("expected productivity-learning check to be ok, got %#v", check)
			}
			required[check.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			t.Fatalf("missing productivity-learning check: %s\nreport: %#v", path, report.Sections)
		}
	}
}

func TestGoldenBenchmarkCoversOfflineRegressionReadinessChecks(t *testing.T) {
	root := discoverSourceRoot()
	if root == "" {
		t.Fatal("expected source root for golden benchmark coverage test")
	}
	benchmark, err := os.ReadFile(filepath.Join(root, "specs", "golden-competitive-benchmark.yaml"))
	if err != nil {
		t.Fatalf("read golden benchmark: %v", err)
	}

	section := buildPublishOfflineRegressionSection(root)
	if section.Status != "ok" {
		t.Fatalf("expected offline-regression section to be ok, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "source checkout" {
			continue
		}
		if !strings.Contains(string(benchmark), `contains: "`+check.Path+`"`) {
			t.Fatalf("golden benchmark does not cover offline-regression readiness check %q", check.Path)
		}
	}
}

func TestGoldenBenchmarkCoversHonestAuditReadinessChecks(t *testing.T) {
	root := discoverSourceRoot()
	if root == "" {
		t.Fatal("expected source root for golden benchmark coverage test")
	}
	benchmark, err := os.ReadFile(filepath.Join(root, "specs", "golden-competitive-benchmark.yaml"))
	if err != nil {
		t.Fatalf("read golden benchmark: %v", err)
	}

	section := buildPublishHonestAuditSection(root)
	if section.Status != "ok" {
		t.Fatalf("expected honest-audit section to be ok, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "source checkout" {
			continue
		}
		if !strings.Contains(string(benchmark), `contains: "`+check.Path+`"`) {
			t.Fatalf("golden benchmark does not cover honest-audit readiness check %q", check.Path)
		}
	}
	for _, claim := range section.Claims {
		if !strings.Contains(string(benchmark), `contains: "`+claim.Claim+`"`) {
			t.Fatalf("golden benchmark does not cover honest-audit claim %q", claim.Claim)
		}
	}
}

func TestGoldenBenchmarkCoversCompetitivePressureReadinessChecks(t *testing.T) {
	root := discoverSourceRoot()
	if root == "" {
		t.Fatal("expected source root for golden benchmark coverage test")
	}
	benchmark, err := os.ReadFile(filepath.Join(root, "specs", "golden-competitive-benchmark.yaml"))
	if err != nil {
		t.Fatalf("read golden benchmark: %v", err)
	}

	section := buildPublishCompetitivePressureSection(root)
	if section.Status != "ok" {
		t.Fatalf("expected competitive-pressure section to be ok, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "source checkout" {
			continue
		}
		if !strings.Contains(string(benchmark), `contains: "`+check.Path+`"`) {
			t.Fatalf("golden benchmark does not cover competitive-pressure readiness check %q", check.Path)
		}
	}
}

func TestGoldenBenchmarkCoversProductivityLearningReadinessChecks(t *testing.T) {
	root := discoverSourceRoot()
	if root == "" {
		t.Fatal("expected source root for golden benchmark coverage test")
	}
	benchmark, err := os.ReadFile(filepath.Join(root, "specs", "golden-competitive-benchmark.yaml"))
	if err != nil {
		t.Fatalf("read golden benchmark: %v", err)
	}

	section := buildPublishProductivityLearningSection(root)
	if section.Status != "ok" {
		t.Fatalf("expected productivity-learning section to be ok, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "source checkout" {
			continue
		}
		if !strings.Contains(string(benchmark), `contains: "`+check.Path+`"`) {
			t.Fatalf("golden benchmark does not cover productivity-learning readiness check %q", check.Path)
		}
	}
}

func TestGoldenBenchmarkCoversAppPlatformReadinessChecks(t *testing.T) {
	root := discoverSourceRoot()
	if root == "" {
		t.Fatal("expected source root for golden benchmark coverage test")
	}
	benchmark, err := os.ReadFile(filepath.Join(root, "specs", "golden-competitive-benchmark.yaml"))
	if err != nil {
		t.Fatalf("read golden benchmark: %v", err)
	}

	section := buildPublishAppPlatformSection(root)
	if section.Status != "ok" {
		t.Fatalf("expected app-platform section to be ok, got %#v", section)
	}
	for _, check := range section.Checks {
		if check.Path == "source checkout" {
			continue
		}
		if !strings.Contains(string(benchmark), `contains: "`+check.Path+`"`) {
			t.Fatalf("golden benchmark does not cover app-platform readiness check %q", check.Path)
		}
	}
}
