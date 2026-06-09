# Number One Platform HLD

Updated: 2026-06-08

This high-level design translates
[number-one-platform-prd.md](./number-one-platform-prd.md) into architecture
boundaries. It is subordinate to the PRD and
[product-settling-decisions.md](./product-settling-decisions.md).

## Release Quality Bar

Jini does not ship an iteration unless the competitor-parity transcript gates
are green for the first-minute use cases users compare against Claude Code,
Codex, ChatGPT, and Gemini CLI:

- simple factual question returns a compact answer
- clear local file edit changes the file or fails closed with exact ambiguity
- configured CLI route invokes the real installed CLI or fails closed
- saved work stays passive until explicitly requested
- route and token diagnostics remain inspectable without startup ceremony

## Architecture Boundaries

Jini is five runtime layers:

- CLI shell: owns prompt rendering, command dispatch, and user-visible receipts.
- Intent boundary: classifies direct answers, file edits, route controls, saved
  work controls, and work creation before any artifact is created.
- Action boundary: performs local file edits, route handoffs, provider/local
  calls, or fail-closed prompts.
- State boundary: persists work state, route receipts, and artifacts only after
  there is real work to preserve.
- Gate boundary: blocks commits, pushes, and releases when golden transcripts,
  PRD drift, scorecard, security, or ship checks fail.

## Request Flow

Every interactive turn follows one path:

1. Parse explicit commands and slash-command errors.
2. Handle safe direct actions and direct answers before current-work rendering.
3. Resolve current-work commands only when the input asks about current work.
4. Route or execute task intent.
5. Create durable work or artifacts only for real work that benefits from reuse.

Simple questions must stop at step 2. They must not enter work creation,
artifact rendering, route ceremony, or saved-work overview.

## State Model

Saved work is passive context. It helps `status`, `continue`, `open`, natural
title matching, and route continuation. It is not the default frame for new or
unrelated input.

The state boundary may persist:

- `current-work.json` for active durable work
- work-thread metadata and artifacts for reusable outputs
- route receipts for handoffs and diagnostics
- dogfood evidence for release validation

The state boundary must not persist simple factual questions or bare-entity
clarifications as work units.

## Non-Goals

The near-term architecture does not include:

- a new conversation grammar
- a visible agent-role tree in the free CLI
- broad desktop/mobile app surfaces
- generic vertical-template routing from entities or questions
- provider API aliases marketed as CLI handoffs
