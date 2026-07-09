# Jini PRD Rebuild — Approved Design

Date: 2026-07-06
Status: Design approved in brainstorming dialogue; awaiting written-spec review, then implementation planning.
Supersedes on landing: this design becomes the rewritten `specs/number-one-platform-prd.md` plus an archive sweep of `specs/` (Section 8).

---

## 1. Product thesis & positioning

**One-liner:** *Jini is the coding agent CLI that's free to run — best-in-class local models built in, your own keys and tools when you need more, and receipts for every dollar it saves you.*

**The wedge: token-savings mastery.** Saving tokens is now on the table at every company that uses AI. The tool that masters that art — in every possible way, with receipts — is the hard sell that sells itself. Jini's identity is the agent that treats every token as money: route down before routing up, reuse before regenerate, compact before send, and prove the savings after every task. Incumbent agent CLIs cannot follow — their business is metered tokens.

**Category:** agent CLI (peer of Claude Code/Codex), differentiated on three axes:

- **Free to run.** A first-party local runtime with one curated, benchmarked model per device class. No API key, no subscription, no third-party runtime required for a working coding agent.
- **Routes everything you already have.** BYO provider keys, gateways (OpenRouter, LiteLLM), and real handoffs to installed CLIs (Claude Code, Codex, Gemini CLI, Aider, OpenCode). Escalation from free-local to paid-cloud is a single transparent ladder the user controls.
- **Proves its economics.** Every task produces a cost receipt; the savings ledger totals what Jini saved versus cloud-default behavior, and receipts are share-ready. Free tier shows the money; the paid tier (autopilot + cross-device continuity) acts on it.

Pure routers (LiteLLM et al.) aren't agents; agent incumbents can't sell token frugality. That intersection is the moat.

**The bar:** the audience is the most discerning developer population there is — engineers at FAANG-class companies who use Claude Code and Codex daily. Within the first minute, they must not be able to tell that Jini isn't a first-party tool from a frontier lab. Anything that reads as amateur — verbose ceremony, invented vocabulary, flaky installs, dishonest numbers — is a defect, not a polish item.

## 2. Target user & jobs to be done

**Primary:** the professional developer at a FAANG-class company — already uses Claude Code/Codex daily, hits quota walls and throttles on provisioned tools, works under org token budgets, and often cannot send sensitive code to arbitrary clouds. This user has zero tolerance for jank and compares every interaction against frontier-lab tooling. Jobs: (1) stretch provisioned quotas by sending only hard tasks to paid routes; (2) do real agentic work at zero marginal cost on a local route that never leaves the machine; (3) never lose a session to a throttle, crash, or tool switch; (4) show the org, with receipts, what the routing discipline saved.

**Also served:** the cost-conscious solo developer who can't justify $20–200/mo (same product, the savings matter even more); multi-CLI power dev using Jini as the front door to tools they already pay for; offline/privacy dev using local-only mode.

**Roadmap-stage buyer (not v1):** the company. Token spend is a budget line at every org running AI; team/enterprise capabilities (org-wide savings ledgers, policy routing, quota governance) monetize the same mastery Jini proves to individuals first. Individuals adopting on receipts are the wedge into that sale.

## 3. Goals, non-goals, roadmap

**v1 goals (release-gated, per the Option A scope below):** working coding agent on all routes (local via detected runtimes in v1; first-party runtime is v1.5); TTFV < 5 min on a fresh machine (macOS/Linux in v1; Windows on its release); **token savings as a first-class capability** — every mechanism in §4.6 shipped and measured in the ledger; **throttle resilience on every route** — detection, suggested fallback, and session resume free; predictive avoidance and automatic re-route paid (§4.3, §4.9); savings ledger + shareable receipts; durable sessions + route-outcome learning; **on-the-fly skills and agents creation** (§4.10); live paywall (Autopilot + Continuity) with fail-closed entitlements; task-success benchmark published per route/model/OS.

**Scope decision — RESOLVED: Option A, wedge first** *(user-decided 2026-07-07 after premortem)*:

- **v1 ships:** the agent on BYO keys, gateways, and CLI handoffs, with sessions, receipts/ledger, throttle fallback (detection + suggested fallback + resume), skills/agents creation, and the live paywall. Local execution in v1 rides *detected third-party runtimes* (Ollama etc.) as an explicitly interim route — a deliberate, temporary inversion of §4.2's "never the foundation" rule, disclosed in-product ("local route via your Ollama; Jini's built-in runtime arrives in v1.5").
- **v1.5 ships:** the first-party embedded runtime + curated benchmarked model matrix (§4.2 in full). "Free to run" marketing waits for v1.5; before that, the honest claim is "free with what you already have."
- **Platform order:** macOS + Linux first; native Windows may trail by one release (§4.8 parity gate applies within each release's declared platform set).
- §4.2's requirements are unchanged as written — they gate v1.5, not v1. The model-research task (§ Next steps) starts during v1 so the matrix is ready.

**Non-goal (permanent):** general chatbot/companion use. Conversation exists only in service of finishing work.

**Roadmap (post-v1, each requires a PRD update):** desktop apps (macOS first), mobile continuation surface, team/enterprise capabilities (org savings ledger, policy routing, quota governance — the "companies want token savings" sell), agent orchestration surfaces, vertical workflow packs.

## 4. Product requirements

### 4.1 Agent core

- Bare `jini` opens a compact task prompt; freeform natural language is the primary interface. No dashboard, no command grammar to learn before value.
- Coding-agent behaviors, release-gated: repo-aware context gathering, multi-file edits, running commands/tests, git-aware workflows (diff, commit on approval), iterative fix loops on test failures.
- Simple questions get compact direct answers; no work artifacts, no ceremony.
- Ambiguous requests fail closed with the exact ambiguity to resolve (e.g., candidate filenames), never a generic draft.
- Side effects follow an approval matrix: reads and workspace edits are direct; commits, pushes, deletions, network sends, and anything destructive require visible approval. Every side-effecting task returns a receipt (files changed, commands run, route used, cost, rollback path).
- Output style is Claude Code/Codex-familiar: answer or action receipt first, no invented vocabulary.

### 4.2 Local execution (the "free to run" pillar)

- First-party inference runtime built into the CLI — no Ollama, no LM Studio, no external daemon required. (Engine choice — llama.cpp-class embedded, MLX on Apple Silicon, CUDA/DirectML elsewhere — is HLD scope.)
- Curated model matrix: exactly one recommended, benchmark-proven coding model per device class (RAM/GPU/OS envelope), maintained in the repo with published task-success scores. Alternates allowed as opt-in; the default is opinionated.
- Model download is consent-based at first run, resumable, checksummed. No silent downloads; no model bundled in the installer.
- Third-party runtimes users already have (Ollama etc.) are detectable as optional extra routes — never the foundation, never required.
- **Model licensing & distribution:** the matrix admits only permissively-licensed, redistribution-safe models (Apache-2.0/MIT-class), selected by documented research and scoring published alongside the benchmark. The model's license is displayed at the download-consent step. Weights are fetched from the model's official distribution source with checksum verification — Jini does not mirror or re-host weights; that third-party availability dependency is accepted and stated. A device class where no model clears both the license bar and the quality floor gets **no recommended model** rather than a compromised one.

### 4.3 Routing & escalation

- One route ladder, always visible: **local model → BYO provider key → gateway → installed CLI handoff**. `jini route` lists, sets, pins, and explains.
- Auto mode picks the cheapest route that clears the task's quality bar, using benchmark data + learned route outcomes; the user can pin anything.
- Route truth-in-labeling: a route named `claude-code` invokes that installed CLI or fails closed with setup guidance. Provider APIs are never disguised as CLIs.
- Escalation is transparent: when local can't clear the bar, Jini says why and what the next rung costs before spending the user's money. (Approved judgment call: the extra confirmation free users see is the point — no silent spend, ever.)
- Throttle/quota errors on any route trigger a suggested fallback (free) or automatic re-route + resume (paid).
- **BYO setup is near-zero effort.** Adding a provider, gateway, or CLI route never requires editing a config file. `jini route add` (or the first-run prompt) accepts a pasted key, validates it with a live test call, and confirms the route is working — one paste, one confirmation, done. Jini auto-detects what's already on the machine: provider env vars (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, …), existing CLI configs (Claude Code, Codex, Gemini CLI), and gateway setups, and offers each detected credential as a ready route rather than asking the user to re-enter it. Keys are stored in the OS keychain, never in plaintext dotfiles. A failed validation states exactly what's wrong (bad key, no quota, wrong region) — never a generic error.

### 4.4 Sessions & memory

- Every task lives in a durable session: resumable after crash, throttle, reboot, or route switch without transcript replay — compact state, not chat history.
- **Resume honesty:** same-route resume is lossless. Cross-route resume (e.g., CLI handoff → local model) is best-effort by construction — different routes carry different context fidelity — and says so: the resume receipt states exactly what carried over and what didn't. Silent context loss is a defect.
- Route-outcome learning: Jini records which route/model succeeded/failed for which task shape on this device and feeds that into auto-routing. Fully inspectable (`jini memory`), local, exportable, deletable.
- No personal profile building in v1.

### 4.5 Savings ledger & shareable proof (the growth engine)

- Every task produces a cost receipt denominated three ways: **dollars** (the headline — dollar savings spread fastest), **tokens** kept off metered routes, and **time** saved (throttle waits avoided, no tool switching). The counterfactual ("this would have cost $X on <default cloud route>") backs the dollar figure.
- **Literal vs imputed, always labeled:** on metered API routes, dollar savings are literal at the provider's posted prices. On subscription routes the user pays the same flat fee regardless — savings there are imputed (quota preserved, valued at API-equivalent prices) and labeled as such; the honest subscription story is quota stretched, limit-hits avoided, and tier upgrades not needed. Mixing the two without labels is the fastest way to get the math publicly debunked.
- The ledger aggregates per repo/week/month; `jini savings` renders it, `jini savings --share` produces a shareable card (image/markdown) with zero private data — numbers only, opt-in.
- Ledger methodology is documented and honest: counterfactuals are labeled estimates. (Approved judgment call: an inflated savings claim is a release-blocking defect — the growth loop dies if the numbers aren't credible.)

### 4.6 Token economy (the mastery discipline)

Every token-saving mechanism is a named, testable requirement — this is the product's craft, not an implementation detail:

- **Route-down bias:** auto mode prefers the cheapest route that clears the quality bar; escalation is the exception path, never the default.
- **Context frugality:** minimal repo context per task — targeted reads over directory dumps; no re-reading unchanged files within a session.
- **Session reuse:** durable compact state means resume and follow-up tasks never replay transcripts or re-derive context.
- **Compact I/O:** prompt and response shapes are token-budgeted; simple prompts must not generate artifacts, plans, or preambles.
- **Quota stretching:** paid routes receive only the work that needs them; the receipt shows what stayed local.
- **Measured, not asserted:** each mechanism's contribution is visible in the ledger, and a token-efficiency regression (same task, materially more tokens) is a gate failure like any UX regression.

### 4.7 First-run experience

- Install: one command per OS (curl script; brew/winget as fast follows), signed/notarized binaries, no source builds.
- First run detects device class + anything already configured, then offers: (a) provision the curated local model, (b) paste a provider key, (c) use an installed CLI. Detected credentials and installed CLIs are pre-offered as ready routes (§4.3) — a user with `ANTHROPIC_API_KEY` set or Claude Code installed should be one confirmation away from a working route.
- **TTFV gate, stated honestly:** install start → first completed task in **under 5 minutes** on the fastest path available to that user, measured on all 3 OSes. The local-model path does not pretend a multi-GB download fits that budget: size and ETA are shown before consent, the download is resumable and runs in the background (a detected BYO/CLI route can serve the first task meanwhile), and the TTFV methodology states explicitly that download time is excluded. Hiding the download inside the metric would violate §5.

### 4.8 Surfaces & platforms

- v1: terminal CLI on macOS (Apple Silicon + Intel), Linux (x64/arm64), Windows (x64, native — not WSL-only). Feature parity across OSes is release-gated; the model matrix covers all three.

### 4.9 Free/paid boundary

- **Free, permanently:** full agent, local execution, all manual routing, BYO keys, gateways, CLI handoffs, sessions + resume, route learning, full ledger + shareable receipts, diagnostics. Free is a complete product, never a crippled trial.
- **Paid (live at launch):** *Autopilot* — predictive throttle avoidance, automatic route/model switching, auto-resume, savings optimization. *Continuity* — account-backed session sync across devices. Both fail closed without entitlement; manual free equivalents always exist.
- **Upgrade trigger, stated honestly:** pay when route management interrupts you more often than the subscription costs. Prices/SKUs live in the commercial repo.
- **Skills/agents tier placement:** creation and use of coding-focused skills and agents (§4.10) is free — it's competitive parity with Claude Code, not a premium feature. The commercial repo retains the productivity-suite/OS feature set. This narrows the previous "all agent/skills features are commercial" doctrine; the repo guidance (`CLAUDE.md`) must be updated in the same change that lands this PRD.

### 4.10 Skills & agents, created on the fly

- Users create skills and agents from inside a session, in natural language: "turn what we just did into a skill" or `jini skill new` produces a reviewable, editable skill file — no scaffolding ceremony, no docs detour. Same for focused agents (`jini agent new`): a named role + instructions + tool permissions.
- Skills and agents are plain, portable files (markdown + frontmatter) stored in the repo or user scope, inspectable and diffable like any other source. Jini reads the formats developers already have where practical (Claude Code-style skills/agents) rather than inventing a new one.
- Created skills are immediately invocable in the same session; agents are dispatchable as sub-tasks with their own tool permissions, subject to the same approval matrix (§4.1) and route ladder (§4.3) — an agent run produces the same receipts as any task.
- Skill/agent creation is free tier (§4.9). Token economy applies: invoking a skill must be cheaper than pasting its instructions inline every time — measured as the skill's token footprint versus the instruction text it replaces.
- **Agents obey the token economy:** dispatch defaults to the cheapest qualifying route; fan-out beyond a single agent requires the same visible cost preview as escalation (§4.3); every agent run itemizes its tokens on the parent task's receipt. Multi-agent orchestration is not an expectation of the local device class — an agent that needs a paid route says so before spending.

## 5. Quality & evidence standards

- **No claim without proof.** Any release-readiness, route-support, or capability claim must trace to a runnable artifact: a gate command, a Go test function, a golden transcript, or a benchmark score. A gate name without a runnable command or named proof reference is planning prose, not evidence.
- **Golden transcript parity.** First-minute behavior is held to Claude Code/Codex expectations via golden transcripts: compact answers to simple questions, direct file edits, no draft/status frames, no invented vocabulary. The 100-prompt first-minute bank remains a commit-gate fixture. The acceptance standard is the FAANG-engineer bar from §1: parity is the floor, not the target.
- **Benchmark honesty.** The task-success benchmark per route/model/device class is published with methodology; scores come from the repo's harness, not vendor claims. The curated model matrix cannot recommend a model whose published score is stale (more than one release old) for its device class.
- **Savings-math honesty.** Counterfactual savings are labeled estimates with documented methodology. An inflated or misleading savings number is a release-blocking defect, same severity as a security issue.
- **Token frugality is P0** on Jini's own behavior (see 4.6); efficiency regressions gate like UX regressions.
- **Security posture:** signed/notarized binaries, secret/dependency scanners wired and gate-protected (CodeQL, govulncheck, OSV-Scanner, TruffleHog, Dependabot), and the public-repo boundary enforced — commercial pricing/entitlement internals never land in this repo.

## 6. Success metrics

**North star:** weekly tasks completed at zero marginal cost per active user — measures the free-to-run pillar actually working, not installs.

| Metric | v1 bar | Rationale |
|---|---|---|
| TTFV (install start → first completed task) | < 5 min, all 3 OSes (download excluded, stated — §4.7) | The release-gating adoption metric |
| First-task success rate on fresh install | release-gated, measured in qualification | The actual survival metric — TTFV means nothing if task one fails |
| Task success rate, local route | ≥ absolute floor per device class, set in the benchmark spec (e.g., ≥70% of the golden coding-task suite); no model clears the floor → no recommended model for that class | "Free to run" must mean "works"; the floor is fixed, not whatever we publish |
| Session resume success (crash/throttle/reboot) | ≥ 99% | The continuity promise |
| Local-route share of completed tasks | ≥ 50% for local-provisioned users | Escalation ladder is honest, not a cloud funnel |
| Receipt share rate | measured, no gamed target | Growth-loop signal; must stay opt-in and honest |
| Free→paid conversion on Autopilot/Continuity | measured post-launch | Paywall validation without corrupting free tier |

Instrumentation is local-first: metrics come from opt-in diagnostics and the ledger, never silent telemetry. Weaker measurement is the accepted cost.

## 7. Release gates

Three-tier structure (commit / push / release) carries forward from `engineering-gate-matrix.md`, with the matrix rewritten to trace to this PRD's requirement numbers. On conflict, this PRD wins.

- **Commit:** `go test ./...`, whitespace/patch checks, security-configuration gate, PRD-drift gate (repointed at this document), customer-value gate, CLI UX regression gate, Claude/Codex use-case gate, scorecard gate.
- **Push:** all commit gates + `jini check ship` with dogfood and smoke evidence for every claimed route.
- **Release (new bars from this PRD):** TTFV measured < 5 min on fresh macOS/Linux/Windows machines from signed artifacts; first-task success measured in release qualification; golden transcripts green; task-success benchmark published for every recommended model, each clearing its device-class floor (§6); model license verification for every matrix entry (§4.2); savings-methodology audit passed, including literal-vs-imputed labeling (§4.5); paywall fail-closed test (no entitlement → paid features cleanly absent, free equivalents work); model download consent/checksum/resume verified; token-efficiency regression suite green (§4.6).

## 8. Superseded documents — archive sweep

### Policy

- One canonical PRD: this design lands as the rewritten `specs/number-one-platform-prd.md`.
- Every spec not in the keep-set moves to `specs/archive/` with a superseded banner naming the PRD section that replaced it. Nothing is deleted.
- Gate scripts (`tools/product_prd_drift_gate.sh` and friends) reference spec files by path; **every move must update the referencing gates in the same change**, verified by running the full commit gate. Archived files must no longer be citable as requirements.

### Disposition table

**Keep — canonical, rewritten to trace to this PRD:**

| File | Role |
|---|---|
| `number-one-platform-prd.md` | Rewritten as this PRD |
| `engineering-gate-matrix.md` | Gate contract, retraced to PRD §7 |
| `golden-competitive-benchmark.yaml` | Competitive bar fixture (§5) |
| `local-model-support-matrix.md` | Curated model matrix (§4.2) |
| `execution-routing-policy.md` | Routing spec (§4.3); absorbs merges below |
| `public-repo-boundary.md` | Free/commercial boundary (§5) |
| `canonical-names.md` | Naming contract |
| `product-rewrite-contract.md` | CLI behavior contract (§4.1), gate-protected |
| `product-settling-decisions.md` | Decision log required by drift gate |
| `prd-implementation-trace.md` | Regenerated against this PRD |
| `claude-codex-prompt-bank.jsonl` | Gate fixture (§5) |
| `open-source-prompt-validation.yaml` | Gate fixture (§5) |
| `dogfood-personas.yaml` | Push-gate evidence fixture (§7) |

**Merge, then archive the source** (content folds into a keep-set doc; source gets superseded banner):

| File(s) | Merge target |
|---|---|
| `runtime-execution-modes.md`, `runtime-selection-heuristics.md`, `device-capability-routing.md`, `research-informed-heuristics.md` | `execution-routing-policy.md` |
| `platform-offline-strategy.md`, `local-slm-frontline-policy.md`, `device-runtime-gate.md` | `local-model-support-matrix.md` |
| `adapter-benchmark-gate.md`, `adapter-capability-benchmarking.md`, `competitive-kpis.yaml` | benchmark methodology alongside `golden-competitive-benchmark.yaml` |
| `dogfood-gates.md`, `lean-platform-gate.md`, `friction-reduction-gate.md`, `engineering-principles.md` | `engineering-gate-matrix.md` |
| `memory-system.md`, `learning-system.md` | PRD §4.4 |
| `install-packaging.md` | PRD §4.7 (operational detail stays as engineering appendix if needed) |
| `client-surfaces-and-free-tier.md` | PRD §4.8–4.9 |

**Archive — superseded PRD generations, plans, audits, and out-of-scope frameworks:**

- Prior PRDs/plans: `full-product-prd.md`, `full-product-prd-execution-plan.md`, `product-consensus-prd-and-plan.md`, `cross-surface-session-platform-prd.md`, `cross-surface-session-system-and-dev-design.md`, `number-one-development-plan.md`, `number-one-platform-hld.md`, `number-one-platform-lld.md` (HLD/LLD regenerate from the new PRD when planning starts), `number-one-product-research.md`, `jini-next-initiative-plan.md`, `cli-replacement-score-plan.md`, `competitive-release-plan.md`, `docs-homepage-rewrite-plan.md`, `product-streamline-redline.md`, `rewrite-guardrails.md`, `rewrite-score-baseline.yaml`
- One-shot audits/reviews: `honest-system-audit.md`, `delight-gap-closure.md`, `friction-reduction-research.md`, `product-review-roles.md`
- Out-of-scope frameworks (PRD §3 non-goals): `travel-curated-experience-framework{,-gate,-review}.md`, `workstream-technical-framework{,-gate,-review}.md`, `adaptive-response-rendering-framework{,-gate,-review}.md`, `personal-os.md`, `work-ontology.md`, `work-state-machine.md`, `operating-profiles.md`, `conversation-and-artifact-ux.md`, `artifact-schemas.md`, `atlassian-target-binding.md`, `launcher-intake-design.md`, `skills-and-delegation-slice.md`, `agentic-development-operating-model.md`, `app-platform-shipping-playbook.md`, `lean-platform-doctrine.md`, `extension-rules.md`, `protocol-core.md` (verify no runtime doc references before archiving)
- Roadmap-stage (revived when roadmap activates): `macos-app-prd.md`, `macos-app-hld.md`, `macos-app-lld.md`, `macos-app-ux-design.md`

Any file whose gate or code references cannot be cleanly repointed gets flagged during planning rather than force-archived.

### Sweep sequencing (mandatory — no mega-change)

The drift gate exists to block exactly the kind of change this sweep is; one giant diff would force weakening the guard at the moment it matters most, and a mid-sweep gate failure would be undiagnosable. Each step lands green through the full commit gate before the next starts:

1. Update `product-settling-decisions.md` and land the new PRD text — old files untouched, gates still pointing at them.
2. Repoint gate scripts to the new canonical set while the old files still exist.
3. Archive in small batches (5–10 files); each batch passes the full commit gate, so a red gate identifies its culprit immediately.
4. Regenerate `prd-implementation-trace.md` last, against the settled state.

## 9. Risk register & competitive response

Premortem-derived. Each risk names its mitigation in this document; a risk without a mitigation is a scope decision, not a footnote.

| Risk | Mitigation |
|---|---|
| Scope death — everything release-gated, nothing ships | Open v1 scope decision (§3 options); sweep sequencing (§8) |
| First task fails on the local model; discerning user uninstalls in minute one | First-task success is release-gated (§6); absolute quality floor per device class — no model clears it, none is recommended (§4.2, §6) |
| Savings math publicly debunked | Three-denomination receipts; literal-vs-imputed labeling; published methodology; inflated claims are release-blocking (§4.5, §5) |
| Model licensing takedown | Permissive-license-only matrix, official-source downloads, license shown at consent (§4.2) |
| Distribution cost / third-party dependency | No mirroring; official sources + checksums; dependency accepted and stated (§4.2) |
| Thin paywall → no revenue | Accepted for v1; enterprise roadmap (§2, §3) is the revenue thesis; conversion measured honestly (§6) |
| Cross-route resume loses context silently → trust incident | Resume honesty rule: receipts disclose fidelity loss; silent loss is a defect (§4.4) |
| Windows/Intel parity holds release hostage | Flagged in v1 scope decision (§3); parity gate scope is part of Option A/B |
| **Incumbent response: a lab ships subsidized local/offline mode** | Jini's position is vendor-neutrality: it routes *all* vendors' tools, including that one — a single-vendor local mode becomes one more rung on Jini's ladder, and the receipts/ledger remain the part nobody whose business is metered tokens will copy |
| Existing free/BYO agents (Cline, Roo, Aider, Goose, OpenCode) close the gap | Differentiation is the *combination*: first-party curated local runtime + real CLI handoffs + savings receipts + throttle resilience; tracked in `golden-competitive-benchmark.yaml`, which must add these tools as scorecard columns |

## Approved judgment calls

1. Escalation quotes the next rung's cost before spending money — the extra confirmation free users see is deliberate.
2. Counterfactual savings math ships with the honesty rule: inflated savings claims are release-blocking defects.
3. North-star metric is weekly zero-marginal-cost tasks per active user; instrumentation is local-first/opt-in, accepting weaker measurement over silent telemetry.
4. Token-savings mastery is the explicit product identity (§1, §4.6) and the future enterprise sell (§2, §3 roadmap).
5. On-the-fly skills/agents creation (§4.10) is a core v1 goal and placed in the **free** tier as competitive parity; the commercial repo keeps the productivity-suite/OS features. This narrows the prior "agent/skills features are commercial" doctrine and requires a `CLAUDE.md` update when the PRD lands. (Flagged for user veto.)
6. Dollar savings is the primary receipt denomination (it spreads fastest), always shown alongside tokens and time saved, with literal-vs-imputed labeling: metered API savings are literal; subscription-route savings are imputed at API-equivalent prices and labeled (§4.5). *(User-decided after premortem.)*
7. Curated matrix is permissive-license-only, selected by documented research/scoring, downloaded from official sources with license shown at consent — no mirroring (§4.2). *(User-approved.)*
8. The archive sweep executes as a gated four-step sequence, never one mega-change (§8). *(User-approved.)*
9. v1 scope is **Option A — wedge first**: BYO/gateway/CLI-handoff agent + sessions + receipts + throttle fallback + skills/agents + paywall now, with third-party local runtimes as a disclosed interim route; first-party runtime + curated matrix in v1.5; native Windows may trail one release. *(User-decided 2026-07-07.)*
10. Post-premortem hardening: TTFV excludes model download and says so (§4.7); local quality bar is an absolute per-device-class floor, not self-referential (§6); first-task success is release-gated (§6, §7); cross-route resume discloses fidelity loss (§4.4); agents obey the token economy with cost previews (§4.10); risk register added (§9).

## Next steps

1. ~~User decides the v1 scope~~ — resolved: Option A (§3).
2. Implementation plan (writing-plans): rewrite `number-one-platform-prd.md` from this design, execute the merge/archive sweep per the §8 sequencing, regenerate `prd-implementation-trace.md`, update `CLAUDE.md` (skills/agents tier doctrine, judgment call 5), add the missing competitor columns to `golden-competitive-benchmark.yaml` (§9), then cross-check implementation against the new PRD section by section.
3. Model research task (Option A/B both need it eventually): thorough find/score/select pass over permissively-licensed coding models per device class, producing the curated matrix with published scores (§4.2).
