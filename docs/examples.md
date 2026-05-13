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
python3 tools/jini.py compile-pack meeting-followup \
  --work-unit-id my-meeting-followup \
  --title "Weekly Product Review Follow-up" \
  --purpose "Turn one meeting into decisions owners and next steps" \
  --owner meeting-owner \
  --output /tmp/my-meeting-followup
python3 tools/jini.py status-pack /tmp/my-meeting-followup
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

## 2. Research To PRD Handoff

**Situation:** research exists, opinions exist, and the team agrees something
should be built.

**What usually goes wrong:** the rationale gets thinned out as the work moves
from source material into a product spec, then into a plan, then into tasks.

**Try it with Jini:**

```bash
python3 tools/jini.py status-pack packs/research-prd/examples/research-prd-v1
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

## 3. Vendor Selection

**Situation:** several vendors or tools look plausible and the team needs to
justify a recommendation.

**What usually goes wrong:** the final answer survives, but the scoring,
tradeoffs, objections, and approval path disappear into meetings or email.

**Try it with Jini:**

```bash
python3 tools/jini.py compile-pack vendor-selection \
  --work-unit-id my-vendor-selection \
  --title "Vendor Evaluation" \
  --purpose "Compare shortlisted vendors and prepare an approval-ready recommendation" \
  --owner procurement-lead \
  --approver finance-approver \
  --output /tmp/my-vendor-selection
python3 tools/jini.py status-pack /tmp/my-vendor-selection
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

## 4. Incident Response

**Situation:** the immediate outage is over, but the operational work is not.

**What usually goes wrong:** timeline clarity, rollback context, customer
impact, verification evidence, and closure requirements get reconstructed after
the fact instead of preserved as the work happens.

**Try it with Jini:**

```bash
python3 tools/jini.py compile-pack incident-response \
  --work-unit-id my-incident-response \
  --title "Checkout Latency Incident" \
  --purpose "Stabilize the checkout path with explicit rollback and verification" \
  --owner incident-commander \
  --approver service-owner \
  --output /tmp/my-incident-response
python3 tools/jini.py status-pack /tmp/my-incident-response
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
