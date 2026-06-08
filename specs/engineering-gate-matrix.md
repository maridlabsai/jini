# Engineering Gate Matrix

Updated: 2026-06-06

This document is a specialized engineering quality-gate contract, not the
top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this gate matrix conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, or automation posture, the canonical PRD wins and
this matrix should be updated.

## Purpose

Jini must have one canonical engineering gate contract that answers three
questions without tribal knowledge:

- what must pass before every commit
- what must pass before every push
- what must pass before every release

The repo must also provide one checked-in runner so these gates are executed
the same way every time.

## Gate Tiers

### Commit gate

This is the minimum required gate for every local commit.

Required commands:

1. `go test ./...`
2. `git diff --check`
3. `git diff --cached --check`
4. `bash tools/security_configuration_gate.sh`
5. `bash tools/product_prd_drift_gate.sh`
6. `bash tools/cli_ux_regression_gate.sh`
7. `jini scorecard-gate --format json`

Required outcome:

- Go runtime regressions are caught immediately
- the migration boundary blocks tracked Python files and Python gate invocations
- staged and unstaged whitespace and patch-format drift are blocked before commit
- scanner wiring for CodeQL, govulncheck, OSV-Scanner, TruffleHog, and
  Dependabot cannot be removed without failing the local gate
- protected PRD and product-positioning surfaces cannot drift unless
  `specs/product-settling-decisions.md` is updated in the same change
- direct CLI edit and simple-question flows cannot regress into draft/status frames, `Start/Keep` choices, or verbose current-work summaries
- competitive scorecard drift is blocked before commit, including required
  coverage for async/background agents, cross-surface continuity, visible
  progress and outputs, permissioned execution, skills/hooks/context routing,
  local/open-model optionality, and scorecard gate wiring
- Outcome gates require executable or named proof references, not just competitor or fixture names.
- A gate name without a runnable command or named proof reference is planning prose, not evidence.

### Push gate

This is the minimum required gate before pushing a branch for broader review or
integration.

Required commands:

1. all commit-gate commands
2. `jini check ship --format json`

Required outcome:

- the branch clears the same Go-only implementation boundary before push
- free security scanning remains configured before the branch reaches CI
- push gate records local shipping evidence, including git repository state and
  required validation evidence
- dirty worktrees are blocked before push

### Release gate

This is the minimum required gate before release packaging, release promotion,
or public shipping claims.

Required commands:

1. all push-gate commands
2. `jini publish-readiness --format json`

Required outcome:

- readiness output is available as a machine-readable proof surface
- release work cannot skip the checked public-contract and readiness layers
- release work cannot claim competitor catch-up while the scorecard gate is
  missing required competitor or pressure-vector coverage

## Canonical Runner

The checked-in runner for these tiers is:

- `bash tools/run_required_gates.sh commit`
- `bash tools/run_required_gates.sh push`
- `bash tools/run_required_gates.sh release`

Convenience aliases must also exist in the repo's `Makefile`:

- `make gates-commit`
- `make gates-push`
- `make gates-release`

## Operating Rules

### Rule 1: No implicit gate definitions

Humans should not have to remember which commands count as the real gate.

### Rule 2: Tier obligations are cumulative

`push` includes `commit`.

`release` includes `push`.

### Rule 3: Expand by policy, not by folklore

If new required checks are added, this matrix and the checked-in runner must be
updated in the same change.

### Rule 4: Narrow local checks are still allowed

Focused tests are encouraged during iteration.

They do not replace the required tier gates before commit, push, or release.

## Promotion Candidates

No separate promotion candidates are currently listed.

Future checks should be added here only when they are not already required by a
commit, push, or release tier. Once promoted, remove them from this section in
the same change that updates the checked-in runner.
