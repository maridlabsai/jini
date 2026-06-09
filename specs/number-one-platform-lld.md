# Number One Platform LLD

Updated: 2026-06-08

This low-level design defines the executable contracts that implement
[number-one-platform-hld.md](./number-one-platform-hld.md).

## Runtime Contracts

### Shell Contract

- `runLauncher` and `runNoCurrentLauncher` render only a minimal task prompt on
  bare startup.
- Answer turns print the answer or action receipt first.
- Simple factual questions must not print `Result ready.`, `Task Snapshot`,
  `Saved:`, `Next: jini ...`, `Goal`, `Working with`, or `AI route`.
- Saved work summaries are available only through `status`, `continue`, `open`,
  natural title matching, or explicit work-state questions.

### Intent Contract

Intent handlers run before work creation:

1. utility commands and setup commands
2. safe local file edit intent
3. simple factual answer intent
4. current-work question intent
5. bare-entity clarification
6. route or work creation

The direct-answer classifier owns typo-tolerant small facts such as `whats teh
capital of france`. Unknown standalone questions may answer `I don't know
locally.` but must still avoid artifacts and saved work.

### Action Contract

- Local file edits modify the matching file or fail closed with candidates.
- CLI handoff routes execute the installed downstream CLI or return setup
  guidance; they must not silently fall back to provider API aliases.
- Provider and local/offline routes are labeled separately from CLI handoff
  routes.
- Side-effect receipts name files changed, commands/tests run, blockers, and
  recovery path when relevant.

### Persistence Contract

- Do not create `current-work.json` for simple factual questions.
- Do not create artifacts for bare entities, greetings, acknowledgements, help
  questions, or unknown standalone questions.
- Create durable work only after task intent needs continuation, artifacts, or
  route receipts.

## Implementation Map

| Contract | Code | Gate |
| --- | --- | --- |
| Minimal shell and saved-work passivity | `internal/app/app.go` | `TestLauncherStartsAsCompactShellWhenCurrentWorkExists`, CLI UX gate |
| Simple-answer bypass | `internal/app/simple_answers.go` | `TestInteractiveTypoCapitalQuestionAnswersDirectlyWithoutArtifactShell`, `TestCurrentWorkTypoCapitalQuestionAnswersDirectly` |
| Direct file edit | `internal/app/local_text_edit.go` | `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting` |
| CLI handoff | `internal/app/cli_handoff.go`, provider decision code | Claude/Codex use-case gate |
| Release quality | `tools/run_required_gates.sh` | commit, push, and release gates |

## Change Rule

If a code change needs output not described here, update the PRD, HLD, LLD, and
tests in the same commit. If the transcript would surprise a Claude Code or
Codex user in the first minute, do not ship it.
