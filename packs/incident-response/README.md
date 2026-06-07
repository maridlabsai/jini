# Incident Response Pack

This pack turns an incident into a canonical stabilization, communication,
verification, and closure surface.

It broadens Jini into operations work without adding new kernel semantics.

## Intended Profile

- `Critical`

## Typical Extensions

- `Business:operations`
- `Modality:incident-response`
- `Environment:repo-local`
- `Risk:service-outage`

## Typical Control Packs

- `Proof`
- `Guard`
- `Rollback`

## Compiled Flow

1. `Scope`
   - create `Brief`
   - capture scope, severity, ownership, and customer impact

2. `Probe`
   - surface assumptions, unknowns, and rollback constraints

3. `Model`
   - create `Spec`
   - define mitigation, communication, verification, and closure requirements

4. `Decide`
   - create `Decision`, `Plan`, and `Tasks`

5. `Make`
   - render the response plan and task surfaces

6. `Verify`
   - bind `Evidence` to the active response revision before closure

Use the incident response workflow with the native Go front door:

```bash
jini
```

Paste the incident summary, current mitigation, rollback options, owner, and
approver as source context. The native pack compiler is tracked as future Go
work; until that command is ported, this pack is a reusable workflow reference.
