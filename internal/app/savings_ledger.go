package app

// Persistent savings ledger — specs/savings-ledger-mvp-design.md. GLOBAL /
// cross-repo (same ~/.jini home as mode.json): savings is a property of the
// user, not one workspace. Bounded to the last N entries verbatim; older
// entries fold into Folded so the file stays small while the grand total stays
// verifiable via the invariant Totals == Folded + sum(Entries). A ledger that
// fails the invariant, or won't parse, loads as nil and never blocks a task.

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
)

const (
	savingsLedgerSchema     = "0.1.0"
	savingsLedgerContext    = "JiniSavingsLedger"
	savingsEntryContext     = "JiniSavingsEntry"
	savingsLedgerMaxEntries = 500
)

type savingsEntry struct {
	SchemaVersion   string  `json:"schema_version"`
	ContextType     string  `json:"context_type"`
	At              string  `json:"at"`
	Repo            string  `json:"repo"`
	RouteLabel      string  `json:"route_label"`
	RouteClass      string  `json:"route_class"`
	EstInputTokens  int     `json:"est_input_tokens"`
	EstOutputTokens int     `json:"est_output_tokens"`
	USDSaved        float64 `json:"usd_saved"`
	Imputed         bool    `json:"imputed"`
	ThrottleDodged  bool    `json:"throttle_dodged"`
	Title           string  `json:"title"`
}

type savingsTotals struct {
	Tasks    int     `json:"tasks"`
	USDSaved float64 `json:"usd_saved"`
	Dodges   int     `json:"dodges"`
}

func (t *savingsTotals) add(e savingsEntry) {
	t.Tasks++
	t.USDSaved += e.USDSaved
	if e.ThrottleDodged {
		t.Dodges++
	}
}

type savingsLedger struct {
	SchemaVersion string         `json:"schema_version"`
	ContextType   string         `json:"context_type"`
	Entries       []savingsEntry `json:"entries"`
	Folded        savingsTotals  `json:"folded"`
	Totals        savingsTotals  `json:"totals"`
}

func savingsLedgerPath() (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jini", "savings-ledger.json"), nil
}

// appendSavingsEntry records one entry, folding the oldest when bounded, and
// recomputes Totals = Folded + sum(Entries).
func appendSavingsEntry(entry savingsEntry) error {
	path, err := savingsLedgerPath()
	if err != nil {
		return err
	}
	ledger := loadSavingsLedger()
	if ledger == nil {
		ledger = &savingsLedger{SchemaVersion: savingsLedgerSchema, ContextType: savingsLedgerContext}
	}
	ledger.Entries = append(ledger.Entries, entry)
	for len(ledger.Entries) > savingsLedgerMaxEntries {
		ledger.Folded.add(ledger.Entries[0])
		ledger.Entries = ledger.Entries[1:]
	}
	ledger.Totals = ledger.Folded
	for _, e := range ledger.Entries {
		ledger.Totals.add(e)
	}
	return persistSavingsLedger(path, *ledger)
}

func persistSavingsLedger(path string, ledger savingsLedger) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "savings-ledger-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// loadSavingsLedger returns nil on absence, corruption, wrong context, or a
// broken integrity invariant — a ledger is best-effort and never blocks work.
func loadSavingsLedger() *savingsLedger {
	path, err := savingsLedgerPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var ledger savingsLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return nil
	}
	if ledger.ContextType != savingsLedgerContext {
		return nil
	}
	if !savingsInvariantHolds(ledger) {
		return nil
	}
	return &ledger
}

// savingsInvariantHolds verifies Totals == Folded + sum(Entries) on every axis.
func savingsInvariantHolds(ledger savingsLedger) bool {
	want := ledger.Folded
	for _, e := range ledger.Entries {
		want.add(e)
	}
	return want.Tasks == ledger.Totals.Tasks &&
		want.Dodges == ledger.Totals.Dodges &&
		math.Abs(want.USDSaved-ledger.Totals.USDSaved) <= 1e-9
}
