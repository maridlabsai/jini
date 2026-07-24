package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func withSavingsHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { executionModeHomeDir = old })
}

func sampleEntry(usd float64) savingsEntry {
	return savingsEntry{
		SchemaVersion:   "0.1.0",
		ContextType:     "JiniSavingsEntry",
		At:              "2026-07-24T10:00:00-07:00",
		Repo:            "jini",
		RouteLabel:      "Claude Code CLI handoff",
		RouteClass:      savingsRouteSubscription,
		EstInputTokens:  100,
		EstOutputTokens: 200,
		USDSaved:        usd,
		Imputed:         true,
	}
}

func TestSavingsLedgerRoundTrip(t *testing.T) {
	withSavingsHome(t)
	if err := appendSavingsEntry(sampleEntry(0.5)); err != nil {
		t.Fatal(err)
	}
	if err := appendSavingsEntry(sampleEntry(1.5)); err != nil {
		t.Fatal(err)
	}
	ledger := loadSavingsLedger()
	if ledger == nil {
		t.Fatal("ledger not loaded")
	}
	if ledger.Totals.Tasks != 2 {
		t.Fatalf("tasks=%d want 2", ledger.Totals.Tasks)
	}
	if ledger.Totals.USDSaved != 2.0 {
		t.Fatalf("usd=%v want 2.0", ledger.Totals.USDSaved)
	}
	path, _ := savingsLedgerPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %o", perm)
	}
}

func TestSavingsLedgerCorruptLoadsNil(t *testing.T) {
	withSavingsHome(t)
	path, _ := savingsLedgerPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if loadSavingsLedger() != nil {
		t.Fatal("corrupt ledger must load nil")
	}
}

func TestSavingsLedgerFoldingPreservesInvariant(t *testing.T) {
	withSavingsHome(t)
	total := 0.0
	for i := 0; i < savingsLedgerMaxEntries+50; i++ {
		if err := appendSavingsEntry(sampleEntry(0.1)); err != nil {
			t.Fatal(err)
		}
		total += 0.1
	}
	ledger := loadSavingsLedger()
	if ledger == nil {
		t.Fatal("ledger nil after many appends")
	}
	if len(ledger.Entries) > savingsLedgerMaxEntries {
		t.Fatalf("entries not bounded: %d", len(ledger.Entries))
	}
	if ledger.Totals.Tasks != savingsLedgerMaxEntries+50 {
		t.Fatalf("grand total task count wrong: %d", ledger.Totals.Tasks)
	}
	// Invariant: Totals == Folded + sum(Entries).
	sum := ledger.Folded.USDSaved
	for _, e := range ledger.Entries {
		sum += e.USDSaved
	}
	if diff := sum - ledger.Totals.USDSaved; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("invariant broken: folded+entries=%v totals=%v", sum, ledger.Totals.USDSaved)
	}
}

func TestSavingsLedgerTamperedTotalsLoadsNil(t *testing.T) {
	withSavingsHome(t)
	if err := appendSavingsEntry(sampleEntry(1.0)); err != nil {
		t.Fatal(err)
	}
	path, _ := savingsLedgerPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var ledger savingsLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	ledger.Totals.USDSaved = 999.0 // tamper
	tampered, _ := json.Marshal(ledger)
	if err := os.WriteFile(path, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if loadSavingsLedger() != nil {
		t.Fatal("tampered totals must fail the integrity invariant → nil")
	}
}
