package app

// Native agentic loop engine — specs/native-agentic-loop-design.md. Bounded
// read→reason→act loop over any model. Reuses generateProviderText for model
// calls (default); tests inject a scripted model. Only tools permitted by the
// posture are offered, so the same jini trust consent governs the loop.

import (
	"context"
	"fmt"
	"io"
	"strings"
)

const (
	agentDefaultMaxSteps   = 12
	agentDefaultMaxRepairs = 2
)

type agentModelFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)

// decisionRecorder is the seam the Pro decision-tree recorder plugs into (via
// the public agentloop package, bridged in agent_seam.go). nil = no-op.
type decisionRecorder interface {
	RecordStep(index int, tool string, args map[string]string, observation string)
}

type agentLoopOptions struct {
	posture    handoffPosture
	tools      []agentTool // full set; loop offers only those <= posture
	workDir    string
	model      agentModelFunc
	maxSteps   int
	maxRepairs int
	recorder   decisionRecorder // nil = no-op (seam for Pro; wired in P4)
	progress   io.Writer        // nil = silent; else one evidence line per action
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
	recorder := effectiveDecisionRecorder(opts.recorder)
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
		observation := runAgentTool(byName, opts, action, step)
		report.Steps = append(report.Steps, agentStep{Tool: action.Tool, Args: action.Args, Observation: observation})
		if recorder != nil {
			recorder.RecordStep(len(report.Steps)-1, action.Tool, action.Args, observation)
		}
		// Evidence: surface each action so autonomous execution is never hidden
		// (visible permissions/progress — the sandbox-execution competitive bar).
		if opts.progress != nil {
			fmt.Fprintln(opts.progress, agentStepLine(action, observation))
		}
		fmt.Fprintf(&transcript, "ACTION: %s\nOBSERVATION: %s\n", action.Tool, observation)
	}
	return "Reached the step limit before finishing.", report, nil
}

// agentStepLine renders one compact evidence line for an executed action. It
// names the action and its target (not the file contents), and marks a failed
// command, so a watching user sees exactly what the loop did.
func agentStepLine(action agentAction, observation string) string {
	failed := strings.HasPrefix(observation, "error:")
	switch action.Tool {
	case "read_file":
		return "· read " + action.Args["path"]
	case "list_dir":
		return "· listed " + firstNonEmpty(action.Args["path"], ".")
	case "search":
		return fmt.Sprintf("· searched %q", action.Args["query"])
	case "edit_file":
		return "· edited " + action.Args["path"]
	case "run_command":
		status := "ok"
		if failed {
			status = "failed"
		}
		return fmt.Sprintf("· ran: %s — %s", compactCommand(action.Args["command"]), status)
	default:
		return "· " + action.Tool
	}
}

// compactCommand trims a command to a single readable line for evidence output.
func compactCommand(cmd string) string {
	cmd = strings.TrimSpace(strings.ReplaceAll(cmd, "\n", " "))
	const max = 80
	if len(cmd) > max {
		return cmd[:max] + "…"
	}
	return cmd
}

// runAgentTool executes a parsed action, returning an observation. A tool the
// posture does not permit (or an unknown tool) yields an observation rather
// than an error, so the model can recover.
func runAgentTool(byName map[string]agentTool, opts agentLoopOptions, action agentAction, step int) string {
	tool, ok := byName[action.Tool]
	if !ok {
		return fmt.Sprintf("unknown or unavailable tool %q at the current permission level", action.Tool)
	}
	// Snapshot before a write step so the Pro backtrack can rewind to it
	// (no-op unless a checkpointer is registered).
	checkpointBeforeWrite(tool, fmt.Sprintf("before %s (step %d)", tool.Name, step+1))
	out, err := tool.Run(opts.workDir, action.Args)
	if err != nil {
		return "error: " + err.Error()
	}
	return out
}
