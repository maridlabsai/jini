package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseActionCapturesFencedBody(t *testing.T) {
	reply := "ACTION: edit_file\npath: calc.py\n```\ndef add(a, b):\n    return a + b\n```"
	a, ok := parseAction(reply)
	if !ok || a.Tool != "edit_file" || a.Args["path"] != "calc.py" {
		t.Fatalf("bad parse: %+v ok=%v", a, ok)
	}
	if a.Args["body"] != "def add(a, b):\n    return a + b" {
		t.Fatalf("body wrong: %q", a.Args["body"])
	}
}

func TestLoopEditsFileUnderSemi(t *testing.T) {
	dir := t.TempDir()
	model := scriptedModel(
		"ACTION: edit_file\npath: calc.py\n```\ndef add(a, b):\n    return a + b\n```",
		"ACTION: finish\nsummary: fixed add",
	)
	_, report, err := runAgentLoop(context.Background(), "fix add", agentLoopOptions{
		posture: postureSemi, tools: allAgentTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Finished {
		t.Fatalf("expected finish, got %+v", report)
	}
	data, err := os.ReadFile(filepath.Join(dir, "calc.py"))
	if err != nil || !strings.Contains(string(data), "return a + b") {
		t.Fatalf("edit not applied: %q err=%v", data, err)
	}
}

func TestLoopEditRefusedUnderPlan(t *testing.T) {
	dir := t.TempDir()
	model := scriptedModel(
		"ACTION: edit_file\npath: calc.py\n```\nx\n```",
		"ACTION: finish\nsummary: done",
	)
	_, report, err := runAgentLoop(context.Background(), "try edit", agentLoopOptions{
		posture: posturePlan, tools: allAgentTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	// edit_file is not offered in plan → observation says unavailable, file not written.
	if len(report.Steps) == 0 || !strings.Contains(report.Steps[0].Observation, "unavailable") {
		t.Fatalf("plan posture must refuse edit_file: %+v", report.Steps)
	}
	if _, err := os.Stat(filepath.Join(dir, "calc.py")); !os.IsNotExist(err) {
		t.Fatal("plan posture must not write a file")
	}
}

func TestEditFileConfinesPath(t *testing.T) {
	dir := t.TempDir()
	tool := toolEditFile()
	if _, err := tool.Run(dir, map[string]string{"path": "../escape.txt", "body": "x"}); err == nil {
		t.Fatal("edit_file must reject a path escape")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "escape.txt")); !os.IsNotExist(err) {
		t.Fatal("escape file must not be written")
	}
}
