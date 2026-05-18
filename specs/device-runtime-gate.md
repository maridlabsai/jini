# Device Runtime Gate

Updated: 2026-05-16

This is the independent gate for device-aware local runtime routing.

It is separate from publish readiness because local capability routing can drift
without obviously breaking the main product surface.

## Gate Categories

### 1. Capability Probe

Must prove code exists for:

- OS detection
- OS version detection
- CPU architecture detection
- memory detection
- accelerator detection
- local runtime class detection

### 2. Versioned Cache

Must prove:

- repo-local device profile path exists
- Jini version is recorded
- capability registry version is recorded
- capture timestamp is recorded
- freshness / re-probe logic exists
- profile invalidates on OS/runtime/endpoint drift, not only time

### 3. Routing Use

Must prove:

- device class reaches route features
- route scoring includes a device capability bias
- local profile availability can downgrade or block expensive local routes
- local profile availability reflects backend readiness, not only hardware class

### 4. Transparency

Must prove:

- provider doctor exposes device class
- provider doctor exposes accelerator class
- provider doctor exposes local runtime class

### 5. Tests

Must prove:

- device class override or equivalent deterministic test path exists
- local route selection tests cover device-aware behavior
- local provider doctor tests cover device-aware output

## Gate Command

The independent validator command should be:

```bash
python3 tools/jini.py validate-device-runtime-gate --format json
```

The gate fails if any category above fails.
