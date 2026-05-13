# Proof

Jini’s public proof is simple:

**done does not mean verified**

Run:

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

This is the difference Jini is trying to preserve:

- tasks can be complete
- evidence can be attached
- the work can still be blocked from advancing

That is not a UI choice. It is the protocol core refusing to collapse
governance into a green checklist.

For the longer walkthrough, see [PROOF_OF_DIFFERENCE.md](../PROOF_OF_DIFFERENCE.md).
