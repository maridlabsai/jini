# Jini Execution Policy

This document is a specialized execution-routing policy, not the
top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this policy conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, or automation posture, the canonical PRD wins and
this policy should be updated.

## 1. Purpose

This document defines how Jini SHOULD choose the cheapest adequate execution
path without sacrificing correctness, completeness, or safety.

The goal is not maximum cleverness.

The goal is:

- avoid unnecessary token and context burn
- avoid avoidable API pressure and rate-limit failures
- use stronger reasoning only when the work actually demands it
- fall back cleanly to local artifacts when external systems are unavailable

## 2. Execution Classes

Jini uses three execution classes.

### `cheap`

Use for:

- mechanical exports
- schema validation
- markdown rendering
- task sync generation
- issue/wiki bundle generation
- bulk deterministic transformations

Rules:

- no child delegation by default
- prefer local files and parsers only
- do not call external publish APIs unless explicitly requested

### `standard`

Use for:

- scoped research
- repo-aware planning
- spec drafting
- PRD synthesis
- normal build execution

Rules:

- targeted context only
- at most one specialist helper by default
- child helpers must return to the parent if deeper reasoning is required

### `deep`

Use for:

- verification in high-control profiles
- production readiness and release gates
- incident handling
- regulated or high-risk tradeoffs
- irreversible decisions

Rules:

- broad but bounded context
- coordinator plus verifier allowed
- max delegation depth is 2
- explicit approval and evidence requirements remain binding

## 3. Routing Rules

Jini SHOULD route by intent first, then adjust by profile and state.

Default intent routing:

- `export`, `issues`, `wiki` -> `cheap`
- `scope`, `probe`, `research` -> `standard`
- `model`, `decide`, `make` -> `standard`
- `verify`, `publish` -> `deep`

Adjustment rules:

- `Critical` and `Regulated` profiles SHOULD increase the class by one step
- `awaiting_verification`, `incident`, and `operational` states SHOULD increase
  the class by one step for non-export intents
- export-only operations SHOULD stay `cheap` even in stricter profiles unless a
  live publish is requested

## 4. Tool Order

Jini SHOULD prefer tools in this order:

1. local artifacts, rendered views, and text parsers
2. structured local exports
3. bounded external text or system fetches
4. authenticated publish APIs
5. vision or screenshot-heavy tools only when text-first paths are insufficient

## 5. Rate-Limit Avoidance

Jini MUST treat rate-limit avoidance as a first-class control concern.

Rules:

- prefer local rendering/export over live API calls when both satisfy the need
- serialize external publish calls
- do not burst Jira and Confluence writes in parallel
- if adapter availability or quota is uncertain, emit markdown/json bundles and
  stop before live publish
- compact and reload the smallest context slice that satisfies the current
  intent
- do not let child workers auto-upshift to a more expensive class on their own

## 6. Fallback Rules

When live systems are unavailable:

- Jira unavailable -> keep the exported Jira issue bundle as the final artifact
- Confluence unavailable -> keep the markdown wiki export as the final artifact
- deeper reasoning unavailable -> return control to the parent and request
  rerouting instead of recursive escalation

## 7. Safety Rules

Execution policy MUST NOT:

- bypass required approvals
- reduce mandatory evidence burden
- bypass forbidden transitions
- claim release readiness from export artifacts alone
- trade away correctness or safety just to reduce cost

## 8. Current Implementation

Jini currently exposes this policy through:

- `recommend-execution`
- `run-pack`
- `export-issues`
- `export-wiki`

When learning is enabled, the policy may also be represented as a local
learning artifact for offline review and bounded routing updates.
