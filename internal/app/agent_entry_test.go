package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// trustCwdAutonomous grants the process cwd autonomous trust for the test and
// returns the cwd. Caller must already be chdir'd into an isolated temp dir.
func trustCwdAutonomous(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := saveTrustGrant(cwd, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	return cwd
}

func TestResolveNativeLoopPosture(t *testing.T) {
	withExecutionModeHome(t)

	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelSemi, "edits"); err != nil {
		t.Fatal(err)
	}
	auto := t.TempDir()
	if err := saveTrustGrant(auto, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}

	// Ask mode: trusted dirs still resolve to plan (loop stays off).
	if err := saveExecutionMode(executionModeAsk); err != nil {
		t.Fatal(err)
	}
	if p := resolveNativeLoopPosture(dir); p != posturePlan {
		t.Fatalf("Ask mode must be plan even when trusted, got %v", p)
	}

	setModeAuto(t)
	if p := resolveNativeLoopPosture(t.TempDir()); p != posturePlan {
		t.Fatalf("untrusted dir must be plan, got %v", p)
	}
	if p := resolveNativeLoopPosture(dir); p != postureSemi {
		t.Fatalf("semi grant → semi, got %v", p)
	}
	if p := resolveNativeLoopPosture(auto); p != postureAutonomous {
		t.Fatalf("autonomous grant → autonomous, got %v", p)
	}
}

// usableDecision is a route decision that resolves to a real (non-preview) model.
func usableDecision() routeDecision {
	return routeDecision{Active: true, Provider: providerConfig{ID: "anthropic", Status: "ok"}}
}

func TestMaybeRunNativeLoop_UntrustedFallsThrough(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	t.Chdir(t.TempDir()) // untrusted cwd

	var out, errBuf strings.Builder
	code, ok := maybeRunNativeLoop(providerGenerationRequest{Source: "do work"}, usableDecision(), &out, &errBuf)
	if ok {
		t.Fatalf("untrusted dir must fall through (ok=false), got ok=true code=%d", code)
	}
	if out.String() != "" {
		t.Fatalf("fall-through must print nothing, got %q", out.String())
	}
}

func TestMaybeRunNativeLoop_LocalPreviewFallsThrough(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	t.Chdir(t.TempDir())
	trustCwdAutonomous(t)

	preview := routeDecision{Active: true, Provider: providerConfig{ID: "local-preview", Status: "ok"}}
	var out, errBuf strings.Builder
	if _, ok := maybeRunNativeLoop(providerGenerationRequest{Source: "do work"}, preview, &out, &errBuf); ok {
		t.Fatalf("local-preview must fall through (no usable model)")
	}
	if out.String() != "" {
		t.Fatalf("fall-through must print nothing, got %q", out.String())
	}
}

func TestMaybeRunNativeLoop_TrustedRunsLoop(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	t.Chdir(t.TempDir())
	trustCwdAutonomous(t)

	// Inject a scripted model: one edit step, then finish.
	prev := newNativeLoopModel
	newNativeLoopModel = func(providerConfig, providerGenerationRequest) agentModelFunc {
		return scriptedModel(
			"ACTION: edit_file\npath: out.txt\n```\nhello\n```",
			"ACTION: finish\nsummary: wrote out.txt",
		)
	}
	t.Cleanup(func() { newNativeLoopModel = prev })

	var out, errBuf strings.Builder
	code, ok := maybeRunNativeLoop(providerGenerationRequest{Source: "write a file"}, usableDecision(), &out, &errBuf)
	if !ok || code != 0 {
		t.Fatalf("trusted dir + usable model must run the loop: ok=%v code=%d stderr=%q", ok, code, errBuf.String())
	}
	if !strings.Contains(out.String(), "autonomous") {
		t.Fatalf("expected the autonomous posture disclosure, got %q", out.String())
	}
	if !strings.Contains(out.String(), "wrote out.txt") {
		t.Fatalf("expected the loop's finish summary, got %q", out.String())
	}
	cwd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(cwd, "out.txt")); err != nil {
		t.Fatalf("edit_file tool should have created out.txt: %v", err)
	}
}
