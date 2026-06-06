# Adaptive Response Rendering Framework

Updated: 2026-06-05

## Purpose

This document defines how Jini should move beyond rigid CLI response templates
while preserving product clarity, testability, artifact continuity, and route
transparency.

This document is a specialized rendering framework, not the top-precedence
product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this framework conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, or operating posture, the canonical PRD wins and
this framework should be reconciled.

The goal is simple:

- keep the data model stable
- let presentation adapt to the user, task, surface, and current stage
- test semantics instead of exact prose
- avoid use-case-specific hard coding in the core product loop

This framework extends:

- [number-one-platform-prd.md](./number-one-platform-prd.md)
- [product-rewrite-contract.md](./product-rewrite-contract.md)
- [conversation-and-artifact-ux.md](./conversation-and-artifact-ux.md)
- [artifact-schemas.md](./artifact-schemas.md)
- [workstream-technical-framework.md](./workstream-technical-framework.md)
- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)

## Product Rule

Jini must harden the semantic contract, not the sentence shape.

The stable contract is what Jini knows:

- user intent
- work thread
- inputs
- artifacts
- decisions
- missing information
- uncertainty
- next action
- route summary
- trust state

The flexible layer is how Jini says it on a specific surface.

## Why This Exists

The current public repo already has strong product primitives: work threads,
artifact cards, route transparency, remembered work, and useful-result-first.
The failure mode is that implementation can still turn those primitives into a
fixed response mold.

Examples of failure:

- every input gets the same "first useful pass" shape
- every continuation repeats the same full status frame
- a greeting creates a work artifact
- a vague request gets a generic draft when one scoping question would improve it
- a detailed request gets blocked by unnecessary intake
- CLI, desktop, and mobile are forced to share the same density

Jini should feel consistent, not repetitive.

## Open Framework Lessons

Jini should learn from open frameworks without adopting their architecture as
the product boundary.

### LangGraph

LangGraph is useful as a reminder that durable workflows need explicit state,
resumability, human checkpoints, and persistence. Jini should adopt the idea of
a durable graph behind the scenes, but keep the user-facing product centered on
work threads and artifacts, not graph terminology.

Reference: https://github.com/langchain-ai/langgraph

### OpenHands

OpenHands shows the value of an agent workspace that can run commands, inspect
state, and make software changes. Jini should adopt the principle of visible,
recoverable work with explicit actions. It should not make terminal-like detail
the normal product surface for non-coding work.

Reference: https://github.com/All-Hands-AI/OpenHands

### Continue

Continue is useful for repo-native customization and checks. Jini should adopt
the idea that teams can encode quality gates in versioned project files. The
public Jini core should provide those gates without making the user understand
internal pack wiring.

Reference: https://docs.continue.dev/

### Mastra

Mastra is useful for composable agent workflows, evals, memory, and tool
integration. Jini should use this as a reminder to separate orchestration from
presentation. A workflow step can be stable while the rendered response changes
by surface and task.

Reference: https://mastra.ai/docs

### AG-UI

AG-UI is useful as a product lesson: agent events and UI rendering should be
decoupled. Jini should move toward an event/envelope model that CLI, desktop,
and mobile can render differently without forking work semantics.

Reference: https://docs.ag-ui.com/

## Core Architecture

Adaptive rendering has four layers.

### Projector And Render Policy Split

The implementation should separate two responsibilities:

- `ThreadProjector`: converts canonical work state, artifacts, turn deltas, route
  facts, and trust facts into a deterministic semantic envelope
- `RenderPolicy`: chooses mode, density, surface emphasis, and visible actions
  from that envelope

This split keeps adaptive behavior testable.

The projector owns truth. The render policy owns emphasis.

Rules:

- the same canonical state must produce the same semantic envelope
- render policy may change order, density, and prominence
- render policy must not reinterpret readiness, approval, evidence, or safety
- renderers must not synthesize continuity from prose summaries when turn
  deltas exist

### 1. Semantic Envelope

The semantic envelope is the canonical meaning of one assistant turn.

Required fields:

- `schema_version`
- `thread_id`
- `turn_id`
- `intent_kind`
- `work_class`
- `stage`
- `complexity`
- `input_quality`
- `surface_recommendation`
- `route_summary`
- `artifacts`
- `decisions`
- `missing`
- `uncertainty`
- `next_action`
- `trust_state`
- `confirmation_required`
- `turn_delta`

Rules:

- renderers consume this envelope, not raw intent strings
- tests assert this envelope first
- prose can vary if the envelope remains correct
- route details must exist even when collapsed in the default UI
- "just finished" and "what changed" come from turn deltas, not static recap
  strings

### 2. Artifact Envelope

The artifact envelope is the canonical payload for openable work.

Required fields:

- `artifact_id`
- `family`
- `title`
- `purpose`
- `status`
- `sections`
- `source_refs`
- `open_decisions`
- `missing_inputs`
- `trust_state`
- `render_hints`

Rules:

- the transcript must not be the only useful output
- every meaningful turn should create, update, or explicitly preserve an artifact
- artifact families are generic, not domain-specific hard-coded branches
- use-case profiles may configure artifact families, but the renderer must not
  contain special-case product behavior for each use case

### 3. Render Request

The render request chooses how to present the envelope.

Required fields:

- `surface`: `cli|desktop|mobile|api`
- `mode`: `first_result|continuation|return_visit|blocked|approval|artifact|recovery|help`
- `density`: `compact|standard|detailed`
- `user_familiarity`: `new|returning|power`
- `risk_level`: `low|medium|high`
- `available_actions`

Rules:

- CLI defaults to compact, useful text and one obvious next action
- desktop defaults to artifact shelf plus working context
- mobile defaults to review, unblock, approve, and resume
- API returns envelopes with render hints, not terminal prose

### 4. Surface Renderer

The surface renderer is allowed to vary:

- section order
- tone
- density
- action labels
- whether route detail is collapsed or visible
- whether artifact previews are inline, card-based, or opened separately

The renderer is not allowed to vary:

- artifact identity
- work-thread identity
- safety state
- route decision facts
- unresolved blockers
- what has or has not been sent, booked, committed, or published

## Response Modes

Jini should choose a response mode from the semantic envelope.

Approved screen modes:

- `preflight`
- `first_result`
- `work_summary`
- `artifact_shelf`
- `ask`
- `multi_thread_home`
- `recovery`

Adaptive rendering chooses which approved screen mode to foreground and which
cards to show. It does not invent a new product model per use case.

### Greeting

Behavior:

- answer like a person
- do not create a work record
- invite the user to give work when ready

Gate:

- a pure greeting must not produce an artifact

### Vague Intent

Behavior:

- make a small useful framing move
- ask one high-impact scoping question when quality would materially improve
- avoid a full artifact if the user has not asked for one

Gate:

- no generic "first useful pass" for a one-word or social input

### Scoped Intent

Behavior:

- produce the first useful artifact quickly
- show assumptions and what remains missing
- avoid blocking on low-impact questions

Gate:

- first result appears before broad status explanation

### Underspecified Complex Intent

Behavior:

- ask one or two scoping questions before drafting if the likely output quality
  would otherwise be generic
- explain why the answer changes the result
- offer a safe default if the user wants to skip

Gate:

- a travel, research, planning, or procurement request must not default to a
  generic plan when scoping inputs would materially change the artifact

### Continuation

Behavior:

- show what changed since the last turn
- show updated artifacts
- avoid replaying the entire summary unless the user asks for status

Gate:

- fewer than 20 percent of normal continuation turns should repeat the full
  recap frame when no major state changed

### Return Visit

Behavior:

- show the current work title, ready artifacts, blockers, and one next step
- make starting new work explicit

Gate:

- user can identify the current work in under 5 seconds in usability testing

### Blocked Or Approval

Behavior:

- present one active ask
- show why it matters
- show what happens if skipped
- keep reversible safety visible

Gate:

- no more than one blocking ask is active in the default UI

## Surface Guidance

### CLI

The CLI should be fast, compact, and direct.

Default shape:

- first useful artifact or concise answer first
- one line for route when route matters
- one active ask or one next action
- optional `jini help` for fuller state

The CLI should not show the full state frame on every launch. Help, return
visits, blocked states, and explicit status requests are the right places for
the fuller frame.

### Desktop

The desktop app should make the artifact shelf the center of gravity.

Default shape:

- left: work threads and active inputs
- center: selected artifact
- right: context, route, missing items, and decisions

The desktop app should feel like a workbench, not a larger chat transcript.

### Mobile

The mobile app should optimize for continuation, review, and unblock.

Default shape:

- current work
- ready artifact preview
- one ask
- approve, defer, revise, or resume actions

Mobile does not need to be the full editing surface in v1.

### API

The API should expose semantic and artifact envelopes. It should not force API
consumers to parse CLI text.

## Testing Strategy

Tests should lock meaning, not phrasing.

In short: test meaning, not prose.

Required test classes:

- semantic envelope tests
- artifact envelope tests
- renderer contract tests
- cross-surface rendering tests
- regression tests for greeting, vague, scoped, underspecified, continuation,
  return visit, blocked, and approval modes
- snapshot tests for artifact payloads, not full prose
- downgrade tests proving public free/local/BYO behavior stays useful

Forbidden test pattern:

- exact full-output golden tests for normal prose unless they target an explicit
  backwards-compatible CLI contract

Allowed test pattern:

- assert required semantic fields
- assert artifact identity and status
- assert missing-state correctness
- assert route facts when route matters
- assert no forbidden internals leak
- assert no work is created for social input

## Migration Plan

### Slice 1: Document And Gate

Add this framework, review, and gate to the public repo. Add doc tests so future
changes cannot remove the contract silently.

### Slice 2: Semantic Envelope Behind Existing Output

Introduce a small internal semantic envelope for the current starter flows. Keep
existing output behavior where needed, but render from the envelope.

### Slice 3: Adaptive CLI Renderer

Add `ThreadProjector` and `RenderPolicy` boundaries. Move greeting, vague input,
scoped intent, and return visit into mode-based CLI rendering. Preserve existing
tests while adding semantic tests.

### Slice 4: Artifact Family Registry

Replace use-case writer functions with generic artifact family builders plus
profile configuration. Profiles may choose artifact families and scoping rules;
the core renderer remains generic.

### Slice 5: Desktop And Mobile Render Contracts

Expose the same envelope to desktop and mobile surfaces. Render differently per
surface while preserving thread, artifact, route, and trust facts.

## Expert Review Synthesis

The architecture, UX, product, and test critique converges on these points:

- artifact shelf must be the product center
- chat should narrate work, not store it
- route and trust detail should be layered, not always fully expanded
- desktop and mobile should share semantics, not identical layouts
- free/public should stay useful through local/BYO, exportable artifacts, and
  transparent routing
- paid/commercial should remove operational friction through sync, hosted
  continuity, collaboration, premium connectors, and governance
- tests must prove semantic correctness without freezing the exact response
  style users see

## Gate Summary

Adaptive rendering is acceptable only if:

- public product invariants still hold
- output is artifact-first when useful work is requested
- pure greetings and social inputs do not create work
- underspecified complex requests ask useful scoping questions before generic
  drafting
- continuation emphasizes what changed
- route and safety facts remain inspectable
- test coverage proves semantics and surface behavior
- no commercial-only strategy or internal notes enter the public repo
