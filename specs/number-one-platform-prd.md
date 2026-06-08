# Number One Platform PRD

Updated: 2026-06-08

This is the canonical near-term PRD. It reduces older broad platform thinking to
the work that matters for GTM. If this document conflicts with exploratory
platform docs, demo docs, or older PRDs, this document and
[product-settling-decisions.md](./product-settling-decisions.md) win.

Delivery chain:

- this PRD defines what matters
- [launcher-intake-design.md](./launcher-intake-design.md) defines the dev design
- [number-one-development-plan.md](./number-one-development-plan.md) defines the active cuts
- drift requires [product-settling-decisions.md](./product-settling-decisions.md)

## Product Thesis

Jini should be built as a CLI-first AI work router and durable session layer for
people already using multiple AI coding tools, CLIs, online models, and local
models.

The near-term GTM product is not the broad OS. The wedge is a high-quality CLI
that reduces token waste, avoids avoidable throttling, preserves useful session
state, and lets users move between configured tools without learning a new conversation style.

Core charter: intent-first Claude/Codex parity outranks feature expansion.

## P0 Jobs

- Start from a natural task in the current directory.
- Edit local files directly when the ask is clear and safe.
- Answer simple questions directly without dumping saved work state.
- Route between configured CLIs, providers, and local/offline models.
- Reuse durable session context without replaying large transcripts.
- Keep route, token, and local runtime diagnostics inspectable.
- Preserve familiar Claude/Codex-style expectations.
- Block regressions with required gates before commit and push.

## UX Contract

Bare `jini` is a task prompt, not a dashboard.

Required behavior:

- Show only `Jini`, `Describe the task.`, and the short help hint on startup.
- Keep saved work hidden until the user asks for `status`, `continue`, `open`,
  or `help`.
- Resume another saved thread when the user naturally types its saved title.
- Treat a new freeform request as work to execute, not as a Start/Keep modal.
- Never show a full current-work overview for a simple factual question.
- Never require users to learn Jini-specific command vocabulary before value.

Forbidden front-door behavior:

- no saved-work dashboard on bare startup
- no visible `Switch` startup control
- no `Start/Keep` interruption model
- no Working Draft for obvious file edits
- no verbose Goal/Working-with/status frame for simple questions
- no visible agent-role theater in the free tier

## Routing And Token Economy

Token frugality is P0. Jini must spend context only when it improves the result.

Routing requirements:

- Default to the cheapest safe route that can complete the task.
- Use local/offline routes when they meet the task quality bar.
- Escalate to stronger online routes when correctness, codebase scope, or
  tool access requires it.
- Keep route choice inspectable through `jini route`.
- Preserve enough session state to continue work without replaying stale chat.

Avoiding throttling is P1. Jini should detect configured CLI or provider
pressure, choose viable alternatives when available, and resume work cleanly.

Power awareness is P1. In powered mode, Jini can choose higher-throughput local
or online routes. In low-battery mode, Jini should avoid wasteful local model
loads unless the user explicitly asks for offline work.

## Competitive And Learning Guards

#### P0.10 Competitive release pressure

Competitor watching remains a P0 feature-selection loop through
[competitive-release-plan.md](./competitive-release-plan.md).

watch packets must decide the next feature candidates.
Each candidate must be classified as copy, integrate, watch, reject, delete.
This is how Jini should reject, downgrade, or delete requirements that create product drift.

#### P0.11 Compounding user productivity learning

Jini must learn user context, usage, habits, and repeated patterns over time,
with inspectable controls. The learning loop should produce repeated prompts, better defaults, better route choices, and faster future sessions without silently weakening privacy, safety, or user control.

## Free Tier

The free tier proves the CLI wedge without giving away the commercial OS.

Free includes:

- direct task intake
- direct local file edits
- manual route inspection and route switching
- compact status, continue, and open flows
- basic offline/local route support when configured
- basic token-frugal session reuse
- setup and provider diagnostics

Free excludes:

- developer-agent fleets
- tester-agent fleets
- commercial skills-based OS productivity suite
- managed throttle prediction and recovery
- governed approval workflows
- team policy, audit, and automation loops

## Commercial Tier

Commercial value must be materially higher than the free CLI.

Commercial includes:

- managed route policy across teams and devices
- throttle-aware and quota-aware recovery
- skills and delegation framework
- developer and tester agents hidden behind normal outcomes
- governed approvals and audit trails
- cross-device and offline-online session continuation
- company automation loops for quality, release, support, and roadmap work

Commercial UX still follows the same rule: simple task in, useful result out.
The user should not manage agent trees to get value.

## Roadmap

P0 now:

- task-first startup even with saved work
- direct file edit reliability
- simple question direct-answer behavior
- route list, route set, route auto, route status
- self-sufficient install from release assets
- CLI UX regression gate in commit gates
- intent/parity golden transcript gate in commit gates
- product PRD drift gate in commit gates

P1 next:

- throttle-aware route switching
- powered-mode and low-battery routing
- offline local-model quality regression harness
- cross-surface session handoff
- clearer diagnostics for configured CLIs and local runtimes

P2 later:

- desktop and mobile app surfaces
- richer commercial agent and skills UI
- team-level policy controls
- broad proof verticals and demo templates

## Gates

Every commit that touches product, CLI UX, routing, or docs must keep these
gates green:

- `go test ./...`
- `bash tools/cli_ux_regression_gate.sh`
- `bash tools/product_prd_drift_gate.sh`
- `jini scorecard-gate --format json`
- `jini check ship --format json` before push/release

The gates exist to prevent old behavior from returning: verbose startup,
Start/Keep modals, hidden Python-era assumptions, stale docs, and broad PRD drift that slows GTM.
