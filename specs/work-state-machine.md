# Jini State Machine

## 1. Overview

Jini is a guarded transition system, not a linear phase checklist.

Two state machines exist:

- WorkUnit state machine
- Artifact lifecycle state machine

## 2. WorkUnit States

- `intake`
- `scoped`
- `probed`
- `modeled`
- `decided`
- `in_make`
- `awaiting_verification`
- `operational`
- `reopened`
- `incident`
- `retired`

## 3. Artifact States

- `draft`
- `reviewed`
- `approved`
- `superseded`
- `invalidated`
- `merged`
- `archived`

## 4. Guarded Transition Rules

The CLI surface SHOULD expose these transitions through guarded commands such as
`advance-pack`, which only promotes a WorkUnit when the required artifacts for
the target state are present and ready.

### 4.1 Intake -> Scoped

Allowed when:

- WorkUnit exists
- Brief is present and internally valid

### 4.2 Scoped -> Probed

Allowed when:

- initial assumptions are recorded
- active profile is assigned
- required control packs for the profile are attached

### 4.3 Probed -> Modeled

Allowed when:

- critical contradictions are either resolved or explicitly deferred
- required probe questions for active extensions are addressed

### 4.4 Modeled -> Decided

Allowed when:

- current Spec revision is approved
- one selected option has a recorded Decision
- required authorities have approved

### 4.5 Decided -> In Make

Allowed when:

- Plan exists
- Tasks exist
- rollback intent exists for profiles that require it

### 4.6 In Make -> Awaiting Verification

Allowed when:

- target work is realized for the active revision
- required evidence collection is complete enough to begin verification

### 4.7 Awaiting Verification -> Operational

Allowed when:

- Evidence validates the current target revision
- Approval exists where required
- Operational artifacts required by the active profile are present and approved

### 4.8 Awaiting Verification -> Reopened

Allowed when:

- evidence fails
- approval is denied
- risk or operational findings invalidate the current revision

### 4.9 Any State -> Incident

Allowed when:

- a production or mission-impacting failure is declared
- incident mode trigger conditions are met under the current profile

Incident mode MUST override normal routing rules, while preserving event history
and requiring mandatory backfill before incident closure.

### 4.10 Incident -> Reopened or Operational

Allowed when:

- incident mitigation is complete
- incident-specific verification passes
- required backfill artifacts and post-incident updates are complete

## 5. Reopen Semantics

The protocol MUST allow reopening from Verify or Incident back into earlier
states.

Common reopen paths:

- `awaiting_verification -> modeled`
- `operational -> reopened`
- `incident -> probed`
- `incident -> decided`

## 6. Branch and Merge

Parallel realization is expected.

### 6.1 Branch Rules

- A WorkUnit MAY branch from any non-retired state.
- Each branch MUST have a distinct branch id.
- Artifacts MUST declare branch lineage.

### 6.2 Merge Rules

- Merge MUST identify conflict class: editorial, semantic, operational, or
  authority conflict.
- Semantic conflicts MUST be resolved by an owner or designated reconciler.
- Merge completion MUST emit supersession and provenance events.

## 7. Invalidation Cascades

If a semantic or breaking change occurs in a Spec, the protocol MUST
determine downstream invalidation.

Potentially stale artifacts include:

- Plan
- Tasks
- Evidence
- Approval
- Runbook
- Signals
- Rollback

Each extension and profile MAY add stricter invalidation rules.

## 8. Incident Mode

Incident mode MUST define:

- severity
- incident commander
- communications channel
- approval override rules
- change freeze or exception handling
- mandatory backfill requirements

Incident mode is not a shortcut around the protocol. It is a protocol branch
with different guards.
