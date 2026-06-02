# Cross-Surface Session System And Developer Design

Updated: 2026-05-21

## Purpose

This document translates the product charter into a buildable system design.

It exists to answer:

- what must be built for Jini to be cost-optimizer first
- what must be built for Jini UX to feel second to none
- what must be built for the same session to survive across macOS, Windows,
  mobile, and CLI

This is not a vision memo. It is the implementation contract that engineering,
UX, and product should build against.

Read this with:

- [cross-surface-session-platform-prd.md](./cross-surface-session-platform-prd.md)
- [lean-platform-doctrine.md](./lean-platform-doctrine.md)
- [engineering-principles.md](./engineering-principles.md)
- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)
- [work-state-machine.md](./work-state-machine.md)
- [artifact-schemas.md](./artifact-schemas.md)
- [protocol-core.md](./protocol-core.md)

## Design Goals

### 1. One Session Everywhere

Jini must preserve one logical session object across:

- macOS
- Windows
- mobile
- CLI

### 2. Cheapest Suitable Continuation

Resume must prefer reuse over rebuild:

- reuse session state before regenerating context
- reuse artifacts before re-summarizing source material
- reuse local route state before escalating to premium routes

Continuation should be cheaper than restart.

### 3. Surface Parity Through Shared Semantics

Surfaces may differ visually, but they must not differ semantically.

The meaning of:

- current session
- ready
- missing
- next
- route evidence
- review-safe state

must remain stable across every surface.

### 4. Inspectable Trust

Users must be able to inspect:

- what is stored
- what was reused
- what route was chosen
- what is still missing
- what is safe to review or share

### 5. Cheap, Plain, Portable Core

The session system should:

- stay local-first
- be plain-file inspectable
- sync cleanly when hosted continuity exists
- avoid binding product truth to one runtime or UI stack

## Non-Goals

This design does not attempt to:

- make mobile the primary authoring surface
- add decorative multi-agent orchestration
- create a provider-specific session model
- require always-on hosted infrastructure for continuity
- hide storage, route, or trust state behind a proprietary black box

## System Overview

Jini should be implemented as one session platform with four layers:

1. session kernel
2. artifact and event persistence
3. routing and cost engine
4. surface adapters

```text
Surfaces
- CLI
- macOS desktop
- Windows desktop
- mobile

Surface adapters
- session list
- session detail
- resume / continue
- review / share

Session kernel
- canonical session object
- state projection
- readiness / missing / next derivation
- resume contract

Persistence
- session envelope
- artifact store
- append-only event log
- sync metadata

Routing and cost engine
- cheapest suitable route selector
- continuation reuse scorer
- route evidence recorder
- premium escalation policy
```

## Developer Structure Rules

The system should be implemented with explicit SOLID and OOP boundaries.

Required shape:

- session store uses a storage-focused contract and does not own routing
- projection code derives view state and does not mutate persistence directly
- routing uses strategy-style policy objects instead of giant mode switches
- surface integration stays behind adapter-style boundaries
- bundle and manifest construction uses factory-style builders
- top-level user flows use facade-style orchestration over narrower services

Implementation defaults:

- composition over inheritance
- narrow interfaces over wide utility objects
- immutable or value-style objects for contracts wherever possible
- no hidden globals as the source of session or route truth

## Canonical Session Object

The session object is the primary product object.

### Session Envelope

Minimum fields:

```yaml
session_id: string
session_version: integer
title: string
goal: string
status: active|waiting|blocked|ready_for_review|done|archived
created_at: timestamp
updated_at: timestamp
last_active_surface: cli|macos|windows|mobile|unknown
last_actor: user|jini|system
current_artifact_id: string|null
current_step: string|null
next_step: string|null
review_safe: boolean
share_boundary: draft_only|review_safe|send_requires_approval|sent
```

### Session Projection

These are derived fields shown to users everywhere:

```yaml
ready:
  - artifact refs ready to open now
missing:
  - unresolved blockers, proofs, approvals, or gaps
assumptions:
  - still-active assumptions that matter
route:
  provider_id: string|null
  route_kind: local|byo_remote|hosted|unknown
  model: string|null
  effort: low|standard|high|unknown
  reason: string|null
connectivity_mode: offline|online|degraded|unknown
online_capability: available|unavailable|unknown
reconciliation_debt_count: integer
reconciliation_summary: string|null
cost_posture:
  current_path: cheap|balanced|premium|unknown
  continuation_saved_work: string|null
```

### Relationship To WorkUnit

The session object is the user-facing shell object.

It may reference one or more protocol-level WorkUnits, but the user-facing
surface should not require protocol fluency.

Rules:

- one session may wrap one primary WorkUnit and supporting sibling work
- the session owns surface continuity and resume semantics
- the WorkUnit model remains the protocol truth for guarded transitions

## Persistence Model

### Local State Root

The canonical local state root remains:

- `.jini/`

Add a stable session tree:

```text
.jini/
  sessions/
    index.json
    <session-id>/
      session.yaml
      projection.json
      events.ndjson
      artifacts/
      attachments/
      sync.json
      route.json
```

### Files

#### `session.yaml`

Source-of-truth session envelope with minimal derived state.

#### `projection.json`

Computed view optimized for fast surface reads.

Contents:

- ready
- missing
- next
- route summary
- trust summary
- artifact summary

#### `events.ndjson`

Append-only event stream.

Examples:

- session_created
- input_added
- artifact_created
- artifact_superseded
- route_chosen
- route_escalated
- review_state_changed
- resumed_on_surface
- sync_merged
- went_offline
- came_online
- reconciliation_debt_accrued
- reconciliation_debt_cleared
- external_framework_context_imported

#### `sync.json`

Sync state, not product truth.

Fields:

- last_local_clock
- last_sync_clock
- last_online_at
- online_capability_status
- reconciliation_debt_count
- reconciliation_debt_summary
- remote_session_etag
- conflict_state

#### `route.json`

Latest route evidence plus continuation-savings data.

## Event Model

The event log is the cheapest portable truth for continuity.

### Event Requirements

Every event must include:

```yaml
event_id: string
session_id: string
event_type: string
emitted_at: timestamp
actor: user|jini|system
surface: cli|macos|windows|mobile|unknown
payload: object
```

### Why Events Matter

They enable:

- resume reconstruction
- sync merge
- trust inspection
- cost analysis
- interruption recovery analysis
- offline debt reconciliation
- imported framework-context reuse while offline

## Resume Contract

`resume` must mean the same thing on every surface.

### Input Contract

Resume should accept:

- explicit session id
- implicit current session when there is one obvious active session
- chooser flow when multiple active sessions exist

### Output Contract

Every surface resume action must return:

- current goal
- latest ready artifacts
- missing blockers
- next action
- route evidence summary
- visible offline/online mode
- reconciliation debt summary when non-empty
- review/share boundary

### Resume Algorithm

1. resolve target session
2. load `session.yaml`
3. load `projection.json`
4. if projection is stale, rebuild from `events.ndjson` and current artifacts
5. emit `resumed_on_surface`
6. show the same semantic frame on the current surface
7. if offline, say so explicitly and surface any reconciliation debt

### Continuation Savings Logic

Before any regeneration:

1. check for reusable ready artifacts
2. check for unresolved missing items
3. check for reusable route/context state
4. decide whether reuse is enough
5. only then consider new generation or route escalation

### Offline Continuity Rules

When online capability is unavailable or degraded:

- Jini should state that it is working in offline mode
- Jini should prefer locally available route, artifact, and session reuse
- Jini may reuse locally stored context imported from Claude Code, Codex, or
  GitHub CLI if that context has already been attached to the current session
- Jini should record reconciliation debt for work that must be synced,
  published, reconciled, or replayed once online capability returns
- Jini should surface that debt before the user is surprised by delayed sync

## Routing And Cost Engine

### Design Principle

The routing engine should optimize continuation cost, not just one-turn cost.

### Inputs

- session state
- current artifact freshness
- route evidence
- local runtime availability
- provider availability
- online capability status
- reconciliation debt pressure
- imported framework context availability
- task class
- risk class
- interruption recovery value

### Outputs

- chosen route
- route reason
- cheaper fallback
- stronger fallback
- continuation savings explanation
- offline mode status
- reconciliation debt update

### Continuation Reuse Scorer

Add a scorer that estimates whether reusing session state is cheaper than
rebuilding.

Factors:

- artifact freshness
- unresolved blockers count
- amount of prior structured context
- local cache availability
- previous route quality
- expected premium token cost if rebuilt

### Route Evidence Recording

Each routed step should record:

```yaml
route_id: string
provider_id: string
model: string
effort: string
route_kind: local|byo_remote|hosted
reason: string
expected_cost_band: cheap|balanced|premium
reused_context: boolean
reused_artifacts: [string]
avoided_rebuild_summary: string|null
```

## Surface Adapter Design

Surfaces are projections over the same session kernel.

### Shared Surface Contract

Every surface adapter must implement:

- list sessions
- open session
- resume session
- open ready artifact
- show missing
- show route evidence
- show next action

### CLI Adapter

Responsibilities:

- fastest cold start
- pathless current-session handling
- compact text rendering
- open artifact handoff to filesystem/editor

Must not:

- require users to know session storage paths
- become the only place where session truth is understandable

### Desktop Adapter

Responsibilities:

- session shelf
- active session detail
- artifact preview
- missing/next panel
- route/trust inspection panel

Must preserve:

- same labels
- same session ids
- same ready/missing meanings

### Mobile Adapter

Responsibilities:

- session list
- latest artifact preview
- approve / defer / continue-lightly flows
- blocker visibility

Must optimize for:

- quick review
- low text density
- no context reconstruction

## Sync And Identity Design

### Local-First Rule

Local state remains the source of truth for free/BYO use.

Hosted sync is a convenience layer, not the only continuity path.

### Sync Unit

The session is the sync unit.

Artifacts and events sync beneath the session id.

### Merge Strategy

Prefer event merge, then rebuild projection.

Rules:

- append-only events merge by clock + event id
- envelope conflicts resolve by newest valid state transition
- artifact conflicts produce sibling revisions, not silent overwrite
- review/send boundary conflicts must never auto-resolve to a riskier state
- reconciliation debt clears only after the required online reconciliation
  action actually succeeds

### Identity

Every synced session needs:

- local stable session id
- optional account-scoped remote id
- exportable session package format

### External Framework Context Import

Jini should be able to reuse locally available context and memory captured in
the current session under other supported framework surfaces when the user goes
offline.

Supported initial imports:

- Claude Code
- Codex
- GitHub CLI

Rules:

- imported context must be attached to the canonical Jini session object, not
  stored as a parallel session model
- imported context must remain inspectable and removable
- imported context may improve offline continuation, but it must not silently
  override Jini artifacts or route evidence

## Security And Privacy

### Storage Rules

Do store:

- artifacts
- route evidence
- review/send boundary
- missing blockers
- sync metadata

Do not store as hidden product magic:

- unbounded private conversation memory
- opaque route decisions with no evidence
- silent send/share actions

### Secret Handling

Secrets must remain outside session truth.

Sessions may reference:

- provider configured
- provider unavailable

They must not store:

- raw tokens
- secret values

## Module Design

### Proposed Python Package Layout

```text
tools/
  jini.py
  session_core.py
  session_store.py
  session_projection.py
  session_events.py
  session_sync.py
  route_engine.py
  route_evidence.py
  surface_contract.py
```

### Module Responsibilities

#### `session_core.py`

- canonical session types
- invariants
- transition helpers

#### `session_store.py`

- load/save session envelope
- artifact path helpers
- index maintenance

#### `session_projection.py`

- derive ready/missing/next
- derive trust summary
- derive cross-surface view state

#### `session_events.py`

- append/read events
- event validation
- replay helpers

#### `session_sync.py`

- sync metadata
- merge logic
- conflict modeling

#### `route_engine.py`

- cheapest suitable route logic
- continuation reuse scorer

#### `route_evidence.py`

- serialize route decisions
- build user-facing evidence summaries

#### `surface_contract.py`

- shared shape that CLI, desktop, and mobile adapters must satisfy

## Data Flow

### New Session

```text
input
-> scope
-> create session envelope
-> emit session_created
-> write first artifact(s)
-> compute projection
-> surface renders first useful result
```

### Resume Session

```text
resume request
-> resolve session
-> load envelope + projection
-> refresh projection if needed
-> compute continuation reuse
-> render current session
-> only generate more if the user continues
```

### Cross-Surface Sync

```text
local event append
-> sync transport
-> remote merge by event log
-> projection rebuild
-> other surface sees same session state
```

## UX Contract For Developers

Developers must preserve these UX invariants:

1. first useful result before diagnostic detail
2. one obvious next action
3. missing state always visible when it matters
4. route evidence inspectable but not noisy by default
5. resume cheaper and simpler than restart
6. session identity stable across surfaces

## Testing Strategy

### Unit Tests

- session envelope validation
- projection derivation
- event append/replay
- route reuse scorer
- merge conflict handling

### Integration Tests

- create session on CLI, resume on desktop contract
- create session on desktop contract, review on mobile contract
- reuse artifacts instead of regenerating
- route escalation only when justified

### Regression Tests

- session id stability
- ready/missing parity across surfaces
- review/send boundary preservation
- route evidence preservation after sync

### Acceptance Tests

Minimum acceptance scenarios:

1. start on CLI, resume on mobile, continue on desktop
2. start on desktop, review on mobile, finish on CLI
3. interrupted work resumes with lower cost than a fresh rebuild
4. a user can inspect what changed and what is still missing at every step

## Delivery Sequence

### Phase 1

- implement canonical session envelope
- implement local session store
- implement event log
- implement projection builder

### Phase 2

- rebind CLI to session kernel
- add explicit resume contract
- record continuation savings

### Phase 3

- implement surface adapter contract
- ship first non-CLI session list/detail surface
- validate parity labels and behavior

### Phase 4

- add sync layer
- add conflict handling
- expose cross-surface resume proof

## First Runtime Slice

The first runtime slice must prove the charter with the smallest possible
kernel. It must not attempt to ship the whole platform plan.

### Slice 1 Goal

Make the CLI speak the canonical session model and prove that continuation is
cheaper than restart.

### Slice 1 Files

Only these new modules should be introduced first:

- `tools/session_core.py`
- `tools/session_store.py`
- `tools/session_projection.py`
- `tools/session_events.py`

The first slice should reuse the current routing surface and artifact rendering
where possible instead of rebuilding everything at once.

### Slice 1 Data

Only these session fields are mandatory in slice 1:

- `session_id`
- `title`
- `goal`
- `status`
- `updated_at`
- `current_artifact_id`
- `next_step`
- `review_safe`
- `share_boundary`

Only these derived projection fields are mandatory in slice 1:

- `ready`
- `missing`
- `next`
- `route.provider_id`
- `route.reason`
- `cost_posture.current_path`
- `cost_posture.continuation_saved_work`

### Slice 1 Commands

Only these CLI flows need to be rebound first:

- `jini`
- `jini status`
- `jini resume`
- `jini open`

`status` and `resume` should become the first public proof that the session
kernel is real.

## Out Of Scope For Slice 1

Do not build these in the first runtime slice:

- hosted sync transport
- account identity
- merge conflict UI
- mobile-native UI
- Windows-native UI
- macOS-native UI
- broad pack/runtime refactors outside session binding
- multi-session collaboration
- advanced route optimization beyond basic continuation reuse evidence

Slice 1 should prove the kernel, not the full product surface.

## Migration From Current Runtime

The new session kernel must replace, not sit beside, the current ad hoc
current-work truth.

### Current Runtime Inputs

The migration layer should read from existing runtime truth where available:

- current-work pointers
- artifact pointers
- work-unit metadata
- route evidence already emitted by the runtime

### Migration Strategy

1. on first session-kernel access, detect whether canonical session state exists
2. if not, synthesize a session envelope from current-work/runtime metadata
3. write canonical session state under `.jini/sessions/<session-id>/`
4. build `projection.json`
5. emit a `session_migrated` event
6. continue serving `status`, `resume`, and `open` from the canonical session

### Migration Rules

- never silently discard existing current-work pointers
- never force users to re-open or re-scope active work
- prefer one migrated session over multiple guessed sessions
- if migration confidence is low, keep the session usable and surface the
  ambiguity as `missing`, not as a hard failure

## Acceptance Proof For Slice 1

Slice 1 is done only when these proofs pass:

### Runtime Proofs

1. `jini` creates or continues a canonical session
2. `jini status` reads from canonical session state
3. `jini resume` reopens the same session without path reconstruction
4. `jini open` opens the ready artifact from canonical session state
5. continuation evidence shows reuse before rebuild when prior work exists

### Regression Proofs

1. existing example flows still work
2. current artifact opening still works
3. pathless status still works
4. route evidence still renders

### Developer Proofs

1. session state is inspectable on disk under `.jini/sessions/`
2. event replay can rebuild projection state
3. the first slice does not require hosted infrastructure
4. the first slice does not add surface-specific workflow drift

## Definition Of Done

The system design is realized only when:

- the same session id survives across supported surfaces
- ready, missing, and next are semantically identical across surfaces
- continuation reuses prior work before buying more intelligence
- route evidence remains inspectable after resume and sync
- users can switch devices without reconstructing context
