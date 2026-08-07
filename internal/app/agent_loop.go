package app

// Native agentic loop engine — specs/native-agentic-loop-design.md. Bounded
// read→reason→act loop over any model. Reuses generateProviderText for model
// calls (default); tests inject a scripted model. Only tools permitted by the
// posture are offered, so the same jini trust consent governs the loop.

import (
	"context"
	"fmt"
	"strings"
)

const (
	agentDefaultMaxSteps   = 12
	agentDefaultMaxRepairs = 2
)

type agentModelFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)

// decisionRecorder is the seam the Pro decision-tree recorder plugs into
// (fleshed out in P4). nil = no-op. Declared here so the loop calls it from
// day one (specs/decision-tree-backtrack-design.md).
type decisionRecorder interface {
	RecordStep(tool string, args map[string]string, observation string)
}

type agentLoopOptions struct {
	posture    handoffPosture
	tools      []agentTool // full set; loop offers only those <= posture
	workDir    string
	model      agentModelFunc
	maxSteps   int
	maxRepairs int
	recorder   decisionRecorder // nil = no-op (seam for Pro; wired in P4)
}

type agentStep struct {
	Tool        string
	Args        map[string]string
	Observation string
}

type agentReport struct {
	Steps    []agentStep
	Finished bool // finished via the finish tool
	Fallback bool // degraded to a plain advisory answer
	Posture  handoffPosture
}

// runAgentLoop runs the bounded loop and returns the final text + a report.
func runAgentLoop(ctx context.Context, task string, opts agentLoopOptions) (string, agentReport, error) {
	if opts.maxSteps <= 0 {
		opts.maxSteps = agentDefaultMaxSteps
	}
	if opts.maxRepairs <= 0 {
		opts.maxRepairs = agentDefaultMaxRepairs
	}
	if opts.model == nil {
		return "", agentReport{Posture: opts.posture}, fmt.Errorf("agent loop requires a model")
	}
	offered := toolsForPosture(opts.tools, opts.posture)
	byName := map[string]agentTool{}
	for _, t := range offered {
		byName[t.Name] = t
	}
	report := agentReport{Posture: opts.posture}
	system := agentSystemPrompt(task, offered)
	var transcript strings.Builder
	repairs := 0

	for step := 0; step < opts.maxSteps; step++ {
		reply, err := opts.model(ctx, system, transcript.String())
		if err != nil {
			return "", report, err
		}
		action, ok := parseAction(reply)
		if !ok {
			repairs++
			if repairs > opts.maxRepairs {
				// Honest degradation: hand back the model's plain reply.
				report.Fallback = true
				return strings.TrimSpace(reply), report, nil
			}
			transcript.WriteString(agentRepairPrompt(offered) + "\n")
			continue
		}
		if action.Tool == "finish" {
			report.Finished = true
			return strings.TrimSpace(action.Args["summary"]), report, nil
		}
		observation := runAgentTool(byName, opts, action)
		report.Steps = append(report.Steps, agentStep{Tool: action.Tool, Args: action.Args, Observation: observation})
		if opts.recorder != nil {
			opts.recorder.RecordStep(action.Tool, action.Args, observation)
		}
		fmt.Fprintf(&transcript, "ACTION: %s\nOBSERVATION: %s\n", action.Tool, observation)
	}
	return "Reached the step limit before finishing.", report, nil
}

// runAgentTool executes a parsed action, returning an observation. A tool the
// posture does not permit (or an unknown tool) yields an observation rather
// than an error, so the model can recover.
func runAgentTool(byName map[string]agentTool, opts agentLoopOptions, action agentAction) string {
	tool, ok := byName[action.Tool]
	if !ok {
		return fmt.Sprintf("unknown or unavailable tool %q at the current permission level", action.Tool)
	}
	out, err := tool.Run(opts.workDir, action.Args)
	if err != nil {
		return "error: " + err.Error()
	}
	return out
}
