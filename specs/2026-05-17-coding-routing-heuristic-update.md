# 2026-05-17 Coding Routing Heuristic Update

## Purpose

This spec updates Jini's runtime selection logic for coding-oriented work based
on recent market/user signals:

- strong model parity across several coding surfaces
- user desire for one stable agent/session
- cost sensitivity
- high practical value of rate-limit headroom

It should be read together with:

- [2026-05-17-coding-tool-market-note.md](./2026-05-17-coding-tool-market-note.md)
- [runtime-selection-heuristics.md](./runtime-selection-heuristics.md)
- [research-informed-heuristics.md](./research-informed-heuristics.md)

## Problem

The current route scorer is good at:

- task class
- depth
- modality
- cohort memory
- local-vs-remote tradeoffs

But for coding work it still underweights:

- continuity of a long-lived coding session
- practical quota headroom
- cost-aware iteration loops
- user resistance to model switching

That can produce correct-but-annoying routing.

## New Policy

For coding-oriented work, Jini should optimize:

1. acceptable quality
2. continuity
3. cost
4. rate-limit headroom
5. latency

Not:

1. strongest possible model first
2. leaderboard prestige first

## New Scoring Terms

The coding route scorer should add four terms.

### 1. Continuity score

Reward staying on the current route when:

- the existing route is still suitable
- the expected quality delta is small
- the task is part of an ongoing coding thread

Penalty when:

- Jini switches route without a meaningful gain in quality, cost, or capability

Intent:

- reduce route churn
- keep one stable coding thread

### 2. Quota headroom score

Reward routes that have more practical iteration capacity for the user.

Examples:

- higher session limits
- cheaper repeat turns
- better support for long back-and-forth code loops

Intent:

- optimize for real project completion, not only per-turn quality

### 3. Route-switch cost

Treat route changes as a cost when:

- the work is already active
- the context is large
- the user is in an iterative debug/review/fix loop

The more continuity matters, the higher this cost should be.

Intent:

- stop unnecessary tool/model hopping

### 4. Override-learning bias

If the user repeatedly overrides Jini for a coding cohort, treat that as a
real signal.

Examples:

- repeatedly forcing Claude-style route for architecture review
- repeatedly forcing cheaper coding route for routine edits

Intent:

- let the system adapt to real user preference and workflow economics

## Updated Coding Route Rule

For coding tasks:

1. use the cheapest suitable route by default
2. if a local route is good enough, prefer local
3. if current route is still suitable, prefer continuity
4. only switch route when one of these is true:
   - the quality gap is material
   - the current route is too expensive for the expected loop
   - the current route is hitting quota/limit pressure
   - the current route lacks a needed capability
5. preserve one visible thread regardless of route change

## Updated Model Rule

For coding work, model selection should prefer:

- practical project completion
- consistency across turns
- cost-aware iteration

This means Jini should not eagerly jump to a "better" model if:

- current model is good enough
- switching would add user confusion
- switching would fragment learned route memory

## Updated Explanation Rule

When Jini changes route/model on an existing coding thread, it should explain
the reason using one of these patterns:

- quality gain justified the change
- cost/limit pressure justified the change
- a missing capability justified the change

It should avoid vague explanations like:

- "better model"
- "stronger route"

Without specifics.

## New Signals To Persist

Persist for coding work:

- prior route on this thread
- route-switch count on this thread
- route-switch reason
- user override count by coding cohort
- user override direction by coding cohort

These should influence later auto routing.

## Suggested Future Inputs

When available, add:

- practical remaining quota estimate
- average cost per accepted artifact for the route
- average acceptance after long coding loops
- context carry efficiency

These are more valuable than static model rankings.

## Guardrails

For coding work, Jini should fail review if:

- it switches route repeatedly with no clear explanation
- it prefers a stronger route where a cheaper route is already accepted and
  stable
- it ignores repeated user overrides in the same cohort
- it fragments one coding project across multiple visible work threads by
  default

## Acceptance Criteria

The implementation should make these true:

1. Coding route selection explicitly includes continuity bias.
2. Coding route selection explicitly includes route-switch cost.
3. Coding route selection explicitly includes quota/headroom bias.
4. Repeated user overrides on coding cohorts influence later auto routing.
5. Existing coding work remains one visible thread even when the backend route
   changes.
6. Route-change explanations are concrete and user-comprehensible.

## Summary

For coding work, Jini should act like a calm engineering lead:

- keep the thread stable
- spend carefully
- switch only when it matters
- remember what the user keeps preferring

That is the right response to a market where raw model gaps are narrowing but
workflow economics still matter a lot.
