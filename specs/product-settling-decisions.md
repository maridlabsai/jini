# Product Settling Decisions

Updated: 2026-06-07

This document records the hard product decisions that reduce ambiguity for GTM,
engineering, docs, and tiering.

If this document conflicts with exploratory specs, examples, demo-flow docs, or
older PRDs, this document wins unless a newer explicit product decision says
otherwise.

## Canonical Category

Jini is a CLI-first AI work router and durable session layer.

It helps people who already use multiple AI tools, CLIs, models, and local
runtimes finish work with less context loss, lower token spend, fewer throttle
interruptions, and cleaner cross-surface continuation.

Jini is not trying to be:

- a general chatbot
- a travel app
- a meeting-notes app
- a project-management app
- a new command grammar users must learn before value
- a visible agent-role theater in the free tier
- a replacement for Claude Code, Codex, or other configured downstream CLIs

Those tools and flows can be routes, adapters, or proof scenarios. They are not
the product identity.

## GTM Wedge

The first product people should notice is the CLI.

The CLI must make these jobs obvious:

- start from a natural task
- edit local files when the request clearly asks for it
- route to the right configured CLI, provider, or local model
- inspect and switch routes with `jini route`
- preserve a durable work thread across continuation
- keep token cost low by reusing saved state instead of replaying transcripts
- avoid throttling and local-device waste when routing can prevent it

Anything that does not improve that wedge is not P0 for GTM.

## Free Tier Decision

The free tier should prove Jini's routing and session value without giving away
the commercial OS.

Free tier includes:

- CLI-first direct task intake
- local preview and configured-route visibility
- manual route switching
- compact status, continue, open, and route inspection
- basic token-frugal session reuse
- clear setup diagnostics

Free tier does not include:

- developer-agent fleets
- tester-agent fleets
- skills-based OS productivity suite
- visible agent trees
- automated company-running workflows
- commercial managed throttle prediction
- commercial cross-device automation policy

## Commercial Tier Decision

Commercial tier is where Jini becomes an agent and skills based OS productivity
suite.

Commercial value must be materially higher than the free CLI:

- managed route policy across teams
- preemptive throttle and quota recovery
- commercial skills and delegation framework
- developer and tester agents hidden behind normal Jini outcomes
- governed approvals and audit trails
- cross-device and offline-online continuation
- company automation loops for support, quality, release, and roadmap work

Commercial interactions must still return normal Jini results. Users should not
need to manage agent role trees to get value.

## UX Decision

No new Jini conversation style.

Jini should align with familiar agent CLI behavior:

- freeform requests execute or answer directly when safe
- bare `jini` starts as a plain task prompt, even when saved work exists
- state inspection is explicit through commands
- route control is explicit through `jini route`
- current work is passive context, not a modal gate
- saved work is resumed through `status`, `continue`, `open`, or natural title matching
- no `Start/Keep` interruption model
- no visible `Switch` startup control
- no full status dump for simple questions
- no product-shaped ceremony before first useful output

If a proposed feature requires teaching new vocabulary before value, it should
be removed, demoted, or hidden behind progressive disclosure.

## Offline Decision

Offline is a route state, not a separate product.

When Jini owns the local/offline route, it must behave as a complete agent CLI
with local model routing, approvals, artifacts, diagnostics, and recovery. When
Jini routes to an already configured online CLI, it should behave like a thin
router and session layer around that CLI.

The user should experience one session either way.

## Roadmap Consequence

Until the CLI wedge is noticeably strong, defer broad expansion.

P0:

- install works without source-build assumptions
- direct file edits work in the current directory
- route list, set, auto, and status are obvious
- current work continuation is compact, familiar, and hidden until requested
- token-frugal context reuse is measurable
- regression gates protect the above

P1:

- throttle-aware route switching
- powered-mode and low-battery route policy
- offline local model quality bars
- cross-surface session handoff

P2:

- desktop and mobile apps
- broad demo verticals
- commercial skills and agent UI surfaces

## Focused Delivery Decision

Jini delivery uses one active chain:

- PRD: `specs/number-one-platform-prd.md`
- dev design: `specs/launcher-intake-design.md`
- implementation plan: `specs/number-one-development-plan.md`

No drift without explicit agreement. A new requirement, product surface,
interaction model, app surface, agent surface, or commercial/free tier boundary
change is not active work until this decision record changes in the same
commit.

Older broad PRDs, research notes, and platform plans are background only. They
may inform decisions, but they do not authorize implementation.

## PRD Drift Control

Protected product and PRD surfaces must not change casually.

Any change to canonical PRD, public positioning, tiering, offline/platform
strategy, skills/delegation boundaries, competitive release pressure, or proof
scenario positioning must update this document in the same change.

The required commit gate enforces this through:

```bash
bash tools/product_prd_drift_gate.sh
```

This makes product drift explicit. If a change does not justify updating the
settled decision record, it should not modify the protected product surface.

## Competitive Scorecard Decision

Competitive scorecards are release gates only when they protect user outcomes.

The scorecard must not pass solely because competitor names are listed. It must
also gate:

- direct current-directory file edits without draft/workflow detours
- compact answers for simple factual questions
- async work receipts with route, model/profile, context, commands, tests,
  blockers, approvals, and rollback evidence
- offline/local route proof or exact setup failure
- adversarial code-review evidence for code-changing cuts
- competitor-watch refresh across benchmark, KPI, release-plan, and gate files
- free/commercial tier boundary protection
- cross-surface session identity
- token-frugal route proof

Outcome gates require executable or named proof references; names alone do not satisfy the scorecard.
Each required outcome must point to a runnable check or a named proof surface that can be inspected later.

New competitor pressure from coding agents, review agents, local/offline tools,
gateways, app builders, and agent frameworks is valid only when it changes one
of those release-critical outcomes. Otherwise it is watchlist noise.

## Release-Facing Copy Decision

Public README and website copy must describe shipped behavior, not future CLI
aspirations.

For `v0.1.2`, the release-facing first-minute story is:

- bare `jini` opens with the compact `Jini` / `Describe the task` prompt
- simple factual questions answer directly
- obvious local text edits update the named file in the current folder
- ambiguous file edits fail closed and list candidate filenames
- saved work stays behind explicit inspection commands or natural title
  matching instead of becoming the startup frame

Do not publish `jini>` live-shell examples, repo-aware startup coaching, or
future prompt behavior in README or website pages until that behavior is
implemented and release-smoked from the public installer.
