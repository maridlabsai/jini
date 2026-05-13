# Jini Atlassian Targets

## 1. Purpose

Jini needs a safe way to bind real Jira and Confluence destinations without:

- hardcoding site details into packs
- guessing project or space targets
- forcing every publish command to repeat the same parameters

This document defines the local target-binding layer for Atlassian-backed
execution.

## 2. Binding File

Jini stores Atlassian targets at:

- `runtime/atlassian-targets.json`

This file is pack-local runtime configuration, not a canonical artifact. It
exists to support external-system publishing while keeping the WorkUnit and its
artifacts portable.

## 3. Bound Fields

The current binding stores:

- `cloud_id`
- `site_url`
- Jira:
  - `project_key`
- Confluence:
  - `space_key`
  - `space_id`

These are sufficient for Jini to produce connector-ready publish plans without
blind target guessing.

## 4. Commands

Jini currently exposes:

- `bind-atlassian`
- `show-atlassian`

### `bind-atlassian`

Persists pack-local Atlassian defaults.

### `show-atlassian`

Shows the current binding so runtime behavior is inspectable before any publish
step.

## 5. Interaction With `run-pack`

When a pack has a binding:

- `run-pack` automatically uses the bound Jira project key if no override is
  provided
- `run-pack` automatically uses the bound Confluence space if no override is
  provided
- `run-pack` can render `connector-ready` publish plans instead of generic
  staged plans

When a pack has no binding:

- Jira publish plans remain unbound staged plans
- Confluence falls back to markdown when no space is configured

## 6. Safety Rules

Jini MUST NOT:

- infer a Jira project from a pack name alone
- infer a Confluence space from a pack name alone
- treat a site binding as permission to publish live automatically
- bypass markdown fallback when Confluence is unconfigured

## 7. Current Scope

This binding layer improves:

- Atlassian readiness
- runtime ergonomics
- publish-plan portability

It does not yet provide:

- live Jira issue creation inside the local CLI
- live Confluence page creation inside the local CLI
- inbound sync from Atlassian back into canonical Jini artifacts
