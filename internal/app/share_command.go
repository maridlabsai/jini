package app

import (
	"fmt"
	"io"
	"strings"
)

// jini share — #2 (viral loops) of specs/pre-viral-readiness.md. Renders a clean,
// copyable "receipt" from the savings ledger so a happy user can paste their result
// to X / a PR / Slack. Each share carries the wedge — un-metered, no throttle walls —
// posted by a real user: the Loom-style loop where the artifact is the ad. It emits
// aggregate totals ONLY (never a prompt, code, or route detail), so sharing is safe.

const shareRepoURL = "github.com/maridlabsai/jini"

func runShare(args []string, stdout, stderr io.Writer) int {
	format := "text"
	for _, a := range args {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "--markdown", "--md", "--format=markdown":
			format = "markdown"
		case "--text", "--format=text":
			format = "text"
		}
	}
	ledger := loadSavingsLedger()
	if shareLedgerEmpty(ledger) {
		fmt.Fprintln(stdout, "Nothing to share yet — run a few tasks through Jini, then `jini share`.")
		return 0
	}
	fmt.Fprintln(stdout, shareableSavingsCard(ledger, format))
	return 0
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
