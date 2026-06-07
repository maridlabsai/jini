# Issue Publish Plan: research-prd

- Adapter: `jira`
- Execution Mode: `staged-only`
- WorkUnit: `research-prd-v1`
- Project Key: `DEMO`

## Rate-Limit Policy
- adapter: `jira`
- dispatch_mode: `serialized`
- max_parallel: `1`
- retry_strategy: `exponential-backoff`
- initial_delay_seconds: `2`
- max_delay_seconds: `60`
- max_attempts: `5`
- on_quota_uncertainty: `fallback-to-file-exports`
- on_rate_limit: `stop-and-preserve-remaining-items`

## Payloads
- [01-validate-source-coverage-and-finalize-research-synthesis.json](./payloads/01-validate-source-coverage-and-finalize-research-synthesis.json) -> `research-prd-v1-task-1`
- [02-review-and-approve-the-rendered-prd.json](./payloads/02-review-and-approve-the-rendered-prd.json) -> `research-prd-v1-task-2`
- [03-confirm-build-ready-requirements-and-task-ownership.json](./payloads/03-confirm-build-ready-requirements-and-task-ownership.json) -> `research-prd-v1-task-3`
