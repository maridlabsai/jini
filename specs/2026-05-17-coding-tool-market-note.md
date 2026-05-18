# 2026-05-17 Coding Tool Market Note

## Purpose

This note captures product and routing lessons from a user-shared discussion
about Claude Code, GitHub Copilot CLI/Desktop, multi-model coding workflows,
cost sensitivity, and long-lived coding sessions.

It is not a benchmark report. It is a product interpretation note for Jini.

## Raw Signals

The conversation points to six recurring themes:

1. Claude Code and GitHub Copilot with Claude are now close enough on many deep
   coding tasks that casual users often do not see a major quality gap.
2. Multi-model availability matters, but mostly as a system capability rather
   than a daily user choice.
3. Cost limits and token ceilings matter materially for real usage.
4. Many users want to stay with one stable agent/session to avoid cognitive
   overload.
5. Long-lived context continuity matters as projects become large.
6. Fast update velocity and tight user-feedback loops create real product pull.

## Product Interpretation

### 1. "Best model" is not the real user goal

The real goal is:

- best useful outcome
- at acceptable cost
- without thinking about models too much
- without losing continuity

That means Jini should optimize for:

- outcome quality
- cost discipline
- continuity
- low cognitive overhead

Not for:

- "always use the strongest model"
- "always expose the user to model choice"

### 2. Multi-model is product infrastructure, not the front door

Users do not want to keep switching among Claude, Codex, Copilot-style routes,
and other options on every task.

They want:

- one front door
- one thread
- one place where the work lives
- one system making good choices underneath

Jini should therefore keep:

- `jini` as the front door
- `Use Auto` as the default recommendation
- route/model overrides as secondary controls

### 3. Cost-sensitive routing is a real differentiator

The chat makes explicit what the product already suspected:

- many users are cost-sensitive
- some users optimize around token ceilings and usage limits
- "good enough, much cheaper" often wins on routine work

This validates Jini's policy:

- cheapest suitable by default
- stronger route only when the work clearly justifies it

### 4. Continuity beats route purity

Users care that the project stays coherent as it grows.

Jini should preserve:

- one durable work thread
- one visible state model
- one artifact shelf

Even when:

- the backend tool changes
- the provider changes
- the model changes

The route may change. The work surface should not.

### 5. Update velocity matters, but should not create user thrash

The chat highlights rapid model, tool, and product evolution. Jini should
benefit from that motion without forcing users to track it manually.

That means:

- fast capability registry updates
- fast heuristic updates
- stable user-facing interaction

Jini should absorb churn so the user does not have to.

### 6. Public leaderboards influence perception, but are not enough

Users do look at leaderboards. They also understand that:

- task fit matters
- speed matters
- context matters
- cost matters
- these systems change quickly

Jini should use public rankings as one weak signal, not as the decision engine.

## Product Decisions For Jini

These are the main product calls reinforced by the chat:

1. Keep `auto` as the default and strongest recommendation.
2. Preserve one stable work thread even when route/model changes internally.
3. Treat model/tool selection as Jini's job, not the user's default burden.
4. Keep strict overrides available for users who want certainty.
5. Prefer route/model continuity unless there is a clear cost or quality reason
   to change.
6. Track user overrides as high-value heuristic feedback.
7. Factor rate limits, cost headroom, and continuity into route scoring.

## Routing Implications

The route scorer should explicitly value:

- task fit
- cost
- latency
- context continuity
- quota headroom
- recent cohort acceptance
- recent route-specific acceptance
- route-switch cost

The new important addition is **route-switch cost**.

Jini should not switch tools/models aggressively when the quality delta is
small and the continuity penalty is real.

## Recommended Heuristic Additions

Add these scoring factors:

### Continuity bias

- reward staying on the current route when the route remains suitable
- penalize unnecessary route churn

### Rate-limit and quota bias

- prefer routes with more practical headroom for the current user/task class
- especially on long coding sessions or iterative review loops

### Model-switch cognitive cost

- prefer hiding model changes unless the change has clear value
- if the route changes, preserve the same work thread and explain the reason

### Override-learning signal

- if the user repeatedly overrides `auto` for a task family, treat that as a
  routing signal, not as noise

## What Jini Should Not Do

- make users babysit daily model selection
- switch routes constantly because a leaderboard moved
- optimize for raw model prestige over practical outcome cost
- fragment a single growing project across multiple visible agent threads by
  default

## Summary

The main lesson is simple:

Users do not want the burden of multi-model choice. They want the benefit of
multi-model choice.

Jini should therefore behave like:

- one stable work surface
- one calm thread
- one auto-routing brain
- one explicit reason when it chooses differently

That is more important than exposing the full tool matrix on the main path.
