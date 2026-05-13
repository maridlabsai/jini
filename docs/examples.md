# Common Examples

Jini is easiest to understand when it is doing work you already recognize.

These four examples are the best public story because they show both sides of
the product:

- the command you run
- the state Jini makes explicit

That is the difference between "another workflow idea" and a framework you can
actually operate.

## 1. Meeting Follow-up

**Situation:** a weekly product, staff, or project meeting ends with notes
scattered across docs, chat, and memory.

**What usually goes wrong:** nobody can later say with confidence what was
actually decided, who owns the next step, what is still open, or whether
anything needs approval before execution.

**Try it with Jini:**

```bash
jini try-example meeting-followup
```

**What Jini shows back:**

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

**Why that helps:** the meeting is no longer "captured" just because notes
exist. Jini turns the follow-up into explicit artifacts, visible task load, and
an honest statement of what is still missing before the work is truly
execution-ready.

**What changes day to day:**

- the meeting owner stops guessing what people heard
- action items stop living only in chat threads
- approvers can see what is still missing before work starts
- the next person inherits state, not just notes

## 2. Research To PRD Handoff

**Situation:** research exists, opinions exist, and the team agrees something
should be built.

**What usually goes wrong:** the rationale gets thinned out as the work moves
from source material into a product spec, then into a plan, then into tasks.

**Try it with Jini:**

```bash
jini try-example research-prd
```

**What Jini shows back:**

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

**Why that helps:** this is the proof scene at the center of Jini. The tasks
are done, but the work is not pretending to be finished. Jini preserves the gap
between completion, verification, and approval, while keeping the evidence
trail attached to the current revision.

**What changes day to day:**

- product and engineering stop arguing from different versions of the rationale
- people can see whether the spec is ready or merely drafted
- verification becomes a visible stage instead of an implied one
- the handoff keeps its source trail attached to the work

## 3. Vendor Selection

**Situation:** several vendors or tools look plausible and the team needs to
justify a recommendation.

**What usually goes wrong:** the final answer survives, but the scoring,
tradeoffs, objections, and approval path disappear into meetings or email.

**Try it with Jini:**

```bash
jini try-example vendor-selection
```

**What Jini shows back:**

```text
HEALTH ready-to-make
STATE  decided
NEXT   Make
CTRL   Proof, Guard, Cost, Approval
MISSING-LATER
  - Approval
  - Evidence
```

**Why that helps:** Jini makes vendor selection look like consequential work,
not a slide deck. The recommendation has control surfaces, a visible approval
path, and an explicit reminder that evidence still has to be bound before the
decision should be treated as durable.

**What changes day to day:**

- the recommendation survives beyond the meeting where it was made
- tradeoffs stay attached to the final answer
- finance or leadership can see the approval path without re-asking for context
- the team can revisit the decision later without reconstructing it from memory

## 4. Incident Response

**Situation:** the immediate outage is over, but the operational work is not.

**What usually goes wrong:** timeline clarity, rollback context, customer
impact, verification evidence, and closure requirements get reconstructed after
the fact instead of preserved as the work happens.

**Try it with Jini:**

```bash
jini try-example incident-response
```

**What Jini shows back:**

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

**Why that helps:** Jini keeps the response honest. "The service is back" is
not the same thing as "the incident is closed." The workflow still has explicit
rollback, proof, and approval requirements before anyone can claim the work is
done.

**What changes day to day:**

- responders stop treating recovery as closure
- rollback context stays visible while pressure is high
- verification evidence gets attached before the story drifts
- closure becomes a real state, not an assumption

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
