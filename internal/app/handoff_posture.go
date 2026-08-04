package app

// Hand-off permission posture — specs/handoff-posture-design.md. Maps execution
// mode + explicit per-directory trust to the downstream CLI's permission mode.
// Fail-safe by construction: with no trust configured (the default) every route
// resolves to plan, byte-identical to today. Escalation requires Auto mode AND
// an explicit trust grant AND a descriptor with verified args for that level;
// it degrades DOWN when a route lacks verified args, never up.

import (
	"os"
	"strings"
)

type handoffPosture int

const (
	posturePlan handoffPosture = iota
	postureSemi
	postureAutonomous
)

func (p handoffPosture) String() string {
	switch p {
	case postureSemi:
		return "semi"
	case postureAutonomous:
		return "autonomous"
	default:
		return "plan"
	}
}

// postureArgs returns the descriptor args for a posture (nil for plan).
func postureArgs(descriptor cliHandoffDescriptor, posture handoffPosture) []string {
	switch posture {
	case postureAutonomous:
		return descriptor.AutonomousArgs
	case postureSemi:
		return descriptor.SemiArgs
	default:
		return nil
	}
}

// resolveHandoffPostureForDir resolves the posture for a specific directory.
// Split out for testability; resolveHandoffPosture uses the process cwd.
func resolveHandoffPostureForDir(descriptor cliHandoffDescriptor, dir string) handoffPosture {
	if effectiveExecutionMode() != executionModeAuto {
		return posturePlan
	}
	level, ok := trustedLevelForDir(dir)
	if !ok {
		return posturePlan
	}
	// Cap at the granted level, then degrade down to the highest posture the
	// descriptor actually has verified args for — never escalate above grant.
	switch level {
	case trustLevelAutonomous:
		if len(descriptor.AutonomousArgs) > 0 {
			return postureAutonomous
		}
		if len(descriptor.SemiArgs) > 0 {
			return postureSemi
		}
	case trustLevelSemi:
		if len(descriptor.SemiArgs) > 0 {
			return postureSemi
		}
	}
	return posturePlan
}

// resolveHandoffPosture resolves the posture for the current working directory.
func resolveHandoffPosture(descriptor cliHandoffDescriptor) handoffPosture {
	cwd, err := os.Getwd()
	if err != nil {
		return posturePlan
	}
	return resolveHandoffPostureForDir(descriptor, cwd)
}

// applyPostureArgs inserts the descriptor's posture args immediately before the
// first {{prompt}} placeholder in its default args (falling back to append
// before the last arg). plan returns the default args unchanged.
func applyPostureArgs(descriptor cliHandoffDescriptor, posture handoffPosture) []string {
	extra := postureArgs(descriptor, posture)
	if len(extra) == 0 {
		return descriptor.DefaultArgs
	}
	out := make([]string, 0, len(descriptor.DefaultArgs)+len(extra))
	inserted := false
	for _, arg := range descriptor.DefaultArgs {
		if !inserted && strings.Contains(arg, "{{prompt}}") {
			out = append(out, extra...)
			inserted = true
		}
		out = append(out, arg)
	}
	if !inserted {
		out = append(out, extra...)
	}
	return out
}
