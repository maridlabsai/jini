# Engineering Gate Matrix

Updated: 2026-07-12

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
6. `bash tools/customer_value_gate.sh`
7. `bash tools/cli_ux_regression_gate.sh`
8. `bash tools/claude_codex_usecase_gate.sh`
9. `bash tools/macos_bundle_hygiene_gate.sh`
10. `jini scorecard-gate --format json`

Required outcome:

- Go runtime regressions are caught immediately
- the migration boundary blocks tracked Python files and Python gate invocations
- staged and unstaged whitespace and patch-format drift are blocked before commit
- macOS sidecar binaries are generated locally and must not be tracked
- scanner wiring for CodeQL, govulncheck, OSV-Scanner, TruffleHog, and
  Dependabot cannot be removed without failing the local gate
- protected PRD and product-positioning surfaces cannot drift unless
  `specs/product-settling-decisions.md` is updated in the same change
- customer-value viability cannot regress into amateur platform claims,
  unsupported route claims, generic workflow scaffolds, or new vocabulary
  before value
- direct CLI edit and simple-question flows cannot regress into draft/status frames, `Start/Keep` choices, or verbose current-work summaries
- the intent/parity golden transcript gate blocks questions, bare entities, and
  explicit task intents from regressing away from Claude/Codex first-minute
  expectations
- Claude and Codex user journeys are exercised as concrete commit-gate use cases, not only as personas in docs: local file edits, repo review, strict CLI handoff, custom Claude args, Codex handoff, failed handoff recovery, compact questions, and the 100-prompt Aryan-derived first-minute prompt bank
- competitive scorecard drift is blocked before commit, including required
  coverage for async/background agents, cross-surface continuity, visible
  progress and outputs, permissioned execution, skills/hooks/context routing,
  local/open-model optionality, and scorecard gate wiring
- Outcome gates require executable or named proof references, not just competitor or fixture names.
- A gate name without a runnable command or named proof reference is planning prose, not evidence.
- Named-proof refs must resolve to existing repository files; executable refs
  must name real Go test functions.
- customer-value gates must name concrete customer outcomes: token savings,
  throttle resilience, configured-tool switching reduction, direct action,
  safety, or session continuity

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
- `jini check ship --format json` exposes Wave 1 CLI handoff setup status,
  dogfood validation status, smoke freshness for claimed routes, missing validation checks,
  relative `.jini/cli-dogfood.json` evidence files, and relative `.jini/cli-smoke.json` evidence files
- installed CLI dogfood before release must verify auth, approvals, output shape, and route receipt privacy
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

Release bars added by the rebuilt PRD (see number-one-platform-prd.md Gates):

- TTFV measured under 5 minutes on fresh macOS/Linux/Windows machines from
  signed artifacts (model download excluded and disclosed).
- First-task success measured in release qualification.
- Task-success benchmark published for every recommended model, clearing its
  device-class floor; every matrix model's license verified (permissive,
  redistribution-safe, official-source download).
- Savings-methodology audit passed, including literal-vs-imputed labeling and
  pricing-table provenance.
- Paywall fail-closed verification: no entitlement -> paid features cleanly
  absent, free equivalents work.
- Token-efficiency regression suite green.
- Crash-free session rate at least 99.5% in qualification.

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

## Absorbed Contracts

These normative rules were merged here from now-archived gate docs during the
PRD rebuild (prd-rebuild-design.md §8). They are traced to the rebuilt PRD's
Routing And Resource Policy and Token Economy sections.

### Lean platform gate (absorbed from lean-platform-gate.md)

Token frugality is P0 and must be treated as a first-order gate, not a generic
cost optimization.

#### 3. Command-Surface Discipline

The product must preserve `one-stable-surface`, low `command-surface-count`, and
`no-compatibility-aliases` in the taught surface.

This discipline also applies to skills and agent interactions. Specialist
helpers must stay reachable through natural intake and progressive disclosure;
they must not become a second command tree or visible agent control plane.
The free tier must not include a skills-based OS productivity suite.

Required regression inputs:

- `token-frugality-p0`
- `offline-online-session-stitching`
- `throttle-driven-platform-switching`
- `task-shaped-model-selection`
- `power-and-battery-aware-routing`
- `skill-agent-interaction-simplicity`

Reject any change that:

- increases token load, transcript replay, or verbose output without measurable
  quality, trust, or safety gain
- removes or weakens powered-mode full power execution when local capability is
  useful and safe
- removes or weakens low-battery or thermal-aware execution so local routes
  keep burning device resources when a smaller, deferred, or remote route would
  preserve the outcome
- splits offline and online execution into separate transcripts, task ids, or
  route histories instead of stitching them into one session timeline
- teaches skill or agent vocabulary as a prerequisite to normal use
- ships developer agents, tester agents, `skills`, `delegate`, or a skills-based OS productivity suite in the free tier
- shows agent trees, role theater, or orchestration logs by default

### Dogfood, friction, and engineering principles (absorbed)

- Persona dogfood rules (from dogfood-gates.md): `jini` must remain the normal
  front door after install; `Auto` is explained in plain language ("Jini picks
  for you"), not as system-state theory.
- Friction reduction (from friction-reduction-gate.md): `jini` starts without
  stale default work output; help shows examples and setup guidance; standard
  one-word commands are never shadowed by taught multiword vocabulary.
- Engineering principles (from engineering-principles.md): SOLID/OOP discipline
  and the reject conditions for over-engineered abstractions remain the code
  review baseline for Go work in this repo.

## Promotion Candidates

No separate promotion candidates are currently listed.

Future checks should be added here only when they are not already required by a
commit, push, or release tier. Once promoted, remove them from this section in
the same change that updates the checked-in runner.
