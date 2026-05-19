# Workstream Technical Framework

Updated: 2026-05-19

## Purpose

This document defines the technical framework for Jini workstreams in the
public repo.

It exists to keep product work, runtime work, surface work, and learning work
moving through one stable architecture instead of fragmenting into parallel
systems.

This is a stable public engineering contract.

## Framework Rule

Every major workstream must build on the same base product contract:

- one work-thread model
- one artifact model
- one routing model
- one trust and verification model
- one install and proof surface

No workstream may invent a separate product architecture that bypasses those
shared contracts.

## Workstream Types

### 1. Surface Workstreams

Examples:

- CLI behavior
- desktop behavior
- mobile continuation/review behavior
- docs and install surfaces
- adaptive response rendering

Responsibilities:

- keep the same product identity across surfaces
- surface the right artifact, missing-state, and next-step information
- preserve free local/BYO usefulness
- render from shared semantic and artifact contracts instead of hard-coded
  use-case response molds

### 2. Runtime Workstreams

Examples:

- provider routing
- local SLM transport
- verification depth
- benchmark and capability learning

Responsibilities:

- choose the cheapest suitable route
- preserve route transparency
- preserve the same work contract regardless of route

### 3. Artifact Workstreams

Examples:

- starter outputs
- follow-up drafts
- readiness checks
- itinerary and recommendation artifacts

Responsibilities:

- return a useful first object before summary
- keep uncertainty visible
- carry continuation state cleanly

### 4. Learning And Policy Workstreams

Examples:

- route feedback
- benchmark learning
- framework review loops
- policy staging and rollback

Responsibilities:

- improve routing and continuation quality without mutating canonical semantics
- keep learning bounded, auditable, and reversible

### 5. Distribution Workstreams

Examples:

- install scripts
- release packaging
- CI
- target-specific shims

Responsibilities:

- keep installation and updates trustworthy
- preserve the same public product contract across targets

## Cross-Cutting Requirements

Every workstream must also answer the same non-functional obligations:

- security and privacy boundaries
- observability and rollback
- migration and compatibility
- user-visible trust and exportability
- competitive usefulness, not just internal elegance

## Shared Invariants

Every workstream must preserve:

1. the remembered-work rule
2. the useful-result-first rule
3. the free local/BYO boundary
4. route and verification transparency
5. one account/work-thread/product identity across surfaces
6. exportability and explicit local-versus-hosted boundaries
7. rollback-safe and observable rollout behavior

## Dependency Order

Work should flow through this order:

1. product contract
2. technical framework
3. workstream-specific design
4. implementation
5. gate and benchmark verification

This prevents local optimization from rewriting the product indirectly.

## Admission Rules

A workstream change is not ready unless it can answer:

- Which shared contract does this rely on?
- Which invariant could this break?
- What proof shows it did not fork the product model?
- What gate or benchmark protects it from regression?
- What is the rollback or downgrade path if it underperforms?
- How does it improve user outcome, trust, or spend posture against the field?

## Required Outputs

Each major workstream should leave behind:

- one design or framework note
- one review trail
- one gate or benchmark hook
- one explicit statement of what changed in user-visible behavior

## Exit Criteria

A workstream should not be called complete unless:

- the shared invariants still hold
- the change is observable in product behavior or protection, not only code
- the rollback path is understood
- the benchmark or gate story is still green
- the change improves or protects competitive posture

## Review Loop

Every major workstream change should go through:

1. critique
2. revision
3. rationalization
4. gate check

That review loop is defined in:

- [adaptive-response-rendering-framework.md](./adaptive-response-rendering-framework.md)
- [adaptive-response-rendering-framework-review.md](./adaptive-response-rendering-framework-review.md)
- [adaptive-response-rendering-framework-gate.md](./adaptive-response-rendering-framework-gate.md)
- [workstream-technical-framework-review.md](./workstream-technical-framework-review.md)
- [workstream-technical-framework-gate.md](./workstream-technical-framework-gate.md)

## What This Framework Rejects

- workstream-specific product models
- shell-specific product contracts
- feature work that bypasses shared artifact or routing semantics
- learning systems that silently rewrite public behavior
- launch plumbing that becomes the real product architecture
- commercial or regulated work that ignores identity, audit, or privacy
  inheritance from the shared product model

## Rationale

Jini is broad enough that drift is the default failure mode.

Without a technical framework, the repo slowly becomes:

- one architecture for CLI
- another for desktop
- another for commercial hosted flows
- another for learning

This document exists to stop that. It makes each workstream answer to the same
product and protocol center.
