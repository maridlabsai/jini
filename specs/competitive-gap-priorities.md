# Competitive Gap Priorities

Status: product-strategy prioritization. **This is not an empirical benchmark and
not a claim that Jini currently matches or outperforms any competitor.** It ranks
where Jini should invest next and ties each gap to a recovery gate. Competitor
descriptions are product observations, not measurements.

Canonical product thesis stays in [number-one-platform-prd.md](./number-one-platform-prd.md);
this file refines *implementation priority*, not the thesis. The R1 contract it
points to is [r1-safe-execution-contract.md](./r1-safe-execution-contract.md).

## Weighting

| Factor | Weight |
| --- | --- |
| Developer adoption impact | 30% |
| Frequency of developer pain | 20% |
| Jini differentiation potential | 20% |
| Trust / safety risk | 20% |
| Implementation leverage | 10% |

## Ranked gaps

### 1. Safe autonomous execution and rollback — 98/100 — JINI-R1
- **Current Jini state:** NOT shipped as safe. The autonomous `run_command` shell tool runs `sh -c` with no workspace confinement, no command classification, no network policy, full environment inheritance (credentials), and no rollback; `edit_file` writes in place with no snapshot. Gated behind Auto + per-dir `trust --autonomous`, but unmediated once reached. (STEP 54/54A audit.)
- **Developer consequence:** an autonomous run can touch files outside the workspace, reach the network, read credentials, and leave ambiguous partial state on interruption.
- **Competitor capability creating the gap:** Cursor/Kiro/Claude Code/Codex offer bounded agentic edit+run loops users already trust for routine changes.
- **Jini differentiation intent:** one enforceable mediation boundary + per-run write-set snapshot rollback + evidence — provider-neutral, user-owned.
- **Prerequisite gates:** none (foundational).
- **Status:** planned (R1).

### 2. Durable journal, crash recovery, and true resume — 93/100 — JINI-R2
- **Current Jini state:** NOT shipped. Throttle-survival holds/self-resumes in-process, but there is no durable, crash-resumable, idempotent journal.
- **Developer consequence:** a crash or kill mid-task can lose progress or risk repeating non-idempotent mutations.
- **Competitor capability:** background/async agents that survive restarts and resume.
- **Jini differentiation intent:** durable, portable task journal with idempotent resume, built on R1's evidence events.
- **Prerequisite gates:** R1 (its transaction/evidence events feed the journal).
- **Status:** planned (R2).

### 3. Model choice, speed modes, Auto routing, and throttle failover — 91/100 — JINI-R3 (+ later JINI-R7 outcome evidence)
- **Current Jini state:** partial. Multi-provider routing and BYO exist; ready-fallback *surfacing* exists; **automatic throttle-driven provider switching is NOT shipped on the free tier** (hold+self-resume + manual `jini route set`; auto-switch is paid Autopilot).
- **Developer consequence:** on throttle the free tier waits or needs a manual switch; no automatic failover.
- **Competitor capability:** Cursor/Kiro expose model/speed choice with smoother routing.
- **Jini differentiation intent:** provider-neutral capability-based routing with honest, user-visible route receipts; outcome learning deferred to R7.
- **Prerequisite gates:** R1 (safe execution), R2 (durable state for resume-across-switch).
- **Status:** partial / planned.

### 4. Mandatory verification before declaring success — 88/100 — safety minimum pulled into R1; depth remains JINI-R4
- **Current Jini state:** NOT enforced. Success is not gated on verification evidence; an exit-zero model/subprocess can be reported as success.
- **Developer consequence:** "done" can be wrong — the core trust failure the recovery program targets.
- **Competitor capability:** test-aware loops that re-run checks before claiming done.
- **Jini differentiation intent:** a minimum "mutating success requires evidence" invariant in R1 (`completed_unverified` otherwise); deeper verifier intelligence in R4.
- **Prerequisite gates:** R1 (the safety-minimum slice); R4 (depth).
- **Status:** planned (R1 minimum; R4 depth).

### 5. Control surfaces and first-task UX — 79/100 — JINI-R9
- **Current Jini state:** CLI-first; no IDE-grade inline feedback; no local status/control surface beyond the CLI (the `jini app serve --stdio` + macOS/Tauri shell are prototypes, not shipped surfaces).
- **Developer consequence:** slower, less legible first-task experience vs editor-native tools; no portable way to watch/approve work from another surface.
- **Competitor capability:** Cursor/Kiro's in-editor latency and feedback; Factory's App/CLI/managed-compute surface breadth.
- **Jini differentiation intent:** fast, legible terminal-first first-task UX plus a frugal, local-first control plane reusing one protocol across surfaces; IDE/remote bridges later. See [frugal-control-plane-roadmap.md](./frugal-control-plane-roadmap.md).
- **Prerequisite gates:** R1–R4 (safe, verified core first).
- **Status:** planned (R9).
- **Decomposition (same 79 score — these are sub-items of R9, not newly ranked gaps):**
  - **R9A** — control protocol and local status.
  - **R9B** — responsive browser/PWA control surface.
  - **R9C** — thin desktop shell reusing the protocol/PWA.
  - **R9D** — native mobile only if PWA evidence proves a material unmet need.
- **Boundary:** only the portable schema **seams** enter R1/R2 (R1A emits the records; R2 adds the durable journal); all UI and network implementation remains R9. Factory's surface breadth is a product observation and **does not reorder the top four** (R1 98 → R2 93 → R3 91 → R4-min 88); no score changes from it.

### 6. MCP / tool-ecosystem invocation — 66/100 — JINI-R11
- **Current Jini state:** listing-only (`jini mcp list`); **invocation NOT supported** (JINI-R0 wording).
- **Developer consequence:** cannot drive MCP tools/servers from Jini yet.
- **Jini differentiation intent:** safe, mediated MCP invocation after the execution core is trustworthy.
- **Prerequisite gates:** R1 (mediation) at minimum.
- **Status:** planned (R11).

## Required honesty statements
- Jini does **not** currently match Cursor or Kiro in model/speed/Auto flexibility.
- Jini does **not** currently provide safe autonomous shell execution.
- Jini does **not** currently provide true durable resume.
- Automatic throttle-driven provider switching is **not** currently shipped (free tier).
- MCP invocation is **not** currently supported.
- Jini's potential advantage is an open, provider-neutral control plane that uses capacity the developer already has, with durable state and evidence — **but that is an intent, not a shipped claim, until the gates pass.**

## Competitor weaknesses as product opportunities (observations, not absolute claims)
- **Cursor:** editor coupling and opaque routing decisions.
- **Kiro:** heavier specification/workflow ceremony and autonomy complexity.
- **Claude Code:** predominantly Anthropic-model-centric; no native multi-provider failover.
- **Codex:** predominantly OpenAI-model-centric.
- **Cross-product opportunity:** portable, user-owned route and verification receipts.

## Recovery order
R1 → R2 → R3 remains mandatory: safe execution (R1) is the foundation; durable resume (R2) builds on R1's evidence; routing/failover (R3) needs both. The existing order is correct; the only refinement is pulling the R4 *minimum* verification invariant into R1 as a safety requirement (depth stays R4).
