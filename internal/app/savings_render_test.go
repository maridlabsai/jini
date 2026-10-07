package app

import (
	"strings"
	"testing"
)

func TestSavingsFooterNilOrZeroIsEmpty(t *testing.T) {
	if savingsFooterLine(nil) != "" {
		t.Fatal("nil entry must render no footer")
	}
	if savingsFooterLine(&savingsEntry{USDSaved: 0}) != "" {
		t.Fatal("zero saving must render no footer")
	}
}

func TestSavingsFooterCarriesUSDWhenLocalized(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "80")
	line := savingsFooterLine(&savingsEntry{USDSaved: 1.5})
	if !strings.Contains(line, "₹120.00") {
		t.Fatalf("expected localized figure, got %q", line)
	}
	if !strings.Contains(line, "US$1.50") {
		t.Fatalf("localized footer must carry USD source, got %q", line)
	}
	if !strings.Contains(line, "imputed") {
		t.Fatalf("footer must label imputed, got %q", line)
	}
	if !strings.Contains(line, "Estimated avoided API spend") {
		t.Fatalf("localized footer must use 'Estimated avoided API spend', got %q", line)
	}
	assertFooterNotFactualSaved(t, line)
}

func TestSavingsFooterUSDLocaleOmitsRedundantUSD(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	line := savingsFooterLine(&savingsEntry{USDSaved: 2.5})
	if !strings.Contains(line, "$2.50") || strings.Contains(line, "US$") {
		t.Fatalf("USD locale should show $2.50 without US prefix, got %q", line)
	}
	if !strings.Contains(line, "Estimated avoided API spend") {
		t.Fatalf("USD footer must use 'Estimated avoided API spend', got %q", line)
	}
	if !strings.Contains(line, "imputed") {
		t.Fatalf("footer must label imputed on the same line, got %q", line)
	}
	assertFooterNotFactualSaved(t, line)
}

// assertFooterNotFactualSaved enforces JINI-R0: a reachable savings footer must
// never present imputed dollars as factual realized savings.
func assertFooterNotFactualSaved(t *testing.T, line string) {
	t.Helper()
	low := strings.ToLower(line)
	for _, banned := range []string{"saved ≈", "saved you", "saved me", "total saved"} {
		if strings.Contains(low, banned) {
			t.Fatalf("footer must not present imputed dollars as money saved (%q): %q", banned, line)
		}
	}
}

// JINI-R0: the startup counter must label its figure estimated/imputed on the
// same line and must not present imputed dollars as money "saved".
func TestSavingsStartupCounterIsQualifiedAndNeutral(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	ledger := &savingsLedger{}
	ledger.Totals = savingsTotals{Tasks: 7, USDSaved: 3.25}
	line := savingsStartupCounterLine(ledger)
	if !strings.Contains(line, "$") {
		t.Fatalf("expected a dollar figure: %q", line)
	}
	low := strings.ToLower(line)
	if !strings.Contains(low, "imputed") && !strings.Contains(low, "estimated") {
		t.Fatalf("counter figure must be labeled estimated/imputed on the same line: %q", line)
	}
	for _, banned := range []string{"has saved you", "saved you", "saved me", "Total saved"} {
		if strings.Contains(line, banned) {
			t.Fatalf("counter must not present imputed dollars as money saved (%q): %q", banned, line)
		}
	}
}

func TestSavingsStartupCounterEmptyLedger(t *testing.T) {
	if savingsStartupCounterLine(nil) != "" {
		t.Fatal("nil ledger must render no counter")
	}
	empty := &savingsLedger{}
	if savingsStartupCounterLine(empty) != "" {
		t.Fatal("empty ledger must render no counter")
	}
}

func TestSavingsStartupCounterReportsTotals(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	ledger := &savingsLedger{}
	ledger.Totals = savingsTotals{Tasks: 12, USDSaved: 34.5}
	line := savingsStartupCounterLine(ledger)
	if !strings.Contains(line, "12 tasks") || !strings.Contains(line, "$34.50") {
		t.Fatalf("counter wrong: %q", line)
	}
	if !strings.Contains(line, "imputed") {
		t.Fatalf("counter must label imputed: %q", line)
	}
}
