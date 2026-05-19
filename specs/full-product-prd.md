# Jini Full Product PRD

Updated: 2026-05-15

## Purpose

This PRD defines the target product shape for Jini after the current rewrite.

It is intentionally stricter than a vision memo. It makes product calls that
must be judged against the rewrite scorecard, golden benchmark, and dogfood
persona gates already established in this repo.

This PRD exists to answer one question clearly:

- what should Jini look like if it is going to beat the current field on real
  user outcomes, not just on internal architecture quality

The execution sequence for this PRD lives in
[full-product-prd-execution-plan.md](./full-product-prd-execution-plan.md).

The runtime-learning interpretation of this PRD lives in
[research-informed-heuristics.md](./research-informed-heuristics.md).

The public surface and free-tier requirements across CLI, desktop, and mobile
live in
[client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md).

The stable engineering framework for major workstreams lives in
[workstream-technical-framework.md](./workstream-technical-framework.md).

The public framework for curated travel work lives in
[travel-curated-experience-framework.md](./travel-curated-experience-framework.md).

Coding workflow strategy should stay aligned with three public product rules:

- preserve one stable work thread even when the route changes underneath
- bias toward the cheapest suitable route with enough quota and continuity for
  long iteration loops
- keep route overrides available without making model choice the user's normal
  burden

## Product Thesis

Jini should be the calm front door to the best available AI runtime for the
job.

Users should not need to understand:

- which AI tool to use
- which provider to use
- which model to use
- which effort level to apply
- whether the work should stay on a local SLM or escalate to a paid route
- which local SLM should handle which kind of job
- where files live
- how work state is stored

Users should be able to:

- start with one command
- paste messy context
- get a useful outcome fast
- see what Jini used
- see what is still missing
- continue later without losing the thread

## Problem

The current market is split across strong but incomplete patterns.

### Claude Code

Wins on terminal immediacy and direct action.

Gap:

- less productized for non-expert users who need state, trust, and gentle
  continuation

### Kiro

Wins on visible workflow structure and disciplined progression.

Gap:

- can feel process-first, especially when the user just wants help finishing
  something

### Hermes

Wins on memory, continuity, and broad agent-platform feel.

Gap:

- breadth and autonomy can overshadow the simplest path to a concrete outcome

### AgentField

Wins on harness orchestration and production infrastructure for coding agents.

Gap:

- too backend-shaped for normal end users

### AI Hero

Wins on transparency, tool-call visibility, and human-in-the-loop trust.

Gap:

- teaching and transparency can dominate over ruthless outcome completion

### Cheap-first coding tools

Win on cost discipline and lean operator loops.

Gap:

- weaker persistent product surface for state, artifacts, and trust

## Product Positioning

Jini should position as:

- the easiest way to finish AI-assisted work with visible state, visible
  artifacts, and visible trust

Jini must not position as:

- a generic multi-agent infrastructure control plane
- a framework that asks the user to manage orchestration internals
- a builder-first protocol browser

## Target Users

Primary users:

- low-literacy first-time users
- pragmatic "just make it work" users
- Claude, Codex, ChatGPT, and Gemini users who want a calmer front door
- AI PMs
- AI engineers
- developers
- architects
- QA testers

Secondary users:

- AWS Bedrock users
- enterprise Azure users
- students
- homemakers
- travel advisors
- software leaders who care about quality, cost, and control

## Jobs To Be Done

### Flagship jobs

These are replacement-critical:

- turn meeting notes into something sendable
- check whether a plan or spec is ready to hand off or build from

### Secondary jobs

These can expand after flagship quality is proven:

- research to PRD
- vendor comparison
- incident follow-up
- trip planning
- general drafting and structured follow-up

## Product Principles

### 1. Outcome first

The first meaningful screen after intake should be a useful object, not a
status report.

### 2. Cheapest suitable by default

Jini should deploy the cheapest route that is likely to do the job well.

If the request clearly demands deeper rigor, Jini should escalate to a better
tool and higher effort automatically.

Commercially usable local SLMs should become the true frontline for most normal
work so paid remote models are used by need, not by habit.

That frontline should be a local model pool with different profiles for
different work classes, not one single local model for everything.

### 3. Visible trust

Jini should always make visible:

- what it used
- what it chose
- why it chose it
- what is missing
- what is safe

### 4. Artifacts first, chat second

Useful work should live in named artifacts, not only in the transcript.

### 5. One front door

The normal product entry is:

- `jini`

Normal users should not need to learn a command family before getting value.

### 6. No path-driven normal flow

Users should never need raw file paths, pack paths, or internal ids in the
normal experience.

### 7. Continuity without hidden drift

Jini should remember work, but never act on remembered work silently.

## User Experience

## Install And First Run

The install and start story must collapse to:

1. install once
2. run `jini`
3. paste what needs to be finished

If setup is required, Jini handles it inline.

The normal user should not need to learn:

- env vars
- provider doctor
- route theory
- model naming
- effort classes

before first success.

## First Minute

The first minute must feel like relief, not setup.

Required first-run shape:

```text
What do you need help finishing?

Jini shell
Paste notes or type what you want finished.

If you need setup help, type `Use Auto`.

Nothing will be sent yet.
```

If the user is unsure, Jini should still accept rough input and produce a first
useful pass instead of forcing early classification.

## First Result

The first visible result should be the thing the user asked Jini to make.

Examples:

- `Sendable Follow-up`
- `Build-Readiness Check`
- `Recommendation Memo`
- `Trip Plan`

The summary comes after the useful result, not before it.

## Persistent Work Thread

Jini should behave like a visible work thread, not a loose chat log.

The stable frame:

- `Goal`
- `Working with`
- `Now`
- `Done`
- `Need`
- `Next`
- `Ready now`
- `Blocked`
- `Safe to do`

When runtime trust matters, the frame also shows:

- `AI route`
- `Model`
- `Route policy`
- `Effort level`
- `Why this was chosen`

## Inputs

Jini should accept and surface:

- pasted text
- files
- images
- audio
- screenshots
- links

Every input becomes a visible `InputItem`.

Processed inputs must show what Jini extracted.

## Artifacts

Jini should produce named, openable artifacts early.

Artifacts should be grouped as:

- `Ready now`
- `Needs input`
- `Blocked`

Each artifact should have:

- title
- status
- short summary
- open action
- export actions

## Multi-Project Use

Jini should support many work threads with one current focus.

If multiple active threads exist, `jini` should surface an active-work chooser.

Users must be able to:

- continue current work
- switch work
- start new work

without filesystem knowledge.

## Runtime Intelligence

Jini should treat these as first-class product behaviors on every request:

- tool selection
- provider selection
- model selection
- effort selection
- local-SLM-first versus remote-escalation selection
- local profile selection when the work stays local

Decision order:

1. classify work type
2. classify depth
3. decide whether the local commercial SLM pool can handle it well enough
4. if local is suitable, choose the correct local profile
5. choose the cheapest suitable route
6. choose model for that route
7. choose effort level
8. show and persist the decision

Effort classes:

- `low`
- `medium`
- `high`
- `extra high`

Escalation triggers include:

- benchmark work
- architecture work
- critique and review
- release-readiness
- exhaustive or root-cause requests

## Functional Requirements

### Install And Setup

- one recommended install path
- `jini` is the normal entry after install
- inline route setup for Claude, Bedrock, Azure, and Auto
- route and secret storage behavior explained plainly
- `provider doctor` remains a support path, not the happy path

### Conversation Handling

- conversation is narration around a durable work thread
- each turn should record what changed
- each turn should show which artifacts changed
- one active blocking ask at a time
- assumptions shown explicitly when Jini proceeds without asking

### Outcome Surfaces

- useful object first
- artifact shelf visible after first result
- blocker and uncertainty visibility preserved
- continuation actions kept small and real

### Trust And Safety

- Jini must not present drafts as final truth
- missing proof and approval gaps must remain visible
- safe and reversible next steps must be stated plainly

### Cost Discipline

- cheapest suitable route is the default
- expensive route or effort choices must be visible and justified
- resumptions should prefer bounded context reloads over full refresh

## Non-Goals

- expose orchestration internals as the normal user surface
- optimize first for general multi-agent backend composition
- expand demo jobs before flagship flows are excellent
- add more public commands instead of improving `jini`

## Success Metrics

North-star:

- time to first useful outcome

Primary:

- install-to-first-result success rate
- first-run completion without docs
- percent of sessions succeeding through `Use Auto`
- percent of sessions where users can identify what is done, missing, and next
- resume success on existing work
- multi-project switch success

Economic:

- cheap-route usage rate on normal work
- cost per successful task
- false-positive deep-work escalation rate

Trust:

- percent of sessions with visible route readout before work begins
- percent of sessions with explicit missing-info callout when needed
- percent of processed inputs with visible extraction summary

## Scorecard Alignment

This PRD is valid only if it preserves or improves the current scorecard lead.

### Workflow rigor

Requirement:

- keep the protocol kernel rigorous, but surface rigor through a clearer product
  loop instead of internal vocabulary

PRD implication:

- visible work thread and artifact model must not weaken canonical protocol
  semantics

### Delivery maturity

Requirement:

- reduce operator friction and compress the path from idea to usable output

PRD implication:

- `jini` must feel complete on first run
- first useful result must come before summary
- continuation must be real, not decorative

### Packaging and install ergonomics

Requirement:

- one install command and one obvious post-install command

PRD implication:

- install and launch UX outrank secondary choice surfaces

### Memory reliability

Requirement:

- returning users should not need to restate context

PRD implication:

- multi-project `WorkThread` continuity is core product behavior

### Adapter portability

Requirement:

- runtime choice should work through the same canonical contract

PRD implication:

- runtime routing is product behavior, but canonical work semantics stay stable

### Token efficiency

Requirement:

- cheapest adequate path is the normal path

PRD implication:

- runtime selection defaults to cheap-first and explains escalation

### Advanced-set breadth

Requirement:

- support broad workflows without bloating the kernel

PRD implication:

- breadth should expand through packs, artifacts, adapters, and routines, not a
  wider public command surface

## Golden Benchmark Alignment

This PRD is designed to improve or protect the benchmark scenarios in
[golden-competitive-benchmark.yaml](./golden-competitive-benchmark.yaml).

### Install trust

PRD response:

- one install path
- one front door
- one obvious first sentence to type

### Cheap default loop

PRD response:

- cheapest suitable route by default
- visible cost discipline
- bounded resumptions

### Portable edges

PRD response:

- stable canonical work contract across runtime and publish edges
- visible route decision without exposing backend plumbing as the main UX

### Guided product loop

PRD response:

- one obvious work loop from messy input to usable result to next step

### Memory and continuity scenarios

PRD response:

- visible current work
- recoverable multi-project continuity
- durable thread and artifact model

## Release Gates

This PRD should not be treated as approved without clearing:

- [product-review-roles.md](./product-review-roles.md)
- [dogfood-gates.md](./dogfood-gates.md)
- [competitive-kpis.yaml](./competitive-kpis.yaml)
- [golden-competitive-benchmark.yaml](./golden-competitive-benchmark.yaml)
- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)

Major product work under this PRD must answer:

- does this improve or protect the benchmark lead
- does this help the first minute feel easier
- does this keep route, model, and effort choices visible
- does this improve outcome quality without forcing more process

## Roadmap

### Phase 1: Finish flagship loops

- make meeting follow-up bulletproof
- make plan/spec readiness bulletproof
- keep result-first flow intact
- preserve clear current-work continuation

### Phase 2: Deepen thread and artifact UX

- attachment enrichment for PDF, image, and audio
- artifact shelf grouping for `Ready now`, `Needs input`, and `Blocked`
- stronger active-work chooser and multi-project continuity

### Phase 3: Deepen runtime intelligence

- provider-native effort controls where possible
- better strict-route handling
- bounded multi-route orchestration for compare, review, and critique

### Phase 4: Expand breadth

- research to PRD
- vendor comparison
- incident follow-up
- trip planning

Expansion happens only after flagship flows clear the same quality and parity
gates.

## Risks

- exposing too much routing theory too early
- adding breadth before the flagship loops are excellent
- letting auto mode feel magical instead of trustworthy
- raising quality while silently raising cost
- turning continuity into clutter

## Product Decision

Jini should not become another AI terminal with nicer labels.

Jini should become:

- a binary CLI with one front door
- a visible work thread
- a visible artifact shelf
- a visible runtime decision policy
- a cheapest-suitable default
- an honest account of what is done, what is missing, and what happens next
