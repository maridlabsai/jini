# Changelog

## v0.1.1

Product-facing CLI correction for first-minute user testing.

What changed:

- bare `jini` now opens as a compact task-first prompt even when saved work exists
- saved work no longer appears as a startup dashboard
- visible `Switch` and `Start/Keep` front-door vocabulary is removed from normal use
- typing a saved work title resumes that thread naturally
- canonical PRD is reduced to the current CLI-first GTM wedge
- CLI UX regression gate now pins the rejected tester scenarios

## v0.1.0

Initial public release of Jini.

What is included:

- strict protocol core for governed, stateful AI work
- proof-first CLI surface built around pack state, verification, approval, and durable memory
- manifest-driven install planning with curated kits and target shims
- public-core packs for research-to-PRD, meeting follow-up, travel planning, budget planning, incident response, compliance audit, and vendor selection
- personal-OS layer for memory, routines, and continuity
- runtime handoff, activation, publish-plan, and evidence-harvest surfaces
- security hardening on repo-derived verification targets and raw shell-backed routines

Fastest proof:

```bash
jini publish-readiness --format json
```

Fastest install:

```bash
bash install.sh
jini commands
```

Distribution note:

- `v0.1.0` officially supports the native Go install path above
- advanced pack compiler and adapter commands are tracked as native Go backlog
  items instead of Python fallback surfaces
