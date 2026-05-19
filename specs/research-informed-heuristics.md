# Research-Informed Heuristics

Updated: 2026-05-17

## Purpose

This document translates useful research and product patterns into practical
runtime rules for Jini.

It is intentionally selective.

Jini should not absorb popular agent patterns wholesale. It should only adopt
what improves:

- outcome quality
- user trust
- cost efficiency
- clarity of progress and next step

## Inputs

This guidance is based on:

- ReAct: [arXiv:2210.03629](https://arxiv.org/abs/2210.03629)
- Self-Refine: [arXiv:2303.17651](https://arxiv.org/abs/2303.17651)
- Reflexion: [arXiv:2303.11366](https://arxiv.org/abs/2303.11366)
- Plan-and-Solve Prompting: [arXiv:2305.04091](https://arxiv.org/abs/2305.04091)
- Self-Consistency: [arXiv:2203.11171](https://arxiv.org/abs/2203.11171)
- Toolformer: [arXiv:2302.04761](https://arxiv.org/abs/2302.04761)
- Anthropic’s public product framing for
  [Claude Cowork](https://www.anthropic.com/product/claude-cowork),
  [Projects](https://www.anthropic.com/news/projects),
  [Artifacts](https://www.anthropic.com/news/artifacts), and
  [the Chat -> Code -> Cowork evolution](https://www.anthropic.com/webinars/future-of-ai-at-work-introducing-cowork)

The two LinkedIn posts the user referenced were not reliably retrievable as
full text during implementation. The official Anthropic Cowork and workflow
materials above expose the same underlying pattern clearly enough to drive
product decisions.

## Product Interpretation

### 1. Outcome-first delegation beats prompt-first chat

The strongest product lesson from Claude Cowork is not "add another mode."
It is:

- users should start from a goal
- Jini should turn that goal into work
- the product should keep the user close to the outcome, not the prompt

For Jini, that means:

- keep `jini` as the one front door
- accept rough goals without asking for tooling choices first
- show `Goal`, `Working with`, `Doing now`, `Done`, `Need`, and `Next`
- produce named artifacts early

Jini should not copy Cowork's surface literally. It should borrow the
goal-to-deliverable framing and keep the CLI calmer than a mode-heavy system.

### 2. Planning should be default for multi-step work, not a special ritual

Plan-and-Solve shows a practical pattern:

- first divide the task into smaller steps
- then carry them out

This belongs in Jini's hidden runtime loop for:

- research synthesis
- plan/spec readiness
- multi-source drafting
- itinerary or workflow building

This does not mean exposing a visible "planner mode" by default.

Product rule:

- if the request is clearly multi-step, Jini should silently create a short
  internal plan before execution
- the user sees the effect through `Doing now`, `Done`, and `Next`
- the user does not need to manage the plan object unless they ask

### 3. ReAct belongs in tool orchestration, not in visible UX

ReAct is useful because it interleaves reasoning and action.

That maps cleanly to:

- deciding when to read an attachment
- deciding when to open an artifact
- deciding when to invoke a connector
- deciding when to pause and ask for missing information

Jini should use the ReAct idea internally, but it should not expose
`Thought / Act / Observe` loops to normal users.

Product rule:

- use action-aware reasoning to choose the next tool or connector step
- keep the visible surface as progress and outcomes, not agent traces

### 4. Toolformer supports connector-aware routing and observation

Toolformer is relevant because it shows that tool usage can be learned as a
decision problem:

- whether to call a tool
- when to call it
- how to incorporate the result

For Jini, this validates:

- connector-aware route scoring
- passive observation of export/open/share/use outcomes
- cohort-specific learning from external workflow signals

Product rule:

- treat connectors as part of the route scorer, not only as integrations
- keep collecting passive and explicit evidence about whether a route really
  helped downstream work

### 5. Self-Refine should be selective, not always-on

Self-Refine improves outputs by adding critique and revision loops.

That is useful for Jini only when:

- the task is deep or high-risk
- the cohort has weak confidence
- the first draft failed acceptance or adoption often enough to justify extra work

It should not be the default path for every task.

Product rule:

- default path stays single-pass plus light structure
- add one focused refine pass when the route scorer predicts material benefit
- cap refinement depth tightly for cost control

### 6. Reflexion should be memory, not theater

Reflexion shows the value of textual feedback improving later attempts.

Jini already benefits from this pattern through:

- cohort memory
- model upvote/downvote
- graded acceptance
- passive edit-distance and semantic rewrite signals
- downstream adoption signals

Product rule:

- preserve outcome memory at the cohort and adapter level
- use it to bias later route choices
- do not turn every session into a verbose self-critique transcript

### 7. Self-Consistency should be reserved for expensive decisions

Self-Consistency can improve difficult reasoning by sampling multiple paths and
choosing the most consistent answer, but it also raises cost.

For Jini, that should be:

- off by default
- enabled only for high-value or high-risk work

Candidate triggers:

- architecture choices
- benchmark claims
- release-readiness judgments
- conflicting evidence across sources
- low confidence on a deep request

Product rule:

- use selective multi-sample verification only when the expected value of a
  better answer is worth the cost
- never pay this tax for normal drafting or everyday follow-up work
- for multimodal work, judge the stronger draft on evidence extraction quality,
  ambiguity visibility, and recommended verification steps, not only generic
  section structure
- make that multimodal rubric subtype-aware, so PDF/scan, screenshot/image,
  and audio/transcript requests can reward the right evidence language
- apply the same subtype split to route choice, so transcript-heavy work can
  stay on cheaper text-capable routes while image/document evidence can justify
  multimodal routing
- apply the same subtype split to local empirical benchmark memory, so one
  strong screenshot run does not over-credit PDF extraction or audio work

## Heuristic Upgrades Required

The runtime scorer should explicitly factor:

1. task family
2. artifact family
3. modality
4. risk level
5. confidence in the candidate route
6. local adapter empirical performance
7. cohort acceptance and downstream adoption
8. cost budget
9. need for selective refinement
10. need for selective multi-sample verification

This yields three policy layers:

- execution policy
- verification policy
- learning policy

### Execution policy

- local-first if quality is good enough
- cheapest suitable route by default
- hidden plan-first for multi-step requests
- connector/tool use only when it advances the artifact

### Verification policy

- no extra verification for low-risk everyday work
- one refine pass for medium-risk or weak-confidence cohorts
- consistency check or stronger route only for high-risk decisions

### Learning policy

- remember route quality by cohort, profile, and connector context
- remember whether outputs were accepted, edited lightly, replaced, shared,
  exported, or observed externally
- decay stale evidence
- promote recovered routes faster when new evidence is strong

## What Jini Should Not Adopt

- default visible chain-of-thought style traces
- always-on multi-agent debate
- always-on self-critique loops
- expensive majority-vote reasoning for everyday work
- mode sprawl in the user-facing command surface

These patterns either raise cost, increase latency, or make the product feel
more process-heavy than outcome-first.

## Concrete Product Decisions

### A. Route selection

Route selection should remain:

- cheapest suitable by default
- capability-aware
- device-aware
- cohort-aware
- connector-aware

For coding-oriented work, route selection should also become:

- continuity-aware
- route-switch-cost-aware
- quota-headroom-aware
- override-learning-aware

### B. Clarification behavior

When a request is multi-step and under-specified:

- Jini should first draft a hidden plan
- infer what is safe
- ask only one blocking question at a time
- show assumptions explicitly

### C. Verification behavior

Verification should be adaptive:

- `low`: single pass
- `medium`: single pass plus structure
- `high`: add one refine pass or a stronger route
- `extra high`: selective multi-sample verification and/or stronger route

In the current implementation, that `extra high` tier is a selective
consistency check with a second independent draft, not an always-on debate or
multi-agent loop.

### D. Surface behavior

The visible shell should still look simple:

- no research-paper jargon
- no visible thought loops
- no agent trace theater

The user should see:

- what Jini chose
- why it chose it
- what it produced
- what is missing
- what happens next

## Gate Implications

Major heuristic changes should now be judged against this document.

A heuristic change fails if it:

- adds cost without measurable acceptance or adoption gain
- adds visible complexity to the normal flow
- adds new route jargon to the beginner surface
- increases verification depth for low-risk routine work
- weakens the local-first cheap-path principle without strong evidence
