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
