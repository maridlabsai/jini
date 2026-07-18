# Jini Platform PRD

Updated: 2026-07-10

This is the canonical PRD, rebuilt from [prd-rebuild-design.md](./prd-rebuild-design.md);
decisions are recorded in [product-settling-decisions.md](./product-settling-decisions.md).
If this document conflicts with exploratory specs, demo docs, or older PRDs, this document
and [product-settling-decisions.md](./product-settling-decisions.md) win.

Delivery chain: this PRD defines what matters; [number-one-platform-hld.md](./number-one-platform-hld.md) defines architecture
boundaries; [number-one-platform-lld.md](./number-one-platform-lld.md) defines runtime contracts;
[launcher-intake-design.md](./launcher-intake-design.md) and [number-one-development-plan.md](./number-one-development-plan.md) define execution;
[macOS app planning](./macos-app-prd.md) is a focused P1 desktop surface over the same session model;
[jini-architecture-blueprint.md](./jini-architecture-blueprint.md) carries the v1 engineering contracts;
drift requires [product-settling-decisions.md](./product-settling-decisions.md).

## Product Thesis

Jini is the coding agent CLI that is free to run: best-in-class local models
built in, your own keys and tools when you need more, and receipts for every
dollar it saves you.
Under the hood, Jini is a CLI-first AI work router and durable session layer
for people already using multiple coding CLIs, online models, and local models.

The wedge is token-savings mastery. Saving tokens is on the table at every
company that uses AI; the tool that masters that art with receipts sells
itself. Jini treats every token as money: route down before routing up, reuse
before regenerate, compact before send, and prove the savings after every
task. Three axes no incumbent follows: free to run, routes everything the user
already has, and proves its economics. Agent incumbents cannot sell token
frugality — their business is metered tokens; pure routers are not agents.

The near-term product is not the broad OS. The bar is commercial-grade: the
primary user is a professional developer at a FAANG-class company who compares
every interaction against frontier-lab tooling; within the first minute they
must not be able to tell Jini is not a first-party tool. Anything that reads
as amateur is a defect, not a polish item.

Core charter: intent-first Claude/Codex parity outranks feature expansion.

Customer value bar: every shipped cut must improve token frugality, throttle
resilience, tool-switching reduction, direct action, safety, or session
continuity for configured tools the user already trusts. If it does not, it is
not P0 work.

## Target Users And Jobs

Primary: the professional developer at a FAANG-class company — daily Claude
Code/Codex user, hits quota walls and throttles on provisioned tools, works
under org token budgets, often cannot send sensitive code to arbitrary clouds.
Jobs: stretch provisioned quotas; do real agentic work at zero marginal cost
on a local route; never lose a session to a throttle, crash, or tool switch;
show the org, with receipts, what the routing discipline saved.

Also served: the cost-conscious solo developer (dollar savings matter even
more), the multi-CLI power dev using Jini as the front door, and the
offline/privacy dev in local-only mode. Roadmap-stage buyer: the company —
org-wide savings ledgers, policy routing, and quota governance monetize the
same mastery individuals adopt first.

## Goals And Scope

v1 (Option A, wedge first — every goal release-gated):

- Working coding agent on BYO provider keys, gateways, and installed-CLI
  handoff routes; local execution rides detected third-party runtimes
  (Ollama etc.) as a disclosed interim route.
- Token savings as a first-class capability: every token-economy mechanism
  shipped and measured in the ledger.
- Throttle resilience on every route: detection, suggested fallback, and
  session resume free; predictive avoidance and automatic re-route paid.
- Savings ledger with receipts on the face of the product, plus shareable
  proof.
- Durable sessions with route-outcome learning and repo-scoped project
  memory.
- On-the-fly skills and agents creation, free tier.
- Live paywall (Autopilot + Continuity) with fail-closed entitlements.
- TTFV under 5 minutes from install start on a fresh machine (model
  downloads excluded and disclosed; a detected key or CLI serves the first
  task meanwhile).

v1.5: first-party embedded inference runtime and the curated,
benchmark-proven, permissive-license-only model matrix per device class,
downloaded from official sources with the license shown at consent. "Free to
run" marketing waits for v1.5; until then the honest claim is "free with what
you already have." Native Windows may trail macOS/Linux by one release.

Non-goal, permanent: general chatbot/companion use. Conversation exists only
in service of finishing work.

## P0 Outcome Requirements

- Number one requirement — autonomous throttle survival: when any route is
  throttled or quota-limited, Jini detects it, holds the session, and resumes
  work on its own without a human babysitting the terminal. The free tier
  self-resumes on the same route the moment capacity returns and names viable
  fallbacks; paid Autopilot additionally switches to a fallback route mid-task
  and resumes there automatically. An agent that stops and waits for a person
  when throttled is a release-blocking defect.
- Start from a natural task in the current directory.
- Edit local files directly when the ask is clear and safe.
- Fail closed with exact ambiguity when a file or code task is unclear.
- Answer simple questions compactly without creating work.
- Ask intent for bare entities without creating artifacts.
- Route between familiar CLIs, providers, gateways, and local/offline models.
- Treat configured CLI routes as real installed-CLI handoffs, not provider API aliases.
- Keep route, token, and local runtime diagnostics inspectable through `jini route`.
- Reuse durable session context without replaying large transcripts.
- Keep saved work hidden until `status`, `continue`, `open`, `help`, or natural title matching.
- Install from release assets without requiring source builds.
- Produce a receipt for every routed work task: tokens, route, time, dollars,
  counterfactual, and side effects with a rollback path.
- Validate BYO credentials with one live call and a typed error taxonomy;
  store keys in the OS keychain, never plaintext dotfiles.
- Quote the next rung's cost before spending the user's money on escalation.
- Resume sessions after crash, throttle, reboot, or route switch; cross-route
  resume discloses exactly what carried over.
- Create skills and agents from inside a session as plain reviewable files.
- Fail closed on paid features without entitlement; a manual free equivalent
  always exists.
- Preserve customer-value viability: reduce token waste, throttle friction,
  tool-switching cost, completion risk, or unsafe side effects.
- Block regressions with required gates before commit and push.

## UX Contract

Bare `jini` is a task prompt, not a dashboard.

- Keep startup to a minimal task prompt; do not teach a new shell before value.
- Return compact answers or action receipts first, without product ceremony.
- Keep shell output precise: name the answer, action, artifact, route, blocker, or next command.
- Treat stale shell vocabulary as a P0 regression.
- Treat a new freeform request as work to execute, not as a Start/Keep modal.
- Resume saved work only through explicit commands or natural title matching.
- Never show a full current-work overview for a simple factual question.
- Never require users to learn Jini-specific command vocabulary before value.
- no saved-work dashboard on bare startup
- no visible `Switch` startup control
- no `Start/Keep` interruption model
- no Working Draft for obvious file edits
- no verbose Goal/Working-with/status frame for simple questions
- no `Result ready`, `Task Snapshot`, `Saved:`, or `Next: jini ...` shell around
  simple factual questions
- no visible agent-role theater in the free tier
- no hard-coded entity-to-template routing

Terminal experience bar (Editor's-Choice guardrails, CLI-idiomatic):

- Terminal-idiomatic, never a generic wrapper: 24-bit color with graceful
  256/16/NO_COLOR fallback, dark/light terminal detection, correct resize
  handling, standard keybindings, shell completions (zsh/bash/fish), man page.
- Perceived performance is a feature: first paint under 100ms, stream results
  as they arrive, and never a bare spinner — long operations narrate what is
  happening ("validating anthropic key, one test call") with polished
  progress states.
- State-adaptive surface: the interface morphs by work state — idle (compact
  prompt), executing (dense progress with route status), completed (receipt
  with savings footer), throttled (fallback suggestion first), resuming
  (compact resume card with fidelity disclosure).
- Output typography is designed: disciplined ANSI hierarchy, semantic color
  that never carries meaning alone, purposeful whitespace; the savings
  dashboard renders magazine-quality inside a terminal.
- Opt-in system notifications for long-task completion and throttle events;
  strict bell discipline.
- Anonymous by default: no account, signup, or tier selection before first
  value; an account exists only for Continuity sync.
- Accessibility is release-gated: NO_COLOR and --plain modes are first-class,
  output survives screen readers and copy-paste, and every layout respects
  the user's terminal width and font scaling.

## Routing And Resource Policy

Token frugality is P0. Jini must spend context only when it improves the result.

Routing requirements:

- Default to the cheapest safe route that can complete the task.
- Use local/offline routes when they meet the task quality bar.
- Escalate to stronger online routes when correctness, codebase scope, or tool access requires it.
- Label provider API routes separately from CLI handoff routes. A route named `codex` or `claude-code` must invoke that CLI or fail closed with setup guidance.
- Preserve enough session state to continue work without replaying stale chat.
- One route ladder, always visible: local model, BYO provider key, gateway,
  installed CLI handoff. `jini route` lists, sets, pins, and explains.
- Escalation is transparent: when a cheaper rung cannot clear the bar, Jini
  says why and what the next rung costs before spending money.
- BYO compatibility matrix, release-gated: Anthropic, OpenAI, Google, DeepSeek,
  Mistral, Groq, and xAI (Grok) keys; OpenRouter and LiteLLM gateways; subscription-backed
  CLIs (Claude Code, Codex/ChatGPT plans, Gemini CLI). Each shape has a
  validation fixture and a receipt-denomination rule; a shape without a
  passing fixture is not claimed.
- BYO setup is near-zero effort: paste a key or confirm a detected credential,
  one live validation call, done. Failed validation names the exact cause —
  bad key, no quota, wrong region, network — never a generic error.

Throttle resilience is release-gated in v1. Detect configured CLI/provider
pressure autonomously — no human babysitting. The free tier holds the
throttled session, self-resumes on the same route when capacity returns, and
suggests viable fallbacks with clean session resume; paid Autopilot performs
the fallback switching and resume automatically. This is the product's number
one requirement (see P0 Outcome Requirements).

Power awareness is P1. In powered mode, Jini can choose higher-throughput local
or online routes. In low-battery mode, Jini should avoid wasteful local model
loads unless the user explicitly asks for offline work.

## Sessions, Memory, And Context

- Every task lives in a durable session: compact state, not chat history;
  resumable after crash, throttle, reboot, or route switch without transcript
  replay.
- Same-route resume is lossless; cross-route resume is best-effort and says
  so — the resume receipt states exactly what carried over and what did not.
  Silent context loss is a defect.
- Project memory is the context engine: per-repo learned context — build and
  test commands, conventions, architecture summaries, known pitfalls —
  accumulated from sessions and injected selectively by task shape instead of
  re-derived by fresh reads. Plain inspectable files, exportable, deletable.
- Route-outcome learning records which route succeeded for which task shape
  on this device and feeds auto-routing; inspectable via `jini memory`.
- No personal profile building in v1. Project memory is repo-scoped working
  context, not a user profile.

## Savings Ledger And Receipts

The growth engine. Every routed work task produces a cost receipt denominated
in dollars (headline), tokens kept off metered routes, time saved, and
throttles dodged — all measured, never estimated multipliers.

- Literal vs imputed, always labeled: metered API savings are literal at
  posted prices; subscription-route savings are imputed at API-equivalent
  prices and labeled as such. An inflated or mislabeled savings claim is a
  release-blocking defect.
- Savings on the face: every routed work task ends with a one-line savings
  footer; session end shows the roll-up; startup shows a one-line running
  counter. Simple questions stay clean — no footer, no counter.
- Throttle dodges count only observed throttle events continued on a fallback
  without waiting out the reset window; Autopilot's predictive avoidances are
  counted separately and labeled predicted.
- `jini savings` renders a terminal dashboard: headline stats and ANSI trend
  charts by week, route, and repo. `jini savings --report` exports a
  self-contained local HTML deep-dive; `jini savings --share` produces an
  opt-in share card built from a fixed six-field allowlist with zero private
  data. Charts render literal and imputed as visually distinct segments.

## Token Economy

Every token-saving mechanism is a named, testable requirement: route-down
bias, context frugality (targeted reads, no unchanged-file re-reads within a
session), session reuse without transcript replay, compact prompt and response
shapes per task class, and quota stretching (paid routes receive only work
that needs them). Each mechanism's contribution is visible in the ledger, and
a token-efficiency regression on a fixed task corpus fails the gate like any
UX regression.

## Skills And Agents

- Users create skills and agents from inside a session, in natural language
  or via `jini skill new` / `jini agent new`: plain, portable markdown files
  with frontmatter, reviewable and diffable, immediately invocable in the
  same session. Creation runs a secret scrub before writing.
- Skills from repetition, proactively suggested: when a flow shape recurs,
  Jini offers one-tap crystallization showing the estimated per-run token
  saving. Never auto-created.
- Agents dispatch with scoped tool permissions, inherit the approval matrix,
  and can never self-approve side effects. Fan-out requires the same visible
  cost preview as escalation; every agent run itemizes its tokens on the
  parent receipt.
- Creation and use of coding-focused skills and agents is free tier.

## Market And Learning Guards

Competitor watching is a P0 feature-selection loop through [competitive-release-plan.md](./competitive-release-plan.md), but it does not create active scope by itself.

- Competitor watch packets can nominate next feature candidates and deletion candidates.
- Each candidate must be classified as copy, integrate, watch, reject, or delete.
- No competitor finding becomes active scope unless the decision record changes.

User productivity learning remains P0 only when it improves the CLI wedge:

- learn stable user context, usage, habits, and repeated patterns
- produce fewer repeated prompts, better defaults, and better route choices
- keep learning inspectable and controllable
- avoid hidden surveillance, broad OS scope, or free-tier agent-suite creep

## Platform Qualities

These principles guide internals without expanding scope: registry-backed routing,
capability-gated scoring, explicit graceful degradation, fail-closed commercial
feature boundaries until entitlement runtime exists, and bounded/inspectable
user-work context learning.

Commercial-grade platform requirements, release-gated:

- Pluggable: providers and routes land behind a versioned adapter API with a
  conformance test suite; new models are adapters, not core releases. MCP
  client support makes existing MCP servers Jini's tool surface.
- Extensible: stable public surfaces — skills/agents file formats, pre/post
  task hooks, and `--format json` on every command.
- Reliable: crash-safe atomic session checkpointing; crash-free session rate
  of at least 99.5% in release qualification; graceful degradation is a
  tested path.
- Scalable: monorepo-class repositories with sub-second context selection;
  thousands of sessions in the store without degradation.
- Requirement language in this document is contractual: "must" with a gate
  named. Unmeasured claims and research hedging are defects.

## Tier Boundary

CLI is available now. App surfaces, when shipped, are available to both free
and commercial users. Subscription gates capabilities, not the ability to
install or open Jini.

Free proves the wedge: direct intake/edits, manual route inspection/switching,
compact status/continue/open, configured offline/local support, token-frugal
session reuse, full ledger with shareable receipts, on-the-fly coding skills
and agents, and setup diagnostics. Free is a complete product, never a
crippled trial. Free excludes: developer-agent fleets, tester-agent fleets,
commercial skills-based OS productivity suite, managed
throttle recovery, governed approvals, team policy, audit, and automation
loops. Commercial value must be materially higher than the free surfaces:
managed route/throttle policy (Autopilot: predictive throttle avoidance,
throttle-aware route switching, auto-resume, savings optimization), governed
skills/delegation, cross-device and offline-online continuation (Continuity),
refreshed pricing and route-benchmark intelligence feeds, team audit, and
automation loops. Both flagship paid features fail closed without entitlement;
manual free equivalents always exist. The honest upgrade trigger: pay when
route management interrupts you more often than the subscription costs.
Prices and SKUs live in the commercial repo. Commercial UX
must still follow the same rule: simple task in, useful result out.

## Risk Register

Premortem-derived; each risk names its mitigation in this document or the
design record: scope death (Option A cut), first-task failure on local models
(absolute per-device-class quality floor; no model clears it, none is
recommended), savings math debunked (literal/imputed labeling, methodology
audit), model licensing takedown (permissive-license-only matrix, official
sources), thin paywall (enterprise roadmap is the revenue thesis), silent
cross-route context loss (resume disclosure rule), incumbent ships subsidized
local mode (vendor-neutral ladder absorbs it; receipts remain uncopyable),
free/BYO agents close the gap (the combination is the moat, tracked in the
competitive scorecard).

## Roadmap

P0 now: intent-first CLI parity, task-first startup even with saved work, direct
file edit reliability, simple question direct-answer behavior, route
list/set/auto/status, real downstream CLI handoff and adapter support waves,
BYO validation matrix, receipts and savings ledger, sessions with project
memory, skills/agents creation, entitlement fail-closed paywall,
release-asset install, and CLI UX, PRD drift, and scorecard gates in commit gates.

P1 next: throttle-aware route switching, powered-mode and low-battery routing,
offline local-model quality regression harness, cross-surface session handoff,
first-party runtime and curated model matrix (v1.5),
macOS app HLD/LLD for a Codex desktop-caliber session and artifact surface
(GUI UX guardrails in [roadmap-app-ux-guardrails.md](./roadmap-app-ux-guardrails.md)), and
clearer CLI/local runtime diagnostics.

Deferred until decision-record update: Windows/mobile apps, richer commercial
agent and skills UI, team policy controls, broad proof verticals/templates, and
company automation loops.

## Gates

Every commit touching product, CLI UX, routing, or docs must keep these gates green:

- `go test ./...`
- `bash tools/cli_ux_regression_gate.sh`
- `bash tools/claude_codex_usecase_gate.sh`
- `bash tools/customer_value_gate.sh`
- `bash tools/product_prd_drift_gate.sh`
- `jini scorecard-gate --format json`
- `jini check ship --format json` before push/release

Release additionally requires: TTFV measured under 5 minutes on fresh machines
from signed artifacts, first-task success measured in qualification, savings
methodology audit including literal-vs-imputed labeling, paywall fail-closed
verification, and the token-efficiency regression suite green.

The gates prevent verbose startup, Start/Keep modals, stale docs, and broad PRD drift.
No release ships unless competitor-parity golden transcript gates for Claude,
Codex, ChatGPT, Gemini-style, and 100-prompt Aryan-derived first-minute use cases are green.
No commit ships unless the customer-value gate can still prove the solution is
useful, route-backed, and non-amateur rather than a generic workflow shell.
