# Adaptive Response Rendering Framework Gate

Updated: 2026-05-19

## Purpose

This gate defines what must remain true as Jini evolves from fixed CLI response
formats toward adaptive response rendering.

## Gate Categories

### 1. Product Contract

Required:

- one work-thread model remains canonical
- one artifact model remains canonical
- useful-result-first remains intact
- remembered work is never switched silently
- route and trust facts stay inspectable

Reject if:

- a surface invents its own work model
- a renderer becomes the source of truth
- normal use requires pack paths, artifact stems, or internal workflow terms

### 2. Adaptive Rendering

Required:

- same canonical state produces the same semantic envelope
- render policy changes emphasis, not truth
- response mode is derived from semantic state
- greeting-only input does not create work
- first useful artifact appears before broad status when work is requested
- continuation focuses on changed artifacts and next step
- blocked states show one active ask and why it matters

Reject if:

- every turn repeats the same full summary frame
- continuity is synthesized from prose when turn deltas exist
- a one-word social input creates a work artifact
- an underspecified complex request gets a generic artifact when a scoping ask
  would materially improve output

### 3. Surface Fit

Required:

- CLI remains compact and action-oriented
- desktop centers the artifact shelf
- mobile supports review, unblock, approve, and resume
- API returns semantic and artifact envelopes rather than terminal prose

Reject if:

- mobile inherits desktop density
- desktop becomes only a larger terminal transcript
- API consumers must parse CLI text to understand work state

### 4. Testability

Required:

- semantic envelope tests exist before broad renderer migration
- artifact identity, status, missing-state, route facts, and trust facts are
  testable without exact prose locks
- existing CLI regressions remain covered where backwards compatibility matters
- publish readiness includes the adaptive framework docs

Reject if:

- tests only compare full prose output
- a renderer can omit blockers, safety state, or route facts without failing
- a profile can bypass artifact or work-thread semantics

### 5. No Core Hard Coding

Required:

- use-case profiles configure artifact families, scoping rules, and examples
- core renderers operate on generic envelopes and modes
- travel, meeting, research, and general work reuse the same rendering contract

Reject if:

- core rendering branches grow one hard-coded format per use case
- new use cases require new prose templates in the main runtime loop
- profile-specific copy becomes the product contract

### 6. Public Boundary

Required:

- public docs describe the base framework and free/local/BYO boundaries
- commercial docs, pricing strategy, private go-to-market notes, and internal
  business reviews stay in the commercial repo

Reject if:

- dated internal strategy notes enter `specs/20xx-*.md`
- commercial pricing or private business plans are committed to the public repo

## Required Regression Inputs

Renderer migration must cover these input classes:

- greeting-only input
- vague one-line input
- already-scoped request
- underspecified complex request
- continuation request
- return visit with current work
- blocked request
- approval or send/publish request

## Required Evidence

Before a renderer migration is considered complete, the implementation must
provide:

- semantic envelope tests
- artifact envelope tests
- CLI renderer tests
- at least one cross-surface contract test or fixture
- publish-readiness pass
- public boundary check pass

## Final Gate

Adaptive rendering passes only when users can feel a more natural product while
the repo still proves the same work truth underneath.
