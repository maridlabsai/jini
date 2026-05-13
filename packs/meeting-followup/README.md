# Meeting Follow-up Pack

This pack turns a meeting into a canonical follow-up workflow with explicit
decisions, owners, tasks, evidence, and approval-ready handoff.

It is a deliberately common public-core pack meant to prove that Jini is not
just for niche workflows. Most teams run this problem every week.

## Intended Profile

- `Delivery`

## Typical Extensions

- `Business:team-operations`
- `Modality:meeting`
- `Environment:docs-local`
- `Risk:collaboration`

## Typical Control Packs

- `Proof`
- `Guard`
- `Cost`

## Compiled Flow

1. `Scope`
   - create `Brief`
   - capture meeting objective, participants, and expected outcomes

2. `Probe`
   - surface assumptions, open questions, and anything that still lacks a clear owner

3. `Model`
   - create `Spec`
   - define the follow-up structure: decisions, tasks, deadlines, and escalation points

4. `Decide`
   - create `Decision`, `Plan`, and `Tasks`

5. `Make`
   - render the follow-up summary and task surfaces

6. `Verify`
   - bind `Evidence` to the active follow-up revision before execution

Compile a meeting follow-up workflow from the benchmark context with:

```bash
python3 tools/jini.py compile-pack meeting-followup \
  --work-unit-id my-meeting-followup \
  --title "Weekly Product Review Follow-up" \
  --purpose "Turn one meeting into decisions, owners, and explicit next steps" \
  --owner meeting-owner \
  --output /tmp/my-meeting-followup
python3 tools/jini.py status-pack /tmp/my-meeting-followup
```
