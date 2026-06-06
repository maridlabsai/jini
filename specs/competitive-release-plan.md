# Competitive Release Plan

Updated: 2026-06-06

This document converts current competitor pressure into release-plan decisions.

It is a supporting strategy document, not the top-precedence product and
operating PRD. The canonical PRD remains
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this document conflicts with the canonical PRD on tenets, requirements,
roadmap order, free-tier boundaries, automation posture, or app-shipping order,
the canonical PRD wins and this document should be reconciled.

## Scope

Jini competes with more than coding agents.

The competitive universe now includes:

- terminal and IDE coding agents
- cloud coding agents that create pull requests
- GitHub-native agent assignment and agent marketplaces
- local/offline model front doors
- LLM gateway and routing infrastructure
- app builders that turn prompts into deployable software
- general autonomous agents and workflow automation products

The goal is not to clone each surface.

The goal is to identify which competitor behavior changes Jini's release
priority, gate, or proof burden.

## Competitive Universe

### Direct Replacement Threats

These products can plausibly become the user's daily AI work front door.

- Claude Code: terminal-native coding agent with memory, hooks, subagents,
  MCP, plugins, and mobile or web remote control for local sessions
  ([overview](https://code.claude.com/docs/en/overview),
  [remote control](https://code.claude.com/docs/en/remote-control),
  [memory](https://docs.claude.com/en/docs/claude-code/memory),
  [hooks](https://code.claude.com/docs/en/hooks)).
- OpenAI Codex: local, desktop, remote, GitHub, and mobile-connected coding
  workflow with review and approval loops
  ([mobile and remote work](https://openai.com/index/work-with-codex-from-anywhere/),
  [Codex upgrades](https://openai.com/index/introducing-upgrades-to-codex/)).
- GitHub Copilot coding agent and Agent HQ: GitHub-native assignment,
  ephemeral GitHub Actions environment, PR creation, model selection, security
  scanning, and multi-agent access to Copilot, Claude, and Codex
  ([coding agent](https://docs.github.com/en/copilot/concepts/coding-agent/coding-agent),
  [Agent HQ](https://github.blog/news-insights/company-news/pick-your-agent-use-claude-and-codex-on-agent-hq/),
  [model selection](https://docs.github.com/en/copilot/concepts/auto-model-selection)).
- Google Jules: asynchronous GitHub-connected coding agent that clones repos in
  a cloud VM, works autonomously, and returns verified changes
  ([docs](https://jules.google/docs/),
  [product](https://jules.google/)).
- Cursor: IDE-native agent modes, background agents, memory/rules, local
  checkpoints, and PR review through Bugbot
  ([agent modes](https://docs.cursor.com/agent),
  [background agents](https://docs.cursor.com/background-agent),
  [rules](https://docs.cursor.com/context/rules),
  [Bugbot](https://docs.cursor.com/bugbot)).
- Windsurf and Devin Desktop: IDE plus agent command center, workspace memory,
  rules, workflows, skills, local Cascade sessions, and cloud Devin sessions
  ([Cascade](https://windsurf.com/cascade),
  [memories and rules](https://docs.windsurf.com/windsurf/cascade/memories),
  [Windsurf product](https://windsurf.com/)).
- Kiro: spec-first agentic IDE with specs, steering, hooks, MCP, privacy
  controls, and structured task execution
  ([docs](https://kiro.dev/docs/),
  [specs](https://kiro.dev/docs/specs/),
  [steering](https://kiro.dev/docs/steering/)).
- JetBrains Junie: IDE-native and CLI coding agent that plans, edits, runs
  tests, and reports progress inside JetBrains workflows
  ([Junie docs](https://www.jetbrains.com/help/ai-assistant/junie-agent.html),
  [Junie product](https://www.jetbrains.com/junie/)).
- Google Gemini CLI: open-source terminal agent that brings Gemini into local
  developer workflows
  ([GitHub](https://github.com/google-gemini/gemini-cli)).
- Zed Agent Panel: editor-native agent thread with code edits, terminal
  context, token visibility, checkpoints, and reviewable changes
  ([Agent Panel](https://zed.dev/docs/ai/agent-panel)).

### Open And BYO Coding-Agent Pressure

These products pressure Jini's free/open posture and provider portability.

- Cline: VS Code agent with file reads, commands, MCP, CLI automation, and
  provider/local-model flexibility
  ([overview](https://docs.cline.bot/introduction/overview)).
- Continue: configurable model, context-provider, local-agent, MCP, and
  self-hosted coding assistant surface
  ([models](https://docs.continue.dev/customize/models),
  [context providers](https://docs.continue.dev/customize/custom-providers),
  [config reference](https://docs.continue.dev/reference)).
- Aider: terminal pair programming in local git repos with model and provider
  configuration
  ([docs](https://aider.chat/docs/)).
- OpenHands: open software-development agent platform and SDK for agents that
  write code, use shell/browser/API tools, and work with repositories
  ([docs](https://docs.all-hands.dev/),
  [SDK](https://docs.openhands.dev/sdk/index)).
- Roo Code: legacy and successor pressure from local/cloud modes, custom modes,
  cloud agents, router, and BYO provider support. Track Roomote below as the
  current operational-agent successor pressure
  ([docs](https://docs.roocode.com/),
  [product](https://roocode.com/)).

### Cloud Autonomous Software Engineers

These products set the bar for "delegate work and come back to a PR."

- Devin: autonomous software engineer for tickets, bugs, features, and internal
  tools
  ([docs](https://docs.devin.ai/),
  [Cognition](https://cognition.ai/)).
- Factory Droids: supervised-to-autonomous coding agents across design docs,
  implementation, and review automation
  ([Droids](https://factory.ai/product/droids)).
- Roomote: interrupt-work agent for support escalations, regressions, repo
  questions, and operational bugs that connects to team systems and hands back
  reviewable work
  ([product](https://roomote.dev/)).
- Replit Agent: plain-language app builder that plans, implements, checks,
  fixes, and deploys
  ([Agent docs](https://docs.replit.com/core-concepts/agent),
  [web apps](https://docs.replit.com/replitai/web-apps)).

### Prompt-To-App Builders

These products compress first useful output for builders and non-engineers.

- Lovable: full-stack AI development platform for building, iterating, and
  deploying web apps with real code and governance
  ([docs](https://docs.lovable.dev/)).
- Bolt.new: prompt, run, edit, and deploy full-stack web apps in the browser
  ([GitHub](https://github.com/stackblitz/bolt.new)).
- Vercel v0: natural-language code and UI generation with deployment to Vercel
  ([docs](https://vercel.com/docs/v0)).
- Base44: AI-powered website and app builder
  ([docs](https://docs.base44.com/Getting-Started/Quick-start-guide)).

### Local And Offline Front Doors

These products commoditize local model setup. Jini should integrate where they
help, not rebuild their model hosting surface.

- Ollama: local model runtime with OpenAI compatibility and integrations
  ([OpenAI compatibility](https://docs.ollama.com/openai)).
- LM Studio: local desktop and server surface for OpenAI-like local endpoints,
  offline use, and local-network serving
  ([docs](https://lmstudio.ai/docs),
  [offline operation](https://www.lmstudio.ai/docs/app/offline)).
- GPT4All: private local desktop LLM and LocalDocs RAG surface
  ([docs](https://docs.gpt4all.io/),
  [LocalDocs](https://docs.gpt4all.io/gpt4all_desktop/localdocs.html)).
- Jan: open-source ChatGPT alternative with desktop, web, server, CLI, local
  OpenAI-compatible server, and offline support
  ([docs](https://www.jan.ai/docs)).
- Open WebUI: self-hosted, offline-capable, provider-agnostic UI for local and
  cloud models
  ([docs](https://docs.openwebui.com/)).
- AnythingLLM: local/offline desktop and agent/doc workspace across local and
  cloud LLM engines
  ([product](https://anythingllm.com/)).

### Routing And Gateway Infrastructure

These products pressure Jini's commercial route optimization story.

- LiteLLM: unified OpenAI-format gateway for 100+ models, retry/fallback,
  load-balancing, spend tracking, budgets, rate limits, and observability
  ([docs](https://docs.litellm.ai/)).
- OpenRouter: unified API and model catalog with hundreds of models and
  OpenAI-compatible request shape
  ([models](https://openrouter.ai/docs/guides/overview/models),
  [API reference](https://openrouter.ai/docs/api/reference)).

### General Agents And Workflow Automation

These products pressure Jini's "company operations should become productized
loops" requirement.

- Manus: general AI agent for autonomous task execution
  ([product](https://manus.is/)).
- Genspark Super Agent: no-code autonomous assistant for real-world tasks such
  as calls, presentations, video, and research
  ([OpenAI case study](https://openai.com/index/genspark/),
  [Genspark blog](https://www.genspark.ai/blog/genspark-super-agent)).
- Lindy: workflow agents built from triggers, actions, conditions, and
  integrations
  ([workflow docs](https://docs.lindy.ai/fundamentals/lindy-101/create-agent)).
- Zapier Agents: AI agents over Zapier's app/action network
  ([docs](https://help.zapier.com/hc/en-us/articles/24393442652557-Build-an-agent-in-Zapier-Agents)).

## Release Lessons

### 1. Async delegation is table stakes

Codex, Copilot, Jules, Cursor, Devin, Factory, and Replit all push users toward
delegating work and returning later.

Jini response:

- P0 release work must prove issue or ask to artifact or PR continuity.
- Mobile and desktop review must become real approval and steering surfaces,
  not read-only status viewers.
- Every async run must produce a receipt: route, model/profile, source context,
  commands run, artifacts changed, tests/gates, blockers, and next approval.

### 2. Cross-surface continuity is now a direct battleground

Codex and Claude Code both now emphasize continuing local or remote work from
mobile/web. GitHub Agent HQ brings agents into GitHub, VS Code, and mobile.

Jini response:

- P0 session graph work moves ahead of broad workflow expansion.
- CLI, desktop, and mobile must all operate over the same session and artifact
  identity.
- Offline and degraded-mode continuation is a release gate, not an app polish
  item.

### 3. Rules, memories, skills, and scoped context are becoming productized

Claude Code, Cursor, Windsurf, Kiro, Continue, and Roo Code all expose some mix
of memory, rules, skills, hooks, scoped configuration, or context providers.

Jini response:

- Context routing discipline remains P0.
- The root context map must stay small.
- Domain files, skills, evidence, metrics, and hypotheses must load on demand.
- Release gates must detect global-context bloat and rule drift.

### 4. Local model hosting is becoming a commodity

Ollama, LM Studio, GPT4All, Jan, Open WebUI, and AnythingLLM are better default
hosts than a new Jini model-hosting surface.

Jini response:

- Jini should adopt local model hosts through adapters and a device-aware model
  registry.
- Jini should not build a new local-model desktop just to match them.
- Commercial value should come from route policy, fallback, savings proof,
  session continuity, and managed recovery.

### 5. Routing gateways commoditize API switching

LiteLLM and OpenRouter already normalize model access, routing, fallback, spend,
and provider diversity.

Jini response:

- Jini must not sell "multi-provider API access" as the differentiator.
- Jini must own work-aware route choice: task class, budget, local/offline
  state, user approval posture, evidence needs, and continuation risk.
- LiteLLM/OpenRouter-style gateways should be optional route adapters, not the
  product brain.

### 6. App builders win first result, but lose durable work by default

Lovable, Bolt, v0, Base44, and Replit compress prompt to running app.

Jini response:

- Jini should not race them on greenfield app generation before flagship loops
  win.
- Jini should absorb their lesson: first useful artifact must appear fast.
- Jini's advantage should be continuation, review, proof, governance,
  handoff, rollback, and session memory after the first artifact exists.

### 7. Security and governance must be visible before autonomy expands

GitHub's coding agent exposes built-in security scanning, and the broader field
is moving more agent work into GitHub Actions, cloud VMs, and IDE sandboxes.

Jini response:

- Security scanners, secret scanning, source-backed research, and release gates
  stay part of the required engineering process.
- Agent instructions from untrusted repositories, docs, comments, or generated
  artifacts must be treated as data unless explicitly admitted as trusted rules.
- Jini must show which instructions and memories were used before high-impact
  actions.

## Requirement Rejection Filter

Competitor research must also remove or block bad requirements.

A competitor capability is not a valid Jini requirement unless it improves at
least one replacement-critical claim without weakening another:

- time to first useful result
- cost per successful task
- cross-surface resume
- async delegation
- local/offline continuity
- GitHub-native handoff
- flagship outcome quality
- visible trust
- free local/BYO usefulness

Reject or downgrade requirements that mainly create:

- feature-count parity without a better user outcome
- a new IDE, browser, local model host, or generic gateway surface that should
  be an adapter instead
- app-specific memory, routing, approval, or artifact semantics
- paid lock-in for local/BYO use, basic continuity, manual route switching, or
  structural token savings
- global context growth when scoped routing, skills, or domain files would work
- autonomy without receipts, approvals, rollback, source trust, and failure
  visibility
- prompt-to-app breadth before flagship quality and continuation proof
- user-facing surface churn that can be solved through policy, learning, or
  adapter changes

Each monthly release plan must mark competitor-derived requirements as:

- adopt: build because it strengthens a replacement-critical claim
- integrate: connect to an existing commodity surface behind Jini's contract
- watch: monitor without changing the current roadmap
- reject: explicitly block because it weakens the PRD
- delete: remove an existing Jini requirement that competitor evidence proves
  is copycat bloat or product debt

The required classification set is adopt, integrate, watch, reject, or delete.

If a requirement cannot survive this filter, keeping it in the PRD is itself a
competitive risk.

## Release Plan Changes

### P0: Next Release Train

Add these items to the active release plan:

1. Competitive watch packet.
   - Weekly packet covering direct replacement threats, local/offline hosts,
     gateways, app builders, and general workflow agents.
   - Each packet records changed competitor capability, source URL, Jini
     impact, required release-plan change, and whether the benchmark/watchlist
     changed.
   - Each packet classifies every proposed requirement as adopt, integrate,
     watch, reject, or delete.

2. Async work receipt.
   - Every delegated or resumed work item gets a machine-readable receipt:
     route, model/profile, context files loaded, artifacts changed, tests run,
     gates run, blockers, approvals, rollback path.

3. Cross-surface session proof.
   - One test fixture must prove the same session can start in CLI, be
     inspected on desktop, be approved or deferred on mobile, and resume on CLI
     without losing artifact identity.

4. Local model host adapters.
   - Keep OpenAI-compatible local routing.
   - Prioritize adapter proof for Ollama and LM Studio first.
   - Treat GPT4All, Jan, Open WebUI, and AnythingLLM as watchlist adapter
     candidates, not P0 runtime dependencies.

5. Gateway adapter boundary.
   - Add explicit support-policy language for optional LiteLLM/OpenRouter
     routes.
   - Keep Jini route policy above the gateway: the gateway may execute route
     switching, but Jini decides why the switch is justified.

6. GitHub-native agent parity slice.
   - Add issue to work-thread to PR/review readiness as a benchmark scenario.
   - Compare against Copilot coding agent, Codex, Claude Code, Jules, and Devin
     style flows.

7. Context-bloat gate.
   - Release cannot add a giant global instruction file as a shortcut.
   - New rules must declare routing scope, evidence status, and load trigger.

### P1: Following Release Train

Add these items after P0 proof exists:

1. IDE and editor bridge strategy.
   - Evaluate VS Code, JetBrains, and Zed/ACP bridges as views over the same
     session graph.
   - Do not create an IDE clone.

2. App-builder handoff lane.
   - Support importing a generated app or design artifact into Jini for review,
     gate checks, cleanup, deployment readiness, and future continuation.
   - Do not make prompt-to-app generation a flagship flow yet.

3. Managed route optimizer.
   - Paid commercial layer may predict throttles, switch among configured
     tools/CLIs/providers/local profiles, and auto-resume.
   - Free layer must keep manual route health, manual switching, local/BYO, and
     saved-state resume useful.

4. Competitor scenario expansion.
   - Add benchmark scenarios for async delegation, GitHub-native PR creation,
     local/offline execution, model-route failover, mobile approval, and
     app-builder handoff.

### P2: Expansion

Add these only after flagship and continuity proof is strong:

1. General workflow-agent integration.
   - Lindy, Zapier Agents, Genspark, and Manus-style flows become integration
     pressure, not core identity.

2. Team operating loops.
   - Product intake, support follow-up, release prep, model watch, and
     competitor watch become Jini automations with receipts.

3. Domain expansion.
   - Add new packs only when they reuse the same session, artifact, routing,
     and proof contracts.

## Benchmark And Watchlist Changes

The core benchmark should stay small enough to run repeatedly.

The watchlist should expand.

Required watch categories:

- terminal coding agents
- IDE-native coding agents
- cloud PR agents
- GitHub-native agent assignment
- local/offline model front doors
- routing gateways
- app builders
- general autonomous agents
- workflow automation agents

Promotion rule:

- A watchlist competitor becomes core only when it threatens one of Jini's
  replacement-critical claims: time to first useful result, cost per successful
  task, cross-surface resume, async delegation, local/offline continuity,
  GitHub-native handoff, or flagship outcome quality.

## Free And Commercial Boundary

Competitive pressure strengthens the existing free-tier rule.

Free must include:

- CLI access
- local/BYO routes
- manual route switching
- basic route and throttle health
- saved-state resume
- structural token savings
- offline/degraded-mode visibility

Paid may include:

- preemptive throttle avoidance
- managed route switching
- cross-surface managed sync and reconciliation
- learned route/compression policy
- auto-resume after failure or limits
- team governance and receipts
- competitor-watch automation for teams

Jini should not monetize basic local usefulness or basic continuity. That would
make the free tier weaker than the open competitors and would damage trust.

## Kill List

Do not:

- build a new IDE to chase Cursor, Windsurf, Kiro, JetBrains, or Zed
- build a new local model host to chase Ollama, LM Studio, GPT4All, Jan, Open
  WebUI, or AnythingLLM
- sell generic multi-provider API switching as the main commercial value
- chase prompt-to-app builders before flagship work quality and continuation
  win
- add autonomy without receipts, approvals, rollback, and source trust
- let competitor pressure bloat the kernel or global context

## Operating Rule

Every monthly release planning pass must answer:

- which competitor changed
- what source proves the change
- which Jini claim is now under pressure
- which P0/P1/P2 item changes
- which benchmark or watchlist entry changes
- what not to copy

If those answers are missing, the competitor update is market noise, not a
release-plan input.
