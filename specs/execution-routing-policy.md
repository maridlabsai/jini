# Jini Execution Policy

Traced to number-one-platform-prd.md §Routing And Resource Policy.

This document is a specialized execution-routing policy, not the
top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

Internal engineering delegation is governed by
[agentic-development-operating-model.md](./agentic-development-operating-model.md).
This execution policy may choose cheap, standard, or deep work classes, but it
does not create public `delegate` commands or expose coordinator/sub-agent
trees in the default CLI.

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
- non-trivial engineering cuts with bounded sub-agent workstreams

Rules:

- targeted context only
- coordinator-owned divide-and-conquer is required for non-trivial Jini
  engineering cuts
- sub-agent write scopes must be disjoint and evidence-bound
- child helpers must return to the coordinator if deeper reasoning is required

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
- disjoint scope ownership and independent review are mandatory
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
- overlapping sub-agent write scopes -> stop parallel work and serialize
  through the coordinator

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

## Absorbed Policies

These normative rules were merged here from now-archived routing docs during
the PRD rebuild (prd-rebuild-design.md §8). Where a source conflicted with the
rebuilt PRD, the PRD wins: "local commercial SLM" phrasing from the old tier
doctrine is dropped — local model routing is a free-tier disclosed route per
the PRD's Tier Boundary and Option A scope. Stale model-name examples were
dropped; model choice follows the chosen route.

### Runtime modes (absorbed from runtime-execution-modes.md)

Jini executes local workflow actions in `supervised` or `autonomous` mode.

- `supervised`: outputs are inspectable before state changes; local exports may
  run with `write` consent; publish plans may be staged with `publish` consent;
  state transitions are planned but not auto-executed.
- `autonomous`: additionally, one legal linear state transition may run when
  `command` consent exists and all guard conditions pass. Autonomous mode MUST
  stop when a required consent category is missing, a guarded transition is
  blocked, or a step would require human-authored semantic input not present in
  canonical artifacts.

First-time consent persists by action category — `write` (deterministic local
file outputs), `command` (deterministic workflow progression; does not waive
lifecycle guards), `publish` (staging external publish plans; serialized
publish rules remain binding) — at `runtime/consent.json`, with the most
recent run report at `runtime/last-run.json`.

Runtime modes MUST NOT: bypass required evidence or approval, skip guarded
transitions, silently escalate execution class, burst publish actions in
parallel, or replace missing human input with invented semantic content.

### Selection heuristics (absorbed from runtime-selection-heuristics.md)

Tool, model, effort, and local-vs-remote choice are first-class runtime
decisions on every request, decided in this order:

1. classify the work type
2. classify the required depth
3. decide whether the local SLM pool can handle it well enough
4. if local is suitable, choose the local profile
5. choose the cheapest suitable tool route by default
6. choose the model for that route
7. choose the effort level for that specific request
8. show the decision and save it with the work

Tool rule: cheapest suitable route by default; a capable local SLM is the
default front line; escalate visibly when local quality risk is too high. For
coding work, add continuity bias, route-switch cost, quota headroom, and
iteration economics so Jini does not churn routes that are still good enough.

Effort levels are `low`, `medium`, `high`, `extra high`: normal work ->
`medium`; quick asks -> `low`; deeper/rigorous asks -> `high`; benchmark,
architecture, root-cause, release-readiness, or exhaustive asks ->
`extra high`.

Visibility and persistence: keep the chosen tool, model, effort, local
profile, and reasons visible; persist tool label, model label, effort level,
and selection reasons with the work item so later screens stay honest and
routes do not drift silently. Coding persistence also supports route
continuity, remembered override tendencies by cohort, and explicit
route-switch reasons.

### Device capability routing (absorbed from device-capability-routing.md)

Local routing must consider task shape, device hardware, OS and OS version,
installed local runtime stack, and measured local reliability together —
task-only routing is not enough.

- Probe and persist a versioned repo-local device profile (OS, OS version,
  architecture, CPU count, memory, accelerator class, runtime class, derived
  device class, profile availability, endpoint signature, Jini version,
  registry version, timestamp).
- Device classes: `mobile-small`, `tiny`, `laptop-light`, `laptop-pro`,
  `workstation`, `gpu-heavy` (`laptop-strong` remains an alias for
  `laptop-pro`). Mobile devices stay in the `mobile-small` envelope and do not
  expose workhorse, deep, or multimodal desktop-local profiles.
- Local profiles `local-fast`, `local-workhorse`, `local-deep`,
  `local-multimodal` resolve to `available`/`limited`/`unavailable` as the
  intersection of hardware potential, runtime presence, and configuration.
- Re-probe when the cached profile is stale or when the Jini version,
  capability registry, OS, architecture, runtime, endpoint, or profile mapping
  changes — Jini rides newly unlocked capabilities instead of freezing to the
  first install state.
- Trust: when local SLM is active, the user can see device class, accelerator
  class, runtime class, profile, model, and why.
- Cost: cheapest suitable route first, but "cheapest suitable" must be
  device-aware — an unusably slow or unstable local route is not cheap in
  productivity terms.

### Research-informed heuristics (absorbed from research-informed-heuristics.md)

Selective adoption of agent-pattern research (ReAct, Self-Refine, Reflexion,
Plan-and-Solve, Self-Consistency, Toolformer):

- Hidden plan-first for clearly multi-step work; the user sees progress and
  outcomes, never a visible planner mode or `Thought/Act/Observe` traces.
- Selective refinement: default path is single-pass plus light structure; add
  one focused refine pass only when the route scorer predicts material
  benefit; cap refinement depth tightly.
- Selective consistency checks only for high-risk work (architecture choices,
  benchmark claims, release-readiness judgments, conflicting evidence, low
  confidence on a deep request) — off by default; never for everyday work. In
  the current implementation this is a second independent draft, not a
  multi-agent debate. Multimodal judging is subtype-aware (PDF/scan,
  screenshot/image, audio/transcript) for verification rubric, route choice,
  and local benchmark memory alike.
- Connector-aware and cohort-aware route learning: connectors are part of the
  route scorer; remember route quality by cohort, profile, and connector
  context; track accepted/edited/replaced/shared/exported outcomes; decay
  stale evidence; promote recovered routes faster on strong new evidence.
- Verification is adaptive by effort level: `low` single pass; `medium` single
  pass plus structure; `high` one refine pass or a stronger route;
  `extra high` selective multi-sample verification and/or stronger route.
- Do NOT adopt: default visible chain-of-thought traces, always-on multi-agent
  debate, always-on self-critique, expensive majority-vote reasoning for
  everyday work, or mode sprawl in the user-facing command surface.
- A heuristic change fails gate review if it adds cost without measurable
  acceptance gain, adds visible complexity to the normal flow, adds route
  jargon to the beginner surface, increases verification depth for low-risk
  work, or weakens the local-first cheap-path principle without strong
  evidence.
