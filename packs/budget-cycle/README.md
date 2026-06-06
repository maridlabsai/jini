# Budget Cycle Pack

`budget-cycle` is a non-software benchmark pack that applies Jini to monthly budgeting.

It exists to prove that the kernel, control packs, compact memory, and verification loop can work outside software delivery.

Use it with the native Go front door:

```bash
jini
```

Paste the monthly budget constraints, obligations, savings goal, and fallback
cuts as source context. The native pack compiler is tracked as future Go work;
until that command is ported, this pack is a reusable workflow reference.

The pack is designed to materialize:

- canonical planning artifacts under `artifacts/`
- a rendered `views/budget.md`
- tasks, issue, and wiki export surfaces via the standard compile pipeline
