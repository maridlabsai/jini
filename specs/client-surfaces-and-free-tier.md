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

## Commercial Boundary

The commercial layer may monetize:

- hosted routing
- sync and continuity convenience
- background execution
- premium connectors
- shared team workflows
- governance and enterprise controls

The public layer must continue to define:

- free access to the surfaces
- free local/BYO operation
- public install and proof path

## Continuity Rule

The user should not need a different mental model per surface.

A user should be able to:

- start on CLI
- continue on desktop
- review on mobile
- return to desktop or CLI

without losing the thread, artifact identity, or routing transparency.

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
