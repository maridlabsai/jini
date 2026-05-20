# Jini Full Product PRD Execution Plan

Updated: 2026-05-15

## Purpose

This plan converts [full-product-prd.md](./full-product-prd.md) into a staged
rewrite sequence with concrete slice goals, scorecard intent, benchmark intent,
and merge gates.

This is not a backlog dump.

It is the approved order of operations for turning the current rewrite into a
product that can beat the current field on:

- ease of start
- visible trust
- outcome quality
- continuity
- cost discipline

## Planning Rules

Every slice in this plan must satisfy all of these:

- improve or preserve the golden benchmark lead
- clear the rewrite score floor
- preserve workflow rigor, packaging-install, and advanced-set breadth
- pass the relevant dogfood persona gates
- improve real user comprehension, not just internal architecture

If a slice makes the product feel simpler while weakening canonical semantics,
it fails this plan.

## Current State

The rewrite already has meaningful pieces in place:

- one human-facing front door centered on `jini`
- inline route setup with `Use Auto` and strict-route choices
- route, model, and effort heuristics as first-class policy
- visible preflight route decision card
- active-work switching for multiple work items
- a first `WorkThread` and `InputItem` runtime projection
- a thread frame with `Goal`, `Working with`, `Now`, `Done`, `Need`, and `Next`

The main remaining gap is not basic architecture.

The gap is product completeness:

- richer conversation handling
- richer attachment handling
- stronger artifact-centered interaction
- stronger flagship work quality
- stronger visible continuity

## Rewrite Tracks

This plan uses five tracks.

### Track A: First Minute And Setup

Goal:

- make install, start, and first useful input obvious enough for low-confidence
  users

### Track B: Work Thread And Conversation

Goal:

- make Jini feel like a live collaborator, not a prompt-plus-recap shell

### Track C: Artifact Shelf And Outcome Quality

Goal:

- make useful artifacts the center of the product, with clear readiness and
  blocker states

### Track D: Runtime Intelligence And Trust

Goal:

- make tool, provider, model, and effort selection both smart and legible
- make commercially usable local SLMs the true cheap-first front line
- make local routing profile-based instead of one-model-fits-all

### Track E: Continuity And Multi-Project Work

Goal:

- make resuming, switching, and continuing work feel reliable and calm

## Phases

## Phase 1: Bulletproof The Two Flagship Loops

Purpose:

- meeting follow-up and plan/spec readiness must become clearly better than
  generic AI chat and competitive enough with the current leaders

Why first:

- if these two loops are not excellent, broader product ambition is noise

### Scope

- improve first-result quality for the two flagship jobs
- ensure the first useful object appears before the summary
- keep continuation real: `Keep going`, `Show what is missing`, `Help me plan this`
- keep missing proof, approval gaps, and uncertainty visible
- align the thread frame and artifact names to the flagship outputs

### Required user-visible outcomes

- meeting notes become a genuinely sendable follow-up
- plan/spec input becomes a genuinely useful build-readiness check
- the user sees what is missing without reading internal state

### Scorecard intent

Expected to improve:

- `delivery-maturity`
- `workflow-rigor`
- `token-efficiency`

At risk:

- `core-simplicity` if continuation UI becomes overloaded

### Benchmark intent

Primary scenarios affected:

- guided product loop
- cheap default loop

### Dogfood gates

Must pass:

- low-literacy first-time user
- pragmatic “just make it work” user
- AI PM
- QA tester

### Exit criteria

- both flagship jobs produce useful first artifacts without extra explanation
- both jobs have real continuation paths
- summary labels and artifact names are human, not system-shaped
- current benchmark lead is preserved

## Phase 2: Upgrade Conversation Into A Real Work Thread

Purpose:

- make Jini conversation feel closer to Codex-style active collaboration

### Scope

- add `TurnRecord` semantics to runtime behavior
- show per-turn deltas: what changed, what was created, what was updated
- add `Just finished`, `Doing now`, and `Up next` behavior inside the thread
- surface one active blocking ask at a time
- show explicit assumptions when Jini proceeds without asking

### Required user-visible outcomes

- users can tell what changed on the current turn
- users can tell which artifact changed
- users are not hit with broad interrogation flows

### Scorecard intent

Expected to improve:

- `delivery-maturity`
- `memory-reliability`
- `workflow-rigor`

At risk:

- `token-efficiency` if narration becomes too verbose

### Benchmark intent

Primary scenarios affected:

- guided product loop
- memory and continuity behavior inside the broader benchmark set

### Dogfood gates

Must pass:

- low-literacy first-time user
- Claude user
- Codex user
- power user
- AI engineer

### Exit criteria

- each turn can identify changed artifacts and changed state
- one active blocking ask maximum
- conversation feels like guided work, not a static summary renderer

## Phase 3: Deepen Attachments And Artifact Shelf

Purpose:

- make inputs and outputs first-class across text, files, images, audio, and
  linked material

### Scope

- enrich `InputItem` processing for PDFs, images, and audio
- surface extraction summaries for processed inputs
- show parsing or transcription failures visibly
- group artifacts under `Ready now`, `Needs input`, and `Blocked`
- make artifact shelf the center of post-result interaction

### Required user-visible outcomes

- user can attach source material without losing track of it
- user can see what Jini extracted from each input
- user can see which outputs are usable and which are blocked

### Scorecard intent

Expected to improve:

- `delivery-maturity`
- `memory-reliability`
- `advanced-set-breadth`

At risk:

- `core-simplicity` if artifact surfaces become cluttered

### Benchmark intent

Primary scenarios affected:

- guided product loop
- portable edges, indirectly through clearer canonical artifact surfaces

### Dogfood gates

Must pass:

- student user
- homemaker user
- travel advisor user
- architect user

### Exit criteria

- image/audio/PDF inputs become visible and meaningfully processed
- failed extraction is explicit
- artifact shelf status is legible without opening every item

## Phase 4: Lock Down Runtime Intelligence As A Trust Surface

Purpose:

- make Jini’s runtime decision policy a product advantage rather than a hidden
  heuristic

### Scope

- add a local commercial SLM frontline policy
- add a visible local SLM route readout
- add local profile routing: `fast | workhorse | deep | multimodal`
- add a stable local SLM transport contract
- harden cheap-first versus best-tool escalation behavior
- improve strict-route handling for Azure, Bedrock, Claude, Codex, ChatGPT, and
  Gemini-style users
- add provider-native effort controls where backend support exists
- expose route, model, and effort reasons consistently before and during work
- keep the user override path simple

### Required user-visible outcomes

- ordinary work stays on a cheap local SLM when it safely can
- Jini chooses the right local profile for the job instead of one generic local model
- users understand what Jini chose and why
- expert users can force a route when policy matters
- deeper work visibly earns stronger runtime use

### Scorecard intent

Expected to improve:

- `token-efficiency`
- `adapter-portability`
- `delivery-maturity`

At risk:

- `packaging-install` if setup becomes too theory-heavy again

### Benchmark intent

Primary scenarios affected:

- cheap default loop
- portable edges
- install trust

### Dogfood gates

Must pass:

- AWS Bedrock user
- enterprise Azure user
- Claude user
- Codex user
- ChatGPT user
- Gemini user
- software VP user

### Exit criteria

- a real local SLM route exists as a front-line option rather than a fake preview
- local profile selection exists and is visible enough to debug or trust
- route, model, and effort readouts are visible and understandable
- auto mode feels legible, not magical
- strict-route users know how to force certainty

## Phase 5: Make Continuity And Multi-Project Work Calm

Purpose:

- make Jini reliable across many concurrent work threads

### Scope

- improve active-work chooser language
- add waiting, paused, and done states for work threads
- preserve progress and next-step summaries across many threads
- make stale current-work handling explicit and recoverable
- support thread history as work history, not only a latest-state recap

### Required user-visible outcomes

- users can manage several projects without path knowledge
- users can resume without restating context
- users can trust that Jini will not silently switch context

### Scorecard intent

Expected to improve:

- `memory-reliability`
- `delivery-maturity`
- `advanced-set-breadth`

At risk:

- `core-simplicity` if multi-project controls sprawl

### Benchmark intent

Primary scenarios affected:

- continuity and memory behavior
- guided product loop for repeat usage

### Dogfood gates

Must pass:

- pragmatic “just make it work” user
- power user
- AI PM
- software VP user

### Exit criteria

- active-work selection feels human and obvious
- users can switch, pause, and resume confidently
- remembered work never acts like magic

## Slice Template

Every major slice taken from this plan should answer these before implementation
starts:

### 1. User-visible change

- what exactly will feel easier, clearer, or more trustworthy

### 2. Scorecard intent

- which dimensions should move up
- which dimensions are at risk

### 3. Benchmark intent

- which benchmark scenarios should improve or be protected

### 4. Dogfood panel

- which personas are the gating audience for this slice

### 5. Done criteria

- what must be true for the slice to count as complete

### 6. Non-goals

- what this slice intentionally does not solve yet

## Working Order

Approved order:

1. flagship loop quality
2. turn-by-turn conversation upgrade
3. attachment and artifact shelf depth
4. runtime intelligence trust hardening
5. multi-project and continuity polish

Expansion flows stay behind these.

Do not reverse this order unless:

- the scorecard gate says a different sequence produces a stronger benchmark
  result, and
- the exception is recorded explicitly

## Push Gate For Plan-Driven Slices

Before pushing a major slice from this plan, record:

- scorecard dimensions improved
- scorecard dimensions at risk
- benchmark scenarios protected or improved
- rewrite floor status
- affected dogfood persona results

If that evidence is missing, the slice is not push-ready.
