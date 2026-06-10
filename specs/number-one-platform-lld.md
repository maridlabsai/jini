# Number One Platform LLD

Updated: 2026-06-10

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
capital of france`. Unknown standalone questions must route through a configured
CLI, provider, or local model when one is available; otherwise they return
compact setup guidance. They still must avoid artifacts and saved work.

### Action Contract

- Local file edits modify the matching file or fail closed with candidates.
- CLI handoff routes execute the installed downstream CLI or return setup
  guidance; they must not silently fall back to provider API aliases.
- Provider and local/offline routes are labeled separately from CLI handoff
  routes.
- Route selection reads from registered adapters, provider readiness, local
  runtime probes, installed CLI checks, capability tags, and recent health
  evidence. It must not infer a hard-coded workflow template from an entity
  name alone.
- Missing routes and unavailable features degrade to the next safe configured
  path or fail closed with setup guidance. Commercial-only feature names fail
  closed in the public CLI until a real entitlement runtime exists.
- Side-effect receipts name files changed, commands/tests run, blockers, and
  recovery path when relevant.

### Side-Effect Approval Matrix

| Action class | Approval rule | Receipt rule |
| --- | --- | --- |
| Read-only inspection, route diagnosis, simple answers | No approval needed | No durable work unless explicitly requested |
| Explicit unambiguous local file edit | No extra approval needed | Name changed file and edit made |
| Ambiguous, multi-file, generated, or risky local edit | Ask first or fail closed | Name candidates, chosen scope, and recovery path |
| Commands that install, mutate, use network, or may be slow/expensive | Ask first unless already covered by an explicit user request | Name command and outcome |
| External send/share/book/pay actions | Require visible approval at the action boundary | Name destination and confirm nothing else was sent |
| Commit, push, release, deploy, destructive change, or credential/policy update | Require visible approval and recovery path | Name irreversible effect, owner, and rollback or next step |
| Paid, managed, or team automation | Fail closed in the public CLI; when implemented, check entitlement before starting | Show free/manual fallback or fail closed |

Approval is not product ceremony. It is a narrow safety gate before effects
that leave the local reversible work path.

### Persistence Contract

- Do not create `current-work.json` for simple factual questions.
- Do not create artifacts for bare entities, greetings, acknowledgements, help
  questions, or unknown standalone questions.
- Create durable work only after task intent needs continuation, artifacts, or
  route receipts.
- Persist learned user/work context only when it improves repeated CLI work,
  route choice, or continuation. Keep it inspectable and avoid broad hidden
  surveillance.

## Implementation Map

| Contract | Code | Gate |
| --- | --- | --- |
| Minimal shell and saved-work passivity | `internal/app/app.go` | `TestLauncherStartsAsCompactShellWhenCurrentWorkExists`, CLI UX gate |
| Simple-answer bypass | `internal/app/simple_answers.go` | `TestInteractiveTypoCapitalQuestionAnswersDirectlyWithoutArtifactShell`, `TestCurrentWorkTypoCapitalQuestionAnswersDirectly` |
| Direct file edit | `internal/app/local_text_edit.go` | `TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting` |
| CLI handoff | `internal/app/cli_handoff.go`, provider decision code | Claude/Codex use-case gate |
| Release quality | `tools/run_required_gates.sh` | commit, push, and release gates |
| Engineering cut control | coordinator-owned work split | `agentic-development-operating-model.md` evidence |

## Change Rule

If a code change needs output not described here, update the PRD, HLD, LLD, and
tests in the same commit. If the transcript would surprise a Claude Code or
Codex user in the first minute, do not ship it.

If a non-trivial cut cannot name disjoint write scopes, integration ownership,
focused checks, and independent review evidence, do not treat it as complete.
Do not add CLI commands or transcript chrome to expose that internal control.

If the required change crosses shell, starter artifact creation, persistence,
and route execution just to answer a simple question or perform a direct file
edit, treat that as a rewrite-trigger candidate under
[product-streamline-redline.md](./product-streamline-redline.md).
