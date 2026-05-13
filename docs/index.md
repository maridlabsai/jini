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
- [See the commercial boundary](./commercial.md)
- [See support and contact paths](./contact.md)
- [See release notes](https://github.com/maridlabsai/jini/releases/tag/v0.1.0)
- [Read the full README](https://github.com/maridlabsai/jini#readme)

## Common Examples

Jini’s public core includes examples people already understand:

- meeting follow-up
- research to PRD
- travel planning
- budget planning

That makes the framework easier to evaluate before you ever reach the more
specialized operational or regulated surfaces.

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
