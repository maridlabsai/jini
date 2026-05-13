# Wiki Publish Plan: research-prd

- Adapter: `confluence`
- Execution Mode: `markdown-fallback`
- WorkUnit: `research-prd-v1`
- Space Key: `unset`
- Markdown Fallback: `packs/research-prd/examples/research-prd-v1/exports/wiki/markdown`

## Rate-Limit Policy
- adapter: `confluence`
- dispatch_mode: `serialized`
- max_parallel: `1`
- retry_strategy: `exponential-backoff`
- initial_delay_seconds: `2`
- max_delay_seconds: `60`
- max_attempts: `5`
- on_quota_uncertainty: `fallback-to-file-exports`
- on_rate_limit: `stop-and-preserve-remaining-items`

## Payloads
- [01-overview.json](./payloads/01-overview.json) -> `overview`
- [02-prd.json](./payloads/02-prd.json) -> `prd`
- [03-tasks.json](./payloads/03-tasks.json) -> `tasks`
