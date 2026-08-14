package app

// Native loop routing entry — specs/native-agentic-loop-design.md, P5. Engages
// Jini's own agentic loop for non-handoff WORK tasks, but only when the cwd is
// trusted (posture semi/autonomous) and a real model route is available. In the
// default/untrusted case posture is plan and this returns ok=false, so today's
// behavior is byte-identical — the fail-safe held throughout the feature.

import (
	"context"
	"fmt"
	"io"
	"os"
)

// resolveNativeLoopPosture maps the granted trust level for a directory straight
// to a loop posture. Unlike the CLI hand-off resolver, there is no downstream
// descriptor to verify args against — Jini itself runs the tools, and it has
// verified tools for every posture (read/edit/run). Auto mode + a grant are
// still required; anything else is plan (loop stays off).
func resolveNativeLoopPosture(dir string) handoffPosture {
	if effectiveExecutionMode() != executionModeAuto {
		return posturePlan
	}
	level, ok := trustedLevelForDir(dir)
	if !ok {
		return posturePlan
	}
	switch level {
	case trustLevelAutonomous:
		return postureAutonomous
	case trustLevelSemi:
		return postureSemi
	}
	return posturePlan
}

// newNativeLoopModel builds the model callback the loop drives. It is a package
// var so tests can inject a scripted model without a live provider; production
// wraps generateProviderText.
var newNativeLoopModel = func(provider providerConfig, request providerGenerationRequest) agentModelFunc {
	return func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
		return generateProviderText(ctx, provider, request, systemPrompt, userPrompt)
	}
}

// maybeRunNativeLoop runs the native agentic loop when the directory is trusted
// and a usable non-handoff model is available. Returns (exitCode, true) when it
// handled the task, or (0, false) to fall through to the existing path.
func maybeRunNativeLoop(request providerGenerationRequest, decision routeDecision, stdout, stderr io.Writer) (int, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return 0, false
	}
	workDir := resolveTrustDir(cwd)

	posture := resolveNativeLoopPosture(cwd)
	if posture == posturePlan {
		return 0, false // untrusted / Ask → keep today's behavior
	}
	provider := providerForDecision(request, decision)
	if provider.ID == "local-preview" || provider.Status != "ok" {
		return 0, false // no usable model → fall through
	}

	if line := postureDisclosureLine(posture, "Jini native loop", workDir); line != "" {
		fmt.Fprintln(stdout, line)
	}

	model := newNativeLoopModel(provider, request)
	result, report, runErr := runAgentLoop(context.Background(), request.Source, agentLoopOptions{
		posture: posture,
		tools:   allAgentTools(),
		workDir: workDir,
		model:   model,
	})
	if runErr != nil {
		fmt.Fprintln(stderr, runErr.Error())
		return 1, true
	}
	if result != "" {
		fmt.Fprintln(stdout, result)
	}
	// A loop run is a work task on a paid-equivalent route → record savings
	// once (persist + attach), then print the footer from the recorded entry.
	recorded := recordSavingsOnDecision(decision, provider, len(request.Source), len(result), request, throttleSurvivalReport{})
	renderSavingsFooter(stdout, recorded.SavingsEntry)
	_ = report
	return 0, true
}
