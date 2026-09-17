package app

import (
	"strings"
	"testing"
)

// Savings honesty gate — specs/number-one-platform-prd.md §Savings Ledger makes
// an inflated or mislabeled savings claim a release-blocking defect. These
// tests encode that contract; a failure here blocks the commit gate.

func TestHonesty_EveryRecordedEntryIsImputed(t *testing.T) {
	// MVP records no real metered spend, so no entry may claim to be literal.
	entry := computeTaskSavings(subscriptionDecision(), providerConfig{}, 8000, 8000, providerGenerationRequest{Title: "t"})
	if entry == nil || !entry.Imputed {
		t.Fatalf("recorded entries must be labeled imputed, got %+v", entry)
	}
}

func TestHonesty_MeteredSpendIsNotASaving(t *testing.T) {
	if e := computeTaskSavings(routeDecision{ToolLabel: "OpenAI API"}, providerConfig{ID: "openai"}, 8000, 8000, providerGenerationRequest{}); e != nil {
		t.Fatalf("money the user paid on a metered route must never be logged as a saving: %+v", e)
	}
}

func TestHonesty_LocalizedFigureAlwaysCarriesUSDSource(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "83")
	// Footer.
	if line := savingsFooterLine(&savingsEntry{USDSaved: 0.5}); !strings.Contains(line, "US$") {
		t.Fatalf("localized footer without USD source: %q", line)
	}
	// Startup counter.
	ledger := &savingsLedger{Totals: savingsTotals{Tasks: 3, USDSaved: 0.5}}
	if line := savingsStartupCounterLine(ledger); !strings.Contains(line, "US$") {
		t.Fatalf("localized counter without USD source: %q", line)
	}
}

func TestHonesty_DisclosureBasisIsAlwaysStated(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "83")
	withSavingsHome(t)
	if err := appendSavingsEntry(sampleEntry(1.0)); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	renderSavingsText(&b, loadSavingsLedger())
	out := b.String()
	for _, want := range []string{"imputed", "chars ÷ 4", "prices as of " + savingsPriceAsOf, "FX INR @ 83 as of " + savingsFXAsOf} {
		if !strings.Contains(out, want) {
			t.Fatalf("basis disclosure missing %q in:\n%s", want, out)
		}
	}
}

func TestHonesty_SubUnitAmountsNeverContradictThemselves(t *testing.T) {
	// A tiny saving must not show a non-zero localized figure beside "$0.00".
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "83")
	line := savingsFooterLine(&savingsEntry{USDSaved: 0.000457})
	if strings.Contains(line, "US$0.00 ") || strings.Contains(line, "US$0.00)") {
		t.Fatalf("localized non-zero shown beside US$0.00: %q", line)
	}
}
