# Number One Platform PRD

Updated: 2026-06-05

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

### 5a. Route context, do not dump it

Jini should prefer a small routing layer over one giant monolithic instruction
file.

The system should load:

- a small root context map
- the relevant domain files
- the relevant skills
- the relevant metrics or evidence

It should not load unrelated rules just because they exist.

The context architecture must reduce token use and improve task relevance over
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

### 9. Disagreement beats fake certainty

For high-impact decisions, Jini should prefer structured critic or debate loops
over single-draft confidence theater.

The product should support disagreement-driven improvement where that raises
quality more than it raises noise.

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

### 2a. Context Architecture

Jini must implement a summary-first context routing architecture.

Minimum required layers:

- root routing context
- domain-specific knowledge files
- workflow or skill files
- evidence and metrics files
- confirmed rules and hypotheses by domain

The product must avoid growing one global context file until it becomes a token
tax.

The system should be able to explain:

- which context files were loaded
- why they were loaded
- which likely-relevant files were intentionally skipped

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

Jini must also support the operator pattern:

1. do the task manually a few times
2. identify the repeated structure
3. extract the SOP and examples
4. convert that pattern into an inspectable reusable automation
5. learn from the outcomes of the automation over time

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
- confirmed-rule and hypothesis repair by domain
- context-bloat detection and routing-table repair

### 5. Upstream Quality Automation

Jini must move quality upstream through automatic checks before users pay for
bad output.

Examples:

- missing evidence before readiness claims
- missing reviewers before handoff
- stale assumptions before continuation
- invalid environment before task execution
- low-confidence route selection before expensive escalation
- missing critic pass before high-stakes decisions

### 5a. Decision Critic System

Jini must support structured critic loops for decisions that are:

- high-cost
- high-risk
- irreversible
- strategy-shaping

These loops should support:

- adversarial review
- disagreement between perspectives
- moderator or synthesis step
- explicit settlement criteria

The product should not simulate committee theater for low-stakes work, but it
should not pretend one draft is enough for important decisions.

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

### 11. Build, Buy, and Integrate Discipline

Jini must decide deliberately when to:

- build native capability
- integrate an existing platform capability
- wrap a platform-native local model or connector

The default should not be "build everything ourselves."

If a platform-native capability materially improves local privacy, latency,
distribution, or user trust, Jini should prefer integrating it behind the same
session and artifact contract.

### 12. Economic Value Measurement

Jini must judge improvements by economic and operator value, not just by model
intelligence.

Required evaluation dimensions include:

- time saved
- interruptions avoided
- rework reduced
- token or provider cost reduced
- quality misses prevented
- follow-through completed

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
- manual SOPs and repeated examples must be convertible into automations
- critic loops must exist for high-impact decisions and company-critical work

#### P0.3 Context routing discipline

- one giant instruction file is not the default operating shape
- root routing context exists
- domain rules, hypotheses, metrics, and procedures are separable
- the system loads only relevant context for the task
- context-bloat regression is visible and repairable

#### P0.4 Local-first cost discipline

- local SLM pool exists as a real frontline route
- form-factor-aware local model support matrix exists
- escalation to paid routes is explicit and justified

#### P0.5 Development system automation

- repo-aware environment learning
- one canonical gate matrix for commit, push, and release
- one checked-in runner for required gate tiers
- quality gates before handoff
- missing-proof and missing-review detection
- docs/help/runtime parity repair

#### P0.6 Deployment and shipping system

- idempotent release and rollback semantics
- signed or attestable build and delivery pipeline where applicable
- release receipts and deployment continuity
- model/runtime update canaries before promotion

#### P0.7 App-surface shipping contract

- desktop and mobile are bound to the same work object as CLI
- platform-by-platform offline strategy exists for macOS, Windows, Android,
  and iOS
- macOS and Windows desktop shipping is part of the committed product roadmap
- mobile continuation, review, and approval are part of the committed roadmap

#### P0.8 Ask-handling system

- inbound asks must become triageable work objects
- support, product asks, bug reports, and follow-up requests must feed one
  prioritized operating loop
- recurring asks must train workflow compression and automation candidates

#### P0.9 Build, buy, and platform leverage

- platform-native capabilities should be adopted when they improve privacy,
  latency, trust, or distribution
- integrations must still resolve into one Jini session, route, and artifact
  contract
- local-model or connector adoption should be judged by measured value, not by
  novelty

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
3. root routing context and domain-context architecture
4. repeated-task compression and reusable automations
5. upstream quality automation before execution or handoff
6. self-correction loops for docs, runtime, benchmark drift, and context bloat

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
5. route asks through the correct domain knowledge and SOP surfaces
6. keep visible backlog, proof, and follow-through state

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
5. update confirmed rules and hypotheses by domain as evidence accumulates
6. use monthly release trains for surface changes and faster policy cadence for
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
- context routing maintenance
- rule and hypothesis maintenance by domain
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
