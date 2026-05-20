# Product Consensus PRD And Plan

Updated: 2026-05-14

This is the short-form consensus version of the fuller product document in
[full-product-prd.md](./full-product-prd.md).

## Product Decision

Jini should be a complete, visible work loop.

The approved user experience is:

1. user gives messy context
2. Jini returns a useful object first
3. Jini shows what is missing, uncertain, and safe
4. user can keep going, inspect gaps, plan next work, inspect trust, or start new work

Jini is not approved as only a launcher into a work system.

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
What do you need help finishing?

Jini shell
Paste messy notes, or type the outcome you want.

Good inputs:
- Turn meeting notes into something I can send
- Check whether a plan is ready to hand off
- Help me plan this
- I am not sure

Nothing will be sent yet.
```

If the user types a clear job or pastes source directly, Jini should start from
that input without making the user choose a command.

If the user types `I am not sure`:

```text
Paste what you have. A rough version is fine.
I will help figure out whether this is follow-up, a plan check, or something else.
Nothing will be sent yet.
```

If the input does not clearly fit a flagship flow, Jini still returns a useful
first object:

- `First Useful Pass`

It must show:

- what the user appears to be trying to finish
- what can be used now
- what Jini needs next
- what is safe because nothing has been sent

Inline examples may appear under the prompt.

The first visible result must be:

- meeting: `Sendable Follow-up`
- plan/spec: `Build-Readiness Check`

The summary appears after the result or when the user asks what is still
missing.

## Post-Result Actions

After the first result, show only real actions:

- `Keep going`
- `Show what is missing`
- `Help me plan this`
- `Start something new`

Do not show actions that are not implemented.

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

## Implementation Plan

### Phase 1: Finish The Two Flagship Loops

- meeting follow-up creates a truly sendable result
- plan/spec readiness creates a clear build-readiness result
- first result appears before summary
- current-work choices are either real or hidden
- meeting follow-up and plan/spec readiness include real `Keep going` and `Show what is missing` actions
- launcher-created work uses shared generation or has golden parity fixtures before it ships or expands
- golden parity fixtures for new and old work exist before cutover

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
