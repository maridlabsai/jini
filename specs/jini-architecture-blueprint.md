# Jini Architecture Blueprint — v1 (Option A scope)

Produced: 2026-07-09 · Ground truth: `specs/prd-rebuild-design.md` (decision record; the rebuilt `number-one-platform-prd.md` carries the same content once the sweep lands). Scope: v1 Option A — BYO keys, gateways, installed-CLI handoffs, sessions, receipts/ledger, throttle fallback, skills/agents, fail-closed paywall; local execution via detected third-party runtimes as a disclosed interim route (§3).

Global invariants binding every contract below:
- Answer/receipt first, Claude Code/Codex-familiar output, no invented vocabulary (§4.1).
- Token frugality is P0 runtime behavior (§4.6).
- Truth-in-labeling for routes (§4.3). Fail closed with exact guidance, never generic errors (§4.1, §4.3).
- Paid features fail closed; a manual free equivalent is always constructible in the same code path (§4.9).
- Savings are typed literal or imputed; the label survives every rendering (§4.5, §5).
- Terminal experience bar (PRD UX contract): state-adaptive surface, sub-100ms first paint, narrated progress instead of spinners, semantic-but-never-color-only output, NO_COLOR/--plain first-class. Every INJECTION PROMPT block below implicitly requires: exact output layout for each state (idle/executing/completed/throttled/resuming), the state transitions between them, and accessibility assertions (a golden-transcript test per state in --plain mode).

---

## Task 1 — Route Engine & Escalation Ladder (PRD §4.3)

### 1.1 Route-ladder state machine (per task)

```
                 ┌──────────┐
                 │ CLASSIFY │  task → TaskShape{PromptClass, quality bar}   (§4.3, §4.6)
                 └────┬─────┘
                      ▼
                 ┌──────────┐  no candidate clears quality bar
                 │  SELECT  ├────────────────────────────────► FAIL_CLOSED
                 └────┬─────┘        (SetupGuidanceError: exact missing rung + setup steps)
        pinned route │ or auto: cheapest rung clearing bar,
                     │ scored by benchmark + OutcomeStore (§4.4)
                      ▼
                 ┌──────────┐  selected rung costs money AND not pre-approved
                 │  QUOTE   ├────────────────────────────────► AWAIT_APPROVAL
                 └────┬─────┘        approve ─► EXECUTE   deny ─► SELECT(next) | FAIL_CLOSED
                      ▼
                 ┌──────────┐  ThrottleEvent (429/529/quota/window)
                 │ EXECUTE  ├────────────────────────────────► THROTTLED
                 └────┬─────┘        free tier:      SUGGEST_FALLBACK (halt; receipt + `jini continue` hint)
                      │              paid+entitled:  AUTO_REROUTE ─► SELECT (resume from compact state §4.4)
                      ▼
                 ┌──────────┐
                 │ RECEIPT  │  always emitted, success or failure (§4.5)
                 └──────────┘
```

Transition invariants:
- `QUOTE → EXECUTE` on a paid rung requires an explicit approval event recorded in the receipt (§4.3 quote-before-spend).
- `THROTTLED → AUTO_REROUTE` exists only behind `Entitlement.Check(FeatureAutopilot)`; the free branch is the same `FallbackPolicy` interface without the decorator (§4.9).
- Every terminal state (COMPLETED, FAIL_CLOSED, halted THROTTLED) emits a receipt. No silent exits.

### 1.2 Go interface contracts

```go
package route

// Ladder order is fixed and product-visible (§4.3).
type Rung int

const (
	RungLocal Rung = iota // v1: detected third-party runtime, disclosed interim (§3)
	RungBYO
	RungGateway
	RungCLIHandoff
)

type RouteKind int

const (
	KindLocalRuntime RouteKind = iota
	KindProviderAPI
	KindGatewayAPI
	KindInstalledCLI
)

type Route interface {
	ID() string   // e.g. "claude-code", "byo:anthropic", "gateway:openrouter"
	Rung() Rung
	Kind() RouteKind
	CostModel() CostModel // -> ledger.LiteralPricing | ledger.SubscriptionImputed (§4.5)
	Probe(ctx context.Context) ProbeResult // executable/key/daemon presence; cached per session
}

// Truth-in-labeling is enforced at registration, not by convention (§4.3):
// an InstalledCLI route can only be built from a resolved executable.
type InstalledCLIRoute struct {
	id, executablePath string
	argsTemplate       []string
}

// NewInstalledCLIRoute fails closed when the executable is absent.
func NewInstalledCLIRoute(id string, probe ExecProbe) (*InstalledCLIRoute, error)
// error is always *SetupGuidanceError — never a generic error.

type SetupGuidanceError struct {
	RouteID string
	Missing string   // "codex executable not found on PATH"
	Steps   []string // exact, copy-pasteable setup steps
}

type TaskShape struct {
	Class      PromptClass // simple_question | file_edit | multi_file | agent_dispatch (§4.6)
	QualityBar Score       // absolute floor per class; local below bar triggers escalation (§6)
	SizeHint   Bucket
}

type Selector interface {
	Select(ctx context.Context, shape TaskShape, prefs Prefs) (Decision, error)
}

type Decision struct {
	Route Route
	Why   Explanation // rendered verbatim by `jini route explain` (§4.3)
	Quote *Quote      // non-nil iff rung spends money without prior approval
}

type Quote struct { // §4.3 quote-before-spend; reused by agent fan-out (§4.10)
	RouteID    string
	EstTokens  TokenRange
	EstCost    ledger.MoneyRange // carries the literal/imputed label (§4.5)
	Escalation string            // "local scored 41 vs bar 70 on multi_file"
}

type ThrottleDetector interface {
	Classify(routeID string, err error) (ThrottleEvent, bool)
}

type ThrottleEvent struct {
	RouteID   string
	Kind      ThrottleKind // RateLimited | QuotaExhausted | WindowReset(t)
	RetryAt   *time.Time
}

// FallbackPolicy is the FREE behavior. Autopilot wraps it; it never replaces it (§4.9).
type FallbackPolicy interface {
	OnThrottle(ev ThrottleEvent, ladder []Route) Suggestion // suggest only, never act
}
```

### 1.3 BYO near-zero-effort layer (§4.3)

Sequence contract (first run and `jini route add`):

```
detect ─► present ready routes ─► [user pastes key OR confirms detected] ─► validate (1 live call)
      ─► RouteProfile ─► confirm ─► store (keychain) ─► route usable; receipt of what was stored where
```

```go
package byo

type Source struct {
	Kind SourceKind // EnvVar | CLIConfig | GatewayConfig
	Name string     // "ANTHROPIC_API_KEY", "claude-code", "openrouter"
}

type DetectedCredential struct {
	Source   Source
	Provider string
	Ref      SecretRef // opaque handle; the secret value never sits in this struct
}

type Detector interface {
	Detect(ctx context.Context) []DetectedCredential
}
// v1 detection matrix (release-gated compatibility, §4.3): env vars ANTHROPIC_API_KEY,
// OPENAI_API_KEY, GEMINI_API_KEY, DEEPSEEK_API_KEY, MISTRAL_API_KEY, GROQ_API_KEY,
// OPENROUTER_API_KEY; CLI configs for Claude Code (Pro/Max), Codex/ChatGPT plans,
// Gemini CLI; LiteLLM/OpenRouter gateway configs. Each shape carries its
// receipt-denomination rule: metered key -> LiteralSavings, subscription CLI ->
// ImputedSavings (§4.5). A shape without a passing validation fixture is not
// claimed (§5). Detected credentials are OFFERED, never auto-activated (§4.7).

type Validator interface {
	// Exactly one minimal live call (models-list or 1-token ping; hard spend cap).
	Validate(ctx context.Context, ref SecretRef) (RouteProfile, error)
}

// Error taxonomy — exact, never generic (§4.3):
type ValidationErrKind int

const (
	ErrBadKey ValidationErrKind = iota // auth rejected
	ErrNoQuota                          // key valid, no credit/quota
	ErrWrongRegion                      // key valid, endpoint/region mismatch
	ErrNetwork                          // transport failure; key not judged
)

type ValidationError struct {
	Kind    ValidationErrKind
	Detail  string   // provider message, secret-scrubbed
	FixHint string   // one exact next step
}

type CredentialStore interface {
	Put(ref SecretRef, secret []byte) error
	Get(ref SecretRef) ([]byte, error)
	Delete(ref SecretRef) error
	Backend() string // "keychain" | "secret-service" | "encrypted-file" — shown in `jini route status`
}
// Backend resolution order: darwin Keychain; Linux Secret Service (DBus); headless
// fallback = age-encrypted file at $JINI_STATE_DIR/credentials.age, mode 0600,
// key from JINI_CREDENTIALS_KEY. Plaintext dotfiles are never written (§4.3).
// Env-var credentials stay in the env — Jini references, never copies, them.
```

### 1.4 `jini route --format json` shape (example instance)

```json
{
  "ladder": [
    {"id": "local:ollama:qwen2.5-coder", "rung": "local", "kind": "local-runtime",
     "status": "ready", "disclosure": "interim third-party runtime; first-party runtime arrives v1.5"},
    {"id": "byo:anthropic", "rung": "byo", "kind": "provider-api", "status": "ready",
     "credential_backend": "keychain"},
    {"id": "gateway:openrouter", "rung": "gateway", "status": "needs-setup",
     "setup": ["export OPENROUTER_API_KEY=... or run: jini route add openrouter"]},
    {"id": "claude-code", "rung": "cli-handoff", "kind": "installed-cli",
     "status": "ready", "executable": true}
  ],
  "pinned": null,
  "auto": {"policy": "cheapest-above-bar", "outcome_samples": 214}
}
```

---

## Task 2 — Durable Sessions & Compact State (PRD §4.4)

### 2.1 Compact session-state schema (NOT chat history)

Checkpointed atomically (write temp + rename) after every tool action and route event; resume = one file read, no transcript replay (<1s target).

```json
{
  "schema_version": "1.0.0",
  "session_id": "s_9f2c",
  "created_at": "2026-07-09T18:02:11Z",
  "task": {
    "goal": "add --json flag to report command",
    "status": "in_progress",
    "constraints": ["no new deps", "keep exit codes"]
  },
  "workspace": {
    "repo_root_hash": "sha256:…",
    "focus_files": [
      {"path": "internal/report/cli.go", "content_hash": "sha256:…",
       "summary": "flag parsing; JSON encoder added at L88"}
    ],
    "dirty": ["internal/report/cli.go"]
  },
  "progress": {
    "done": ["flag added", "encoder wired"],
    "next": ["update golden test"],
    "blockers": []
  },
  "route_history": [
    {"route_id": "byo:anthropic", "outcome": "throttled",
     "tokens": {"prompt": 5210, "completion": 902}, "ended_at": "…"}
  ],
  "route_native_context": {"claude-code": {"ref": "opaque-session-ref", "expires_at": "…"}},
  "fidelity": {
    "carried": ["task", "workspace.focus_files", "progress"],
    "dropped": ["model_scratch_reasoning"]
  }
}
```

### 2.2 Resume flows and fidelity disclosure

- **Same-route resume (lossless guarantee):** if `route_native_context[route]` is present and unexpired, resume via the route's native continuation; compact state is the checkpoint of record. Receipt states `resume: lossless (native continuation)`.
- **Cross-route resume (best-effort, disclosed):** rebuild working context from compact state only. The resume receipt is mandatory and silent context loss is a defect (§4.4):

```json
{
  "resumed_from": "byo:anthropic",
  "resumed_to": "claude-code",
  "carried": ["goal", "progress", "focus_files (3)"],
  "lost": ["in-flight reasoning", "unsummarized reads (2 files)"],
  "note": "cross-route resume is best-effort; re-verify the last edit before continuing"
}
```

- Crash/reboot resume: identical to same-route with native context expired → falls to compact-state path with disclosure.

### 2.3 Route-outcome learning store (§4.4)

```go
package memory

type TaskFingerprint struct {
	Class      route.PromptClass
	LangHints  []string // ["go"], ["ts","css"]
	SizeBucket Bucket   // S | M | L by files touched + prompt tokens
	RepoKey    string   // salted local hash; never leaves the machine
}

type Outcome struct {
	RouteID   string
	Success   bool
	Tokens    ledger.Usage
	Throttled bool
	At        time.Time
}

type OutcomeStore interface {
	Record(fp TaskFingerprint, oc Outcome) error
	Score(fp TaskFingerprint, routeID string) (score float64, samples int) // feeds Selector
	List(f Filter) ([]Entry, error)  // `jini memory`
	Export(w io.Writer) error        // `jini memory export` (JSON)
	Delete(f Filter) error           // `jini memory delete` — full wipe supported (§4.4)
}
```

Storage: single JSON-lines file under `$JINI_STATE_DIR/memory/outcomes.jsonl`; no personal profile fields exist in the schema (§4.4 "no personal profile building in v1").

### 2.4 Project memory — the context engine (§4.4, primary §4.6 saver)

Per-repo learned context that substitutes for repeated discovery reads. Plain markdown files under `.jini/memory/` (inspectable, diffable, deletable): `commands.md` (build/test/run), `conventions.md`, `architecture.md` (learned summaries with content-hash freshness), `pitfalls.md`. Injection is selective by task shape:

```go
type ProjectMemory interface {
	// Recall returns at most Budget-permitted tokens of memory relevant to the
	// task shape — never the whole store (§4.6).
	Recall(shape route.TaskShape, budget int) ([]MemoryChunk, error)
	// Learn appends distilled facts after task completion; content-hash keyed
	// so stale summaries are invalidated when files change.
	Learn(sessionID string, facts []Fact) error
}
```

Token accounting: every `Recall` that replaces a would-have-been file read credits `tokens.kept_off_metered` via the receipt — project memory's savings are measured, not asserted (§4.5 I4-adjacent).

---

## Task 3 — Savings Ledger & Receipt Engine (PRD §4.5)

### 3.1 Literal vs imputed as a type-level distinction

```go
package ledger

type Money struct{ Cents int64; Currency string }

type PriceRef struct { // provenance for every dollar figure (§5)
	SourceURL   string
	RetrievedAt time.Time
}

// Savings is a sealed sum type: exactly two implementations. They cannot be
// summed or compared without going through Label() — the compiler carries
// the honesty rule (§4.5, §5).
type Savings interface {
	Amount() Money
	Label() SavingsLabel // "literal" | "imputed" — no default, no omission
	Provenance() PriceRef
	sealed()
}

type LiteralSavings struct { /* metered API route: posted prices */ }
type ImputedSavings struct { // subscription route: quota preserved, valued at API-equivalent prices
	Basis string // "anthropic API-equivalent pricing for claude-sonnet-5"
}
```

### 3.2 Receipt schema (every task, success or failure)

```json
{
  "schema_version": "1.0.0",
  "task_id": "t_41ab",
  "route_id": "local:ollama:qwen2.5-coder",
  "outcome": "completed",
  "tokens": {"prompt": 6100, "completion": 840, "kept_off_metered": 6940},
  "time_saved": {"throttle_wait_avoided_s": 0, "tool_switches_avoided": 1},
  "counterfactual": {
    "route_id": "byo:anthropic",
    "cost": {"cents": 9, "currency": "USD"},
    "label": "literal",
    "pricing": {"source_url": "https://…/pricing", "retrieved_at": "2026-07-01T00:00:00Z"}
  },
  "dollars_saved": {"cents": 9, "currency": "USD", "label": "literal"},
  "side_effects": {"files_changed": ["…"], "commands_run": ["go test ./…"], "rollback": "git checkout -- …"}
}
```

`time_saved` fields come only from measured events (an observed throttle window avoided, an observed resume) — never estimated multipliers (honesty invariant I4).

### 3.3 Share-card allowlist (complete field enumeration — nothing else exists)

`jini savings --share` builds a card ONLY from:

| Field | Type |
|---|---|
| `period` | ISO week/month string |
| `task_count` | int |
| `dollars_saved` | Money + label (literal/imputed/mixed — mixed shows both figures separately) |
| `tokens_kept_local` | int |
| `time_saved_minutes` | int (measured only) |
| `jini_version` | semver |

Enforcement: the generator's input is `type AllowlistedCard struct` containing exactly these fields; a reflection test asserts the struct has no other fields and the renderer touches no other data source. No repo names, no paths, no prompts, no hostnames. Opt-in only.

### 3.4 Honesty invariants (executable; wired into the release gate, §5/§7)

- **I1** — type test: no code path sums `LiteralSavings` with `ImputedSavings` into one labeled-literal figure (compile-level + a summation audit test).
- **I2** — provenance: every counterfactual resolves to a pricing-table entry with `SourceURL` + `RetrievedAt`; entries older than 30 days render "stale pricing" on receipts and fail the release gate.
- **I3** — label survival: golden transcripts assert the literal/imputed label appears in text, JSON, and share-card renderings.
- **I4** — measured time only: `time_saved` producers are enumerable; a test walks producers and rejects any constant/multiplier source.

Pricing table: `internal/ledger/pricing.json`, checked in with provenance fields; updated by a scheduled job that opens a PR (never silent). Release gate runs `tools/savings_methodology_gate.sh` = invariant suite I1–I4.

---

## Task 4 — Token Economy Enforcement (PRD §4.6)

### 4.1 Budgets and frugality contracts

```go
package tokeneconomy

type Budget struct {
	Class           route.PromptClass
	MaxPromptTokens int
	MaxOutputTokens int
	MaxFileReads    int
}

// v1 defaults (tunable only with benchmark evidence, §5):
//   simple_question {2000, 300, 0}
//   file_edit       {8000, 1000, 3}
//   multi_file      {32000, 4000, 12}
//   agent_dispatch  {from parent Quote — see §4.10 contract}

type Meter interface {
	// ErrBudgetExceeded forces an explicit choice: degrade scope or escalate
	// WITH a route.Quote. Silent overspend is not a representable state.
	Charge(class route.PromptClass, u ledger.Usage) error
}

type ContextLedger interface {
	// Frugality as a violable interface: strict mode (tests/gate) returns
	// ErrRedundantRead when an unchanged path is re-read in-session;
	// prod mode logs a frugality warning on the receipt.
	NoteRead(path, contentHash string) error
}
```

### 4.2 Token-efficiency regression harness (release-gating)

- Corpus: the 100-prompt first-minute bank + a fixed task corpus, one entry per `PromptClass`.
- Runner: `tools/token_efficiency_gate.sh` executes the corpus against the candidate binary with a fake deterministic route, writing `.jini/token-baseline.json` (`task_id → {prompt, completion}`).
- Pass/fail: per-class median regression > 10% vs the previous release's committed baseline, or any single task > 25%, fails. Baselines are committed per release tag; a baseline update requires the gate output in the PR body (§5 evidence rule).

---

## Task 5 — On-the-fly Skills & Agents + Paywall Boundary (PRD §4.10, §4.9)

### 5.1 Skill/agent file format (portable; Claude Code-compatible frontmatter)

```markdown
---
name: run-migration-checks
description: Use when touching db/migrations — runs the check sequence and summarizes failures
tools: [read, bash]        # permission scope; must be ⊆ creator's scope
---
1. Run `make migrate-dry` …
2. …
```

Agents add `role:` and `budget_tokens:`. Locations: `.jini/skills/` + `.jini/agents/` (repo-scoped) and `~/.jini/skills|agents/` (user-scoped). Files are plain markdown — diffable, reviewable, portable (§4.10).

Creation flow (`jini skill new` / "turn what we just did into a skill"): source is the compact-state `progress` log — never the raw transcript (§4.4); a secret-scrub pass runs before write (see FLAG-3); the file opens for review; the registry watches both dirs so it is invocable in the same session with no restart.

### 5.2 Agent dispatch contracts

```go
package agents

type AgentSpec struct {
	Name        string
	Role        string
	Tools       []Tool       // ⊆ parent scope; enforced at construction
	BudgetTokens int
}

type DispatchRequest struct {
	Spec  AgentSpec
	Quote *route.Quote // REQUIRED for fan-out > 1 or any paid route (§4.10) — same
	                   // quote-before-spend shape as Task 1
}

type DispatchReceipt struct {
	AgentName string
	RouteID   string
	Tokens    ledger.Usage // itemized onto the parent task's receipt (§4.10)
	Outcome   string
}
```

Approval-matrix inheritance (§4.1): an agent's side effects route through the parent's approval flow; an agent can never self-approve commits, pushes, deletions, or network sends.

Repetition-triggered crystallization (§4.10):

```go
type RepetitionDetector interface {
	// Observe consumes compact-state progress logs; when >= 2 similar
	// multi-step sequences recur, emit a Suggestion with the estimated
	// per-run token saving (skill tokens vs re-explanation tokens).
	Observe(sessionID string, progress []Step) []Suggestion
}
```

Suggestions render as one-tap offers ("save this as a skill — saves ~1.8k tokens/run"); never auto-created.

### 5.3 Entitlement interface (§4.9)

```go
package entitlement

type Feature int

const (
	FeatureAutopilot Feature = iota
	FeatureContinuity
)

type Checker interface {
	// Fail-closed: any verification ambiguity (offline, malformed token,
	// clock skew) = not entitled. The error always names the free equivalent.
	Check(f Feature) error // -> *NotEntitledError
}

type NotEntitledError struct {
	Feature        Feature
	FreeEquivalent string // "run `jini continue` after switching routes manually"
}
```

- The binary embeds only a public verification key; token issuance, pricing, and SKUs live in the commercial repo (§4.9 — zero pricing data here).
- Paid features are decorators over free implementations (`Autopilot` wraps `FallbackPolicy`; `Continuity` wraps the local session store). The free path is always constructible without any entitlement object — enforced by a test that builds the full free product with `entitlement.None()`.

---

## FLAGS (boundary/honesty conflicts surfaced, not silently resolved)

1. **FLAG-1 (§4.5/§5):** "Dollars headline" for subscription-only users means the headline number is *imputed*. Any rendering that drops or shrinks the label to make receipts/cards punchier violates §5. Resolution required at UX review: label is part of the number's rendering primitive, not a caption.
2. **FLAG-2 (§4.1/§4.9):** Autopilot auto-reroute mid-task can move code context to a *different provider* than the user last approved. Cost is quoted, but data destination changes too. The approval matrix treats "network sends" as approval-worthy — Autopilot settings must include a destination allowlist, not just spend limits.
3. **FLAG-3 (§4.10):** Skill creation from a session can capture secrets that transited the session (pasted keys, tokens in command output). The creation flow must run a secret-scrub (same scrubber as receipt privacy) before writing the file. Added to the §5.1 flow above as mandatory.

---

## INJECTION PROMPT FOR DOWNSTREAM CODER — Block 1 (Route Engine)

Context: Expert Go systems developer in the Jini repo. Read `specs/number-one-platform-prd.md` §4.3 and `internal/app/cli_handoff.go` (reference for truth-in-labeling) first.
Implement Task 1's contracts in `internal/route/` and `internal/byo/`. TDD every state transition including throttle classification and fail-closed constructors. Validation errors must be the typed taxonomy — a test asserts no generic error string reaches the user. Done means `bash tools/run_required_gates.sh commit` exits 0; paste its tail in your report.

## INJECTION PROMPT FOR DOWNSTREAM CODER — Block 2 (Sessions)

Context: as Block 1; read PRD §4.4. Implement `internal/session/` per Task 2: atomic checkpointing, same-route lossless resume, cross-route resume with mandatory fidelity receipt, and `internal/memory/` OutcomeStore with export/delete. Tests: kill -9 mid-task then resume < 1s; cross-route resume asserts the `lost` list is truthful (seed known-unsummarized reads). Gate green before done.

## INJECTION PROMPT FOR DOWNSTREAM CODER — Block 3 (Ledger)

Context: as Block 1; read PRD §4.5, §5. Implement `internal/ledger/` per Task 3: sealed Savings types, receipt emission on every terminal state, share-card allowlist with reflection test, pricing table with provenance, and `tools/savings_methodology_gate.sh` running invariants I1–I4. Gate green before done.

## INJECTION PROMPT FOR DOWNSTREAM CODER — Block 4 (Token Economy)

Context: as Block 1; read PRD §4.6. Implement `internal/tokeneconomy/` per Task 4 plus `tools/token_efficiency_gate.sh` wired into `tools/run_required_gates.sh` (preserve that script's existing required fragments). Strict-mode ContextLedger tests deliberately violate each frugality rule and assert the typed error. Gate green before done.

## INJECTION PROMPT FOR DOWNSTREAM CODER — Block 5 (Skills/Agents + Entitlements)

Context: as Block 1; read PRD §4.10, §4.9, §4.1. Implement `internal/skills/`, `internal/agents/`, `internal/entitlement/` per Task 5. Tests: full free product constructs with `entitlement.None()`; agent cannot exceed parent tool scope (construction error); dispatch tokens itemize onto parent receipt; skill-creation secret-scrub catches a seeded key. Gate green before done.
