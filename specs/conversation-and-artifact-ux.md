# Jini Conversation And Artifact UX

Updated: 2026-05-15

## Purpose

This document defines how Jini should handle conversation, attachments, and
artifacts in the user-facing product.

The goal is to stop Jini from behaving like a loose chat transcript.

Jini should behave like a visible work thread:

- the goal stays visible
- the current state stays visible
- the inputs stay visible
- the produced artifacts stay visible
- the next step stays visible

This spec is informed directly by the Codex collaboration pattern used in this
rewrite:

- durable work thread
- progress updates while work moves
- early concrete artifacts
- explicit missing information
- recoverable snapshots

## Product Rule

Conversation is not the storage model.

Conversation is the narration layer around a work thread.

The durable product objects are:

- `WorkUnit`: canonical protocol root
- `WorkThread`: user-facing runtime projection of that `WorkUnit`
- `InputItem`: user-provided or AI-derived working input
- `ArtifactCard`: openable result object
- `TurnRecord`: one user/AI exchange with structured deltas
- `Ask`: one bounded request for user clarification
- `ProgressSnapshot`: stable summary of where work stands now

## UX Principles

### 1. One visible frame

At all times, Jini should keep this frame visible:

- `Goal`
- `Working with`
- `Just finished`
- `Doing now`
- `Up next`
- `Now`
- `Done`
- `Need`
- `Next`

This frame must survive across turns.

### 2. Artifacts first, chat second

Jini should produce named artifacts as early as possible.

Examples:

- `Sendable Follow-up`
- `Build-Readiness Check`
- `Recommendation Memo`
- `Trip Plan`
- `Budget Sketch`
- `Booking Checklist`

The chat transcript must never be the only place where useful work exists.

### 3. Inputs are first-class

Users must be able to provide more than text:

- pasted text
- files
- images
- audio
- screenshots
- links

Every input becomes a visible `InputItem`.

### 4. Ask only what matters

Jini must not interrogate the user with long question chains.

Rules:

- infer what is safe
- show assumptions explicitly
- ask only for high-impact missing inputs
- keep at most one blocking ask active at a time

### 5. No path-driven normal use

Users should never need filesystem paths, raw artifact ids, or internal pack
structure in the normal flow.

## Runtime Model

## WorkThread

`WorkThread` is the user-facing runtime object for one active piece of work.

It is a projection over one `WorkUnit`.

Required fields:

- `thread_id`
- `work_unit_id`
- `title`
- `goal`
- `current_stage`
- `progress_snapshot`
- `working_with`
- `open_artifact_ids`
- `active_ask_id`
- `current_route`
- `last_turn_id`
- `created_at`
- `updated_at`

Rules:

- one `WorkThread` maps to exactly one active `WorkUnit`
- users interact with `WorkThread`, not raw `WorkUnit` internals
- a user may have many `WorkThread` objects, but one is current focus

## TurnRecord

One `TurnRecord` captures one user input and Jini response cycle.

Required fields:

- `turn_id`
- `thread_id`
- `user_input_ids`
- `assistant_message`
- `artifacts_created`
- `artifacts_updated`
- `state_changes`
- `asks_opened`
- `asks_resolved`
- `route_decision`
- `started_at`
- `completed_at`

Rules:

- turns must record what changed, not just free text
- the user should be able to see which artifacts changed on a turn

## InputItem

`InputItem` represents any working input.

Required fields:

- `input_id`
- `thread_id`
- `kind`: `text|file|image|audio|link|derived`
- `title`
- `source_actor`
- `status`: `received|processed|failed|superseded`
- `preview`
- `origin_ref`
- `derived_artifact_ids`
- `created_at`
- `updated_at`

Examples:

- `meeting-notes.txt`
- `budget.png`
- `requirements.pdf`
- `voice-note.m4a`
- `hotel-shortlist.xlsx`

Rules:

- every uploaded item must become visible immediately
- Jini must report what it extracted from each processed input
- failed parsing/transcription must be visible, not silent

## ArtifactCard

`ArtifactCard` is the user-facing representation of an artifact.

It may wrap any canonical artifact from [artifact-schemas.md](./artifact-schemas.md).

Required fields:

- `artifact_id`
- `thread_id`
- `artifact_type`
- `title`
- `status`: `draft|ready|needs_input|blocked|approved|archived`
- `summary`
- `preview`
- `open_action`
- `export_actions`
- `source_input_ids`
- `updated_at`

User-facing status mapping:

- `draft`: useful but still provisional
- `ready`: usable now
- `needs input`: blocked on user answer
- `blocked`: blocked on external dependency or missing proof

Rules:

- artifact cards must be openable inline
- users must see artifact title and status without opening it
- artifact cards must be grouped under `Ready now`, `Needs input`, or `Blocked`

## Ask

`Ask` is a bounded request for user clarification.

Required fields:

- `ask_id`
- `thread_id`
- `prompt`
- `reason`
- `required`
- `options`
- `assumptions_if_skipped`
- `blocking`
- `created_at`
- `resolved_at`

Rules:

- one active blocking ask at a time
- multiple low-impact asks must be bundled into one compact form
- every ask must say why the answer matters
- if the user skips the ask, Jini must say what assumption or draft limit will remain

## ProgressSnapshot

`ProgressSnapshot` is the stable state header for the current thread.

Required fields:

- `goal`
- `working_with_summary`
- `now`
- `done`
- `need`
- `next`
- `safe_to_do`

This object is the source for the persistent shell frame.

## Conversation Structure

Jini should structure the conversation in three layers:

1. `Thread`
   The durable work context.
2. `Turn`
   What changed this exchange.
3. `Artifact`
   The durable outputs the user can open later.

The transcript itself is not the product.

The transcript supports:

- explanation
- progress narration
- assumptions
- warnings
- next-step guidance

## Screen Shapes

## 1. Empty Thread / First Run

```text
What do you need help finishing?

Jini shell
Paste notes or type what you want finished.

Working with
Local preview

If you need setup help, type `Use Auto`.
If you are not sure, type `help me finish this`.

Good inputs:
- Turn meeting notes into something I can send
- Check whether a plan is ready to hand off
- Plan
- I am not sure
```

Rules:

- one obvious first move: paste work
- setup help is secondary
- examples are hints, not a forced menu

## 2. Preflight Decision Card

Shown only before new work begins.

```text
Jini will start with

Tool
ChatGPT

Provider
Azure OpenAI / gpt-4o-prod

Route policy
Automatic

Model
gpt-4o-prod

Effort level
Medium

Why this was chosen
Auto mode prefers the cheapest suitable planning tool by default.

Change route
Type `Use Claude Code`, `Use Bedrock Sonnet`, `Use ChatGPT`, `Use Codex`, `Use Azure OpenAI`, or `Use Auto`.
```

Rules:

- this card appears before the first draft
- this card explains route trust, not provider plumbing
- this card should be concise enough to skim

## 3. First Result Screen

```text
Your first draft is ready.

Trip Plan

[artifact preview]

Actions
- Continue
- Show missing
- Plan
- Start new
```

Rules:

- useful result appears before summary
- no output-size question before value
- follow-on actions must all be real

## 4. Current Thread Screen

```text
Goal
7 day Paris trip for a couple under $2500

Working with
Trip notes, hotel screenshot, budget voice note

Now
Comparing trip shape and refining itinerary

Done
- Dates inferred
- Budget draft created
- Day-by-day itinerary drafted

Need
Confirm museum-heavy, food-heavy, or mixed

Next
Finalize itinerary and booking checklist

Ready now
- Trip Plan
- Budget Sketch
- Booking Checklist

Blocked
- Exact hotel area
```

Rules:

- this is the default return view
- it must feel like a calm recap, not telemetry
- users should know what happened without reading chat history

## 5. Multi-Project Home

```text
Your work

1. Paris trip
   Next: confirm trip style

2. Vendor review
   Next: choose one option

3. Q3 memo
   Next: review final draft

Type a number to open one, or type `Start`.
```

Rules:

- if multiple threads exist, Jini shows the thread chooser first
- the chooser uses user titles and next-step summaries
- no pack ids, no file paths

## 6. Attachment Strip

Every thread should show a compact input strip under `Working with`.

Example:

```text
Working with
- Notes.txt
- Hotel options.png
- Budget voice note.m4a
- Flight shortlist link
```

When processed, Jini should add extraction notes:

- `Budget voice note.m4a -> transcribed`
- `Hotel options.png -> prices extracted`

## 7. Artifact Shelf

Artifact shelf groups:

- `Ready now`
- `Needs input`
- `Blocked`
- `Sent / shared`

Each card shows:

- title
- type
- status
- last updated
- open action

## Input Handling

## Text

Plain pasted text is the default input path.

Rules:

- text should work without setup knowledge
- Jini should classify probable work type from the text

## Files

Supported intent:

- document review
- spreadsheet extraction
- image reading
- note ingestion

Rules:

- file name must be shown back to the user
- parsed extraction summary must be shown
- unreadable file state must be visible

## Images

Supported intent:

- screenshot analysis
- whiteboard/photo transcription
- form extraction
- visual comparison

Rules:

- Jini must say what it observed
- do not silently convert image input into hidden text

## Audio

Supported intent:

- meeting note transcription
- voice memo capture
- spoken plan extraction

Rules:

- Jini must show transcription status
- user must be able to open transcript artifact
- if confidence is low, Jini must say so

## Links

Links become `InputItem(kind=link)`.

Rules:

- show source domain
- show whether content was fetched or not
- if a link cannot be read, keep the link visible and mark it unresolved

## Clarification Rules

Jini should converge on missing information like this:

1. classify what kind of work this is
2. create the first useful artifact
3. expose the highest-impact missing input
4. ask one bounded question
5. continue the artifact, not the chat

Bad:

- five separate questions before any output

Good:

```text
Need
Choose one trip style

Options
1. Classic sights
2. Food and neighborhoods
3. Mixed pace

If you skip this
Jini will assume mixed pace.
```

## Artifact Sharing Back To The User

Artifacts returned by Jini should support:

- open
- compare
- revise
- export
- mark ready

Export formats depend on artifact type:

- markdown
- copyable plain text
- structured JSON
- document export when relevant

Rules:

- artifact title must be human-readable
- exported shape must match artifact purpose

## Safety And Trust

Jini must explicitly show:

- assumptions
- unresolved blockers
- uncertainty
- whether the output is still safe to review before sending

The product must not present a draft as final when material gaps remain.

## Mapping To Existing Protocol

The user-facing thread model maps to protocol objects like this:

- `WorkThread -> WorkUnit`
- `ArtifactCard -> Artifact`
- `TurnRecord -> Event bundle over Operation + StateTransition`
- `Ask -> unresolved Requirement/Constraint/Assumption gap`
- `ProgressSnapshot -> runtime projection over WorkUnit + artifacts + evidence`

This means Jini does not need a second product database conceptually.
It needs a UI/runtime projection over the canonical protocol.

## Implementation Order

### Phase 1

- add `WorkThread` runtime projection
- add visible `InputItem` strip
- add `ArtifactCard` shelf
- keep persistent `Goal / Working with / Now / Done / Need / Next`

### Phase 2

- support file, image, audio, and link intake in one thread
- add extraction status and transcript/provenance artifacts
- add bounded `Ask` handling

### Phase 3

- add artifact compare/revise flows
- add sent/shared state
- add portfolio-level thread home

## Non-Goals

This spec does not define:

- exact provider APIs
- storage adapter mechanics
- full document rendering/export stack
- permission model for external publication

Those belong in execution, memory, and publication specs.
