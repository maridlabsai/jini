# Public Repo Boundary

Updated: 2026-05-18

## Public Repo Rule

The public `jini` repo is for code, public product docs, stable technical
contracts, public examples, install paths, and tests.

Anything in this repo must be safe to index, quote, fork, and review in public.

## Private Commercial Repo Rule

Internal business and review material belongs in the private commercial repo,
not here.

That includes pricing strategy, GTM plans, customer-specific rollout notes,
internal gate reviews, candid competitive notes, and commercialization drafts.

## Forbidden In The Public Repo

The public repo must not contain:

- dated internal notes in `specs/20xx-*.md`
- commercialization docs such as `COMMERCIAL.md`
- internal business directories like `commercial/`, `gtm/`, or `sales/`
- candid internal review writeups that are not stable public doctrine

## Allowed In The Public Repo

The public repo may contain:

- stable product contracts
- stable routing or architecture policy
- public install and usage docs
- public benchmark harnesses and scorecards
- public examples and tests

The distinction is:

- public repo: durable product truth
- private repo: candid internal operating material

## Enforcement

This boundary is enforced by:

- `python3 tools/jini.py validate-public-boundary --format json`
- `python3 tools/jini.py publish-readiness --format json`

Any new internal business or review material should be written in the private
commercial repo first, not migrated out later.
