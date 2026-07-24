package app

import (
	"strings"
	"testing"
)

func clearLocaleEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"JINI_CURRENCY", "JINI_FX_RATE", "LC_MONETARY", "LC_ALL", "LANG"} {
		t.Setenv(k, "")
	}
}

func TestRegionFromLocale(t *testing.T) {
	cases := map[string]string{
		"en_IN.UTF-8": "IN",
		"de_DE":       "DE",
		"en_GB.UTF-8": "GB",
		"C":           "",
		"":            "",
		"en":          "",
	}
	for in, want := range cases {
		if got := regionFromLocale(in); got != want {
			t.Errorf("regionFromLocale(%q)=%q want %q", in, got, want)
		}
	}
}

func TestDetectDisplayCurrencyFromLocale(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	if got := detectDisplayCurrency(); got != "INR" {
		t.Fatalf("expected INR from en_IN, got %q", got)
	}
}

func TestDetectDisplayCurrencyDefaultsUSD(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "C")
	if got := detectDisplayCurrency(); got != "USD" {
		t.Fatalf("expected USD default, got %q", got)
	}
}

func TestDetectDisplayCurrencyOverrideWins(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_CURRENCY", "gbp")
	if got := detectDisplayCurrency(); got != "GBP" {
		t.Fatalf("JINI_CURRENCY override should win, got %q", got)
	}
}

func TestLocalizeAlwaysCarriesUSD(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "80")
	amt := localize(1.5)
	if amt.Currency != "INR" {
		t.Fatalf("expected INR, got %q", amt.Currency)
	}
	if amt.USD != 1.5 {
		t.Fatalf("USD source of truth must be preserved, got %v", amt.USD)
	}
	if amt.Rate != 80 {
		t.Fatalf("JINI_FX_RATE override not applied, got %v", amt.Rate)
	}
	if !strings.HasPrefix(amt.Local, "₹") {
		t.Fatalf("expected ₹ formatting, got %q", amt.Local)
	}
}

func TestLocalizeUSDLocaleNoFallbackFlag(t *testing.T) {
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	amt := localize(2.0)
	if amt.Currency != "USD" || amt.FellBack {
		t.Fatalf("US locale is not a fallback: %+v", amt)
	}
	if amt.Local != "$2.00" {
		t.Fatalf("expected $2.00, got %q", amt.Local)
	}
}

func TestLocalizeUnknownCurrencyFallsBackToUSD(t *testing.T) {
	clearLocaleEnv(t)
	// A region whose currency has no rate entry: force via JINI_CURRENCY that
	// is not in the table so detect ignores it, then a locale with no mapping.
	t.Setenv("LANG", "xx_ZZ.UTF-8")
	amt := localize(3.0)
	if amt.Currency != "USD" {
		t.Fatalf("unmapped locale must yield USD, got %q", amt.Currency)
	}
}

func TestFormatMoneyIndianGrouping(t *testing.T) {
	if got := formatMoney("INR", 990000); got != "₹9,90,000.00" {
		t.Fatalf("lakh grouping wrong: %q", got)
	}
	if got := formatMoney("INR", 9900); got != "₹9,900.00" {
		t.Fatalf("small INR wrong: %q", got)
	}
}

func TestFormatMoneyWesternGroupingAndDecimals(t *testing.T) {
	if got := formatMoney("USD", 1234567.5); got != "$1,234,567.50" {
		t.Fatalf("western grouping wrong: %q", got)
	}
	if got := formatMoney("JPY", 1234567); got != "¥1,234,567" {
		t.Fatalf("zero-decimal currency wrong: %q", got)
	}
}
