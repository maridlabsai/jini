# Vendor Selection Pack

This pack compiles a commercial evaluation workflow into the same Jini core
semantics used by delivery, personal planning, incident response, and regulated
audit work.

It is meant for:

- selecting between vendors or partner tools
- making approval-ready commercial recommendations
- keeping scoring, tradeoffs, and follow-through visible in one surface

The pack emits canonical artifacts plus a rendered selection view and portable
publish bundles.

## What it proves

- advanced-set breadth can grow through packs instead of kernel growth
- commercial decision workflows can stay evidence-bound and approval-ready
- portable markdown handoff works for non-software evaluation work too

## Compile

```bash
jini compile-pack vendor-selection \
  --work-unit-id my-vendor-selection \
  --title "Vendor Evaluation" \
  --purpose "Compare shortlisted vendors and prepare an approval-ready recommendation" \
  --owner procurement-lead \
  --approver finance-approver \
  --output /tmp/my-vendor-selection
jini status-pack /tmp/my-vendor-selection
```
