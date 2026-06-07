# Incident Response Benchmark Context

The service is showing a sustained degradation on a customer-visible checkout
path. Operators need a compact, governed response surface that keeps scope,
mitigation, communication, verification, and rollback explicit under pressure.

The benchmark context is intentionally local-first:

- repo or runtime verification can be harvested through bounded local checks
- publish outputs should remain portable through staged or local-apply adapters
- closure should be blocked until recovery proof and residual risks are visible

The pack should compile into a usable response plan with no manual YAML repair.
