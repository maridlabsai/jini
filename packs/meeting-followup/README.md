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

Use the meeting follow-up workflow with the native Go front door:

```bash
jini
```

Paste the meeting notes, decisions, owners, and follow-up constraints as source
context. The native pack compiler is tracked as future Go work; until that
command is ported, this pack is a reusable workflow reference.
