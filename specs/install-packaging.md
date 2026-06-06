# Jini Install Packaging

## 1. Purpose

Jini should be installable as a product, not only cloned as a repo.

The first packaging surface in this repo is intentionally conservative:

- a machine-readable install manifest
- curated install kits
- a discoverable bundle catalog
- a dry-run planner
- explicit target shims
- a receipt-oriented summary
- doctor output that reflects receipt and activation state

This is the trust layer that should exist before Jini writes into multiple
agent environments.

## 2. Design Rules

Packaging should follow these rules:

- universal payloads stay canonical and shared
- target-specific behavior lives in thin shims
- install scope is previewable before any write
- provenance and risk are surfaced inline
- install, update, uninstall, and verify should eventually form one symmetric lifecycle

## 3. Bundle Model

The current install manifest groups installable capability into bundles.

Each bundle declares:

- `id`
- `label`
- `kind`
- `version`
- `summary`
- `permission_risk`
- `review_before_use`
- `compatible_targets`
- `universal_paths`
- optional `depends_on`
- optional `activation_steps`
- optional `migration_notes`
- optional `target_shims`

This keeps compatibility and risk visible in one place.

The manifest can also declare curated install kits.

Each kit groups bundles for a common adoption path, such as:

- starter surface
- delivery benchmark surface
- personal-OS surface

Kits can also carry an intended audience:

- `beginner`
- `power-user`
- `hybrid`

This matters because most operators should not need bundle-level repo knowledge
before they can preview or install a useful Jini surface.

## 4. Target Model

Each target declares:

- `id`
- `label`
- destination root
- shim root
- default link mode
- target-specific risk notice
- target-specific activation steps

The target model should remain thin.

Targets do not redefine bundle semantics.
They only define where and how Jini binds into a local environment.

## 5. Dry-Run Planner

The CLI command is:

```bash
jini get-started --target codex
jini plan-install --kit starter-kit --target codex
jini install-bundles --kit starter-kit --target codex --prefix /tmp/jini-stage
jini doctor-install --kit starter-kit --target codex --prefix /tmp/jini-stage
jini catalog-bundles
jini catalog-bundles --target codex --format json
jini plan-install --kit operations-response-kit --target codex
jini plan-install --kit regulated-readiness-kit --target codex
jini plan-install --kit vendor-decision-kit --target codex
jini update-bundles --kit starter-kit --target codex --prefix /tmp/jini-stage
jini uninstall-bundles --kit starter-kit --target codex --prefix /tmp/jini-stage
jini plan-install --bundle jini-core --target codex --target kiro-cli
jini plan-install --format json
```

The planner should show:

- source and revision
- the manifest default kit when no explicit bundle or kit is selected
- selected kits
- selected bundles
- selected targets
- copy or symlink strategy
- universal payload destinations
- shim destinations
- permission and review notices
- an install receipt id
- a manifest digest for auditability
- the next install and doctor commands that complete the trust path

The planner does not write files.

If no explicit `--bundle` or `--kit` is provided, Jini should prefer the
manifest default kit instead of expanding to every known bundle. Broad installs
are still available, but they should be explicit.

Curated kits should also cover materially different advanced surfaces. The
install manifest now exposes an `operations-response-kit` so incident-response
adoption does not require bundle-by-bundle assembly.
It also exposes a `regulated-readiness-kit` so governed audit and approval
workflows are installable through the same curated path.
It also exposes a `vendor-decision-kit` so commercial evaluation and approval
workflows are installable through the same curated path.

Install trust should not stop at path materialization. `doctor-install` should
report target-specific health semantics, including receipt presence, shim
documentation quality, manifest freshness, link-mode behavior, activation-target
consistency, runtime activation readiness when a target has already been
activated, and lightweight behavioral probes like `plan-install` and
`resolve-adapter` so target bindings are validated as working surfaces rather
than only as files on disk.

## 5.1 Install Lifecycle

The current lifecycle is:

- `get-started`: show the curated trust path first and demote raw bundle detail
- `catalog-bundles`: discover curated kits first and use JSON mode for deeper bundle inspection
- `plan-install`: preview source, targets, paths, shims, risk, and receipt id
- `install-bundles`: materialize universal payloads and target shims
- `update-bundles`: refresh installed bundle content from source truth
- `doctor-install`: verify installed paths, locate matching receipts, and surface activation hints
- `uninstall-bundles`: remove installed bundle roots and emit an uninstall receipt

For tests and local staging, use `--prefix` so install paths are remapped into a
safe writable root.

The actual install write model is still intentionally conservative:

- bundle payloads are installed under a bundle root
- target shims are installed separately under a shim root
- install and uninstall both emit receipts
- doctor verifies presence of expected installed paths, matching receipts, latest receipt status, and target-specific smoke probes
- manifest digests make the planned install auditable against the source manifest used at install time

## 6. Why This Matters

Packaging is product maturity.

Without it, Jini remains a source checkout for repo-literate operators.

With it, Jini can become:

- easier to adopt
- easier to trust
- easier to distribute across runtimes
- easier to audit when install behavior changes
