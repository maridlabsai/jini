# Budget Cycle Pack

`budget-cycle` is a non-software benchmark pack that applies Jini to monthly budgeting.

It exists to prove that the kernel, control packs, compact memory, and verification loop can work outside software delivery.

Compile it with:

```bash
python3 tools/jini.py compile-pack budget-cycle \
  --work-unit-id monthly-budget-v1 \
  --title "Monthly Budget" \
  --purpose "Build a monthly budget with explicit savings, obligations, and fallback cuts" \
  --owner finance-owner \
  --output /tmp/monthly-budget
```

The compiled pack materializes:

- canonical planning artifacts under `artifacts/`
- a rendered `views/budget.md`
- tasks, issue, and wiki export surfaces via the standard compile pipeline
