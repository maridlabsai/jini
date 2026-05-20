# Friction Reduction Gate

## Gate Purpose

This gate prevents Jini from becoming harder to use than Codex, ChatGPT, or
Claude while trying to be more capable. Any feature that adds surface area must
show how it reduces user effort, preserves context, improves trust, lowers cost,
or creates a richer artifact experience.

## Gate Categories

### 1. First-Minute Simplicity

Pass criteria:

- `jini` starts without stale default work output.
- `jini help` or an explicit help request shows examples and setup guidance.
- A new user can type natural language before learning packs, routes, or
  provider names.
- Familiar tactical commands such as `/help`, `/status`, `/doctor`, `/model`,
  `/init`, `/memory`, `/permissions`, and `/cost` are safe aliases and never
  create work artifacts by themselves.
- The first useful result appears before status or model metadata.

Regression inputs:

- `hello`
- empty shell start
- `help`
- `/status`
- `/doctor`
- `/memory`
- `7 day paris trip`
- `turn these notes into a follow-up`

### 2. Natural Intent Handling

Pass criteria:

- Greeting-only input stays conversational.
- Ambiguous work asks minimal clarifying questions.
- Already-scoped work drafts immediately.
- Complex multi-artifact work creates a durable artifact plan.
- Core behavior is driven by semantic envelope fields, not use-case-specific
  hard-coded response blocks.

Regression inputs:

- greeting
- underspecified travel request
- scoped travel request
- raw meeting notes
- launch-plan critique
- vendor comparison

### 3. Continue-Anywhere Work State

Pass criteria:

- A work item has one stable identity across CLI, desktop, mobile, and hosted
  commercial sync.
- Resume does not require a path when one active item is obvious.
- When multiple active items exist, Jini asks the smallest useful chooser.
- The resume view carries goal, ready artifacts, blockers, last action, next
  action, pending approvals, and route/cost posture.

Required markers:

- `continue-anywhere`
- `single-work-identity`
- `minimal-active-work-chooser`
- `cross-surface-resume`

### 4. Artifact Escalation

Pass criteria:

- Long, reusable, editable, or multi-part outputs become artifacts.
- Terminal output summarizes the artifact instead of dumping all detail.
- Artifact metadata includes type, title, source context, readiness, missing
  inputs, safe actions, and smart links.
- Native/commercial surfaces can render the same artifact envelope without
  changing the core work state.

Required markers:

- `artifact-escalation`
- `terminal-summary-not-dump`
- `surface-independent-artifact-envelope`
- `smart-links`

### 5. Setup Doctor And Self-Healing

Pass criteria:

- PATH, runtime, provider, OS, accelerator, token, and subscription issues map
  to doctor checks.
- Doctor output explains what failed, why it matters, and the next command or
  user action.
- Provider setup never leaks secret values.
- A failed strict route offers local/free fallback when safe.

Required markers:

- `setup-doctor`
- `path-self-healing`
- `provider-secret-redaction`
- `local-or-cheap-fallback`

### 6. Cost And Route Minimalism

Pass criteria:

- Default routing chooses least-expense capable route.
- Premium routes are justified by quality, safety, context size, or external
  action risk.
- The user can inspect model/API/provider choice without seeing route metadata
  before useful content.
- Subscription exhaustion degrades to local/free capability rather than a dead
  end.

Required markers:

- `best-productivity-least-expense`
- `route-explain-on-demand`
- `subscription-aware-fallback`
- `premium-route-justification`

### 7. Trust Without Ceremony

Pass criteria:

- Risky write, publish, payment, booking, and external-action paths require
  confirmation.
- Low-risk continuation does not repeatedly ask for the same permission.
- Trust grants are scoped, inspectable, expirable, and reversible.
- Jini shows what is safe to review before sharing.

Required markers:

- `visible-trust`
- `scoped-approval-memory`
- `review-before-share`
- `no-repeated-permission-ceremony`

### 8. Prompt Adoption Parity

Pass criteria:

- The first-turn experience feels like natural task intake, not workflow setup.
- Provider, route, model, pack, or profile labels stay hidden until the user
  asks or cost, safety, or capability constraints make them relevant.
- Greeting, thanks, tactical commands, and lightweight social inputs never
  create work or expose internal artifact taxonomy.
- Internal product terms such as `First Useful Pass` are not required to
  understand or navigate default user flows.
- Clarification is driven by shared scope logic and asks only the smallest
  useful set of high-yield questions.
- Core output shaping does not depend on use-case-specific hard coding in the
  primary intake path.

Required markers:

- `natural-task-intake`
- `no-early-provider-leak`
- `no-internal-taxonomy-reliance`
- `generic-scope-planner`
- `no-core-use-case-hard-coding`

## Reject Conditions

Reject a change if any of these are true:

- It adds a required command before the first useful result.
- It treats a tactical command, greeting, acknowledgement, or help request as a
  new work artifact.
- It prints route/model/provider/state metadata before useful content without a
  safety reason.
- It adds a use-case-specific response mold to core rendering.
- It creates a detailed terminal-only answer when an artifact envelope is
  required.
- It makes resume depend on remembering a file path when a current work item is
  known.
- It hides subscription or route fallback state until after failure.
- It adds an approval prompt that cannot be remembered, scoped, or explained.
- It teaches provider, route, pack, or command taxonomy before the user has
  stated the work.
- It relies on product-internal terms such as `First Useful Pass` in the normal
  first-minute journey.
- It adds or preserves use-case-specific hard coding in core intake when the
  same behavior should be driven by the shared work envelope.

## Required Regression Inputs

- `jini`
- `jini help`
- `jini /status`
- `jini /doctor`
- `jini /memory`
- `hello`
- `thanks`
- `what can you do?`
- `7 day paris trip`
- `Plan a 5 day Lisbon trip for two adults in October, food/design focused,
  moderate budget, no rental car.`
- `Plan a 5 day Lisbon trip for two adults in October, food/design focused,
  moderate budget, no rental car, Alfama stay, no museums, Sintra optional.`
- `turn these notes into a follow-up`
- `rewrite this note to be sendable`
- `critique this launch plan`
- `continue this`
- `continue`
- provider missing-token setup
- subscription exhausted with local runtime available

## Gate Summary

The gate passes only when Jini keeps the first minute lighter than a workflow
tool, the resume path clearer than chat history, the artifact path richer than a
terminal dump, and the route path cheaper by default than premium-model habit.
It also requires prompt behavior that feels native to Claude, ChatGPT, and
Codex users without making them learn Jini-specific taxonomy first.
