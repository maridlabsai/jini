# Jini Launcher Intake Design

Status: review
Updated: 2026-05-14

## Goal

Make `jini` the real first-run product.

A new user should be able to:

1. run `jini`
2. paste what they have or choose a plain-language job
3. give messy context without formatting it for the tool
4. see the first useful result
5. then see what is missing and what is safe to do next

This removes the biggest remaining gap against Claude and Hermes:
first-run payoff.

## Scope

This slice covers the first-run launcher and initial work creation.

Replacement-critical flows:

- meeting follow-up
- spec readiness

Demo flow:

- trip planning

It does not yet cover:

- full conversational continuation
- multi-step in-session commands
- live streaming run UI
- advanced operator flows

## User Flow

### No current work

When the user runs `jini` with no current work:

1. invite natural paste-first input
2. show the job list for users who want choices
3. ask for the situation in plain language
4. ask only blocking questions
5. create a starter work directory under `.jini/work/...`
6. remember that work in `.jini/current-work.json`
7. show the first useful object
8. show the calm summary after the object or when the user asks what is missing

Exact first prompt:

```text
What do you need help finishing?

Jini shell
Paste messy notes, or type the outcome you want.

Good inputs:
- Turn meeting notes into something I can send
- Check whether a plan is ready to hand off
- Plan this first
- I am not sure

Nothing will be sent yet.
```

Exact source prompt:

```text
Paste what you have. A rough version is fine.
```

If the user chooses `I am not sure`, Jini says:

```text
Paste what you have. A rough version is fine.
I will help figure out whether this is follow-up, a plan check, or something else.
Nothing will be sent yet.
```

If the input does not clearly match meeting follow-up or plan/spec readiness,
Jini still shows a useful first object:

- `First Useful Pass`

It must include what the user appears to be trying to finish, what can be used
now, what Jini needs next, and what is safe because nothing has been sent.

Example inputs:

- meeting: `Weekly product review. Need owners, due dates, and open questions.`
- plan/spec: `Notifications PRD draft. Need to know if it is safe to hand off.`
- trip: `7-day Paris trip in June for two people. Mid-range budget.`

Jini must not ask for output size before the first result. It defaults to a
quick useful draft and offers `Make it fuller` after the result exists.

For meeting follow-up and plan/spec readiness, Phase 1 must include real `Keep
going` and `See what is still missing` actions after the first result. They must
not be placeholder menu items.

### Current work exists

When current work exists:

1. show that work is already in progress
2. show what is ready, what is missing, and the next step
3. offer three real choices:
   - continue current work
   - open what is ready
   - start something new

The three choices must not be shown as selectable actions until they work.

Recommended copy:

```text
You already have work in progress.

Current work
Research to PRD handoff

Ready now
- Build-Readiness Check
- Handoff Brief

Still missing
- Product approval
- Rollback note

Next step
Open Build-Readiness Check

Jini shell
What do you want to do?

- Keep going
- Open ready work
- See what is still missing
- Plan this first
- Start something else
```

## Output Rules

The first opened object must be usable, not structural.

### Meeting follow-up

Starter output:

- `Sendable Follow-up`
- `Owners and Due Points`
- `Decisions Made`
- `Open Questions`

The follow-up must contain:

- concrete decisions
- owners
- due points
- open questions

### Spec readiness

Starter output:

- `Build-Readiness Check`
- `Handoff Brief`
- `Missing Pieces Before Build`
- `Risk List`

The summary must contain:

- build-readiness answer
- missing approval or rollback gaps
- first implementation slice

### Trip planning

Starter output:

- `Trip Plan`
- `Budget Sketch`
- `Travel Logistics`
- `Still To Book`

The itinerary must contain:

- day-by-day draft
- budget sketch
- logistics
- contingencies

## Runtime Design

The Go runtime owns the launcher behavior.

### Entry

- `cmd/jini/main.go` calls `app.RunInteractive(...)`

### Core runtime

- `RunInteractive(args, stdin, stdout, stderr)` is now the public runtime entry
- `Run(...)` remains as a thin wrapper for existing tests and callers

### First-run intake

- `runLauncher(...)` checks for remembered work
- if none exists and `stdin` is available, it calls `runNewWorkIntake(...)`
- `runNewWorkIntake(...)`:
  - renders the job menu
  - reads choice
  - renders job-specific example input as inline helper copy under the source prompt
  - reads source line
  - asks only blocking follow-up questions
  - creates starter work
  - writes current-work state
  - shows the first useful object
  - then loads and renders the calm summary on request

### Work creation

Starter work is created locally under:

- `.jini/work/<pack>-<slug>/`

It writes:

- `work-unit.yaml`
- `views/*.md`
- `artifacts/*.yaml`

This slice intentionally uses a small local starter writer instead of calling
the old Python compatibility layer. The point is to prove the product flow in
the new runtime first.

## Why This Design

### Why not keep the old example/setup path?

Because it preserves the old product smell:

- framework-first
- seeded-path-first
- not truly comparable to Claude or Hermes

### Why not port the full Python engine first?

Because that would delay the user-facing win.

The higher-value move is:

- make first-run motion real
- preserve existing contracts
- improve benchmark score

### Why local starter writers?

Because this slice is about the product loop, not full engine parity.

The launcher needs to:

- ask little
- create something fast
- show something useful

That is more important right now than deep canonical completeness for every
possible flow.

This is temporary. Local starter writers must not become a second truth model.
The next replacement slice must add parity fixtures or move generation behind a
shared service before the launcher expands.

No launcher-created work may ship or expand beyond the two flagship flows unless
one of these is true:

- it uses shared generation
- golden parity fixtures prove `work-unit.yaml`, readiness logic, human output
  labels, and `check` / `open` behavior match canonical work

Golden parity fixtures for old and new work must exist before cutover.

## Consensus Review Roles

The launcher must pass four product-review roles before it is treated as the
approved replacement shape:

- Competitive Analyst: Jini must beat competitors stage by stage, not just in
  positioning.
- UX Researcher: low-confidence users must understand the first minute without
  translating product categories.
- UX Designer: the useful result must be centered before summary, shelf, or
  system detail.
- Program Manager: the slice must remain narrow, parity-tested, and score-safe.

The detailed role definitions live in:

- [product-review-roles.md](./product-review-roles.md)

The consensus PRD and implementation plan live in:

- [product-consensus-prd-and-plan.md](./product-consensus-prd-and-plan.md)

## Must-Have Framework Merits

These are required behaviors, not optional inspiration.

### Claude Code: Immediate Payoff

- the first minute should feel like asking for work and getting useful work back
- no setup vocabulary before the result
- no `short` / `full` decision before the result
- the useful result appears before the summary screen
- post-result actions are real and immediately usable

### Kiro: Visible Progression

- the work stays visible while it moves
- the user can see what is ready and what is still missing
- output names are human objects, not files or internal stems
- deeper task/proof detail exists without crowding the first screen
- visible progression exists as work movement, not only as a post-hoc summary

### Hermes: Continuity

- returning to `jini` shows the current work before acting
- continue, open, and start-new choices must be real before cutover
- remembered work must never feel hidden or magical
- stale remembered work is explained and recoverable
- park, switch, and stale-work behavior are specified before public cutover

### AgentField: Inspectability

- Jini shows what source it used
- Jini shows assumptions that affect trust
- Jini shows missing approval, evidence, ownership, or blockers
- provenance and artifacts remain openable on demand
- inspectability is a visible trust surface, not only an advanced file path

### Jini: Honest Closure

- a draft can be useful without being final
- missing truth stays visible in plain language
- uncertainty is shown when it affects safety
- approvals and evidence are simplified, not removed

## Guardrails

This design must keep:

- remembered-work visibility
- plain-language summary
- no path-driven normal flow
- human-readable output names
- useful result before summary on first run
- competitor must-haves above
- consensus role approval
- shared generation or golden parity evidence before shipping launcher-created work
- scorecard lead
- rewrite score ratchet

It must not:

- rewrite canonical pack formats wholesale
- remove advanced surfaces
- hide current work silently
- show placeholder choices that do not work
- treat local starter writers as permanent canonical generation
- replace artifact truth with pure chat output

## Acceptance

This slice is done when:

1. `jini` can create meeting/spec starter work from stdin
2. the new work is remembered automatically
3. the first visible result is the useful object
4. the summary screen is available after the result
5. current-work continuation choices are either fully implemented or not shown
6. the first object is usable
7. launcher-created work passes golden parity checks or uses shared generation
8. the competitive analyst, UX researcher, UX designer, and program manager agree the shape is approvable
9. the overall benchmark score increases
10. the rewrite score baseline is ratcheted upward

## Current Result

This slice currently achieved:

- real first-run launcher in Go
- meeting/spec/trip starter work creation
- remembered-work persistence
- usable starter objects
- benchmark score increase from `8.89` to `8.92`

## Open Questions

1. Should `jini` support true in-session commands next, or should it remain a
single-turn launcher plus scriptable subcommands for now?
2. Should meeting/spec/trip starter writers stay local in Go until the full
runtime is ported, or should they be replaced quickly with deeper canonical
generation?
3. Should the next slice prioritize:
   - better in-session continuation
   - stronger meeting/spec/trip object quality
   - live run progress
4. Should trip planning stay in the launcher while meeting/spec replacement
quality is still being hardened, or move back to demo status until parity is
proven?
