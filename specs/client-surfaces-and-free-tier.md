# Client Surfaces And Free Tier

Updated: 2026-05-19

## Purpose

This document defines the public product requirements for Jini across:

- CLI
- desktop app
- mobile app

It also defines the public free-tier rule and the public boundary between
free-core product access and commercial hosted value.

This is a public product contract, not a pricing memo.

## Product Rule

Jini should behave like one product with multiple surfaces, not three
separately priced products.

The product should have:

- one account model
- one work thread model
- one route and artifact model
- one continuity model across surfaces

## Shipping Status And Public Commitment

Jini needs a clear distinction between what is shipping now and what is a
public product commitment.

### Shipping Now

- CLI is the live primary surface.
- public install, proof, and routing transparency are already part of the
  current product story.

### Public Product Commitment

- desktop support for macOS and Windows is part of the public product
  direction, but it should not be described as generally available until the
  real surface exists and passes the same work-thread and trust checks as CLI.
- mobile continuation/review is also a public commitment, not a claim that the
  full mobile surface is already shipped.

Public docs should never blur the difference between:

- available now
- publicly committed next
- speculative later

## Surface Requirements

### CLI

The CLI is the primary power-user surface and the most important proof surface.

Public requirements:

- must remain installable without paid access
- must support local SLM use
- must support BYO provider/API configuration
- must support public flagship flows
- must remain useful on the free tier

### Desktop App

The desktop app is the primary managed day-to-day surface.

Public requirements:

- must exist for macOS and Windows
- must support the same work-thread model as the CLI
- must expose artifacts, missing state, and next-step state clearly
- must remain usable on the free tier
- must preserve local-first and BYO flows where the platform allows them

### Mobile App

The mobile app is a continuation, review, and triage surface rather than the
primary creation surface.

Public requirements:

- must let users review work, continue lightweight work, and handle follow-up
- must keep the same work-thread identity as CLI and desktop
- must preserve visibility of route, artifacts, blockers, and next step
- must support the free tier

## Platform Coverage

### macOS

Public requirement:

- desktop support is required

### Windows

Public requirement:

- desktop support is required

### Mobile Platforms

Public requirement:

- at least one mobile surface is required as the continuation/review client
- exact platform order can be staged separately from this public contract

## Free Tier Rule

The free tier must be permanently useful.

That means:

- CLI access is free
- desktop access is free
- mobile access is free
- local SLM usage is free
- BYO provider/key usage is free
- the public core workflows remain usable without subscription

The free tier must not be a fake evaluation shell that becomes useless without
hosted spend.

## Open Version Value Proposition

The open version should be valuable even if a user never pays Jini.

The open version should give users:

- one stable shell across supported surfaces
- local SLM and BYO control
- structural token savings that work across providers and tools
- artifact-first continuation
- compact resume and status views
- checkpoints, projections, and visible route proof

The open version should adopt the common efficiency patterns that serious tools
already converge on:

- context compaction
- lazy context loading
- checkpointed resume
- artifact reuse instead of transcript replay
- subtask isolation where the runtime can support it

These are table stakes for a trustworthy open shell and should not be treated
as paid lock-in.

## Public Upgrade Logic

The public repo should explain the upgrade ladder in product terms without
turning into a pricing page.

The stable public story is:

- free gives real local and BYO use plus structural savings
- hosted proof-of-value may exist without making free deceptive
- paid product adds adaptive optimization, limit avoidance, auto recovery, and
  stronger continuity convenience
- paid services and support are separate from paid product capability

That adaptive optimization must include:

- changing the underlying platform or runtime target when throttle pressure,
  quota pressure, or availability drift makes another route better
- changing the model or local profile for the task based on task shape,
  modality, and complexity instead of assuming one fixed model is right for all
  work

## Upgrade Trigger

The user should upgrade only when the open version is already useful and the
commercial layer can clearly save more money or interruption than it costs.

In plain language:

- use open Jini when local, BYO, and structural efficiency are enough
- pay for Jini when provider limits, throttles, and route decisions are the
  expensive part
- do not pay just to unlock basic session continuity or basic shell use

The product should never depend on a weak free tier to force conversion.

## Commercial Boundary

The commercial layer may monetize:

- provider-specific optimization
- learned routing and compression policy
- subscription-limit forecasting
- preemptive throttle avoidance
- automatic fallback and resume
- automatic platform switching across managed routes when throttling or quotas
  make the current route the wrong choice
- task-shaped model and profile selection across local and managed routes
- sync and continuity convenience
- shared team workflows
- governance and enterprise controls

The public layer must continue to define:

- free access to the surfaces
- free local/BYO operation
- free structural efficiency patterns
- public install and proof path

## Continuity Rule

The user should not need a different mental model per surface.

A user should be able to:

- start on CLI
- continue on desktop
- review on mobile
- return to desktop or CLI

without losing the thread, artifact identity, or routing transparency.

## Trust, Identity, And Export Boundary

If Jini offers signed-in continuity across surfaces, it must preserve:

- visible route and verification transparency
- exportability of user work
- clear separation between local/BYO behavior and hosted behavior
- explicit privacy boundaries for cross-surface identity and sync

Commercial convenience may extend continuity. It may not quietly erase:

- what is running locally
- what is running through Jini-hosted infrastructure
- what the user can export or take with them

## Public Non-Negotiables

1. Do not price CLI, desktop, and mobile as separate products.
2. Do not paywall local-first usage.
3. Do not make the free tier too weak to build trust.
4. Do not split the product identity across different user models by surface.
5. Do not turn the public repo into a pricing or GTM surface.

## Relationship To Commercial Repo

The commercial repo may define:

- SKUs
- hosted usage budgets
- subscription enrollment
- billing operations
- financial viability rules

Those details do not belong here.

This public document only fixes the product requirements and free-tier
boundary those commercial systems must respect.
