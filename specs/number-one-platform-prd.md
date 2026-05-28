# Number One Platform PRD

Updated: 2026-05-27

## Product Decision

Jini should be built as a self-learning and self-correcting work operating
system.

It should win through:

- faster first useful result
- lower interruption cost
- better upstream quality
- lower total cost to successful outcome
- stronger cross-surface continuity

It should not try to win by matching every competitor feature release one by
one.

## Mission

Jini exists to learn the user's environment, context, and work patterns over
time so it can:

- automate repetitive generic work
- move quality upstream
- stay frugal by default
- preserve one durable work session across CLI and apps
- reach the user's desired outcome faster than competing frameworks

## Category Definition

Jini is not a chatbot, not just a coding agent, and not just an installer.

Jini is:

- a session system
- an artifact system
- a workflow-learning system
- a self-correcting orchestration system
- a cross-surface review and approval system

## User Problem

Current AI tooling makes users pay repeatedly through:

- restating context
- switching tools
- fixing preventable quality misses
- manually wiring repetitive workflows
- rechecking trust and provenance every time
- relearning each surface separately

The product must remove those costs.

## Core Product Thesis

If Jini owns the session graph, artifact graph, quality graph, and route policy
graph, it can improve user outcomes without forcing users to constantly adapt
to new product surfaces.

That means the framework becomes better primarily through:

- data and policy improvement
- learned workflow compression
- stronger self-correction loops
- broader environment understanding

instead of only through visible feature shipping.

## Product Requirements

### 1. One Session Graph

Jini must maintain one canonical session graph that spans:

- goal
- current focus
- artifacts
- blockers
- unresolved tasks
- approvals
- route history
- quality state
- environment state

CLI, desktop, and mobile must all read and act on the same object.

### 2. Environment Learning

Jini must learn stable features of the user's working environment:

- repo family
- build and test patterns
- deployment and review conventions
- preferred tools
- approval posture
- recurring failure modes

This learning must reduce future user input and improve automation decisions.

### 3. Workflow Learning

Jini must detect repeated patterns such as:

- notes to follow-up
- plan to readiness
- issue to implementation
- PR to review to merge
- release preparation
- incident triage

Repeated flows should become reusable, inspectable automations with clear
rollback and override rules.

### 4. Self-Correction Engine

Jini must correct itself when outcome quality or product contract drifts.

Required self-correction loops:

- docs/help/runtime parity repair
- route-policy regret detection
- repeated clarification detection
- stale-memory and stale-evidence detection
- benchmark regression detection
- cross-surface continuity failure detection
- repeated workflow failure-pattern detection

### 5. Upstream Quality Automation

Jini must move quality upstream through automatic checks before users pay for
bad output.

Examples:

- missing evidence before readiness claims
- missing reviewers before handoff
- stale assumptions before continuation
- invalid environment before task execution
- low-confidence route selection before expensive escalation

### 6. CLI Contract

The CLI must remain:

- the fastest front door
- the cheapest control surface
- the most stable recovery surface

The CLI must:

- accept natural intake
- return useful results before status
- expose one continuation model
- stay compact under token pressure
- keep cost, route, and next action legible

### 7. Desktop And Mobile Apps

The apps should not duplicate the CLI.

They should specialize in:

- artifact inspection
- artifact editing
- approvals
- review
- long-running work supervision
- session switching
- interruption recovery
- push-based continuation

### 8. GitHub-Native System Of Record

For engineering workflows, Jini must become first-class with GitHub:

- issues
- pull requests
- reviews
- checks
- runbooks
- release state

The goal is to beat raw shell plus `gh` by preserving context, trust, and
repeatability end to end.

### 9. Frugal Route Policy

Jini must choose the cheapest adequate route by default and make that choice
visible.

Premium or deeper routes should be used only when justified by:

- quality delta
- time savings
- safety gain
- context size
- automation payoff

### 10. Trust And Governance

Jini must preserve trust by making decisions inspectable:

- why a route was chosen
- what memory was used
- what evidence is missing
- what approvals are needed
- what action was taken
- how to undo or override it

## Non Goals

Jini should not:

- become a feature clone of every fast-moving agent shell
- rely on users writing large instruction files to get value
- add surface-specific dialects for CLI, desktop, and mobile
- ship many demo flows before the flagship flows are dominant
- hide cost, route, or quality tradeoffs

## Success Metrics

Primary metrics:

- time to first useful result
- cost per successful task
- interruption recovery time
- cross-surface resume success rate
- repeated workflow automation rate
- upstream defect catch rate
- clarification-turn reduction
- route regret rate

## Score Exit Criteria

The product does not declare a win until:

- `delivery-maturity >= 9.0`
- `memory-reliability >= 9.0`
- `adapter-portability >= 9.0`
- `token-efficiency >= 9.0`
- overall score margin over the strongest competitor is at least `0.8`
- flagship flows beat direct competitor workflows in benchmark evidence

## Release Philosophy

Jini should keep a slower user-facing release cadence than the fastest
competitors by pushing more improvement into the learning and policy layer.

Target model:

- daily trace capture and evaluation
- weekly shadow-policy trials
- weekly policy-bundle promotion when score-positive
- monthly product-surface releases
- quarterly architecture changes only when benchmark and score evidence justify them

Users should experience steadier behavior, fewer surface churn events, and
better outcomes over time.

## Risks

- overfitting learning to narrow workflows
- hidden automation reducing trust
- policy drift outrunning evaluation discipline
- cross-surface complexity growing faster than session coherence
- GitHub-native depth outpacing non-engineering applicability

## PRD Decision

Jini should be built as the self-learning and self-correcting work operating
system that wins on outcome speed, frugality, and continuity rather than on
feature-count parity.
