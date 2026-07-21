# Auto/Ask Execution Mode — Design (2026-07-21)

Implements PRD §Execution modes (`Auto`/`Ask`) and the "Ask-mode approval
before throttled-work resume" backlog item from
specs/prd-implementation-trace.md. Design ratified after a three-round
adversarial review (round 1 REJECT — false save-promise, fail-open safety
toggle; round 2 6.5; round 3 7.4 converged); the review's must-fix list is
fully folded in below.

## Context

PRD "Execution modes" (specs/number-one-platform-prd.md ~145): `Auto` default
(autonomous incl. throttle hold/self-resume, escalation rule never waived);
`Ask` asks first — before side effects and before resuming throttled work; maps
to `supervised` runtime mode (specs/execution-routing-policy.md ~181).
Switchable mid-session with one obvious action, never loses state, no second
command tree. Free-tier throttle survival core shipped in 5e5d8d7
(internal/app/throttle_survival.go). Maturity deadline ~2026-07-25.

## User-ratified decisions

1. Surface: top-level `jini mode [auto|ask]`. Bare `jini mode` prints effective
   mode + switch hint.
2. Slice scope: Ask enforces ONLY throttle-resume approval. File-edit/handoff
   gates trail with permissions work.
3. No-TTY / unpromptable contexts: fail closed with a real park record (below).
   Never silently self-resume against the chosen mode.
4. **AMENDED**: approval is remembered per throttle-surviving call (one
   `runWithThrottleSurvival` invocation), not per abstract "task" — that layer
   wraps a single generation call and has no task boundary. With prompting
   scoped to the one-shot path (decision 6) one invocation ≈ one task, so the
   user-visible behavior matches the original intent: first throttle asks,
   subsequent holds in the same call don't.
5. **NEW**: mode is GLOBAL — `~/.jini/mode.json` (literal home dir, os.UserHomeDir),
   deliberately diverging from the per-project sessionStateRoot() pattern:
   supervision is a property of the user, not the workspace. A per-project
   safety toggle silently leaving other projects autonomous is a trust hazard.
   `JINI_MODE` env overrides for scripts/CI. Divergence documented in the file
   header and the settling doc.
6. **NEW**: this slice prompts ONLY on the one-shot CLI path, where jini owns
   stdin. Interactive launcher (bufio.Scanner owns stdin, app.go:440) and app
   sidecar (stdin is a protocol pipe, app_sidecar.go:57) fail closed in Ask
   mode with park + guidance. Their interactive approval arrives with the paid
   remote-approval transport work, through the same approval interface.
7. **NEW**: prompt is `[y/N]` — default DECLINE on bare Enter, buffered
   newlines drained before reading. A stray newline must never auto-approve.
   Declining is safe: work is parked and resumable.

## Design

### internal/app/execution_mode.go (new)

- Values: "auto" (default), "ask". Anything else is INVALID, never coerced.
- Precedence: JINI_MODE env → ~/.jini/mode.json → "auto".
- **Fail-closed parsing**: unreadable file, unmarshal error, unknown value
  (file or env), or os.UserHomeDir failure resolves to "ask" with a one-line
  stderr warning naming the bad source ("mode setting unreadable; treating as
  Ask (supervised)"). Absence of the file is NOT an error — that is the
  normal "auto" default.
  Rationale: a corrupted safety toggle must degrade to supervision, not
  autonomy. (Explicitly diverges from router_settings.go's fail-open load.)
- **Atomic write**: temp file in same dir + os.Rename, 0o600. JSON shape:
  {schema_version "0.1.0", context_type "JiniExecutionMode", mode}.
- `runMode(args, stdout, stderr)`:
  - bare: prints effective mode, and when JINI_MODE overrides the file, names
    the winning source: "Ask (from JINI_MODE; saved setting is Auto)."
  - `jini mode ask|auto`: writes atomically, confirms. If JINI_MODE is set,
    the confirmation warns that the env var still wins in this shell.
  - invalid arg: `Unknown mode "banana". Use `jini mode auto` or `jini mode ask`.`
    exit 1.
- Registration (4 sites): dispatch switch (app.go ~112), canonicalTopLevelCommand
  (app.go ~5393), validateNativeArgs (loose: len ≤ 2, any second arg passes —
  so the friendly rejection copy in runMode is actually reachable rather than
  being shadowed by the generic "Unsupported arguments" pre-dispatch error at
  app.go:104-110), help listing. PLUS the real gate: the allowed-command map
  in go_migration_test.go:90 gains "mode" — a named spec amendment recorded
  in specs/product-settling-decisions.md, not a quiet test edit.

### Approval seam in throttle_survival.go

A small interface, not a bare func — sized for the two known future consumers
(paid Autopilot route switching, remote Ask-mode mobile approvals):

    type throttleApprovalRequest struct {
        Label        string        // route label
        Wait         time.Duration // advertised/backoff wait for this hold
        FallbackHint string        // lazily-probed viable fallback ("" if none)
        Hold         int           // 1-based hold number
        TaskTitle    string        // short prompt excerpt — what is being approved
                                   // (needed later by the mobile approval card)
    }
    type throttleApprovalDecision int // approvalGranted | approvalDeclined
    type throttleApprover interface {
        Approve(ctx context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error)
    }

- `runWithThrottleSurvival` gains an approver parameter. Consulted once,
  before the FIRST hold only; the grant is cached loop-locally (decision 4).
- Auto mode passes `autoApprover` (always grants, zero IO) — the shipped
  free-tier path stays byte-identical in behavior and output.
- One-shot CLI path in Ask mode passes `cliPromptApprover`: injectable TTY
  check (default: term check on os.Stdin fd) + injectable reader/writer
  (default os.Stdin/os.Stderr). Not a TTY → decline without prompting.
  Prompt: `{label} is throttled (retry in {wait}). Resume automatically when
  capacity returns? [y/N] `. Drains pending buffered input first; only an
  explicit y/yes grants.
- Interactive launcher and sidecar paths in Ask mode pass `failClosedApprover`
  (declines without prompting). No raw stdin read ever occurs under the
  launcher's Scanner or the sidecar's protocol pipe.
- Decline (any source) returns a typed `throttleDeclinedError` carrying label,
  reason, and fallback hint — never the generic exhaustion error.

**Approver-selection conduit (round-2 must-fix 3).** The shared callers in
`generateWithConfiguredProviderDecision` (provider.go:129,155) cannot know
which entry path invoked them, so the approver is a process-level default —
a package var mirroring the existing `throttleNarration`/`throttleSleep`
injection pattern:

    var throttleApproverForProcess throttleApprover = autoApprover{}

Set at the DISPATCH BRANCHES, not at `Run`/`RunInteractive` — `Run`
unconditionally delegates to `RunInteractive` (app.go:69-71) and the
one-shot/launcher split happens inside it (app.go:82-99), so an entry-function
set-site would hand the launcher the prompting approver. Concretely: the
one-shot direct-task intake and standalone-question branches set
`cliPromptApprover` when effective mode is "ask"; `runLauncher`,
`runNewWorkIntake`, and the sidecar serve loop set `failClosedApprover` when
mode is "ask"; mode "auto" leaves the default everywhere. Each invocation
reaches exactly one of these branches before provider work runs, so the
package var is set-once-then-read — same safety profile as the existing
narration/sleep vars, and tests override it the same way. This is a
deliberate choice over threading an approver parameter through the provider
call stack (no-plumbing virtue kept, but the mechanism is now named, not
implied).

### Park record — making the guidance true

Round-1 veto: "session is saved; `jini continue` resumes it" was false on both
dominant paths (simple_answers.go:39-44 swallows the error;
app.go:2460→2475 saves current-work only AFTER the provider call). Fix:

- New park file at sessionStateRoot()/throttle-park.json (per-project, like
  other session state), 0o600 (it stores the user's prompt): {schema_version,
  context_type "JiniThrottlePark", prompt, route label, reason, fallback hint,
  parked_at}.
- **Layering (round-2 must-fix 1)**: the CALLER owns the park, not the
  survival loop. On every promptable path with the original prompt in hand
  (provider path via request.Source, provider.go:124-127; standalone question
  via raw, simple_answers.go:34), when effective mode is "ask", the caller
  atomically pre-writes the park BEFORE invoking `runWithThrottleSurvival`.
  The park is deleted only on SUCCESS of the overall call — not on approval
  grant. Consequences: the approver needs no park callback (its request
  struct stays as specified); Ctrl-C at any point — during the prompt, during
  a post-grant hold (throttle_survival.go:199-201 returns sleepErr directly),
  or mid-generation — leaves the park in place and the work resumable. A
  clean non-throttle completion or non-throttle error also deletes the park
  (nothing to resume). The pre-write necessarily omits `reason` and
  `fallback hint` (a throttle hasn't happened yet; the fallback is lazily
  probed on first hold) — the caller updates the park with both on the error
  return, before surfacing the error. Auto mode never writes a park
  (unchanged shipped behavior); the exhaustion path in Auto keeps its current
  copy, which only mentions `jini continue` where current-work provides it.
- **Lifecycle (round-2 must-fix 2)**: starting any new work clears a stale
  park — a `clearThrottlePark()` call at the same sites that call
  `saveCurrentWork`, so a park can never shadow newer work. `runContinue`
  checks throttle-park.json first; if present it discloses age when older
  than one hour ("parked 3 days ago: <prompt excerpt> — resuming") and
  re-runs the parked prompt through the normal route engine (mode consulted
  fresh — the user may have switched to Auto), deleting the park on success.
  Corrupt park file → clean fallthrough to current-work behavior. Two
  concurrent processes in one directory share one park slot: last-writer-wins
  on write, and one process's delete-on-success may remove the other's park;
  atomic rename prevents tearing, and both accepted loss modes are documented
  in the file header (single-user CLI).
- Error copy discipline: the "session is saved; `jini continue` resumes it"
  sentence appears ONLY when a park record was actually written. Where no park
  is possible, the error names the throttle reason and the fallback
  (`jini route set {fallback}`) without claiming a save.
- `standaloneQuestionSetupMessage` (simple_answers.go:39-44) is changed to
  pass throttle-family errors (throttled/declined/exhausted) through verbatim
  instead of replacing them with the generic "Route unavailable" message.

### Context-deadline interaction

Round-1 finding: the standalone-question path's 10s deadline
(simple_answers.go:58-65) kills the prompt AND already kills the shipped 20s
first hold — a pre-existing defect in the shipped P0 on that path.

- **Mechanism (round-2 must-fix 4)**: `runWithThrottleSurvival` gains an
  `attemptTimeout time.Duration` parameter (0 = none, current behavior). When
  set, each `attempt()` runs under `context.WithTimeout(parentCtx, budget)` —
  derived from the parent so Ctrl-C cancellation is preserved across holds;
  holds and the approval prompt run under the parent ctx, outside the budget.
  Blast radius acknowledged: touches both call sites in provider.go and the
  sidecar path (which passes 0, keeping context.Background() semantics).
  The standalone-question path creates no whole-call deadline anymore; it
  passes its 10s as attemptTimeout.
- **Wall-clock cap for simple questions**: the standalone path uses a
  single-hold ladder (one 20s hold, or the advertised Retry-After capped at
  60s) instead of the full 20/40/80 ladder — worst case ≈ 90s for a simple
  question, honoring "simple questions get simple answers". The full ladder
  stays on the work paths. Ladder selection is a parameter alongside
  attemptTimeout, not a global.
- The approval prompt honors ctx cancellation: the blocking read runs in a
  goroutine; select on ctx.Done() vs the answer, mirroring sleepWithContext's
  shape. The goroutine blocked on stdin leaks on cancellation — accepted for
  a one-shot process, noted in a comment. Park ordering is the caller-owned
  pre-write above: cancellation or decline at any stage leaves the park in
  place, so Ctrl-C never loses parked work.

### Known exemption (documented, backlogged)

generateConsistencyDraft/generateRefinedDraft call generateProviderText
directly (provider.go:174,181) and bypass throttle survival entirely — a gap
in the SHIPPED core, out of this slice's scope. Recorded in the trace backlog
as residual hardening; wrapping them goes with the throttle-dodge counter
work.

### Testing

- Auto path: byte-identical output on the shipped fake-throttle harness
  (regression guard).
- Ask one-shot: prompts once, grant covers later holds; explicit-n declines;
  bare-Enter declines; buffered type-ahead newline does NOT approve.
- Ask no-TTY / launcher / sidecar: decline without prompting, park written,
  error copy contains `jini continue` only when park exists.
- Park roundtrip: decline → park file → `jini continue` re-runs prompt →
  park deleted on success. Corrupt park file → clean "nothing parked"
  fallthrough. Grant then Ctrl-C during a later hold → park still present.
  New work start clears a stale park. Old park discloses age on resume.
- Standalone single-hold ladder: simple question worst case bounded (~90s).
- Mode parsing: corrupt file → "ask" + warning; JINI_MODE=garbage → "ask" +
  warning; env-over-file precedence; atomic write leaves no temp droppings.
- `jini mode` copy incl. env-override annotation; invalid-arg rejection.
- Per-attempt timeout: standalone question with one 20s hold now survives
  (was: killed at 10s — covers the pre-existing defect).
- Live transcript: one-shot Ask run against the fake-throttling CLI —
  prompt shown, `n` declines, park written, `jini continue` resumes.

### Sequencing (4 days to deadline)

1. This design lands only after the dogfood-fix agent's app.go/cli_handoff.go
   diff is reviewed and committed (avoid merge hell in app.go).
2. First commit is the small registration + allowed-map + settling-doc
   amendment; feature code lives in execution_mode.go / throttle_survival.go
   where contention is nil.
3. Remaining backlog unaffected: the approver interface is the seam Autopilot
   switching and remote mobile approvals extend.
