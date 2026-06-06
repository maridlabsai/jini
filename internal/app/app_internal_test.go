package app

import (
	"bytes"
	"encoding/json"
	"os"
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
			return
		}
	}
	t.Fatalf("expected promotion loop check to be incomplete, got %#v", section.Checks)
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
		{"commands", "extra"},
		{"status", "extra"},
		{"route", "extra"},
		{"observe", "status", "extra"},
		{"open", "--print-path"},
		{"publish-readiness", "extra"},
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
