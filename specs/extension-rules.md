# Jini Extension Algebra

## 1. Purpose

Extensions allow Jini to support verticals, risk regimes, and operating contexts
without mutating the kernel.

Extensions MUST NOT alter kernel operation names or base object kinds.

## 2. Extension Types

Jini defines these extension classes:

- `Business`
- `Modality`
- `Risk`
- `Environment`
- `Regulation`

Control packs are also extensions, but with mandatory protocol-level effects.

## 3. Extension Declaration

Every extension MUST declare:

```yaml
extension_id: string
extension_type: string
version: string
targets: [object_type]
required_fields: [field_ref]
required_guards: [guard_ref]
required_artifacts: [artifact_type]
required_evidence: [evidence_rule]
dependencies: [extension_id]
incompatibilities: [extension_id]
precedence: integer
merge_strategy: additive|override|reject
```

## 4. Allowed Effects

An extension MAY:

- add fields to target artifact schemas
- require additional artifacts
- require stronger transition guards
- require stronger evidence burden
- forbid transitions under certain conditions

An extension MUST NOT:

- rename kernel operations
- silently downgrade profile requirements
- remove provenance fields from artifacts

## 5. Composition Rules

### 5.1 Dependencies

If extension A depends on extension B, A MUST NOT be active unless B is also
active.

### 5.2 Incompatibilities

If extension A conflicts with extension B, the system MUST reject the pair or
resolve the conflict through a higher-precedence profile rule.

### 5.3 Precedence

When two extensions modify the same target field or guard, the higher precedence
rule applies unless merge_strategy is `reject`, in which case composition fails.

### 5.4 Merge Strategy

- `additive`: both modifications apply
- `override`: higher precedence wins
- `reject`: composition is invalid

## 6. Control Packs

The following packs SHOULD exist as first-class extensions:

- `Proof`
- `Guard`
- `Cost`
- `Authority`
- `Resilience`

These packs MAY be mandatory under certain profiles.

## 7. Adapter Conformance

A runtime adapter MUST understand:

- active extension set
- required artifacts and guards
- extension-induced invalidation behavior

An adapter that ignores active extensions MUST fail conformance.

## 8. Anti-Sprawl Rule

A proposed extension MUST include:

- prevented failure mode
- measurable benefit
- compatibility story
- sunset criteria

Without these, the extension MUST be rejected.
