# Run Alpha Smoke Review

Check whether the current Jini cut is credible for alpha testers.

## Procedure

1. Inspect `specs/claude-codex-prompt-bank.jsonl`.
2. Run the CLI UX and Claude/Codex use-case gates if relevant:

```bash
bash tools/cli_ux_regression_gate.sh
bash tools/claude_codex_usecase_gate.sh
```

3. For any changed user flow, inspect the expected transcript shape.
4. Block if simple prompts can still produce saved task snapshots, generic drafts, or stale shell language.

## Output

Return:

- Alpha readiness verdict: `ready`, `needs-work`, or `block`.
- Evidence.
- Top fix before tester handoff.

$ARGUMENTS
