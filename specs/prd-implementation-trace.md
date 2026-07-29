# PRD Implementation Trace

Updated: 2026-07-18

This file maps the canonical P0 requirements in
[number-one-platform-prd.md](./number-one-platform-prd.md) to implementation
surfaces and gates. If a requirement cannot name code and a gate, it is not
implementation-aligned.

Architecture background lives in the archived
[number-one-platform-hld.md](./archive/number-one-platform-hld.md) and
[number-one-platform-lld.md](./archive/number-one-platform-lld.md); the live
engineering contracts are in
[jini-architecture-blueprint.md](./jini-architecture-blueprint.md). The PRD
states the outcome and the gates below prove it.

## Implemented

| P0 requirement | Runtime surface | Proof |
| --- | --- | --- |
| Start from a natural task in the current directory | `RunInteractive`, `runLauncher`, direct task intake | `TestDirectTaskArgumentsStartNativeIntake`, CLI UX gate |
| Edit local files directly when clear and safe | local text edit intent handler | `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting`, CLI UX gate |
| Fail closed with exact ambiguity | local edit ambiguity handling and route setup errors | local text edit tests, route missing-CLI tests |
| Answer simple questions compactly | simple answer classifier before work creation | simple factual question tests including typo transcript, CLI UX gate |
| Ask intent for bare entities without artifacts | bare entity classifier before starter packs | intent-first routing fixture, CLI UX gate |
| Route between familiar CLIs, providers, gateways, and local/offline models | adapter registry, router settings, route list/set/auto/status | route command tests, Claude/Codex use-case gate, scorecard gate |
| Treat configured CLI routes as installed-CLI handoffs | `cli_handoff.go`, `generateWithConfiguredProviderDecision` | fake downstream CLI command-shape tests for Claude Code and Codex, signed smoke evidence tests, failed-execution receipt regression, and Gatekeeper rejection fail-closed regression |
| Keep route, token, and runtime diagnostics inspectable | `jini route`, `jini status`, `jini doctor`, route receipt state | route status/list tests, privacy-preserving CLI handoff receipt status test, publish readiness |
| Reuse durable session context without transcript replay | saved work state, metadata-only route receipts, compact status/open/continue | saved work and route receipt tests |
| Keep saved work hidden until explicit commands or title matching | launcher and current-work interruption handling | startup and current-work regression tests |
| Install from release assets without source builds | `install.sh`, release manifest, publish checks | install tests and release gate |
| Preserve customer-value viability and anti-amateur scope | `tools/customer_value_gate.sh`, product settling decisions, competitive benchmark outcome gate | `TestProductViabilityGatePinsCustomerValueAndAntiAmateurBoundary`, customer value gate, scorecard gate |
| Block regressions before commit and push | `tools/run_required_gates.sh`, scorecard PRD completion summary | commit/push/release gate tests, Claude/Codex use-case gate, scorecard PRD implementation completion tests |
| Autonomous throttle survival — free-tier core: detect throttle on any route, hold the session, self-resume the same route, name a viable fallback | `throttle_survival.go` (`runWithThrottleSurvival`, throttle classifiers, Retry-After honoring, narration, receipt reason), wired into provider and CLI-handoff paths in `provider.go`/`cli_handoff.go` | `TestRunWithThrottleSurvival*`, `TestIsThrottleSignal*`, `TestClassifyCLIThrottleOutput*`; live transcript: fake throttling downstream CLI, hold narrated, advertised wait honored, same-route resume, work saved |
| `Auto`/`Ask` execution mode, switchable mid-session, plus Ask-mode approval before throttled-work resume with fail-closed parking | `execution_mode.go` (fail-closed setting, `runMode`), `throttle_survival.go` approver seam (`throttleApprover`, `autoApprover`/`failClosedApprover`/`cliPromptApprover`, `configureThrottleApproverForEntry`), `throttle_park.go` (resumable park), `jini continue` park-resume in `app.go` | `TestRunMode*`, `TestModeIsARoutedTopLevelCommand`, `TestConfigureThrottleApproverForEntry`, `TestCLIPromptApprover*`, `TestRunWithThrottleSurvivalDeclineReturnsTypedError`/`*FailClosedApprover*`, `TestThrottlePark*`, `TestRunContinueResumesPark`, `TestStandaloneThrottleFamilyErrorPassesThrough`; live transcript: Ask decline parks + `jini continue` resumes, Auto silent hold honors advertised 2s wait then resumes same route |
| Savings ledger MVP — dollar-primary, OS-currency-localized, imputed-and-labeled per-task receipts with a running counter and dashboard | `savings_pricing.go` (dated price table), `savings_currency.go` (OS-currency detection, dated FX, formatting), `savings_ledger.go` (global ledger, folding integrity invariant), `savings_receipt.go` (compute + one-entry-per-task wiring), `savings_render.go` (footer + startup counter), `savings_command.go` (`jini savings` text/JSON) | `TestBaselineForRoute`, `TestSavingsUSD*`, `TestDetectDisplayCurrency*`, `TestLocalize*`, `TestFormatMoney*`, `TestSavingsLedger*` (round-trip, folding invariant, tampered→nil), `TestComputeTaskSavings*`, `TestRecordSavings*`, `TestSavingsFooter*`, `TestSavingsStartupCounter*`, `TestRunSavings*`, `TestHonesty_*`; live transcript: work task footer `₹0.04 (US$0.0005)`, `jini savings` totals + disclosure, JSON report |

## Not Yet Implemented (v1 backlog)

These rebuilt-PRD requirements have no runtime surface today. Per this
trace's own rule they are explicitly not implementation-aligned yet — this
list is the input to the implementation phase, not a claim. Each names the
PRD section and the gate that will eventually prove it. (Rendered as a list,
not a table: the scorecard trace parser counts any three-cell table row as an
implemented P0 row.)

- Autonomous throttle survival, remaining slices (§P0 Outcome Requirements,
  §Routing And Resource Policy): the free-tier same-route core and the
  Ask-mode resume approval are implemented (see Implemented table); still
  unbuilt are paid Autopilot mid-task route switching and the throttle-dodge
  counter feeding the savings ledger. Future proof: throttle-resilience
  release gate.
- Savings ledger and receipts, remaining slices (§Savings Ledger And
  Receipts): the MVP core is implemented (see Implemented table) — per-task
  imputed receipt, work-task footer, startup counter, and `jini savings`
  text/JSON, all dollar-primary with OS-currency localization and a disclosed
  estimation basis. Still unbuilt: `jini savings --report` HTML export,
  `--share` card, ANSI trend charts, and real metered-usage capture (the seam
  for literal rows). Future proof: those surfaces plus a literal-capture
  fixture.
- Token-economy regression suite (§Token Economy). Future proof:
  token-efficiency regression gate in the release tier.
- On-the-fly skills and agents as plain reviewable files (§Skills And
  Agents). Future proof: skills/agents creation fixture in the CLI UX gate.
- User preference envelope — never/prefer/pin per model/route, speed bias,
  plain-file persistence (execution-routing-policy §Absorbed Policies).
  Future proof: preference-constraint routing tests.
- BYO credential validation with typed errors and OS keychain storage,
  including the xAI (Grok) shape (§Routing And Resource Policy). Future
  proof: per-shape validation fixtures; a shape without a passing fixture is
  not claimed.
- Paywall entitlements failing closed with manual free equivalents (§Tier
  Boundary). Future proof: paywall fail-closed test.
- Escalation cost quote before spending on a stronger rung (§P0 Outcome
  Requirements). Future proof: escalation-quote fixture.
- TTFV under five minutes on all three OSes and first-task success
  benchmarks (§Goals And Scope, §Gates). Future proof: TTFV measurement and
  first-task success release bars in the gate matrix.

Residual hardening:

- Wave 1 command templates still use fake downstream CLIs for automated
  command-shape coverage, but release readiness now requires signed
  `.jini/cli-smoke.json` evidence, recent `.jini/cli-dogfood.json`
  validation evidence, and `jini check ship --format json` setup status for
  claimed routes. Real installed CLI dogfood remains required on tester
  machines for auth, approvals, output-shape differences, route receipt
  privacy, and signed smoke freshness.
- Selective-consistency and refinement drafts
  (`generateConsistencyDraft`/refine paths in `provider.go`) still call the
  providers directly, bypassing `runWithThrottleSurvival`. A throttle during
  a draft fails that draft rather than holding; the primary answer is
  unaffected. Future proof: route these auxiliary drafts through the survival
  wrapper (with a draft-scoped hold budget) or drop them under throttle.
