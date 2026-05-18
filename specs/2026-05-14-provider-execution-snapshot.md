# Provider Execution Snapshot

Date: 2026-05-14

## Purpose

Capture the exact state of the Go rewrite after wiring live first-draft provider
execution for Azure OpenAI and Amazon Bedrock, so work can resume without
reconstructing context from chat history.

## Current State

- Go binary supports `jini provider doctor`.
- Go starter flow can use live Azure OpenAI or Amazon Bedrock calls for the
  first useful draft.
- Local preview remains deterministic and offline.
- The shell shows `Working with` using the selected provider label.
- Stale remembered work no longer leaks raw filesystem errors.

## Implemented In This Slice

- Added provider execution layer in [provider.go](/Users/sharad.sharma/Developer/jini/internal/app/provider.go).
- Wired starter bootstrap in [app.go](/Users/sharad.sharma/Developer/jini/internal/app/app.go) so provider-backed drafts can overwrite the primary starter view.
- Added safe AWS SigV4 signing for Bedrock Converse requests.
- Added AWS profile credential and region loading from shared AWS config files.
- Kept the public surface unchanged: `jini` remains the front door.

## Verified Behavior

- Azure OpenAI path:
  - deployment chat completions request is formed correctly
  - `api-key` header is sent
  - secret value is not leaked into request body or UI
- Bedrock path:
  - Converse request is formed correctly
  - SigV4 `Authorization` header is generated
  - static credentials and `AWS_PROFILE` flows are both covered
  - secret value is not leaked into request body or UI
- Provider setup failure blocks work creation with user-facing guidance to run
  `jini provider doctor`.
- Provider-generated first draft is written into the primary starter view.
- Stale current-work state is cleared and replaced with recovery text.

## Verification Results

- `go test ./...` passed
- `python3 tests/test_jini_cli.py` passed
  - `Ran 105 tests`
  - `OK`
- `python3 tools/jini.py publish-readiness --format json`
  - status: `ok`
  - novice: `ok`
  - consensus gates: `ok`
- `python3 tools/jini.py validate-golden-benchmark --format json`
  - overall status: `leading`
  - Jini score: `8.93`
  - failed scenarios: none
- `git diff --check` passed for the files touched in this slice

## Docs Updated

- [README.md](/Users/sharad.sharma/Developer/jini/README.md)
- [docs/cli.md](/Users/sharad.sharma/Developer/jini/docs/cli.md)
- [docs/install.md](/Users/sharad.sharma/Developer/jini/docs/install.md)

The docs now state that the Go binary supports live first-draft calls through
Azure OpenAI and Amazon Bedrock, while local preview remains offline.

## Known Boundaries

- This slice wires provider-backed first useful drafts into the starter flow.
- It does not yet stream progress while the provider is generating.
- It does not yet expose a richer in-shell progress event model during remote
  generation.
- Python compatibility CLI still documents and validates provider setup, but the
  live provider execution path is implemented in the Go runtime.

## Clean Resume Point

Next highest-leverage slice:

1. Add visible in-shell progress while provider generation is running.
2. Improve prompt shaping for meeting follow-up and spec-readiness so remote
   providers produce stronger first drafts.
3. Add one end-to-end smoke path for a visible user scenario such as `Plan 7
   day Paris trip` using a fake transport in tests and a real provider behind a
   feature flag when networked verification is available.
