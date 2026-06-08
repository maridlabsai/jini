# Website Execution Tracker

Last updated: 2026-06-08

This tracker exists so the public website can move in its own slices instead of
being treated as a side effect of app or billing work.

## Current State

| Surface | Status | What is real now | Next cut |
| --- | --- | --- | --- |
| Homepage | In progress | Product landing now reflects the shipped `v0.1.2` first-minute CLI wedge: compact prompt, simple answers, named local edits, and route-aware work | Add public links to live release packets when public downloads actually exist |
| Commercial page | In progress | Free-vs-paid split, plan comparison, packet-fed surface badges, and more present-tense rollout language now replace hand-maintained status rows | Add public links to live release packets when public downloads actually exist |
| Proof page | In progress | Sanitized public-proof ingestion contract now feeds the page instead of hardcoded snapshot copy | Replace checked-in sanitized proof snapshots with live ingestion once telemetry is ready |
| Install page | In progress | CLI install now shows the actual `v0.1.2` terminal transcript and avoids future prompt-shape claims | Add public links to live release packets when public downloads actually exist |
| Command catalog | In progress | Public command page now keeps support commands behind the shipped `jini` front door and removes future prompt examples | Keep pruning commands that are operator plumbing rather than first-run product surface |

## Website Done Criteria

- the homepage reads like a product site, not a repository front page
- release-facing examples match the public installer binary
- free orchestration value is legible without reading multiple pages
- paid value is proven, not merely described
- app availability remains honest about preview vs live state
- design quality is modern enough to carry the product story without apology

## Current Blocking Reality

- live public proof telemetry is not wired yet
- commercial checkout is still control-plane work, not a live website flow
- app delivery is still preview-only, so website promise boundaries must stay strict
