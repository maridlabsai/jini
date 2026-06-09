# Product Settling Decisions

Updated: 2026-06-09

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
- route to the right configured CLI, provider API, or local model without
  disguising one as another
- inspect and switch routes with `jini route`
- preserve a durable work thread across continuation
- keep token cost low by reusing saved state instead of replaying transcripts
- avoid throttling and local-device waste when routing can prevent it

Anything that does not improve that wedge is not P0 for GTM.

## Core Development Charter

Claude Code and Codex first-minute parity is the highest-precedence
development bar for Jini.

If this charter conflicts with vertical templates, broad app surfaces, demo
flows, agent-role UX, or roadmap expansion, this charter wins until a newer
explicit product decision changes it.

The streamline-or-rewrite boundary lives in
[product-streamline-redline.md](./product-streamline-redline.md). Current work
continues only while the Go kernel can preserve the familiar first-minute
transcript through clear shell, intent, action, state, and gate boundaries.

The internal engineering operating model lives in
[agentic-development-operating-model.md](./agentic-development-operating-model.md).
It is mandatory for non-trivial Jini engineering cuts, but it is
a coordinator-owned process. It must not become public UX, free-tier command
grammar, or a visible agent-role layer.

Non-negotiable invariants:

- Questions answer compactly and do not create work.
- Bare entities ask for intent and do not create work.
- Explicit task intent edits, runs, routes, or fails closed with the exact
  ambiguity to resolve.
- File and code requests produce real side effects, receipts, or fail-closed
  ambiguity; they never become generic drafts.
- Current work is passive context, not the default frame for unrelated input.
- Route decisions stay inspectable, but routine answers avoid route ceremony.
- Default output follows a lightweight Claude/Codex-style transcript: answer or
  action receipt first, no status or artifact frame unless the user asks for it.
- Simple factual questions never print `Result ready.`, `Task Snapshot`,
  `Saved:`, or follow-on command chrome.
- Configured CLI route names require real installed-CLI handoff or fail-closed
  setup guidance; provider API routing is a separate route type.
- Adapter breadth is P0 only for familiar tools users already trust; it must
  not create a new workflow taxonomy or first-minute command surface.
- Side-effecting work reports files changed, commands or tests run, blockers,
  approvals, and rollback or recovery path when relevant.
- No hard-coded entity-to-template routing.

## CLI Handoff Decision

Aryan's core expectation is that Jini behaves like a familiar terminal agent.

That means:

- local/offline mode must act like a full Jini agent CLI
- configured CLI routes must invoke the installed downstream CLI or fail closed
  with exact setup guidance
- provider API routes may exist, but they must not be marketed as Codex,
  Claude Code, or another CLI unless that CLI is actually invoked
- route output must make the difference between CLI handoff, provider API, and
  local/offline execution visible without adding startup ceremony

This is P0. A provider-backed adapter can be an internal fallback or prototype,
but it does not satisfy the configured-CLI route requirement.

Adapter support ships in waves:

- Wave 0, foundation: one handoff contract, route receipts, `doctor` detection,
  fail-closed setup guidance, and no first-minute adapter taxonomy.
- Wave 1, terminal agents: Codex, Claude Code, Gemini CLI, Aider, and OpenCode.
- Wave 2, local and gateway routes: Ollama, LM Studio, OpenRouter, and LiteLLM.
- Wave 3, editor and cloud agents: Continue, Cline/Roo, Cursor, Windsurf, and
  GitHub Copilot coding agent.

An adapter moves from planned to supported only after smoke tests cover
available, missing, and failed handoff states.

Wave 0 and Wave 1 are runtime-supported in the free CLI when the downstream
tool is installed and trusted by the OS. Missing or rejected CLIs must fail
closed and must not fall back to provider API aliases.

Release claims require a separate dogfood evidence layer. Executable detection
means a route is ready to validate; it does not prove auth, approvals, output
shape, or route-receipt privacy. `jini check ship --format json` must keep
those states separate and may read local `.jini/cli-dogfood.json` evidence for
validated installed-CLI runs.

Implementation plan:

- P0.1: keep a golden CLI transcript suite for simple questions, bare
  entities, direct file edits, current-work interruption, route inspection, and
  explicit vertical opt-in.
- P0.2: run that suite in `bash tools/cli_ux_regression_gate.sh` on every
  commit gate.
- P0.3: require the scorecard to include an intent-first routing outcome gate
  with executable proof references.
- P0.4: block releases that reintroduce status dumps, `Start/Keep`, `Switch`,
  generic drafts for file edits, or template routing from questions.
- P0.5: accept new product surfaces only when they preserve this first-minute
  contract or are hidden behind explicit progressive disclosure.
- P0.6: require adapter support waves to be scorecard-gated before breadth is
  treated as a viral-adoption claim.

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
- intent-first Claude/Codex parity is protected by golden transcript gates
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
- HLD: `specs/number-one-platform-hld.md`
- LLD: `specs/number-one-platform-lld.md`
- streamline redline: `specs/product-streamline-redline.md`
- internal operating model: `specs/agentic-development-operating-model.md`
- front-door dev design: `specs/launcher-intake-design.md`
- implementation plan: `specs/number-one-development-plan.md`

The PRD states product outcomes. The HLD owns architecture boundaries. The LLD
owns executable runtime contracts. Implementation must be traceable to all
three before it can be treated as release-ready.

No release ships unless golden transcript gates prove first-minute quality at
the level users expect from Claude Code, Codex, ChatGPT, and Gemini-style tools.
Architecture quality does not compensate for a bad transcript.

No non-trivial engineering cut is release-ready unless its work split, disjoint
write ownership, integration decision, and evidence are traceable through the
internal operating model.

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

## PRD Sharpness Decision

The canonical near-term PRD must stay smaller than the older platform plans and
must not carry broad aspirational scope as active requirements.

Requirements belong in `specs/number-one-platform-prd.md` only when they are:

- active P0/P1 work for the CLI wedge
- a release-blocking UX, routing, tiering, or gate constraint
- a guardrail that prevents known product drift

Older requirements must be removed or demoted when they:

- imply desktop, mobile, broad agent OS, company automation, or demo verticals
  are active GTM scope
- duplicate tier boundaries already settled in this document
- preserve stale numbering or headings after a higher-precedence charter exists
- make competitor research or user learning sound like automatic feature scope
  instead of gated input to the decision record

## Competitive Scorecard Decision

Competitive scorecards are release gates only when they protect user outcomes.

The scorecard must not pass solely because competitor names are listed. It must
also gate:

- direct current-directory file edits without draft/workflow detours
- compact answers for simple factual questions
- intent-first routing for general questions, bare entities, and explicit task
  requests
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
Named-proof refs must resolve to existing repository files, and executable refs
must name real Go test functions.

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
