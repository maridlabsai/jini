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

Compile a travel plan from the benchmark context with:

```bash
python3 tools/jini.py compile-pack travel-plan \
  --work-unit-id my-travel-plan \
  --title "Spring Travel Plan" \
  --purpose "Plan a constrained trip with explicit itinerary and contingencies" \
  --owner traveler \
  --output /tmp/my-travel-plan
python3 tools/jini.py status-pack /tmp/my-travel-plan
```
