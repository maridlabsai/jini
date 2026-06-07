# Engineering Principles

Jini should not rely on taste alone for implementation quality. Development
should follow explicit engineering rules that keep the product dependable,
cost-effective, and frictionless as the codebase grows.

## Core Rule

Favor simple object boundaries, explicit contracts, and composable design over
cleverness, global state, or feature-shaped code sprawl.

## SOLID Rules

### Single Responsibility Principle

Each module, class, and helper should have one reason to change.

Examples:

- session persistence should not also decide routing
- rendering code should not also mutate entitlement state
- storefront planning should not also compute release blockers

### Open/Closed Principle

New surfaces, providers, and policies should extend existing contracts instead
of rewriting the core object model.

Examples:

- add a new adapter or strategy instead of branching the kernel
- add a new policy implementation instead of growing one giant conditional

### Liskov Substitution Principle

Any implementation behind a shared contract should preserve the same semantics.

Examples:

- every surface adapter must mean the same thing by `ready`, `missing`, and
  `resume`
- every route policy must return a valid route decision object, not ad hoc
  shape changes

### Interface Segregation Principle

Contracts should stay narrow and task-specific.

Examples:

- a host-surface contract should not force payment logic into preview-only
  callers
- a storage contract should not require UI concerns

### Dependency Inversion Principle

High-level workflow logic should depend on explicit contracts, not concrete
transport or UI details.

Examples:

- session logic depends on storage and projection interfaces
- shell rendering depends on availability/build contracts
- routing depends on policy interfaces, not direct provider branches

## OOP Rules

- use objects and dataclasses for stable domain concepts, not incidental
  wrappers
- keep state transitions explicit and inspectable
- prefer composition over inheritance
- keep constructors small and intention-revealing
- avoid mutable global state as product truth

## Preferred Design Patterns

Jini should default to a small set of patterns:

- Adapter: for surfaces, providers, and external integrations
- Strategy: for routing, pricing, entitlement, and recovery policy
- Factory: for building manifests, session objects, and generated bundles
- Facade: for user-facing orchestration surfaces over deeper subsystems
- Value Object: for explicit immutable contracts such as session, route, and
  availability shapes

These patterns should be used to reduce coupling, not to add ceremony.

## Reject Conditions

Reject changes that introduce:

- god objects that own unrelated concerns
- giant switch statements where a strategy or adapter should exist
- UI copy as the source of runtime truth
- hidden globals for session, entitlement, or route state
- inheritance hierarchies where composition would be clearer
- boolean-flag APIs that hide multiple modes in one object
- business logic spread across rendering, storage, and transport layers

## Review Checklist

Before merging, ask:

- does each changed module have one clear responsibility
- did this change extend a contract instead of branching core semantics
- are adapters and strategies used where variation is expected
- is composition clearer than inheritance here
- is runtime truth stored in one explicit contract instead of repeated strings
- would a new engineer understand where to add the next platform or policy
