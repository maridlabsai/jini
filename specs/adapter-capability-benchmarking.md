# Adapter Capability Benchmarking

Updated: 2026-05-16

## Purpose

Jini should not route local work only from static machine class heuristics.

It should keep a measured capability report for local adapters so routing can
account for:

- load success
- response latency
- steady-state throughput
- cold-start cost
- structured-output reliability
- basic output quality
- current endpoint/runtime shape

## Core Rule

Routing should use three layers together:

1. task shape
2. device/runtime probe
3. measured adapter capability

If measured capability disagrees with static heuristics, measured capability
should win.

## Adapter Registry

Jini should keep a real adapter registry with, at minimum:

- adapter id
- label
- provider mode
- locality
- cost tier
- modalities
- benchmark support

This becomes the routing source of truth instead of scattered switch blocks.

## Capability Report

Jini should persist a repo-local `local-runtime-capabilities.json` with:

- schema version
- capability registry version
- Jini version
- device probe fingerprint
- endpoint signature
- local runtime class
- per-adapter rows

Each adapter row should include:

- model id
- status
- latency ms
- warm latency ms
- cold-start cost ms
- output chars
- output tokens
- tokens per second
- quality class
- structured reliability
- benchmark timestamp

Jini should also keep a short rolling history per adapter so routing can
notice regressions, instability, or improving trends over time.
Jini should also keep direct cohort history for real Local SLM completions,
grouped by adapter and request cohort such as `trip-itinerary`,
`sendable-followup`, or `build-readiness`.
Jini should also keep explicit per-cohort feedback for Local SLM routes so
user upvotes/downvotes can shape later cohort routing. It should also accept
graded artifact usefulness signals such as `accepted-as-is`,
`needed-light-edits`, and `not-useful` for matching cohorts.
Jini should also learn passively from edit distance between the generated
artifact and the later accepted artifact, so tiny cleanup and substantive
rewrite are not treated as the same outcome.
That passive learning should be section-aware, so title/header cleanup,
supporting-section edits, and core content rewrites can be weighted
differently.
Within core sections, it should also separate wording-only edits from actual
decision or recommendation changes.
Jini should also record downstream artifact outcome signals such as
`used-this`, `shared-this`, and `replaced-this` so actual adoption influences
routing more than edit patterns alone.
It should also record narrow passive workflow signals from repeated artifact
opens, export opens, and substantive reopen-after-rewrite events, so routing
keeps learning even when users never label those outcomes explicitly.
For downstream work outside Jini, this should extend to opt-in observed
external copies, so edits or substantive replacement in those files can inform
the same cohort memory without pretending to monitor arbitrary user activity.
History penalties should be confidence-weighted so one noisy sample does not
count as heavily as sustained degradation across the recent window.
Older degraded samples should also decay so stale problems matter less than
fresh ones.
If a route rebounds strongly after a recent degraded streak, Jini should
recognize recovery and restore score faster than simple penalty decay alone.

## Benchmark Rule

For benchmarkable local adapters, Jini should run a small real request and
measure:

- request success/failure
- cold-call latency
- warm-call latency
- output token throughput
- basic output compliance
- repeatability of structured output across repeated calls

Jini does not need perfect quality scoring here. It needs enough signal to
avoid obviously slow, broken, or weak local routes.

If a Local SLM endpoint is already configured and the report is stale or
missing, Jini should try to warm this capability report in the background
during interactive startup so the next normal request can often use measured
scoring without requiring an explicit doctor step.

## Routing Rule

Measured capability should affect route scoring.

Examples:

- failed local benchmark should heavily penalize that route
- strong and fast local benchmark should boost that route
- weak or very slow benchmark should reduce its score
- high cold-start cost should matter for short tasks
- unstable structured output should matter for artifact-producing work
- repeated regression across recent samples should matter more than one bad run
- one noisy sample should not overcorrect as hard as sustained regression
- stale regressions should matter less than fresh regressions
- strong recovery after recent degradation should restore score faster
- benchmark bias should be weighted by request shape so recovered structured
  drafting performance does not over-promote unrelated work classes
- request-shape weighting should narrow again by artifact family or cohort, so
  readiness-check evidence does not get full credit on unrelated trip or
  narrative-draft tasks unless measured evidence supports that shape
- direct cohort evidence from real Local SLM completions should override
  discounted transfer when a matching cohort history exists
- explicit cohort feedback should bias matching Local SLM cohorts without
  spilling into unrelated cohorts
- graded artifact usefulness feedback should bias matching cohorts more
  precisely than binary voting alone
- passive edit-distance learning should distinguish tiny cleanup from
  substantive rewrite on matching cohorts
- passive section-aware learning should distinguish cosmetic header edits from
  supporting-section changes and core-content rewrites
- semantic section-aware learning should distinguish core wording cleanup from
  actual decision or recommendation changes
- downstream artifact outcome signals should bias matching cohorts more
  strongly than passive edit patterns alone

## Acceptance Criteria

This slice is only complete when:

- adapter registry exists in code
- measured local capability report exists in code
- routing uses measured benchmark bias
- provider doctor exposes measured local capability lines
- tests cover benchmark capture and benchmark-based scoring
- an independent validator gate exists
