# Native Agentic Loop — Design

Status: Draft (brainstorming output; awaiting user review → plan-and-execute).
Anchors: PRD Product Thesis ("Jini is the coding agent CLI ... local models
built in"), §Tier Boundary. Prior decisions: [[tiering-and-pricing]],
[[paid-autopilot-boundary]], [[handoff-posture-complete]],
`specs/decision-tree-backtrack-design.md`.

## Purpose

Give Jini its OWN edit→reason→act loop over any model (including free local
models), so it is a coding agent itself, not only a router that hands off to
Claude Code/Codex. This closes the gap between the PRD thesis and the
implementation, and it is what makes "hand the buildout to Jini" literally true
without a premium CLI installed.

## Locked decisions (user-ratified)

- **Bounded multi-step, edits-first.** A real read→reason→edit→(verify) loop,
  step-capped. Free ceiling is edit-only (semi); command-running/self-verify
  unlocks under an autonomous trust grant.
- **Text-protocol tool driving** (ReAct-style ACTION blocks the model emits and
  Jini parses) — works on ANY model incl. weak local ones. Designed so
  provider-native function-calling can be added later for capable models.

## Reuse (why this de-risks a large build)

- Model calls → existing `generateProviderText` (every provider + local
  OpenAI-compatible endpoint already works). No new provider infra.
- Safety → the posture/trust model already shipped (`handoff_posture.go`,
  `handoff_trust.go`).
- Cost → the savings ledger (`savings_*`); each loop run is a task.

## Components

### Tool layer (`agent_tools.go`)
Fixed toolset, each with a MINIMUM posture and cwd-confined paths:

| Tool | Min posture | Notes |
|---|---|---|
| `read_file`, `list_dir`, `search` | plan | always available |
| `edit_file` | semi | applies a file edit |
| `run_command` | autonomous | runs a shell command |
| `finish(summary)` | plan | ends the loop |

Path confinement: every path is resolved and must stay within the cwd; reject
`..` escapes and out-of-dir absolute paths. The tool's min posture is checked
against the resolved hand-off posture (`resolveHandoffPosture`) — so the SAME
`jini trust` consent governs the loop. In a plan directory the loop is
read-only analysis (reads + proposes, no edits); semi unlocks edits;
autonomous unlocks commands.

### Text protocol (`agent_protocol.go`)
- System prompt: tool list + a strict response format — one `THOUGHT` then one
  `ACTION` block (fenced, `action:` + JSON `args:`).
- Parser extracts the action. Malformed → bounded **repair** re-prompt with the
  exact format. Exhausting repairs → honest degradation to a plain advisory
  answer (never silent failure).
- Interface designed so a native function-calling driver can replace the
  text driver for capable models later (hybrid over time).

### Loop engine (`agent_loop.go`)
`runAgentLoop(ctx, request, opts) (result, report, error)`:
build prompt → call model (`generateProviderText`) → parse action → gate by
posture → execute tool → append observation → repeat. Stop on `finish`, step
cap (default ~12), repair cap, or budget (tokens/time). Deterministic bounds by
construction. Emits a report (steps, tools used, edits, cost) for the ledger +
receipts.

### Pro-backtrack seams (baked in day one — the key requirement)
- `type DecisionRecorder interface { RecordStep(step) }` — public seam (like
  the autopilot seam). Free/public: a compact linear recorder. Commercial:
  full decision-tree recorder with enumerated alternatives.
- `type Checkpointer interface { Checkpoint(node) }` — optional hook before
  edit/command steps (git snapshot). No-op in public; commercial implements.
- The loop calls these at each step so the Pro capability is a layer, not a
  later rewrite (per `specs/decision-tree-backtrack-design.md`).

### Routing / entry
The native loop is the actionable path for NON-handoff routes (local/BYO model,
no coding CLI configured) on WORK tasks — replacing today's "return text" with
"read→reason→act". Rules:
- Simple/standalone questions stay single-shot (unchanged).
- If a CLI-handoff route is active/available, it still wins (delegated) unless
  the user opts into native.
- Posture gates actions: plan → read-only advisory; semi → edits; autonomous →
  commands.

### Tiering
- **Free / public:** loop engine + tools + text protocol + posture gating +
  recorder/checkpoint SEAMS. Edit-only ceiling (semi); autonomous under trust.
- **Plus / Pro (commercial):** managed unattended runs, self-verify, throttle-
  survival integration, larger budgets; Pro adds decision-tree record/view/
  backtrack via the seams.

## Error handling
- Model unreachable / throttled → existing throttle survival + provider errors
  apply (the loop's model calls go through the same path).
- Malformed action → repair, then advisory fallback.
- Tool error (bad path, edit conflict, command failure) → fed back as an
  observation so the model can recover; hard errors stop the loop with a
  summary.
- Posture-denied tool → the tool is simply absent from the offered set for that
  posture (the model can't call what it isn't given), plus a guard that refuses
  if it tries.

## Testing
- Tool layer: path confinement (reject `..`/outside), each tool's min-posture
  gate, edit/read round-trips (temp dir).
- Protocol: parse well-formed actions; repair on malformed; advisory fallback
  after repair cap.
- Loop engine (model stubbed): bounded steps; `finish` stops; step/repair/budget
  caps; observation threading; posture gating end-to-end (plan→read-only,
  semi→edits, autonomous→commands) using a fake model that emits scripted
  actions.
- Seams: recorder receives each step; checkpointer called before write steps;
  no-op recorder/checkpointer leave behavior unchanged.
- Routing: work task on a non-handoff local route enters the loop; simple
  question does not; handoff route still delegates.
- Safety regression: no trust configured → loop is read-only advisory (no edits,
  no commands) — byte-safe default.

## Top risks (honest)
1. **Quality tracks the model** — the loop works on weak local models but
   results may be poor. Framing: "works on any model; quality tracks the
   model." No overpromising.
2. **Protocol brittleness** on small models → strict format + repair +
   advisory fallback.
3. **Large build** → phased (below).

## Phasing (plan-and-execute)
1. Tool layer + text protocol + loop engine, **read-only** (plan posture): the
   agent can read/search/reason and PROPOSE, fully tested with a stubbed model.
2. `edit_file` under **semi** posture (+ path confinement hardening).
3. `run_command` under **autonomous** posture + self-verify pattern.
4. Recorder + checkpointer **seams** (public no-op/linear).
5. Routing entry (non-handoff actionable path) + savings/receipt integration.
6. Trace/settling bookkeeping.

## Out of scope
- Native function-calling driver (text protocol first; hybrid later).
- Managed unattended runs, decision-tree view/backtrack (commercial).
- Multi-agent / parallelism.

## Open assumptions (flagged for review)
- Two decisions with the widest blast radius, taken as designed unless the user
  redirects: (a) a tool's minimum posture reuses the `jini trust` consent gate
  (loop-safety == posture model); (b) the native loop takes over the
  non-handoff actionable path (replacing "return text" for work tasks).
- Step cap ~12 and repair cap are tunable defaults.
