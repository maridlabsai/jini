package app

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestRunSavingsEmptyLedger(t *testing.T) {
	withSavingsHome(t)
	clearLocaleEnv(t)
	var out bytes.Buffer
	if code := runSavings(nil, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "No savings recorded yet.") {
		t.Fatalf("empty ledger copy wrong: %q", out.String())
	}
}

func TestRunSavingsTextTotalsAndDisclosure(t *testing.T) {
	withSavingsHome(t)
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	if err := appendSavingsEntry(sampleEntry(1.25)); err != nil {
		t.Fatal(err)
	}
	if err := appendSavingsEntry(sampleEntry(0.75)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := runSavings(nil, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := out.String()
	// Total == ledger sum (measurement-integrity).
	if !strings.Contains(got, "$2.00 across 2 tasks") {
		t.Fatalf("totals wrong: %q", got)
	}
	if !strings.Contains(got, "100% imputed") {
		t.Fatalf("imputed split missing: %q", got)
	}
	if !strings.Contains(got, "chars ÷ 4") || !strings.Contains(got, "prices as of "+savingsPriceAsOf) {
		t.Fatalf("basis disclosure missing: %q", got)
	}
}

func TestRunSavingsLocalizedCarriesFXAndUSD(t *testing.T) {
	withSavingsHome(t)
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_IN.UTF-8")
	t.Setenv("JINI_FX_RATE", "80")
	if err := appendSavingsEntry(sampleEntry(1.5)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	runSavings(nil, &out, io.Discard)
	got := out.String()
	if !strings.Contains(got, "₹120.00") || !strings.Contains(got, "US$1.50") {
		t.Fatalf("localized total must carry USD source: %q", got)
	}
	if !strings.Contains(got, "FX INR @ 80 as of "+savingsFXAsOf) {
		t.Fatalf("FX disclosure with value+date missing: %q", got)
	}
}

func TestRunSavingsJSONShape(t *testing.T) {
	withSavingsHome(t)
	clearLocaleEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	if err := appendSavingsEntry(sampleEntry(2.0)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := runSavings([]string{"--format", "json"}, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	var report savingsReportJSON
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if report.ContextType != "JiniSavingsReport" || report.Tasks != 1 || report.TotalUSDSaved != 2.0 {
		t.Fatalf("bad report: %+v", report)
	}
	if !report.Imputed || report.CharsPerToken != savingsCharsPerToken || report.FXAsOf != savingsFXAsOf {
		t.Fatalf("disclosure fields missing: %+v", report)
	}
}

func TestRunSavingsRejectsUnknownArg(t *testing.T) {
	withSavingsHome(t)
	var errOut bytes.Buffer
	if code := runSavings([]string{"--bogus"}, io.Discard, &errOut); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "Unknown argument") {
		t.Fatalf("wrong rejection: %q", errOut.String())
	}
}

func TestSavingsIsARoutedTopLevelCommand(t *testing.T) {
	if canonicalTopLevelCommand("savings") != "savings" {
		t.Fatal("savings not canonical top-level command")
	}
	if err := validateNativeArgs([]string{"savings", "--format", "json"}); err != nil {
		t.Fatalf("savings --format json must validate: %v", err)
	}
}
