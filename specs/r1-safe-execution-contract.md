# JINI-R1 — Safe-Execution Contract

Status: **NORMATIVE, NOT YET IMPLEMENTED.** This contract specifies what R1 must
build. Nothing here describes current behavior; the current state is the STEP
54/54A audit (unmediated autonomous shell, no rollback, no verification gate).
Priority rationale: [competitive-gap-priorities.md](./competitive-gap-priorities.md).

## 1. Objective
R1 establishes **one enforceable mediation boundary** for every agent-controlled
side effect. R1 does **not** claim complete production readiness or general-purpose
OS isolation.

## 2. Execution modes

**PLAN** — read-only. No task-controlled file writes, subprocesses, network,
Git mutations, credential access, external-agent handoffs, or self-update.
Produces a proposed plan only. (Behavior-compatible with today's untrusted/plan path.)

**SUPERVISED** — side effects require a concrete preview and explicit approval.
Approval binds to the exact operation: capability, argv, cwd, environment
capability set, network capability, and expected write scope shown. Materially
changed operations require fresh approval. Approved operations still pass through
the mediator and evidence capture.

**AUTO** — **default deny.** Permits only operations backed by an active, tested
mediation/sandbox capability. Raw shell, external-CLI mutation, unrestricted
network, self-update, Git mutation, outside-workspace access, and unbounded
environment inheritance **fail closed** unless a later contract + tested backend
authorize them. Until such a backend exists, **`run_command` via `sh -c` is
unavailable in Auto.** Trust grants *eligibility* to request Auto; it never
bypasses mediation.

## 3. Privilege and configuration
- Repository-controlled files **cannot** elevate trust.
- Environment variables **cannot** elevate execution privilege.
- `JINI_MODE` may **reduce** privilege but must never raise it above the
  persisted/user-selected posture.
- Per-directory autonomous trust is **necessary but not sufficient**.
- Resumed/retried work must **re-evaluate** current trust and policy.
- Test overrides must be unreachable through production input.

## 4. Central mediator
Every agent-controlled operation passes through one conceptual side-effect
mediator. **Operation classes:** read, mediated file write, structured
subprocess, raw shell, network, credential access, Git mutation, external-agent
handoff, self-update.

For each request the mediator records/evaluates: run/operation ID; mode;
requested capability; canonical workspace root; target paths; executable+argv;
cwd; environment allowlist; network policy; expected write scope;
timeout/cancellation; preview/approval evidence; transaction/snapshot state;
verification requirement; final outcome. **Unknown operation types fail closed.**

Internal Jini bookkeeping (ledgers, config) is **not** automatically a
task-controlled workspace mutation, but must still use safe atomic persistence
and must never become a bypass for agent-provided paths or commands.

## 5. Filesystem policy
- Canonical workspace root established once per run.
- Relative task paths only; no absolute or parent traversal.
- No symlink-following write along the **complete** target path (every component, not only the final element); **hard-linked mutation targets fail closed in Auto unless proven safe.**
- No write across Git submodule/worktree boundaries without separately granted scope.
- Platform-correct for macOS, Linux, Windows.
- Race-resistant, descriptor-relative access where supported; **no
  validate-path-then-open TOCTOU sequence.**
- Atomic temp-file + replace for updates; preserve appropriate mode/metadata.
- Explicit create/update/delete operation types; every target declared before mutation.
- Unsupported secure-path primitives **fail closed in Auto**.
- A descriptor-relative root primitive (e.g. `os.Root`) does **not** by itself prevent mount-point/filesystem-boundary traversal or special-file (device/FIFO/socket) access and is **not** a complete sandbox; Auto mutation must **reject** unsupported boundary cases and non-regular (special) files.
- The descriptor-relative confinement above assumes a **patched toolchain**: R1's implementation must require Go ≥ 1.26.5 within the 1.26 line (earlier 1.26 patches are affected by GO-2026-4970), enforced by an **executable gate** rather than a `toolchain` directive alone. The exact go.mod + gate mechanism is settled at implementation; this contract does not change go.mod.
- `confinePath` as it exists today does **not** satisfy this contract (TOCTOU + unproven hard-link/worktree/case-insensitive/Windows handling).

## 6. Subprocess / environment policy
- **Scrub environment by default**; start from an explicit safe allowlist; never inherit all credentials.
- Secrets supplied only through an explicitly declared capability; **redact** secrets from previews, logs, errors, receipts.
- Structured executable + argv preferred over shell. **In Auto, every subprocess — structured or raw shell — fails closed; R1's Auto initially supports only mediated filesystem operations, and a subprocess may run in Auto only if an active tested sandbox/isolation backend proves its capabilities (scrubbed env, structured argv, or classification do not by themselves make an Auto subprocess safe).**
- cwd alone is not confinement; command classification cannot make arbitrary shell safe.
- Network denied unless explicitly declared and allowed.
- Timeouts and cancellation mandatory; **child-process-tree termination required.**
- Expected write scope captured before execution.
- If no tested sandbox backend satisfies the requested capabilities, Auto **fails closed and offers Supervised**.
- No universal mechanism is mandated (not `sandbox-exec`/Landlock/bubblewrap/Docker/a Windows API). Define a **backend capability interface**; each backend must **prove** its guarantees.

## 7. Transaction and rollback policy
Use **per-run write-set snapshots, not `git stash`.**
- Capture pre-run Git HEAD + porcelain status.
- Preserve pre-image bytes + relevant metadata before each mediated mutation; record before/intended-after hashes.
- Atomic writes; multi-file commit/rollback semantics.
- On failure/timeout/cancellation/failed verification → attempt rollback.
- Roll back **only when the current target still matches the agent-produced after-hash**; concurrent user modification → **conflict stop, never overwrite**.
- Preserve pre-existing dirty and untracked user files.
- **Never** `git reset --hard`, `git clean`, or implicit stash.
- Transaction metadata suffices for R1 rollback but is **not** the full R2 durable journal.
- Incomplete transaction state must be **detectable and honestly reported**.
- Transaction artifacts (snapshots, metadata, evidence) persist **outside the repository** under a permission-protected `JINI_HOME` safety area, organized by workspace identity and run ID: directories owner-only, snapshot/evidence files owner read/write only. They are **not** an OS temporary directory unless that path is explicitly rooted under the protected `JINI_HOME`.
- Artifacts **exclude or redact secrets**. Failure/recovery artifacts are retained per explicit policy; successful-run artifacts may be compacted or removed only after verified completion.

## 8. Minimum verification pulled into R1
- A mutating task **cannot** report "success" solely because a model/subprocess exited zero.
- Required minimum: mutation-integrity checks; expected-write-scope comparison; repository-state comparison; `git diff --check` when inside Git; configured focused verification when available.
- If no meaningful verifier exists → final state `completed_unverified` (not success).
- Failed verification → rollback when safe; if rollback conflicts with concurrent user work → `recovery_required`.
- Deeper test selection, evidence quality, and mandatory project verification remain **JINI-R4**.

**Terminal outcomes:** `planned`, `denied`, `approved`, `completed_verified`,
`completed_unverified`, `failed_rolled_back`, `failed_partial`,
`recovery_required`, `cancelled_rolled_back`, `cancelled_partial`.

## 9. Compatibility / fail-closed during R1 migration
- Existing Plan/read-only behavior remains available.
- `edit_file` and `run_command` may **not** execute in Auto until mediated.
- Mutation-capable CLI handoff is **unavailable in Auto** until a sandboxed contract exists.
- Verification commands are structured, env-scrubbed, timed, mediated.
- `jini update` stays an explicit administrative command **outside** task Auto.
- Keychain/provider configuration stays explicit, **outside** task Auto.
- Read-only system/Git probes may stay outside the mutation transaction, but their arguments must not be agent-controlled unsafe input.
- Unsupported platform guarantees **fail closed** rather than silently degrade.

## 10. Explicit non-goals (deferred)
Crash-resumable full journal / idempotent resume → **R2**; model/speed/Auto route
selection + throttle failover → **R3**; comprehensive verification intelligence →
**R4**; route outcome learning → **R7**; IDE/first-task UX → **R9**; proof receipts
beyond R1 safety evidence → **R10**; MCP server invocation → **R11**;
unrestricted cross-platform arbitrary-shell autonomy → not in R1.

## 11. Acceptance gates (required tests)
Dispatch/integration tests for all modes; attempted trust/mode escalation;
shell-unavailable-in-Auto; supervised exact-operation approval; environment
secret scrubbing; network denial; absolute/traversal/symlink/TOCTOU/hard-link/
worktree/submodule cases; macOS/Linux/Windows path semantics; atomic single- and
multi-file mutation; timeout/cancellation + child cleanup; disk/write/verification
failure rollback; concurrent-user-edit rollback conflict; dirty/untracked
preservation; expected-write-scope violation; verification-required terminal
outcomes; external-handoff fail-closed; **direct lower-level mediator-bypass tests.**

**R1 is not complete until production-reachable dispatch tests and the canonical
required gate pass.**
