# Check Jini PRD Drift

Compare the current work against the canonical Jini product direction.

## Canonical Inputs

- `specs/number-one-platform-prd.md`
- `specs/prd-implementation-trace.md`
- `specs/product-settling-decisions.md`
- `specs/product-rewrite-contract.md`
- `specs/golden-competitive-benchmark.yaml`

## Procedure

1. Inspect changed files.
2. Read only the canonical inputs needed for those files.
3. Identify stale, contradictory, or undesired requirements.
4. Prefer removing drift over adding a new planning document.

## Output

Return `gap`, `risk`, `fix`, and `verification`.

$ARGUMENTS
