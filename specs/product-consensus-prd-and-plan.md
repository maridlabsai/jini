# Product Consensus PRD And Plan

Updated: 2026-06-05

This is the short-form consensus version of the fuller product document in
[full-product-prd.md](./full-product-prd.md).

The canonical product and operating PRD now lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

This document remains useful as a short-form critique and consensus snapshot,
but it is not the top-precedence roadmap or operating charter.

If this document conflicts with the canonical PRD on tenets, priorities,
requirements, or roadmap, the canonical PRD wins.

Deeper follow-on strategy artifacts:

- [number-one-product-research.md](./number-one-product-research.md)
- [number-one-platform-prd.md](./number-one-platform-prd.md)
- [number-one-development-plan.md](./number-one-development-plan.md)

## Product Decision

Jini should be a complete, visible work loop.

The approved user experience is:

1. user gives messy context
2. Jini routes to the configured downstream CLI when that is the right path, or acts locally when safe
3. Jini shows what is missing, uncertain, and safe
4. user can keep going with familiar commands, inspect gaps, plan next work, inspect trust, or start new work

Jini is not approved as only a launcher into a work system.
Jini is also not approved to create a new conversation grammar when common
agent CLI behavior is enough.

## Critique Resolution Rule

Jini should act in favor of better outcomes for the user, not in favor of
preserving internal shapes.

If more than one critique source independently identifies the same friction,
confusion, or low-value surface:

- the default decision is to remove it
- or demote it behind explicit user intent
- or simplify it until the critique no longer holds

Keeping the criticized behavior requires explicit proof that it materially
improves user outcome through speed, trust, clarity, quality, or cost posture.
Implementation effort, prior investment, and architectural tidiness are not
sufficient reasons to keep it.

## Flagship Scope

Replacement-critical flows:

- meeting follow-up
- plan/spec readiness

Demo-only until parity is proven:

- trip planning
- vendor comparison
- incident cleanup
- general work

## First Minute

The first minute must feel like relief, not setup.

Required flow:

```text
Jini
Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.

Examples:
- Add a line to the matching .txt file in this folder
- Turn meeting notes into something I can send
- Check whether a plan is ready to hand off
- Plan a 7 day Paris trip for two adults in October
- Compare these vendors and recommend one

Jini will not send, book, commit, or run destructive changes without a visible step.
Type `help` for examples and commands.
```

If the user types a clear job or pastes source directly, Jini should start from
that input without making the user choose a command.

If the user types `I am not sure`:

```text
Describe the task in one sentence.
Jini can answer, edit files, route to a configured tool, or ask one clarification.
Nothing will be sent, booked, committed, or changed without a visible step.
```

If the input does not clearly fit a flagship flow, Jini still returns a useful
first object:

- `Task Snapshot`

It must show:

- what the user appears to be trying to finish
- what is usable now
- what Jini needs next
- what is safe because nothing has been sent

The first visible result must be:

- meeting: `Sendable Follow-up`
- plan/spec: `Build-Readiness Check`

The summary appears after the result or when the user asks what is still
missing.

## Post-Result Commands

After the first result, use familiar commands instead of a custom action
taxonomy:

- `jini continue`
- `jini open`
- `jini status`
- `jini start`

Do not show commands or actions that are not implemented.

## Competitive Requirements

Jini must adopt the strongest competitor behaviors as requirements.

### Claude Code

Immediate payoff:

- useful result before status
- no setup vocabulary
- no output-size gate before value

### Kiro

Visible progression:

- show what Jini used
- show what is ready
- show what is missing
- keep deeper task/proof detail available

### Hermes

Dependable continuity:

- current work is visible before action
- current work can be continued, opened, or parked
- stale remembered work is explained and recoverable

### AgentField

Inspectable trust:

- show source
- show assumptions
- show missing proof
- keep provenance openable on demand

### Jini

Honest closure:

- useful drafts are not treated as final
- missing truth stays visible
- approval and evidence gaps are simplified, not removed

## Product Review Consensus Gate

The product shape is approved only when all four roles accept it:

- Competitive Analyst: stage-by-stage competitor position is strong enough
- UX Researcher: first minute works for tired or unsure users
- UX Designer: screen order and copy are obvious
- Program Manager: scope, sequence, parity, and score lead are safe
- repeated critique convergence triggers removal, demotion, or simplification unless user-outcome proof says otherwise

## Implementation Plan

### Phase 1: Finish The Two Flagship Loops

- meeting follow-up creates a truly sendable result
- plan/spec readiness creates a clear build-readiness result
- first result appears before summary
- current-work choices are either real or hidden
- meeting follow-up and plan/spec readiness include real `Continue` and `Missing` actions
- launcher-created work uses shared generation or has golden parity fixtures before it ships or expands
- golden parity fixtures for new and old work exist before cutover
- repeated multi-role critique is resolved by simplification, not carried forward as known UX debt

### Phase 2: Add Parity And Current-Work Safety

- parity tests for launcher-created and canonical work
- stale and missing current-work handling
- explicit park/switch behavior
- no local starter writer can expand without parity evidence

Phase 2 can deepen parity and recovery behavior, but the Phase 1 replacement
slice must already prove that launcher-created flagship work does not drift
from canonical work.

### Phase 3: Add Visible Continuation

- continue current work
- make result fuller
- see missing proof and assumptions
- inspect provenance on demand

### Phase 4: Expand Demo Flows

- trip planning
- vendor comparison
- incident cleanup
- general work

Expansion only happens after the two flagship flows pass the consensus gate.
