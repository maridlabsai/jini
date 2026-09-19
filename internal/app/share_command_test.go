package app

import (
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
