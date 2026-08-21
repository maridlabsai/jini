# Product Maturity Coverage — Test & Guardrail Program

Status tracker for the maturity loop: comprehensive test coverage plus new
quality-guardrail layers, run iteratively until Jini is mature across domains,
task complexity, request kinds, and response qualities. One coherent, gate-green,
committed slice per loop pass.

Scope decision (2026-08-14): build the **missing quality layers** (validators),
not only tests — while keeping the free/commercial split, token frugality, and
the no-verbose-ceremony non-negotiables. Guardrails are **detection-first**
(pure invariants enforced in tests, zero runtime cost) and **redaction-only**
where they touch runtime (they may remove sensitive content, never add verbosity
or rewrite tone silently).

## Dimension × surface matrix

Legend: ✅ built+tested · 🟡 partial · ⬜ gap · N/A not Jini's layer (downstream
model/CLI owns it)

| Dimension | Jini surface | Status |
| --- | --- | --- |
| Domains (code/prose/data/devops/math/sql/regex/git/k8s/…) | intent routing + response shape | ✅ (P4) 23-case corpus |
| Task complexity (simple→multi-step) | simple-answer / work / native-loop | ✅ (P4) trivial-compact + multi-step cases |
| Request kinds (question/edit/ambiguous/attachment) | classifiers, inputItems | ✅ (P3/P4) |
| Security — secret leakage in Jini output | `respguard` leaked-secret detector | ✅ (P1) |
| Security — sandbox/permission, receipt privacy | posture, route receipts, loop step evidence | ✅ (P5) native-loop actions visible |
| Accessibility — plain-text / no ANSI reliance | `respguard` ansi/control detector | ✅ (P1) |
| Readability — no run-on walls, plain | `respguard` run-on detector | ✅ (P2) enforced over corpus |
| Tone — neutral, no fear/hype | `respguard` tone detector | ✅ (P2) enforced over corpus (0 violations) |
| Citations / references — cited paths are real | `respguard` broken-reference detector | ✅ (P7) enforced over corpus |
| Attachments — prompt + `@file` intake | `attachments.go` (validate/forward/inline) | ✅ (P3) |
| Attachments — images to native/provider model | `anthropicUserContent` base64 blocks | ✅ (P8) Anthropic vision; others staged |
| Response qualities on model answers (citations, style) | downstream model/CLI | N/A (Jini can only guard, not author) |

## Functional harness (always-on health)

A single scenario table (`selfCheckScenarios` in `internal/app/selfcheck.go`) is
the source of truth for "is Jini functional across all scenarios". It backs two
surfaces that never drift:

- **`jini check functional`** — a runtime product capability: drives Jini's core
  scenarios end-to-end, offline (local-preview) and hermetically (each scenario
  in a throwaway cwd/home, no workspace side effects), and prints `ok/FAIL` per
  scenario with an `N/N passed` summary. A user — or Jini itself in a dogfood
  loop — can confirm health at any time, no network.
- **`TestFunctionalHarness`** — the same scenarios in-process as a Go test.
- **`tools/functional_smoke.sh`** — runs the harness (core tier) and, with
  `--live`, the real installed-CLI route smoke/dogfood that the push gate needs.

Scenarios covered: arithmetic (symbol + word), capital, ambiguous entity,
offline standalone question, work-task draft, direct file edit (+ side-effect
check), attachment present/missing, help, doctor, route list/status, savings,
trust list, unknown-command error, unicode. Add a scenario in one place
(`selfCheckScenarios`) and both surfaces pick it up.

## Loop protocol (per pass)

1. Pick the highest-value ⬜/🟡 cell.
2. Build the guardrail (pure detector or redactor) and/or coverage corpus.
3. Unit-test the detector; enforce it across the black-box corpus where safe.
4. Premortem + adversarial review; run `tools/run_required_gates.sh commit`.
5. Commit one slice; update this matrix + the iteration log.

## jini-through-jini (dogfood) track

Commitment: further Jini development runs *through* Jini, proxying to an
installed Claude Code under the hood. This track must stay unblocked at all
times. Setup: `jini mode auto`, route `claude-code` (auto-detected ok), and — for
autonomous edits — `jini trust --autonomous` in the repo (interactive-terminal
consent only, by fail-safe design; a non-TTY grants nothing).

- **D1 (2026-08-16)** — proved jini→claude proxy end-to-end (jini handed a repo
  question to claude, which read the file and answered). Fixed the blocker it
  surfaced: the standalone-question path capped every attempt at 10s, killing a
  `claude --print` subprocess mid-flight. Now CLI hand-offs get a generous
  per-attempt budget (`cliHandoffAttemptTimeout`, 3m default, override-able); an
  explicit `JINI_STANDALONE_QUESTION_TIMEOUT` still wins for any route.

## Iteration log

- **P8 (2026-08-19)** — native multimodal (competitive priority #5, closes the
  P3 image gap): `anthropicUserContent` builds base64 image content blocks for a
  vision-capable route, so an image `@attachment` reaches a native/provider model
  (BYO Claude) — not only hand-off CLIs. `providerGenerationRequest.Images`
  carries resolved images; the intake attaches them when `routeSupportsVision`
  (Anthropic today; local/preview/bedrock/azure stay the honest can't-read note).
  Fail-safe: no/unsupported/oversize/unreadable images → single text block,
  byte-identical to the prior payload. Per-image cap 5MB; png/jpeg/gif/webp.
- **P7 (2026-08-18)** — reference/citation integrity (last named dimension):
  `brokenPathReferences` flags a file Jini *claims* to have acted on ("Updated
  X", "wrote N bytes to X", "· edited X") that doesn't exist under the cwd —
  applied only to explicit file-claim phrases, never model-answer prose, so
  there are no false positives on placeholders/flags/URLs/mentions. Enforced
  across the corpus (Jini's edit-intent path correctly creates the file it
  cites) and unit-tested for conservatism.
- **P6 (2026-08-18)** — token-frugality regression suite (competitive vector
  #4, PRD P0): `TestMaturityCorpus_TokenFrugality` budgets trivial answers to
  ≤40 chars and forbids replaying the question (transcript-replay avoidance), and
  caps every corpus case at a 4000-char anti-bloat ceiling. Baselines are tiny
  today (trivial 3–7 chars, work drafts ~150, help 443, doctor 655), so the
  suite locks frugality against regression rather than chasing a fix.
- **P5 (2026-08-17)** — permissioned-sandbox execution evidence (competitive
  vector #2): the native loop now emits a compact line per action as it runs
  (`· read x`, `· edited y`, `· ran: <cmd> — ok/failed`) via an optional
  `progress` writer, wired to stdout in `maybeRunNativeLoop`. Autonomous
  execution is no longer hidden behind a final-result-only view — visible
  permissions/progress, matching Claude Code / Codex. Names targets and command
  outcomes, never file contents. nil progress stays a silent no-op.
- **P4 (2026-08-15)** — deepened the intent/domain corpus to 23 cases
  (sql/regex/git/shell/docker/k8s/security/translate/explain/compare/debug/
  multi-step) and added response-shape parity for `intent-first-cli-parity`:
  `TestMaturityCorpus_TrivialPromptsAreCompact` asserts trivial prompts answer in
  one compact line with no work-draft ceremony. Fixed a real parity gap it
  surfaced — word-form arithmetic ("17 times 23", "2 plus 2", "divided by") now
  answers offline like a familiar CLI instead of falling through to "no route".
- **P3 (2026-08-15)** — attachments intake (`attachments.go`): a prompt can now
  carry `@file`/`@image` references alongside text. Jini validates them up front
  and **fails closed** with an exact message on a missing path; acknowledges
  resolved attachments compactly; **forwards** the prompt verbatim so Claude Code
  / Codex read `@path` (incl. images) natively; and **inlines text** file content
  (capped) for local/provider routes, noting honestly when an image/audio can't
  be read by a non-handoff route. Limits (P0 frugality): ≤10 attachments,
  ≤64KB/file, ≤192KB total inline. Social handles (`@alice`) are not mistaken for
  files. Native/provider models: text attachments yes (inlined); image/audio not
  yet wired to provider payloads (staged) — route to a hand-off CLI for those.
- **P2 (2026-08-15)** — enforced tone + readability over the corpus. Tone: 0
  violations (copy already neutral). Readability: refined the detector to flag
  only multi-sentence run-on walls (a single long sentence soft-wraps in the
  terminal and is fine); split the two flagged lines. Fixed the `jini trust`
  consent paragraph to one sentence per line (wording unchanged) so an important
  consent screen stays scannable; the doctor AUTO_ROUTE line is a single
  sentence and passes. `TestMaturityCorpus_ToneAndReadability` now green.
- **P1 (2026-08-14)** — `respguard.go`: pure user-facing-output audit library
  (`AuditUserFacingOutput`) with detectors for ANSI/control chars
  (accessibility), leaked secrets (security), overlong lines/walls (readability),
  and fear/hype lexicon (tone). Unit-tested. Enforced HARD invariants (no ANSI,
  no control chars, no leaked secrets) across a black-box command corpus; tone +
  readability enforcement staged for P2 after auditing current copy.
