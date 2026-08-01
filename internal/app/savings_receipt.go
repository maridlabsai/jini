package app

// Per-task savings computation — specs/savings-ledger-mvp-design.md. The ledger
// records IMPUTED savings from free/subscription routes: money the user did
// NOT spend by routing off a metered API. Metered BYO usage the user actually
// paid for is not a saving and writes no entry (baselineForRoute → not ok).

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// attemptOnRoute runs a single-shot generation on a named fallback route mode,
// used by the throttle-survival switch executor (paid Autopilot). It does NOT
// wrap itself in another survival loop — one attempt, no recursion. Returns an
// error the survival loop treats as "switch failed, hold the original route".
func attemptOnRoute(ctx context.Context, mode string, request providerGenerationRequest) (string, error) {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return "", fmt.Errorf("no switch route")
	}
	if cliHandoffMode(mode) {
		prompt := firstNonEmpty(strings.TrimSpace(request.Source), providerUserPrompt(request))
		text, _, err := runCLIHandoff(ctx, mode, prompt)
		return text, err
	}
	fallback := enrichRouteDecisionForRequest(request, detectRouteForToolMode(mode, true))
	provider := providerForDecision(request, fallback)
	if provider.ID == "local-preview" || provider.Status != "ok" {
		return "", fmt.Errorf("switch route %q not ready", mode)
	}
	return generateProviderText(ctx, provider, request, providerSystemPrompt(), providerUserPrompt(request))
}

// routeClassForDecision classifies the route that answered the task.
func routeClassForDecision(decision routeDecision, provider providerConfig) string {
	if cliHandoffMode(decision.ToolMode) {
		return savingsRouteSubscription
	}
	if provider.ID == "local-preview" {
		return savingsRouteLocal
	}
	return savingsRouteMetered
}

func savingsRepoName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

// computeTaskSavings builds the imputed savings entry for a completed work
// task, or nil when the route earns no imputed saving (metered/unknown, or a
// zero estimate).
func computeTaskSavings(decision routeDecision, provider providerConfig, inChars, outChars int, request providerGenerationRequest) *savingsEntry {
	class := routeClassForDecision(decision, provider)
	label := firstNonEmpty(decision.ToolLabel, provider.Label, provider.ID)
	if cliHandoffMode(decision.ToolMode) {
		label = cliHandoffLabel(decision.ToolMode)
	}
	baseline, ok := baselineForRoute(class, label)
	if !ok {
		return nil
	}
	inTok := estTokensFromChars(inChars)
	outTok := estTokensFromChars(outChars)
	usd := savingsUSD(baseline, inTok, outTok)
	if usd <= 0 {
		return nil
	}
	return &savingsEntry{
		SchemaVersion:   savingsLedgerSchema,
		ContextType:     savingsEntryContext,
		At:              time.Now().Format(time.RFC3339),
		Repo:            savingsRepoName(),
		RouteLabel:      label,
		RouteClass:      class,
		EstInputTokens:  inTok,
		EstOutputTokens: outTok,
		USDSaved:        usd,
		Imputed:         true,
		Title:           request.Title,
	}
}

// recordSavingsOnDecision computes, persists (best-effort), and attaches the
// savings entry. Simple questions stay clean — no entry, no footer. Persistence
// failures never block a task. Called exactly once, at a task's success return.
// When survival records a throttle dodge (paid Autopilot switched routes), the
// entry is flagged dodged and attributed to the route that actually answered.
func recordSavingsOnDecision(decision routeDecision, provider providerConfig, inChars, outChars int, request providerGenerationRequest, survival throttleSurvivalReport) routeDecision {
	if request.Standalone {
		return decision
	}
	entry := computeTaskSavings(decision, provider, inChars, outChars, request)
	if entry == nil {
		return decision
	}
	if survival.Dodged {
		entry.ThrottleDodged = true
		if survival.SwitchedTo != "" {
			entry.RouteLabel = survival.SwitchedTo
		}
	}
	_ = appendSavingsEntry(*entry)
	decision.SavingsEntry = entry
	return decision
}
