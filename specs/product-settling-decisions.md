# Product Settling Decisions

Updated: 2026-06-10

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

## Customer Value Bar

Jini is viable only when it makes the user's existing AI tools easier to use
with less wasted context, fewer interruptions, and less tool-switching friction.
It is not viable if it merely adds a new workflow layer, a new command grammar,
or a polished-looking scaffold around unsupported behavior.

Every product, routing, CLI UX, or docs cut must map to at least one customer
value outcome:

- less token spend through compact context reuse instead of transcript replay
- fewer throttle stalls through route choice, fallback, or explicit setup
  guidance
- faster completion through direct local action or correct downstream handoff
- safer side effects through clear approvals, receipts, and recovery paths
- lower switching cost across Claude Code, Codex, Gemini CLI, Aider, OpenCode,
  local models, provider APIs, and gateways the user configured

Anti-amateur constraints:

- Do not claim Kiro-scale IDE, team platform, or managed automation scope while
  shipping only a CLI router.
- Do not claim support for a framework, model, or CLI unless Jini can detect it,
  invoke it when selected, and fail closed when it is unavailable.
- Do not answer with generic templates, dashboards, task snapshots, or workflow
  ceremony when the user asked a simple question or clear file/task action.
- Do not hide fake intelligence behind hard-coded entity routing or canned
  vertical flows.
- Do not ship commercial-suite language into the free CLI unless entitlement,
  fallback, and customer-value boundaries are implemented.

The customer-value gate is required on every commit. If a cut cannot explain its
value in terms of token savings, throttle resilience, tool-switching reduction,
direct action, safety, or session continuity, it should not ship.

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
- Shell output must stay precise and professional: name the actual action,
  artifact, route, blocker, or next command instead of vague readiness,
  conversational filler, or product ceremony.
- Stale shell vocabulary is disallowed in runtime output; the CLI UX gate owns
  the forbidden list and must fail when old scaffolding reappears.
- Simple factual questions never print `Result ready.`, `Task Snapshot`,
  `Saved:`, or follow-on command chrome.
- Unknown standalone questions route through a configured CLI, provider, or
  local model when one is available; otherwise they return compact setup
  guidance and still create no work.
- Configured CLI route names require real installed-CLI handoff or fail-closed
  setup guidance; provider API routing is a separate route type.
- Adapter breadth is P0 only for familiar tools users already trust; it must
  not create a new workflow taxonomy or first-minute command surface.
- Side-effecting work reports files changed, commands or tests run, blockers,
  approvals, and rollback or recovery path when relevant.
- No hard-coded entity-to-template routing.

## Dynamic Platform Decision

Jini should remove hard-coded scaffolding wherever it affects first-minute
behavior, route claims, feature availability, or pricing boundaries.

Dynamic platform rules:

- Routing is registry-backed. Adapter descriptors, installed CLI checks,
  provider readiness, local runtime probes, model/profile eligibility, health
  history, and user route choices feed route selection.
- Route scoring must be capability-gated before it is preference-scored. An
  unavailable local profile, rejected CLI, missing provider, or denied paid
  feature cannot win because of score bias.
- Graceful degradation is explicit. Jini chooses the next safe configured route
  or fails closed with setup guidance; it must not fake support with a generic
  scaffold.
- Subscription boundaries happen at the feature boundary. Installing and using
  the core CLI stays available; commercial-only feature names fail closed in the
  public CLI until managed commercial automation has real entitlement checks and
  manual free fallbacks where possible.
- User and work context learning is bounded to repeated preferences, route
  choices, local runtime outcomes, and resumable work. It must be inspectable
  and must not become hidden surveillance or public agent-role UX.

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
- route setup guidance is public through `jini route help`; it lists CLI
  handoff env vars and provider/local route setup without printing secrets

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
- P0.2a: keep `specs/claude-codex-prompt-bank.jsonl` at exactly 100 Aryan-derived prompts with diverse domain, age, gender, and race or ethnicity coverage.
- P0.2b: run the prompt bank validator through `bash tools/claude_codex_usecase_gate.sh` on every commit gate.
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
future commercial automation.

CLI is available now. App surfaces, when shipped, are available to both free and
commercial users. Download, install, launch, and basic session review are not
the subscription boundary. Subscription gates capabilities inside those
surfaces.

Free tier includes:

- CLI and app access when each surface is live
- direct task intake
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

Commercial tier is the future managed automation layer for skills, delegation,
agents, policy, and continuity.

Commercial value must be materially higher than the free surfaces:

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

Offline local-model selection is Jini intelligence, not user homework. Default
auto mode should discover common loopback OpenAI-compatible local runtimes,
inspect available chat models, and choose the smallest suitable profile for the
task, device class, and battery posture before spending remote tokens. Explicit
route, provider, or model pins remain overrides. Zero-config discovery does not
authorize silent model downloads or bundling large model assets into the CLI
installer.

Offline mode is active when no usable remote provider or CLI API is configured
or when connectivity is unavailable. CLI and app surfaces must read the same
runtime availability signal and route engine. Automatic remote routes may fail
over to local SLM or local preview after a network-class provider failure;
explicit user-pinned remote routes must fail closed instead of silently
switching providers. Route-time connectivity detection should use a fast
no-payload network-route probe; generation-time provider errors remain the
source of truth for DNS, provider, captive-network, and mid-flight failures.

Discovered local models must obey form-factor envelopes. Mobile-class devices
may use only lightweight fine-tuned local SLMs for bounded work and must not
silently expose desktop workhorse, deep, or multimodal profiles. Laptop-class
devices split into light and pro policy tiers from measured SKU signals; light
laptops avoid pro-sized discovered models, while pro laptops may use stronger
mid-size local workhorse models when the device and power posture support it.

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
- macOS app PRD, UX design, HLD, and LLD for a Codex desktop-caliber surface
  over the same Go session graph

P2:

- Windows and mobile apps
- broad demo verticals
- commercial skills and agent UI surfaces

The macOS app may proceed only as a focused desktop surface over the existing
CLI/session/router core. It must use the same session model as the CLI, preserve
the lightweight Claude/Codex transcript contract, and avoid Start/Keep, Switch,
dashboard-first startup, generic drafting shells, or free-tier agent-role UX.

## Focused Delivery Decision

Focused implementation is the development philosophy. Every cut should be the
smallest change that advances the active CLI wedge or removes a release-blocking
quality risk.

Jini delivery uses one active chain:

- PRD: `specs/number-one-platform-prd.md`
- macOS app PRD: `specs/macos-app-prd.md`
- macOS app UX design: `specs/macos-app-ux-design.md`
- macOS app HLD: `specs/macos-app-hld.md`
- macOS app LLD: `specs/macos-app-lld.md`
- HLD: `specs/number-one-platform-hld.md`
- LLD: `specs/number-one-platform-lld.md`
- streamline redline: `specs/product-streamline-redline.md`
- internal operating model: `specs/agentic-development-operating-model.md`
- front-door dev design: `specs/launcher-intake-design.md`
- implementation plan: `specs/number-one-development-plan.md`

macOS app design review feedback is active scope only where it hardens the
existing macOS PRD/HLD/LLD chain: transient simple answers, session identity
mapping, macOS file-access posture, sidecar protocol idempotency/replay,
approval scope/audit contracts, and removal of stale or ambiguous app wording.

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
