---
title: State And Artifacts
description: The shortest explanation of how Jini keeps interaction, state, and durable artifacts visible.
---

Jini should never leave you guessing about three things:

- what you can do now
- what state the work is in
- what artifacts back that state up

Those are not there for ceremony. They are there so the user can get to an
outcome instead of stopping at a draft, a recommendation, or a false green
light.

<div class="section-card">
  <h3>The three commands that matter most</h3>
  <div class="on-this-page">
    <a href="#status-pack"><code>jini status-pack</code>: see the truth of the work right now</a>
    <a href="#execution-checklist"><code>jini execution-checklist</code>: see the next honest step toward the outcome</a>
    <a href="#compact-context"><code>jini compact-context</code>: hand the important context to the next person or agent</a>
  </div>
</div>

## What Jini Always Shows

When Jini is doing its job, these surfaces stay visible:

<div class="example-snapshot">
  <div class="snapshot-card">
    <span class="snapshot-label">Interaction</span>
    <p>You always have a clear next command, not a vague suggestion.</p>
  </div>
  <div class="snapshot-card">
    <span class="snapshot-label">State</span>
    <p>You can see the current state, health, next move, and what is still missing.</p>
  </div>
  <div class="snapshot-card">
    <span class="snapshot-label">Artifacts</span>
    <p>You can see which durable artifacts support the work instead of trusting chat memory.</p>
  </div>
</div>

## `status-pack` {#status-pack}

Run:

```bash
jini status-pack /path/to/work
```

This is the everyday truth screen.

You should expect to see things like:

```text
STATE  awaiting_verification
HEALTH ready-to-verify
NEXT   Verify
MISSING-LATER
  - Approval
TASKS
  done:       3/3
  unresolved: 0/3
```

Read it like this:

- `STATE`: where the work really is now
- `HEALTH`: whether the current stage is healthy enough to advance toward an outcome
- `NEXT`: the next honest move
- `MISSING-LATER`: future blockers already visible now
- `TASKS`: whether the tasks are truly done or still unresolved

## `execution-checklist` {#execution-checklist}

Run:

```bash
jini execution-checklist /path/to/work --repo /path/to/repo --intent verify
```

Use this when you want the next step turned into an explicit checklist instead
of relying on memory or chat scrollback.

This is the bridge between state and action.

## `compact-context` {#compact-context}

Run:

```bash
jini compact-context /path/to/work --repo /path/to/repo --intent verify --max-chars 900
```

Use this when the next person or agent needs the essential context without a
full reload.

This is the bridge between artifacts and handoff.

## The Core Artifacts

Jini does not treat work as “whatever the last chat said.”

It keeps durable artifacts such as:

- `Brief`
- `Plan`
- `Tasks`
- `Evidence`
- `Approval`
- `Publication`

You do not need to memorize all of them on day one. The important point is
that state should be backed by artifacts, so the team can reach an outcome
without reconstructing the truth from memory.

## The Practical Rule

If someone asks:

- “What state is this in?”
- “What happens next?”
- “Why do we think this is ready?”

you should be able to answer from Jini directly, not by reconstructing the
story from messages, meetings, or memory.

## Go Next

<div class="section-card">
  <div class="on-this-page">
    <a href="./proof.md">See the proof screen</a>
    <a href="./examples.md">See example workflows</a>
    <a href="./cli.md">See the grouped CLI guide</a>
    <a href="./simple.md">Read the simple guide</a>
  </div>
</div>
