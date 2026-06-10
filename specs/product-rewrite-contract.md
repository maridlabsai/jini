# Jini Product Rewrite Contract

Updated: 2026-06-10

## Purpose

This document freezes the product contract for the next Jini rewrite phase.
It exists to stop architectural drift, command-surface sprawl, and example
design that proves internal structure more than user value.

This document is a rewrite guardrail, not the top-precedence product and
operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this rewrite contract and the canonical PRD disagree on product tenets,
priorities, requirements, or roadmap order, the canonical PRD wins.

Current status: rewrite-era claims about a launcher dashboard, startup
`Start`/`Keep` choices, `Task Snapshot` fallback artifacts, broad flow
scaffolding, or app-wide surfaces are stale unless the canonical PRD restates
them. The current release contract is CLI-first: compact task prompt, direct
answers, safe local edits, explicit saved-work commands, real route handoff or
fail-closed setup guidance, and no broad OS or agent-suite claim.

The merge-time guardrails for this contract live in:

- [rewrite-guardrails.md](./rewrite-guardrails.md)
- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)
- [product-review-roles.md](./product-review-roles.md)
- [number-one-platform-prd.md](./number-one-platform-prd.md)
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

Small support commands:

- `jini status`
- `jini continue`
- `jini open`
- `jini route`
- `jini doctor`

Human-facing docs should center `jini` first. Older command families and
admin plumbing should stay out of the beginner path.

Inside `jini`, natural language remains primary. Planning is a behavior Jini
may choose for complex work, not a mandatory startup action label.

## Remembered-Work Rule

Jini should behave like a single-entry product that remembers the thing the
user is currently working on.

Users should not need to manage:

- pack paths
- temp directories
- artifact locations
- internal file stems

Current R0 rule: bare `jini` still starts at the task prompt. Remembered work
is surfaced through `status`, `continue`, `open`, explicit work-state
questions, or natural title matching. Jini must never switch remembered work
silently, but it must not turn every new input into a saved-work dashboard.

## First Result Screen

The first screen after intake must be the thing the user asked Jini to make.

Durable artifact first-result examples:

- meeting: `Sendable Follow-up`
- plan/spec: `Build-Readiness Check`
- vendor/options: `Recommendation Memo`
- incident: `Closure Checklist`
- trip: `Trip Plan`

Rules:

- Do not show the work summary before the first useful result on first run.
- Do not ask the user to choose output size before the first useful result.
- Default silently to a quick useful result when the request truly needs a
  durable artifact.
- Do not offer decorative continuation actions. If Jini cannot materially
  change or open a useful artifact from the action, remove it from the default
  surface.
- The first result must include enough content to be useful without opening a second screen.

After the first result, Jini should show one useful next step or the compact
action receipt. It should not add decorative continuation labels that do not
change, open, or inspect real work.

## Work Summary Screen

The work summary screen should read like a calm recap, not telemetry. It is a
durable-work view only. It appears after useful artifact work exists and the
user asks for status, missing pieces, open outputs, or continuation.

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
4. user can continue, inspect gaps, plan next work, inspect trust, or start new work

The product must not stop at launcher intake.

For the replacement-critical path, only two flows are approved:

- meeting follow-up
- plan/spec readiness

Trip planning can remain a demo until the two flagship flows pass the same
quality and parity gates.

### First Run

1. User runs `jini`.
2. If no current work exists, Jini asks:
   - `Jini`
3. Jini supports natural paste-first intake:
   - `Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.`
4. Jini shows example inputs without requiring command selection:
   - Add a line to the matching .txt file in this folder
   - Turn meeting notes into something I can send
   - Check whether a plan is ready to hand off
   - Plan a 7 day Paris trip for two adults in October
   - Compare these vendors and recommend one
5. If the input is unclear, Jini asks the shortest blocking question or fails
   closed with candidates.
6. Jini does not create a generic snapshot, working draft, or saved-work
   scaffold for simple questions, bare entities, greetings, acknowledgements,
   or unclear input.
7. Jini asks only blocking questions after that.
8. Jini defaults to the cheapest safe route that can complete the task.
9. The first visible result is the compact answer, file-edit receipt, setup
   guidance, or useful work object itself.
10. Work summaries appear only after useful work exists and the user asks for
    status, continuation, opening, or missing pieces.

### Post-Result Continuation

At least one real continuation path must exist in the replacement-critical
slice.

Continuation labels from the rewrite phase are examples, not current startup
chrome. Any exposed continuation must be real: continue the work, inspect a
missing piece, open an artifact, or ask a blocking question. Placeholders are
not allowed.

### Continuing Current Work

1. User asks for `jini status`, `jini continue`, `jini open`, a work-state
   question, or a natural saved-work title.
2. Jini shows only the compact recap needed for that request.
3. If continuing, Jini resumes in place.
4. If opening, Jini shows the output shelf.
5. If switching remembered work, Jini makes the switch explicit.

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
- better final posture than keeping an interpreted public runtime

The runtime cutover is complete: public CLI behavior must be native Go, and
future parity work must be implemented as Go tests and Go command handlers.

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
- add native Go golden-output tests for selected scenarios

### Wave 1: Port the Fast Read Path

- remembered-work store
- `check`
- `open`
- output shelf naming
- artifact catalog resolution

### Wave 2: Ship the Binary Shell

- package and ship a real binary CLI
- keep only the public command surface in the shell
- remove fallback for unported advanced commands; unsupported commands fail fast
  until they are ported natively

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

- remove the compatibility runtime from the public runtime story
- keep future parity checks native Go-only

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

- interpreted-runtime public install story
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

- `jini` should be a compact task prompt, not a dashboard or second shell
- the screen should use plain phrases, not operator labels
- the first output must be useful before the system explains itself
- remembered work must be visible and controllable
- users need reassurance about what Jini used, what it is unsure about, and what is safe to do next
