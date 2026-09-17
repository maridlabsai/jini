# Savings Ledger MVP — Design

Status: Draft (brainstorming output; awaiting user review → plan-and-execute)
Author session: 2026-07-23
PRD anchor: `specs/number-one-platform-prd.md` §Savings Ledger And Receipts,
§Token Economy. Trace backlog item: "Savings ledger and receipts".

## Purpose

Make Jini's core value — routing work off metered APIs — visible in the
user's own money. Every routed work task ends with a one-line savings footer;
startup shows a running counter; session end shows a roll-up; `jini savings`
renders a text summary. Figures are dollar-primary, localized to the OS
currency, and **honest**: an inflated or mislabeled savings claim is a
release-blocking defect (PRD).

## Scope (user-ratified 2026-07-23)

IN: per-task receipt → persistent ledger; one-line footer on work tasks;
startup running counter; session-end roll-up; `jini savings` text summary
with literal-vs-imputed labeling and OS-currency localization.

OUT (later slices): `jini savings --report` HTML export; `--share` card; ANSI
trend charts; real metered-usage capture (the seam for "literal" rows exists
but is not populated in MVP); paid Autopilot predictive-avoidance counting.

## Measurement model (the honesty core)

We capture **no real token counts today**: `generateProviderText` returns
`(string, error)` and discards usage; CLI handoffs report only char counts +
duration. Therefore every MVP savings figure is **imputed** and labeled so.
There are no "literal" rows in MVP — that is honest, because we record no real
metered spend. Literal rows appear only when real-usage capture lands later.

Chain (each step disclosed):

1. `est_tokens ≈ (prompt_chars + output_chars) / charsPerToken` where
   `charsPerToken = 4` (conservative, documented). Estimate, not a
   measurement — never presented as a multiplier.
2. `usd_saved = est_input_tokens * inUSDPer1M/1e6 + est_output_tokens *
   outUSDPer1M/1e6`, priced against the route's baseline metered model at
   posted API-equivalent prices (subscription/local routes cost the user $0
   incrementally, so the counterfactual metered cost IS the saving).
3. Localized display: `local = usd_saved * fxRate`, shown as
   `≈ <local> (USD $<usd>, imputed; FX <rate> as of <fxAsOf>)`. USD is
   never hidden — it is the auditable source-of-truth figure.

Route classes: `subscription` (Codex/Claude Code handoffs — imputed savings),
`local` (local model — imputed savings vs frontier equivalent), `metered`
(BYO API — user paid; MVP records $0 saved rather than guess a cheaper
baseline; the literal-capture slice refines this).

## Components

### `savings_pricing.go` — auditable price table
- `type modelPrice struct { InUSDPer1M, OutUSDPer1M float64 }`
- `savingsPriceAsOf = "2026-07"` and a `map[baseline]modelPrice` sourced from
  official posted prices, with a comment citing the source per row.
- `baselineForRoute(routeClass, routeLabel string) (baseline string, ok bool)`
  maps Codex→GPT-class, Claude Code→Claude-class, local→nearest frontier
  equivalent. Unknown → no saving recorded (fail honest, not fail generous).

### `savings_currency.go` — OS-currency localization
- `detectDisplayCurrency() string`: `JINI_CURRENCY` override → parse
  `LC_MONETARY`/`LC_ALL`/`LANG` region (e.g. `en_IN`→INR) via region→currency
  map → default `USD`.
- Checked-in dated FX: `usdToRate map[string]float64`, `fxAsOf = "2026-07"`;
  `JINI_FX_RATE` override (applies to the detected currency, for
  testing/correctness). Missing rate → fall back to USD + note.
- `formatMoney(currency string, amount float64) string`: small built-in
  per-currency format map (symbol, decimals, grouping; INR uses lakh/crore
  grouping and `₹`). No external i18n dependency.
- `localize(usd float64) localizedAmount` returning both the formatted local
  string and the USD figure + FX disclosure fields.

### `savings_ledger.go` — persistent store
- Location: `~/.jini/savings-ledger.json` (GLOBAL/cross-repo, same home as
  `mode.json`; savings is a property of the user, not one workspace).
- `type savingsEntry struct { SchemaVersion, ContextType ("JiniSavingsEntry"),
  At (RFC3339), Repo, RouteLabel, RouteClass, EstInputTokens, EstOutputTokens
  int, USDSaved float64, Imputed bool (always true in MVP), ThrottleDodged
  bool, Title string }`.
- `type savingsLedger struct { SchemaVersion, ContextType ("JiniSavingsLedger"),
  Entries []savingsEntry, Folded savingsTotals, Totals savingsTotals }`.
  `Folded` = aggregates of entries already dropped from `Entries` (task count,
  USD saved, dodges); `Totals` = grand total. **Integrity invariant, always
  checkable on load:** `Totals == Folded + sum(Entries)`. A ledger that fails
  this is treated as corrupt (→ nil). Without a separate `Folded` the grand
  total would be unverifiable once bounding drops entries — that gap is the
  reason `Folded` exists.
- `appendSavingsEntry(entry)`: read-modify-write with atomic rename (park
  discipline), 0600. Bounded: keep the last N entries verbatim (N documented,
  e.g. 500); when trimming, add the dropped entry's contribution to `Folded`.
  `Totals` is recomputed as `Folded + sum(Entries)` on every write.
- `loadSavingsLedger() *savingsLedger`: corrupt → nil (never blocks a task).
- Concurrency: single-user last-writer-wins accepted; documented.

### Wiring (`provider.go`, `simple_answers.go`, launcher/status)
- After a successful routed **work** task in
  `generateWithConfiguredProviderDecision`, compute + append an entry, using
  char counts already available (CLI handoff receipt chars; provider path
  estimates from prompt/output strings). Gate: skip when
  `request.Standalone` (simple questions stay clean — no footer, no entry).
  **Invariant: exactly one entry per user-visible work task.** The entry is
  written at the single primary-answer success return; auxiliary generations
  (selective-consistency / refinement drafts, which call `generateProviderText`
  directly) must not write entries. Verified by test to prevent double-count.
- Footer: rendered by the caller that prints work-task output, one line,
  localized, `(imputed)`, pointer to `jini savings`.
- Startup counter: one line from `Totals` in the launcher/status path; absent
  when the ledger is empty.
- Session roll-up: totals for entries added this process, shown at session
  end where the current end-of-work summary renders.

### `jini savings` command
- New top-level command (added to the taught surface + settling-doc amendment,
  same pattern as `jini mode`). Text summary:
  - headline: localized total + `(USD $X, imputed)`;
  - task count, throttles dodged (0 in MVP);
  - breakdown by route class;
  - imputed vs literal split (100% imputed in MVP, stated plainly);
  - disclosure footer: `charsPerToken`, price-table `as of`, FX `as of`.
- `--format json` for machine-readability and gate assertions.

## Error handling
- Ledger/pricing/FX failures never fail a task: a receipt is best-effort. A
  task that can't be priced records no entry rather than a wrong one.
- Unknown currency/route/baseline → degrade to USD / no-entry, never to an
  inflated or fabricated figure.

## Testing (honesty gate)
- `savings_pricing_test.go`: baseline mapping; unknown route → no saving.
- `savings_currency_test.go`: locale→currency parsing (`en_IN`, `de_DE`,
  `en_GB`, unset→USD); `JINI_CURRENCY`/`JINI_FX_RATE` overrides; INR lakh
  formatting; missing-rate → USD fallback + note; USD always present.
- `savings_ledger_test.go`: append round-trip, atomic rename, 0600, corrupt→
  nil; bounded folding preserves the invariant `Totals == Folded +
  sum(Entries)` across many appends; a hand-tampered `Totals` load → nil.
- `savings_receipt_test.go`: work task writes one entry; **standalone question
  writes none and shows no footer**; imputed flag always true; USD figure
  accompanies every localized figure.
- `savings_command_test.go`: `jini savings` totals == ledger sum
  (measurement-integrity); disclosure lines present; JSON shape.
- Gate: extend `tools/run_required_gates.sh` coverage / a savings-methodology
  assertion so a literal label without recorded spend, or a localized figure
  without its USD source, fails the gate.

## Trace / drift bookkeeping
- Move "Savings ledger and receipts" from trace backlog to Implemented
  (scoped to MVP; note deferred slices). Pair with a
  `product-settling-decisions.md` entry (drift gate requires it). Bump the
  pinned P0 count + residual-hardening assertions in `app_internal_test.go`
  (lesson from the Auto/Ask trace commit).

## Refinements from stress-test (2026-07-23)

- **Figure threading:** `generateWithConfiguredProviderDecision` attaches the
  computed entry to `routeDecision.SavingsEntry *savingsEntry` (same pattern
  as `CLIHandoffReceipt`/`Reason`). The output caller renders the footer from
  it — no ledger re-read. `nil` ⇒ no footer (simple questions, unpriced
  routes).
- **Route class:** `routeClassForDecision(decision, provider)` —
  `cliHandoffMode`→`subscription`, `provider.ID == "local-preview"`→`local`,
  else `metered`.
- **Input chars:** provider path uses `len(systemPrompt)+len(userPrompt)` in +
  `len(text)` out; CLI path uses the receipt's `prompt_chars` in + `stdout_chars`
  out. Both disclosed as estimates.
- **Offline failover:** the entry records the route that actually answered
  (fallback decision), written once at the single success return — never for a
  throttled primary that failed over.
- **Footer seam:** Phase 4 must first locate the exact work-task output render
  path (distinct from the simple-answer path) before wiring the footer.

## Open assumptions (flagged for reviewer)
- `charsPerToken = 4` is a rough industry heuristic; conservative by intent.
- Price table + FX table are point-in-time; disclosed with dates; staleness
  accepted for an offline-first free CLI (user-ratified 2026-07-23).
- `metered` route savings = $0 in MVP (no guessed cheaper baseline).
