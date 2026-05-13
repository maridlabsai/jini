# Changelog

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
python3 tools/jini.py status-pack packs/research-prd/examples/research-prd-v1
```

Fastest install:

```bash
pipx install --editable git+https://github.com/maridlabsai/jini.git
jini get-started --target codex
```

Distribution note:

- `v0.1.0` officially supports the editable source-backed install path above
- conventional wheel-style distribution is intentionally deferred until the
  packaged runtime boundary is separated cleanly from writable operator state
