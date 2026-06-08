# Number One Development Plan

Updated: 2026-06-08

This is the active implementation plan. It is intentionally narrow.

Authority chain:

1. PRD: [number-one-platform-prd.md](./number-one-platform-prd.md)
2. Decision record: [product-settling-decisions.md](./product-settling-decisions.md)
3. Dev design: [launcher-intake-design.md](./launcher-intake-design.md)
4. Implementation plan: this file

The current competitor-derived release pressure lives in
[competitive-release-plan.md](./competitive-release-plan.md), but competitor
ideas do not become active work until the decision record changes.

## Active Goal

Make the CLI good enough for users to prefer trying Jini again.

This means:

- install succeeds from release assets
- `jini` starts task-first
- local file edits work
- simple questions answer directly
- route inspection and route switching are clear
- configured CLI routes actually hand off to configured CLIs
- saved work helps continuation without becoming a startup dashboard
- tests and gates catch regressions before commit

The core charter is intent-first Claude/Codex parity. If a cut makes Jini feel
less like a familiar coding CLI, or turns questions and file tasks into generic
artifacts, that cut stops until the regression is removed.

## Active Cuts

### Cut 1: Intent-First CLI Parity

Status: active.

Deliver:

- keep bare startup compact with and without saved work
- keep file-edit, simple-question, malformed-question, and bare-entity flows
  direct or fail-closed
- keep explicit vertical choices working without auto-routing bare entities
- remove stale public examples that teach old interaction models
- expand CLI UX regression tests before changing behavior

Exit evidence:

- `bash tools/cli_ux_regression_gate.sh`
- `jini scorecard-gate --format json`
- `go test ./...`
- public `curl | bash` install smoke for release builds

Release blockers:

- simple questions create work, route into templates, or dump saved-work status
- bare entities create `Task Snapshot`, itinerary, or other artifacts
- file/code tasks produce drafts instead of side effects, receipts, or exact
  ambiguity
- startup reintroduces `Start/Keep`, `Switch`, or saved-work dashboards

### Cut 2: Configured CLI Handoff And Token-Frugality Proof

Status: next.

Deliver:

- keep `jini route` distinguishing CLI handoff, provider API, and local/offline routes
- invoke installed Codex, Claude Code, or other configured CLI routes when selected
- preserve fail-closed guidance when a named CLI route is unavailable
- keep provider API routes separately labeled instead of using CLI names
- record route choice compactly so continuation does not replay stale context

Exit evidence:

- route tests cover CLI handoff success, missing CLI setup, reserved-route
  fail-closed behavior, provider API routes, auto, manual, local preview,
  local SLM, Codex, and Claude
- docs show route inspection without teaching a new workflow first
- scorecard pressure vector `token-frugality-p0` remains green

Release blocker:

- `codex`, `claude-code`, or another CLI route is implemented as a provider API
  alias or lacks visible handoff/fail-closed setup guidance

### Cut 3: Saved Work Continuity Without Dashboard

Status: next.

Deliver:

- make `status`, `continue`, and `open` useful after direct task execution
- support natural saved-title resume
- keep sibling work out of startup unless explicitly requested

Exit evidence:

- regression tests prove startup stays compact
- continuation tests prove saved state still helps
- no public docs recommend a saved-work dashboard

### Cut 4: User Feedback Loop

Status: next after the first tester pass.

Deliver:

- capture the top user rejection reasons as testable scenarios
- add only the smallest product change that removes the rejection
- update PRD/design only when a new requirement is explicitly accepted

Exit evidence:

- each accepted user finding maps to one test, one implementation change, and
  one doc or decision update if user-facing

## Paused Work

These are not active implementation work:

- desktop and mobile apps
- broad agent OS surfaces
- free-tier skills or delegation commands
- visible developer/tester agent fleets
- broad vertical demos
- new command grammar
- company automation loops

Paused work can restart only through `product-settling-decisions.md`.

## Implementation Rule

Every cut must stay traceable:

- PRD requirement
- dev design behavior
- implementation surface
- regression test
- gate evidence

If that trace cannot be written in one short paragraph, the cut is too broad.
