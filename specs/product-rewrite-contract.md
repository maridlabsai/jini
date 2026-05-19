# Jini Product Rewrite Contract

Updated: 2026-05-13

## Purpose

This document freezes the product contract for the next Jini rewrite phase.
It exists to stop architectural drift, command-surface sprawl, and example
design that proves internal structure more than user value.

The merge-time guardrails for this contract live in:

- [rewrite-guardrails.md](./rewrite-guardrails.md)
- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)
- [product-review-roles.md](./product-review-roles.md)
- [product-consensus-prd-and-plan.md](./product-consensus-prd-and-plan.md)
- [full-product-prd.md](./full-product-prd.md)
- [full-product-prd-execution-plan.md](./full-product-prd-execution-plan.md)
- [local-slm-frontline-policy.md](./local-slm-frontline-policy.md)
- [conversation-and-artifact-ux.md](./conversation-and-artifact-ux.md)
- [adaptive-response-rendering-framework.md](./adaptive-response-rendering-framework.md)
- [adaptive-response-rendering-framework-review.md](./adaptive-response-rendering-framework-review.md)
- [adaptive-response-rendering-framework-gate.md](./adaptive-response-rendering-framework-gate.md)
- [client-surfaces-and-free-tier.md](./client-surfaces-and-free-tier.md)
- [workstream-technical-framework.md](./workstream-technical-framework.md)
- [workstream-technical-framework-review.md](./workstream-technical-framework-review.md)
- [workstream-technical-framework-gate.md](./workstream-technical-framework-gate.md)
- [public-repo-boundary.md](./public-repo-boundary.md)

The rewrite goal is to make Jini the easiest way to turn messy AI work into a
usable, trustworthy result.

Major product decisions under this contract must clear the scorecard gate in:

- [competitive-kpis.yaml](./competitive-kpis.yaml)
- [golden-competitive-benchmark.yaml](./golden-competitive-benchmark.yaml)
- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)

No major UX, routing, artifact, or command-surface decision is considered done
until it has been checked against those score sources.

The public repo boundary is also a hard gate: internal business strategy,
commercialization drafts, and candid internal review notes belong in the
private commercial repo, not in this public tree.

## Product Identity

Jini should help people finish work without losing track of what matters.

Jini should help users:

- get something usable fast
- see what is happening while work moves
- know what is still missing before calling the work done
- continue work later without losing context

Jini should not lead publicly with:

- builder-first language
- internal workflow wiring
- storage-model jargon
- command-family sprawl
- file taxonomy

Those may remain true internally, but they are not the front-door product.

The front-door promise should sound like this:

- tell Jini what you are trying to finish
- Jini asks only what it needs
- Jini gives you something usable
- Jini shows what is still missing
- Jini tells you the next step

## Public Command Surface

Primary human-facing entry:

- `jini`

Scriptable commands:

- `jini run`
- `jini check`
- `jini open`

Human-facing docs should center `jini` first.

Older command families should stay out of the human-facing surface.

Inside `jini`, the public actions are plain-language shell actions:

- `Keep going`
- `Show what's ready`
- `Show what is missing`
- `Help me plan this`
- `Start new work`

When the problem needs planning, `Help me plan this` moves the work into a
Kiro-like sequence:

- goal
- requirements
- design
- steps
- run

## Remembered-Work Rule

Jini should behave like a single-entry product that remembers the thing the
user is currently working on.

Users should not need to manage:

- pack paths
- temp directories
- artifact locations
- internal file stems

If current work exists, `jini` should offer:

- continue current work
- open what is ready
- start something new

Jini must never switch current work silently.

Jini must always make the remembered work visible before acting on it.

## First Result Screen

The first screen after intake must be the thing the user asked Jini to make.

Default first result examples:

- meeting: `Sendable Follow-up`
- plan/spec: `Build-Readiness Check`
- vendor/options: `Recommendation Memo`
- incident: `Closure Checklist`
- trip: `Trip Plan`

Rules:

- Do not show the work summary before the first useful result on first run.
- Do not ask the user to choose output size before the first useful result.
- Default silently to a quick useful draft.
- Offer `Make it fuller` only after the first result exists.
- The first result must include enough content to be useful without opening a second screen.

After the first result, Jini should offer:

- `Keep going`
- `Make it fuller`
- `Show what is missing`
- `Start something new`

## Work Summary Screen

The work summary screen should read like a calm recap, not telemetry.
It appears after the first result, when the user asks what is missing, or when
the user returns to current work.

Default sections:

- `You're working on`
- `Jini is using`
- `Jini is doing`
- `Ready now`
- `Still missing`
- `Not sure about`
- `Next step`
- `Safe to do`

Layout:

```text
You're working on
Research to PRD handoff

Jini is using
Latest PRD draft and review comments

Jini is doing
Checking assumptions and approval gaps
2 of 4 steps done

Ready now
- Build-readiness check
- Handoff brief

Still missing
- Product approval
- Rollback note for notification frequency

Not sure about
- Whether approval was already granted in the review thread

Next step
Open Build-readiness check

Safe to do
Nothing has been sent yet. You can review before sharing.
```

Rules:

- Do not show `STATE`, `HEALTH`, `TASKS`, or raw file paths on the default screen.
- Do not show internal artifact stems on the default screen.
- Show what Jini is working from in plain language.
- Show uncertainty explicitly when it affects trust.
- `Still missing` must list real missing proof, blockers, or unresolved ownership only.
- `Next step` must be one recommendation, not a menu.
- `Safe to do` must reassure the user when the output is still a draft or reversible.

## Interaction Model

## Approved Product Shape

The approved shape is a complete work loop:

1. user gives Jini messy context
2. Jini returns a useful object
3. Jini shows what is missing and what is safe
4. user can continue, make it fuller, inspect trust, or start new work

The product must not stop at launcher intake.

For the replacement-critical path, only two flows are approved:

- meeting follow-up
- plan/spec readiness

Trip planning can remain a demo until the two flagship flows pass the same
quality and parity gates.

### First Run

1. User runs `jini`.
2. If no current work exists, Jini asks:
   - `What do you need help finishing?`
3. Jini supports natural paste-first intake:
   - `Jini shell`
   - `Paste messy notes, or type the outcome you want.`
4. Jini shows example inputs without requiring command selection:
   - Turn meeting notes into something I can send
   - Check whether a plan is ready to hand off
   - Help me plan this
   - I am not sure
5. If the user chooses `I am not sure`, Jini says:
   - `Paste what you have. A rough version is fine.`
   - `I will help figure out whether this is follow-up, a plan check, or something else.`
   - `Nothing will be sent yet.`
   If the input does not clearly match a flagship flow, Jini still returns a
   useful first object called `First Useful Pass` with:
   - what the user appears to be trying to finish
   - what can be used now
   - what Jini needs next
   - what is safe because nothing has been sent
6. If the user chooses a job, Jini asks one plain source question:
   - `Paste what you have. A rough version is fine.`
7. Inline helper text may show one example for the selected job.
8. Jini asks only blocking questions after that.
9. Jini silently defaults to a quick useful draft.
10. The first visible result is the useful object itself.
11. The work summary appears after the useful object, or when the user asks what is still missing.
12. Jini ends the first pass with clear actions:
   - keep going
   - make it fuller
   - see what is still missing
   - start something new

### Post-Result Continuation

At least one real continuation path must exist in the replacement-critical
slice.

For meeting follow-up and plan/spec readiness, these actions are required in
Phase 1:

- `Keep going`
- `Show what is missing`

Both must be real. They must not be placeholders.

### Continuing Current Work

1. User runs `jini`.
2. If current work exists, Jini shows a compact recap:
   - what the work is
   - what is ready
   - what is still missing
   - what to do next
3. Jini offers exactly three choices:
   - continue current work
   - open what is ready
   - start something new
4. If continuing, Jini returns to the stable status screen and resumes in place.
5. If opening, Jini shows the output shelf.
6. If starting new work, Jini explicitly parks the previous work.

## Must-Have Competitor Merits

The rewrite must treat the strongest competitor behaviors as product
requirements, not inspiration.

### Claude Code: Immediate Payoff

Jini must match the feeling of asking for work and seeing useful work back.

Required behavior:

- no required setup vocabulary in the first minute
- no `short` / `full` gate before the first result
- useful output appears before internal state explanation
- one obvious next action after the first result

### Kiro: Visible Progression

Jini must keep the work visible without making the user inspect internals.

Required behavior:

- show what Jini is working from
- show what Jini is doing in plain language
- show what is ready to open
- show what is still missing before the user treats work as done

### Hermes: Continuity

Jini must make returning to work feel dependable, not magical.

Required behavior:

- current work is shown before Jini acts
- current work can be continued, opened, or parked
- Jini never silently switches remembered work
- stale or missing remembered work is handled in plain language

### AgentField: Inspectability

Jini must expose enough trace to trust the output without making inspection the
default experience.

Required behavior:

- show the source used
- show assumptions that affect trust
- show missing approval, evidence, ownership, or blockers
- keep deeper artifacts and provenance openable on demand

### Jini: Honest Closure

Jini must preserve its own lead: work can look done while important truth is
still missing.

Required behavior:

- no useful object is treated as final just because it exists
- `Still missing` and `Not sure about` remain visible when they matter
- `Safe to do` explains what has not been sent, changed, or committed
- approvals and evidence are simplified in language, not removed

## Output Shelf

The output shelf is a product surface, not a file browser.

Users should open human-ready objects such as:

- `Sendable Follow-up`
- `Owners and Due Points`
- `Decisions Made`
- `Open Questions`
- `Build-Readiness Check`
- `Handoff Brief`
- `Missing Pieces Before Build`
- `Recommendation Memo`
- `Tradeoff Table`
- `Closure Checklist`
- `Risk List`

Do not surface internal stems such as:

- `followup`
- `prd`
- `selection`
- `response`

Shelf groups:

- `Ready to use`
- `Useful next`
- `Background`

Default shelf rules:

- `Ready to use` appears first.
- `Background` is hidden unless requested.
- Avoid generic labels like `Task List` when a job-specific label is clearer.

## Hero Flows

The rewrite should optimize first for two hero flows:

1. meeting follow-up
2. spec readiness

### Meeting Follow-up

The first useful outcome must be:

- a sendable follow-up
- an owner list
- open questions
- decisions made

The current failure mode is producing a plan-shaped structure instead of a
sendable work object.

### Spec Readiness

The first useful outcome must be:

- a build-readiness answer
- a handoff brief
- a missing-pieces list
- a risk list

This is the strongest current Jini example and should become the flagship flow.

## UX Principles

The rewrite should follow these rules:

1. usable output beats elegant structure
2. visible progress beats hidden rigor
3. natural intake beats command taxonomy
4. one next step beats state explanation
5. human-ready outputs beat file-oriented outputs
6. continuity must feel helpful, not magical
7. deep detail stays available, but hidden by default
8. relief matters more than conceptual neatness
9. the user should feel safer, not inspected

## Architecture Direction

End-state runtime:

- Go binary CLI

Why:

- fast cold start
- strong local filesystem and process model
- easy static distribution
- simpler than Rust for this product stage
- better final posture than keeping Python public

Python should remain a migration oracle, not the final public runtime.

## Rewrite Principles

- Preserve on-disk contracts first.
- Do not rewrite the work format and runtime at the same time.
- Keep the public face smaller than the internal operator surface.
- Port the read path before the write path.
- Keep the system local-first and deterministic where possible.
- Preserve governance, evidence, and approval semantics.
- Preserve the small-kernel rule.

## Migration Waves

### Wave 0: Freeze Contracts

- create golden fixtures for work directories, manifests, receipts, and exports
- create golden snapshots for public CLI outputs
- add dual-runtime diff tests against Python for selected scenarios

### Wave 1: Port the Fast Read Path

- remembered-work store
- `check`
- `open`
- output shelf naming
- artifact catalog resolution

### Wave 2: Ship the Binary Shell

- package and ship a real binary CLI
- keep only the public command surface in the shell
- allow Python fallback for unported advanced commands

### Wave 3: Port the Run Loop

- `run`
- stable live screen
- blocking-question loop
- ready-now shelf updates
- continue-or-start-new behavior

### Wave 4: Port Install And Admin Surfaces

- manifest planner
- install/update/uninstall/doctor
- receipts and provenance

These must remain operationally strong, but out of the beginner path.

### Wave 5: Port Harness And Publish Adapters

- harness selection and execution
- adapter bridges
- conformance checks

### Wave 6: Cutover

- Python becomes compatibility oracle only
- remove Python from the public runtime story

## Keep / Hide / Delete

### Keep

- work-unit contracts
- artifacts, views, and exports
- manifests and receipts
- governance, evidence, and approval semantics
- execution-routing policy
- pack system and adapter registry
- regression suite, expanded into golden compatibility tests

### Hide

- install/admin internals from the beginner path
- raw bundle, kit, and manifest language
- scorecard and improvement tooling
- low-level activation/handoff semantics
- builder-first language on the public face

### Delete

- Python-first public install story
- editable install as the primary product path
- giant monolithic public argparse surface
- duplicate public verbs as first-class product concepts
- any rewrite plan that changes the work format and runtime simultaneously

## Success Bar

The rewrite succeeds if:

- install feels native
- startup feels instant
- first useful output appears in minutes
- normal use requires no file paths
- the work remains visible while it moves
- outputs are openable and human-ready
- users understand what is still missing before calling work done
- Jini feels simpler without losing rigor

## Benchmark Intent

The rewrite should move Jini toward:

- Claude Code’s immediacy
- Kiro’s visible progression
- Hermes’s continuity
- AgentField’s inspectability
- Jini’s own honest-completion truth

That combination is the target product posture.

## What End-User Research Changed

The rewrite contract originally leaned too far toward builder taste:

- compact command symmetry
- shell-within-a-shell thinking
- short telemetry labels
- truth-first screens before relief-first outputs

End-user critique changed that.

The revised rules are:

- `jini` should be a launcher or dashboard, not a second shell
- the screen should use plain phrases, not operator labels
- the first output must be useful before the system explains itself
- remembered work must be visible and controllable
- users need reassurance about what Jini used, what it is unsure about, and what is safe to do next
