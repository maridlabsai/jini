# Honest System Audit

Updated: 2026-06-07

This audit keeps product language honest. It separates implemented behavior
from specified intent, guarded documentation, and future roadmap work.

## Status Taxonomy

- `implemented`: runtime behavior exists, tests cover it, and a required gate
  exercises it.
- `guarded`: release or benchmark checks protect the requirement from being
  removed, but the runtime behavior may still be incomplete.
- `partial`: some runtime behavior or test coverage exists, but the full
  user-facing requirement is not proven end to end.
- `specified`: the requirement is written down, but no runtime behavior or
  release gate proves it yet.
- `not-started`: the repo does not contain meaningful implementation or guard
  coverage yet.
- `blocked`: progress needs a dependency, access, external decision, or
  platform capability.

Guarded is not implemented.

## Current Implementation Reality

| Area | Honest status | Evidence | Gap | Next cut |
| --- | --- | --- | --- | --- |
| Native Go CLI | implemented | Go runtime, Go tests, no tracked Python source, required gates | CLI still needs first-minute dogfood across more personas | Keep reducing command and launcher friction |
| Configured CLI handoff | implemented | Wave 0 handoff contract, Wave 1 route registry, `doctor` detection, fake downstream CLI smoke tests, fail-closed missing/trust checks, route receipts | Each downstream CLI still needs real-world dogfood for command templates, approvals, and output shape | Harden Wave 1 command templates against real installed CLIs without broadening the first-minute UX |
| Simplicity as UX tenet | guarded | Canonical PRD, lean gate, product simplicity test | The runtime still exposes too much internal vocabulary in some flows | Prefer natural `jini` intake and progressive disclosure |
| Repo review snapshot | implemented | Direct repo-review test and porcelain parser coverage | It is a model-free first pass, not a full code-review agent | Add richer changed-file focus and security/test prompts |
| P0 competitor watching | guarded | PRD, competitive release plan, readiness guard, benchmark coverage | No watch packet generator, scheduler, or feature-selection ingestion loop exists yet | Build a watch-packet artifact and release-plan ingestion path |
| P0 compounding user productivity learning | guarded | PRD, learning-system spec, readiness guard, benchmark coverage | Runtime learning is partial: multimodal and route signals exist, but stable user context, habits, and inspect/revoke controls are not implemented | Build inspectable user-context learning store with opt-out and productivity metrics |
| Offline and local model story | partial | Local SLM profiles, device/runtime gates, offline strategy, tests | Local/offline behavior is not proven across shipped macOS, Windows, iOS, and Android apps | Add device runtime smoke fixtures and app-surface proof |
| Skills and delegation | specified | Commercial-tier boundary in skills-and-delegation slice and simplicity gate | The agent and skills OS productivity suite is commercial-only and not a free-tier runtime feature | Implement the commercial suite from the commercial PRD without adding `skills` or `delegate` commands to the free tier |
| App surfaces | specified | App platform playbook and app-surface PRDs | Desktop and mobile apps are not shipped clients yet | Build app-shell proof over the same work object |
| Publish readiness | partial | `jini publish-readiness` checks docs, specs, benchmark coverage, runtime counts, and claim evidence | Many checks are fragment checks, not functional proofs | Require every P0 claim to list implementation evidence or say it is only guarded |

## Core Feedback Accommodations

The core feedback is that Jini must be simpler, more honest, and more useful
over time instead of accumulating impressive but ambiguous requirements.

Accommodations now required:

- Every major claim needs claim, status, evidence, and next cut.
- No P0 is complete from documentation alone.
- Release readiness may guard a requirement without implying it is implemented.
- Competitor watching must decide features, integrations, watch items,
  rejections, and deletions.
- User learning must improve productivity through inspectable, scoped, and
  reversible context, not hidden personalization.
- Skills and agents must reduce user work without creating a second command
  tree.
- Simplicity remains the UX tenet: one obvious next action, fewer visible
  choices, and advanced controls only when they protect trust, cost, recovery,
  or safety.

## Machine-Readable Evidence Contract

`jini publish-readiness --format json` must expose honest audit claims as
machine-readable objects with:

- `claim`: the product or engineering claim being made.
- `status`: one of the audit taxonomy statuses.
- `evidence`: the best current repo evidence for the claim.
- `gap`: the highest-risk missing proof.
- `next_cut`: the next concrete implementation or validation cut.
- `runtime_implemented`: whether runtime behavior exists today.

Readiness `ok` does not mean every claim is implemented. It means the repo has
the required guardrails and the current claim status is visible enough for a
release decision.

## Progressive Harsh Review

| Persona | Verdict | Core criticism |
| --- | --- | --- |
| JuniorDevNitpicker | pass with fixes | The repo now has clear specs, but some readiness labels can still sound more complete than the system is. |
| SeniorArchCritic | pass with fixes | Several P0 items are guarded by documentation checks before they are runtime systems. The taxonomy must be visible in release work. |
| ProdOpsHardass | pass with fixes | A future release can still overclaim if release notes do not separate implementation evidence from guardrail coverage. |

Overall verdict: pass with fixes. The immediate fix is this audit plus a
publish-readiness guard. The next fix is implementation-evidence scoring for
each P0 requirement.
