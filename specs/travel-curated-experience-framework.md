# Travel Curated Experience Framework

Updated: 2026-05-19

## Purpose

This document defines the public product framework for Jini travel work.

Its job is to make trip planning feel curated, resumable, and trustworthy
without turning the public product into a booking engine before the
infrastructure exists.

## Why This Exists

Generic itinerary dumps are not competitive.

The public travel experience should feel closer to:

- Layla's curated leisure planning and trip shaping
- Navan Edge's preference-aware, confirmation-first execution posture

References:

- [Layla About](https://layla.ai/about)
- [Navan Edge](https://navan.com/product/navan-edge)
- [Introducing Navan Edge](https://navan.com/blog/introducing-navan-edge)

The goal is not to copy their surfaces literally. The goal is to absorb the
right product lessons and apply them to Jini's local-first, cross-surface,
work-thread model.

The strategic posture is:

- Layla-like curation for the public planning experience
- Navan-like confirmation discipline for any future execution layer

## Product Rule

Jini travel should act like one persistent travel work thread, not a sequence
of unrelated prompts.

That means the user should be able to:

- start a trip on CLI
- continue it on desktop
- review or unblock it on mobile
- return later without losing the brief, options, itinerary, blockers, or
  route history

## Benchmark Lessons

The most useful lessons from the market are:

1. do not start with an unscoped itinerary
2. turn preferences into a reusable trip brief
3. present curated options, not only prose
4. keep execution gated by explicit confirmation
5. remember traveler preferences across sessions

## Public Experience Layers

### 1. Scoped Travel Brief

Before a first draft, Jini should collect or infer the smallest set of
high-value constraints needed to avoid a generic result.

Typical dimensions:

- destination or route
- trip length
- travelers
- date window or season
- budget range
- pace
- hotel or stay preference
- must-dos
- constraints such as kids, accessibility, or limited walking

Jini should ask only for what is missing.

### 2. Curated Option Set

The first useful result should not be a monolith. It should usually include a
small option set such as:

- two or three trip styles
- two or three neighborhood or stay directions
- major anchor choices
- tradeoffs between pace, budget, and geography

This keeps the user in decision-making mode instead of forcing them to rewrite
a full itinerary immediately.

### 3. Day-By-Day Trip Object

Once scope is good enough, Jini should produce a trip artifact that includes:

- trip brief
- day-by-day plan
- likely budget shape
- logistics to lock
- still-to-confirm items
- contingency logic when weather, energy, or booking availability changes

### 4. Smart Reference Layer

Travel artifacts should include useful links or references on first mention of
major destinations, museums, transit entities, or booking-sensitive anchors.

The rule is:

- one helpful link beats many noisy links
- links should support action, not decorate text

### 5. Continuity And Resume

Travel work must keep:

- the current trip brief
- selected options
- itinerary version
- blockers
- next step
- what is already ready to book or review

This is especially important because trip planning naturally happens across
many sessions and devices.

### 6. Confirmation-First Trust

The travel experience should be explicit about the difference between:

- planning
- recommendation
- draft execution intent
- real booking or change

Jini should never blur those boundaries.

If later commercial integrations add live booking or changes, they must still
require explicit user confirmation before execution.

## Shipping Now Vs Later

### Shipping Now

The public travel commitment today is:

- scoped trip intake
- curated itinerary artifacts
- smart links and references
- resumable trip threads
- visible missing state and next steps

### Later, Not Yet Assumed Publicly Shipped

The public repo should not imply that Jini already ships:

- live flight or hotel pricing
- direct booking
- loyalty program syncing
- disruption rebooking
- human concierge

Those belong to a later execution layer.

## Shared Invariants

Travel work must preserve:

1. useful-result-first
2. one persistent work thread
3. explicit missing-state visibility
4. clear planning-versus-execution boundaries
5. exportable trip artifacts
6. local/BYO usefulness on the free tier

## What This Framework Rejects

- destination-specific hardcoding as the main architecture
- generic itinerary-first behavior for underspecified requests
- trip output that hides missing constraints
- booking-like language without explicit confirmation semantics
- travel work that forks into a separate continuity model from the rest of Jini

## Relationship To Commercial Repo

This public framework defines the curated travel product contract.

The commercial repo may build on it with:

- traveler profiles
- loyalty memory
- live pricing
- booking and confirmation systems
- disruption handling
- subscription or metered execution

It may not replace this framework with a shell-specific or transaction-first
product model.
