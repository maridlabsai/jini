# Product Review Roles

Updated: 2026-05-14

## Purpose

Jini product changes must be reviewed through four roles before they are treated
as ready for implementation or public positioning.

The goal is consensus on the user experience, not agreement that the design is
architecturally tidy.

## Roles

### Competitive Analyst

The competitive analyst judges whether the proposed experience can beat the
best recent tools at the stage where they are strongest.

Required comparisons:

- Claude Code: first payoff and terminal immediacy
- Kiro: visible progression and actionable work artifacts
- Hermes: continuity, memory, and return-to-work feel
- AgentField: inspectability, provenance, and operational trust
- AI Hero-style intake: natural capture without forcing taxonomy too early

Approval condition:

- Jini must be a complete, visible, trustworthy work loop.
- Jini must not be merely a better launcher for a rigorous internal system.

### UX Researcher

The UX researcher represents tired, unsure, low-confidence users.

They reject designs that make users:

- classify their problem before they understand it
- compress messy context too early
- interpret system state before seeing value
- trust remembered work without visible control
- learn internal vocabulary before getting help

Approval condition:

- the first minute feels like relief
- the user sees value before structure
- the user always knows what is safe, what is missing, and what to do next

### UX Designer

The UX designer owns the screen order, copy, and information hierarchy.

They reject designs that:

- show summaries before the useful result
- show choices that do not work
- expose implementation storage or file language
- use generic output names when a human object name is clearer
- add a separate explanation screen where inline help would do

Approval condition:

- the screen flow is one obvious path
- the first result is centered
- follow-up actions are few, real, and written in user language

### Program Manager

The program manager protects scope, sequence, and scorecard lead.

They reject designs that:

- expand beyond the flagship flows before parity is proven
- create a second generation path without fixtures
- ship remembered-work behavior without park/switch/stale recovery rules
- rely on benchmark uplift without compatibility proof
- damage workflow rigor, governance, learning maturity, or core simplicity

Approval condition:

- the replacement slice is narrow enough to ship safely
- parity tests exist for new and old work
- score momentum improves without weakening the guarded lead

## Consensus Gate

A product change is not approved until all four roles accept these conditions:

- first useful result appears before the summary
- first-run choices are limited to real shipped paths
- natural paste-first intake is supported before strict taxonomy
- current work is visible, controllable, recoverable, and never switched silently
- source, assumptions, missing proof, and safe next action are visible
- launcher-created work is parity-tested or uses shared generation
- meeting follow-up and plan readiness are bulletproof before broader launcher scope

