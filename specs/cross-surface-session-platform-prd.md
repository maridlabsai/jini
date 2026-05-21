# Cross-Surface Session Platform PRD

Updated: 2026-05-21

## Purpose

This PRD defines the stricter product charter Jini must satisfy:

- cost optimizer first
- UX second to none
- one session preserved across every supported surface

This document exists because the older product materials describe continuity
and multi-surface support, but they do not yet make those three requirements
the dominant product contract.

This PRD should be read alongside:

- [full-product-prd.md](./full-product-prd.md)
- [lean-platform-doctrine.md](./lean-platform-doctrine.md)
- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)
- [work-state-machine.md](./work-state-machine.md)
- [artifact-schemas.md](./artifact-schemas.md)

## Product Charter

Jini should be the lowest-cost, highest-clarity shell for work that must
persist, resume, and finish safely across devices and interfaces.

Jini is not a chat app with memory bolted on later.

Jini is not a desktop wrapper around a CLI.

Jini is not a premium-model launcher that happens to remember files.

Jini is one session-preserving shell that users can enter from:

- macOS
- Windows
- mobile
- CLI

## Non-Negotiable Product Order

### 1. Cost Optimizer First

Jini should minimize the total cost to useful outcome:

- model spend
- operator time
- retry count
- clarification burden
- handoff overhead
- interruption recovery cost

The default route should be the cheapest suitable route, not the fanciest
route.

### 2. UX Second To None

Jini should feel easier than raw model shells, not more complicated than them.

The UX bar is:

- faster first useful result
- less ambiguity
- lower cognitive load
- clearer next action
- lower interruption anxiety
- easier recovery after switching surfaces

### 3. Session Continuity Across Form Factors

The same session should be visible and resumable from every supported form
factor.

Supported form factors:

- macOS
- Windows
- mobile
- CLI

Resume should not mean "start a similar task again." Resume should mean
"continue the same session with the same context, state, artifacts, and route
evidence."

## Category Boundary

Jini is for work with an awkward middle:

- there is messy input
- there is a deliverable to finish
- the work may pause and resume
- the work may move across devices
- the work may need review or handoff

Jini is not optimized for:

- one-shot chat
- novelty prompts
- provider playground usage
- users who only want raw model answers with no artifact, continuity, or
  review posture

## Core User Promise

The user should be able to:

1. start anywhere
2. get a useful object fast
3. leave
4. come back on another supported surface
5. see exactly what is ready, what is missing, and what to do next
6. continue without rebuilding context

## Canonical Session Model

Every supported surface should read and write the same logical session object.

Minimum session fields:

- stable session id
- title
- user goal
- current status
- ready state
- missing state
- next action
- artifact list
- source references
- route evidence
- review-safe state
- approval/sending boundary
- last updated time
- last active surface

### Session States

The session model should preserve:

- active
- waiting
- blocked
- ready for review
- done
- archived

### Session Guarantees

Every session should preserve:

- a usable latest deliverable
- explicit missing items
- explicit assumptions
- explicit risks
- explicit route choice evidence
- explicit review/send boundary

## Surface Contract

Every supported surface should present the same mental model.

### Required Surface Questions

Every primary Jini surface should answer:

- what is this session
- what happened last
- what is ready now
- what is missing now
- what should happen next
- what is safe to review or share

### CLI Contract

The CLI should remain the thinnest, fastest surface.

The CLI must:

- show current sessions quickly
- resume the same session model
- open artifacts directly
- expose route and cost evidence on demand
- avoid path or provider jargon in the default flow

### Desktop Contract

Desktop should expose the same session model with richer visibility, not a
different workflow model.

Desktop must:

- show active and recent sessions
- show artifacts and ready/missing state
- resume sessions directly
- keep route evidence inspectable

### Mobile Contract

Mobile should focus on continuity, quick review, and lightweight continuation.

Mobile must:

- show recent sessions fast
- show latest ready deliverable fast
- show what is missing
- allow lightweight continuation and handoff
- preserve the exact same session identity

### Windows And macOS Contract

Platform availability is a product requirement, not a stretch goal.

Windows and macOS should:

- support the same session model
- support the same command vocabulary where relevant
- support the same resume semantics
- avoid platform-exclusive concepts in the primary workflow

## Routing And Cost Contract

The session model must carry route evidence across surfaces.

Every surface should be able to show:

- what route was used
- why it was chosen
- whether it stayed local or escalated
- whether the current route is still the cheapest suitable route

Jini should optimize:

- continuation cost over restart cost
- local reuse over fresh premium calls
- concise clarification over broad reprompting

## UX Rules

### Default Path

The default path should work for users across:

- technical backgrounds
- education levels
- domains
- provider familiarity levels

The default user should not need to know:

- provider names
- model names
- route policy internals
- filesystem internals
- artifact schema internals

### Action Clarity

Jini should favor short, standard actions.

Every important interaction should expose:

- one obvious next step
- one obvious way to inspect missing state
- one obvious way to resume later

### Interruption Recovery

Switching surfaces should not create:

- hidden drift
- duplicate sessions
- ambiguous latest state
- artifact confusion

## Trust Rules

The user should be able to inspect:

- what is stored locally
- what is not stored as product magic
- what was generated
- what is still missing
- what is safe to share
- what requires approval before send

Trust should come from inspectability, not branding language.

## Primary Metrics

Jini should track:

- cost-per-successful-task
- time-to-first-useful-result
- continuation-success-rate
- cross-surface-resume-success-rate
- recovery-time-after-interruption
- artifact-open-rate
- premium-route-regret-rate
- command-surface-count
- clarification-turn-count

## Reject Conditions

Reject work that:

- improves intelligence but increases cost without clear outcome gain
- improves architecture but worsens default UX clarity
- adds a supported surface without shared session continuity
- creates platform-specific workflow drift
- forces users to reconstruct context when switching surfaces
- hides route or trust evidence behind expert-only paths
- adds product ceremony before first useful result

## Rollout Priorities

### Phase 1: Canonical Session Contract

- define the stable session object
- define ready/missing/next semantics
- define review-safe and send-boundary semantics
- define cross-surface resume semantics

### Phase 2: Surface Parity

- CLI parity with canonical session contract
- macOS parity
- Windows parity
- mobile review and resume parity

### Phase 3: Cost And Continuity Evidence

- expose route evidence everywhere
- expose continuation savings everywhere
- expose interruption recovery evidence everywhere

### Phase 4: UX Hardening

- reduce cognitive load on every surface
- simplify next-step presentation
- validate low-expertise usability
- validate cross-device recovery speed

## Definition Of Done

This charter is only satisfied when:

- a user can start on one supported surface and resume on another without
  rebuilding context
- the session object remains stable across those surfaces
- Jini still chooses the cheapest suitable route by default
- the default path remains understandable for non-expert users
- trust, route, and artifact state remain inspectable everywhere
