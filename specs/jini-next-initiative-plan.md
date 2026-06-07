# Jini Next Initiative Plan

Updated: 2026-06-02
Status: proposed-for-execution

## Purpose

This plan curates the next major Jini initiative around one question:

- how do we turn Jini into the state-of-the-art operating surface for agentic
  work without losing simplicity, offline strength, or trust?

This is not a feature wishlist. It is the execution contract for:

- product direction
- architecture migration
- persona coverage
- offline excellence
- accessibility
- security
- traceability
- self-learning
- extensibility
- scorecards, SLOs, and SLAs

This initiative plan is subordinate to the canonical product and operating PRD
in [number-one-platform-prd.md](./number-one-platform-prd.md).

If this initiative plan and the canonical PRD disagree on priorities,
requirements, app-shipping order, automation posture, or operating rules, the
canonical PRD wins and this plan should be updated before execution continues.

Read this with:

- [number-one-platform-prd.md](./number-one-platform-prd.md)
- [lean-platform-doctrine.md](./lean-platform-doctrine.md)
- [cross-surface-session-system-and-dev-design.md](./cross-surface-session-system-and-dev-design.md)
- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)
- [dogfood-gates.md](./dogfood-gates.md)
- [dogfood-personas.yaml](./dogfood-personas.yaml)
- [../docs/archive/2026-06-02-codebase-snapshot-manifest.md](../docs/archive/2026-06-02-codebase-snapshot-manifest.md)

## Archive Before Change

The current codebase has been archived before this initiative proceeds.

- free repo snapshot: `fe5b94e1aa1d58b013f71d1843935ec795edf2e8`
- commercial repo snapshot: `b40ebffd4a01b2b389b0413117332129b973e4cd`
- archive bundles were created and verified before planning
- restore commands and bundle paths live in
  [../docs/archive/2026-06-02-codebase-snapshot-manifest.md](../docs/archive/2026-06-02-codebase-snapshot-manifest.md)

## Planner Verdict

Jini should keep the good parts:

- one calm front door
- durable work sessions
- artifact-first continuation
- local-first and cheap-first routing
- visible route truth
- offline continuation

Jini should let go of the controversial or low-value parts:

- product jargon in the default path
- duplicate command dialects
- status-heavy first screens
- agent ceremony for normal users
- overly generic personas that do not map to clear customer outcomes
- giant monolithic runtime growth without boundary correction

The product should be judged by whether these four users are happy:

- software engineer
- college student
- high school student
- realtor

If any one of them needs a different mental model to succeed, the design is
too complicated.

## Architect Verdict

Go is now the core language for Jini because execution efficiency, local
reliability, and cross-platform packaging are primary product requirements.

Decision:

- keep the public CLI and execution kernel native Go
- keep product semantics stable while expanding native command parity
- treat missing advanced commands as explicit Go backlog, not fallback work

Why Go:

- faster CLI startup
- easier static distribution
- lower memory overhead
- stronger concurrency for orchestration and probes
- simpler cross-platform delivery than an interpreted runtime
- lower friction for offline device-class execution

Why not a full Rust rewrite first:

- the efficiency upside is real, but the implementation complexity is too high
  for the current product-risk profile
- Jini needs product consolidation and boundary cleanup at the same time as
  performance work

## Product Outcome

Jini should become:

- the shell you start in for durable work
- the system that absorbs token exhaustion, throttling, slow churn, and weak
  connectivity
- the product that treats token frugality as P0 and avoids replaying or loading
  context unless it is necessary for the outcome
- the product that stays productive in the wilderness when the device is
  capable
- the adaptive execution layer that can switch platforms, models, and profiles
  without making the user re-decide the stack
- the execution layer that can go full power when plugged in and thermally safe,
  then become battery-conscious under low battery or thermal pressure

## Persona Outcomes

### Software Engineer

Success means:

- `jini` starts fast
- command truth is stable
- route selection is inspectable
- proofs and artifacts are first-class
- work can continue offline and reconcile later

### College Student

Success means:

- Jini helps turn notes, assignments, project briefs, and research into usable
  outputs quickly
- campus Wi-Fi failure does not end the session
- the product does not require provider or model vocabulary

### High School Student

Success means:

- the default flow is safe, plain-language, and low-jargon
- accessibility and clarity are better than competing agent shells
- the product can work offline on constrained devices when tasks are local
  enough

### Realtor

Success means:

- notes from calls, showings, and follow-ups become sendable artifacts
- mobile and desktop continuity is real
- privacy and trust boundaries are obvious
- offline capture and later reconciliation are calm, not surprising

## Contradictions To Resolve

### 1. Shell-Anywhere vs Proof-Heavy Work

Resolution:

- keep `jini` shell-anywhere
- move proof, blockers, approvals, and route detail into progressive layers
- first useful result still comes before state narration

### 2. Offline-First vs Managed Convenience

Resolution:

- the open/free product must stay genuinely useful offline
- the commercial layer may improve recovery, switching, forecasting, and sync
- the commercial layer must not own the only usable continuation path

### 3. Adaptive Intelligence vs Simple UX

Resolution:

- Jini should make more decisions internally
- Jini should expose fewer decisions externally
- any visible complexity must correspond to trust, safety, or recovery value

### 4. Self-Learning vs Privacy And Trust

Resolution:

- session learning should be inspectable, scoped, and reversible
- local-first memory remains the default
- hosted learning should be explicit and boundary-aware

### 5. Multiple Engines vs One Stable Product

Resolution:

- Claude Code, Codex, GitHub CLI, local SLMs, and commercial managed routes
  are execution targets
- Jini is the product surface
- users should not need to think in backend dialects

## State-Of-The-Art Requirements

Jini should be state of the art in:

- agentic workflow continuity
- shell UX
- offline execution
- accessibility
- security
- self-learning
- evolution under feedback
- traceability
- extensibility
- contextual adaptation by user, domain, use case, and device

## Target Architecture

### Layer 1: Go Session Kernel

Owns:

- canonical session object
- event log
- projection engine
- compact context builder
- route/runtime state
- reconciliation debt ledger

### Layer 2: Go Routing And Execution Core

Owns:

- local capability probes
- offline/online state detection
- adaptive platform switching
- task-shaped model/profile selection
- rate-limit and throttle avoidance
- runtime health scoring

### Layer 3: Native Command Surface

Owns after cutover:

- CLI surface preservation
- compatibility aliases that are implemented in Go
- docs and test harness integration
- fail-fast handling for commands not yet ported natively

### Layer 4: Surface Adapters

Owns:

- CLI
- desktop
- mobile
- commercial managed continuity

### Layer 5: Extensibility Plane

Owns:

- adapters
- skills
- artifact renderers
- publish/export edges
- domain-specific workflow packs

## Capability Pillars

### Offline Excellence

Jini must:

- state when it is in offline mode
- state whether online capability is available
- continue using imported context already attached to the session
- track reconciliation debt created while offline
- reconcile safely and visibly when connectivity returns

### Accessibility

Jini must:

- keep CLI output screen-reader-friendly
- preserve keyboard-only flows
- use plain-language default copy
- target WCAG 2.2 AA on desktop and mobile surfaces
- pass a low-literacy comprehension gate for first-minute usage

### Security

Jini must:

- keep secrets redacted in all normal output paths
- isolate local and hosted trust boundaries
- require explicit confirmation for publish, payment, booking, and external
  writes
- preserve an auditable route, artifact, and action trail

### Self-Learning

Jini must:

- learn from route regret
- learn from task outcomes
- learn from recurring repo, domain, and device patterns
- improve suggestions and routing without changing the user's mental model

### Traceability

Jini must:

- keep event lineage for route choice, memory use, artifacts, and approvals
- make route and artifact provenance visible on demand
- preserve proof under compact and offline projections

### Extensibility

Jini must:

- make adapters pluggable
- keep domain workflows modular
- avoid hard-coding use-case-specific response molds into the core shell

## Free And Commercial Split

### Free/Open Must Keep

- CLI access
- local SLM support
- BYO provider support
- offline-first continuation
- imported-context reuse offline
- artifact-first continuation
- compact status and resume
- visible route truth

### Commercial May Add

- managed sync convenience
- proactive throttle avoidance
- automatic platform switching across managed targets
- subscription forecasting
- reconciliation scheduling
- advanced team governance
- stronger hosted recovery and supervision

## Scorecard

The initiative succeeds only if the measured scorecard improves.

### Product Scorecard

- first useful result quality: `>= 9.3`
- continuation clarity: `>= 9.3`
- offline usefulness: `>= 9.2`
- trust and provenance clarity: `>= 9.3`
- accessibility clarity: `>= 9.1`
- security posture clarity: `>= 9.2`
- extensibility fitness: `>= 9.0`

### Engineering Scorecard

- CLI cold start score: `>= 9.4`
- runtime memory footprint score: `>= 9.2`
- adapter portability score: `>= 9.1`
- token efficiency score: `>= 9.5`
- route-switch trust score: `>= 9.2`
- power and battery-aware route score: `>= 9.0`

## SLO And SLA Framework

### User-Facing SLOs

- CLI cold start p50: `<= 120 ms`
- CLI cold start p95: `<= 250 ms`
- interactive prompt paint after `jini`: `<= 150 ms`
- compact `resume` generation p95: `<= 300 ms`
- offline continuation success rate: `>= 99%`
- cross-surface resume success rate: `>= 99%`
- reconciliation debt visibility rate: `100%`
- route explanation presence when routing changes: `100%`
- battery-aware route-regret rate: `<= 5%`

### Managed Reliability SLOs

- adaptive platform switch success rate: `>= 95%`
- throttle-avoided interruption rate: `>= 85%`
- local capability detection freshness: `>= 99%` within policy window

### SLA Posture

Open/free:

- no hard uptime SLA promise
- strong local/offline continuation as the reliability baseline

Commercial:

- explicit sync and managed recovery SLA
- explicit escalation-path SLA for hosted continuity
- explicit reconciliation backlog clearing SLA after connectivity recovery

## Phased Execution

### Phase 0: Snapshot, Freeze, And Contract Cleanup

- archive free and commercial repos
- remove duplicated doctrine language
- refresh persona catalog and gates
- write the single initiative plan and acceptance tests

### Phase 1: Shell Contract And Simplicity

- keep one stable front door
- remove command/surface drift
- preserve Claude Code, Codex, and GitHub CLI parity habits where useful
- keep first useful result ahead of state narration

### Phase 2: Go Kernel Slice

- implement Go session kernel
- implement Go projection/resume/status engine
- preserve current CLI semantics through native Go handlers and golden tests

### Phase 3: Offline-First Excellence

- finalize local capability probes
- strengthen imported-context reuse
- make debt accrual and reconciliation visible everywhere
- ensure pathless saved-session and saved-snapshot flows stay honest

### Phase 4: Persona Flagship Loops

- engineer: repo review, failing-test repair, spec-to-plan
- college student: research-to-brief, notes-to-study-pack
- high school student: explain-and-rewrite, assignment helper
- realtor: call notes to follow-up, listing/showing summary, offline capture

### Phase 5: Accessibility, Security, And Trust

- screen-reader and keyboard audits
- low-literacy copy audit
- secret-handling and route-proof audits
- publish/write approval hardening

### Phase 6: Commercial Adaptive Orchestration

- managed throttle avoidance
- team continuity
- governed sync
- route forecasting
- hosted recovery convenience

## Acceptance Gates

This initiative does not ship as “done” unless:

- the archive manifest exists and points to verified snapshots
- the Go-core migration path is defined and phaseable
- the four named personas have explicit success outcomes
- offline mode is still first-class
- free/open remains genuinely useful
- commercial value comes from optimization and continuity convenience, not from
  degrading the free tier
- scorecards, SLOs, and SLA posture are written and testable

## Non-Goals

- rewriting every helper in one cut
- inventing a visible multi-agent control plane for normal users
- shipping more UI surfaces before the session kernel is simplified
- making Jini depend on constant connectivity
- making the commercial tier the only trustworthy path
