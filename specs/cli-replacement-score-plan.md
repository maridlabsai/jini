---
title: CLI replacement score plan
updated_at: "2026-05-27"
status: active
---

# CLI Replacement Score Plan

This document is intentionally blunt.

Jini does not need to make users abandon Claude Code or Codex as engines.
It needs to make them start in Jini because Jini is the better operating
surface for finishing, resuming, routing, proving, and handing off work.

That means the target is not:

- "be a better foundation model shell"

The target is:

- "be the better default shell for real work that has to survive the first draft"

Today Jini has a real product idea, but it is not yet replacement-grade at the
CLI surface. The score floor says `8.92`. If every replacement-critical score
must rise above `9.0`, the next work has to be subtractive, public, and
behavioral.

## Current blunt assessment

### What is already strong

- The product direction is coherent: Jini is for the part after the first draft
  when users need state, proof, handoff, continuity, and artifacts.
- The shell framing is materially better than raw chat when it works:
  `Goal`, `Working with`, `Doing now`, `Up next`, `Ready now`, `Blocked`,
  `Need`, and `Why this matters` are the right surfaces.
- The routing story is directionally correct: cheap adequate path first,
  stronger route only when rigor or policy requires it.
- The paid story is narrow enough to be credible: paid only when savings or
  continuity can be measured.

### What is still below replacement grade

- The public command contract is not trustworthy enough.
- The installed artifact can drift away from the public docs and still ship.
- Too much of the value story is explained in docs instead of proven in the
  live binary.
- Jini still reads like a product vision in places where it should read like a
  stable operator tool.

## Replacement-critical evidence

The current public materials teach a command surface that the installed CLI does
not reliably honor.

Examples from the repo:

- `README.md` teaches `jini commands`, `jini admin help`, and `jini doctor`.
- `docs/cli.md` teaches `jini commands`, `jini admin help`, `jini doctor`,
  `jini status`, and `jini metrics`.
- The installed binary in local use accepted `jini check` and
  `jini provider doctor`, but rejected:
  - `jini commands`
  - `jini admin help`
  - `jini doctor`
  - `jini status`
  - `jini --help`

This is not a cosmetic mismatch.

For a CLI, command-contract drift is a trust failure. A user who hits
"docs said X, binary says unknown command" will not grant Jini replacement
status over Claude Code or Codex. They will demote it to "interesting wrapper."

## What Jini is and is not

### What Jini should be

- The shell you start in when the work has to persist.
- The surface that chooses the right engine and makes that choice visible.
- The shell that leaves behind sendable, reviewable, reusable work.
- The shell that makes resume, handoff, and proof legible.

### What Jini should not be

- A bigger command tree.
- A thinner clone of Claude Code.
- A wrapper that adds explanation but not operational gain.
- A product that asks the user to memorize a public surface that the binary
  itself does not honor.

## The score problem

The current tracked sub-9 dimensions already point at the right work:

- `delivery-maturity`: `8.9`
- `memory-reliability`: `8.9`
- `adapter-portability`: `8.9`
- `token-efficiency`: `8.8`

Those scores should not be raised by argument. They should rise only when the
CLI behaves like a calmer and more dependable front door than adjacent tools.

## What must change to get every score above 9.0

## P0: Command truth before feature growth

If the public docs teach a command, the install artifact must support it.

Mandatory moves:

- Ship one replacement-safe public command contract and freeze it:
  - `jini`
  - `jini help`
  - `jini check`
  - `jini open`
  - `jini provider doctor`
- Either:
  - add compatibility aliases for taught commands like `jini commands`,
    `jini status`, `jini doctor`, and `jini --help`
  - or remove those teachings from the public docs until the live binary
    actually supports them
- Generate the public command catalog from the binary or from a single checked
  source of truth. Do not hand-maintain the docs and parser separately.
- Add release smoke tests against the install artifact, not just the repo code.

Score effect:

- delivery maturity up
- packaging/install trust up
- product trust up

Without this, no other score increase matters.

## P0: Replacement flows must beat raw shells on one screen

Jini will not win by being generally nicer. It will win only if two or three
flagship flows are obviously better than starting from Claude Code or Codex.

The required flagship flows are:

- After a meeting -> sendable follow-up with owners and open questions
- Before handoff -> build-readiness or plan-readiness check
- Before decision -> recommendation plus tradeoff summary

For each flagship flow, the first screen after generation must make four things
obvious:

- what is usable now
- what is still blocked
- what artifact should be opened first
- what the next safe action is

If that loop is not visibly better than "paste prompt into model shell and copy
the result," Jini does not earn a replacement score above 9.

## P1: Help and routing must be product surfaces, not fallback surfaces

`--help` failure is unacceptable for a CLI that wants replacement status.

Mandatory moves:

- `jini --help` and `jini -h` must always work
- `jini help` must always print the same public contract
- route inspection should be visible in a public-safe form
- provider health should never require hidden operator knowledge

The right posture is:

- public help is calm and bounded
- operator depth exists, but does not contaminate the beginner path

## P1: Docs must become executable promises

Every public promise should be tied to one of:

- CLI smoke test
- docs contract test
- release artifact test
- golden benchmark scenario

New rule:

- no public page may teach a command that is not exercised against the built
  install artifact in CI

New rule:

- no homepage claim about route, proof, or resume may survive without a
  benchmark scenario that exercises it

This is how trust gets above 9.0: visible proof, not more copy.

## P1: Position Jini as the front door, not the engine

Jini does not need users to "leave Claude Code."
It needs users to stop starting there by default.

The winning framing is:

- start in Jini
- let Jini choose the cheapest adequate route
- escalate to Claude Code, Codex, Bedrock, or another stronger engine only when
  the work justifies it
- keep the artifact, proof, and continuity layer in Jini

If Jini keeps presenting itself as a direct model-shell replacement, it will be
judged on the wrong axis and lose.

If it presents itself as the better operator shell for durable work, it has a
credible lane.

## P1: Release/install drift must become a tracked blocker

Today the repo and the installed binary can tell different stories.

That should be treated as a release-blocking defect class.

Required controls:

- compare docs-taught commands against the installed artifact
- compare release README snippets against the installed artifact
- fail publish-readiness when the built binary and public docs disagree

This is not optional polish. It is replacement trust.

## P2: Memory and adapter scores only rise if they change behavior at the right step

Sub-9 scores in memory and adapters should not be fixed with more machinery.
They should be fixed only when they improve the default loop.

Memory must:

- resurface the right prior decision at the step where it changes action
- avoid forcing the user to restate recent context on normal resumptions
- influence the next recommendation or next artifact, not merely exist

Adapters must:

- make one live issue/wiki/runtime edge feel native
- preserve the same semantics across edges
- avoid making the user relearn the system per integration

If they remain architecture-heavy and behavior-light, they should stay below 9.

## P2: Token efficiency should become visible and user-legible

Jini's routing story is good on paper.
It needs to become obvious at runtime.

To exceed 9.0:

- the cheap path must be the visible default
- expensive escalation must show a short reason
- resumptions must visibly avoid full reloads
- cost reduction must come from architecture, not from weaker output quality

The user should be able to answer:

- why this route
- why this effort level
- why not a cheaper route

without reading deep docs.

## Score-raising order

1. Fix command truth and release drift.
2. Lock two flagship replacement flows until they are obviously better than raw
   shells.
3. Make help, doctor, and routing public-safe and stable.
4. Tie docs and homepage claims to executable tests.
5. Only then spend more energy on memory, adapters, and deeper automation.

## Definition of above-9 replacement quality

Jini can claim every replacement-critical score is above `9.0` only when all of
these are true at the same time:

- a new user can install and run the public docs commands without drift
- the default `jini` path produces a better durable work loop than raw model
  shells for at least two flagship workflows
- route choice, proof, and blocked state are visible before the user has to
  trust them
- the installed artifact, public docs, and benchmark scenarios all agree on the
  product surface
- escalating to Claude Code or Codex feels like Jini doing the right thing, not
  Jini failing to be enough

## Immediate next build slices

### Slice 1

Make the command contract true.

- support the taught public aliases or remove them from public docs
- make `--help` work
- add install-artifact smoke tests for every taught public command

### Slice 2

Choose two replacement flows and make them unbeatable.

- meeting notes -> follow-up
- PRD/plan -> readiness check

### Slice 3

Generate public command docs from the actual binary contract.

### Slice 4

Bind homepage and proof claims to executable benchmark scenarios.

Until those four slices land, the right product verdict is:

- strong idea
- real value
- not yet a >9.0 replacement-grade CLI
