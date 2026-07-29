package app

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// `jini savings` — specs/savings-ledger-mvp-design.md. Renders the ledger as a
// text summary (default) or JSON. Every figure is labeled imputed, the
// localized total always carries its USD source, and the basis line discloses
// the estimation divisor, price date, and FX rate + date so the number is
// auditable. Empty ledger prints a clean line, never an error.

func runSavings(args []string, stdout, stderr io.Writer) int {
	format := "text"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--format" && i+1 < len(args):
			format = normalizeName(args[i+1])
			i++
		case strings.HasPrefix(arg, "--format="):
			format = normalizeName(strings.TrimPrefix(arg, "--format="))
		default:
			fmt.Fprintf(stderr, "Unknown argument %q. Use `jini savings` or `jini savings --format json`.\n", arg)
			return 1
		}
	}
	if format != "text" && format != "json" {
		fmt.Fprintln(stderr, "Unsupported savings format. Try `jini savings` or `jini savings --format json`.")
		return 1
	}
	ledger := loadSavingsLedger()
	if format == "json" {
		return renderSavingsJSON(stdout, ledger)
	}
	renderSavingsText(stdout, ledger)
	return 0
}

func usdSuffix(amt localizedAmount) string {
	if amt.Currency == "USD" {
		return ""
	}
	return " (US" + formatMoney("USD", amt.USD) + ")"
}

func fxDisclosure(amt localizedAmount) string {
	if amt.Currency == "USD" {
		return "figures in USD"
	}
	return fmt.Sprintf("FX %s @ %g as of %s", amt.Currency, amt.Rate, savingsFXAsOf)
}

func savingsIsEmpty(ledger *savingsLedger) bool {
	return ledger == nil || ledger.Totals.Tasks == 0 || ledger.Totals.USDSaved <= 0
}

func renderSavingsText(w io.Writer, ledger *savingsLedger) {
	if savingsIsEmpty(ledger) {
		fmt.Fprintln(w, "No savings recorded yet.")
		fmt.Fprintln(w, "Jini logs imputed savings when it routes work off metered APIs.")
		return
	}
	total := localize(ledger.Totals.USDSaved)
	fmt.Fprintln(w, "Jini savings — imputed")
	fmt.Fprintf(w, "Total saved ≈ %s%s across %d tasks.\n", total.Local, usdSuffix(total), ledger.Totals.Tasks)
	fmt.Fprintf(w, "Throttles dodged: %d\n", ledger.Totals.Dodges)

	byClass := map[string]*savingsTotals{}
	order := []string{}
	for _, e := range ledger.Entries {
		t, ok := byClass[e.RouteClass]
		if !ok {
			t = &savingsTotals{}
			byClass[e.RouteClass] = t
			order = append(order, e.RouteClass)
		}
		t.add(e)
	}
	if len(order) > 0 {
		sort.Strings(order)
		fmt.Fprintln(w, "By route:")
		for _, class := range order {
			t := byClass[class]
			amt := localize(t.USDSaved)
			fmt.Fprintf(w, "  %-13s ≈ %s%s — %d tasks\n", class, amt.Local, usdSuffix(amt), t.Tasks)
		}
	}
	if ledger.Folded.Tasks > 0 {
		fmt.Fprintf(w, "(+ %d earlier tasks folded into the total; route detail not retained)\n", ledger.Folded.Tasks)
	}
	fmt.Fprintln(w, "Literal vs imputed: 100% imputed (no metered spend recorded).")
	fmt.Fprintf(w, "Basis: est. tokens = chars ÷ %d; API-equivalent prices as of %s; %s.\n",
		savingsCharsPerToken, savingsPriceAsOf, fxDisclosure(total))
}

type savingsReportJSON struct {
	ContextType   string  `json:"context_type"`
	Currency      string  `json:"currency"`
	TotalUSDSaved float64 `json:"total_usd_saved"`
	TotalLocal    string  `json:"total_local"`
	FXRate        float64 `json:"fx_rate"`
	FXAsOf        string  `json:"fx_as_of"`
	Tasks         int     `json:"tasks"`
	Dodges        int     `json:"dodges"`
	Imputed       bool    `json:"imputed"`
	CharsPerToken int     `json:"chars_per_token"`
	PriceAsOf     string  `json:"price_as_of"`
}

func renderSavingsJSON(w io.Writer, ledger *savingsLedger) int {
	report := savingsReportJSON{
		ContextType:   "JiniSavingsReport",
		Currency:      "USD",
		Imputed:       true,
		CharsPerToken: savingsCharsPerToken,
		PriceAsOf:     savingsPriceAsOf,
		FXAsOf:        savingsFXAsOf,
		FXRate:        1.0,
	}
	if !savingsIsEmpty(ledger) {
		total := localize(ledger.Totals.USDSaved)
		report.Currency = total.Currency
		report.TotalUSDSaved = total.USD
		report.TotalLocal = total.Local
		report.FXRate = total.Rate
		report.Tasks = ledger.Totals.Tasks
		report.Dodges = ledger.Totals.Dodges
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return 1
	}
	fmt.Fprintln(w, string(data))
	return 0
}
