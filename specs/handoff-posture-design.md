# Hand-off Permission Posture — Design

Status: Draft (brainstorming output; awaiting user review → plan-and-execute)
Author session: 2026-07-31
Anchors: PRD §UX Contract (Auto/Ask), §Tier Boundary (free "direct
intake/edits"); `specs/auto-ask-execution-mode-design.md`. Empirical basis: the
hand-off ceiling experiment (2026-07-31) showed jini's static `--print` args
make the claude-code handoff plan-only; `--permission-mode acceptEdits` applies
edits (no commands); `--dangerously-skip-permissions` edits AND verifies.

## Problem

Jini's CLI handoff passes static args (`claude --print {{prompt}}`), so no
matter the execution mode it is plan-only: the downstream agent diagnoses and
proposes but cannot apply edits. Execution mode (Auto/Ask) has zero influence
on the downstream permission posture. This is the gap between "proven autonomy"
and "hand the buildout to Jini."

## Decision (user-ratified)

Three postures, with GRADUATED CONSENT:

- **plan** — today's behavior (diagnose/propose; no edits). Default everywhere.
- **semi** — `acceptEdits`: applies file edits, never runs commands. Safer;
  cannot self-verify (running tests stays a human/CI step).
- **autonomous** — full edit + verify (runs commands too). Truly unattended.

Selection is by an explicit per-directory **trust level**, not the mode alone:

- `jini trust` blesses the cwd at **semi** (the safer default).
- `jini trust --autonomous` escalates the cwd to **autonomous** (edits +
  commands), with a stronger warning.
- Execution mode gates: **Ask → always plan**; **Auto → the cwd's trust
  level**. Untrusted cwd → plan.

Nothing escalates without both Auto mode and an explicit trust grant at that
level.

**Informed explicit consent (hard requirement):** a trust grant is only
recorded after the user is shown the concrete repercussions of that level and
actively confirms. Consent is per-level (semi vs autonomous are separate
grants), auditable (the store records what was acknowledged and when), and
revocable (`jini trust remove`). The runtime disclosure line keeps the user
aware on every non-plan hand-off; the grant-time prompt is where consent is
obtained.

## Scope: which routes posture applies to

Posture is a property of a DOWNSTREAM CLI's permission mode. It applies only to
CLI-handoff routes (codex/claude-code/…). In the FREE tier, Jini's autonomy
comes from the coding CLI it hands off to — Jini does not run an edit/execute
loop of its own:

- Provider-API routes and local-model routes return TEXT (`generateProviderText`
  → an answer). In the free tier they never edit files or run commands as part
  of the route, so there is nothing to escalate — posture is always plan
  (advisory), even in a trusted directory. Enforced structurally: only
  `cliHandoffDescriptor`s carry `SemiArgs`/`AutonomousArgs`; a route with no
  such descriptor has no verified args → plan.
- The narrow `maybeHandleLocalTextFileEditIntent` path (jini editing a specific
  file the user directly named in the command) is request-scoped explicit
  consent, not a standing unattended grant, and is OUT of this model.

Autonomous edit/execute WITHOUT an underlying CLI — Jini running its own
agentic loop over any model, including free local models — is a PAID
capability, not a free-tier gap. It lives in the commercial repo and is a
NEW capability beyond the PRD's current Autopilot (which is managed
route/throttle policy, not a coding-agent loop); it is out of scope for this
spec. This free posture feature only governs the downstream CLI's permission
mode, and trusting a directory changes nothing about the free/paid boundary.

## Posture resolution (safety truth table)

| Execution mode | cwd trust level | descriptor verified for that level? | Posture |
|---|---|---|---|
| Ask  | any        | any | plan |
| Auto | (untrusted)| any | plan |
| Auto | semi       | semi ok | semi |
| Auto | autonomous | autonomous ok | autonomous |
| Auto | autonomous | only semi verified | semi (degrade down, never up) |
| Auto | semi       | semi not verified | plan |

**Invariants:**
- No trust configured (the default) ⇒ every row is plan ⇒ byte-identical to
  today.
- Degrade only DOWNWARD when a descriptor lacks verified args for the granted
  level — never silently escalate above the granted level.
- Ask mode is always plan ("check with me").

## Components

### Posture type + resolver (`handoff_posture.go`)
```go
type handoffPosture int
const (
    posturePlan handoffPosture = iota
    postureSemi
    postureAutonomous
)
func resolveHandoffPosture(descriptor cliHandoffDescriptor) handoffPosture
```
Resolver: if `effectiveExecutionMode() != auto` → plan. Else read the cwd's
trust level; map to the highest posture the descriptor has verified args for,
capped at the granted level; default plan.

### Per-descriptor posture args (`cli_handoff.go`)
Add `SemiArgs`, `AutonomousArgs []string` to `cliHandoffDescriptor`.
Verified-only:
- `claude-code`: `SemiArgs = ["--permission-mode","acceptEdits"]`,
  `AutonomousArgs = ["--dangerously-skip-permissions"]` (both proven
  2026-07-31).
- All others: empty → stay plan-only until dogfooded.

Applied in `resolveCLIHandoffCommand`: append the resolved posture's args to
`DefaultArgs`, before the `{{prompt}}` arg, ONLY when the user has not set
`JINI_*_ARGS` (explicit user args always win and are never modified).

### Trust store (`handoff_trust.go`)
- `~/.jini/trusted-dirs.json` (global; same home as mode.json). Map of
  absolute, symlink-resolved path → record `{level ("semi"|"autonomous"),
  granted_at (RFC3339), acknowledged ("edits" | "edits+commands")}`. The
  `acknowledged` field is the audit trail of what the user consented to.
  Fail-safe: unreadable/corrupt → empty → plan-only.
- `trustedLevelForDir(dir) (level string, ok bool)`: resolve symlinks, exact
  match.

### `jini trust` command — informed explicit consent
Taught surface + settling-doc amendment, like `mode`/`savings`.

- `jini trust` (default semi) and `jini trust --autonomous` (full) both run a
  consent flow before recording anything. Tone is plain and factual — state
  what the level does and its scope; do not dramatize.
  1. Print a scoped description of the level. Copy is ROUTE-AGNOSTIC — a trust
     grant is per-directory and Auto may select different hand-off CLIs there,
     so it names "the configured hand-off CLI", not a specific tool:
     - semi: "In this directory, Auto mode lets the coding CLI it hands off to
       (such as Claude Code or Codex) apply file edits without asking first. It
       won't run commands. This applies only here and only in Auto mode; Ask
       mode and other directories are unchanged."
     - autonomous: "In this directory, Auto mode lets the coding CLI it hands
       off to (such as Claude Code or Codex) apply file edits and run commands
       without asking first. This applies only here and only in Auto mode; Ask
       mode and other directories are unchanged. Use it where you're
       comfortable with unattended edits and commands."
   - Examples in copy must be REAL hand-off CLIs (Claude Code, Codex, Gemini
     CLI, Aider, opencode) — never BYO text-API providers (e.g. xAI/Grok) or
     competitors (e.g. GitHub Copilot), which are not edit/run hand-off routes.
  2. Prompt `Trust this directory for <level> hand-off? [y/N] ` — default No;
     only an explicit `y`/`yes` records the grant. Anything else → record
     nothing.
  3. No TTY (piped stdin) → record nothing and print "Trust needs an
     interactive terminal; nothing changed.", exit non-zero — fail closed,
     never grant without a human. (Reuses the throttle-prompt TTY/first-line
     discipline.)
  4. On `y`, record the grant with `granted_at` + `acknowledged`, and print a
     one-line confirmation naming the level, the directory, and
     `jini trust remove` to revoke.
- `jini trust list` — show each trusted dir, its level, and when granted.
- `jini trust remove [dir]` — untrust (cwd default); no consent prompt needed
  to REDUCE privilege.
- Re-granting at a different level re-runs the consent flow for the new level.

### Honesty / no silent escalation
`resolveHandoffPosture` is a pure function, called at BOTH sites: (1)
`resolveCLIHandoffCommand` to append args (no writer there), and (2) the
direct-answer path (`runDirectCLIHandoffAnswer`, which has stdout) to print the
disclosure line before the hand-off. Cheap to call twice; avoids threading
posture through the provider stack. Narration is scoped to the direct-answer
path for MVP; other paths (draft flow) can adopt it later.

Emit one line before a non-plan handoff, naming the ACTUAL route via
`cliHandoffLabel(mode)` (the run-time path knows which CLI answered):
- semi: `Auto mode: applying edits in <dir> via <route label> (semi; no commands run).`
- autonomous: `Auto mode: applying edits and running commands in <dir> via <route label> (autonomous).`
When Auto + untrusted, a one-line hint after the answer:
`Plan-only here — run \`jini trust\` (edits) or \`jini trust --autonomous\` (edits+commands) to enable Auto hand-off in this directory.`
When Auto + trusted but the active route lacks verified args for the granted
level (posture degraded to plan): `Plan-only — <route label> doesn't support
<level> hand-off yet.` So a trusted-dir user is never left silently wondering
why nothing escalated.

## Error handling
- Trust store errors never block a task → treated as untrusted (plan).
- Posture resolution best-effort; any uncertainty → plan.
- `JINI_*_ARGS` set → posture logic is a complete no-op (user owns args).

## Testing
- `handoff_posture_test.go`: every truth-table row incl. fail-safe + the
  downward-degrade row; Ask-always-plan; unverified-descriptor→plan/degrade;
  `JINI_*_ARGS`-set→no-op.
- `handoff_trust_test.go`: round-trip with levels + consent metadata,
  corrupt→empty, symlink normalization, list/remove.
- `jini trust` consent flow: scoped description shown for each level; `[y/N]`
  defaults No (non-y → nothing recorded); `y` → grant recorded with
  `acknowledged`; no-TTY → record nothing + non-zero exit; `--autonomous`
  states edits+commands in a neutral tone; `remove` needs no prompt.
- Copy audit: consent + disclosure strings contain no fear framing (no caps
  emphasis, no threat lists) — a test asserts the strings are plain/factual
  (e.g. no ALL-CAPS capability words).
- `cli_handoff` arg construction: claude-code semi appends
  `--permission-mode acceptEdits`; autonomous appends
  `--dangerously-skip-permissions`; both before the prompt; plan path
  unchanged.
- Honesty: semi/autonomous runs emit the correct disclosure; untrusted Auto
  emits the hint.
- Regression: existing handoff/smoke/dogfood tests unchanged (default = plan).

## Trace / drift bookkeeping
- New settling-decisions entry (command surface + graduated-consent posture
  policy).
- Trace: advances free "direct intake/edits"; record Implemented/residual per
  scope; bump pinned counts if a P0 row is added.

## Out of scope (explicit)
- Autopilot / paid throttle switching (separate, commercial).
- Posture args for CLIs other than claude-code until each is dogfooded.
- Recursive/subdirectory trust (MVP is per-exact-directory).

## Open assumptions (flagged for review)
- claude-code flags verified empirically 2026-07-31: `acceptEdits` = edits, no
  commands; `--dangerously-skip-permissions` = edits + commands.
- Trust is per-exact-directory for MVP; recursive trust is a future refinement.
- Default trust level for bare `jini trust` is semi (safer); autonomous
  requires the explicit flag.
