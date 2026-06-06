# Number One Development Plan

Updated: 2026-06-05

This document is a supporting development plan, not the top-precedence product
and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

The current competitor-derived release pressure lives in
[competitive-release-plan.md](./competitive-release-plan.md).

If this development plan conflicts with the canonical PRD on priorities,
requirements, roadmap order, automation posture, or app-shipping order, the
canonical PRD wins and this plan should be updated before execution continues.

## Goal

This is Jini's only mission until all scores declare a win:

- make Jini number 1 in its category
- make the CLI and apps feel like quantum jumps, not gradual polish
- make the framework self-learning and self-correcting enough that it needs
  fewer user-facing updates than fast-moving competitors

## Operating Cadence

Jini should separate:

- fast learning cadence
- medium policy cadence
- slower product-surface cadence

Target cadence:

- daily benchmark and trace ingestion
- weekly competitor watch and shadow evaluation
- weekly policy-bundle promotion if score-positive
- monthly public release train
- quarterly structural simplification review

## Workstreams

### Workstream A: Canonical Session And Artifact Graph

Deliverables:

- one canonical session graph
- one artifact graph
- cross-surface current-focus recovery
- one continuation model across CLI and apps

Score impact:

- delivery-maturity
- memory-reliability
- token-efficiency

Exit criteria:

- no surface asks users to reconstruct current work manually
- continue, open, and review work identically across live and degraded paths

### Workstream B: Environment And Workflow Learning

Deliverables:

- environment fingerprinting
- learned repo and workflow profiles
- repeated-task compression
- reusable automation suggestions with explicit approval and rollback

Score impact:

- memory-reliability
- delivery-maturity
- advanced-set-breadth

Exit criteria:

- repeated work is detected and compressed automatically
- future tasks need less user re-explanation

### Workstream C: Self-Correction And Contract Repair

Deliverables:

- docs/help/runtime parity repair
- route-regret detection
- benchmark regression blocking
- stale-evidence and stale-memory repair flows
- automatic UX contract checks on flagship flows

Score impact:

- delivery-maturity
- token-efficiency
- workflow-rigor

Exit criteria:

- product drift is caught before release
- score regressions are blocked by evidence, not noticed after shipping

### Workstream D: Upstream Quality Automation

Deliverables:

- readiness preflight
- missing-proof detection
- missing-review detection
- action-risk classification
- quality gate suggestions before execution or handoff

Score impact:

- delivery-maturity
- workflow-rigor
- memory-reliability

Exit criteria:

- flagship flows catch quality misses before users or reviewers do

### Workstream E: CLI And App Surface Specialization

Deliverables:

- CLI as the fastest universal control surface
- desktop as the rich review and artifact-edit surface
- mobile as the continuity, approval, and interruption surface
- parity fixtures proving they are views over one session graph

Score impact:

- delivery-maturity
- token-efficiency
- advanced-set-breadth

Exit criteria:

- CLI, desktop, and mobile feel different in strengths but identical in work identity

### Workstream F: GitHub-Native System Of Record

Deliverables:

- issue to implementation continuity
- PR review and merge continuity
- release and verification continuity
- idempotent publish and rollback semantics

Score impact:

- adapter-portability
- delivery-maturity
- workflow-rigor

Exit criteria:

- Jini beats raw shell plus `gh` on continuity, trust, and review cost in benchmarks

### Workstream G: Score And Competitive Operations

Deliverables:

- stronger benchmark harness
- competitor-watch review loop
- score movement evidence
- shadow-policy comparisons against archived traces
- source-backed competitor release packet covering terminal agents, IDE agents,
  cloud PR agents, local/offline hosts, routing gateways, app builders, and
  general workflow agents
- release-plan deltas that explicitly say what Jini will copy, integrate,
  avoid, or defer

Score impact:

- all replacement-critical dimensions

Exit criteria:

- every score change is backed by measured evidence
- the overall lead margin stays above target instead of oscillating

## Phased Plan

### Phase 0: Score Truth And Benchmark Control

Objective:

- make every score traceable to runtime evidence

Deliverables:

- benchmark coverage for flagship flows
- measured cross-surface recovery metrics
- measured route-regret and resume-cost metrics
- automated competitor-watch packet
- expanded competitor watch categories from the competitive release plan
- canonical commit, push, and release gate matrix with one checked-in runner

Required score movement:

- stabilize `token-efficiency`
- stabilize `delivery-maturity`

### Phase 1: Session Graph And Artifact-First Continuation

Objective:

- make Jini feel like one persistent work object everywhere

Deliverables:

- canonical session graph
- canonical current focus
- artifact-first continue/open/show
- app surfaces bound to the same session graph

Required score movement:

- `delivery-maturity -> 9.0`
- `memory-reliability -> 9.0`

### Phase 2: Workflow Learning And Upstream Quality

Objective:

- make Jini reduce user work and prevent rework

Deliverables:

- learned environment patterns
- learned workflow templates
- P0 local model support matrix by form factor and profile role
- successor-model watch and canary promotion loop for the local SLM pool
- readiness and quality automation
- repeated-task compression suggestions

Required score movement:

- `memory-reliability -> 9.0`
- `workflow-rigor` margin expansion

### Phase 3: GitHub-Native Execution And Review

Objective:

- make Jini the best front door for engineering work, not just another shell

Deliverables:

- issue, PR, review, and release continuity
- review-ready artifacts
- rollback-safe automation receipts
- policy-aware GitHub action surfaces

Required score movement:

- `adapter-portability -> 9.0`
- `delivery-maturity` lead expansion

### Phase 4: Self-Correcting Policy Engine

Objective:

- reduce the need for constant product shipping

Deliverables:

- shadow-policy evaluation
- safe policy promotion
- route-regret repair
- docs/help/runtime contract regeneration
- automated competitor-delta review

Required score movement:

- `token-efficiency -> 9.0`
- overall margin growth

### Phase 5: Score-Declare Win

Objective:

- prove Jini is number 1 by benchmark, not by narrative

Deliverables:

- all replacement-critical scores at `9.0+`
- overall lead margin above `0.8`
- published proof of flagship-flow wins
- sustainment loop that keeps the scores there

## Kill List

Do not spend cycles on:

- broadening demo flows before flagship dominance
- adding public commands instead of collapsing them
- manual setup burdens that learning can remove
- new app surfaces that are not bound to the canonical session graph
- high-ceremony UX improvements that do not reduce user work
- connector breadth before GitHub-native depth is excellent

## Sustainability Model

Jini should keep pace with competitors through:

- automated competitor research refresh
- benchmark replays over archived traces
- policy updates before code releases
- generated help and docs from canonical command/state schemas
- score-gated promotion and rollback

The framework should get better because it learns and corrects itself, not
because humans race to ship visible changes every week.

## Final Exit Gate

The mission is complete only when:

- all replacement-critical scores declare `9.0+`
- the overall score lead is durable
- flagship user outcomes are measurably better than direct competitor paths
- the framework can keep pace through self-learning and self-correction without
  relying on constant product churn
