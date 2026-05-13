# Jini

**Jini is a framework with a strict protocol core for governed, stateful AI work.**

Most AI systems are good at first drafts. Jini is built for what happens after
that: revisions, handoffs, verification, approvals, and memory that survives
tool changes.

## Start Here

If you want the fastest proof that Jini is different, run this:

```bash
python3 tools/jini.py status-pack packs/research-prd/examples/research-prd-v1
```

The important lines are:

```text
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
teams already run:

- **Meeting follow-up.** One meeting produces notes, decisions, owners, open
  questions, and action items, but those usually end up split across docs,
  chat, and memory.
- **Research to PRD handoff.** A team knows what it wants to build, but still
  needs a clean path from sources and synthesis into a scoped spec, plan, and
  task set.
- **Vendor selection.** Several options look plausible, stakeholders have
  opinions, and someone still has to preserve the rationale behind the final
  recommendation.
- **Incident response.** The service may be back up, but the operational work
  is not over until owners, evidence, rollback context, and closure state are
  explicit.

Those four examples cover the most common Jini shape: the draft exists, but the
real work still needs ownership, verification, approval, and continuity.

The public repo also proves the same kernel can stretch into personal planning,
such as travel and budgeting, and into more formal workflows such as compliance
audits.

If you want the full examples breakdown, [see the detailed examples page](./examples.md).

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
