package app

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// jini streak — #2 (viral loops) of specs/pre-viral-readiness.md, the Duolingo-style
// momentum surface. It reads the same savings ledger as `jini savings`/`jini share`
// and turns it into a compact "keep going" signal: current momentum (tasks + $ saved),
// progress toward the next round-number milestone, and — when a milestone lands — a
// copyable one-line share (the artifact-is-the-ad loop). Aggregate totals ONLY; never
// a prompt, code, or route detail, so it is safe to paste. Empty ledger prints a
// friendly nudge, never an error.

// streakTaskMilestones are the round-number task counts that mark momentum. Money
// alone is noisy (per-task savings vary); task-count milestones are the stable,
// gamifiable spine, mirroring how Duolingo counts days, not minutes.
var streakTaskMilestones = []int{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

func runStreak(args []string, stdout, stderr io.Writer) int {
	format := "text"
	for _, a := range args {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "--markdown", "--md", "--format=markdown":
			format = "markdown"
		case "--text", "--format=text":
			format = "text"
		}
	}
	ledger := loadSavingsLedger()
	if streakLedgerEmpty(ledger) {
		fmt.Fprintln(stdout, "No streak yet — run a few tasks through Jini, then check `jini streak`.")
		return 0
	}
	fmt.Fprintln(stdout, streakCard(ledger, format))
	return 0
}

// streakLedgerEmpty gates on task count only: a user who has routed tasks has
// momentum worth showing even if the imputed dollars round to zero.
func streakLedgerEmpty(ledger *savingsLedger) bool {
	return ledger == nil || ledger.Totals.Tasks <= 0
}

// nextTaskMilestone returns the smallest milestone strictly greater than tasks.
// Beyond the table it rolls to the next multiple of the largest tier so momentum
// never runs out of a "next" to chase.
func nextTaskMilestone(tasks int) int {
	for _, m := range streakTaskMilestones {
		if m > tasks {
			return m
		}
	}
	last := streakTaskMilestones[len(streakTaskMilestones)-1]
	return ((tasks / last) + 1) * last
}

// prevTaskMilestone returns the largest milestone at or below tasks (0 if none),
// used as the baseline for the progress bar toward the next milestone.
func prevTaskMilestone(tasks int) int {
	prev := 0
	for _, m := range streakTaskMilestones {
		if m <= tasks {
			prev = m
		}
	}
	last := streakTaskMilestones[len(streakTaskMilestones)-1]
	if tasks >= last {
		prev = (tasks / last) * last
	}
	return prev
}

// taskMilestoneReached reports whether tasks lands exactly on a milestone — the
// moment worth celebrating and offering a copyable share.
func taskMilestoneReached(tasks int) bool {
	if tasks <= 0 {
		return false
	}
	for _, m := range streakTaskMilestones {
		if tasks == m {
			return true
		}
	}
	last := streakTaskMilestones[len(streakTaskMilestones)-1]
	return tasks > last && tasks%last == 0
}

// consecutiveDayStreak counts consecutive calendar days ending on the most recent
// recorded entry. savingsEntry carries an RFC3339 timestamp (At), so a real
// day-streak is computable; folded/older entries drop their timestamps, so this is
// best-effort over the retained entries and returns 0 when none parse.
func consecutiveDayStreak(entries []savingsEntry) int {
	days := map[string]bool{}
	for _, e := range entries {
		t, err := time.Parse(time.RFC3339, e.At)
		if err != nil {
			continue
		}
		days[t.Format("2006-01-02")] = true
	}
	if len(days) == 0 {
		return 0
	}
	var maxDay time.Time
	first := true
	for d := range days {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		if first || t.After(maxDay) {
			maxDay, first = t, false
		}
	}
	streak := 0
	for days[maxDay.AddDate(0, 0, -streak).Format("2006-01-02")] {
		streak++
	}
	return streak
}

func streakProgressBar(tasks, prev, next int) string {
	const cells = 10
	span := next - prev
	filled := 0
	if span > 0 {
		filled = (tasks - prev) * cells / span
	}
	if filled < 0 {
		filled = 0
	}
	if filled > cells {
		filled = cells
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", cells-filled)
}

// streakCard renders the momentum block (+ a copyable share when a milestone lands)
// in "text" or "markdown". Money rendering reuses localize()/usdSuffix() so the
// figure matches `jini savings` exactly, USD source and all.
func streakCard(ledger *savingsLedger, format string) string {
	tasks := ledger.Totals.Tasks
	total := localize(ledger.Totals.USDSaved)
	amount := total.Local + usdSuffix(total)
	next := nextTaskMilestone(tasks)
	prev := prevTaskMilestone(tasks)
	toGo := next - tasks

	var b strings.Builder
	fmt.Fprintln(&b, "🧞 Jini streak — momentum")

	momentum := fmt.Sprintf("%d tasks · ≈ %s saved", tasks, amount)
	if ledger.Totals.Dodges > 0 {
		word := "throttles"
		if ledger.Totals.Dodges == 1 {
			word = "throttle"
		}
		momentum += fmt.Sprintf(" · %d %s dodged", ledger.Totals.Dodges, word)
	}
	fmt.Fprintln(&b, momentum)

	if days := consecutiveDayStreak(ledger.Entries); days > 0 {
		unit := "days"
		if days == 1 {
			unit = "day"
		}
		fmt.Fprintf(&b, "🔥 %d-%s active streak — don't break the chain.\n", days, unit)
	}

	fmt.Fprintf(&b, "%s  next milestone: %d tasks (%d to go)\n", streakProgressBar(tasks, prev, next), next, toGo)

	if taskMilestoneReached(tasks) {
		fmt.Fprintln(&b, "🎉 Milestone reached! Copyable:")
		fmt.Fprintln(&b, streakMilestoneShare(tasks, amount, ledger.Totals.Dodges, format))
	}

	return strings.TrimRight(b.String(), "\n")
}

// streakMilestoneShare is the paste-ready one-liner emitted when a milestone lands.
func streakMilestoneShare(tasks int, amount string, dodges int, format string) string {
	headline := fmt.Sprintf("Jini just hit %d AI-coding tasks — %s saved", tasks, amount)
	if dodges > 0 {
		headline += ", 0 throttle walls hit"
	}
	if format == "markdown" {
		return fmt.Sprintf("🧞 **%s.** → [%s](https://%s)", headline, shareRepoURL, shareRepoURL)
	}
	return fmt.Sprintf("🧞 %s. → https://%s", headline, shareRepoURL)
}
