package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestShareCardTextCarriesTheWedge(t *testing.T) {
	ledger := &savingsLedger{Totals: savingsTotals{Tasks: 47, USDSaved: 41.90, Dodges: 3}}
	card := shareableSavingsCard(ledger, "text")
	for _, want := range []string{"Jini saved me", "47", "survived 3 throttles", "0 walls hit", shareRepoURL} {
		if !strings.Contains(card, want) {
			t.Fatalf("share card missing %q:\n%s", want, card)
		}
	}
}

func TestShareCardMarkdownHasLink(t *testing.T) {
	ledger := &savingsLedger{Totals: savingsTotals{Tasks: 5, USDSaved: 2.5}}
	card := shareableSavingsCard(ledger, "markdown")
	if !strings.Contains(card, "](https://"+shareRepoURL+")") {
		t.Fatalf("markdown card must carry a link:\n%s", card)
	}
}

func TestShareCardSingularAndZeroDodges(t *testing.T) {
	one := shareableSavingsCard(&savingsLedger{Totals: savingsTotals{Tasks: 2, USDSaved: 1, Dodges: 1}}, "text")
	if !strings.Contains(one, "survived 1 throttle with") || strings.Contains(one, "throttles") {
		t.Fatalf("one dodge must read singular:\n%s", one)
	}
	zero := shareableSavingsCard(&savingsLedger{Totals: savingsTotals{Tasks: 2, USDSaved: 1, Dodges: 0}}, "text")
	if strings.Contains(zero, "throttle") {
		t.Fatalf("zero dodges → omit the survival line:\n%s", zero)
	}
}

func TestShareLedgerEmptyGate(t *testing.T) {
	if !shareLedgerEmpty(nil) {
		t.Fatalf("nil ledger is empty")
	}
	if !shareLedgerEmpty(&savingsLedger{Totals: savingsTotals{Tasks: 0}}) {
		t.Fatalf("zero-task ledger is empty")
	}
	if shareLedgerEmpty(&savingsLedger{Totals: savingsTotals{Tasks: 3, USDSaved: 1.0}}) {
		t.Fatalf("a real ledger is not empty")
	}
}

// JINI-R0: jini share is disabled. Dispatch-level test — it must be recognized
// (exit 1, fail closed), never fall through to prompt handling, and emit no
// promotional/savings claim.
func TestShareDispatchDisabledFailsClosed(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"share"}, &out, &errb); code != 1 {
		t.Fatalf("jini share must fail closed with exit 1, got %d\nstdout=%q stderr=%q", code, out.String(), errb.String())
	}
	combined := strings.ToLower(out.String() + errb.String())
	for _, banned := range []string{"saved", "un-metered", "unmetered", "0 walls", "walls avoided", "$"} {
		if strings.Contains(combined, strings.ToLower(banned)) {
			t.Fatalf("unavailable share output must not contain %q; got:\n%s", banned, combined)
		}
	}
	// No internal roadmap identifiers or recovery-program terminology in user copy.
	for _, internal := range []string{"jini-r8", "jini-r10", "trust-first", "honest ledger", "recovery"} {
		if strings.Contains(combined, internal) {
			t.Fatalf("share copy must not leak internal term %q; got:\n%s", internal, combined)
		}
	}
	if !strings.Contains(combined, "temporarily unavailable") {
		t.Fatalf("share must announce it is temporarily unavailable; got:\n%s", combined)
	}
	if !strings.Contains(combined, "estimates, not verified proof") {
		t.Fatalf("share must state metrics are estimates, not verified proof; got:\n%s", combined)
	}
}

// Regression guard (not an R0 behavior change): share and streak were never
// advertised in `jini commands` discovery, and must stay absent. This asserts
// they do not get surfaced later, not that R0 removed them.
func TestShareAndStreakNotInCommandDiscovery(t *testing.T) {
	var out, errb bytes.Buffer
	Run([]string{"commands"}, &out, &errb)
	help := out.String() + errb.String()
	for _, hidden := range []string{"jini share", "jini streak"} {
		if strings.Contains(help, hidden) {
			t.Fatalf("command discovery must not advertise %q; got:\n%s", hidden, help)
		}
	}
}
