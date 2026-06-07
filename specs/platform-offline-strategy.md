# Platform Offline Strategy

Updated: 2026-06-05

This document is a specialized platform offline strategy, not the
top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this strategy conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, automation posture, app-shipping order, or route
policy, the canonical PRD wins and this strategy should be updated.

Read alongside:

- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)
- [app-platform-shipping-playbook.md](./app-platform-shipping-playbook.md)
- [cross-surface-session-platform-prd.md](./cross-surface-session-platform-prd.md)
- [cross-surface-session-system-and-dev-design.md](./cross-surface-session-system-and-dev-design.md)
- [device-capability-routing.md](./device-capability-routing.md)
- [local-model-support-matrix.md](./local-model-support-matrix.md)
- [local-slm-frontline-policy.md](./local-slm-frontline-policy.md)
- [runtime-selection-heuristics.md](./runtime-selection-heuristics.md)

## Purpose

Jini needs one concrete offline strategy for:

- macOS
- Windows
- Android
- iOS

The strategy must define:

- exact offline guarantees
- local-model expectations
- sync semantics
- route policy
- shipping prerequisites
- future update behavior

The goal is not to make every platform identical.

The goal is to make every platform read and write the same work object while
using the best local capability that form factor can support.

## Product Rule

Jini should behave like one work operating system across platforms.

Platform differences are allowed only in:

- interaction density
- local model capacity
- offline execution depth
- distribution constraints
- review and approval ergonomics

Platform differences are not allowed in:

- session identity
- artifact identity
- route evidence
- review and send boundary
- offline debt visibility
- sync conflict rules

## Cross-Platform Guarantees

Every shipped surface must preserve these guarantees.

### Guarantee 1: Same Work Object

Every platform must act on the same logical session object.

Minimum state:

- stable session id
- title
- goal
- current status
- current artifact
- ready state
- missing state
- next action
- route evidence
- review-safe state
- approval and send boundary
- offline and sync status

### Guarantee 2: Offline Mode Is Explicit

When the device is offline or remote capability is unavailable, Jini must show:

- offline mode
- available local route
- unavailable remote route or connector
- work that can continue now
- work that is blocked until reconnection
- reconciliation debt that will need sync

### Guarantee 3: Local Work Does Not Fork The Session

Offline work must append events to the same session timeline.

It must not create a second hidden transcript, second task id, or detached
artifact family.

### Guarantee 4: Route Evidence Survives Sync

After sync, the user must still be able to inspect:

- which device acted
- which route was used
- which local profile or model class was used
- what was generated offline
- what was reconciled later

### Guarantee 4a: Offline And Online Toggle Seamlessly

Jini must treat offline and online as route states inside one session, not as
separate products.

The same session timeline must stitch together:

- local model work performed offline
- queued approvals or annotations captured on mobile
- downstream CLI work resumed online
- managed-route recovery after throttling or provider limits
- sync and reconciliation events after connectivity returns

Cross-navigation must preserve the same current artifact, next action, route
evidence, device capability state, battery or thermal posture, online
capability state, configured CLI throttle state, and offline debt.

### Guarantee 5: Mobile Is Not Desktop Parity

Mobile should be excellent at continuation, review, approval, defer, capture,
and light transforms.

Mobile should not be planned as the default host for deep coding, long
multi-step agent loops, or large local multimodal inference.

### Guarantee 6: Desktop Is The Offline Authoring Host

macOS and Windows should be planned as the primary offline authoring,
inspection, and artifact-editing hosts.

Desktop should support deeper local model profiles than mobile when the
machine can run them reliably.

## Platform Matrix

| Platform | Primary role | Offline depth | Local profile expectation | Distribution posture |
| --- | --- | --- | --- | --- |
| macOS | desktop authoring, review, artifact editing, supervision | deep local work when hardware allows | `desktop-fast`, `desktop-workhorse`, `desktop-multimodal` | direct-first |
| Windows | desktop authoring, review, artifact editing, supervision | deep local work when runtime stack allows | `desktop-fast`, `desktop-workhorse`, `desktop-multimodal` | direct-first |
| Android | continuation, review, approval, capture, light local transforms | bounded local work | `mobile-small` | direct-first where policy allows, store secondary |
| iOS | continuation, review, approval, defer, interruption recovery | bounded local work | `mobile-small` when practical | App Store constrained |

## macOS Strategy

### Product Role

macOS is the first-class local desktop authoring host.

It should support:

- artifact inspection
- artifact editing
- session continuation
- local-first drafting
- local multimodal first pass when the machine supports it
- deeper review before paid escalation

### Offline Guarantees

macOS must support offline:

- open current session
- inspect artifacts
- edit local artifacts
- continue with attached local context
- run local model profiles when configured
- record route evidence and offline events
- show reconciliation debt before sync

macOS may block offline:

- remote connector calls
- hosted sync
- hosted managed-route switching
- actions that require live provider or account state

### Local Model Expectations

Supported profile roles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`

Expected model classes:

- `desktop-fast` uses Phi-class small text models or successors with better
  measured latency and reliability.
- `desktop-workhorse` uses Gemma 4 12B class or Qwen-class instruct models
  when they win on the actual machine.
- `desktop-multimodal` uses Gemma 4 12B class or a stronger successor when
  local multimodal score is positive.

macOS should prefer platform-native acceleration where it improves latency,
privacy, install friction, or reliability.

### Sync Semantics

macOS should keep a local session store and append events while offline.

When online capability returns:

- upload local event log
- merge by session id and event order
- rebuild projection
- preserve route evidence
- flag conflicts that cannot be auto-merged

### Route Policy

Default route order:

1. local profile that can satisfy the task
2. BYO provider route when configured and cheaper or higher quality
3. managed remote route when policy, quality, or availability justifies it

Escalate visibly when:

- local model lacks required modality
- local output fails structured reliability checks
- task risk exceeds local profile confidence
- user explicitly requests stronger route

## Windows Strategy

### Product Role

Windows should be symmetric with macOS at the product-policy level.

It should support:

- artifact inspection
- artifact editing
- session continuation
- local-first drafting
- local multimodal first pass when the runtime stack supports it
- route evidence that is comparable to macOS

### Offline Guarantees

Windows must support offline:

- open current session
- inspect artifacts
- edit local artifacts
- continue with attached local context
- run supported local model profiles when configured
- record route evidence and offline events
- show reconciliation debt before sync

Windows may block offline:

- remote connector calls
- hosted sync
- hosted managed-route switching
- actions that require live provider or account state

### Local Model Expectations

Supported profile roles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`

Expected model classes:

- `desktop-fast` uses Phi-class small text models or successors.
- `desktop-workhorse` uses Gemma 4 12B class or Qwen-class instruct models
  chosen by measured local score.
- `desktop-multimodal` uses Gemma 4 12B class, Phi multimodal class, or a
  successor that fits the local runtime stack better.

Windows should not expose a different product contract just because packaging
or acceleration differs from macOS.

### Sync Semantics

Windows should use the same session envelope and event-log semantics as macOS.

When online capability returns:

- sync local events by session id
- merge artifacts by artifact id and version
- preserve device-specific route evidence
- show conflicts before overwriting user-visible artifacts

### Route Policy

Default route order:

1. local profile that can satisfy the task
2. BYO provider route when configured and appropriate
3. managed remote route when policy, quality, or availability justifies it

Windows should choose local profiles from measured device capability, not from
OS name alone.

## Android Strategy

### Product Role

Android is a continuation, review, approval, capture, and lightweight local
transform surface.

It should support:

- quick session review
- approval and defer
- lightweight continuation
- note or voice capture where available
- small local transforms
- offline triage of existing artifacts

### Offline Guarantees

Android must support offline:

- view recent synced sessions
- view latest ready artifact
- view missing state
- approve, defer, or annotate locally
- capture new input into the session queue
- run bounded local transforms when a mobile profile is available
- show pending sync and reconciliation debt

Android may block offline:

- deep generation
- long coding sessions
- large multimodal transforms
- connector writes
- account and entitlement changes

### Local Model Expectations

Supported profile role:

- `mobile-small`

Expected model classes:

- platform-native on-device model class when available
- Gemma 3n class or successor when Jini needs a portable open-weight path
- small text, voice, or image transform model only when latency and battery
  stay inside the mobile envelope

Android should not be the default deep reasoning host.

### Sync Semantics

Android should cache a bounded local projection of each recent session.

Offline events should include:

- review action
- approval action
- defer action
- annotation
- captured input
- local transform output

When online capability returns:

- upload mobile events
- reconcile with desktop or CLI changes
- mark any conflicting approval or artifact edit for review
- keep the original mobile action visible in history

### Route Policy

Default route order:

1. local `mobile-small` route for bounded transforms
2. defer to desktop or CLI local route for heavier work
3. managed remote route only when online and justified

Android should offer a handoff action when the task exceeds mobile-local
capacity.

## iOS Strategy

### Product Role

iOS is an interruption-safe continuation and review surface.

It should support:

- quick review
- approval and defer
- lightweight continuation
- artifact inspection
- offline annotations
- capture into a session queue

It should not be planned as a desktop-equivalent local inference host.

### Offline Guarantees

iOS must support offline:

- view recent synced sessions
- view latest ready artifact
- view missing state
- approve, defer, or annotate locally
- capture new input into the session queue
- show pending sync and reconciliation debt

iOS may block offline:

- deep generation
- large local multimodal inference
- connector writes
- account and entitlement changes
- managed remote route switching

### Local Model Expectations

Supported profile role:

- `mobile-small`

Expected model classes:

- small local transform class only when runtime maturity, battery behavior, and
  app constraints make it practical
- Gemma 3n class or successor only if the measured mobile envelope is strong
  enough for the specific bounded task

iOS value should come first from continuity, trust, and review ergonomics.

### Sync Semantics

iOS should cache a bounded local session projection and append offline review
events.

When online capability returns:

- upload review and annotation events
- reconcile pending approvals against latest artifact version
- require re-confirmation if the artifact changed after the offline approval
- preserve the reason the device was offline

### Route Policy

Default route order:

1. local `mobile-small` route for bounded transforms when available
2. defer or hand off to desktop, CLI, or Android when the task exceeds mobile
   capacity
3. managed remote route only when online and justified

iOS should make handoff cheap instead of pretending every task belongs on the
phone.

## Sync Semantics

Jini should sync events, not raw chat transcripts.

The core sync object is:

- session envelope
- event log
- artifact metadata
- artifact versions
- route evidence
- offline debt
- conflict markers

### Event Categories

Supported offline events:

- `artifact_created`
- `artifact_updated`
- `review_added`
- `approval_recorded`
- `approval_deferred`
- `input_captured`
- `route_used`
- `offline_debt_created`
- `offline_debt_cleared`

### Merge Rules

Default merge rules:

- merge by session id
- preserve every event with device id and timestamp
- rebuild projection after merge
- never discard route evidence
- never silently overwrite the current artifact
- require user review when two devices edited the same artifact version

### Reconciliation Debt

Offline debt must be visible when:

- connector write is queued
- hosted sync has not completed
- approval was recorded against an older artifact version
- route evidence is incomplete
- artifact merge conflict exists

## Route Policy

Jini should use one route policy across all platforms.

The route decision should consider:

- task shape
- modality
- risk
- user preference
- device class
- local profile availability
- local runtime health
- battery and thermal envelope where relevant
- offline state
- provider availability
- online CLI throttle level and quota pressure
- downstream CLI route availability
- prior route regret

### Local-First Rule

Use the cheapest suitable local route when:

- a local profile can satisfy the task
- the route is reliable enough
- the task risk is acceptable
- the user has not pinned a stronger route

### Escalation Rule

Escalate or hand off when:

- the local profile is unavailable
- local latency makes the route expensive in practice
- required modality is unavailable locally
- task risk requires stronger reasoning
- connector write requires online capability
- user asks for a managed or remote route

### Mobile Handoff Rule

Mobile should hand off rather than overrun its role when:

- generation is long-running
- artifact edit is complex
- local model profile is too weak
- battery, thermal, or memory constraints make local work wasteful
- the user needs desktop inspection

## Shipping Prerequisites

### Desktop

macOS and Windows are shippable only when:

- session envelope is shared with CLI
- local artifact store is durable
- offline events append correctly
- route evidence is inspectable
- sync reconciliation is visible
- local model profile selection is device-aware
- installer or app distribution does not hide preview posture

### Mobile

Android and iOS are shippable only when:

- mobile can show the same session identity as CLI and desktop
- latest ready artifact is available offline after sync
- review, approval, defer, and annotation events survive offline mode
- pending sync is visible
- stale approval is detected before send or publish
- handoff to desktop or CLI is obvious when mobile capacity is too small

## Future Update Policy

Jini should improve local capability through the local model registry and
canary loop, not through platform-specific product rewrites.

Future model updates should:

- map to stable profile roles
- run the same offline and continuation checks as current defaults
- preserve route evidence shape
- preserve session and artifact identity
- deprecate old mappings explicitly

Future app updates should:

- keep CLI, desktop, and mobile bound to one session graph
- make offline debt more visible
- reduce handoff cost
- improve local route selection through measured evidence
- avoid adding platform-specific session semantics

## Acceptance Criteria

This strategy is complete only when all are true:

- macOS has explicit offline guarantees, local-model expectations, sync
  semantics, and route policy
- Windows has explicit offline guarantees, local-model expectations, sync
  semantics, and route policy
- Android has explicit offline guarantees, local-model expectations, sync
  semantics, and route policy
- iOS has explicit offline guarantees, local-model expectations, sync semantics,
  and route policy
- all platforms share one session envelope and one route-evidence story
- mobile is positioned as continuation and review, not desktop inference parity
- desktop is positioned as the main offline authoring and artifact host
- future model updates flow through the registry, canary, promote, and
  deprecate loop
