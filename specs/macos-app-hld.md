# macOS App HLD

Updated: 2026-06-09

This high-level design translates [macos-app-prd.md](./macos-app-prd.md) and
[macos-app-ux-design.md](./macos-app-ux-design.md) into architecture
boundaries. It is subordinate to [number-one-platform-prd.md](./number-one-platform-prd.md)
and [product-settling-decisions.md](./product-settling-decisions.md).

## Scope

This HLD covers the macOS app only.

It does not activate:

- Windows desktop implementation
- mobile implementation
- commercial agent-suite UI
- company automation loops
- a second session model
- a new Jini conversation grammar

## Research Inputs

The design uses these current platform constraints:

- Codex desktop sets the competitive bar for a desktop command center with
  projects, threads, worktrees, diff review, approvals, shared CLI history, and
  skills.
- Codex CLI and Claude Code set the first-minute bar: compact answers, local
  file work, explicit command approval, and familiar terminal-agent behavior.
- Apple direct distribution requires Developer ID signing, hardened runtime,
  notarization, launch testing under Gatekeeper, and a trustworthy update path.
- Tauri 2 requires explicit capabilities per window and signed update artifacts.

## Stack Decision

Phase 1 uses a Tauri 2 shell over the Go core.

Reason:

- The current repo already selected Tauri 2 as the default desktop shell path.
- Jini's product logic must stay in Go: intent, routing, sessions, approvals,
  offline state, diagnostics, tier boundaries, and gates.
- A Tauri shell can give a small desktop surface while reusing the Go core and
  preserving a future Windows path.
- A native SwiftUI/AppKit rewrite is not justified before a prototype proves
  Tauri cannot meet native trust, accessibility, performance, or desktop
  fidelity.

Decision boundary:

- Go owns product runtime behavior.
- Tauri owns the macOS window, shell packaging, updater, webview, and native
  integration glue.
- TypeScript/Rust code must stay presentation and IPC only. It must not own
  tier policy, route choice, prompt construction, session state, approval
  policy, or diagnostics redaction.
- SwiftUI/AppKit may be added only for narrow native controls when the HLD/LLD
  contract remains unchanged.

Native escape hatch:

- If the prototype fails keyboard/VoiceOver quality, sandbox constraints,
  file-open affordances, signed update safety, or desktop performance targets,
  stop app implementation and update this HLD with a SwiftUI/AppKit decision.
- Do not patch around a failed shell with duplicated product logic in the UI.

## Architecture Overview

```text
macOS app

Tauri desktop shell
- window, sidebar, thread, inspector, composer
- menus, shortcuts, notifications, sheets
- updater and platform packaging
- capability-scoped IPC client

Desktop bridge
- launches signed Go sidecar
- speaks versioned JSON-RPC over stdio or Unix domain socket
- maps app events to Go core commands
- streams turn, progress, diff, route, approval, and diagnostic events

Go app core
- intent and command classification
- simple answers
- local file edits
- routing and CLI handoff
- approvals and command policy
- session projection
- artifact and diff projection
- offline state and sync debt
- diagnostics and redaction
- gates and ship checks

Local state
- project files
- .jini work/session state
- route evidence
- artifacts and diffs
- diagnostics cache
```

## Runtime Layers

### 1. macOS Presentation Shell

Responsibilities:

- render the three-pane UX
- own toolbar, sidebar, thread, inspector, composer, menus, and sheets
- render streamed Go events
- collect user input and approval decisions
- request file/folder access through native affordances
- preserve accessibility and keyboard focus
- show offline, route, power, and approval state

Must not:

- classify intent
- infer whether a question is simple
- decide whether to create work
- choose a route
- mutate project files directly
- call downstream CLIs directly
- redact diagnostics itself as the only protection
- persist a private session store

### 2. Desktop Bridge

Responsibilities:

- start and supervise the Go sidecar
- perform bridge handshake and protocol-version negotiation
- enforce request/response correlation
- route UI requests to Go commands
- stream Go events to UI view models
- fail closed when protocol version, signature, or executable trust is invalid

Transport:

- Phase 1 uses stdio JSON-RPC for the Go sidecar.
- Unix domain socket is allowed if stdio cannot meet streaming or cancellation
  needs.
- Local TCP is not allowed for Phase 1 because it expands the attack surface
  without product value.

### 3. Go Core

Responsibilities:

- provide a single app-facing API over existing Go runtime capabilities
- preserve the same semantics as CLI front-door behavior
- project existing and future session state into desktop view models
- execute local file edits, route handoffs, and local/provider work through
  the same action boundary as CLI
- create approvals before side effects
- write route evidence, events, artifacts, and offline debt
- produce redacted diagnostics

### 4. Local Persistence

Responsibilities:

- keep the app and CLI on the same state graph
- support existing `.jini/work` and current-work state during transition
- support the target `.jini/sessions` event model when the session graph lands
- keep state plain-file inspectable where possible
- keep secrets out of state files

## Trust Boundary

```text
User input
  -> Tauri renderer
  -> capability-scoped bridge command
  -> Go sidecar API
  -> Go intent/action/state boundary
  -> explicit approval if side effect
  -> project files, routes, CLIs, providers, local models, or state store
```

Only the Go sidecar may perform side effects.
Renderer can request side effects but cannot perform them.

Side effects include:

- file writes
- shell commands
- downstream CLI execution
- provider or gateway network calls
- local model runtime launch when it consumes material resources
- commit, push, PR, send, publish, or export
- diagnostics export

## Tauri Capability Model

App windows:

- main window: thread, sidebar, inspector, composer
- approval sheet: explicit approval decisions
- diagnostics export sheet: support-bundle preview and export

Capabilities:

- allow only named bridge commands
- no broad filesystem read/write from renderer
- no shell execution from renderer
- no unrestricted external link opening
- no secrets in renderer-local storage
- no app-wide command that bypasses the Go approval boundary

Every new capability requires:

- PRD or HLD trace
- threat note
- focused test or manual security checklist item
- product-settling update if it changes user-visible scope

## Primary Data Flows

### First Launch

1. Tauri launches main window.
2. Bridge starts Go sidecar.
3. UI sends `app.handshake`.
4. Go returns app snapshot: version, protocol, route state, recent projects,
   setup warnings, and no selected session.
5. UI renders compact prompt.

Invariant: saved work is passive. The launch flow must not force a current-work
frame.

### Open Project

1. User chooses folder through native picker.
2. UI sends `project.open`.
3. Go validates path, trust posture, repo metadata, and `.jini` state.
4. Go returns project snapshot, session summaries, route health, and setup
   warnings.
5. UI updates sidebar and toolbar without creating a session.

### Submit Turn

1. User enters text and optional attachments.
2. UI sends `turn.submit` with project id, optional session id, input text,
   attachment refs, route preference, and approval mode.
3. Go classifies intent.
4. Go may return one of:
   - compact answer
   - clarification ask
   - plan
   - approval request
   - action receipt
   - streaming progress
   - error with recovery
5. UI renders the thread and inspector from streamed events.

Invariant: simple factual questions stop before work creation, route ceremony,
and artifact projection.

### Direct File Edit

1. Go resolves file intent.
2. Ambiguous target returns a bounded ask with candidate files.
3. Clear target enters approval policy if required.
4. Go edits file.
5. Go returns changed-file receipt and diff projection.
6. UI opens Diffs inspector.

Invariant: no Working Draft for clear file edits.

### Approval

1. Go emits `approval.requested`.
2. UI renders sheet with action, target, route, risk, and choices.
3. User chooses allow once, scoped allow, or deny.
4. UI sends `approval.resolve`.
5. Go verifies request id and scope, then continues or aborts.

Invariant: approval ids are one-time tokens scoped to a session, action, and
target.

### Route Handoff

1. Go evaluates route policy.
2. For CLI handoff, Go verifies executable, trust, args template, auth/dogfood
   evidence where available, and Gatekeeper state.
3. If route is unavailable, Go returns setup guidance.
4. If route is approved and executable, Go launches downstream CLI through Go.
5. Go records route receipt and streams progress.

Invariant: `codex` route never silently becomes provider API.

### Offline Resume

1. Connectivity or selected route changes to offline.
2. Go returns offline state and available local routes.
3. Existing sessions and artifacts remain openable.
4. New local events accrue sync debt.
5. UI shows offline indicator and sync debt in Route or Progress inspector.

Invariant: offline work appends to the same session graph.

### Diagnostics Export

1. UI requests diagnostics preview.
2. Go builds redacted manifest.
3. UI shows included categories and exclusions.
4. User approves export.
5. Go writes bundle and returns path.

Invariant: prompts, provider payloads, secrets, and raw artifact bodies are
excluded by default.

## Session And State Strategy

The app must bridge two states:

- current implementation: `.jini/work`, `current-work.json`, route receipts,
  artifacts, and CLI-focused projections
- target implementation: stable `.jini/sessions/<session-id>` with append-only
  events, projections, artifacts, route evidence, and sync debt

Phase 1 strategy:

- build Go projectors that expose one desktop view model over current state
- do not let the UI learn current-state file layout
- keep desktop view models compatible with the target session graph
- migrate persistence later behind the Go API without changing the app UX
- introduce a `SessionIdentityMap` that binds current `.jini/work` pack dirs,
  `current-work.json` records, target `.jini/sessions/<session-id>` records,
  and app-facing ids without creating duplicate sessions
- preserve one stable app-facing `session_id` across CLI and app even while
  persistence is backed by legacy current-work state
- never create a session identity for transient simple answers

## macOS File Access Posture

Phase 1 direct distribution uses Developer ID signing, hardened runtime,
notarization, stapling, and Gatekeeper launch checks.

The sandbox decision is explicit:

- Internal dogfood may start hardened-runtime-only while the app is distributed
  directly and file access stays user-selected through the project picker.
- Public alpha must either keep hardened-runtime-only direct distribution with a
  documented entitlement rationale or adopt App Sandbox with security-scoped project bookmarks.
- If App Sandbox is adopted, the Tauri shell obtains the bookmark and the Go
  sidecar receives only a validated project access token plus a resolved
  project path through the bridge.
- The renderer never receives a broad filesystem capability.
- File writes remain project-scoped and approval-gated in Go.

The implementation cannot proceed to public alpha until the chosen entitlement
posture is captured in release gates and smoke-tested on a clean Mac.

## Error Handling

Failures must be actionable:

- missing CLI: executable name and setup path
- Gatekeeper rejected CLI: fail closed and point to `jini doctor`
- protocol mismatch: require app/core update
- sidecar crash: show recovery, preserve unsent input, offer restart
- file permission denied: show target and permission recovery
- route offline: show local alternatives and sync debt
- state corruption: preserve bundle, show safe reset or repair path

## Security Architecture

Release builds must satisfy:

- Developer ID signing
- notarization
- stapling
- hardened runtime
- Gatekeeper launch smoke
- updater signature verification
- least-privilege Tauri capabilities
- sandbox entitlement review before public alpha
- no secrets in renderer storage, logs, crash reports, update manifests, or
  diagnostics bundles
- explicit approval before side effects

Secrets:

- provider credentials remain outside the renderer
- app never prints secrets in route setup guidance
- future account tokens use platform keychain or the existing secure storage
  decision for the account surface

## Performance Architecture

Targets for prototype:

- cold launch to prompt: under 2 seconds on target dogfood Mac
- warm resume to latest session: under 700 ms
- open diff inspector for small edit: under 300 ms
- route status refresh: under 1 second for local checks
- diagnostics preview: under 3 seconds for normal project

Policies:

- do not prewarm local models on launch
- do not scan full repos on launch
- lazy-load inspector data
- keep first launch independent of route dogfood checks
- stream long operations

## Accessibility Architecture

The presentation shell must expose:

- deterministic focus order: sidebar, thread, inspector, composer
- labeled toolbar route and offline state
- accessible approval sheets
- diff additions/deletions as semantic labels, not color only
- reduced motion handling
- keyboard-only completion of first launch, project open, edit, approval, route
  help, diff review, and diagnostics export

## Commercial Boundary

Free macOS app:

- same session model
- project/session shell
- direct answers and local file work
- manual route visibility and switching
- basic offline/local route support when configured
- diagnostics and setup help

Commercial macOS app:

- managed route policy
- governed skills and delegation
- automatic throttle recovery
- team audit and approval policy
- managed cross-device sync

The shell may render commercial capabilities later, but the Phase 1 free app must not expose agent-role theater or a skills OS as default UX.

## HLD Gates

Before implementation starts:

- [macos-app-lld.md](./macos-app-lld.md) defines the app protocol and view
  models.
- Product tests protect this HLD and LLD as drift-controlled docs.
- Prototype plan validates Tauri accessibility, IPC, update, and signing risks.

Before alpha:

- app-to-CLI and CLI-to-app session continuity test passes
- simple answer and direct edit golden transcript pass inside app shell
- route handoff unavailable and Gatekeeper-rejected states are visible
- redacted diagnostics export passes
- signed and notarized clean-machine launch passes
