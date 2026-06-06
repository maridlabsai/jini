package app

import (
	"bytes"
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
