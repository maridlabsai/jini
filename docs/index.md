# Jini

<p class="hero-mark-wrap">
  <img class="hero-mark" src="{{ '/assets/brand/jini-mark-512.png' | relative_url }}" alt="Jini mark">
</p>

**Jini is a framework with a strict protocol core for governed, stateful AI work.**

Most AI systems are good at first drafts. Jini is built for what happens after
that: revisions, handoffs, verification, approvals, and memory that survives
tool changes.

## Start Here

If you want the fastest proof that Jini is different, run this:

```bash
jini try-example research-prd
```

The important lines are:

```text
EXAMPLE Research To PRD Handoff
HEALTH ready-to-verify
STATE  awaiting_verification
NEXT   Verify
MISSING-LATER
  - Approval
TASKS
  done:       3/3
  unresolved: 0/3
```

That is the core idea in one screen: completed tasks are not the same thing as
verified, approved, publishable work.

![Jini research to PRD proof demo](./assets/examples/research-prd.gif)

## Pick Your Workflow

Start with the workflow you already ran this week. Jini is easiest to judge
when it is making ordinary work more truthful.

<div class="workflow-grid">
  <a class="workflow-card" href="{{ '/examples.html#meeting-followup' | relative_url }}">
    <span class="workflow-meta">Meeting Follow-up</span>
    <h3>Turn notes into decisions, owners, and missing approvals.</h3>
    <p>Use this when one meeting ended with vague next steps and everyone is pretending the follow-up is obvious.</p>
    <code>jini try-example meeting-followup</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#research-prd' | relative_url }}">
    <span class="workflow-meta">Research To PRD</span>
    <h3>Show whether a spec is ready, or only drafted.</h3>
    <p>Use this when research exists, tasks are done, and you need to know whether the handoff is actually verified.</p>
    <code>jini try-example research-prd</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#vendor-selection' | relative_url }}">
    <span class="workflow-meta">Vendor Selection</span>
    <h3>Keep the rationale attached to the recommendation.</h3>
    <p>Use this when several options look plausible and the team needs an approval-ready answer instead of a meeting recap.</p>
    <code>jini try-example vendor-selection</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#incident-response' | relative_url }}">
    <span class="workflow-meta">Incident Response</span>
    <h3>Separate service recovery from real closure.</h3>
    <p>Use this when the outage is over but rollback, proof, and closure work still need to be made explicit.</p>
    <code>jini try-example incident-response</code>
  </a>
</div>

<div class="value-strip">
  <p><strong>What changes in one run:</strong> Jini gives you one truthful state surface. You can see what happened, what comes next, and what is still missing before the work should be treated as durable.</p>
</div>

## Why These Four

Those four workflows are the shortest honest explanation of Jini:

- work happened
- a draft or decision exists
- people are tempted to call it done
- Jini shows what is still missing before the work should be treated as durable

The public repo also proves the same kernel can stretch into personal planning,
such as travel and budgeting, and into more formal workflows such as compliance
audits. If you want the full examples breakdown with commands, output, and GIF
walkthroughs, [see the detailed examples page](./examples.md).

## Day-To-Day Value

Jini should earn its keep in normal work, not only in edge cases.

Here is the practical payoff in day-to-day flows:

- **After a meeting:** you can turn loose notes into explicit decisions,
  owners, tasks, and missing requirements instead of hoping the follow-up stays
  coherent.
- **Before engineering starts:** you can check whether a research-backed spec
  is actually verified, or whether the team is about to build from an
  unverified draft.
- **Before asking for approval:** you can show the current state, missing
  evidence, and rationale in one place instead of assembling it from docs,
  chat, and memory.
- **During handoffs:** the next person can run one command and see what is
  done, what is missing, and what happens next.
- **After an incident:** you can tell the difference between "the service is
  back" and "the work is actually ready for closure."

The recurring Jini move is simple:

```bash
jini status-pack <path-to-work>
```

That one screen answers the questions teams ask every day:

- What state is this in?
- What happens next?
- What is still missing?
- Are the tasks merely done, or is the work actually verified?

If you want to feel that payoff before you have your own work pack, start with:

```bash
jini try-example meeting-followup
```

## Explore More

- install Jini with `pipx install --editable git+https://github.com/maridlabsai/jini.git`
- [See the proof path](./proof.md)
- [See the install path](./install.md)
- [See detailed examples](./examples.md)
- [See the commercial boundary](./commercial.md)
- [See support and contact paths](./contact.md)
- [See release notes](https://github.com/maridlabsai/jini/releases/tag/v0.1.0)
- [Read the full README](https://github.com/maridlabsai/jini#readme)

## Public Core Boundary

Free and public:

- framework code
- CLI
- protocol and schema docs
- install path
- example packs
- proof path
- tests

Paid later:

- implementation help
- onboarding workshops
- design-partner work
- premium domain or control surfaces after repeated demand proves they are worth building
- enterprise support, integrations, and governance

The public repo is meant to be usable on its own. The paid path is for
acceleration, customization, and enterprise trust.
