# Delight Gap Closure

Updated: 2026-05-19

## Purpose

This document selects the smallest set of user-delighting features that can
materially narrow the adoption gap between Jini and stronger day-to-day
competitors without growing Jini's kernel surface.

The standard here is not novelty. The standard is felt user value:

- less friction
- better continuity
- richer artifacts
- fewer accidental context mistakes

## Official Source Set

- Claude Code overview: https://docs.anthropic.com/en/docs/claude-code/overview
- Claude artifacts: https://support.anthropic.com/en/articles/9487310-what-are-artifacts-and-how-do-i-use-them
- ChatGPT Projects: https://help.openai.com/en/articles/10169521-using-projects-in-chatgpt
- ChatGPT Canvas: https://help.openai.com/en/articles/9930697-what-is-the-canvas-feature-in-chatgpt-and-how-do-i-use-it
- OpenAI Codex docs: https://platform.openai.com/docs/codex
- Work with Codex from anywhere: https://openai.com/index/work-with-codex-from-anywhere/

## Selection Rule

Only ship delight features that meet all three tests:

1. the user feels the value in the same session
2. the feature reduces friction or confusion instead of adding setup
3. the behavior generalizes across work types instead of depending on
   use-case-specific hard coding

## Critique Rule

Delight work does not get special protection just because it feels polished.

If two or more critique sources independently say the same delight feature:

- adds cognitive load
- teaches too much
- duplicates an easier action
- creates anxiety about making a mistake
- adds a branch the user does not actually need

then the default action is to remove, demote, hide, or collapse it.

Do not keep a criticized delight feature only because:

- it is already implemented
- it demonstrates technical cleverness
- it looks richer in a screenshot
- it helps explain an internal architecture

Keeping the feature requires explicit proof that it improves user outcome
through speed, trust, clarity, quality, or cost posture.

## Selected Features

### 1. Interruption-Safe New Work

#### Competitor pattern

- Codex and ChatGPT keep continuity strong by preserving active work state
  across surfaces instead of silently discarding context.
- Claude Code keeps the working environment legible and does not require the
  user to rediscover what was active.

#### Jini gap

When current work exists, pasting a fresh request can silently change focus.
That is fast, but it is not trustworthy.

#### UX design

If the user enters what looks like a genuine new task while current work is
active, Jini should pause and show a compact interrupt card:

- current work title
- incoming request preview
- `Start new work`
- `Keep current work`
- `Switch project`

The old work stays saved unless the user explicitly switches.

#### Technical design

- intercept fallback freeform input in current-work mode before a new work unit
  is created
- preserve the raw request so `Start new work` can continue without retyping
- route `Switch project` through the existing active-work picker
- keep the interrupt card generic and independent of pack-specific behavior

#### Pass criteria

- current work is never switched silently
- a pasted new request can become new work without retyping
- the interrupt flow remains one short decision, not a wizard

### 2. Interactive Artifact Shelf

#### Competitor pattern

- Claude artifacts and ChatGPT canvas turn large outputs into openable surfaces,
  not passive transcript text.
- Codex keeps rich work objects inspectable across app, web, mobile, and local
  task contexts.

#### Jini gap

Jini can list ready artifacts, but the shelf is still too passive. Users see
names, then have to infer the next command instead of simply opening the thing
they want.

#### UX design

When the user chooses `Show what's ready` or `Open what's ready`, Jini should
show a numbered shelf and let the next input open an item directly by:

- number
- title
- alias

The shelf should work for:

- ready artifacts
- send/share exports
- additional detail artifacts

#### Technical design

- build one ordered open-shelf list from views, exports, and details
- render numeric affordances without exposing internal storage paths
- allow follow-up selection by number or label
- record the open as an artifact observation just like scriptable `jini open`

#### Pass criteria

- the shelf is immediately actionable
- typed numbers do not collide with top-level actions because selection happens
  inside the shelf flow
- artifact opens keep telemetry and passive outcome tracking intact

### 3. Fuller Next Pass

#### Competitor pattern

- ChatGPT canvas and Claude artifacts make iterative expansion feel natural.
- Codex continuation flows encourage moving from the first useful object to the
  next useful surface without restating the task.

#### Jini gap

After the first useful artifact, Jini can still feel binary: either keep going
or inspect missing state. There is not enough support for the common user ask
"show me the fuller version."

#### UX design

Jini should accept natural expansion requests such as:

- `Make it fuller`
- `Show more`
- `Expand this`

The action should open the best richer companion surface already available:

- next ready artifact
- supporting detail artifact
- export/detail fallback when no richer ready artifact exists

#### Technical design

- add a generic richer-surface selector over the current work summary
- keep the action read-only and fast; this is a navigation affordance, not a
  second generation pass
- support the action in both post-result and current-work flows

#### Pass criteria

- expansion opens a richer artifact when one exists
- the action does not create new work
- the action uses generic artifact ordering, not use-case branches

## Rejected Candidates

These are intentionally not in this slice:

- broader route/model explanations by default
- more starter profiles
- more inline examples
- new domain-specific canned workflows
- richer surface claims without a real artifact interaction path

They either add teaching overhead or solve a weaker problem than the three
selected features above.

## Delivery Order

1. interruption-safe new work
2. interactive artifact shelf
3. fuller next pass

This order protects trust first, then makes artifacts feel more native.

## Next Delight Slice

The next highest-leverage delight work is not more startup scaffolding. It is
making the artifact feel more inspectable, more revisable, and safer to revise.

### 1. Inspectable Context Capsule

#### Competitor pattern

- ChatGPT Projects and memory make stored context and project-scoped sources a
  user-facing concept instead of hidden prompt state.
- Claude Code makes memory files and project context inspectable.
- Codex continuation emphasizes visible task context, terminal state, and
  approval context across surfaces.

#### UX design

Jini should let the user ask:

- `Show what Jini used`
- `Show context`
- `What did you use`

The answer should stay compact and show:

- direct user inputs
- clarifications or attachments
- links or source references when present
- what Jini intentionally kept visible as missing or uncertain
- route and continuity context when it materially shaped the work

#### Technical design

- use existing `inputItems`, `thread-state`, route fields, and source-link
  extraction
- avoid pack-specific rendering; the capsule is generic across work types
- keep it read-only and available both immediately after first draft and later
  during current-work resume

#### Pass criteria

- users can inspect what shaped a draft without opening JSON files
- the capsule explains context without dumping internal implementation details
- the same command works for text, clarified scope, and file-backed inputs

### 2. Quick Artifact Rewrite Shortcuts

#### Competitor pattern

- ChatGPT Canvas supports fast editing and rewrite-style iteration.
- Claude Artifacts encourages iterative reshaping of the same work object.
- Codex continuation flows keep users moving from draft to improved draft
  without rebuilding context from scratch.

#### UX design

Jini should support natural rewrite shortcuts on the current artifact:

- `Make it shorter`
- `Make it executive`
- `Turn this into a checklist`

These should feel like immediate refinements of the current artifact, not new
work records or provider setup flows.

#### Technical design

- treat the current primary artifact as the editable object
- apply generic markdown transforms over headings, bullets, gaps, and next-step
  state instead of pack-specific rewrite templates
- keep transforms deterministic so they work even when no external provider is
  configured

#### Pass criteria

- rewrite shortcuts work on any ready markdown artifact
- the output becomes immediately usable in the same session
- the transform does not create a second parallel work unit

### 3. Revision History And Undo

#### Competitor pattern

- ChatGPT Canvas exposes revision-oriented editing and restore flows.
- Claude Artifacts supports iterative updates and encourages safe reuse instead
  of one-shot transcript text.
- Rich assistant surfaces generally make users comfortable revising because the
  previous state is not lost.

#### UX design

After Jini rewrites an artifact, the user should be able to say:

- `Show versions`
- `Undo last change`

This keeps experimentation safe and lowers the fear of asking for a rewrite.

#### Technical design

- save a lightweight snapshot before every in-place artifact rewrite
- keep a small per-work ledger for artifact snapshots
- restore the latest snapshot in one step without requiring file paths or git

#### Pass criteria

- every shortcut rewrite creates a restorable version first
- `Show versions` is human-readable, not file-system-oriented
- `Undo last change` restores the prior artifact content in one step

## Why These Three

These features were chosen because they combine directly:

- trust, by making context inspectable
- delight, by making artifact refinement immediate
- safety, by making revision reversible

Together they move Jini closer to competitor strengths without copying their
entire product surfaces or adding more startup ceremony.
