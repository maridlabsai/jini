# Jini Local Model Support Matrix

Updated: 2026-07-15

Traced to number-one-platform-prd.md §Routing And Resource Policy (v1.5
curated-matrix scope; v1 local execution rides detected third-party runtimes
as a disclosed interim route per §Goals And Scope, Option A).

## P0 Decision

Jini should treat the local model support matrix as a P0 product surface.

This is not a tuning appendix.

It is the concrete implementation of three existing product rules:

- local SLMs should become the cheap-first frontline for ordinary work
- the right local route depends on form factor, device class, and modality
- Jini should improve over time without changing the public product contract

## Purpose

This document defines:

- which local model classes fit which form factors best
- which of those should be adopted in the commercial tier
- how Jini should watch for successor versions and promote them safely

The platform-by-platform offline guarantees, sync semantics, and route policy
are absorbed below in §Absorbed Policies (from the archived
platform-offline-strategy.md).

The product goal is not to chase model brands.

The product goal is to keep one stable Jini route contract while the underlying
local model pool improves.

This document is a specialized local-routing and platform-support matrix, not
the top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this document conflicts with the canonical PRD on priorities, form-factor
commitments, automation posture, or route policy, the canonical PRD wins and
this matrix should be updated.

## Product Rule

Jini should bind product behavior to stable local profile roles, not to one
hard-coded model family.

Stable local profile roles:

- `mobile-small`
- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`
- `workstation-deep`

Users should see the profile and route reason first.

Advanced users may inspect the exact model mapping.

## Shipping Rule

Current public reality:

- the CLI is the only live installable surface today
- desktop and mobile remain public product commitments, not live public installs

This matrix still matters now because:

- CLI local routing already exists
- commercial desktop and mobile planning should not wait for ad hoc model picks
- future app releases should inherit one stable routing policy instead of
  inventing separate per-surface model logic

## Form Factor Matrix

### 1. Android

Primary Jini role:

- offline continuation
- review and approval
- bounded local transforms
- light capture and extraction

Preferred profile:

- `mobile-small`

Recommended primary model classes:

- platform-native on-device model such as Gemini Nano when available
- open-weight mobile class such as Gemma 3n when Jini needs a portable local
  path outside the platform-native stack
- only lightweight fine-tuned models should be eligible for discovered local
  mobile routes

Recommended use:

- summarization
- rewriting
- proofreading
- small extraction
- lightweight voice or image follow-up where the runtime supports it

Avoid treating Android as:

- the default deep-reasoning host
- the default long multi-step coding host

### 2. iPhone and iPad

Primary Jini role:

- offline continuation
- review, triage, approve, defer
- bounded local transforms where the runtime and policy allow them

Preferred profile:

- `mobile-small`

Recommended primary model classes:

- small local transform class only
- open-weight mobile class such as Gemma 3n only when runtime maturity and app
  constraints make it practical
- discovered local routes must reject desktop workhorse or deep models on iOS

Product rule:

- iOS should not be planned as a parity desktop inference surface
- iOS should be planned as the strongest interruption-safe continuation surface

### 3. macOS Laptop, Light SKU

Primary Jini role:

- day-to-day local authoring
- cheap-first drafting
- offline-first desktop work

Preferred profiles:

- `desktop-fast`
- `desktop-workhorse`

Recommended primary model classes:

- `desktop-fast` -> Phi-class small text model
- `desktop-workhorse` -> strong instruct model inside the light laptop envelope
  rather than a pro-sized model

Product rule:

- light laptops should prefer fast and small workhorse models and avoid
  pro-sized local models unless the user explicitly overrides policy

### 4. macOS Laptop, Pro SKU

Primary Jini role:

- day-to-day local authoring
- cheap-first drafting
- multimodal first pass when the local runtime is ready
- offline-first desktop work

Preferred profiles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`
- `workstation-deep` only when measured capability supports it

Recommended primary model classes:

- `desktop-fast` -> Phi-class small text model
- `desktop-workhorse` -> Gemma or strong Qwen-class instruct model in the
  mid-size local envelope
- `desktop-multimodal` -> Gemma-class multimodal model when latency and memory
  are acceptable

Product rule:

- pro laptops may use stronger local models than light laptops, but should not
  silently select workstation-sized models

### 5. Windows Laptop, 16GB To 32GB Class

Primary Jini role:

- day-to-day local authoring
- cheap-first drafting
- multimodal first pass
- offline-first desktop work

Preferred profiles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`

Recommended primary model classes:

- `desktop-fast` -> Phi-class small text model
- `desktop-workhorse` -> Gemma 4 12B or Qwen-class instruct model, chosen by
  measured local score
- `desktop-multimodal` -> Gemma 4 12B or Phi multimodal class when runtime fit
  is stronger

Product rule:

- Windows should stay symmetric with macOS at the routing-policy level even if
  runtime packaging differs

### 6. Workstation Or Strong Desktop

Primary Jini role:

- higher-rigor local drafting
- local critique before paid escalation
- stronger coding and reasoning
- broader multimodal work when the runtime supports it

Preferred profiles:

- `desktop-workhorse`
- `desktop-multimodal`
- `workstation-deep`

Recommended primary model classes:

- `workstation-deep` -> strongest supported local reasoning model such as
  Qwen3 30B-A3B class
- `desktop-multimodal` -> strongest supported local multimodal model that still
  meets startup and memory constraints for normal use

## Commercial Tier Adoption Rule

Jini should adopt these model classes in the commercial tier.

Adoption should happen through a managed local runtime registry, not through
hard-coding one brand into the product contract.

Commercial tier should support:

- downloadable or discoverable local model packs
- profile-to-model mapping per platform
- measured route scoring
- managed fallback when a local route is missing or degraded

Commercial tier should not require:

- bundling all large weights into the installer
- one identical default model across every platform
- users understanding model-brand debates before first success

## Support Tiers

### Tier A: Default Candidates

These should be first canary candidates for supported local profiles:

- mobile-small
  - Gemini Nano class where platform-native
  - Gemma 3n class where open-weight portable
- desktop-fast
  - Phi-class small text model
- desktop-workhorse
  - Gemma 4 12B class
  - Qwen-class mid-size instruct model
- desktop-multimodal
  - Gemma 4 12B class
  - Phi multimodal class where runtime fit is better
- workstation-deep
  - Qwen3 30B-A3B class or strongest supported successor

### Tier B: Experimental Candidates

These may be tested but should not become defaults without measured uplift:

- newly released local MoE variants
- niche runtime-specific forks
- model families without stable local serving paths across the supported hosts

## Admission Criteria

Per number-one-platform-prd.md §Goals And Scope, matrix admission is
license-gated:

- only permissively-licensed, redistribution-safe models are admitted
- model downloads come from official sources only
- the license is shown to the user at consent time, before download

A model that fails any of these is not a candidate, regardless of benchmark
score.

## Registry Contract

Jini should maintain a versioned local model registry with fields including:

- `family`
- `variant`
- `license`
- `profile_role`
- `modalities`
- `form_factor_fit`
- `minimum_device_class`
- `preferred_runtimes`
- `context_window`
- `status`
- `introduced_at`
- `deprecated_at`

Suggested statuses:

- `candidate`
- `canary`
- `supported`
- `deprecated`
- `blocked`

## Promotion Loop

Jini should look for successor versions continuously and promote them through a
fixed loop.

### 1. Watch

Track official release channels for:

- Google Gemma
- Android on-device model stack
- Microsoft Phi
- Qwen

### 2. Ingest

For each new candidate, capture:

- release identifier
- official license
- supported modalities
- recommended runtimes
- stated hardware targets

### 3. Canary

Run the candidate on the same Jini benchmark slices used for local routing:

- intake classification
- follow-up drafting
- checklist shaping
- spec or PRD readiness first pass
- bounded coding support
- multimodal extraction where relevant

### 4. Score

Promotion must consider:

- warm latency
- cold-start cost
- structured-output reliability
- token throughput
- artifact acceptance rate
- edit-distance after generation
- route-regret rate
- crash or transport instability

### 5. Promote

A new model version becomes the new default for a profile only if:

- it is score-positive for that profile and form factor
- it does not regress trust or startup cost beyond the allowed envelope
- it passes the same offline and continuation checks as the current winner

### 6. Deprecate

Older model mappings should move to `deprecated` instead of disappearing
silently, so existing installs and receipts remain explainable.

## Release Cadence

This matrix should run on:

- every monthly release train
- every material local runtime integration update
- every official new-version release from a Tier A model family

## Acceptance Criteria

This P0 is complete only when all are true:

- each major form factor has a preferred local profile mapping
- the commercial tier can express those mappings without hard-coded brand logic
- a versioned registry exists
- a watch and canary loop exists
- successor models can be promoted without rewriting the product contract
- offline continuation and trust surfaces stay stable while local model picks
  evolve underneath them

## Absorbed Policies

These normative rules were merged here from now-archived local-execution docs
during the PRD rebuild (prd-rebuild-design.md §8). Where a source conflicted
with the rebuilt PRD, the PRD wins: "commercially usable local SLM" phrasing
from the old tier doctrine is read as plain local SLM routing — local model
routes are free-tier per the PRD Tier Boundary, with the curated matrix as
v1.5 scope.

### Local SLM frontline (absorbed from local-slm-frontline-policy.md)

Default runtime order: local SLM pool first, stronger paid remote route only
when needed. Jini should not spend frontier-model budget on work a local small
model can complete well enough, and should treat local inference as a routed
pool (`fast`, `workhorse`, `deep`, `multimodal`), not one fixed model.

- Frontline work classes (local first attempt): intake classification, first
  useful pass, follow-up drafting, plan/spec readiness first pass, extraction
  from text-like inputs, summarization, gap detection, checklist shaping,
  rewrite/cleanup without deep external reasoning.
- Escalation work classes: deep critique, architecture review, benchmark or
  exhaustive work, codebase-wide reasoning with stronger correctness
  expectations, stronger tool use or provider-bound execution, unsupported
  modality, policy-constrained cloud routing.
- Runtime decision order: (1) can the local pool handle this well enough,
  (2) which local profile fits, (3) use it; otherwise (4) cheapest suitable
  stronger route, and (5) explicit deep asks prefer the best suitable route.
- Trust readout: when local runs, the user sees `AI route`, `Model`,
  `Local profile`, `Route policy`, and `Why this was chosen`.
- Configuration layers: local SLM mode `off | prefer | require`; profile
  selection `auto | fast | workhorse | deep | multimodal`; one stable local
  transport contract; profile-to-model mapping. Users never need these terms
  before first success.
- Non-goals: pretending local preview is real local inference; exposing
  model-brand debates as the normal experience; requiring a local model
  install before Jini is useful.

### Cross-platform offline strategy (absorbed from platform-offline-strategy.md)

Jini behaves like one work operating system across macOS, Windows, Android,
and iOS. Platform differences are allowed only in interaction density, local
model capacity, offline execution depth, distribution constraints, and review
ergonomics — never in session identity, artifact identity, route evidence,
review/send boundary, offline debt visibility, or sync conflict rules.

Guarantees every shipped surface must preserve:

1. Same work object: every platform acts on the same logical session (stable
   session id, goal, status, current artifact, ready/missing state, next
   action, route evidence, review-safe state, approval boundary, offline and
   sync status).
2. Offline mode is explicit: show offline mode, available local route,
   unavailable remote routes, work that can continue, work that is blocked,
   and reconciliation debt.
3. Local work does not fork the session: offline work appends events to the
   same session timeline — no second transcript, task id, or detached
   artifact family.
4. Route evidence survives sync: after sync the user can inspect which device
   acted, which route and local profile were used, what was generated
   offline, and what was reconciled later.

### Guarantee 4a: Offline And Online Toggle Seamlessly

Offline and online are route states inside one session, not separate
products. The same session timeline must stitch together:

- local model work performed offline
- queued approvals or annotations captured on mobile
- downstream CLI work resumed online
- managed-route recovery after throttling or provider limits
- sync and reconciliation events after connectivity returns

Cross-navigation must preserve the same current artifact, next action, route
evidence, device capability state, battery or thermal posture, online
capability state, configured CLI throttle state, and offline debt.

Further guarantees: mobile is not desktop parity (excellent at continuation,
review, approval, defer, capture, light transforms — not deep coding or large
local inference); desktop (macOS/Windows) is the offline authoring host with
deeper local profiles when the machine supports them.

Route policy is unified across platforms. The route decision considers task
shape, modality, risk, user preference, device class, local profile
availability, local runtime health, battery and thermal envelope, offline
state, provider availability, online CLI throttle level and quota pressure,
downstream CLI route availability, and prior route regret.

- Local-first rule: cheapest suitable local route when a local profile can
  satisfy the task reliably at acceptable risk and no stronger route is
  pinned.
- Escalation rule: escalate or hand off when the local profile is
  unavailable, local latency makes the route expensive in practice, required
  modality is missing locally, task risk requires stronger reasoning, a
  connector write requires online capability, or the user asks.
- Mobile handoff rule: mobile hands off rather than overruns its role for
  long-running generation, complex artifact edits, weak local profiles, or
  battery/thermal/memory pressure.

Sync semantics: sync events, not raw transcripts. The core sync object is the
session envelope, event log, artifact metadata and versions, route evidence,
offline debt, and conflict markers. Merge by session id; preserve every event
with device id and timestamp; rebuild projection after merge; never discard
route evidence; never silently overwrite the current artifact; require user
review when two devices edited the same artifact version. Offline debt is
visible whenever a connector write is queued, hosted sync is incomplete, an
approval targets an older artifact version, route evidence is incomplete, or
a merge conflict exists.

Shipping prerequisites: desktop ships only with a shared session envelope,
durable local artifact store, correct offline event append, inspectable route
evidence, visible sync reconciliation, device-aware profile selection, and
honest preview posture. Mobile ships only with shared session identity,
offline latest-ready artifact, offline-surviving review/approval/defer/
annotation events, visible pending sync, stale-approval detection before
send, and obvious handoff when mobile capacity is too small.

### Future Update Policy

Jini improves local capability through the registry and canary loop above,
not through platform-specific product rewrites.

Future model updates should:

- map to stable profile roles
- run the same offline and continuation checks as current defaults
- preserve route evidence shape
- preserve session and artifact identity
- deprecate old mappings explicitly

Future app updates should keep CLI, desktop, and mobile bound to one session
graph, make offline debt more visible, reduce handoff cost, improve local
route selection through measured evidence, and avoid platform-specific
session semantics.

### Device runtime gate (absorbed from device-runtime-gate.md)

The independent gate for device-aware local runtime routing — separate from
publish readiness because capability routing can drift without breaking the
main product surface. Categories, each of which must be proven:

1. Capability probe: OS, OS version, CPU architecture, memory, accelerator,
   and local runtime class detection all exist in code.
2. Versioned cache: repo-local device profile path; Jini version, capability
   registry version, and capture timestamp recorded; freshness/re-probe
   logic; profile invalidates on OS/runtime/endpoint drift, not only time.
3. Routing use: device class reaches route features; route scoring includes a
   device capability bias; profile availability can downgrade or block
   expensive local routes and reflects backend readiness, not only hardware.
4. Transparency: provider doctor exposes device class, accelerator class, and
   local runtime class.
5. Tests: deterministic device-class override path; device-aware route
   selection tests; device-aware provider doctor tests.

Gate command: `jini validate-device-runtime-gate --format json`. The gate
fails if any category fails.
