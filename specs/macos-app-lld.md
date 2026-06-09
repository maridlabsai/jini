# macOS App LLD

Updated: 2026-06-09

This low-level design defines the executable contracts for the Jini macOS app.
It implements [macos-app-hld.md](./macos-app-hld.md) and
[macos-app-ux-design.md](./macos-app-ux-design.md).

The LLD is a contract for implementation. It does not claim that the macOS app
exists yet.

## Design Rules

- Go owns product logic.
- The renderer owns presentation only.
- The app and CLI use the same session model.
- Side effects flow through Go approvals.
- Simple questions do not create work.
- File edits produce diffs and receipts, not Working Drafts.
- Route names are truthful: CLI handoff routes invoke real CLIs or fail closed.
- Offline appends sync debt to the same session.
- Diagnostics are redacted by default.

## Process Model

```text
Jini.app
  Tauri shell process
    renderer window
    native menus
    updater
    capability-scoped bridge commands
  Go sidecar process
    jini app serve --stdio
    existing Go app runtime
    project/session projector
    action executor
```

Sidecar command:

```bash
jini app serve --stdio --surface macos
```

This command is planned by this LLD. Until implemented, app work stays in
design/prototype state.

Sidecar lifecycle:

1. Tauri starts sidecar from the signed app bundle.
2. Tauri sends `app.handshake`.
3. Go validates protocol and returns `AppSnapshot`.
4. Tauri subscribes to app events.
5. Tauri restarts sidecar only after a crash or explicit update.

Sidecar crash handling:

- preserve unsent composer text in renderer memory
- mark current operation interrupted
- offer `Restart Jini Core`
- do not replay side effects automatically

## Protocol

Transport:

- JSON-RPC-like messages over stdout/stdin, one JSON object per line
- UTF-8
- request/response plus async events
- no local TCP in Phase 1

Envelope:

```json
{
  "protocol_version": "macos-app-v1",
  "id": "req_01",
  "method": "turn.submit",
  "params": {},
  "surface": "macos",
  "sent_at": "2026-06-09T00:00:00Z"
}
```

Response:

```json
{
  "protocol_version": "macos-app-v1",
  "id": "req_01",
  "ok": true,
  "result": {},
  "events": []
}
```

Error:

```json
{
  "protocol_version": "macos-app-v1",
  "id": "req_01",
  "ok": false,
  "error": {
    "code": "route_unavailable",
    "message": "Codex CLI is blocked by macOS trust checks.",
    "recovery": "Run jini doctor after reinstalling Codex from a trusted source.",
    "safe_to_retry": true
  }
}
```

Async event:

```json
{
  "protocol_version": "macos-app-v1",
  "event_id": "evt_01",
  "event_type": "turn.progress",
  "session_id": "ses_01",
  "payload": {},
  "emitted_at": "2026-06-09T00:00:01Z"
}
```

Required envelope fields:

- `protocol_version`
- `id` for requests and responses
- `method` for requests
- `ok` for responses
- `event_id` and `event_type` for async events
- `surface: macos`

## Method Surface

### App

| Method | Purpose |
| --- | --- |
| `app.handshake` | protocol negotiation and initial app snapshot |
| `app.snapshot` | refresh toolbar/sidebar/global app state |
| `app.subscribe` | subscribe to async state events |
| `app.shutdown` | graceful sidecar shutdown |

### Project

| Method | Purpose |
| --- | --- |
| `project.listRecent` | recent local projects |
| `project.open` | validate and open local folder |
| `project.close` | clear selected project |
| `project.refresh` | refresh repo, route, and session summaries |

### Session

| Method | Purpose |
| --- | --- |
| `session.list` | list sessions for project |
| `session.open` | open session projection |
| `session.resume` | append resume event and return current projection |
| `session.search` | search sessions, artifacts, route receipts |

### Turn

| Method | Purpose |
| --- | --- |
| `turn.submit` | submit user input |
| `turn.cancel` | request cancellation |
| `turn.retry` | retry safe failed turn |

### Approval

| Method | Purpose |
| --- | --- |
| `approval.resolve` | allow once, scoped allow, or deny |
| `approval.listPending` | pending approvals for session |

### Artifact And Diff

| Method | Purpose |
| --- | --- |
| `artifact.list` | list artifacts for session |
| `artifact.open` | open artifact projection |
| `artifact.export` | export artifact through Go |
| `diff.list` | list changed files |
| `diff.open` | open file diff projection |

### Route

| Method | Purpose |
| --- | --- |
| `route.status` | list route health and selected route |
| `route.set` | request route preference |
| `route.help` | setup guidance |
| `route.doctor` | route trust and executable checks |

### Diagnostics

| Method | Purpose |
| --- | --- |
| `diagnostics.preview` | redacted support-bundle manifest |
| `diagnostics.export` | write support bundle after approval |

## View Models

All view models are returned by Go. The renderer may cache them for rendering
but must not treat cached values as product truth.

### AppSnapshot

```json
{
  "app_version": "0.1.0",
  "core_version": "0.1.0",
  "protocol_version": "macos-app-v1",
  "selected_project_id": "proj_01",
  "selected_session_id": "",
  "online_state": "online",
  "offline_state": "available",
  "route_summary": {},
  "setup_warnings": [],
  "recent_projects": [],
  "safe_start_mode": "task_prompt"
}
```

Rules:

- `safe_start_mode` must be `task_prompt` on normal launch.
- No session is auto-opened unless launch context names one.
- Setup warnings are passive unless they block the requested task.

### ProjectVM

```json
{
  "project_id": "proj_01",
  "display_name": "jini",
  "root_label": "jini",
  "root_path_redacted": "~/Developer/jini",
  "git_branch": "main",
  "git_dirty": false,
  "jini_state_state": "ready",
  "route_health": "needs_setup",
  "last_opened_at": "2026-06-09T00:00:00Z"
}
```

Rules:

- UI gets redacted path labels by default.
- Full paths are only shown after explicit user action.
- Opening a project does not create a session.

### SessionSummaryVM

```json
{
  "session_id": "ses_01",
  "title": "Review route setup help",
  "status": "active",
  "last_receipt": "Edited route help test",
  "updated_at": "2026-06-09T00:00:00Z",
  "current_artifact_title": "Route Help Notes",
  "route_label": "local shell",
  "offline_debt_count": 0,
  "requires_attention": false
}
```

Rules:

- Sidebar shows title and last receipt, not raw work ids.
- Session summaries are passive on launch.
- Selecting a session opens it directly.

### SessionDetailVM

```json
{
  "session_id": "ses_01",
  "project_id": "proj_01",
  "title": "Review route setup help",
  "goal": "add regression coverage for route setup help",
  "status": "active",
  "turns": [],
  "ready": [],
  "missing": [],
  "next_action": "Review diff",
  "current_route": {},
  "offline_debt": {},
  "approval_summary": {},
  "selected_inspector_tab": "diffs"
}
```

Rules:

- `selected_inspector_tab` is a Go recommendation; UI can change it locally.
- Opening a session does not regenerate model output.
- Stale projections are rebuilt by Go before response.

### TurnVM

```json
{
  "turn_id": "turn_01",
  "session_id": "ses_01",
  "kind": "action_receipt",
  "user_text": "add tests for route setup help",
  "assistant_text": "Edited internal/app/app_test.go",
  "status": "done",
  "created_at": "2026-06-09T00:00:00Z",
  "artifacts_changed": [],
  "files_changed": ["internal/app/app_test.go"],
  "route_receipt_id": "route_01",
  "approval_request_id": ""
}
```

Allowed `kind` values:

- `compact_answer`
- `clarification_ask`
- `plan`
- `approval_request`
- `progress`
- `action_receipt`
- `route_receipt`
- `error`

Rules:

- `compact_answer` cannot create session state by itself.
- `approval_request` must have an approval id.
- `action_receipt` must state what changed or what ran.

### InspectorVM

```json
{
  "selected_tab": "diffs",
  "progress": {},
  "diffs": [],
  "artifacts": [],
  "route": {},
  "approvals": [],
  "diagnostics": {},
  "sources": []
}
```

Rules:

- Inspector data is lazy-loadable by tab.
- Empty states come from Go so copy stays consistent across surfaces.
- Inspector must not be required for simple answers.

### DiffVM

```json
{
  "diff_id": "diff_01",
  "session_id": "ses_01",
  "file_label": "internal/app/app_test.go",
  "file_path_redacted": "internal/app/app_test.go",
  "change_kind": "modified",
  "additions": 12,
  "deletions": 2,
  "hunks": [],
  "generated_by_turn_id": "turn_01"
}
```

Rules:

- Full paths are hidden unless requested.
- Hunks are loaded on demand.
- Diff colors must have text labels for accessibility.

### RouteEvidenceVM

```json
{
  "route_id": "codex",
  "route_kind": "cli_handoff",
  "label": "Codex CLI handoff",
  "status": "needs_setup",
  "selected": false,
  "reason": "configured route is blocked by local trust checks",
  "model_profile": "",
  "token_posture": "frugal",
  "power_posture": "normal",
  "offline_state": "online",
  "setup_guidance": [
    "Codex CLI is blocked by macOS trust checks.",
    "Run jini doctor after reinstalling Codex from a trusted source."
  ],
  "receipt_id": "route_01"
}
```

Rules:

- `route_kind` must distinguish `cli_handoff`, `provider_api`, `gateway`,
  `local_model`, and `offline`.
- CLI handoff routes do not fall back to provider APIs.
- Setup guidance never includes secrets.

### ApprovalRequestVM

```json
{
  "approval_request_id": "apr_01",
  "session_id": "ses_01",
  "action": "run_command",
  "target_label": "go test ./internal/app -run TestRouteHelp",
  "route_label": "local shell",
  "risk": "reads project files and runs tests",
  "choices": ["allow_once", "allow_for_project_tests", "deny"],
  "expires_at": "2026-06-09T00:10:00Z"
}
```

Rules:

- Approval ids are single-use.
- Scoped allow choices are generated by Go policy, not renderer inference.
- Commit and push approvals are separate requests.

### OfflineDebtVM

```json
{
  "mode": "online",
  "debt_count": 0,
  "summary": "",
  "blocked_routes": [],
  "available_local_routes": [],
  "last_sync_at": "2026-06-09T00:00:00Z"
}
```

Rules:

- Offline work appends debt to the session.
- Remote-route blockers name local alternatives when available.
- Debt clears only after Go confirms reconciliation.

### DiagnosticsPreviewVM

```json
{
  "bundle_id": "diag_01",
  "included": [
    "app version",
    "core version",
    "session ids",
    "route ids",
    "redacted logs"
  ],
  "excluded": [
    "provider keys",
    "raw prompts",
    "artifact bodies",
    "full local paths"
  ],
  "requires_approval": true
}
```

Rules:

- Export requires approval.
- Raw prompts and artifact bodies are excluded by default.
- The preview must match exported bundle contents.

## State Machines

### Turn State

```text
idle
  -> classifying
  -> answering
  -> asking
  -> planning
  -> awaiting_approval
  -> executing
  -> done
  -> failed
  -> cancelled
```

Rules:

- `answering` for simple factual questions ends without session persistence.
- `awaiting_approval` has no side effects until approval resolves.
- `cancelled` does not roll back completed side effects; it stops future work.

### Approval State

```text
requested -> allowed_once -> consumed
requested -> allowed_scoped -> consumed
requested -> denied
requested -> expired
```

Rules:

- `consumed` is terminal.
- `expired` requires a fresh approval request.
- denied approval returns a thread receipt, not a crash.

### Route State

```text
available
needs_setup
blocked_by_trust
blocked_by_offline
blocked_by_policy
running
failed
completed
```

Rules:

- `blocked_by_trust` cannot auto-run.
- `blocked_by_offline` must show local alternatives.
- `failed` includes recovery and safe-to-retry.

### Offline State

```text
online
degraded
offline_available
offline_blocked
sync_pending
sync_conflict
```

Rules:

- `offline_available` means local work can continue.
- `offline_blocked` names the unavailable route or connector.
- `sync_conflict` must not auto-resolve to a riskier review/send boundary.

## Event Types

Required async events:

- `app.ready`
- `project.opened`
- `project.refreshed`
- `session.opened`
- `session.projection_updated`
- `turn.started`
- `turn.delta`
- `turn.progress`
- `turn.completed`
- `turn.failed`
- `approval.requested`
- `approval.resolved`
- `diff.updated`
- `artifact.updated`
- `route.status_updated`
- `offline.state_changed`
- `diagnostics.preview_ready`
- `diagnostics.exported`

Event rule:

- Every event includes `event_id`, `session_id` when applicable, `surface`, and
  `emitted_at`.

## Command Mapping

| UX Action | Method | Go Owner |
| --- | --- | --- |
| first launch | `app.handshake` | app service facade |
| open project | `project.open` | project service |
| submit task | `turn.submit` | intent and action boundary |
| answer simple question | `turn.submit` | simple answer classifier |
| edit file | `turn.submit` | local text edit and action boundary |
| approve command | `approval.resolve` | approval policy |
| route help | `route.help` | route settings and CLI handoff registry |
| route status | `route.status` | router and ship check projections |
| open diff | `diff.open` | diff projector |
| export diagnostics | `diagnostics.export` | diagnostics redactor |

## Go Package Plan

Phase 1 can add packages or files behind existing `internal/app` boundaries:

- `app_bridge.go`: sidecar method dispatch
- `app_protocol.go`: request, response, event, error envelopes
- `app_projector.go`: AppSnapshot, ProjectVM, SessionSummaryVM
- `app_turns.go`: turn submission facade over existing launcher/runtime logic
- `app_approvals.go`: approval request and resolution contracts
- `app_diffs.go`: diff projection
- `app_routes.go`: route evidence projection for desktop
- `app_diagnostics.go`: diagnostics preview and export projection

Rules:

- Existing CLI behavior remains the source of golden transcript truth.
- New app code calls existing core functions or extracts shared services from
  CLI paths.
- Do not duplicate simple-answer, file-edit, route, or provider logic for the
  app.
- If a CLI function cannot be reused without printing CLI chrome, extract the
  product operation behind a renderer-neutral function.

## Renderer Plan

The renderer may use React or the Tauri webview framework selected by the app
prototype.

Renderer components:

- `AppShell`
- `Toolbar`
- `ProjectSidebar`
- `SessionList`
- `ThreadView`
- `Composer`
- `Inspector`
- `DiffPanel`
- `ArtifactPanel`
- `RoutePanel`
- `ApprovalSheet`
- `DiagnosticsSheet`

Rules:

- Components receive Go view models.
- Components do not compute product state.
- Components may keep UI-only state: selected tab, collapsed pane, draft input,
  local focus, scroll position.
- Components do not store secrets or full prompts in local storage.

## Persistence Contract

Phase 1 reads and writes through Go only.

Current-state adapter:

- reads current `.jini/work` and current-work state
- exposes `SessionSummaryVM` and `SessionDetailVM`
- hides existing file layout from renderer
- avoids creating sessions for compact answers

Target-state adapter:

- writes `.jini/sessions/<session-id>/events.ndjson`
- rebuilds projection from events
- stores route evidence and offline debt under session id

Migration rule:

- Renderer must not change when persistence migrates from current work state to
  target session graph.

## Security Contracts

IPC:

- allowlisted methods only
- schema validation on every request
- request size limit
- attachment path validation
- no arbitrary shell command method
- no arbitrary file read/write method

Files:

- file access must be project-scoped or explicitly selected by the user
- writes happen only after Go approval policy passes
- diff projection never grants write access to renderer

Routes:

- downstream CLI execution happens only in Go
- provider calls happen only in Go
- local model runtime launch happens only in Go
- route setup output is redacted

Diagnostics:

- preview before export
- approval before export
- no raw secrets
- no raw prompts by default
- no raw artifact bodies by default
- full local paths redacted unless user opts in

## Update Contract

The app bundle includes:

- Tauri shell version
- Go sidecar version
- protocol version
- minimum compatible session schema
- update channel

Update rules:

- updater artifacts are signed
- app refuses unsigned update payloads
- protocol mismatch reports app/core update need
- session schema migration must be reversible or guarded by compatibility check
- rollback preserves local session state

## Accessibility Contract

The renderer must expose:

- semantic landmarks for sidebar, thread, inspector, composer
- keyboard focus restoration after sheets
- labels for route status and offline state
- labels for diff additions/deletions
- readable approval risk text
- reduced-motion compliance

Keyboard tests must cover:

- Command-N
- Command-O
- Command-P
- Command-F
- Command-1, Command-2, Command-3
- Command-Return for focused safe approval
- Escape to close sheet or stop focus trap

## Testing Contract

Unit tests:

- protocol envelope validation
- method dispatch
- project projection
- session projection
- simple answer returns compact turn with no session creation
- direct file edit returns receipt and diff
- ambiguous file edit returns bounded ask
- route status distinguishes CLI handoff from provider API
- Gatekeeper-rejected CLI returns `blocked_by_trust`
- approval id is single-use
- diagnostics export redacts required fields

Integration tests:

- app handshake returns task prompt snapshot
- open project does not create session
- submit simple question returns `Paris.`
- submit direct file edit changes file and returns diff
- route `codex` unavailable state does not fallback to provider API
- offline resume opens existing artifact and records sync debt
- sidecar crash preserves unsent input in renderer test harness

Manual release checks:

- signed and notarized clean-machine launch
- Gatekeeper launch
- updater signature verification
- keyboard-only full walkthrough
- VoiceOver smoke
- reduced-motion smoke
- diagnostics bundle inspection

## Error Codes

Required codes:

- `protocol_mismatch`
- `invalid_request`
- `project_not_selected`
- `project_permission_denied`
- `session_not_found`
- `simple_answer_completed`
- `ambiguous_file_target`
- `approval_required`
- `approval_denied`
- `approval_expired`
- `route_unavailable`
- `route_blocked_by_trust`
- `route_blocked_by_offline`
- `sidecar_interrupted`
- `diagnostics_redaction_failed`
- `update_signature_invalid`

Error copy rule:

- user-facing `message` is short
- `recovery` names the next action
- logs may include structured error class but not secrets

## Acceptance Gates

Implementation is not ready until:

- HLD and LLD are protected by PRD drift gate.
- Product tests pin HLD/LLD non-negotiables.
- App protocol tests pass.
- Golden app transcripts match CLI first-minute behavior.
- Security checklist passes for Tauri capabilities and Go sidecar boundary.
- Signing/notarization/Gatekeeper smoke passes before public alpha.
