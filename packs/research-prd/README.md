# Research PRD Pack

This pack turns validated research into a canonical Brief, a rendered PRD view,
and a build-ready Spec/Plan/Tasks handoff.

It is the flagship bridge for:

`research -> synthesis -> brief -> PRD -> spec -> plan -> tasks`

## Intended Profile

- `Delivery`

## Typical Extensions

- `Business:product-discovery`
- `Modality:software`
- `Environment:docs-local`
- `Risk:user-facing`

## Typical Control Packs

- `Proof`
- `Guard`
- `Authority`
- `Cost`

## Compiled Flow

1. `Scope`
   - create `Brief`
   - capture boundaries and research-backed objective

2. `Probe`
   - surface assumptions and unsupported claims

3. `Model`
   - create `Sources`, `Literature`, `Method`, and `Spec`

4. `Decide`
   - create `Decision`, `Plan`, and `Tasks`

5. `Make`
   - render the PRD view and handoff surfaces

6. `Verify`
   - bind research-backed `Evidence` to the working revision

## Example

Validated example artifacts live in:

- [examples/research-prd-v1](examples/research-prd-v1)

Validate them with:

```bash
python3 tools/jini_validate.py validate-pack packs/research-prd/examples/research-prd-v1
python3 tools/jini_validate.py status-pack packs/research-prd/examples/research-prd-v1
python3 tools/jini.py recommend-execution packs/research-prd/examples/research-prd-v1 --intent wiki
python3 tools/jini.py bind-atlassian packs/research-prd/examples/research-prd-v1 --cloud-id 11111111-2222-3333-4444-555555555555 --site-url https://example.atlassian.net --project-key DEMO --space-key DEMO --space-id 123456
python3 tools/jini.py show-atlassian packs/research-prd/examples/research-prd-v1
python3 tools/jini.py run-pack packs/research-prd/examples/research-prd-v1 --mode supervised --consent write
python3 tools/jini.py export-tasks packs/research-prd/examples/research-prd-v1
python3 tools/jini.py sync-tasks packs/research-prd/examples/research-prd-v1
python3 tools/jini.py export-issues packs/research-prd/examples/research-prd-v1 --adapter jira
python3 tools/jini.py export-wiki packs/research-prd/examples/research-prd-v1 --adapter confluence
python3 tools/jini.py export-wiki packs/research-prd/examples/research-prd-v1 --adapter markdown
python3 tools/jini.py publish-issues packs/research-prd/examples/research-prd-v1 --adapter jira --project-key DEMO
python3 tools/jini.py publish-wiki packs/research-prd/examples/research-prd-v1 --adapter confluence --space-key DEMO
python3 tools/jini.py capture-publication packs/research-prd/examples/research-prd-v1 --author release-coordinator --input /tmp/publication-result.json --scope atlassian-publish
python3 tools/jini.py capture-output packs/research-prd/examples/research-prd-v1 \
  --author product-lead \
  --task-index 1 \
  --status done \
  --note "Reviewed source coverage and finalized the research synthesis"
python3 tools/jini.py capture-approval packs/research-prd/examples/research-prd-v1 \
  --author product-ops \
  --approver-actor eng-manager \
  --scope handoff-acceptance
```

The rendered PRD view lives in:

- [views/prd.md](examples/research-prd-v1/views/prd.md)

Compile a research-backed pack from the benchmark context with:

```bash
python3 tools/jini.py compile-pack research-prd \
  --work-unit-id my-research-prd \
  --title "Jini Research To PRD" \
  --purpose "Turn validated research into a PRD and build-ready handoff" \
  --owner product-lead \
  --approver eng-manager \
  --stakeholder research-lead \
  --stakeholder design-lead \
  --output /tmp/my-research-prd
python3 tools/jini.py status-pack /tmp/my-research-prd
```

The compiled output includes:

- canonical research artifacts
- a build-ready `Spec`, `Plan`, and `Tasks`
- a rendered `views/prd.md`
