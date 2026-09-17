package app

import (
	"fmt"
	"io"
)

// Savings surfaces — specs/savings-ledger-mvp-design.md. Every line is honest:
// the localized amount is always accompanied by its USD source when the
// display currency is not USD, and every figure is labeled imputed. Surfaces
// stay clean when there is nothing to show (nil entry / empty ledger).

func savingsFooterLine(entry *savingsEntry) string {
	if entry == nil || entry.USDSaved <= 0 {
		return ""
	}
	amt := localize(entry.USDSaved)
	if amt.Currency == "USD" {
		return fmt.Sprintf("Saved ≈ %s this task · imputed · jini savings", amt.Local)
	}
	return fmt.Sprintf("Saved ≈ %s (US%s) this task · imputed · jini savings",
		amt.Local, formatMoney("USD", amt.USD))
}

func renderSavingsFooter(w io.Writer, entry *savingsEntry) {
	if line := savingsFooterLine(entry); line != "" {
		fmt.Fprintln(w, line)
	}
}

// savingsStartupCounterLine is the one-line all-time running counter; empty
// when nothing has been saved yet, so startup stays clean.
func savingsStartupCounterLine(ledger *savingsLedger) string {
	if ledger == nil || ledger.Totals.Tasks == 0 || ledger.Totals.USDSaved <= 0 {
		return ""
	}
	amt := localize(ledger.Totals.USDSaved)
	if amt.Currency == "USD" {
		return fmt.Sprintf("Jini has saved you ≈ %s across %d tasks (imputed).",
			amt.Local, ledger.Totals.Tasks)
	}
	return fmt.Sprintf("Jini has saved you ≈ %s (US%s) across %d tasks (imputed).",
		amt.Local, formatMoney("USD", amt.USD), ledger.Totals.Tasks)
}

func renderSavingsStartupCounter(w io.Writer) {
	if line := savingsStartupCounterLine(loadSavingsLedger()); line != "" {
		fmt.Fprintln(w, line)
	}
}
