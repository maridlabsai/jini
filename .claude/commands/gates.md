# Run Jini Commit Gates

Run the public Jini commit gate and report the exact outcome.

## Procedure

1. Inspect `git status --short --branch`.
2. Run:

```bash
bash tools/run_required_gates.sh commit
```

3. If the gate fails, stop and summarize the first actionable failure.
4. If it passes, report the important gate evidence, including PRD completion status if printed.

## Output

Return:

- Gate result.
- Important evidence.
- Remaining risk.

$ARGUMENTS
