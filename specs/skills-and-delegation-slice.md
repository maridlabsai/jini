# Skills And Delegation Slice

Updated: 2026-05-21

## Purpose

This document defines the next implementation slice for skills and delegated
work in Jini.

The goal is not to turn Jini into a visible multi-agent control plane.

The goal is to let Codex and Claude Code users bring familiar specialist-helper
patterns into Jini without adding orchestration theory, product-shaped jargon,
or a second interaction model.

## Product Rule

Phase 1 must be explicit, file-backed, and low-ceremony.

Jini should ship:

- one discovery command: `skills`
- one execution command: `delegate`
- project-scoped skills
- user-scoped skills
- normal Jini artifacts as outputs

Jini should not yet ship:

- automatic delegation by default
- visible subagent trees
- agent-role theater
- prompt-only ephemeral skills
- workflow branches that require users to understand orchestration internals

## UX Contract

### 1. Normal Users Do Not Need Skills

The default Jini path remains:

- `jini`
- paste the work
- get a useful result

Skills are optional.

### 2. Explicit Delegation Uses One Verb

The explicit command is:

- `delegate`

Examples:

- `delegate reviewer`
- `delegate debugger`
- `delegate research`

The user should not need `use-skill`, `run-agent`, `subagent`, or similar
product-shaped commands.

### 3. Discovery Uses One Noun

The discovery command is:

- `skills`

Examples:

- `skills`
- `skills reviewer`

### 4. Results Look Like Normal Jini Work

A delegated result must come back as normal Jini output:

- a named artifact
- a short result summary
- what changed
- what still needs attention

The result must not dump internal orchestration logs by default.

### 5. Delegation Must Stay Attached To Current Work

Delegated work is not a second thread.

It is attached to the current work object and must preserve:

- current work id
- current artifact focus
- continuation
- history

### 6. Hidden Mechanics, Visible Trust

Jini may hide orchestration mechanics, but it must still show:

- which skill ran
- what it was asked to do
- what it produced
- whether the result is ready, partial, or blocked

## File Format

## Skill Discovery Roots

Jini should load skills from two roots:

- project: `.jini/skills/`
- user: `~/.jini/skills/`

Project scope wins on id collision.

### Layout

Each skill lives in its own directory:

```text
.jini/skills/reviewer/
  skill.yaml
  prompt.md
```

The same layout applies to `~/.jini/skills/`.

### skill.yaml

Required fields:

- `schema_version`
- `skill_id`
- `label`
- `purpose`
- `when_to_use`
- `allowed_tools`
- `input_contract`
- `output_contract`
- `prompt_file`

Optional fields:

- `enabled`
- `scope`
- `tags`
- `preferred_artifact_types`
- `preferred_intents`
- `max_cost_class`
- `requires_approval`

### Minimal Example

```yaml
schema_version: "0.1.0"
skill_id: reviewer
label: Reviewer
purpose: Review a draft artifact for correctness, risk, and missing follow-through.
when_to_use:
  - The user asks for critique or review.
  - The current artifact is close to shareable but may hide risks.
allowed_tools:
  - read-artifacts
  - write-artifacts
  - status
input_contract:
  requires_current_work: true
  accepts_artifact_focus: true
  accepts_freeform_instruction: true
output_contract:
  primary_artifact_type: review
  returns_summary: true
  returns_findings: true
prompt_file: prompt.md
enabled: true
scope: project
tags:
  - review
  - critique
preferred_artifact_types:
  - followup
  - prd
  - checklist
preferred_intents:
  - verify
requires_approval: false
```

### prompt.md

`prompt.md` contains the skill-specific working instructions.

It should be plain text or markdown, small, and local to the skill.

It must not redefine Jini’s global user contract.

## Runtime Behavior

### 1. `skills`

`skills` returns discovered skills in canonical order:

1. project skills
2. user skills

The output must show:

- `skill_id`
- `label`
- `scope`
- `purpose`
- whether it is enabled

If an id is shadowed by project scope, only the winning entry is shown in the
default list.

### 2. `delegate <skill-id>`

`delegate` must:

1. resolve the skill id
2. require current work unless explicitly supported otherwise
3. capture current artifact focus
4. build a bounded request
5. run the delegated job
6. store the result under the same work object
7. return a normal Jini result surface

### 3. Delegation Storage

Delegated runs should be recorded under the current work directory:

```text
work/<work-id>/delegations/<timestamp>-<skill-id>/
  request.json
  result.json
  summary.md
```

Required `request.json` fields:

- `delegation_id`
- `work_unit_id`
- `skill_id`
- `scope`
- `artifact_focus`
- `instruction`
- `created_at`

Required `result.json` fields:

- `delegation_id`
- `work_unit_id`
- `skill_id`
- `status`
- `created_artifacts`
- `updated_artifacts`
- `summary_path`
- `completed_at`

### 4. Current Work State

`thread-state.json` should track the latest delegation on the active work:

- `active_delegation_id`
- `active_skill_id`
- `active_delegation_status`

This should behave like other current-work continuation state, not as a separate
subsystem.

### 5. First-Phase Routing Rule

Phase 1 delegation is explicit only.

Jini may suggest delegation, but it should not auto-run a skill unless:

- the user explicitly invoked `delegate`
- or a future guarded rollout adds auto-delegation behind a separate gate

### 6. Error Behavior

Canonical failures:

- unknown skill id
- disabled skill
- invalid skill file
- current work required but missing
- required artifact focus missing

Errors should be plain and actionable:

- `Unknown skill 'reviewerx'.`
- `Skill 'reviewer' is disabled.`
- `Skill 'reviewer' is invalid: missing output_contract.`
- `Current work is required for delegate reviewer.`

## Output Contract

Successful `delegate` output should include:

- skill label
- what it produced
- ready or blocked state
- `Open`
- `Status`
- `Continue`

If findings are returned, they should be first.

If an artifact was created, it should become the focused artifact.

## Shipping Sequence

### Phase 1

- file-backed skill discovery
- `skills`
- `delegate`
- work-attached delegation records
- explicit-only delegation

### Phase 2

- `delegate` with artifact targeting
- skill filtering by current artifact family
- approval gates for high-cost or write-heavy skills

### Phase 3

- guarded auto-delegation suggestions
- confidence thresholds
- delegated-result comparison and learning hooks

## Tests

## Discovery Tests

- project skill is discovered
- user skill is discovered
- project scope overrides user scope on id collision
- disabled skills do not appear in default output
- malformed skill files fail cleanly

## Command Tests

- `skills` lists canonical fields
- `skills reviewer` filters correctly
- `delegate reviewer` fails without current work
- `delegate reviewer` succeeds with current work
- unknown skill id returns a clean error

## State Tests

- delegated run writes `request.json`
- delegated run writes `result.json`
- delegated run writes `summary.md`
- current thread state records `active_delegation_id`
- focused artifact updates when a delegated artifact is created

## UX Tests

- delegated results render as normal Jini output
- no visible `subagent` or `agent` jargon in default output
- `delegate` does not open a second work thread
- `resume` returns to the delegated artifact if it is the active focus

## Safety Tests

- disabled skill cannot run
- invalid skill schema cannot run
- skill prompt file must exist
- project skill shadowing user skill is deterministic

## Non-Goals

This slice does not include:

- cloud multi-agent orchestration
- visible task trees
- background worker fleets
- auto delegation on first run
- skill marketplaces

## Success Criteria

This slice is complete when:

- Jini supports file-backed project and user skills
- `skills` and `delegate` work end-to-end
- delegated work stays attached to the same work object
- results feel like normal Jini artifacts
- Codex and Claude Code users can map their existing “specialist helper” habit
  into Jini without learning a new orchestration model
