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

Compile a compliance audit plan from the benchmark context with:

```bash
jini compile-pack compliance-audit \
  --work-unit-id my-compliance-audit \
  --title "Quarterly Controls Audit" \
  --purpose "Review controls, evidence, and signoff readiness for a regulated surface" \
  --owner compliance-lead \
  --approver risk-officer \
  --output /tmp/my-compliance-audit
jini status-pack /tmp/my-compliance-audit
```
