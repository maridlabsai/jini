package app

import (
	"context"
	"testing"

	"github.com/maridlabsai/jini/agentloop"
)

type fakeRecorder struct{ steps []agentloop.Step }

func (f *fakeRecorder) RecordStep(s agentloop.Step) { f.steps = append(f.steps, s) }

type fakeCheckpointer struct{ labels []string }

func (f *fakeCheckpointer) Checkpoint(label string) error {
	f.labels = append(f.labels, label)
	return nil
}

func TestRegisteredRecorderReceivesSteps(t *testing.T) {
	rec := &fakeRecorder{}
	agentloop.RegisterRecorder(rec)
	t.Cleanup(func() { agentloop.RegisterRecorder(nil) })

	dir := t.TempDir()
	model := scriptedModel(
		"ACTION: list_dir\npath: .",
		"ACTION: finish\nsummary: done",
	)
	_, _, err := runAgentLoop(context.Background(), "look", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.steps) != 1 || rec.steps[0].Tool != "list_dir" || rec.steps[0].Index != 0 {
		t.Fatalf("recorder did not receive the step: %+v", rec.steps)
	}
}

func TestCheckpointerCalledBeforeWriteStepsOnly(t *testing.T) {
	cp := &fakeCheckpointer{}
	agentloop.RegisterCheckpointer(cp)
	t.Cleanup(func() { agentloop.RegisterCheckpointer(nil) })

	dir := t.TempDir()
	model := scriptedModel(
		"ACTION: list_dir\npath: .", // read — no checkpoint
		"ACTION: edit_file\npath: f.txt\n```\nhi\n```", // write — checkpoint
		"ACTION: finish\nsummary: done",
	)
	_, _, err := runAgentLoop(context.Background(), "edit", agentLoopOptions{
		posture: postureSemi, tools: allAgentTools(), workDir: dir, model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cp.labels) != 1 {
		t.Fatalf("checkpoint should fire once (before the edit), got %d: %v", len(cp.labels), cp.labels)
	}
}

func TestNoRegistrationIsNoOp(t *testing.T) {
	agentloop.RegisterRecorder(nil)
	agentloop.RegisterCheckpointer(nil)
	dir := t.TempDir()
	model := scriptedModel("ACTION: finish\nsummary: done")
	_, report, err := runAgentLoop(context.Background(), "noop", agentLoopOptions{
		posture: posturePlan, tools: readOnlyTools(), workDir: dir, model: model,
	})
	if err != nil || !report.Finished {
		t.Fatalf("unregistered seams must leave the loop working: %+v err=%v", report, err)
	}
}
