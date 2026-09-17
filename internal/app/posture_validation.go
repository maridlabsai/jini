package app

// Behavioral posture validation — proves a hand-off CLI actually honors each
// posture (specs/cli-handoff-compatibility-premortem.md) instead of trusting the
// docs. For each posture it runs the CLI in a throwaway workspace with a probe
// prompt and OBSERVES the filesystem effect:
//   - plan       → must leave the workspace unchanged (read-only; safety-critical)
//   - semi       → must apply the edit (when the route claims SemiArgs)
//   - autonomous → must apply the edit (when the route claims AutonomousArgs)
// Command execution is CLI-specific (aider runs no shell), so it is reported but
// not required. A route that passes can be marked PostureVerified.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// postureProbePrompt instructs the CLI to make one observable edit and run one
// observable command. Real CLIs interpret it via their model; test fakes act on
// the posture flags. Effects are judged by the marker files, not the output.
const postureProbePrompt = "Create a file named PROOF.txt containing OK, then run the shell command: echo done > CMD.txt"

const (
	postureProbeEditMarker    = "PROOF.txt"
	postureProbeCommandMarker = "CMD.txt"
)

type postureProbe struct {
	Posture    handoffPosture
	Edited     bool // the edit marker appeared
	CommandRan bool // the command marker appeared
	RunErr     error
}

// runCLIHandoffAtPosture runs the descriptor's CLI at an explicit posture (not
// the ambient trust/mode resolution) inside workDir. Effects are observed on
// disk, so the run error is captured but not fatal.
func runCLIHandoffAtPosture(ctx context.Context, descriptor cliHandoffDescriptor, posture handoffPosture, prompt, workDir string) error {
	executable := firstNonEmpty(configValue(descriptor.ExecutableEnv), descriptor.DefaultExecutable)
	args := cliHandoffArgsWithPrompt(applyPostureArgs(descriptor, posture), prompt)
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = workDir
	return cmd.Run()
}

// validateRoutePosture runs the three probes and returns them plus whether the
// route's posture behaves as claimed and a list of issues (empty when verified).
func validateRoutePosture(ctx context.Context, descriptor cliHandoffDescriptor) (probes []postureProbe, verified bool, issues []string) {
	for _, p := range []handoffPosture{posturePlan, postureSemi, postureAutonomous} {
		probes = append(probes, runOnePostureProbe(ctx, descriptor, p))
	}
	plan, semi, autonomous := probes[0], probes[1], probes[2]

	// Safety-critical, universal: plan must not touch the workspace.
	if plan.Edited || plan.CommandRan {
		issues = append(issues, "plan posture modified the workspace — it must be read-only")
	}
	// Edits must apply where the route claims an escalation posture.
	if len(descriptor.SemiArgs) > 0 && !semi.Edited {
		issues = append(issues, "semi posture did not apply the edit")
	}
	if len(descriptor.AutonomousArgs) > 0 && !autonomous.Edited {
		issues = append(issues, "autonomous posture did not apply the edit")
	}
	return probes, len(issues) == 0, issues
}

func runOnePostureProbe(ctx context.Context, descriptor cliHandoffDescriptor, posture handoffPosture) postureProbe {
	dir, err := os.MkdirTemp("", "jini-posture-")
	if err != nil {
		return postureProbe{Posture: posture, RunErr: err}
	}
	defer os.RemoveAll(dir)
	runErr := runCLIHandoffAtPosture(ctx, descriptor, posture, postureProbePrompt, dir)
	return postureProbe{
		Posture:    posture,
		Edited:     fileExistsInDir(dir, postureProbeEditMarker),
		CommandRan: fileExistsInDir(dir, postureProbeCommandMarker),
		RunErr:     runErr,
	}
}

func fileExistsInDir(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
