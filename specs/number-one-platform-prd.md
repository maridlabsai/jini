# Number One Platform PRD

Updated: 2026-06-10

This is the canonical near-term PRD. It defines the work that matters for GTM.
If it conflicts with exploratory specs, demo docs, or older PRDs, this document
and [product-settling-decisions.md](./product-settling-decisions.md) win.

Delivery chain: this PRD defines what matters; [number-one-platform-hld.md](./number-one-platform-hld.md) defines architecture
boundaries; [number-one-platform-lld.md](./number-one-platform-lld.md) defines runtime contracts;
[launcher-intake-design.md](./launcher-intake-design.md) and [number-one-development-plan.md](./number-one-development-plan.md) define execution;
[macOS app planning](./macos-app-prd.md) is a focused P1 desktop surface over the same session model;
drift requires [product-settling-decisions.md](./product-settling-decisions.md).

## Product Thesis

Jini is a CLI-first AI work router and durable session layer for people already
using multiple coding CLIs, online models, and local models.

The near-term product is not the broad OS. The wedge is a high-quality CLI that
reduces token waste, avoids avoidable throttling, preserves useful session
state, and lets users move between configured tools without learning a new
conversation style.

Core charter: intent-first Claude/Codex parity outranks feature expansion.

Customer value bar: every shipped cut must improve token frugality, throttle
resilience, tool-switching reduction, direct action, safety, or session
continuity for configured tools the user already trusts. If it does not, it is
not P0 work.

## Current Release Contract

R0 is CLI-first: minimal task prompt, compact answers, clear local edits,
explicit route/status/continue/open/doctor flows, real provider/local/CLI route
execution or fail-closed setup guidance, and release-asset install. R0 must not
claim dashboards, `Start/Keep`, `Task Snapshot` scaffolds, generally available
desktop/mobile/team/agent-suite surfaces, or automatic commercial optimization.
Side effects follow the LLD approval matrix.

## P0 Outcome Requirements

- Start from a natural task in the current directory.
- Edit local files directly when the ask is clear and safe.
- Fail closed with exact ambiguity when a file or code task is unclear.
- Answer simple questions compactly without creating work.
- Ask intent for bare entities without creating artifacts.
- Route between familiar CLIs, providers, gateways, and local/offline models.
- Treat configured CLI routes as real installed-CLI handoffs, not provider API aliases.
- Keep route, token, and local runtime diagnostics inspectable through `jini route`.
- Reuse durable session context without replaying large transcripts.
- Keep saved work hidden until `status`, `continue`, `open`, `help`, or natural title matching.
- Install from release assets without requiring source builds.
- Preserve customer-value viability: reduce token waste, throttle friction,
  tool-switching cost, completion risk, or unsafe side effects.
- Block regressions with required gates before commit and push.

## UX Contract

Bare `jini` is a task prompt, not a dashboard.

- Keep startup to a minimal task prompt; do not teach a new shell before value.
- Return compact answers or action receipts first, without product ceremony.
- Keep shell output precise: name the answer, action, artifact, route, blocker, or next command.
- Treat stale shell vocabulary as a P0 regression.
- Treat a new freeform request as work to execute, not as a Start/Keep modal.
- Resume saved work only through explicit commands or natural title matching.
- Never show a full current-work overview for a simple factual question.
- Never require users to learn Jini-specific command vocabulary before value.
- no saved-work dashboard on bare startup
- no visible `Switch` startup control
- no `Start/Keep` interruption model
- no Working Draft for obvious file edits
- no verbose Goal/Working-with/status frame for simple questions
- no `Result ready`, `Task Snapshot`, `Saved:`, or `Next: jini ...` shell around
  simple factual questions
- no visible agent-role theater in the free tier
- no hard-coded entity-to-template routing

## Routing And Resource Policy

Token frugality is P0. Jini must spend context only when it improves the result.

Routing requirements:

- Default to the cheapest safe route that can complete the task.
- Use local/offline routes when they meet the task quality bar.
- Escalate to stronger online routes when correctness, codebase scope, or tool access requires it.
- Label provider API routes separately from CLI handoff routes. A route named `codex` or `claude-code` must invoke that CLI or fail closed with setup guidance.
- Preserve enough session state to continue work without replaying stale chat.

Avoiding throttling is P1. Detect configured CLI/provider pressure, choose
viable alternatives, and resume work cleanly.

Power awareness is P1. In powered mode, Jini can choose higher-throughput local
or online routes. In low-battery mode, Jini should avoid wasteful local model
loads unless the user explicitly asks for offline work.

## Market And Learning Guards

Competitor watching is a P0 feature-selection loop through [competitive-release-plan.md](./competitive-release-plan.md), but it does not create active scope by itself.

- Competitor watch packets can nominate next feature candidates and deletion candidates.
- Each candidate must be classified as copy, integrate, watch, reject, or delete.
- No competitor finding becomes active scope unless the decision record changes.

User productivity learning remains P0 only when it improves the CLI wedge:

- learn stable user context, usage, habits, and repeated patterns
- produce fewer repeated prompts, better defaults, and better route choices
- keep learning inspectable and controllable
- avoid hidden surveillance, broad OS scope, or free-tier agent-suite creep

## Dynamic Platform Principles

These principles guide internals without expanding R0: registry-backed routing,
capability-gated scoring, explicit graceful degradation, fail-closed commercial
feature boundaries until entitlement runtime exists, and bounded/inspectable
user-work context learning.

## Tier Boundary

CLI is available now. App surfaces, when shipped, are available to both free
and commercial users. Subscription gates capabilities, not the ability to
install or open Jini.

Free proves the wedge: direct intake/edits, manual route inspection/switching,
compact status/continue/open, configured offline/local support, token-frugal
session reuse, and setup diagnostics. Free excludes: developer-agent fleets,
tester-agent fleets, commercial skills-based OS productivity suite, managed
throttle recovery, governed approvals, team policy, audit, and automation
loops. Commercial value must be materially higher than the free surfaces:
managed route/throttle policy, governed skills/delegation, cross-device and
offline-online continuation, team audit, and automation loops. Commercial UX
must still follow the same rule: simple task in, useful result out.

## Roadmap

P0 now: intent-first CLI parity, task-first startup even with saved work, direct
file edit reliability, simple question direct-answer behavior, route
list/set/auto/status, real downstream CLI handoff and adapter support waves,
release-asset install, and CLI UX, PRD drift, and scorecard gates in commit gates.

P1 next: throttle-aware route switching, powered-mode and low-battery routing,
offline local-model quality regression harness, cross-surface session handoff,
macOS app HLD/LLD for a Codex desktop-caliber session and artifact surface, and
clearer CLI/local runtime diagnostics.

Deferred until decision-record update: Windows/mobile apps, richer commercial
agent and skills UI, team policy controls, broad proof verticals/templates, and
company automation loops.

## Gates

Every commit touching product, CLI UX, routing, or docs must keep these gates green:

- `go test ./...`
- `bash tools/cli_ux_regression_gate.sh`
- `bash tools/claude_codex_usecase_gate.sh`
- `bash tools/customer_value_gate.sh`
- `bash tools/product_prd_drift_gate.sh`
- `jini scorecard-gate --format json`
- `jini check ship --format json` before push/release

The gates prevent verbose startup, Start/Keep modals, stale docs, and broad PRD drift.
No release ships unless competitor-parity golden transcript gates for Claude,
Codex, ChatGPT, Gemini-style, and 100-prompt Aryan-derived first-minute use cases are green.
No commit ships unless the customer-value gate can still prove the solution is
useful, route-backed, and non-amateur rather than a generic workflow shell.
