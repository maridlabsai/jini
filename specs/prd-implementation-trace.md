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

## Not Yet Implemented (v1 backlog)

These rebuilt-PRD requirements have no runtime surface today. Per this
trace's own rule they are explicitly not implementation-aligned yet — this
list is the input to the implementation phase, not a claim. Each names the
PRD section and the gate that will eventually prove it. (Rendered as a list,
not a table: the scorecard trace parser counts any three-cell table row as an
implemented P0 row.)

- Autonomous throttle survival, the number one P0 (§P0 Outcome Requirements,
  §Routing And Resource Policy): detect throttle, hold session, self-resume
  same-route free / auto-switch paid. Future proof: throttle-resilience
  release gate with transcript evidence.
- Savings ledger and receipts (§Savings Ledger And Receipts): per-task
  receipt, session roll-up, startup counter, `jini savings` dashboard.
  Future proof: savings-methodology audit gate (literal-vs-imputed labels).
- Token-economy regression suite (§Token Economy). Future proof:
  token-efficiency regression gate in the release tier.
- On-the-fly skills and agents as plain reviewable files (§Skills And
  Agents). Future proof: skills/agents creation fixture in the CLI UX gate.
- `Auto`/`Ask` execution mode surface, switchable mid-session (§UX
  Contract). Future proof: mode-switch fixture preserving session state.
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
