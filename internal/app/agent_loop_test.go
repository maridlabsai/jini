package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scriptedModel returns queued replies in order, then repeats the last.
func scriptedModel(replies ...string) agentModelFunc {
	i := 0
	return func(_ context.Context, _, _ string) (string, error) {
		r := replies[i]
		if i < len(replies)-1 {
			i++
		}
		return r, nil
	}
}

func TestParseActionForgiving(t *testing.T) {
	a, ok := parseAction("THOUGHT: I'll read it\nACTION: read_file\npath: main.go")
	if !ok || a.Tool != "read_file" || a.Args["path"] != "main.go" {
		t.Fatalf("parse failed: %+v ok=%v", a, ok)
	}
	if _, ok := parseAction("no action here"); ok {
		t.Fatal("must not parse an action from prose")
	}
	// Last ACTION wins over earlier chatter.
	a2, _ := parseAction("ACTION: list_dir\nACTION: finish\nsummary: done")
	if a2.Tool != "finish" || a2.Args["summary"] != "done" {
		t.Fatalf("last action should win: %+v", a2)
	}
}

func TestLoopReadsThenFinishes(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("world"), 0o644)
	model := scriptedModel(
		"ACTION: read_file\npath: hello.txt",
		"ACTION: finish\nsummary: the file says world",
	)
	out, report, err := runAgentLoop(context.Background(), "read hello.txt", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Finished || !strings.Contains(out, "world") {
		t.Fatalf("expected finish with summary, got %q report=%+v", out, report)
	}
	if len(report.Steps) != 1 || report.Steps[0].Observation != "world" {
		t.Fatalf("read observation wrong: %+v", report.Steps)
	}
}

func TestLoopRepairsThenAdvisoryFallback(t *testing.T) {
	model := scriptedModel("I think you should refactor it.") // never a valid action
	out, report, err := runAgentLoop(context.Background(), "do something", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: t.TempDir(), model: model, maxRepairs: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Fallback || !strings.Contains(out, "refactor") {
		t.Fatalf("expected advisory fallback after repair cap, got %q report=%+v", out, report)
	}
}

func TestLoopStepCap(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644)
	model := scriptedModel("ACTION: read_file\npath: a.txt") // never finishes
	_, report, err := runAgentLoop(context.Background(), "loop", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: dir, model: model, maxSteps: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Finished || len(report.Steps) != 3 {
		t.Fatalf("expected 3 steps then cap, got %+v", report)
	}
}

func TestConfinePathRejectsEscapes(t *testing.T) {
	dir := t.TempDir()
	if _, err := confinePath(dir, "../secret"); err == nil {
		t.Fatal("must reject parent escape")
	}
	if _, err := confinePath(dir, "/etc/passwd"); err == nil {
		t.Fatal("must reject absolute outside path")
	}
	if _, err := confinePath(dir, "sub/ok.txt"); err != nil {
		t.Fatalf("must allow in-dir path: %v", err)
	}
}

func TestPostureGatesToolOffering(t *testing.T) {
	// A write tool at semi must not be offered in plan posture.
	writeTool := agentTool{Name: "edit_file", MinPosture: postureSemi}
	all := append(readOnlyTools(), writeTool)
	planTools := toolsForPosture(all, posturePlan)
	for _, tl := range planTools {
		if tl.Name == "edit_file" {
			t.Fatal("edit_file must not be offered in plan posture")
		}
	}
	semiTools := toolsForPosture(all, postureSemi)
	found := false
	for _, tl := range semiTools {
		if tl.Name == "edit_file" {
			found = true
		}
	}
	if !found {
		t.Fatal("edit_file must be offered in semi posture")
	}
}
