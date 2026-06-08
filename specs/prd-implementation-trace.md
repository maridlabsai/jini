# PRD Implementation Trace

Updated: 2026-06-08

This file maps the canonical P0 requirements in
[number-one-platform-prd.md](./number-one-platform-prd.md) to implementation
surfaces and gates. If a requirement cannot name code and a gate, it is not
implementation-aligned.

| P0 requirement | Runtime surface | Proof |
| --- | --- | --- |
| Start from a natural task in the current directory | `RunInteractive`, `runLauncher`, direct task intake | `TestDirectTaskArgumentsStartNativeIntake`, CLI UX gate |
| Edit local files directly when clear and safe | local text edit intent handler | `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting`, CLI UX gate |
| Fail closed with exact ambiguity | local edit ambiguity handling and route setup errors | local text edit tests, route missing-CLI tests |
| Answer simple questions compactly | simple answer classifier before work creation | simple factual question tests, CLI UX gate |
| Ask intent for bare entities without artifacts | bare entity classifier before starter packs | intent-first routing fixture, CLI UX gate |
| Route between familiar CLIs, providers, gateways, and local/offline models | adapter registry, router settings, route list/set/auto/status | route command tests, scorecard gate |
| Treat configured CLI routes as installed-CLI handoffs | `cli_handoff.go`, `generateWithConfiguredProviderDecision` | fake downstream CLI handoff smoke test |
| Keep route, token, and runtime diagnostics inspectable | `jini route`, `jini doctor`, route receipt state | route status/list tests, publish readiness |
| Reuse durable session context without transcript replay | saved work state, route receipts, compact status/open/continue | saved work and route receipt tests |
| Keep saved work hidden until explicit commands or title matching | launcher and current-work interruption handling | startup and current-work regression tests |
| Install from release assets without source builds | `install.sh`, release manifest, publish checks | install tests and release gate |
| Block regressions before commit and push | `tools/run_required_gates.sh` | commit/push/release gate tests |

Residual hardening:

- Wave 1 command templates use fake downstream CLIs in automated tests. Real
  installed CLI dogfood remains required for auth, approvals, and output-shape
  differences.
- macOS CLI handoff runs a Gatekeeper trust check before execution. Rejected
  binaries fail closed instead of triggering downstream execution.
