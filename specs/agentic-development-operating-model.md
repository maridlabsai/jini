# Agentic Development Operating Model

Updated: 2026-06-09

## Purpose

This document defines how Jini development work is divided across sub-agents.
It is an internal engineering operating model, not a public product surface.

Jini's user-facing contract stays simple:

- run `jini`
- paste or describe the work
- get a useful result
- see what changed, what is missing, and what is safe to do next

Sub-agent orchestration must improve engineering quality without leaking agent
trees, delegation vocabulary, or orchestration logs into the default CLI.

## Precedence

The product-streamline redline wins over delegation cleverness.

Reject any development process or implementation that:

- adds public command grammar to explain internal agent work
- makes users choose agents before getting value
- exposes agent trees by default
- weakens first-minute CLI transcripts
- makes a simple question or direct file edit enter artifact, skill, or
  delegation machinery before it has to
- cannot map changes back to PRD outcome, architecture boundary, implementation
  contract, and executable gate

## Core Rule

Use divide-and-conquer sub-agents by default for non-trivial Jini engineering
cuts.

The lead agent remains accountable for the final answer, final diff, gate
evidence, and product fit. Sub-agents supply bounded analysis, plans, designs,
patches, tests, or reviews. They do not replace integration judgment.

This rule is mandatory for material cuts that change product behavior,
architecture boundaries, runtime contracts, tests, gates, release claims, or
protected docs.

## Spec Trace

Every non-trivial cut must name its trace before completion:

- PRD outcome:
  [number-one-platform-prd.md](./number-one-platform-prd.md)
- HLD boundary:
  [number-one-platform-hld.md](./number-one-platform-hld.md)
- LLD contract:
  [number-one-platform-lld.md](./number-one-platform-lld.md)
- streamline redline:
  [product-streamline-redline.md](./product-streamline-redline.md)
- execution class:
  [execution-routing-policy.md](./execution-routing-policy.md)
- implementation slice:
  [number-one-development-plan.md](./number-one-development-plan.md)

## Required Sub-Agent Coverage

For material work, split these concerns instead of letting one agent reason
through everything alone:

- Reasoning: map the problem, constraints, invariants, and likely failure modes.
- Planning: turn the goal into ordered slices with explicit scope and gates.
- Design: define contracts, boundaries, data flow, and user-facing behavior.
- Coding: implement bounded changes inside an assigned write set.
- Testing: create or run focused regression checks and required gates.
- Code review: independently inspect the diff for bugs, regressions, missing
  tests, product drift, and unsafe assumptions.

The lead agent may combine concerns only when the work is small enough that the
split would add more coordination cost than risk reduction.

## When To Spawn

Spawn sub-agents when any condition is true:

- the change touches more than one architecture boundary
- the task requires both product judgment and code changes
- the task changes CLI behavior, transcripts, routing, state, artifacts,
  gates, security, packaging, or release claims
- the work has unclear blast radius
- the implementation can be divided into independent files or modules
- a regression would be expensive to notice late
- review requires a different mindset than implementation
- the lead agent has made two failed attempts or is guessing

Default split for medium work:

- one planner or design reviewer before edits
- one implementer per disjoint write set
- one tester focused on proof
- one reviewer who did not write the code

## When Not To Spawn

Do not spawn sub-agents for work that is simpler than the coordination:

- single-line or mechanical edits with obvious blast radius
- formatting-only changes
- reading one file to answer a narrow question
- running one known command and reporting its output
- reverting or resolving a user-directed local-only edit
- tasks where the user explicitly asks for solo, synchronous work

Do not spawn when delegation would violate the workspace contract, create
overlapping write sets, hide uncertainty, or slow a production fix without
reducing risk.

## Two-Level Delegation

Delegation is capped at two levels.

Level 1 is the lead agent delegating bounded workstreams:

- research or reasoning probe
- plan critique
- design sketch
- implementation slice
- test and gate runner
- independent code review

Level 2 is a workstream agent delegating one narrower probe only when it needs
evidence it cannot gather safely itself. Level 2 work must be read-only unless
the lead agent explicitly assigns a write set.

No Level 3 delegation is allowed. Nested teams create theater and make evidence
harder to audit.

Every delegated task must state:

- objective
- allowed files or read-only scope
- expected output
- required evidence
- known redlines
- deadline or stopping condition

## Disjoint Write Sets

The lead agent must assign write ownership before coding begins.

Rules:

- one owner per writable file or glob
- no two coding agents may edit the same file concurrently
- generated files count as writes and need an owner
- integration files are owned by the lead agent unless explicitly reassigned
- shared contracts, public CLI copy, gate runners, and release docs require
  serialized edits
- if a write set overlaps, stop parallel work and either split the file,
  serialize the edits, or make the lead agent the sole integrator

Sub-agents may read outside their write set for context, but they must not edit
outside it. If a needed edit falls outside the assigned set, they report the
need instead of making the change.

## Integration Rules

The lead agent integrates all sub-agent outputs.

Before accepting a sub-agent patch or recommendation, the lead agent checks:

- the output satisfies the delegated objective
- changed files are inside the assigned write set
- assumptions are explicit
- failures and skipped checks are reported
- the change preserves Jini's `jini`-first CLI shape
- the change does not weaken the product-streamline redline
- evidence is attached to claims

The lead agent resolves conflicts by priority:

1. user instruction and repository ownership boundaries
2. safety, security, and data-loss prevention
3. product-streamline redline and first-minute transcript quality
4. correctness and regression coverage
5. simplicity and maintainability
6. implementation convenience

## Review Rules

Review must be independent from implementation for material work.

Reviewers inspect:

- behavioral regressions
- missing or brittle tests
- boundary violations
- write-set violations
- hidden public command-surface expansion
- verbose or magical CLI output
- route, state, artifact, or permission ambiguity
- evidence gaps between claims and executed checks

Two or more independent reviewers raising the same product-friction issue means
the default action is to remove, demote, or simplify the offending surface
unless explicit evidence proves it improves user outcome.

## Gates And Evidence

Sub-agent output is not evidence by itself. Evidence is a named file, command
result, transcript, fixture, test, or explicit unresolved risk.

Minimum evidence for any material change:

- problem statement and intended user or engineering outcome
- delegation map showing spawned agents and their write sets
- changed-path list
- focused tests or inspections relevant to the slice
- required gate results, or explicit reason a gate was not run
- product-streamline check for CLI-facing changes
- independent review findings or "no findings" statement
- unresolved risks and follow-up owner, if any

Use focused checks during iteration, then run the repository's required gate tier
before commit, push, or release. A named gate without a runnable command, test
function, transcript, or proof file is planning prose, not evidence.

## Public Surface Rule

Sub-agents can make Jini development stronger. They must not make Jini harder to
use.

Sub-agent divide-and-conquer is internal and coordinator-owned. It is not a
free-tier command surface, public workflow taxonomy, or default transcript
layer. Tier and product UX permissions belong in
[skills-and-delegation-slice.md](./skills-and-delegation-slice.md), not in this
internal engineering-process spec.

The product should keep hiding internal orchestration behind the normal result
shape:

- named artifact
- short result summary
- what changed
- what still needs attention
- safe next action

If users need to understand the agent tree to trust or operate the default CLI,
the design has failed this operating model.
