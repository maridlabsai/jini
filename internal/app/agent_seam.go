package app

// Bridge from the public agentloop seam into the internal loop. When the
// commercial module has registered a recorder/checkpointer, the loop records
// each step and checkpoints before write steps; otherwise both are no-ops.

import "github.com/maridlabsai/jini/agentloop"

type publicRecorderAdapter struct{ r agentloop.DecisionRecorder }

func (a publicRecorderAdapter) RecordStep(index int, tool string, args map[string]string, observation string) {
	a.r.RecordStep(agentloop.Step{Index: index, Tool: tool, Args: args, Observation: observation})
}

// effectiveDecisionRecorder returns the explicit opts recorder, else a bridge
// to the registered public recorder, else nil.
func effectiveDecisionRecorder(explicit decisionRecorder) decisionRecorder {
	if explicit != nil {
		return explicit
	}
	if r, ok := agentloop.RegisteredRecorder(); ok {
		return publicRecorderAdapter{r: r}
	}
	return nil
}

// checkpointBeforeWrite snapshots state before a write tool runs (best-effort).
func checkpointBeforeWrite(tool agentTool, label string) {
	if tool.MinPosture <= posturePlan {
		return // read-only tools need no checkpoint
	}
	if c, ok := agentloop.RegisteredCheckpointer(); ok {
		_ = c.Checkpoint(label)
	}
}
