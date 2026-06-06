# Lean Platform Doctrine

Updated: 2026-06-05

This document is a specialized doctrine for lean, cost-effective, and
offline-capable platform behavior, not the top-precedence product and
operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this doctrine conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, or automation posture, the canonical PRD wins and
this doctrine should be updated.

Jini should be a lean, efficient, cost-saving platform that teams want to buy because it is easy to learn, easy to govern, and hard to replace.

## Mission

Jini exists to be the most dependable, cost-effective, and frictionless
platform for automating complex AI workflows across any environment.

That mission should be judged literally:

- dependable under interruption, throttling, and device switching
- cost-effective across local, BYO, and hosted routes
- frictionless for expert and non-expert users alike
- platform-wide across CLI, desktop, mobile, and governed environments

## Core Principles

### 1. Lowest Total Cost To Useful Outcome

Jini optimizes for the full cost of getting useful work done:

- model cost
- operator time
- retries
- context rebuilding
- review overhead
- avoidable clarification turns

Success is not a cheap request. Success is a cheap successful outcome.

### 2. One Stable Surface

Jini should present one stable surface:

- one front door
- one work model
- one continuation model
- one command vocabulary
- one mission across any environment

Users should not need to learn separate product dialects for setup, runtime, or review.

### 3. Cheap By Default, Strong When Needed

Jini should stay local or low-cost when quality is good enough, then escalate only when the expected gain is worth the spend.

Every paid escalation should be justified by one of:

- materially better quality
- materially lower time to useful result
- materially lower operator burden
- materially lower failure risk

This policy must apply at two levels:

- the underlying platform/runtime target, such as `claude-code`, `codex`,
  `github-copilot`, or `kiro-cli`
- the model/profile used inside that route, such as `local-fast`,
  `local-workhorse`, `local-deep`, or the managed provider/model selected for
  the task

Jini should be able to switch platform when throttling, quota pressure,
availability drift, or interruption risk makes the current route worse for the
user. Jini should also be able to adjust the model or local profile based on
task shape and complexity instead of pinning one model across all work.

### 4. Visible Efficiency

Jini should make efficiency legible.

Users should be able to see:

- what route was chosen
- why it stayed cheap or escalated
- what state was reused
- what work was avoided
- whether Jini is currently operating in offline mode
- whether online capability is currently available
- what reconciliation debt was accrued while working offline

Efficiency that cannot be explained will not be trusted.

### 5. UX Second To None

Jini should feel easier than the alternatives, not richer than the alternatives.

Great Jini UX means:

- fewer user decisions
- faster first useful result
- clearer next action
- safer handoff
- lower interruption cost
- lower recovery cost

UX polish that adds ceremony, hiding, or ambiguity is a failure.

### 6. Sessions First, Surface Second

Jini should preserve one session across every supported surface:

- macOS
- Windows
- mobile
- CLI

The session is the product object. Each surface is only a view over that same
object.

Users should be able to:

- see current sessions
- resume from any supported surface
- inspect ready and missing state everywhere
- continue without rebuilding context

When Jini is offline, it should still be able to benefit from locally available
context and memory already curated under supported framework surfaces, such as
Claude Code, Codex, or GitHub CLI, once that context has been bound into the
same session model.

### 7. Fewer Product Ideas, Better Execution

Jini should win through clarity and repeatability, not feature count.

The default action for extra ceremony is removal. The default action for duplicate commands is removal. The default action for product-shaped wording is simplification.

## Buying Posture

Jini should be easy to buy because it is:

- easy to explain
- easy to deploy
- easy to govern
- easy to benchmark
- easy to justify against cost and latency

## Operating Rules

### Engineering Discipline

- implementation should follow SOLID by default
- stable domain objects should use explicit OOP boundaries instead of ambient
  mutable state
- prefer composition over inheritance
- use adapter, strategy, factory, facade, and value-object patterns where they
  reduce coupling and repetition
- reject god objects, mixed-responsibility modules, and giant mode-switch
  branches in core workflow code

### Command-Surface Discipline

- taught commands must be canonical
- taught commands should be one word
- multiword taught commands are a failure unless there is no simpler standard term
- compatibility aliases should not be taught, and should be removed when the canonical surface is stable

### Latency Discipline

- first-turn paths should stay fast
- continuation should be cheaper than restarting
- help and status should not dump unnecessary detail
- added structure must justify its response-time cost

### UX Discipline

- every primary screen must answer what happened, what is ready, what is
  missing, and what to do next
- advanced controls should not leak into the default path
- the default path should work for non-expert users without provider jargon
- intelligence should remove decisions, not create them

### Continuity Discipline

- resume must mean the same thing on every surface
- the current session should be visible before the user is asked to start over
- switching devices must not require context reconstruction
- artifacts, route evidence, and ready/missing state must travel with the
  session
- offline mode should be visible to the user instead of being inferred later
- online capability should be checked efficiently and surfaced when it changes
- reconciliation debt from offline work should be tracked and shown before sync
- locally imported context from Claude Code, Codex, or GitHub CLI should remain
  reusable offline when it has already been attached to the session

### Cost Discipline

- local-first when good enough
- premium routing only when justified
- no hidden escalation as the default happy path
- no duplicate shells or duplicate steps when one stable path is enough
- throttle-driven platform switching should be automatic when a better route is
  available
- task-shaped model selection should choose the smallest model or profile that
  can do the job well, then escalate only when the task truly needs more depth

## Measures

Jini should track:

- cost-per-successful-task
- time-to-first-useful-result
- clarification-turn-count
- resume-cost
- cross-surface-resume-success-rate
- recovery-time-after-interruption
- command-surface-count
- premium-route-regret-rate
- throttle-avoided-interruption-rate
- platform-switch-success-rate
- task-to-model-fit accuracy
- offline-mode-transparency-rate
- reconciliation-debt-clearance-rate
- imported-context-reuse-rate
- offline-mode-transparency-rate
- reconciliation-debt-clearance-rate
- imported-context-reuse-rate

## Reject Conditions

Reject changes that:

- increase command-surface-count without a matching reduction elsewhere
- reintroduce compatibility aliases into the taught surface
- add multiword commands when a standard one-word command exists
- increase time-to-first-useful-result without a clear user-outcome gain
- increase premium-route usage without a measurable quality or speed gain
- add workflow ceremony that users must learn before getting value
- add a supported surface that cannot resume the same session model
- add UX complexity that makes route, state, or next action harder to
  understand
- pin users to one managed platform when a cheaper or more available route is
  clearly better under throttle pressure
- choose a stronger or more expensive model by default when task-shaped
  selection could have completed the work well
- hide that Jini is operating offline when route or sync behavior materially
  changed because of it
- lose or silently ignore reconciliation debt accrued while the user was
  productively offline
- hide that Jini is operating offline when route or sync behavior materially
  changed because of it
- lose or silently ignore reconciliation debt accrued while the user was
  productively offline
