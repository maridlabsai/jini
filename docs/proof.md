---
title: Proof
description: One command shows the difference between finished tasks and a real usable outcome.
---

Jini’s public proof is simple:

**done does not mean outcome**

Run:

```bash
jini example research-prd
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

This is the difference Jini is trying to preserve:

- tasks can be complete
- evidence can be attached
- the work can still fall short of a safe outcome

![Jini research to PRD proof demo](./assets/examples/research-prd.gif)

That is not a UI choice. It is the protocol core refusing to confuse activity
with outcome.

If you want the same proof shape on a more ordinary workflow, run:

```bash
jini example meeting-followup
```

For the longer walkthrough, see [PROOF_OF_DIFFERENCE.md](https://github.com/maridlabsai/jini/blob/main/PROOF_OF_DIFFERENCE.md).
