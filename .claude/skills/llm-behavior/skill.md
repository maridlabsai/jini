# LLM Behavior

Use this skill when Jini responds with the wrong intent, wrong route, or wrong output shape.

## Procedure

1. Capture the smallest prompt transcript that reproduces the issue.
2. Classify the failure as intent, routing, context carryover, rendering, prompt, or tool-selection.
3. Inspect the shared engine before proposing a one-off fix.
4. Add or update a prompt-bank or CLI UX gate fixture.

## Output

Return `repro`, `classification`, `root cause`, `fix`, and `test`.
