# PRD Implementation Trace

Updated: 2026-06-08

This file maps the canonical P0 requirements in
[number-one-platform-prd.md](./number-one-platform-prd.md) to implementation
surfaces and gates. If a requirement cannot name code and a gate, it is not
implementation-aligned.

The trace must stay aligned with
[number-one-platform-hld.md](./number-one-platform-hld.md) and
[number-one-platform-lld.md](./number-one-platform-lld.md). The PRD states the
outcome, the HLD states the architecture boundary, and the LLD states the
runtime contract that tests enforce.

| P0 requirement | Runtime surface | Proof |
| --- | --- | --- |
| Start from a natural task in the current directory | `RunInteractive`, `runLauncher`, direct task intake | `TestDirectTaskArgumentsStartNativeIntake`, CLI UX gate |
| Edit local files directly when clear and safe | local text edit intent handler | `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting`, CLI UX gate |
| Fail closed with exact ambiguity | local edit ambiguity handling and route setup errors | local text edit tests, route missing-CLI tests |
| Answer simple questions compactly | simple answer classifier before work creation | simple factual question tests including typo transcript, CLI UX gate |
| Ask intent for bare entities without artifacts | bare entity classifier before starter packs | intent-first routing fixture, CLI UX gate |
| Route between familiar CLIs, providers, gateways, and local/offline models | adapter registry, router settings, route list/set/auto/status | route command tests, Claude/Codex use-case gate, scorecard gate |
| Treat configured CLI routes as installed-CLI handoffs | `cli_handoff.go`, `generateWithConfiguredProviderDecision` | fake downstream CLI handoff smoke tests for Claude Code and Codex, failed-execution receipt regression, and Gatekeeper rejection fail-closed regression |
| Keep route, token, and runtime diagnostics inspectable | `jini route`, `jini status`, `jini doctor`, route receipt state | route status/list tests, privacy-preserving CLI handoff receipt status test, publish readiness |
| Reuse durable session context without transcript replay | saved work state, metadata-only route receipts, compact status/open/continue | saved work and route receipt tests |
| Keep saved work hidden until explicit commands or title matching | launcher and current-work interruption handling | startup and current-work regression tests |
| Install from release assets without source builds | `install.sh`, release manifest, publish checks | install tests and release gate |
| Preserve customer-value viability and anti-amateur scope | `tools/customer_value_gate.sh`, product settling decisions, competitive benchmark outcome gate | `TestProductViabilityGatePinsCustomerValueAndAntiAmateurBoundary`, customer value gate, scorecard gate |
| Block regressions before commit and push | `tools/run_required_gates.sh`, scorecard PRD completion summary | commit/push/release gate tests, Claude/Codex use-case gate, scorecard PRD implementation completion tests |

Residual hardening:

- Wave 1 command templates use fake downstream CLIs in automated tests and now
  expose `jini check ship --format json` setup status plus local
  `.jini/cli-dogfood.json` validation evidence. Real installed CLI dogfood
  remains required before release claims for auth, approvals,
  output-shape differences, and route receipt privacy.
