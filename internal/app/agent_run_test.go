package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCommandExecutesUnderAutonomous(t *testing.T) {
	out, err := toolRunCommand().Run(t.TempDir(), map[string]string{"command": "echo hello-loop"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello-loop") {
		t.Fatalf("command output wrong: %q", out)
	}
}

func TestRunCommandFailureFedBackNotHardError(t *testing.T) {
	out, err := toolRunCommand().Run(t.TempDir(), map[string]string{"command": "exit 3"})
	if err != nil {
		t.Fatalf("command failure must be an observation, not a hard error: %v", err)
	}
	if !strings.Contains(out, "error") {
		t.Fatalf("failure observation should note the error: %q", out)
	}
}

func TestRunCommandRefusedUnderSemi(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran.txt")
	model := scriptedModel(
		"ACTION: run_command\ncommand: touch "+marker,
		"ACTION: finish\nsummary: done",
	)
	_, report, err := runAgentLoop(context.Background(), "try run", agentLoopOptions{
		posture: postureSemi, tools: allAgentTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Steps) == 0 || !strings.Contains(report.Steps[0].Observation, "unavailable") {
		t.Fatalf("semi posture must refuse run_command: %+v", report.Steps)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("semi posture must not run the command")
	}
}

func TestLoopSelfVerifyEditThenRun(t *testing.T) {
	dir := t.TempDir()
	// Edit a script, then run it to verify — the self-verify pattern.
	model := scriptedModel(
		"ACTION: edit_file\npath: check.sh\n```\necho VERIFIED\n```",
		"ACTION: run_command\ncommand: sh check.sh",
		"ACTION: finish\nsummary: verified",
	)
	_, report, err := runAgentLoop(context.Background(), "edit then verify", agentLoopOptions{
		posture: postureAutonomous, tools: allAgentTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Finished || len(report.Steps) != 2 {
		t.Fatalf("expected edit+run then finish, got %+v", report)
	}
	if !strings.Contains(report.Steps[1].Observation, "VERIFIED") {
		t.Fatalf("self-verify run did not observe output: %q", report.Steps[1].Observation)
	}
}
