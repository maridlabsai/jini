# Pre-Viral Readiness Backlog

The critical review (2026-09-18) found Jini has a viral-worthy **wedge** and strong
**engineering**, but is not yet a complete or quality-*proven* product. It can seed
a focused beta on the cost story; it cannot yet survive broad virality as a Claude
Code peer. This backlog closes that gap. Sequenced by **impact ÷ effort** — do the
cheap high-leverage move first, then the load-bearing one.

## The four gaps, prioritized

| # | Gap | Impact | Effort | Priority | Why this rank |
| --- | --- | --- | --- | --- | --- |
| 4 | **Positioning honesty** — "un-metered cost-first router", not "free Claude Code" | High | Trivial | **P0 — now** | A doc/GTM change that defuses the feature-parity critique users will make. Near-zero cost, immediate. |
| 1 | **Coding quality** — confidence-based routing + verification loop | Critical | Med-High | **P0 — load-bearing** | Converts a viral spike into retention. "Cheap *and* verified" answers the #1 developer pain (hallucination/accuracy). Without it, virality churns. |
| 2 | **Viral loops** — shareable receipts + skill/agent gallery | High | Medium | **P1** | The flywheel. The wedge earns attention; loops make it self-propagate. The ledger exists; the *shareable* surface doesn't. |
| 3 | **Feature depth** — MCP client + hook-enforced cheap delegation | Medium | Med-High | **P2** | Closes the "obviously thinner than Claude Code" gap. Hooks-delegation also cuts our own token cost ~90% (dogfood ROI). |

Rationale for the order: **#4 is a free ratchet** (do immediately). **#1 is load-bearing**
for everything downstream — no point amplifying a product that churns on accuracy, so
it precedes the viral loops. **#2** makes the wedge self-propagate once #1 secures
retention. **#3** is parity/depth, valuable but not gating a focused beta.

---

## #1 (deep design) — Confidence-Based Routing + Verification Loop

**Problem.** Jini routes *cheapest-capable* first. On hard tasks a cheap model may be
*less accurate*, and today there is **no confidence signal and no self-correction** —
so Jini can ship wrong code cheaply. 2026 developer sentiment is explicit: *cost
efficiency + hallucination control now outrank raw capability*, and *hallucination ∝
cost* (wasted runs = spend). The fix turns Jini's weakness into its second wedge:
**"cheap AND verified."**

Two mechanisms, built in this order (objective first, cheap-to-earn-trust):

### Mechanism A — Verification loop (build first; objective; highest ROI)
After a coding task, run **objective** checks before declaring success:
- **Buildability:** does the changed code compile / typecheck? (language-aware:
  `go build`, `tsc --noEmit`, `python -m py_compile`, etc.)
- **Tests:** do existing tests still pass? Did requested new tests pass?
- **Constraint satisfaction:** reuse the `jini check model` objective-probe idiom
  (the 4-probe framework already exists) — did the output meet the request's
  explicit, checkable constraints (file exists, function present, N items)?

Flow: task → produce → **verify** → on pass, ship with an honest receipt
(`verified: build ✓ tests ✓`); on fail, **one same-model retry** with the failure fed
back, then **escalate** (Mechanism B), then surface to the user with the failure. Never
present unverified output as done.

*Why first:* objective (no model self-rating), reuses the check-model framework, and it
alone is a real accuracy story. It also produces the **measurement** that answers
"product-quality standards unproven": *% tasks verified-correct without escalation*.

### Mechanism B — Confidence-based routing (escalation)
Add a **quality axis** to the existing cost-first route engine (which today ranks by
cost + the throttle fallback ladder). Escalate cheap → mid → premium when:
1. **Verification fails** (objective trigger — the primary signal), OR
2. **Difficulty pre-estimate is high** — cheap heuristics only (task type, multi-file
   scope, code-artifact complexity, ambiguity), so a task *likely* hard starts a tier
   up instead of failing down, OR
3. The cheap model **declines / emits incomplete or structurally-invalid** output.

**Deliberately NOT used as the confidence signal:** the model's own self-rating —
unreliable, and reasoning-block replay is itself a documented hallucination source.
Confidence is **earned by verification**, not asserted by the model.

Guardrails (frugality stays P0):
- **Escalation cap** per task (don't burn premium on everything); default start = cheap.
- The existing **soft-constraint savings knob** becomes a quality↔cost preference: a
  user can bias toward "always verify + escalate" or "cheapest, best-effort."
- **Honest receipt + ledger:** every escalation is logged and shown
  (`escalated to <model> because verification failed`) — never a silent paid adoption
  (preserves the free/commercial boundary and the BYO-no-silent-cost rule).

### Integration points
- **Route engine** — add a quality tier + escalation decision alongside cost ranking.
- **Native loop** — insert the verify step after task completion; escalation re-enters
  the loop with the stronger route.
- **Savings ledger + receipt** — log escalations honestly; report verification status.
- **`jini check model`** — its probe framework is the seed of the verifier.

### Phased build
1. **Verify-only** (Mechanism A): post-task build/test/constraint checks + honest
   receipt. Ship the accuracy *measurement*. — *first PR*
2. **Escalate-on-failure** (Mechanism B trigger 1): route up when verification fails,
   capped, ledger-logged. — *second PR*
3. **Difficulty pre-estimate** (trigger 2): start at the right tier. — *third PR*

### Risks & mitigations
- *Verification needs to run code* → time/sandbox budget; skip gracefully for
  non-verifiable outputs (prose) and mark the receipt `unverified: not runnable`.
- *Escalation cost* → caps + failure-only + honest logging.
- *Difficulty mis-estimate* → start conservative; verification corrects it downstream.

### Success metric (the product-quality number we currently lack)
`verified_without_escalation%` (efficiency) + `verified_after_escalation%` (accuracy
ceiling). Publishing these turns "quality is unproven" into a tracked, improvable stat.

---

## #2 (design) — Viral loops
- **Shareable savings receipt:** extend the ledger to emit a clean, copyable artifact
  (text + markdown + an image card) — "saved $X, dodged N throttles, verified ✓" — one
  command, paste to X / a PR / Slack. Each is an ad carrying the wedge + (post-#1) the
  *verified* proof. Modeled on Loom (the artifact is the ad).
- **Skill/agent gallery:** a published index (reuse the catalog-merge machinery) where
  users share coding skills/agents; installing a popular one requires Jini. Modeled on
  Notion templates. *Effort: medium; depends on the existing skills/agents surface.*

## #3 (design) — Feature depth
- **MCP client:** let Jini consume MCP servers (tools/resources) like Claude Code —
  the single biggest ecosystem-parity gap. *Effort: med-high.*
- **Hook-enforced cheap delegation** (the Spotify 92% pattern, [[staying-current-is-vital]]):
  PreToolUse hooks intercept large reads / boilerplate writes → cheap/free model,
  keeping the premium context lean. Do it for our **own dogfood first** (immediate
  ~90% token cut + dogfoods the feature). Enforce via hooks, not advisory rules.

## #4 (do now) — Positioning honesty
- Reframe all messaging (landing, GTM, README) to **"the un-metered, cost-first AI
  coding router — cheap *and* (post-#1) verified"**, explicitly *not* a feature-complete
  Claude Code replacement. Over-claiming parity invites the exact critique above.
- Lead with the wedge + the savings/verification receipt; concede the GUI/IDE lane and
  frontier-max-capability lane by design. Honesty is the moat, externally too.

---

**Sequence to execute:** #4 (now, doc/GTM) → #1 Phase 1 (verify-only, the load-bearing
build) → #1 Phases 2–3 → #2 (loops) → #3 (depth). Each lands through the protected PR +
quality-battery pipeline. Related: [[gtm-and-open-threads]], [[staying-current-is-vital]],
[[beta-launch-checklist]], [[ai-cost-routing-lessons]].
