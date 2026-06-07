# Compliance Audit Pack

This pack turns a regulated or compliance-oriented review into a canonical
audit, evidence, approval, and closure surface.

It broadens Jini into governed review work without adding new kernel
semantics.

## Intended Profile

- `Regulated`

## Typical Extensions

- `Business:governance`
- `Modality:compliance`
- `Environment:docs-local`
- `Risk:regulated`

## Typical Control Packs

- `Proof`
- `Guard`
- `Approval`

## Compiled Flow

1. `Scope`
   - create `Brief`
   - capture scope, obligations, reviewer roles, and signoff path

2. `Probe`
   - surface assumptions, evidence gaps, and residual-risk questions

3. `Model`
   - create `Spec`
   - define review, remediation, approval, and closure requirements

4. `Decide`
   - create `Decision`, `Plan`, and `Tasks`

5. `Make`
   - render the audit plan and task surfaces

6. `Verify`
   - bind `Evidence` and `Approval` to the active audit revision before closure

Use the compliance audit workflow with the native Go front door:

```bash
jini
```

Paste the audit scope, control surface, evidence set, owner, and approver as
source context. The native pack compiler is tracked as future Go work; until
that command is ported, this pack is a reusable workflow reference.
