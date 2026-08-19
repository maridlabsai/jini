package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAgentStepLine_Rendering(t *testing.T) {
	cases := []struct {
		tool, obs string
		args      map[string]string
		want      string
	}{
		{"read_file", "…", map[string]string{"path": "a.go"}, "· read a.go"},
		{"list_dir", "…", map[string]string{}, "· listed ."},
		{"search", "…", map[string]string{"query": "TODO"}, `· searched "TODO"`},
		{"edit_file", "wrote 3 bytes", map[string]string{"path": "b.go"}, "· edited b.go"},
		{"run_command", "ok", map[string]string{"command": "go build"}, "· ran: go build — ok"},
		{"run_command", "error: exit 1", map[string]string{"command": "go test"}, "· ran: go test — failed"},
	}
	for _, c := range cases {
		got := agentStepLine(agentAction{Tool: c.tool, Args: c.args}, c.obs)
		if got != c.want {
			t.Fatalf("tool %s → %q, want %q", c.tool, got, c.want)
		}
	}
}

func TestAgentLoop_EmitsProgressEvidence(t *testing.T) {
	dir := t.TempDir()
	var progress bytes.Buffer
	model := scriptedModel(
		"ACTION: list_dir\npath: .",
		"ACTION: edit_file\npath: note.txt\n```\nhi\n```",
		"ACTION: finish\nsummary: done",
	)
	_, _, err := runAgentLoop(context.Background(), "make a note", agentLoopOptions{
		posture: postureSemi, tools: allAgentTools(), workDir: dir, model: model, progress: &progress,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := progress.String()
	if !strings.Contains(out, "· listed .") || !strings.Contains(out, "· edited note.txt") {
		t.Fatalf("expected per-step evidence lines, got:\n%q", out)
	}
	// finish is not an executed tool → no evidence line for it.
	if strings.Contains(out, "finish") {
		t.Fatalf("finish should not emit an evidence line, got:\n%q", out)
	}
}

func TestAgentLoop_NilProgressIsSilent(t *testing.T) {
	dir := t.TempDir()
	model := scriptedModel("ACTION: list_dir\npath: .", "ACTION: finish\nsummary: done")
	// progress nil → must not panic and must still complete.
	if _, report, err := runAgentLoop(context.Background(), "x", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: dir, model: model,
	}); err != nil || !report.Finished {
		t.Fatalf("nil progress must be a silent no-op: report=%+v err=%v", report, err)
	}
}
