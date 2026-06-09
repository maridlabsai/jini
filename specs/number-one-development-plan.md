# Number One Development Plan

Updated: 2026-06-09

This is the active implementation plan. It is intentionally narrow.

Authority chain:

1. PRD: [number-one-platform-prd.md](./number-one-platform-prd.md)
2. Decision record: [product-settling-decisions.md](./product-settling-decisions.md)
3. HLD: [number-one-platform-hld.md](./number-one-platform-hld.md)
4. LLD: [number-one-platform-lld.md](./number-one-platform-lld.md)
5. Internal operating model:
   [agentic-development-operating-model.md](./agentic-development-operating-model.md)
6. Dev design: [launcher-intake-design.md](./launcher-intake-design.md)
7. Implementation plan: this file

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
- familiar tools are supported through adapters without adding a new UX curve
- saved work helps continuation without becoming a startup dashboard
- tests and gates catch regressions before commit

The core charter is intent-first Claude/Codex parity. If a cut makes Jini feel
less like a familiar coding CLI, or turns questions and file tasks into generic
artifacts, that cut stops until the regression is removed.

Non-trivial engineering cuts must use the internal sub-agent divide-and-conquer
model. The coordinator owns scope splits, disjoint write sets, integration, and
evidence; users must not see this as a new command surface or role tree.

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
- simple answers print `Result ready`, `Task Snapshot`, `Saved:`, or follow-on
  command chrome instead of a compact answer

### Cut 2: Configured CLI Handoff And Token-Frugality Proof

Status: next.

Deliver:

- keep `jini route` distinguishing CLI handoff, provider API, and local/offline routes
- invoke installed Wave 1 CLIs when selected: Codex, Claude Code, Gemini CLI,
  Aider, or OpenCode
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

### Cut 3: Familiar-Tool Adapter Breadth

Status: next.

Deliver:

- define one subprocess handoff contract for cwd, prompt, stdin/stdout,
  approvals, changed files, exit status, and receipts
- ship Wave 0 foundation: subprocess contract, route receipts, `doctor`
  detection, and fail-closed setup guidance
- ship Wave 1 terminal-agent adapters: Codex, Claude Code, Gemini CLI, Aider,
  and OpenCode
- plan Wave 2 local/gateway routes: Ollama, LM Studio, OpenRouter, and LiteLLM
- plan Wave 3 editor/cloud-agent handoff: Continue, Cline/Roo, Cursor,
  Windsurf, and GitHub Copilot coding agent
- keep adapter setup behind `jini route` and `doctor`; no first-minute taxonomy

Exit evidence:

- each supported wave has adapter smoke tests for available, missing, and
  failed handoff states
- scorecard pressure vector `familiar-tool-adapter-breadth` remains green
- docs list supported, reserved, and planned adapters without overclaiming

Release blocker:

- claiming broad framework support before the relevant wave has smoke tests

### Cut 4: Saved Work Continuity Without Dashboard

Status: next.

Deliver:

- make `status`, `continue`, and `open` useful after direct task execution
- support natural saved-title resume
- keep sibling work out of startup unless explicitly requested

Exit evidence:

- regression tests prove startup stays compact
- continuation tests prove saved state still helps
- no public docs recommend a saved-work dashboard

### Cut 5: User Feedback Loop

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
- operating-model scope split and evidence
- implementation surface
- regression test
- gate evidence

If that trace cannot be written in one short paragraph, the cut is too broad.
