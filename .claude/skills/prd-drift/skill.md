# PRD Drift

Use this skill when implementation, docs, or release planning may be diverging from the canonical Jini product direction.

## Canonical Inputs

- `specs/number-one-platform-prd.md`
- `specs/prd-implementation-trace.md`
- `specs/product-settling-decisions.md`
- `specs/product-rewrite-contract.md`
- `specs/golden-competitive-benchmark.yaml`

## Process

1. Compare the changed surface against the canonical inputs.
2. Identify stale, contradictory, or undesired requirements.
3. Decide whether the fix belongs in code, docs, PRD, gates, or scorecards.
4. Prefer removing drift over adding another planning artifact.

## Output

Return a concise table with `gap`, `risk`, `fix`, and `owner surface`.
