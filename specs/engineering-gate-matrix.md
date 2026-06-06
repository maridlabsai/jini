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
3. `bash tools/security_configuration_gate.sh`

Required outcome:

- Go runtime regressions are caught immediately
- the migration boundary blocks tracked Python files and Python gate invocations
- whitespace and patch-format drift is blocked before commit
- scanner wiring for CodeQL, govulncheck, OSV-Scanner, and Dependabot cannot
  be removed without failing the local gate

### Push gate

This is the minimum required gate before pushing a branch for broader review or
integration.

Required commands:

1. all commit-gate commands

Required outcome:

- the branch clears the same Go-only implementation boundary before push
- free security scanning remains configured before the branch reaches CI

### Release gate

This is the minimum required gate before release packaging, release promotion,
or public shipping claims.

Required commands:

1. all push-gate commands
2. `jini publish-readiness --format json`

Required outcome:

- readiness output is available as a machine-readable proof surface
- release work cannot skip the checked public-contract and readiness layers

## Canonical Runner

The checked-in runner for these tiers is:

- `bash tools/run_required_gates.sh commit`
- `bash tools/run_required_gates.sh push`
- `bash tools/run_required_gates.sh release`

Convenience aliases should also exist in the repo's `Makefile`.

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

These suites are important, but they are not yet part of the canonical required
gate tiers until they are baseline-green and explicitly promoted in this
matrix:

- `GOCACHE=/private/tmp/jini-go-cache GOMODCACHE=/private/tmp/jini-go-mod /Users/sharad.sharma/Developer/.local-go/bin/go test ./...`

They should be treated as promotion candidates, not quietly implied required
gates.
