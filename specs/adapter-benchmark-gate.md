# Adapter Benchmark Gate

Updated: 2026-05-16

This is the independent gate for the adapter registry and empirical local
benchmarking layer.

## Gate Categories

### 1. Registry

Must prove:

- adapter registry exists
- adapter metadata includes provider mode and benchmark support
- local adapter modes come from the registry

### 2. Capability Report

Must prove:

- local runtime capability report path exists
- report records device/runtime fingerprint
- report records per-adapter benchmark rows
- freshness logic exists

### 3. Benchmarking

Must prove:

- code runs a real benchmark request for local adapters
- benchmark captures latency and quality class
- benchmark captures throughput and cold-start cost
- benchmark captures structured-output repeatability
- benchmark history is persisted per adapter
- direct cohort history is persisted per adapter and request cohort
- explicit cohort feedback is persisted for matching Local SLM cohorts
- graded artifact usefulness feedback is persisted for matching Local SLM
  cohorts
- passive edit-distance signals are persisted for matching Local SLM cohorts
- passive section-aware edit signals are persisted for matching Local SLM
  cohorts
- semantic core-edit signals are persisted for matching Local SLM cohorts
- downstream artifact outcome signals are persisted for matching Local SLM
  cohorts
- passive workflow outcome signals from repeated opens, export opens, and
  substantive reopen-after-rewrite events are persisted for matching Local SLM
  cohorts
- opt-in external-copy observation can persist matching passive signals for
  Local SLM cohorts without enabling hidden background monitoring
- benchmark can mark adapters as ok, degraded, failed, or not configured

### 4. Routing Use

Must prove:

- route scoring includes measured benchmark bias
- measured failures can suppress a local route
- measured success can improve a local route
- weak throughput or heavy cold-start cost can reduce a local route
- unstable structured output can reduce a local route
- repeated regression across recent samples can reduce a local route
- one noisy sample is confidence-weighted lower than sustained regression
- stale regressions decay and count less than fresh regressions
- strong recovery after degradation can restore route score faster
- benchmark recovery is weighted by current work shape instead of applied
  uniformly across unrelated tasks
- benchmark weighting narrows by artifact family or cohort inside a broad work
  class, so planning evidence does not transfer uniformly across all planning
  subtypes
- direct cohort evidence from real Local SLM completions can outrank discounted
  transferred adapter evidence for matching request cohorts
- explicit user feedback can bias matching cohorts without affecting unrelated
  cohorts
- graded artifact usefulness feedback can bias matching cohorts more precisely
- passive workflow behavior can bias matching cohorts even when the user never
  labels the outcome explicitly
- opt-in external-copy edits or substantive replacement can bias matching
  cohorts through the same passive signal path
  than binary voting alone
- passive edit-distance learning can distinguish tiny cleanup from substantive
  rewrite for matching cohorts
- passive section-aware learning can distinguish cosmetic header edits from
  supporting-section changes and core rewrites
- semantic section-aware learning can distinguish core wording cleanup from
  actual decision changes
- downstream artifact adoption can influence cohort routing more strongly than
  passive edit patterns alone

### 5. Transparency And Tests

Must prove:

- provider doctor exposes benchmark summary lines
- benchmark capture tests exist
- benchmark-based routing tests exist

## Gate Command

The independent validator command should be:

```bash
jini validate-adapter-benchmark-gate --format json
```
