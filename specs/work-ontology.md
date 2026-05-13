# Jini Ontology

## 1. Purpose

This document defines the object model used by Jini. The goal is to prevent
category errors between artifacts, actions, decisions, approvals, runtime
contexts, and domain extensions.

## 2. Upper Ontology

Jini defines the following peer object kinds.

- `WorkUnit`: the canonical aggregate root for a piece of work
- `Operation`: an invocation of a kernel operation
- `Decision`: a chosen option with rationale and scope
- `Authorization`: a statement that a named actor may approve, waive, release,
  roll back, or otherwise authorize a transition
- `StateTransition`: a change of WorkUnit state or artifact state
- `Artifact`: a versioned semantic document or structured object
- `Evidence`: a claim-supporting result bound to a specific target revision
- `ActorRole`: a scoped role such as owner, approver, operator, reviewer, or
  rollback authority
- `Environment`: a runtime context such as development, staging, production,
  air-gapped, field, or regulated environment
- `Extension`: a typed modifier that augments protocol behavior without changing
  kernel semantics

## 3. WorkUnit

`WorkUnit` is the aggregate root. All canonical protocol state attaches to a
WorkUnit.

### 3.1 Required Fields

- `work_unit_id`
- `title`
- `purpose`
- `current_state`
- `profile_id`
- `active_extensions`
- `branch_id`
- `parent_work_unit_id` when derived
- `owner_actor_id`
- `approver_actor_ids`
- `operator_actor_id` when relevant
- `rollback_authority_actor_id` when relevant
- `service_owner_actor_id` when relevant
- `stakeholder_actor_ids`
- `created_at`
- `updated_at`

### 3.2 WorkUnit Rules

- A canonical artifact MUST belong to exactly one active WorkUnit revision.
- A WorkUnit MAY branch.
- A WorkUnit MAY be reopened after verification failure or incident findings.
- A WorkUnit MUST maintain lineage through parent and supersession links.

## 4. Actor Roles

Jini models roles explicitly.

### 4.1 Canonical Roles

- `owner`
- `approver`
- `reviewer`
- `operator`
- `rollback_authority`
- `waiver_authority`
- `service_owner`
- `incident_commander`

### 4.2 Role Rules

- Roles MAY be held by human actors, automated agents, or service principals.
- A profile MAY require segregation of duties between owner and approver.
- An actor MUST NOT be inferred as approver merely because they authored an
  artifact.

## 5. Semantic Relations

Jini uses typed relations between objects.

Required relation verbs include:

- `derived_from`
- `supersedes`
- `implements`
- `satisfies`
- `tests`
- `mitigates`
- `approves`
- `owned_by`
- `operated_by`
- `affects`
- `depends_on`

## 6. Atomic Domain Objects

Inside WorkUnits and artifacts, Jini recognizes typed semantic objects such as:

- `Requirement`
- `Constraint`
- `Assumption`
- `Option`
- `Risk`
- `Control`
- `Claim`
- `Metric`
- `TestCase`
- `Dependency`
- `Change`
- `Incident`
- `Service`

These are not peer protocol roots; they live within or are referenced by
artifacts and evidence.

## 7. Events

Every meaningful protocol mutation SHOULD produce an event.

Examples:

- WorkUnit created
- Spec approved
- Decision superseded
- Evidence invalidated
- Incident opened
- Rollback authorized

Events MUST carry:

- event id
- timestamp
- actor id
- affected object ids
- prior state
- resulting state
- reason or transition basis
