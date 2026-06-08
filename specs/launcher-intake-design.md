# Launcher Intake Dev Design

Updated: 2026-06-08

This is the active dev design for the CLI front door. It is subordinate to
[number-one-platform-prd.md](./number-one-platform-prd.md) and
[product-settling-decisions.md](./product-settling-decisions.md).

## Goal

Make `jini` feel like a familiar agent CLI, not a new workflow.

The user should be able to:

1. run `jini`
2. describe the task
3. get a direct answer, direct edit, route handoff, or one short question

## Invariants

- Bare `jini` renders the same task prompt with or without saved work.
- Startup is not a saved-work dashboard.
- Saved work is passive context until the user asks for it.
- Typing a saved work title resumes that thread naturally.
- `status`, `continue`, `open`, and `help` are explicit inspection paths.
- Slash commands are rejected instead of becoming work.
- Greetings, acknowledgements, and help questions do not create work.
- Clear local file edits execute instead of creating a draft artifact.
- Simple factual questions answer directly without a work-state dump.

## Forbidden Front-Door UX

- no compact resume card on startup
- no visible `Switch` startup control
- no `Start/Keep` modal
- no Working Draft for obvious file edits
- no Goal/Working-with/status frame for simple questions
- no agent tree or skills OS surface in the free CLI

## Implementation Surfaces

Primary code:

- `internal/app/app.go`
- `runLauncher`
- `runNoCurrentLauncher`
- `handleCurrentWorkAction`
- `maybeHandleLocalTextFileEditIntent`
- `maybeHandleSimpleAnswer`
- `resolveActiveWorkSelection`

Regression tests:

- `TestLauncherStartsAsCompactShellWhenCurrentWorkExists`
- `TestCurrentWorkInteractiveLauncherIsCompactByDefault`
- `TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist`
- `TestInteractiveLauncherCanResumeNamedActiveProject`
- `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting`
- `TestCurrentWorkSimpleFactualQuestionAnswersDirectly`

Required gate:

```bash
bash tools/cli_ux_regression_gate.sh
```

## Change Rule

Do not add a new front-door interaction pattern from implementation alone.

If a change needs a new startup concept, modal, command category, saved-work
surface, or route behavior, update `product-settling-decisions.md` first in the
same commit. Otherwise, treat it as drift and reject it.
