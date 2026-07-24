package app

// Per-task savings computation — specs/savings-ledger-mvp-design.md. The ledger
// records IMPUTED savings from free/subscription routes: money the user did
// NOT spend by routing off a metered API. Metered BYO usage the user actually
// paid for is not a saving and writes no entry (baselineForRoute → not ok).

import (
	"os"
	"path/filepath"
	"time"
)

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
func recordSavingsOnDecision(decision routeDecision, provider providerConfig, inChars, outChars int, request providerGenerationRequest) routeDecision {
	if request.Standalone {
		return decision
	}
	entry := computeTaskSavings(decision, provider, inChars, outChars, request)
	if entry == nil {
		return decision
	}
	_ = appendSavingsEntry(*entry)
	decision.SavingsEntry = entry
	return decision
}
