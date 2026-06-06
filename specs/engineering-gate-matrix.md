# Engineering Gate Matrix

Updated: 2026-06-05

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

1. `python3 tools/language_gate.py`
2. `python3 -m unittest tests.test_go_first_launcher -v`
3. `GOCACHE=/private/tmp/jini-go-cache GOMODCACHE=/private/tmp/jini-go-mod /Users/sharad.sharma/Developer/.local-go/bin/go test ./internal/app`
4. `git diff --check`

Required outcome:

- changed user-facing files are scanned for blocked language without relying on
  a machine-local helper path
- launcher boundary regressions are caught immediately
- Go runtime regressions in the main app package are caught immediately
- whitespace and patch-format drift is blocked before commit

### Push gate

This is the minimum required gate before pushing a branch for broader review or
integration.

Required commands:

1. all commit-gate commands
2. `python3 -m unittest discover -s tests -p 'test_*docs.py' -v`

Required outcome:

- the public docs and documentation-contract suites are green
- the branch clears both the implementation boundary gate and the public-doc
  contract gate before push

### Release gate

This is the minimum required gate before release packaging, release promotion,
or public shipping claims.

Required commands:

1. all push-gate commands
2. `python3 tools/jini.py publish-readiness --format json`

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

- `python3 -m unittest tests.test_jini_cli -v`
- `python3 -m unittest tests.test_install_sh -v`
- `python3 -m unittest discover -s tests -v`
- `GOCACHE=/private/tmp/jini-go-cache GOMODCACHE=/private/tmp/jini-go-mod /Users/sharad.sharma/Developer/.local-go/bin/go test ./...`

They should be treated as promotion candidates, not quietly implied required
gates.
