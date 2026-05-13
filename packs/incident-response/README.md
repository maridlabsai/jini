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

Compile an incident response plan from the benchmark context with:

```bash
python3 tools/jini.py compile-pack incident-response \
  --work-unit-id my-incident-response \
  --title "Checkout Latency Incident" \
  --purpose "Stabilize the checkout path with explicit rollback and verification" \
  --owner incident-commander \
  --approver service-owner \
  --output /tmp/my-incident-response
python3 tools/jini.py status-pack /tmp/my-incident-response
```
