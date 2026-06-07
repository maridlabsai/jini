# Travel Plan Pack

This pack turns a travel goal into a canonical itinerary, plan, tasks, and
verification surface.

It is a deliberate non-software pack used to prove Jini can support personal
planning without changing the kernel.

## Intended Profile

- `Explore`

## Typical Extensions

- `Business:personal-planning`
- `Modality:travel`
- `Environment:docs-local`
- `Risk:consumer`

## Typical Control Packs

- `Proof`
- `Guard`
- `Cost`

## Compiled Flow

1. `Scope`
   - create `Brief`
   - capture destination, date, and budget intent

2. `Probe`
   - surface assumptions, hard constraints, and missing trip details

3. `Model`
   - create `Spec`
   - define itinerary, logistics, and contingency requirements

4. `Decide`
   - create `Decision`, `Plan`, and `Tasks`

5. `Make`
   - render the itinerary and task surfaces

6. `Verify`
   - bind `Evidence` to the active itinerary revision before execution

Use the travel plan workflow with the native Go front door:

```bash
jini
```

Paste the destination, dates, budget, pace, constraints, and contingency needs
as source context. The native pack compiler is tracked as future Go work; until
that command is ported, this pack is a reusable workflow reference.
