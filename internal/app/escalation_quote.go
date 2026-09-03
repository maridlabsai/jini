package app

import "fmt"

// escalationCostQuoteLine quotes the next rung's cost before the user's money
// is spent on it — PRD requirement: escalation is never silently billed.
//
// Only metered tiers get a quote. External CLI handoffs, local models, and the
// free preview cost the user nothing at the point of use, so quoting them would
// be noise. An unknown or unpriceable route returns "" rather than a guess:
// fail honest, not fail generous.
func escalationCostQuoteLine(decision routeDecision, systemPrompt, userPrompt string) string {
	descriptor, ok := adapterDescriptorForMode(decision.ToolMode)
	if !ok {
		return ""
	}
	var baseline string
	switch descriptor.CostTier {
	case "premium":
		baseline = "claude-frontier"
	case "standard":
		baseline = "gpt-frontier"
	default:
		return ""
	}
	price, ok := savingsModelPrices[baseline]
	if !ok {
		return ""
	}
	inTokens := estTokensFromChars(len(systemPrompt) + len(userPrompt))
	const outTokens = 1600 // disclosed flat estimate; a quote, not a meter
	est := savingsUSD(baseline, inTokens, outTokens)
	return fmt.Sprintf(
		"Cost note: %s is a paid route (~$%.2f/1M in, $%.2f/1M out) — est ~$%.4f for this task before it runs. Local and free routes cost $0.",
		firstNonEmpty(decision.ToolLabel, decision.ToolMode),
		price.InUSDPer1M,
		price.OutUSDPer1M,
		est,
	)
}
