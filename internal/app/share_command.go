package app

import (
	"fmt"
	"io"
)

// jini share — DISABLED during the trust-first recovery (JINI-R0). Its historical
// receipt figures are estimated, not verified proof, so the command no longer emits
// them. The receipt helpers below are RETAINED (not deleted) for the JINI-R8 honest
// ledger and JINI-R10 proof-receipt work, which will replace this surface. Direct
// invocation stays recognized so it fails closed rather than falling through to
// prompt handling.

const shareRepoURL = "github.com/maridlabsai/jini"

func runShare(args []string, stdout, stderr io.Writer) int {
	_ = args
	fmt.Fprintln(stderr, "jini share is temporarily unavailable.")
	fmt.Fprintln(stderr, "Past metrics are estimates, not verified proof, so Jini does not publish them yet.")
	return 1
}

func shareLedgerEmpty(ledger *savingsLedger) bool {
	return ledger == nil || ledger.Totals.Tasks == 0 || ledger.Totals.USDSaved <= 0
}

// shareableSavingsCard renders the copyable receipt in "text" or "markdown".
func shareableSavingsCard(ledger *savingsLedger, format string) string {
	amt := localize(ledger.Totals.USDSaved)
	amount := amt.Local + usdSuffix(amt)
	tasks := ledger.Totals.Tasks
	dodges := ledger.Totals.Dodges

	headline := fmt.Sprintf("Jini saved me %s across %d AI-coding tasks", amount, tasks)
	if dodges > 0 {
		word := "throttles"
		if dodges == 1 {
			word = "throttle"
		}
		headline += fmt.Sprintf(" — survived %d %s with 0 walls hit", dodges, word)
	}
	pitch := "The un-metered AI coding CLI: routes across free providers, survives rate limits, shows you what it saved."

	if format == "markdown" {
		return fmt.Sprintf("🧞 **%s.**\n\n%s\n\n→ [%s](https://%s)", headline, pitch, shareRepoURL, shareRepoURL)
	}
	return fmt.Sprintf("🧞 %s.\n%s\n→ https://%s", headline, pitch, shareRepoURL)
}
