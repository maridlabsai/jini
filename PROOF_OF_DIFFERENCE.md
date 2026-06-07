# Jini: Proof of Difference

This file is not a category argument.

It is one concrete artifact trail from the repo that shows what Jini is trying
to keep coherent as work moves from research into handoff.

## One Inspectable Moment

Run:

```bash
jini open packs/research-prd/examples/research-prd-v1/views/tasks.md
```

The important lines are:

```text
State: awaiting_verification
Health: ready-to-verify
Next: Verify
Done: 3
Unresolved: 0
```

This is the moment that matters.

All tasks are done.

The work still does not advance.

Why?

Because task completion is not the same thing as verified, publish-ready work.

That is not a UX flourish.

That is the protocol core refusing to collapse governance into a green
checklist.

## The Example

The example is the `research-prd-v1` work unit in:

- [work-unit.yaml](./packs/research-prd/examples/research-prd-v1/work-unit.yaml)
- [artifacts/09-tasks.yaml](./packs/research-prd/examples/research-prd-v1/artifacts/09-tasks.yaml)
- [artifacts/10-evidence.yaml](./packs/research-prd/examples/research-prd-v1/artifacts/10-evidence.yaml)
- [views/tasks.md](./packs/research-prd/examples/research-prd-v1/views/tasks.md)
- [exports/issues/github/issues.json](./packs/research-prd/examples/research-prd-v1/exports/issues/github/issues.json)
- [exports/wiki/markdown/prd.md](./packs/research-prd/examples/research-prd-v1/exports/wiki/markdown/prd.md)

## What Survives Across The Trail

### 1. The work unit still has explicit state

The root work unit is still visible as a first-class object:

- `work_unit_id`: `research-prd-v1`
- `current_state`: `awaiting_verification`
- `profile_id`: `Delivery`
- `purpose`: `Turn validated research into a PRD and build-ready handoff`

Source:
- [work-unit.yaml](./packs/research-prd/examples/research-prd-v1/work-unit.yaml)

### 2. The task artifact preserves progress and references

The task artifact is not just a checklist. It carries:

- `revision: 4`
- `status: reviewed`
- all `3` tasks marked `done`
- output notes describing what happened
- output refs pointing back to canonical artifacts and downstream exports

The third task explicitly ties the handoff to:

- `artifacts/06-spec.yaml`
- `artifacts/08-plan.yaml`
- `artifacts/09-tasks.yaml`
- `exports/issues/jira/issues.json`

Source:
- [artifacts/09-tasks.yaml](./packs/research-prd/examples/research-prd-v1/artifacts/09-tasks.yaml)

### 3. Evidence is attached to the work, not floating around it

The evidence artifact is bound to the same work unit and target artifact:

- `artifact_type: Evidence`
- `status: reviewed`
- `target_artifact_id: spec-research-prd-v1`
- validated claims
- review results
- residual risks

This means the work does not just have output. It has attached justification
and remaining uncertainty.

Source:
- [artifacts/10-evidence.yaml](./packs/research-prd/examples/research-prd-v1/artifacts/10-evidence.yaml)

### 4. The rendered task view stays consistent with the canonical state

The human-readable task board still reflects the same work:

- `State: awaiting_verification`
- `Health: ready-to-verify`
- `Next: Verify`
- `Done: 3`
- `Unresolved: 0`

Source:
- [views/tasks.md](./packs/research-prd/examples/research-prd-v1/views/tasks.md)

### 5. The same context survives export into downstream systems

The exported GitHub issue payload still carries:

- the same `work_unit_id`
- the same `state`
- the same `health`
- task-level evidence context
- current output refs

That means the downstream issue surface is not detached from the work that
produced it.

Source:
- [exports/issues/github/issues.json](./packs/research-prd/examples/research-prd-v1/exports/issues/github/issues.json)

### 6. The PRD stays derived from the same source of truth

The rendered PRD export carries forward:

- the same problem statement
- explicit stakeholders
- success criteria
- evidence highlights
- requirements
- delivery slices

Source:
- [exports/wiki/markdown/prd.md](./packs/research-prd/examples/research-prd-v1/exports/wiki/markdown/prd.md)

## Why This Matters

This example is not exciting because it is dramatic.

It is useful because it is ordinary.

Research became a PRD, a task surface, an evidence record, an issue export, and
a wiki export without losing:

- what the work is
- what state it is in
- which claims were validated
- which artifacts are canonical
- what still happens next

That is the difference Jini is trying to preserve.
