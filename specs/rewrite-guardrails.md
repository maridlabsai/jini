# Jini Rewrite Guardrails

Updated: 2026-05-14

## Purpose

This checklist exists to keep the rewrite from accidentally improving the
front-door experience while damaging the scorecard lead underneath it.

Use it before merging any rewrite change that touches:

- command surface
- current-work behavior
- work-unit semantics
- artifacts, views, or exports
- install/distribution
- packs or adapters
- governance, evidence, or approval behavior
- rewrite score floor or benchmark baselines

This is a stop/go document.

If a change fails the guardrails below, it should not merge until the risk is
removed or a deliberate exception is recorded.

## Lead To Preserve

The current lead or tie position comes from these dimensions:

- `workflow-rigor` — tied
- `packaging-install` — tied
- `advanced-set-breadth` — tied
- `learning-maturity` — ahead
- `governance` — ahead
- `core-simplicity` — ahead

These are the dimensions the rewrite must not degrade.

The dimensions that should improve during the rewrite are:

- `delivery-maturity`
- `token-efficiency`
- `adapter-portability`
- `memory-reliability`
- `flexibility`

The rewrite should also move the overall golden benchmark score upward over
time, not only protect individual dimension ties.

The locked floor for the next rewrite slice lives in:

- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)

The product-review role definitions live in:

- [product-review-roles.md](./product-review-roles.md)

The standing install/config/usage dogfood gate lives in:

- [dogfood-gates.md](./dogfood-gates.md)

The standing scorecard sources for major product decisions and push readiness live in:

- [competitive-kpis.yaml](./competitive-kpis.yaml)
- [golden-competitive-benchmark.yaml](./golden-competitive-benchmark.yaml)
- [rewrite-score-baseline.yaml](./rewrite-score-baseline.yaml)

## Scorecard Decision Gate

Use the scorecard gate before any major decision that changes:

- public command surface
- conversation structure
- current-work behavior
- artifact shelf behavior
- route/model/effort policy
- install/setup flow
- multi-project behavior
- benchmark scenarios, competitor framing, or KPI weighting

Major decisions must answer these questions explicitly:

- which scorecard dimensions are expected to improve?
- which scorecard dimensions are at risk?
- does the golden benchmark still preserve or improve the lead?
- does the rewrite score floor still hold after this choice?

If those answers are not written down, the decision is not ready.

## Push Gate

Before pushing a meaningful rewrite slice, run the scorecard gate.

Minimum push evidence:

- current golden benchmark result
- current rewrite score floor check
- current competitor lead margin check
- dogfood gate result for the affected personas when install/setup/usage changed

For major slices, pushing is blocked unless the work either:

- preserves the current scorecard lead and floor, or
- records an explicit temporary exception with recovery criteria

Do not rely on intuition alone for “this still feels better.”
Pushes that change product shape must clear the scorecard gate.

## Allowed Simplifications

These are good simplifications. They improve UX without weakening the system.

- Replace public builder jargon with plain language.
- Reduce front-door command count while keeping advanced surfaces reachable.
- Hide advanced commands from beginner docs while preserving aliases and scriptability.
- Replace file stems with human-readable output names.
- Replace path-driven normal flow with remembered current work, as long as the current work is always visible and controllable.
- Collapse multi-step beginner flows into one launcher flow if the underlying semantics still execute.
- Move the public runtime from Python to a compiled binary while preserving work contracts.
- Prefer deterministic local reads over expensive orchestration for `check` and `open`.
- Show approvals, evidence gaps, and uncertainty in friendlier language instead of removing them.
- Move capability growth into packs, adapters, and views instead of growing the kernel.
- Add human-facing summaries over existing canonical artifacts instead of inventing a second truth model.
- Keep the public face small while leaving operator tooling intact behind it.
- Remove or demote any surface that attracts the same critique from more than one review role unless there is explicit user-outcome proof to keep it.

## Lead-Damaging Simplifications

These are bad simplifications. They make the product easier to explain by
making it weaker.

- Changing `work-unit.yaml`, canonical artifact semantics, or export contracts during the runtime rewrite.
- Deleting evidence, approval, provenance, or rollback concepts from the core flow.
- Replacing deterministic state or readiness logic with ungrounded model judgment.
- Removing advanced packs, adapters, or routines just because they are not part of the front door.
- Removing install receipts, provenance, smoke checks, or target verification to make install look simpler.
- Deleting scriptable commands and leaving only an interactive launcher.
- Making current work magical or implicit without showing what Jini selected.
- Flattening domain differences into new kernel concepts instead of keeping them in packs.
- Replacing canonical artifacts with ephemeral chat-only output.
- Deleting review, backtest, or governed learning surfaces that preserve the `learning-maturity` lead.
- Simplifying by narrowing the product to only one or two workflows and losing breadth.
- Merging runtime rewrite and work-format rewrite into the same phase.

## Mechanical Merge Gate

Every rewrite change should pass this checklist.

If any answer is `no`, stop and fix it before merge.

### A. Workflow Rigor

- [ ] Does this change preserve canonical work-unit state semantics?
- [ ] Does this change preserve artifact readiness logic rather than hand-waving it?
- [ ] Can a real work unit still move from idea to verification without manual YAML surgery?
- [ ] Does the change make the flow easier to use without making the lifecycle less legible?

### B. Packaging And Install

- [ ] Does this change preserve install receipts, provenance, and target visibility?
- [ ] Does this change keep installation auditable and reversible?
- [ ] Does this change improve or preserve first-run trust rather than just hiding complexity?
- [ ] Does this change avoid introducing a fake binary that still behaves like a fragile script wrapper?

### C. Advanced Breadth

- [ ] Does this change preserve the existing pack and adapter breadth?
- [ ] If it hides advanced capability, does it remain reachable and testable?
- [ ] Does this change avoid shrinking Jini into a single-workflow tool?
- [ ] Does this change keep breadth outside the kernel where possible?

### D. Learning Maturity

- [ ] Does this change preserve governed learning, review, experiment, and rollback surfaces?
- [ ] Does this change avoid adding ungoverned adaptive behavior?
- [ ] Does this change keep enough telemetry or evidence for bounded future learning?

### E. Governance

- [ ] Does this change preserve approval and evidence semantics?
- [ ] Does this change keep proof visible before important transitions?
- [ ] Does this change preserve replayability and provenance across publishes or adapters?
- [ ] Does this change reduce governance ceremony without removing governance truth?

### F. Core Simplicity

- [ ] Does this change reduce front-door complexity without adding kernel complexity?
- [ ] Does this change avoid introducing new universal concepts when a pack or adapter could carry the behavior?
- [ ] Does this change keep the kernel smaller or equal in conceptual weight?

### G. User Trust

- [ ] Does Jini still show what it is using, what it is doing, and what is still missing?
- [ ] Is current work visible before Jini acts on it?
- [ ] Are uncertainty and reversibility visible where they matter?
- [ ] Does the first useful output appear before the system asks the user to learn internal concepts?
- [ ] Does the `I am not sure` path reduce confusion without forcing problem classification?
- [ ] If `I am not sure` cannot be classified, does Jini still return a useful first object instead of a clarification dead end?

### H. Competitor Must-Haves

- [ ] Does the first minute deliver a useful result before state explanation, matching Claude-style immediacy?
- [ ] Does the flow show ready work and missing work clearly enough to match Kiro-style progression?
- [ ] Does remembered work feel visible, controllable, and recoverable enough to match Hermes-style continuity?
- [ ] Does Jini expose source, assumptions, missing proof, and provenance enough to match AgentField-style inspectability?
- [ ] Does Jini preserve its own honest-closure lead by keeping missing truth visible after useful output exists?
- [ ] Does at least one real post-result continuation path exist in the replacement-critical slice?
- [ ] Do both replacement-critical flows have real `Continue` and `Missing` actions before release?

### I. Cross-Role Consensus

- [ ] Did the competitive analyst approve Jini's stage-by-stage position against Claude, Kiro, Hermes, AgentField, and natural-intake tools?
- [ ] Did the UX researcher approve the first minute for tired or low-confidence users?
- [ ] Did the UX designer approve the exact screen order, copy, and output shelf labels?
- [ ] Did the program manager approve scope, sequencing, parity evidence, and lead preservation?
- [ ] Are all role objections either resolved in the design or recorded as explicit non-goals for this slice?
- [ ] If two or more roles raised the same friction, was the default action remove, demote, or simplify it unless user-outcome proof justified keeping it?
- [ ] Is launcher-created work blocked from shipping or expanding without shared generation or golden parity fixtures?
- [ ] Do old/new parity fixtures exist before public cutover?

### I2. Dogfood Personas

- [ ] Did the low-literacy first-time user pass the install and first-run gate?
- [ ] Did the AWS Bedrock user pass the strict-route trust gate?
- [ ] Did the Azure enterprise user pass the Azure-only confidence gate?
- [ ] Did the pragmatic “just make it work” user pass the no-jargon value gate?
- [ ] Did the Claude, Codex, ChatGPT, and Gemini personas pass the platform-path gate?
- [ ] Did the power user, software engineer, hardcore developer, AI engineer, QA tester, and architect personas pass the expert-operator gate?
- [ ] Did the AI PM and software VP personas pass the product-clarity and cost-governance gate?
- [ ] Did the college-student, high-school-student, household-manager, realtor, and travel-advisor personas pass the beginner/domain clarity gate?
- [ ] If any persona failed, was the failure fixed or explicitly recorded as a non-goal before merge?

### J. Rewrite Momentum

- [ ] Does the current overall golden benchmark score clear the locked rewrite score floor?
- [ ] Does the current overall lead margin still clear the locked competitor margin floor?
- [ ] If this refactor improved the overall score, did you update `rewrite-score-baseline.yaml` so the next refactor must beat the new floor?

## Fast Heuristics

Use these heuristics when a change looks attractive but feels risky.

### Safe

- “This hides complexity but keeps the same truth underneath.”
- “This renames a system term in user language.”
- “This moves capability out of the beginner path without deleting it.”
- “This speeds up the common read path without reducing rigor.”
- “This makes outputs more usable without changing their canonical source.”

### Unsafe

- “Users do not need to know about approvals/evidence anymore.”
- “We can remove this adapter/publish/review path because beginners do not use it.”
- “We can make readiness a model guess instead of checking artifacts.”
- “We can simplify by collapsing several states into one vague status.”
- “We can hide current work selection and trust the tool to infer it.”
- “We can ship only the launcher now and rebuild scriptability later.”

## Rewrite Order Guardrail

The rewrite should happen in this order:

1. freeze contracts
2. port `check`
3. port `open`
4. ship binary shell
5. port `run`
6. port install/admin surfaces
7. port adapters and advanced surfaces
8. cut over public runtime

Do not skip ahead by:

- rewriting work formats early
- deleting Python before parity exists
- widening the public feature set before the fast read path is solid

## Exception Rule

If a change fails a guardrail but still seems necessary, document:

- which dimension takes the hit
- why the hit is temporary
- how parity will be restored
- what test or metric will prove recovery

No silent tradeoffs.

## One-Line Rule

Simplify the experience.

Do not simplify away the truth that gives Jini its lead.
