# Jini Core

## 1. Purpose

Jini defines a universal protocol for consequential work. "Consequential" means
work where correctness, traceability, cost, safety, legal standing, or
operational outcomes matter enough that purely conversational execution is not
sufficient.

Jini MUST:

- separate durable work state from runtime-specific prompting
- support humans and machines as cooperating actors
- preserve provenance, authority, evidence, and decision history
- scale from solo work to regulated enterprise work through profiles
- survive model, vendor, and tool churn

## 2. Normative Terms

The keywords MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are normative.

## 3. Canonical Naming

Jini uses the short canonical names defined in
[canonical-names.md](./canonical-names.md).

## 4. Kernel Operations

Jini exposes exactly six kernel operations.

### 4.1 Scope

`Scope` normalizes an ask into a coherent work object.

Outputs typically include:

- Brief
- initial Assumptions
- initial scope boundaries


### 4.2 Probe

`Probe` tests assumptions, contradictions, risk, trust, economics, and
reversibility before commitment.

Outputs typically include:

- updated Assumptions
- identified risks and unresolved questions
- decision pressure on weak plans


### 4.3 Model

`Model` shapes the work into entities, workflows, contracts, options,
dependencies, and boundaries.

Outputs typically include:

- Spec
- Decision alternatives
- dependency and environment models

### 4.4 Decide

`Decide` selects a path and binds responsibility, acceptance, and rollback
intent.

Outputs typically include:

- Decision
- Plan
- Tasks
- Approval when required


### 4.5 Make

`Make` produces the target artifact, service, filing, document, process, or
system change.

Outputs typically include:

- implemented change or produced artifact
- updated task state
- generated evidence inputs


### 4.6 Verify

`Verify` validates evidence, authorizes transitions, releases or submits when
allowed, and reopens work when reality disagrees.

Outputs typically include:

- Evidence
- Approval
- Runbook or Submission when relevant
- transition to operational, retired, or reopened state


## 5. Core Principles

### 5.1 Tiny Kernel

The kernel MUST remain small. Domain growth MUST be represented through profiles
and extensions rather than new kernel operations.

### 5.2 Artifact-First State

Prompts and conversations MAY inform work, but canonical work state MUST live in
typed artifacts and the event log.

### 5.3 Guarded Transitions

A WorkUnit MUST NOT advance based solely on artifact existence. It advances only
when transition guards pass.

### 5.4 Explicit Authority

Approval, waiver, rollback, and operational responsibilities MUST be modeled
explicitly.

### 5.5 First-Class Operations

Production reality MUST NOT be treated as a late-stage concern. Incident,
degraded-mode, rollback, and support semantics MUST integrate with the protocol.

## 6. Product Surface vs Protocol Core

Jini distinguishes between:

- protocol core: durable semantics
- compiled workflows: user-facing conveniences
- adapters: runtime-specific glue

Workflow packs MAY expose commands such as `to-prd`, `triage`, or `tdd`, but
those commands MUST compile into Jini artifacts, transitions, and controls.

## 7. Compliance Requirement

A system MAY claim Jini compatibility only if it can:

- create and update canonical WorkUnits
- emit versioned artifacts conforming to schema
- honor guarded state transitions
- preserve provenance and authority metadata
- produce an append-only event history
