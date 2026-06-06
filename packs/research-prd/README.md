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

Inspect them directly:

```bash
jini open packs/research-prd/examples/research-prd-v1/views/prd.md
```

The native publish, adapter, and capture commands are tracked as future Go
work. Until those commands are ported, this pack keeps the canonical artifacts
and rendered views available as a reusable workflow reference.

The rendered PRD view lives in:

- [views/prd.md](examples/research-prd-v1/views/prd.md)

Use the research-to-PRD workflow with the native Go front door:

```bash
jini
```

Paste the research brief, source set, owner, approver, stakeholders, and target
handoff as source context. The native pack compiler is tracked as future Go
work; until that command is ported, this pack is a reusable workflow reference.

The intended output includes:

- canonical research artifacts
- a build-ready `Spec`, `Plan`, and `Tasks`
- a rendered `views/prd.md`
