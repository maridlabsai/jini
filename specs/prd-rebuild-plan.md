# Jini PRD Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land the new canonical PRD from `specs/prd-rebuild-design.md`, execute the specs archive sweep with gate repointing, and regenerate all tracing — without ever breaking the commit gate.

**Architecture:** Four-phase gated sequence (design §8): (1) land PRD + settling decisions; (2) repoint gate scripts while old files still exist; (3) merge-then-archive and archive in small batches, fixing Go references per batch; (4) regenerate trace, update doctrine. Every task ends with the full commit gate green and a commit.

**Tech Stack:** Markdown specs, bash gate scripts (`tools/*.sh`), Go tests (`internal/app/`), `jini scorecard-gate`.

## Global Constraints

- **Gate command after every task:** `bash tools/run_required_gates.sh commit` from repo root. A task is not done until it exits 0.
- **Drift-gate law:** any change touching a file listed in `tools/product_prd_drift_gate.sh` lines 28–52 (includes `README.md` and `specs/number-one-platform-prd.md`) MUST modify `specs/product-settling-decisions.md` in the same change.
- **Customer-value-gate law:** these four fragments must exist VERBATIM in `specs/number-one-platform-prd.md` at every commit (from `tools/customer_value_gate.sh:51-54`):
  1. `Customer value bar: every shipped cut must improve token frugality, throttle`
  2. `Preserve customer-value viability: reduce token waste, throttle friction,`
  3. `bash tools/customer_value_gate.sh`
  4. `No commit ships unless the customer-value gate can still prove the solution is`
- **Go-reference law:** ~30 spec files are referenced from `internal/app/*.go` (tests and runtime). Never `git mv` a spec file without fixing every Go reference in the same commit. Discovery command: `grep -rn 'specs/<filename>' --include='*.go' internal/`
- **Archive banner (verbatim template, fill the two bracketed slots):**
  ```markdown
  > **SUPERSEDED (2026-07-07).** Archived by the PRD rebuild; no longer a requirement source.
  > Replaced by: [<target file>](<relative path>) <§section if applicable>.
  > Rationale: `specs/prd-rebuild-design.md` §8. Do not cite this file in new work.
  ```
- **Free/commercial boundary:** no prices, SKUs, or entitlement internals in any file in this public repo.
- **Nothing is deleted.** Archived files move to `specs/archive/` with the banner prepended; content is otherwise untouched.
- Commit messages end with: `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`

---

### Task 1: Baseline — commit design + plan, verify green start

**Files:**
- Create (already on disk): `specs/prd-rebuild-design.md`, `specs/prd-rebuild-plan.md`

**Interfaces:**
- Produces: a green baseline commit every later task diffs against.

- [ ] **Step 1: Verify the gate is green before any work**

Run: `bash tools/run_required_gates.sh commit`
Expected: exit 0. If red on a clean tree, STOP — fix the environment first; nothing in this plan proceeds on a red baseline.

- [ ] **Step 2: Commit the design and plan docs**

Neither file is in the drift-gate protected list, so no settling-decisions edit is needed here.

```bash
git add specs/prd-rebuild-design.md specs/prd-rebuild-plan.md
git commit -m "docs: add PRD rebuild design and implementation plan

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

- [ ] **Step 3: Re-run gate on the commit**

Run: `bash tools/run_required_gates.sh commit`
Expected: exit 0.

### Task 2: Reference inventory (read-only)

**Files:**
- Create: `/private/tmp` scratch only — no repo files.

**Interfaces:**
- Produces: `sweep-refs.txt`, the per-file entanglement map every later task consults before moving a file.

- [ ] **Step 1: Generate the map**

```bash
for f in $(ls specs/ | grep -v '^archive$'); do
  n=$(grep -rln "specs/$f" --include='*.go' internal/ tools/ 2>/dev/null | wc -l | tr -d ' ')
  echo "$n $f"
done | sort -rn > "${TMPDIR:-/tmp}/sweep-refs.txt"
cat "${TMPDIR:-/tmp}/sweep-refs.txt"
```

Expected: a count per spec file. Known-heavy entries to confirm: `app-platform-shipping-playbook.md`, `competitive-release-plan.md`, `honest-system-audit.md`, `skills-and-delegation-slice.md`, `lean-platform-gate.md`, `learning-system.md`, `platform-offline-strategy.md`.

- [ ] **Step 2: Extract every PRD-pinned string in Go tests**

```bash
grep -rn 'specs/number-one-platform-prd.md' --include='*_test.go' internal/ | cut -d: -f1 | sort -u
```

Then open each listed test file and copy every string literal it asserts against PRD content into a checklist. Task 3's rewrite must preserve each pinned string verbatim OR the same commit must update that test with an equivalent assertion against the new PRD text (prefer preserving; update tests only where the design explicitly changed the requirement, and say which design section justifies it).

### Task 3: Land the new PRD + settling decision (sweep step 1)

**Files:**
- Modify: `specs/number-one-platform-prd.md` (full rewrite)
- Modify: `specs/product-settling-decisions.md` (append section)
- Modify: `internal/app/*_test.go` only where Task 2's checklist requires
- Possibly modify: `README.md` if a pinned positioning test couples it to PRD phrases (Task 2 tells you)

**Interfaces:**
- Consumes: Task 2's pinned-string checklist.
- Produces: the canonical PRD every later task's banners point at.

- [ ] **Step 1: Append the settling decision (drift-gate key — do this first)**

Append to `specs/product-settling-decisions.md`, after the last section, and bump the `Updated:` line to `2026-07-07`:

```markdown
## PRD Rebuild Decision (2026-07-07)

The canonical PRD is rewritten from specs/prd-rebuild-design.md. Decisions:

- Product identity: token-savings mastery with receipts; free to run; routes
  everything; dollar savings is the primary receipt denomination, always shown
  with tokens and time saved, labeled literal (metered API routes) vs imputed
  (subscription routes, valued at API-equivalent prices).
- Primary user: professional developer at a FAANG-class company; the quality
  bar is first-minute indistinguishability from frontier-lab tooling.
- v1 scope is Option A (wedge first): BYO/gateway/CLI-handoff agent, sessions,
  receipts/ledger, throttle fallback, on-the-fly skills/agents, live paywall
  (Autopilot + Continuity, fail-closed). Local execution rides detected
  third-party runtimes as a disclosed interim route; the first-party runtime
  and curated permissive-license-only model matrix are v1.5.
- On-the-fly skills/agents creation is a free-tier feature (competitive
  parity). This narrows the earlier "agent/skills features are commercial"
  doctrine; the commercial repo keeps the productivity-suite/OS feature set.
- The specs/ directory is swept: one canonical PRD, a small keep-set, and
  specs/archive/ for everything superseded, per prd-rebuild-design.md §8,
  executed as a gated sequence, never one mega-change.
```

- [ ] **Step 2: Rewrite the PRD**

Replace the body of `specs/number-one-platform-prd.md` with design sections 1–7 of `specs/prd-rebuild-design.md`, applying these transformations:

1. Title: `# Jini Platform PRD` with `Updated: 2026-07-07` and one line: `Rebuilt from specs/prd-rebuild-design.md; decisions recorded in specs/product-settling-decisions.md.`
2. Drop design-process artifacts: the "Approved judgment calls", "Next steps", and §8 sweep sections stay in the design doc only. Keep §9 (risk register) — a PRD with a risk register is deliberate.
3. Convert "Scope decision — RESOLVED" prose into plain scope statements (v1 ships X; v1.5 ships Y) without the decision-history framing.
4. Insert the four customer-value-gate fragments verbatim where they fit naturally:
   - In §5 add the bullet: `Customer value bar: every shipped cut must improve token frugality, throttle` `resilience, switching cost, direct action, safety, or session continuity — or it does not ship.`
   - In §5 add the bullet: `Preserve customer-value viability: reduce token waste, throttle friction,` `and tool-switching cost in every shipped change.`
   - In §7's commit-gate list keep the literal command text `bash tools/customer_value_gate.sh`.
   - In §7 add the sentence: `No commit ships unless the customer-value gate can still prove the solution is` `useful, route-backed, and non-amateur rather than a generic workflow shell.`
5. Preserve every string from Task 2's pinned checklist (or update the pinning test in this same commit, citing the design section).

- [ ] **Step 3: Run the two content gates directly for fast feedback**

Run: `bash tools/customer_value_gate.sh && bash tools/product_prd_drift_gate.sh`
Expected: both exit 0 (drift gate passes because settling doc changed).

- [ ] **Step 4: Run the full gate**

Run: `bash tools/run_required_gates.sh commit`
Expected: exit 0. If a Go test fails, it is a pinned-string miss — reconcile per Task 2 rule; do not weaken tests to pass.

- [ ] **Step 5: Commit**

```bash
git add specs/number-one-platform-prd.md specs/product-settling-decisions.md internal/ README.md
git commit -m "docs: land rebuilt canonical PRD (Option A scope, token-savings identity)

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

### Task 4: Repoint the drift gate (sweep step 2)

**Files:**
- Modify: `tools/product_prd_drift_gate.sh:28-52`
- Modify: `specs/product-settling-decisions.md` (the gate protects itself — one-line note under the Task 3 section: `- Drift-gate protected set repointed to the post-rebuild keep-set.`)

**Interfaces:**
- Produces: the protected set later archive tasks rely on — files leaving `specs/` must already be off this list.

- [ ] **Step 1: Replace the case-statement list**

In `tools/product_prd_drift_gate.sh`, replace the pattern list inside `is_protected_product_surface()` (currently lines 28–52) with the keep-set:

```bash
    README.md | \
    specs/number-one-platform-prd.md | \
    specs/prd-rebuild-design.md | \
    specs/product-settling-decisions.md | \
    specs/product-rewrite-contract.md | \
    specs/engineering-gate-matrix.md | \
    specs/execution-routing-policy.md | \
    specs/local-model-support-matrix.md | \
    specs/public-repo-boundary.md | \
    specs/canonical-names.md | \
    specs/prd-implementation-trace.md)
```

Old files drop off the protected list NOW, while still in place — so later archive batches don't trip the drift gate, exactly as design §8 sequencing intends.

- [ ] **Step 2: Verify the gate still works both ways**

```bash
bash tools/product_prd_drift_gate.sh   # clean tree state: exit 0
echo "x" >> specs/number-one-platform-prd.md
bash tools/product_prd_drift_gate.sh && echo "GATE BROKEN" || echo "gate correctly failed"
git checkout -- specs/number-one-platform-prd.md
```

Expected: second invocation fails (protected file changed, settling doc unchanged in tree) — if it prints `GATE BROKEN`, the case syntax is wrong; fix before proceeding.

- [ ] **Step 3: Full gate + commit**

Run: `bash tools/run_required_gates.sh commit` → exit 0, then:

```bash
git add tools/product_prd_drift_gate.sh specs/product-settling-decisions.md
git commit -m "gates: repoint drift-gate protected set to post-rebuild keep-set

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

### Task 5: Rewrite the gate matrix (absorb its merge group)

**Files:**
- Modify: `specs/engineering-gate-matrix.md`
- Sources to absorb (archived in Task 6): `specs/dogfood-gates.md`, `specs/lean-platform-gate.md`, `specs/friction-reduction-gate.md`, `specs/engineering-principles.md`

**Interfaces:**
- Produces: gate matrix traced to PRD §7; Task 6 archives the four sources against it.

- [ ] **Step 1: Rewrite**

Keep the three-tier structure and every required command list unchanged. Apply:
1. `Updated: 2026-07-07`; pointer paragraph now says the canonical PRD is `number-one-platform-prd.md` **as rebuilt**, and each gate tier cites the PRD §7 bullet it implements.
2. Add release-tier bars from PRD §7: TTFV measurement, first-task success, model license verification, savings-methodology audit (literal-vs-imputed labeling), paywall fail-closed test, token-efficiency regression suite.
3. Append one `## Absorbed Contracts` section: for each of the four source docs, copy its still-true normative rules (the requirements Go tests reference — check Task 2 map for `lean-platform-gate.md` refs and keep those rule names word-for-word so tests keep matching), attributed like `(absorbed from dogfood-gates.md)`. Drop anything the new PRD contradicts; list what was dropped and why in one line each.

- [ ] **Step 2: Full gate + commit**

Run: `bash tools/run_required_gates.sh commit` → exit 0, then:

```bash
git add specs/engineering-gate-matrix.md specs/product-settling-decisions.md
git commit -m "docs: retrace gate matrix to rebuilt PRD and absorb gate-doc merge group

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

(`engineering-gate-matrix.md` is now protected — include the settling-doc one-liner `- Gate matrix retraced to PRD §7.` under the Task 3 section.)

### Task 6: First archive batch — the absorbed gate docs

**Files:**
- Move: `specs/{dogfood-gates,lean-platform-gate,friction-reduction-gate,engineering-principles}.md` → `specs/archive/`
- Modify: every Go/tools file referencing them (Task 2 map; `lean-platform-gate.md` has ~6 refs)

- [ ] **Step 1: Create archive dir and move with banners**

```bash
mkdir -p specs/archive
for f in dogfood-gates lean-platform-gate friction-reduction-gate engineering-principles; do
  git mv "specs/$f.md" "specs/archive/$f.md"
done
```

Prepend to each moved file (after its `# title` line) the Global-Constraints banner with `Replaced by: [engineering-gate-matrix.md](../engineering-gate-matrix.md) §Absorbed Contracts`.

- [ ] **Step 2: Fix references**

```bash
grep -rn 'specs/dogfood-gates\|specs/lean-platform-gate\|specs/friction-reduction-gate\|specs/engineering-principles' --include='*.go' internal/ tools/
```

For each hit: if the code checks the file exists/contains rules, repoint the path to `specs/engineering-gate-matrix.md` (the rules were absorbed verbatim); if a test asserts the *old file* specifically, repoint the path in the test. Never delete an assertion — repoint it.

- [ ] **Step 3: Full gate + commit**

Run: `bash tools/run_required_gates.sh commit` → exit 0, then commit as `docs: archive absorbed gate docs (batch 1)` with the standard trailer.

### Task 7: Routing merge group

**Files:**
- Modify: `specs/execution-routing-policy.md` (+ settling one-liner, it's protected)
- Move → `specs/archive/`: `runtime-execution-modes.md`, `runtime-selection-heuristics.md`, `device-capability-routing.md`, `research-informed-heuristics.md`
- Modify: Go refs (`execution-routing-policy.md` itself has ~3 Go refs — it stays; the moved four, per Task 2 map)

- [ ] **Step 1: Merge** — append an `## Absorbed Policies` section to `execution-routing-policy.md` with each source's normative content (mode definitions, selection heuristics, device-capability rules), deduplicated against what the file already says; every rule name a Go test greps for is kept word-for-word. Add a top note: `Traced to number-one-platform-prd.md §4.3.`
- [ ] **Step 2: Move + banner** — same pattern as Task 6, `Replaced by: [execution-routing-policy.md](../execution-routing-policy.md) §Absorbed Policies`.
- [ ] **Step 3: Fix Go refs** — `grep -rn 'specs/runtime-execution-modes\|specs/runtime-selection-heuristics\|specs/device-capability-routing\|specs/research-informed-heuristics' --include='*.go' internal/ tools/` → repoint to `execution-routing-policy.md`.
- [ ] **Step 4: Full gate + commit** — `docs: merge routing specs into execution-routing-policy (batch 2)`.

### Task 8: Local-execution merge group

**Files:**
- Modify: `specs/local-model-support-matrix.md` (+ settling one-liner, protected)
- Move → `specs/archive/`: `platform-offline-strategy.md` (~6 Go refs), `local-slm-frontline-policy.md`, `device-runtime-gate.md`

Same 4-step pattern as Task 7. Merge target gets `## Absorbed Policies` + `Traced to number-one-platform-prd.md §4.2 (v1.5 scope; v1 interim third-party-runtime rule per §3 Option A).` Add PRD §4.2's licensing rule as a matrix admission criterion: only permissively-licensed, redistribution-safe models; official-source downloads; license shown at consent. Commit: `docs: merge local-execution specs into local-model-support-matrix (batch 3)`.

### Task 9: Benchmark merge group + competitor promotion

**Files:**
- Modify: `specs/golden-competitive-benchmark.yaml`
- Move → `specs/archive/`: `adapter-benchmark-gate.md` (~3 Go refs), `adapter-capability-benchmarking.md`, `competitive-kpis.yaml` (~1 Go ref)

- [ ] **Step 1: Promote the free/BYO agent competitors (PRD §9 requirement)**

In `golden-competitive-benchmark.yaml` `comparison_model.core_benchmark_set`, add (and remove from `watchlist`): `Cline`, `Aider`, `Roo Code`, `Goose`, `OpenCode`, `Continue`. Inspect how existing core entries are scored in the scenarios sections of the same file and add entries for the six in the identical schema — copy an existing competitor's block shape, fill honest scores. Do NOT remove the three fragments `customer_value_gate.sh` pins in this file (`customer-value-viability-fixture`, `product-viability-customer-value`, `A green product cut must map to customer value, not just implementation activity`).

- [ ] **Step 2: Verify the scorecard gate specifically**

Run: `jini scorecard-gate --format json`
Expected: exit 0. This gate has required-coverage checks; if it names a missing category for the promoted competitors, fill that category before proceeding.

- [ ] **Step 3: Archive the two benchmark docs + kpis yaml** with banner `Replaced by: [golden-competitive-benchmark.yaml](../golden-competitive-benchmark.yaml)`; fix Go refs; append absorbed methodology notes (adapter benchmarking rules) as YAML comments or a `methodology:` block consistent with the file's existing style.

- [ ] **Step 4: Full gate + commit** — `bench: promote free/BYO agents to core set; absorb benchmark method docs (batch 4)`.

### Task 10: Sessions/surfaces merge group

**Files:**
- Move → `specs/archive/`: `memory-system.md`, `learning-system.md` (~5 Go refs), `install-packaging.md` (~1 Go ref), `client-surfaces-and-free-tier.md` (~2 Go refs)

These merge into the PRD itself (already written in Task 3: §4.4 sessions/memory, §4.7 install, §4.9 tiering). So this is archive + banner (`Replaced by: [number-one-platform-prd.md](../number-one-platform-prd.md) §4.4 / §4.7 / §4.9` respectively) + Go-ref repointing to `specs/number-one-platform-prd.md`. Before archiving each file, skim it for any normative rule the new PRD lacks; if found, add the rule to the PRD (settling one-liner required — protected file) rather than losing it. Full gate + commit: `docs: archive session/install/tier specs superseded by PRD (batch 5)`.

### Task 11: Heavy singletons — one commit each

Four files with deep Go entanglement get individual commits so a red gate identifies its culprit (design §8 rule). For each: move + banner + repoint every Go ref + full gate + commit.

- [ ] **Step 1:** `app-platform-shipping-playbook.md` (~27 refs → mostly `publish_readiness.go` and release tests; repoint to `engineering-gate-matrix.md` release tier; any playbook rule the matrix lacks gets absorbed into the matrix in the same commit, settling one-liner included). Commit: `docs: archive shipping playbook into gate matrix (batch 6a)`.
- [ ] **Step 2:** `competitive-release-plan.md` (~18 refs incl. `competitive_release_plan_test.go` — repoint the test to the PRD §3 roadmap + §9 competitive table; rename the test only if its name references the file, keep assertions). Banner → `number-one-platform-prd.md §3, §9`. Commit: `docs: archive competitive release plan (batch 6b)`.
- [ ] **Step 3:** `honest-system-audit.md` (~9 refs). Banner → `number-one-platform-prd.md §5`. Commit: `docs: archive honest-system-audit (batch 6c)`.
- [ ] **Step 4:** `skills-and-delegation-slice.md` (~8 refs). Banner → `number-one-platform-prd.md §4.10`. Any still-true normative skill/agent rule missing from PRD §4.10 gets added to the PRD (settling one-liner). Commit: `docs: archive skills slice superseded by PRD §4.10 (batch 6d)`.

### Task 12: Low-entanglement archive batches

Move → `specs/archive/` with banners, in three batches of ≤10 files, checking each file's refs in the Task 2 map first (expected 0–3 each; repoint or, where a test exists purely to pin a now-archived doc's existence, repoint it to the archive path — existence pins stay valid since nothing is deleted):

- [ ] **Batch 7a — prior PRD generations:** `full-product-prd.md`, `full-product-prd-execution-plan.md`, `product-consensus-prd-and-plan.md`, `cross-surface-session-platform-prd.md`, `cross-surface-session-system-and-dev-design.md`, `number-one-development-plan.md`, `number-one-platform-hld.md`, `number-one-platform-lld.md`, `number-one-product-research.md`. Banner → `number-one-platform-prd.md`. Commit per batch.
- [ ] **Batch 7b — plans/audits/reviews:** `jini-next-initiative-plan.md`, `cli-replacement-score-plan.md`, `docs-homepage-rewrite-plan.md`, `product-streamline-redline.md`, `rewrite-guardrails.md`, `rewrite-score-baseline.yaml`, `delight-gap-closure.md`, `friction-reduction-research.md`, `product-review-roles.md`. Banner → `number-one-platform-prd.md §5`.
- [ ] **Batch 7c — out-of-scope frameworks:** `travel-curated-experience-framework{,-gate,-review}.md`, `workstream-technical-framework{,-gate,-review}.md`, `adaptive-response-rendering-framework{,-gate,-review}.md`, `personal-os.md`, `work-ontology.md`, `work-state-machine.md`, `operating-profiles.md`, `conversation-and-artifact-ux.md`, `artifact-schemas.md`, `atlassian-target-binding.md`, `extension-rules.md`, `launcher-intake-design.md`, `agentic-development-operating-model.md`, `lean-platform-doctrine.md`. Banner → `number-one-platform-prd.md §3 (non-goals)`. Split into two commits if the gate run surfaces >3 ref fixes.
- [ ] **`protocol-core.md` special check:** run `grep -rn 'specs/protocol-core' --include='*.go' internal/ tools/` — if runtime (non-test) code references it, KEEP it in `specs/` and add `Traced to number-one-platform-prd.md §4.1` instead of archiving; note the deviation in the final task's summary.

### Task 13: macOS app docs — roadmap archive

**Files:**
- Move → `specs/archive/`: `macos-app-prd.md`, `macos-app-hld.md`, `macos-app-lld.md`, `macos-app-ux-design.md` (~3 Go refs each)

Banner variant (roadmap, not superseded-forever): `Replaced by: [number-one-platform-prd.md](../number-one-platform-prd.md) §3 roadmap — parked until the desktop-app roadmap item activates; content remains the starting point.` Repoint Go refs; full gate; commit `docs: park macOS app docs as roadmap-stage (batch 8)`.

### Task 14: Regenerate the implementation trace

**Files:**
- Modify: `specs/prd-implementation-trace.md` (+ settling one-liner, protected as of Task 4)

- [ ] **Step 1: Rewrite** with `Updated: 2026-07-07`, mapping the rebuilt PRD's requirements to surfaces and proofs. Two tables:
  1. **Implemented** — carry forward every row of the current trace whose surface/proof still exists (verify each named Go test still exists: `grep -rn '<TestName>' --include='*_test.go' internal/`), re-keyed to new PRD section numbers (§4.1 agent core, §4.3 routing, §4.4 sessions, §4.7 first-run, §7 gates).
  2. **Not yet implemented (v1 backlog)** — honest gap list from the new PRD with no code today; at minimum: savings ledger + receipts (§4.5), token-economy regression suite (§4.6), skills/agents creation (§4.10), throttle fallback UX (§4.3), paywall entitlements (§4.9), BYO auto-detect + keychain (§4.3). Each row names the PRD section and the gate that will eventually prove it. Per the trace's own rule, these rows are explicitly marked not implementation-aligned yet — that is the input to the implementation-rewrite phase, not a claim.
- [ ] **Step 2: Full gate + commit** — `docs: regenerate PRD implementation trace against rebuilt PRD`.

### Task 15: Update CLAUDE.md doctrine + close out

**Files:**
- Modify: `CLAUDE.md`
- Move: `specs/prd-rebuild-design.md` stays in `specs/` (it is the decision record the settling doc cites — do NOT archive it; it's in the protected list)

- [ ] **Step 1: Update `CLAUDE.md`**
  1. Canonical Files section → the Task 4 keep-set (drop archived paths: `product-rewrite-contract.md` stays; `platform-offline-strategy.md`, `local-slm-frontline-policy.md`, `competitive-release-plan.md` lines now point to their merge targets).
  2. Replace the boundary sentence `Free tier may route configured tools and local models; commercial agent/skills OS productivity features stay private.` with: `Free tier may route configured tools and local models, and includes on-the-fly creation of coding-focused skills and agents (PRD §4.10); the commercial repo keeps the productivity-suite/OS feature set.` (Judgment call 5 in the design; decision recorded in the settling doc in Task 3.)
- [ ] **Step 2: Verify no live doc cites an archived path**

```bash
grep -rn 'specs/' --include='*.md' README.md CLAUDE.md CONTRIBUTING.md specs/ --exclude-dir=archive | grep -f <(ls specs/archive/ | sed 's|^|specs/|; s|$|\\b|') || echo "clean"
```

Expected: `clean` (or only banner self-references). Fix any stragglers.
- [ ] **Step 3: Full gate + push-level check + commit**

```bash
bash tools/run_required_gates.sh commit
git add CLAUDE.md
git commit -m "docs: update contributor doctrine for rebuilt PRD and sweep

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
bash tools/run_required_gates.sh push
```

Expected: both gate runs exit 0. The push gate (`jini check ship`) may flag dogfood/smoke evidence freshness — that is pre-existing release posture, not caused by this plan; report it, don't chase it here.

---

## Explicitly out of this plan

- **Model research/selection task** (design Next-steps #3): separate research effort with its own deliverable (curated matrix update for v1.5). Different skill set (research + benchmarking), no repo entanglement with the sweep. Run it as its own plan.
- **Implementation rewrite against the new PRD** (ledger, receipts, skills/agents, throttle fallback, BYO auto-detect, paywall): Task 14's backlog table is its input; it needs brainstorming + its own plan per feature.

## Self-review notes

- Spec coverage: design §1–§7 → Task 3; §8 policy/sequencing → Tasks 4–13; §8 keep-set → Task 4 list; §9 competitor columns → Task 9; trace → Task 14; CLAUDE.md → Task 15; model research + implementation cross-check → explicitly out (stated in design Next steps as separate efforts).
- Files in disposition table not named in any task: none — every keep/merge/archive entry appears in Tasks 4–13; `prd-rebuild-design.md` disposition corrected from "archive" to "keep, protected" (it is the settling doc's cited rationale).
- Type-consistency equivalent here is path-consistency: all banners point at files the keep-set retains; verified against Task 4's list.
