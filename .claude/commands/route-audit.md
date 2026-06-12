# Audit Jini Routing

Inspect routing, offline fallback, provider selection, and route diagnostics.

## Canonical Inputs

- `specs/execution-routing-policy.md`
- `specs/runtime-execution-modes.md`
- `specs/runtime-selection-heuristics.md`
- `specs/device-capability-routing.md`
- `specs/local-slm-frontline-policy.md`

## Procedure

1. Read only the canonical inputs needed for the changed route surface.
2. Inspect the route engine implementation under `internal/`.
3. Check no-config, no-internet, provider-failure, battery, throttling, and user override behavior.
4. Confirm CLI and app use the same route contract.

## Output

Return route gaps as `scenario`, `expected`, `actual`, `fix`, and `test`.

$ARGUMENTS
