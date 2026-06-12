# Jini Claude Code Guide

## Project Identity

Jini is a public, free-tier-first AI CLI and app framework written in Go. It should feel familiar to Claude Code, Codex, ChatGPT, and Gemini users: compact answers by default, direct action when safe, explicit approval for side effects, and no custom conversational ceremony.

Commercial-only productivity-suite work belongs in `../jini-commercial`, not this repo. Claude Code agents and skills in this repo are development helpers only; they are not free-tier runtime features.

## Canonical Files

- Product contract: `specs/number-one-platform-prd.md`
- PRD trace: `specs/prd-implementation-trace.md`
- CLI behavior contract: `specs/product-rewrite-contract.md`
- Runtime routing: `specs/execution-routing-policy.md`, `specs/runtime-execution-modes.md`, `specs/runtime-selection-heuristics.md`
- Offline/local model policy: `specs/platform-offline-strategy.md`, `specs/local-slm-frontline-policy.md`, `specs/local-model-support-matrix.md`
- Competitive bar: `specs/golden-competitive-benchmark.yaml`, `specs/competitive-release-plan.md`
- Gates: `specs/engineering-gate-matrix.md`, `tools/run_required_gates.sh`

## Non-Negotiables

- Keep shell output precise and Claude/Codex-like. Simple questions should get simple answers.
- Do not reintroduce Start/Keep, work snapshots, generic drafts, or verbose safety blocks for trivial prompts.
- Treat token frugality as P0. Avoid unnecessary file reads, generated artifacts, and long responses.
- Prefer the shared route engine for CLI and app behavior. Do not hard-code prompt classes as one-off fixes.
- Preserve the free/commercial boundary. Free tier may route configured tools and local models; commercial agent/skills OS productivity features stay private.
- Before claiming release readiness, run the required gates and inspect the changed UX path directly.

## Development Workflow

1. Read the smallest canonical spec set for the touched area.
2. Inspect the implementation and tests before editing.
3. Make focused changes; avoid PRD drift and broad rewrites unless explicitly requested.
4. Add or update tests for behavior changes.
5. Run `tools/run_required_gates.sh commit` from the repo root before commit.
6. Commit only after the worktree diff has been reviewed.

## Claude Code Agents

Use project agents under `.claude/agents/` when the work benefits from a focused second pass:

- `jini-cli-quality-reviewer`: CLI transcript quality, first-minute UX, simple answers, file-edit behavior.
- `jini-route-runtime-reviewer`: offline/online routing, local model fallback, provider/tool selection.
- `jini-prd-drift-reviewer`: PRD alignment, stale requirement removal, competitor-scorecard consistency.
- `jini-release-gate-reviewer`: install, signing/notarization readiness, docs, website, public release assets.

## Claude Code Skills

Use project skills under `.claude/skills/` for repeatable work:

- `debugger`: diagnose a failing Jini path with the smallest useful reproduction.
- `reviewer`: adversarial review of current work before continuing.
- `research`: frame missing facts and evidence needed for a decision.
- `prd-drift`: compare implementation or docs against canonical PRD and remove stale requirements.
