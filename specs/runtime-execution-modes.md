# Jini Runtime Modes

## 1. Purpose

This document defines how Jini executes local workflow actions in:

- `supervised` mode
- `autonomous` mode

The goal is to increase execution capability without sacrificing traceability,
approval boundaries, or rate-limit discipline.

## 2. Modes

### `supervised`

Use when:

- a human wants to inspect outputs before state changes
- first-time runtime consent has not been granted yet
- the work is still shaping execution surfaces rather than progressing lifecycle
  state

Behavior:

- local exports may run if `write` consent exists
- publish plans may be staged if `publish` consent exists
- state transitions are planned but not auto-executed

### `autonomous`

Use when:

- the workflow is sufficiently deterministic
- first-time runtime consent exists for the relevant action categories
- advancing one legal state is desirable

Behavior:

- local exports may run if `write` consent exists
- publish plans may be staged if `publish` consent exists
- one legal linear state transition may run if `command` consent exists and all
  guard conditions pass

Autonomous mode MUST stop when:

- a required consent category is missing
- a guarded state transition is blocked
- a step would require human-authored semantic input not present in canonical
  artifacts

## 3. Consent Categories

Jini persists first-time consent by action category.

Categories:

- `write`
- `command`
- `publish`

### `write`

Allows deterministic local file outputs such as:

- task board rendering
- task sync export
- issue bundle export
- wiki bundle export

### `command`

Allows deterministic workflow progression commands such as:

- `advance-pack`

This does not waive lifecycle guards. It only permits Jini to attempt the
command.

### `publish`

Allows staging external-system publish plans such as:

- Jira publish bundles
- Confluence publish bundles

This does not allow uncontrolled API bursting. Serialized publish rules remain
binding.

## 4. Consent Persistence

Jini stores runtime consent at:

- `runtime/consent.json`

Jini stores the most recent run report at:

- `runtime/last-run.json`

This creates a durable local execution record without introducing a second
source of truth for the WorkUnit itself.

## 5. Current `run-pack` Behavior

`run-pack` is the runtime orchestrator over existing Jini commands.

It currently:

- loads the current execution recommendation
- applies and persists any newly granted consent categories
- executes deterministic local exports when allowed
- stages serialized publish plans when allowed
- advances one legal state in `autonomous` mode when allowed
- writes a run report under `runtime/`

It does not yet:

- publish live to Jira or Confluence
- auto-author Evidence or Approval semantics
- perform multi-step autonomous execution across multiple guarded states

## 6. Safety Rules

Jini runtime modes MUST NOT:

- bypass required evidence
- bypass required approval
- skip guarded transitions
- silently escalate execution class
- burst publish actions in parallel
- replace missing human input with invented semantic content
