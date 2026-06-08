# Number One Platform PRD

Updated: 2026-06-08

This is the canonical near-term PRD. It defines the work that matters for GTM.
If it conflicts with exploratory specs, demo docs, or older PRDs, this document
and [product-settling-decisions.md](./product-settling-decisions.md) win.

Delivery chain:

- this PRD defines what matters
- [launcher-intake-design.md](./launcher-intake-design.md) defines the dev design
- [number-one-development-plan.md](./number-one-development-plan.md) defines active cuts
- drift requires [product-settling-decisions.md](./product-settling-decisions.md)

## Product Thesis

Jini is a CLI-first AI work router and durable session layer for people already
using multiple coding CLIs, online models, and local models.

The near-term product is not the broad OS. The wedge is a high-quality CLI that
reduces token waste, avoids avoidable throttling, preserves useful session
state, and lets users move between configured tools without learning a new
conversation style.

Core charter: intent-first Claude/Codex parity outranks feature expansion.

## P0 Outcome Requirements

- Start from a natural task in the current directory.
- Edit local files directly when the ask is clear and safe.
- Fail closed with exact ambiguity when a file or code task is unclear.
- Answer simple questions compactly without creating work.
- Ask intent for bare entities without creating artifacts.
- Route between familiar CLIs, providers, gateways, and local/offline models.
- Treat configured CLI routes as real installed-CLI handoffs, not provider API
  aliases.
- Keep route, token, and local runtime diagnostics inspectable through `jini route`.
- Reuse durable session context without replaying large transcripts.
- Keep saved work hidden until `status`, `continue`, `open`, `help`, or natural
  title matching.
- Install from release assets without requiring source builds.
- Block regressions with required gates before commit and push.

## UX Contract

Bare `jini` is a task prompt, not a dashboard.

Required behavior:

- Show only `Jini`, `Describe the task.`, and the short help hint on startup.
- Treat a new freeform request as work to execute, not as a Start/Keep modal.
- Resume saved work only through explicit commands or natural title matching.
- Never show a full current-work overview for a simple factual question.
- Never require users to learn Jini-specific command vocabulary before value.

Forbidden front-door behavior:

- no saved-work dashboard on bare startup
- no visible `Switch` startup control
- no `Start/Keep` interruption model
- no Working Draft for obvious file edits
- no verbose Goal/Working-with/status frame for simple questions
- no visible agent-role theater in the free tier
- no hard-coded entity-to-template routing

## Routing And Resource Policy

Token frugality is P0. Jini must spend context only when it improves the result.

Routing requirements:

- Default to the cheapest safe route that can complete the task.
- Use local/offline routes when they meet the task quality bar.
- Escalate to stronger online routes when correctness, codebase scope, or tool
  access requires it.
- Label provider API routes separately from CLI handoff routes. A route named
  `codex` or `claude-code` must invoke that CLI or fail closed with setup
  guidance.
- Preserve enough session state to continue work without replaying stale chat.

Avoiding throttling is P1. Detect configured CLI/provider pressure, choose
viable alternatives, and resume work cleanly.

Power awareness is P1. In powered mode, Jini can choose higher-throughput local
or online routes. In low-battery mode, Jini should avoid wasteful local model
loads unless the user explicitly asks for offline work.

## Market And Learning Guards

Competitor watching is a P0 feature-selection loop through
[competitive-release-plan.md](./competitive-release-plan.md), but it does not
create active scope by itself.

- Competitor watch packets can nominate next feature candidates and deletion
  candidates.
- Each candidate must be classified as copy, integrate, watch, reject, or
  delete.
- No competitor finding becomes active scope unless the decision record changes.

User productivity learning remains P0 only when it improves the CLI wedge:

- learn stable user context, usage, habits, and repeated patterns
- produce fewer repeated prompts, better defaults, and better route choices
- keep learning inspectable and controllable
- avoid hidden surveillance, broad OS scope, or free-tier agent-suite creep

## Tier Boundary

Free proves the CLI wedge:

- direct task intake and local file edits
- manual route inspection and route switching
- compact status, continue, and open flows
- basic offline/local route support when configured
- basic token-frugal session reuse
- setup and provider diagnostics

Free excludes:

- developer-agent fleets, tester-agent fleets, a commercial skills-based OS productivity suite
- managed throttle recovery, governed approval workflows, team policy, audit,
  and automation loops

Commercial value must be materially higher than the free CLI: managed
route/throttle policy, governed skills and delegation, cross-device and
offline-online continuation, team audit, and automation loops. Commercial UX
must still follow the same rule: simple task in, useful result out.

## Roadmap

P0 now:
- intent-first CLI parity
- task-first startup even with saved work
- direct file edit reliability
- simple question direct-answer behavior
- route list, route set, route auto, route status
- real downstream CLI handoff and familiar-tool adapter breadth
- self-sufficient install from release assets
- CLI UX, PRD drift, and scorecard gates in commit gates

P1 next:
- throttle-aware route switching
- powered-mode and low-battery routing
- offline local-model quality regression harness
- cross-surface session handoff
- clearer diagnostics for configured CLIs and local runtimes

Deferred until a decision-record update:

- desktop and mobile app surfaces
- richer commercial agent and skills UI
- team-level policy controls
- broad proof verticals and demo templates
- company automation loops

## Gates

Every commit that touches product, CLI UX, routing, or docs must keep these
gates green:

- `go test ./...`
- `bash tools/cli_ux_regression_gate.sh`
- `bash tools/product_prd_drift_gate.sh`
- `jini scorecard-gate --format json`
- `jini check ship --format json` before push/release

The gates exist to prevent old behavior from returning: verbose startup,
Start/Keep modals, hidden Python-era assumptions, stale docs, and broad PRD drift.
