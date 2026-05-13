---
title: Jini
description: See what is done, what happens next, and what is still missing before work should be treated as durable.
---

<div class="hero-panel">
  <p class="hero-kicker">Truthful AI Workflows</p>
  <h1 class="hero-title">See what is done, what happens next, and what is still missing.</h1>
  <p class="hero-summary">Jini is for the part after the exciting first draft: handoffs, checking, sign-off, and the quiet missing work that turns “done” into a real outcome instead of rework later.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/proof.html' | relative_url }}">See the 30-second proof</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Start with the simple guide</a>
    <a class="cta-button cta-button-secondary" href="{{ '/install.html' | relative_url }}">Install Jini</a>
  </div>
  <div class="fact-strip">
    <div class="fact-pill">
      <strong>Question 1</strong>
      <p>What is actually done?</p>
    </div>
    <div class="fact-pill">
      <strong>Question 2</strong>
      <p>What happens next?</p>
    </div>
    <div class="fact-pill">
      <strong>Question 3</strong>
      <p>What is still missing?</p>
    </div>
  </div>
</div>

**In plain words:** Jini helps you get work across the line by showing what is
done, what comes next, and what is still missing.

If you want the simplest version first, read the [Simple Guide](./simple.md).

Most AI tools are good at starting work. Jini is for the harder part after
that: handoffs, checking, sign-off, and keeping the truth clear as work moves
between people and tools.

For technical readers: Jini is a framework with a small protocol core for AI
work that needs durable state, approvals, evidence, memory, and portability.

If you want the shortest explanation of how Jini keeps interaction, state, and
artifacts visible, read [State And Artifacts](./state-and-artifacts.md).

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
a usable outcome.

![Jini research to PRD proof demo](./assets/examples/research-prd.gif)

<div class="section-card">
  <strong>Read this screen in plain English:</strong>
  <ul>
    <li><code>done: 3/3</code> means the tasks were completed</li>
    <li><code>STATE awaiting_verification</code> means the work still needs one more step before it becomes a safe outcome</li>
    <li><code>MISSING-LATER Approval</code> means a later blocker is already visible now</li>
  </ul>
</div>

## Pick Your Workflow

Start with the workflow you already ran this week. Jini is easiest to judge
when it is making ordinary work more truthful.

Each example answers a user question, not a framework question.

<div class="workflow-grid">
  <a class="workflow-card" href="{{ '/examples.html#meeting-followup' | relative_url }}">
    <span class="workflow-meta">Meeting Follow-up</span>
    <h3>Your meeting ended, but the follow-through still is not real.</h3>
    <p>Use this when notes, owners, and approvals are about to get scattered before any actual follow-up happens.</p>
    <code>jini try-example meeting-followup</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#research-prd' | relative_url }}">
    <span class="workflow-meta">Research To PRD</span>
    <h3>Your spec looks done, but the team still does not have a safe build handoff.</h3>
    <p>Use this when research exists, tasks are complete, and the handoff is starting to look more finished than it really is.</p>
    <code>jini try-example research-prd</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#vendor-selection' | relative_url }}">
    <span class="workflow-meta">Vendor Selection</span>
    <h3>You need a vendor choice that can actually move into approval and action.</h3>
    <p>Use this when several vendors look plausible and you need a choice that survives beyond the meeting where it was made.</p>
    <code>jini try-example vendor-selection</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#incident-response' | relative_url }}">
    <span class="workflow-meta">Incident Response</span>
    <h3>The outage is over, but the work still is not ready for true closure.</h3>
    <p>Use this when rollback, proof, and final signoff still matter after service recovery.</p>
    <code>jini try-example incident-response</code>
  </a>
</div>

<div class="value-strip">
  <p><strong>What changes in one run:</strong> Jini gives you one truthful path from draft to outcome. You can see what happened, what comes next, and what is still missing before the work should be treated as durable.</p>
</div>

## Why These Four

Those four workflows are the shortest honest explanation of Jini:

- work happened
- a draft or tentative answer exists
- people are tempted to call it done
- Jini shows what is still missing before the work becomes a real outcome

The public repo also proves the same kernel can stretch into personal planning,
such as travel and budgeting, and into more formal workflows such as compliance
audits. If you want the full examples breakdown with commands, output, and GIF
walkthroughs, [see the detailed examples page](./examples.md).

## Day-To-Day Value

Jini should earn its keep in normal work, not only in edge cases.

Here is the practical payoff in day-to-day flows:

- **After a meeting:** you can turn loose notes into explicit follow-up,
  owners, tasks, and missing requirements instead of hoping the follow-up stays
  coherent.
- **Before engineering starts:** you can check whether a research-backed spec
  can become a safe handoff, or whether the team is about to build from an
  unfinished draft.
- **Before asking for approval:** you can show the current state, missing
  evidence, and rationale in one place instead of assembling the outcome story from docs,
  chat, and memory.
- **During handoffs:** the next person can run one command and see what is
  done, what is missing, and what happens next to move the work forward.
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
- Are the tasks merely done, or do we have a real outcome yet?

If you want to feel that payoff before you have your own work pack, start with:

```bash
jini try-example meeting-followup
```

## Explore More

<div class="section-card">
  <h3>Go next</h3>
  <div class="on-this-page">
    <a href="./simple.md">Read the Simple Guide</a>
    <a href="./state-and-artifacts.md">See state and artifacts clearly</a>
    <a href="./cli.md">See the short CLI guide</a>
    <a href="./proof.md">See the proof path</a>
    <a href="./install.md">See the install path</a>
    <a href="./examples.md">See detailed examples</a>
    <a href="./contact.md">See support and contact paths</a>
    <a href="https://github.com/maridlabsai/jini/releases/tag/v0.1.0">See release notes</a>
  </div>
</div>

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
