---
title: Jini
description: See what is done, what happens next, and what is still missing before work should be treated as durable.
---

<div class="hero-panel">
  <p class="hero-kicker">Less Friction In Daily Work</p>
  <h1 class="hero-title">Finish work without status hunting or late surprises.</h1>
  <p class="hero-summary">Spend less time chasing status, rebuilding context, and discovering missing work late. Jini is a harness orchestration CLI that keeps the next step, blockers, and proof visible while work is still moving.</p>
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

**In plain words:** Jini helps you finish work with less rework by showing what
is done, what comes next, and what is still missing.

If you want the simplest version first, read the [Simple Guide](./simple.md).

Most AI tools are good at starting work. The harder part is finishing without
confusion, repeated explanation, or late surprises. That is the problem this
site is trying to solve.

For technical readers: Jini is a framework with a small protocol core for AI
work that needs durable state, approvals, evidence, memory, and portability.

If you want the shortest explanation of how Jini keeps interaction, state, and
artifacts visible, read [State And Artifacts](./state-and-artifacts.md).

## Start Here

If you want the fastest proof that Jini is different, run this:

```bash
jini example research-prd
jini outcome
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

## What Gets Easier Right Away

Start with the kind of work that already costs you time. Judge Jini by whether
it reduces status hunting, follow-up drift, and cleanup work.

Each example is framed around a productivity gain, not a product claim.

<div class="workflow-grid">
  <a class="workflow-card" href="{{ '/examples.html#meeting-followup' | relative_url }}">
    <span class="workflow-meta">Meeting Follow-up</span>
    <h3>Leave the meeting with clear owners and real follow-through.</h3>
    <p>Spend less time cleaning up scattered notes, implied owners, and missing approvals.</p>
    <code>jini example meeting-followup</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#research-prd' | relative_url }}">
    <span class="workflow-meta">Research To PRD</span>
    <h3>Hand off a spec people can build from without second-guessing it.</h3>
    <p>Reduce rework caused by polished drafts that still hide missing verification or approval.</p>
    <code>jini example research-prd</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#vendor-selection' | relative_url }}">
    <span class="workflow-meta">Vendor Selection</span>
    <h3>Move from comparison to approval without losing the reasoning.</h3>
    <p>Spend less time rebuilding tradeoffs and rationale when someone asks why this option won.</p>
    <code>jini example vendor-selection</code>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#incident-response' | relative_url }}">
    <span class="workflow-meta">Incident Response</span>
    <h3>Recover fast and still close the incident cleanly.</h3>
    <p>Avoid the second wave of work that comes from skipping proof, rollback checks, or final closure steps.</p>
    <code>jini example incident-response</code>
  </a>
</div>

<div class="value-strip">
  <p><strong>What changes in one run:</strong> you spend less time guessing, re-explaining, and fixing preventable misses because the next step and missing work are visible early.</p>
</div>

## What Users Get Back

Those four examples show the same payoff in different kinds of work:

- less time chasing status
- less rework from false “done”
- cleaner handoffs between people
- fewer surprises at approval or closure time

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
jini outcome
jini artifacts
jini show prd
```

That one screen answers the questions teams ask every day:

- What state is this in?
- What happens next?
- What is still missing?
- Are the tasks merely done, or do we have a real outcome yet?

If you want to feel that payoff before you have your own work pack, start with:

```bash
jini example meeting-followup
```

## Bring Your Own Harness

Use the coding harness you already prefer to execute the work. Jini sits above
that harness and keeps the state, artifacts, and next honest step coherent.

```bash
jini harnesses
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
