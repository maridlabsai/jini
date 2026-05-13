# Jini

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

## What You Can Do Next

- install Jini with `pipx install --editable git+https://github.com/maridlabsai/jini.git`
- [See the proof path](./proof.md)
- [See the install path](./install.md)
- [See detailed examples](./examples.md)
- [See the commercial boundary](./commercial.md)
- [See support and contact paths](./contact.md)
- [See release notes](https://github.com/maridlabsai/jini/releases/tag/v0.1.0)
- [Read the full README](https://github.com/maridlabsai/jini#readme)

## Common Examples

The best way to understand Jini is through a small set of scenarios that most
teams already run, plus the exact state Jini makes visible.

- **Meeting follow-up.** Jini turns one meeting into a tracked follow-up pack
  with explicit decisions, tasks, and missing execution requirements like
  `Approval` and `Evidence` with `jini try-example meeting-followup`.
- **Research to PRD handoff.** Jini shows the most important proof scene in one
  screen: tasks can be `done: 3/3` while the pack is still
  `awaiting_verification` and still missing `Approval` with
  `jini try-example research-prd`.
- **Vendor selection.** Jini compiles a recommendation into an approval-ready
  workflow with visible control packs for `Proof`, `Guard`, `Cost`, and
  `Approval` with `jini try-example vendor-selection`.
- **Incident response.** Jini keeps the response honest by preserving rollback,
  evidence, and closure state after the immediate firefight is over with
  `jini try-example incident-response`.

Those four examples cover the most common Jini shape: the draft exists, but the
real work still needs ownership, verification, approval, and continuity.

The public repo also proves the same kernel can stretch into personal planning,
such as travel and budgeting, and into more formal workflows such as compliance
audits.

If you want the full examples breakdown with commands and output, [see the
detailed examples page](./examples.md).

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
