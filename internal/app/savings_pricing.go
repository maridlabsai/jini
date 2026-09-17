package app

import "strings"

// Auditable API-equivalent price table — specs/savings-ledger-mvp-design.md.
// Values are USD list prices per 1,000,000 tokens for the baseline metered
// model each free/subscription route is counterfactually compared against.
// Point-in-time and disclosed: savingsPriceAsOf travels with every figure on
// `jini savings`. Update a price and savingsPriceAsOf together, citing source.

const (
	savingsPriceAsOf     = "2026-07"
	savingsCharsPerToken = 4 // conservative estimate; disclosed, never a multiplier
)

// Route classes for savings attribution.
const (
	savingsRouteSubscription = "subscription" // flat-fee CLI handoff; imputed savings
	savingsRouteLocal        = "local"        // local model; imputed savings vs frontier
	savingsRouteMetered      = "metered"      // BYO API; user paid — $0 saved in MVP
)

type modelPrice struct {
	InUSDPer1M  float64
	OutUSDPer1M float64
}

// Posted frontier-API list prices as of savingsPriceAsOf.
var savingsModelPrices = map[string]modelPrice{
	"gpt-frontier":    {InUSDPer1M: 2.50, OutUSDPer1M: 10.00},
	"claude-frontier": {InUSDPer1M: 3.00, OutUSDPer1M: 15.00},
}

// baselineForRoute picks the counterfactual metered model for a route class.
// metered/unknown routes return ok=false: no saving is imputed for money the
// user actually spent, and an unpriceable route records nothing rather than a
// guess (fail honest, not fail generous).
func baselineForRoute(routeClass, routeLabel string) (string, bool) {
	switch routeClass {
	case savingsRouteSubscription:
		if strings.Contains(strings.ToLower(routeLabel), "claude") {
			return "claude-frontier", true
		}
		return "gpt-frontier", true
	case savingsRouteLocal:
		// Local runs cost $0; the counterfactual is a frontier metered API.
		return "claude-frontier", true
	default:
		return "", false
	}
}

// savingsUSD imputes the counterfactual metered cost of a task from estimated
// token counts. Returns 0 when the baseline is unknown.
func savingsUSD(baseline string, inTokens, outTokens int) float64 {
	price, ok := savingsModelPrices[baseline]
	if !ok {
		return 0
	}
	return float64(inTokens)*price.InUSDPer1M/1e6 + float64(outTokens)*price.OutUSDPer1M/1e6
}

// estTokensFromChars converts a char count to an estimated token count.
func estTokensFromChars(chars int) int {
	if chars <= 0 {
		return 0
	}
	return chars / savingsCharsPerToken
}
