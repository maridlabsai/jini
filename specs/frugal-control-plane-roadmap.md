# Frugal Control-Plane and Incremental-Delivery Roadmap

Status: **STRATEGY AND PLANNING CONTRACT — NOT YET IMPLEMENTED.**

- Nothing in this document describes shipped behavior. It records the settled
  direction for how Jini becomes a frugal, local-first control plane, delivered
  as independent customer increments.
- **Pricing and competitive targets here are non-binding planning hypotheses
  that require measurement — not offers, commitments, benchmarks, or achieved
  results. This is a public repository: nothing here is secret, and no figure
  below is a price or a promise.**
- **Current, prototype, planned, and deferred capabilities must never be
  conflated.** Every capability statement below is tagged with its state. A
  prototype is not a shipped feature; a planned increment is not a current one.
- Canonical product thesis stays in
  [number-one-platform-prd.md](./number-one-platform-prd.md); priority ranking in
  [competitive-gap-priorities.md](./competitive-gap-priorities.md); the execution
  safety contract in [r1-safe-execution-contract.md](./r1-safe-execution-contract.md).
  This file refines *delivery strategy*, not the thesis or the scores.
- Settled 2026-10-21 (STEP 68 owner decisions). Supersedes nothing; adds the
  control-plane and incremental-delivery direction on top of the merged R1 work.

## a. Positioning and boundary

**Positioning (verbatim, owner-approved):**

> Jini is the local-first control plane that keeps coding work safe, resumable,
> and portable across agents and models — from the terminal first and from any
> device later.

### Jini competes on
- safe execution and rollback;
- durable resume;
- portability across agents, models, and capacity the user already has;
- honest verification and terminal outcomes;
- provider-neutral routing;
- frugal use of local, free-tier, BYOK, and existing-subscription capacity;
- one reusable control protocol serving the CLI first and other surfaces later.

### The initial product does NOT compete by
- replacing the full IDE;
- owning a cloud development computer;
- subsidizing inference;
- bundling proprietary frontier models;
- building native mobile before validating a PWA;
- providing unrestricted autonomous shell/network execution;
- providing hosted multi-user collaboration or a Jini-hosted relay in the first
  release.

### Scope (from STEP 67, preserved)

**Included ~80% scope** (planning hypothesis, not a measured claim): repository
understanding; scoped single-file and multi-file changes; bug fixes and focused
refactors; test creation and repair; documentation and review work;
build/lint/test verification; safe interruption and resume; model / provider /
installed-agent selection; throttle-aware fallback; remote status,
clarification, approval, rejection, pause, resume, and cancel.

**Excluded initial ~20%:** Jini-operated cloud development computers; full IDE
replacement; native iOS/Android apps; large enterprise agent fleets; massive
autonomous migrations; proprietary frontier-model development; bundled high-cost
inference; complex multi-agent missions; enterprise SSO/SCIM and fleet
analytics; deployment automation beyond currently proven scope.

**Remote status/control is split deliberately:**
- the portable operation/event/approval/outcome **protocol and records are
  designed now** (as a reusable seam — see §e);
- **networked remote access is deferred** until the R1 safety and R2 durability
  guarantees exist;
- a **hosted relay remains outside the initial release** and is a separately
  costed later decision (§d.11, §h).

## b. Execution-route abstraction

The durable selection unit is an **execution route, not a model**. A route must
be able to describe:

1. harness/agent;
2. provider;
3. model;
4. execution location;
5. capability set;
6. task/mutation class;
7. safety/isolation guarantees;
8. trust and approval requirements;
9. privacy/data-handling policy;
10. network policy;
11. cost class;
12. quota/rate-limit state;
13. health/throttle state;
14. verification capability;
15. proven task classes from Jini's own evidence.

**Field 15 remains empty / unasserted until R4 (verification depth) and R7
(route-outcome learning) produce real evidence.** Jini must not claim a route is
"proven" for a task class before its own measurement exists.

### Planned routing policies (with honest state)

| Policy | Intended behavior | Current state |
| --- | --- | --- |
| `manual` | user pins a route | **implemented** (`jini route set`) |
| `free-first` | prefer local, existing-subscription, and free-tier capacity; never silently escalate to paid | **planned** — cheap-first heuristics exist, but not a named/complete free-first policy |
| `fast` | lowest-latency acceptable route | **partial** — depth/effort classes influence selection; not a settled named mode |
| `balanced` | cost/latency/quality balance | **planned** |
| `deep` | highest-rigor route for hard tasks | **partial** — `deep` depth class exists as a heuristic input |
| `private` | restrict to local/privacy-preserving routes | **planned** — not implemented |

Existing manual/cheap/depth behavior **must not be relabeled** as completed
free-first or private routing. Those named policies are planned.

## c. Free / open / local capacity policy

Every route is tagged with **one or more** of this **five-way taxonomy** — the
categories may overlap where reality requires (for example, a large open-weight
model served through a paid metered API is both open-weight and paid-metered):

- **open-weight** — weights are publicly available;
- **free-tier API** — a hosted API offering a no-cost quota;
- **locally runnable at practical consumer cost** — runs on typical developer
  hardware without special expense;
- **zero-marginal-cost through existing paid capacity/subscription** — the user
  already pays for it (e.g. an installed Claude Code / Codex / Gemini
  subscription reached via handoff);
- **paid metered API** — billed per use.

Explicit rules:
- **open-weight does not mean free to run.**
- A very large open-weight model (e.g. a 501B-parameter model) **must not be
  presented as a practical free-local route merely because its weights are
  available** — serving it has real cost and hardware requirements.
- Provider limits and free tiers are **volatile**; they must be discovered and
  reported honestly, never hard-coded as guarantees.
- **Jini must never guarantee third-party free capacity.**
- `free-first` must prefer local, existing-subscription, and free-tier capacity.
- **Jini must never silently escalate to a paid metered route.**
- Paid escalation requires **explicit informed approval**.
- **Capability compatibility must be checked before routing** (a route unable to
  perform the operation class is not selected).
- Throttling, fallback, and the **final route actually used must be reported
  honestly**.

Groq, OpenRouter, Beam, and any other named model or provider appear here only
as **dated planning examples** — not permanent promises, defaults, endorsements,
or benchmark claims. (Example, as observed 2026-10-20: Groq and OpenRouter
publish free-tier APIs with per-organization/payment-gated limits; Beam 1.0 is
an Apache-2.0 open-weight model that is open-weight but not practically
free-to-run locally. These are illustrations of the taxonomy, not commitments.)

## d. Frugal multi-surface architecture

One engine, many surfaces, in this order:

1. **CLI remains authoritative** — the first and primary customer surface.
2. **Extract only the minimum reusable execution core** needed by the customer
   increments below; avoid a broad horizontal refactor of `internal/app`.
3. **R1A emits portable structured records** (§e).
4. **R2 adds the durable journal and resume** on top of those records.
5. **R3 adds route policy, health, quota, and failover.**
6. **R9A adds local status/control** against the shared protocol.
7. **R9B adds a responsive local PWA.**
8. **Remote PWA access may later use user-configured private access** (e.g. the
   user's own private networking or an explicitly configured secure tunnel).
9. **R9C adds a thin desktop shell** reusing the same protocol/PWA.
10. **R9D native mobile is conditional** on evidence that the PWA is insufficient.
11. **An optional hosted encrypted coordination relay** is a later,
    separately-costed decision — not part of the initial release.
12. **Repositories, source content, and inference must not transit a future
    relay merely for convenience.**
13. **Multi-user / team administration is deferred.**

**Prototype honesty:** `jini app serve --stdio --surface macos`
(`internal/app/app_sidecar.go`) and the macOS/Tauri shell under `apps/macos` are
**prototypes** — a line-delimited JSON request/response + event seam driven over
stdio. They are **not** a shipped network server, remote-control surface, web
app, or mobile app. They inform §e; they must never be described as production.

## e. Portable R1A seam

R1A must emit a **minimum stable schema** shared by the terminal and any future
surface. Fields:

- schema version;
- operation ID;
- run ID;
- session ID (where applicable);
- monotonically ordered event sequence;
- timestamp;
- mode;
- operation class;
- requested capability;
- declared target / write set;
- structured policy decision and reason;
- approval request/reference and disposition;
- snapshot/transaction evidence reference;
- verification-evidence reference;
- terminal outcome;
- redacted error/recovery information (where applicable).

Requirements:
- **appendable events** (append-only, never rewritten in place);
- **stable identifiers**;
- **idempotency** where a mutation can be retried;
- **no dependency on terminal rendering** (terminal output is a projection);
- **redaction of secrets** from every record;
- **no remote transport requirement**;
- **no daemon/server requirement**;
- **compatibility with the existing sidecar seam** (`app_sidecar.go` already
  carries operation methods, idempotency keys, ordered events, session IDs, and
  an `approval.resolve` method) **without declaring that prototype
  production-ready.**

This is a reusable seam only. **It authorizes no daemon, network API, browser
UI, remote access, desktop work, mobile work, or routing implementation in
R1A.**

**Minimal R1A depth:** R1A may implement these fields only to the depth
Increment 1 requires, provided the schema remains **versioned** and
**extensible**, omitted future capabilities are **never falsely emitted as
evidence**, and the change requires **no remote transport, durable R2
journaling, routing orchestration, daemon, or UI work**.

## f. Incremental customer-delivery ladder

Twelve independently deliverable increments. Release states are used
consistently: **Experimental**, **Beta**, **Stable**. Planned increments are
never described as shipped.

The twelve numbered increments form a **capability / dependency sequence, not a
strict implementation chronology**. Increment 5 documents an existing
manual-route foundation for completeness and is **not** claimed as a future
milestone. Customer-usable delivery should still proceed in the smallest safe
increments.

### 1. Safe supervised single-file edit
- **Customer job:** change one file safely and know it was done or cleanly undone.
- **Observable behavior:** preview → exact approval → snapshot → atomic
  replacement → minimum verification → rollback or evidence.
- **Operation classes:** mediated file read; mediated single-file write.
- **Safety mode:** SUPERVISED.
- **Non-goals:** multi-file transactions; raw shell; network; Auto.
- **Failure behavior:** on verification failure or interruption, roll back when
  the target still matches the produced after-hash; on concurrent user edit,
  conflict-stop (never overwrite).
- **Tests/gates:** R1 acceptance subset — single-file atomic write, concurrent-
  edit conflict, snapshot rollback, verification-required outcome.
- **Rollback:** per-run write-set snapshot; never `git reset --hard`/`clean`/stash.
- **Release channel:** Experimental → Beta.
- **Promotion criterion:** single-file R1 acceptance tests green on
  macOS/Linux/Windows.
- **Independently revertible:** yes.

### 2. Safe supervised multi-file transaction
- **Customer job:** make a coherent change across several files, all-or-nothing.
- **Observable behavior:** declared write set → all-or-nothing mutation →
  conflict-safe rollback.
- **Operation classes:** mediated multi-file write.
- **Safety mode:** SUPERVISED.
- **Non-goals:** shell; network; Auto; cross-worktree/submodule writes.
- **Failure behavior:** partial-failure → roll back the whole set; conflict →
  stop without overwrite.
- **Tests/gates:** multi-file atomic commit/rollback, partial-failure rollback.
- **Rollback:** multi-file write-set snapshot.
- **Release channel:** Beta.
- **Promotion criterion:** all-or-nothing + partial-failure tests green.
- **Independently revertible:** yes.

### 3. Honest terminal outcomes wired end to end
- **Customer job:** trust what "done" means.
- **Observable behavior:** `completed_verified`, `completed_unverified`,
  `failed_rolled_back`, `failed_partial`, `recovery_required` (and the other R1
  contract outcomes) surfaced clearly in the terminal.
- **Operation classes:** n/a (projection of the §e records).
- **Safety mode:** n/a.
- **Non-goals:** deeper verifier intelligence (R4).
- **Failure behavior:** never report success without verification evidence.
- **Tests/gates:** terminal-outcome vocabulary round-trip from the §e records.
- **Rollback:** n/a.
- **Release channel:** Beta.
- **Promotion criterion:** every mutating path maps to exactly one terminal
  outcome, tested.
- **Independently revertible:** yes.

### 4. Same-machine resume after interruption
- **Customer job:** pick up an interrupted task on the same machine.
- **Observable behavior:** `jini resume`/`continue` restores task state and
  **re-evaluates current trust and policy** before continuing.
- **Operation classes:** read; mediated write on resume.
- **Safety mode:** SUPERVISED.
- **Non-goals:** crash-resumable durable journal (R2); cross-machine resume.
- **Failure behavior:** if trust/policy no longer permits, stop and report.
- **Tests/gates:** resume re-evaluates trust (R1 §3); no stale-privilege resume.
- **Rollback:** interrupted transaction rolls back or reports `recovery_required`.
- **Release channel:** Beta.
- **Promotion criterion:** resume never elevates privilege; interrupted state is
  honestly reported.
- **Independently revertible:** yes.

### 5. Manual execution-route selection
- **Customer job:** choose exactly which route runs the work.
- **Observable behavior:** `jini route set <route>`; honest route receipt.
- **Operation classes:** routing (selection only).
- **Safety mode:** inherits the task's mode.
- **Non-goals:** automatic selection; failover.
- **Failure behavior:** incompatible route rejected before use.
- **Tests/gates:** capability check before selection; receipt accuracy.
- **Rollback:** n/a (selection).
- **Release channel:** Stable (manual pin largely exists today).
- **Promotion criterion:** route receipts match the route actually used.
- **Independently revertible:** yes.

### 6. Free-first read-only routing
- **Customer job:** run read-only work on free/owned capacity first.
- **Observable behavior:** local/free/existing-capacity routes tried first; no
  silent paid escalation; explicit approval before any paid route.
- **Operation classes:** read-only routing.
- **Safety mode:** read-only.
- **Non-goals:** mutating failover (increment 8).
- **Failure behavior:** on exhaustion, honest report + explicit paid-approval
  prompt; never silent escalation.
- **Tests/gates:** no-silent-paid-escalation; 429/Retry-After awareness;
  capability validation; honest fallback report.
- **Rollback:** n/a (read-only).
- **Release channel:** Beta.
- **Promotion criterion:** zero silent paid escalations in tests.
- **Independently revertible:** yes.

### 7. Route health / throttle classification and read-only failover
- **Customer job:** keep read-only work moving when a route is throttled/down.
- **Observable behavior:** 429/Retry-After classified; read-only work fails over
  to another eligible route; final route reported.
- **Operation classes:** read-only routing + health.
- **Safety mode:** read-only.
- **Non-goals:** mutating failover (increment 8).
- **Failure behavior:** when no eligible route remains, hold/park + honest report.
- **Tests/gates:** throttle classification, read-only failover, honest receipts.
- **Rollback:** n/a.
- **Release channel:** Beta.
- **Promotion criterion:** throttle handled without a false success.
- **Independently revertible:** yes.

### 8. Mutating-task failover (gated)
- **Customer job:** continue a *mutating* task across a route switch safely.
- **Observable behavior:** switch mid-task only with durable state + safe
  rollback guarantees intact.
- **Operation classes:** mediated write + routing.
- **Safety mode:** SUPERVISED → AUTO (only where a tested backend exists).
- **Non-goals:** any switch that risks duplicate non-idempotent mutation.
- **Failure behavior:** if guarantees are absent, refuse to fail over; park.
- **Tests/gates:** requires **R1 and R2 guarantees** proven first.
- **Rollback:** durable-journal-backed (R2).
- **Release channel:** Experimental.
- **Promotion criterion:** no duplicate/partial mutation across a switch.
- **Independently revertible:** yes.

### 9. Loopback-only, read-only browser status
- **Customer job:** glance at run status in a browser on the same machine.
- **Observable behavior:** a loopback-only page shows status/diff/evidence; no
  mutation.
- **Operation classes:** read-only projection of §e records.
- **Safety mode:** read-only.
- **Non-goals:** approvals; mutation; remote access.
- **Failure behavior:** page unreachable → CLI unaffected.
- **Tests/gates:** loopback bind only; no mutating endpoint.
- **Rollback:** n/a.
- **Release channel:** Experimental.
- **Promotion criterion:** binds loopback only; strictly read-only.
- **Independently revertible:** yes.

### 10. Authenticated local approval/control PWA
- **Customer job:** approve/reject/pause/resume/cancel from a local browser.
- **Observable behavior:** authenticated local PWA issues control operations
  with a **single-writer lease** preventing conflicting approvals.
- **Operation classes:** control operations against the protocol.
- **Safety mode:** SUPERVISED.
- **Non-goals:** remote access; multi-user collaboration.
- **Failure behavior:** lost lease → read-only; no conflicting action applied.
- **Tests/gates:** auth required; single-writer-lease correctness; idempotent
  control ops.
- **Rollback:** control ops are idempotent/append-only.
- **Release channel:** Experimental.
- **Promotion criterion:** no conflicting concurrent approval possible.
- **Independently revertible:** yes.

### 11. Remote PWA through explicitly configured private access
- **Customer job:** use the same PWA from another device the user owns.
- **Observable behavior:** reachable only over user-managed private networking or
  an explicitly configured secure tunnel.
- **Operation classes:** same as increment 10, over private access.
- **Safety mode:** SUPERVISED.
- **Non-goals:** Jini-hosted relay; public exposure by default.
- **Failure behavior:** without configured private access, remote use is simply
  unavailable (no default public surface).
- **Tests/gates:** no default public bind; explicit-config requirement.
- **Rollback:** inherits increment 10.
- **Release channel:** Experimental.
- **Promotion criterion:** never reachable without explicit user configuration.
- **Independently revertible:** yes.

### 12. Thin desktop shell reusing the control protocol/PWA
- **Customer job:** a lightweight desktop window over the same engine.
- **Observable behavior:** a thin shell (extending `apps/macos`) reuses the same
  protocol/PWA — **not a second execution core**.
- **Operation classes:** same control operations.
- **Safety mode:** SUPERVISED.
- **Non-goals:** a separate product implementation; a divergent engine.
- **Failure behavior:** shell failure never affects the CLI engine.
- **Tests/gates:** shell issues only protocol operations; no duplicated engine
  logic.
- **Rollback:** inherits the protocol.
- **Release channel:** Experimental.
- **Promotion criterion:** zero engine logic duplicated in the shell.
- **Independently revertible:** yes.

## g. First customer-visible R1A delivery

The next implementation target is **Increment 1**:

> Safe supervised single-file edit with preview, exact approval, transaction
> snapshot, atomic write, minimum verification, rollback-or-evidence, and a
> portable outcome record.

It must be a **complete, customer-usable vertical slice**. It must **not** be
delayed by first implementing the whole mediator, a daemon, a web UI, the
routing system, or a broad package extraction. Deliver the narrow slice end to
end; extend horizontally only as later increments require.

## h. Pricing and unit-economics direction — non-binding planning hypotheses

**Non-binding planning direction — not an offer, a price, or a commitment.
(Public repository: this is direction, not a secret.)**

- **Jini Core:** free.
- **Jini Router:** a paid low-monthly subscription; the exact amount remains
  **unset pending the economics evaluation below**.
- **Router Beta:** free while evidence is collected.
- **No bundled or marked-up inference** initially.
- Execution is **BYOK / local / free-tier / existing-subscription** capacity.
- **No Jini-funded cloud computer.**
- A **relay, if ever offered, is measured separately**.
- **Team / enterprise pricing remain unset.**

### Viability assumptions (what must hold for a low monthly subscription)
- execution remains on user-owned capacity by default;
- no Jini-funded inference;
- static / low-cost client surfaces (PWA assets);
- any relay handles **encrypted coordination only, not repositories**;
- bounded storage and notification expense;
- bounded support burden;
- provider / free-tier volatility is **not absorbed as a Jini service
  guarantee**.

**Competitive data point (dated observation, may change):** Factory's entry
paid individual tier was listed at **$20/month** (official `docs.factory.com`,
accessed 2026-10-20). The **strategic hypothesis** is that Jini can economically
serve the common ~80% of coding work at roughly **one-third** of such a
competitor entry-price benchmark. This is a **non-binding hypothesis that
requires measurement** (§i) — not an achieved result, a benchmark, or a
customer-facing claim; "one-third" must never ship as a claim before the §i
evaluation supports it.

### Measurements required before offering any price

No paid price may be offered until a dated evaluation records evidence for at
least:

- active users;
- route mix;
- support minutes per active user;
- relay cost;
- storage cost;
- notification cost;
- payment-processing cost;
- provider-funded versus user-funded inference mix;
- gross-margin sensitivity;
- conversion;
- churn.

The eventual subscription amount stays unset until these measurements support
it. No target values are asserted here.

## i. Measurement before claims

No external **"80%", "one-third", "3× cheaper"**, savings, performance-
superiority, security, or readiness claim may ship until a **dated,
representative evaluation** records all of:

- eligible task set;
- verified completion rate;
- fallback success rate;
- task latency;
- human interventions per task;
- resumed-task success rate;
- free/existing-capacity utilization percentage;
- provider cost;
- Jini infrastructure cost;
- comparison tier and date;
- sample size and exclusions.

These claims are bound by the existing claim-language discipline
([engineering-gate-matrix.md](./engineering-gate-matrix.md), the Commit-gate
claim-language gate). **The current claim-language gate does not itself measure
these outcomes** — it blocks unqualified public wording; the evaluation above is
a separate prerequisite before any such claim is made.

**Gate-coverage disclosure:** `tools/claim_language_gate.sh` deliberately
excludes `specs/**`, so the live claim-language gate does **not** enforce the
claims written in this roadmap. Honesty and evidence in these specs therefore
remain a **review responsibility**, not an automated guarantee. (This step does
not expand or modify that gate.)

## j. Risks and explicit deferrals

- **Free-tier / provider volatility:** limits change without notice; classify
  and report honestly, never guarantee third-party free capacity.
- **Low-subscription economics breaking through funded compute:** any
  Jini-funded inference or cloud computer invalidates the hypothesis — hard
  invariant against both.
- **R1A scope creep into remote/UI work:** R1A emits the §e seam only;
  daemon/web/remote stay out until R1+R2 hold.
- **`internal/app` coupling blocking reuse:** extract only the minimum reusable
  core; a broad refactor is itself a risk.
- **Unsupported competitive claims:** route every external figure through the
  claim-language discipline and the §i evaluation.
- **Remote authentication and single-writer correctness:** deferred to the PWA
  increments; must be proven before remote control is offered.
- **Cross-provider semantic differences:** routes differ in tool-calling,
  structured output, and context limits; capability checks precede routing.
- **Support burden from many providers:** must stay bounded (viability
  assumption); provider breadth is not a support guarantee.

### Deferrals (not in the initial release)
- networked remote control until R1 + R2 guarantees exist;
- any Jini-hosted relay (separately costed, later);
- native mobile until PWA evidence proves a material unmet need (R9D);
- multi-user / team administration.
