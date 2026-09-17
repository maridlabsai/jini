package app

import (
	"os"
	"strings"
	"testing"
)

// claudeDescriptor is the verified route (both semi + autonomous args).
func claudeDescriptor(t *testing.T) cliHandoffDescriptor {
	t.Helper()
	d, ok := cliHandoffDescriptorForMode("claude-code")
	if !ok {
		t.Fatal("claude-code descriptor missing")
	}
	return d
}

func setModeAuto(t *testing.T) {
	t.Helper()
	if err := saveExecutionMode(executionModeAuto); err != nil {
		t.Fatal(err)
	}
}

func TestPosture_AskModeAlwaysPlan(t *testing.T) {
	withExecutionModeHome(t)
	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	if err := saveExecutionMode(executionModeAsk); err != nil {
		t.Fatal(err)
	}
	if p := resolveHandoffPostureForDir(claudeDescriptor(t), dir); p != posturePlan {
		t.Fatalf("Ask mode must be plan even when trusted, got %v", p)
	}
}

func TestPosture_AutoUntrustedIsPlan(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	if p := resolveHandoffPostureForDir(claudeDescriptor(t), t.TempDir()); p != posturePlan {
		t.Fatalf("untrusted dir must be plan, got %v", p)
	}
}

func TestPosture_AutoTrustedSemi(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelSemi, "edits"); err != nil {
		t.Fatal(err)
	}
	if p := resolveHandoffPostureForDir(claudeDescriptor(t), dir); p != postureSemi {
		t.Fatalf("trusted semi + verified route → semi, got %v", p)
	}
}

func TestPosture_AutoTrustedAutonomous(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	dir := t.TempDir()
	if err := saveTrustGrant(dir, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	if p := resolveHandoffPostureForDir(claudeDescriptor(t), dir); p != postureAutonomous {
		t.Fatalf("trusted autonomous + verified route → autonomous, got %v", p)
	}
}

func TestPosture_DegradesDownNeverUp(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	dir := t.TempDir()
	// Route verified for semi only; grant is autonomous → must degrade to semi.
	semiOnly := cliHandoffDescriptor{Mode: "semi-only", SemiArgs: []string{"--edits"}}
	if err := saveTrustGrant(dir, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	if p := resolveHandoffPostureForDir(semiOnly, dir); p != postureSemi {
		t.Fatalf("autonomous grant on semi-only route must degrade to semi, got %v", p)
	}
	// Route with no verified args → plan even when trusted.
	unverified := cliHandoffDescriptor{Mode: "unverified"}
	if p := resolveHandoffPostureForDir(unverified, dir); p != posturePlan {
		t.Fatalf("unverified route must be plan, got %v", p)
	}
}

func TestApplyPostureArgs_InsertsBeforePrompt(t *testing.T) {
	d := claudeDescriptor(t)
	got := applyPostureArgs(d, postureSemi)
	want := []string{"--print", "--permission-mode", "acceptEdits", "{{prompt}}"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("semi args wrong: %v", got)
	}
	if strings.Join(applyPostureArgs(d, posturePlan), " ") != "--print {{prompt}}" {
		t.Fatalf("plan must leave default args unchanged: %v", applyPostureArgs(d, posturePlan))
	}
	autonomous := applyPostureArgs(d, postureAutonomous)
	if strings.Join(autonomous, " ") != "--print --dangerously-skip-permissions {{prompt}}" {
		t.Fatalf("autonomous args wrong: %v", autonomous)
	}
}

func TestResolveCLIHandoffCommand_PostureAppliedAndOverrideWins(t *testing.T) {
	withExecutionModeHome(t)
	setModeAuto(t)
	// Trust cwd autonomous so the claude route escalates.
	cwd, _ := os.Getwd()
	if err := saveTrustGrant(cwd, trustLevelAutonomous, "edits+commands"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { removeTrustGrant(cwd) })
	d := claudeDescriptor(t)
	cmd, _ := resolveCLIHandoffCommand(d)
	if !containsArg(cmd.Args, "--dangerously-skip-permissions") {
		t.Fatalf("autonomous posture args not applied: %v", cmd.Args)
	}
	// Explicit ArgsEnv override → posture is a no-op.
	t.Setenv("JINI_CLAUDE_CODE_ARGS", "--print {{prompt}}")
	cmd2, _ := resolveCLIHandoffCommand(d)
	if containsArg(cmd2.Args, "--dangerously-skip-permissions") {
		t.Fatalf("explicit ArgsEnv must suppress posture args: %v", cmd2.Args)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
