# Number One Platform PRD

Updated: 2026-05-27

The current execution contract for the next major product and architecture
initiative lives in
[jini-next-initiative-plan.md](./jini-next-initiative-plan.md).

This document is the canonical product and operating PRD for Jini.

If another plan, roadmap, or strategy note conflicts with this document, this
document wins unless a newer explicit PRD decision supersedes it.

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

## Non-Negotiable Rule

Agentic automation in all aspects of running this company is non-negotiable.

That does not mean reckless autonomy.

It means Jini must be built so the company increasingly runs through:

- inspectable automation
- reversible automation
- measurable automation
- policy-bounded automation
- automation that reduces human toil instead of creating hidden operator debt

This applies to:

- product development
- testing and quality gates
- release and deployment
- support and inbound asks
- roadmap maintenance
- competitive monitoring
- model and runtime updates
- customer continuity and follow-through

The company should not depend on manual babysitting as the default operating
mode once a loop is understood well enough to automate safely.

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

## Tenets

### 1. One work object everywhere

The same work must survive across CLI, desktop, mobile, offline mode, and
provider changes without becoming a second-class copy.

### 2. Cheapest suitable route first

Jini should stay frugal by default, with commercially usable local SLMs as the
front line whenever they are good enough.

### 3. Outcome before narration

Jini should produce a useful artifact or decision surface before it produces a
framework lecture or a diagnostic transcript.

### 4. Agentic automation with visible trust

Automation is required, but it must always remain inspectable, bounded, and
reversible.

### 5. Learn once, reuse many times

The product should compound from environment learning, workflow learning,
artifact reuse, and policy improvement so repeated work becomes cheaper over
time.

### 6. Apps are specialized views, not separate products

CLI, desktop, and mobile should differ in strengths, not in identity, memory,
or trust semantics.

### 7. Shipping quality beats shipping theater

Jini should not ship flashy surfaces that increase drift, trust debt, or
manual cleanup burden.

### 8. Company operations should become productized loops

Support, release, follow-up, roadmap intake, and model refresh should become
first-class Jini loops, not side chores around the product.

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

## Prioritized Requirements

### P0: Company-Critical Requirements

These are the requirements Jini must satisfy before broad expansion work is
considered successful.

#### P0.1 Session and continuity truth

- one canonical session graph
- one artifact graph
- one route and trust story across surfaces
- offline continuation that remains legible and recoverable

#### P0.2 Agentic automation backbone

- repeated workflows become reusable automations
- automations have explicit approval, rollback, and proof
- human babysitting is not the default for known loops

#### P0.3 Local-first cost discipline

- local SLM pool exists as a real frontline route
- form-factor-aware local model support matrix exists
- escalation to paid routes is explicit and justified

#### P0.4 Development system automation

- repo-aware environment learning
- quality gates before handoff
- missing-proof and missing-review detection
- docs/help/runtime parity repair

#### P0.5 Deployment and shipping system

- idempotent release and rollback semantics
- signed or attestable build and delivery pipeline where applicable
- release receipts and deployment continuity
- model/runtime update canaries before promotion

#### P0.6 App-surface shipping contract

- desktop and mobile are bound to the same work object as CLI
- macOS and Windows desktop shipping is part of the committed product roadmap
- mobile continuation, review, and approval are part of the committed roadmap

#### P0.7 Ask-handling system

- inbound asks must become triageable work objects
- support, product asks, bug reports, and follow-up requests must feed one
  prioritized operating loop
- recurring asks must train workflow compression and automation candidates

### P1: Scale and differentiation requirements

- stronger GitHub-native engineering continuity
- broader multimodal local workflows
- richer artifact editing surfaces on desktop
- managed continuity convenience across devices
- stronger benchmark and competitor-watch automation

### P2: Expansion requirements

- wider domain packs after flagship dominance is proven
- deeper team and enterprise governance surfaces
- more specialized workflow packs once the core operating loop is stable

## Roadmap

### Track A: Development

Objective:

- make Jini the default operating system for building and improving itself

Roadmap:

1. canonical session and artifact graph
2. repo and environment learning
3. repeated-task compression and reusable automations
4. upstream quality automation before execution or handoff
5. self-correction loops for docs, runtime, and benchmark drift

Exit condition:

- most repeated product-development work is either automated or compressed into
  inspectable reusable flows

### Track B: Deployment and Release

Objective:

- make deployment, release, rollback, and verification agentic by default

Roadmap:

1. release receipts and proof surfaces
2. stable deployment and rollback contracts
3. automated regression and benchmark gates
4. model and runtime canary promotion loop
5. reconciliation-safe managed recovery when online capability drifts

Exit condition:

- shipping a release does not require heroic manual coordination to remain safe

### Track C: App Shipping

Objective:

- ship specialized surfaces over the same work system

Roadmap:

1. CLI remains the canonical live surface
2. macOS desktop
3. Windows desktop
4. Android continuation/review
5. iOS continuation/review under App Store constraints

Role split:

- CLI -> fastest universal control surface
- desktop -> artifact inspection, editing, review, supervision
- mobile -> interruption-safe review, approval, defer, lightweight continuation

Exit condition:

- every shipped app feels like the same Jini session, not a disconnected client

### Track D: Addressing Asks

Objective:

- turn inbound requests into one governed company operating loop

Sources of asks:

- users
- customers
- internal teams
- dogfood feedback
- benchmark regressions
- competitor deltas
- runtime/model ecosystem changes

Roadmap:

1. normalize asks into one intake and triage surface
2. classify asks by urgency, domain, and repeatability
3. attach asks to canonical work objects and artifacts
4. detect repeated asks and generate automation or workflow candidates
5. keep visible backlog, proof, and follow-through state

Exit condition:

- the company does not lose or re-litigate important asks because intake is too
  manual or fragmented

### Track E: Future Updates

Objective:

- ship future improvements through policy, learning, and measured promotion

Roadmap:

1. maintain official-source watch loops for supported model families and
   runtimes
2. run canary evaluations for successor local and managed routes
3. promote only score-positive route changes
4. keep user-facing product contracts stable while internal policies improve
5. use monthly release trains for surface changes and faster policy cadence for
   measured route improvements

Exit condition:

- Jini gets materially better over time without constant product-surface churn

## Operating Cadence

Development, deployment, asks, and model updates should run on separate
cadences instead of one noisy release loop.

- daily: traces, asks, regressions, automation receipts, benchmark evidence
- weekly: triage review, policy candidates, competitor delta, model watch
- monthly: user-facing releases across CLI and shipped apps
- quarterly: architectural simplification and roadmap reset

## Shipping Philosophy

Jini should ship improvements in this order:

1. reliability and continuity
2. automation that removes real toil
3. cost reduction that does not hide quality risk
4. new surfaces only when they preserve work identity
5. broader flows only after flagship dominance and trust are proven

## Company Operations Requirement

The company itself should increasingly run on Jini loops.

Minimum required loops:

- product intake and prioritization
- bug and ask triage
- roadmap maintenance
- release preparation
- release verification
- model and runtime update review
- support follow-up
- benchmark and competitor watch

These loops must become:

- visible
- measurable
- automatable
- inspectable
- overrideable

If a core company loop remains permanently manual, that is product debt.

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
