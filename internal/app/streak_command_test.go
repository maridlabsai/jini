package app

import (
	"strings"
	"testing"
)

func TestStreakCardShowsMomentumAndNextMilestone(t *testing.T) {
	ledger := &savingsLedger{Totals: savingsTotals{Tasks: 12, USDSaved: 8.40, Dodges: 2}}
	card := streakCard(ledger, "text")
	for _, want := range []string{"12 tasks", "saved", "2 throttles dodged", "next milestone: 25 tasks", "13 to go"} {
		if !strings.Contains(card, want) {
			t.Fatalf("streak card missing %q:\n%s", want, card)
		}
	}
	// Not on a milestone → no copyable share block.
	if strings.Contains(card, "Milestone reached") {
		t.Fatalf("mid-progress card must not claim a milestone:\n%s", card)
	}
}

func TestStreakCardMilestoneEmitsShare(t *testing.T) {
	ledger := &savingsLedger{Totals: savingsTotals{Tasks: 50, USDSaved: 44.10, Dodges: 4}}
	card := streakCard(ledger, "text")
	if !strings.Contains(card, "Milestone reached") {
		t.Fatalf("landing exactly on 50 must celebrate:\n%s", card)
	}
	if !strings.Contains(card, "Jini just hit 50 AI-coding tasks") || !strings.Contains(card, shareRepoURL) {
		t.Fatalf("milestone share line missing headline/link:\n%s", card)
	}
	// Next milestone after 50 is 100, 50 to go.
	if !strings.Contains(card, "next milestone: 100 tasks (50 to go)") {
		t.Fatalf("milestone card must still point at the next tier:\n%s", card)
	}
}

func TestStreakMilestoneShareMarkdownHasLink(t *testing.T) {
	md := streakMilestoneShare(100, "$88.00", 0, "markdown")
	if !strings.Contains(md, "](https://"+shareRepoURL+")") {
		t.Fatalf("markdown share must carry a link:\n%s", md)
	}
	if strings.Contains(md, "throttle walls") {
		t.Fatalf("zero dodges must omit the walls clause:\n%s", md)
	}
}

func TestStreakNextAndPrevMilestoneBoundaries(t *testing.T) {
	cases := []struct {
		tasks, next, prev int
	}{
		{0, 5, 0},
		{5, 10, 5},
		{7, 10, 5},
		{100, 250, 100},
		{10000, 20000, 10000},
		{15000, 20000, 10000},
	}
	for _, c := range cases {
		if got := nextTaskMilestone(c.tasks); got != c.next {
			t.Fatalf("nextTaskMilestone(%d) = %d, want %d", c.tasks, got, c.next)
		}
		if got := prevTaskMilestone(c.tasks); got != c.prev {
			t.Fatalf("prevTaskMilestone(%d) = %d, want %d", c.tasks, got, c.prev)
		}
	}
}

func TestStreakMilestoneReached(t *testing.T) {
	for _, m := range []int{5, 50, 10000, 20000} {
		if !taskMilestoneReached(m) {
			t.Fatalf("%d should be a milestone", m)
		}
	}
	for _, n := range []int{0, 3, 12, 49, 15000} {
		if taskMilestoneReached(n) {
			t.Fatalf("%d should not be a milestone", n)
		}
	}
}

func TestStreakSingularThrottleAndZeroDodges(t *testing.T) {
	one := streakCard(&savingsLedger{Totals: savingsTotals{Tasks: 12, USDSaved: 1, Dodges: 1}}, "text")
	if !strings.Contains(one, "1 throttle dodged") || strings.Contains(one, "throttles") {
		t.Fatalf("single dodge must read singular:\n%s", one)
	}
	zero := streakCard(&savingsLedger{Totals: savingsTotals{Tasks: 12, USDSaved: 1, Dodges: 0}}, "text")
	if strings.Contains(zero, "dodged") {
		t.Fatalf("zero dodges must omit the dodge clause:\n%s", zero)
	}
}

func TestStreakConsecutiveDayStreak(t *testing.T) {
	entries := []savingsEntry{
		{At: "2026-09-17T09:00:00-07:00"},
		{At: "2026-09-18T09:00:00-07:00"},
		{At: "2026-09-18T18:00:00-07:00"}, // same day, still counts once
		{At: "2026-09-19T09:00:00-07:00"},
	}
	if got := consecutiveDayStreak(entries); got != 3 {
		t.Fatalf("three consecutive days = 3, got %d", got)
	}
	// A gap breaks the run measured back from the most recent day.
	gapped := []savingsEntry{
		{At: "2026-09-10T09:00:00-07:00"},
		{At: "2026-09-19T09:00:00-07:00"},
	}
	if got := consecutiveDayStreak(gapped); got != 1 {
		t.Fatalf("gap should reset to 1, got %d", got)
	}
	if got := consecutiveDayStreak(nil); got != 0 {
		t.Fatalf("no entries = 0, got %d", got)
	}
}

func TestStreakLedgerEmptyGate(t *testing.T) {
	if !streakLedgerEmpty(nil) {
		t.Fatalf("nil ledger is empty")
	}
	if !streakLedgerEmpty(&savingsLedger{Totals: savingsTotals{Tasks: 0}}) {
		t.Fatalf("zero-task ledger is empty")
	}
	if streakLedgerEmpty(&savingsLedger{Totals: savingsTotals{Tasks: 1}}) {
		t.Fatalf("a single-task ledger has momentum, even at $0")
	}
}
