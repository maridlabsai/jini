# Common Examples

These examples are the fastest way to judge whether Jini is useful for your
kind of work.

Use the same reading pattern each time:

1. run one command
2. look at `STATE`, `NEXT`, and `MISSING-LATER`
3. decide whether that missing truth would have mattered in your last real workflow

<div class="workflow-jump">
  <div class="workflow-grid">
    <a class="workflow-card" href="#meeting-followup">
      <span class="workflow-meta">Meeting Follow-up</span>
      <h3>See how Jini turns a meeting into explicit execution state.</h3>
      <p>Best for teams losing decisions, owners, and missing approvals in notes and chat.</p>
      <code>jini try-example meeting-followup</code>
    </a>
    <a class="workflow-card" href="#research-prd">
      <span class="workflow-meta">Research To PRD</span>
      <h3>See the difference between tasks being done and work being verified.</h3>
      <p>Best for product and engineering handoffs where the rationale thins out as the work moves forward.</p>
      <code>jini try-example research-prd</code>
    </a>
    <a class="workflow-card" href="#vendor-selection">
      <span class="workflow-meta">Vendor Selection</span>
      <h3>See how recommendations keep their tradeoffs and approval path attached.</h3>
      <p>Best for teams making expensive decisions that should survive beyond the meeting where they were made.</p>
      <code>jini try-example vendor-selection</code>
    </a>
    <a class="workflow-card" href="#incident-response">
      <span class="workflow-meta">Incident Response</span>
      <h3>See how Jini keeps proof, rollback, and closure visible after the firefight.</h3>
      <p>Best for operational work where “service is back” is not the same thing as “the work is actually closed.”</p>
      <code>jini try-example incident-response</code>
    </a>
  </div>
</div>

## 1. Meeting Follow-up {#meeting-followup}

**When this matters:** a weekly product, staff, or project meeting ends with
notes scattered across docs, chat, and memory.

**Run:**

```bash
jini try-example meeting-followup
```

![Jini meeting follow-up demo](./assets/examples/meeting-followup.gif)

**Look for these lines:**

```text
HEALTH ready-to-make
STATE  decided
NEXT   Make
MISSING-LATER
  - Approval
  - Evidence
TASKS
  done:       0/3
  unresolved: 3/3
```

**What Jini makes obvious:**

- the meeting exists, but the work is not ready
- approval and evidence are still missing
- the next person inherits state, not just notes

## 2. Research To PRD Handoff {#research-prd}

**When this matters:** research exists, the team agrees something should be
built, and the handoff is starting to look more finished than it really is.

**Run:**

```bash
jini try-example research-prd
```

![Jini research to PRD demo](./assets/examples/research-prd.gif)

**Look for these lines:**

```text
HEALTH ready-to-verify
STATE  awaiting_verification
NEXT   Verify
MISSING-LATER
  - Approval
TASKS
  done:       3/3
  unresolved: 0/3
EVIDENCE
  target: spec-research-prd-v1 r1
  claims: 3
  risks:  1
```

**What Jini makes obvious:**

- tasks can be done while the work is still waiting on verification
- approval is still missing even though the draft looks complete
- the handoff keeps its source trail attached to the work

## 3. Vendor Selection {#vendor-selection}

**When this matters:** several vendors look plausible and the team needs an
approval-ready recommendation instead of another meeting recap.

**Run:**

```bash
jini try-example vendor-selection
```

![Jini vendor selection demo](./assets/examples/vendor-selection.gif)

**Look for these lines:**

```text
HEALTH ready-to-make
STATE  decided
NEXT   Make
CTRL   Proof, Guard, Cost, Approval
MISSING-LATER
  - Approval
  - Evidence
```

**What Jini makes obvious:**

- the recommendation exists, but the proof trail is still incomplete
- approval is a visible part of the workflow, not a side conversation
- tradeoffs stay attached to the decision instead of disappearing into slides

## 4. Incident Response {#incident-response}

**When this matters:** the immediate outage is over, but the operational work
still needs rollback, proof, and honest closure.

**Run:**

```bash
jini try-example incident-response
```

![Jini incident response demo](./assets/examples/incident-response.gif)

**Look for these lines:**

```text
HEALTH ready-to-make
STATE  decided
NEXT   Make
PROF   Critical
CTRL   Proof, Guard, Rollback
MISSING-LATER
  - Approval
  - Evidence
```

**What Jini makes obvious:**

- recovery is not the same thing as closure
- rollback context stays visible while pressure is still high
- proof and approval are still part of the work even after the service is back

## Breadth, After The Core Story

Once those four examples make sense, the broader public packs are easier to
understand:

- travel planning
- budget planning
- compliance audit

Those are not the best first explanation of Jini. They are proof that the same
kernel can stretch beyond one narrow class of work.

## What To Look For

If you are deciding whether Jini is relevant, ask whether one of these failure
patterns is already painful in your environment:

- the draft exists, but no one trusts the current state
- tasks are marked done, but the work is not actually verified
- ownership is implied, but not explicit
- approvals happen in chat or meetings, not in the work trail
- the next person loses the rationale behind the current state

That is the common shape Jini is designed for.

## The Smallest Useful Habit

If you adopt only one Jini habit, make it this:

```bash
jini status-pack <path-to-work>
```

Use it:

- after a meeting
- before a handoff
- before asking for approval
- after a major task set is marked done
- before declaring an incident closed

That is where Jini starts feeling useful in daily work. It gives one truthful
state surface for consequential work instead of making you infer status from
documents, tickets, and chat.
