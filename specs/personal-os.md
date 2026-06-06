# Personal OS Home

## Purpose

Jini now has a small personal-OS surface for durable memory, tool inventory,
and repeatable routines.

This layer is meant to make the system easier to start, resume, and operate
without promoting new concepts into the kernel.

The rule is:

- canonical pack state remains the source of truth for governed workflows
- personal memory supports recall, setup, and continuity
- routines package repeated behavior without mutating protocol semantics

## Home Layout

A bootstrapped home currently materializes:

- `home.yaml`
- `soul.md`
- `user.md`
- `tools.md`
- `memory/daily/`
- `memory/long-term.md`
- `routines/local/`
- `routines/remote/`
- `outputs/briefs/`
- `outputs/benchmarks/`
- `outputs/reviews/`
- `outputs/release/`
- `runtime/remote-runs/`

## Files

### `home.yaml`

Machine-readable manifest for the home.

It records:

- owner and assistant identity
- memory paths
- tool inventory
- routine directories
- update timestamps

### `soul.md`

Human-readable tone and operating style for the assistant.

This is a personalization surface, not a protocol override.

### `user.md`

Human-readable notes about the operator.

It is for durable preference and working-context hints, not canonical workflow
approval or evidence.

### `tools.md`

Human-readable inventory of available tools and registries.

This is intended to make the active surface legible before the user has to
inspect manifests or adapters directly.

## Memory Model

The current memory loop is intentionally simple.

### Daily Memory

`append-memory` writes one durable line into `memory/daily/YYYY-MM-DD.md`.

This is the cheap write path for facts worth keeping after a session or routine.

### Long-Term Memory

`dream-memory` compresses daily files into `memory/long-term.md`.

The current implementation:

- reads daily memory files
- strips formatting noise
- deduplicates lines
- writes a compact durable summary with provenance

### Truth Hierarchy

The hierarchy is:

1. canonical pack artifacts and runtime evidence
2. pack-local runtime logs and generated reports
3. long-term memory
4. daily memory

Long-term memory can support retrieval and resumption, but it should not
override canonical artifact state.

## Routine Model

Routines are split by execution locality.

### Local Routines

Local routines run on the operator machine.

Current built-ins include:

- `dream-memory`
- `daily-brief`
- `golden-benchmark`
- `framework-review`
- `publish-readiness`

Local routines can also be argv-backed. Raw shell-backed routines are an
explicit trusted-local escape hatch and should be treated as unsafe for shared
or untrusted home manifests.

### Remote Routines

Remote routines are represented explicitly even when live cloud execution is
not yet available.

Current remote execution is staged, not performed. `run-routine --mode remote`
creates an auditable receipt under `runtime/remote-runs/`.

That keeps the contract honest:

- remote intent is modeled now
- fake cloud execution is avoided
- receipts remain replayable and reviewable

## Native Go Surface

The current native Go surface for this layer is the shared product surface:

```bash
jini
jini status
jini continue
jini open
jini doctor
```

The old personal-OS verbs remain backlog concepts, not runnable public CLI:

- bootstrap home
- append memory
- dream memory
- list tools
- list routines
- run routine locally or remotely

If those capabilities return, they should graduate as native Go commands only
after the public command-surface and publish-readiness gates accept them.

## Why This Layer Exists

This slice improves:

- delivery maturity by reducing startup friction
- memory reliability by making durable notes and compression explicit
- advanced-set breadth by adding a reusable routine layer

It does not replace the roadmap items that still matter:

- automatic post-session memory writes from runtime flows
- richer retrieval into `recommend-execution` and `compact-context`
- real remote routine execution
- stronger adapter-backed tool inventory

The benchmark and framework-review routines are intentionally operational rather
than abstract. They keep competitor comparison and framework evolution on the
same lightweight home surface as memory and release routines, instead of
turning them into one-off CLI rituals.
