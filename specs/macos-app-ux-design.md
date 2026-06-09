# macOS App UX Design

Updated: 2026-06-09

This is the active UX design for the Jini macOS app. It implements
[macos-app-prd.md](./macos-app-prd.md) and stays subordinate to
[number-one-platform-prd.md](./number-one-platform-prd.md) and
[product-settling-decisions.md](./product-settling-decisions.md).

## Design Goal

Make the macOS app feel immediately familiar to Codex and Claude Code users
while making Jini's differentiator visible: one frugal, route-aware session
layer across local models, downstream CLIs, provider routes, and offline work.

The app should feel like a calm desktop command center, not a dashboard, chat
toy, project manager, or Jini-specific workflow shell.

The UX must preserve the same session model as the CLI.

## Selected Approach

Use a three-pane desktop supervision shell:

- left sidebar for projects, sessions, and search
- center thread for compact task input, answers, receipts, and active work
- right inspector for progress, diffs, artifacts, route evidence, approvals,
  diagnostics, and sources

Rejected alternatives:

- Dashboard-first: too slow to first useful output and repeats the old
  saved-work startup mistake.
- Document-editor-first: strong for artifacts but weak for route handoff,
  approvals, and coding supervision.
- Terminal-only wrapper: familiar, but fails the desktop app purpose because
  diffs, artifacts, approvals, and route evidence stay too hidden.

## First-Minute Contract

The first minute is the release bar. A user who knows Codex should not need to
learn Jini concepts before value.

Default launch:

```text
Jini
Describe the task.
>
```

Simple question:

```text
> what is the capital of france?
Paris.
```

Direct edit:

```text
> add "jini was here" to pear fellow script.txt
Edited pear fellow script.txt
```

Coding task:

```text
> review this repo for release blockers
Plan
- inspect gates
- review changed files
- run focused tests

Approval needed: run git diff --stat
```

Forbidden first-minute states:

- Start/Keep modal
- Switch control
- saved-work dashboard
- Task Snapshot for factual questions
- Working Draft for clear file edits
- route ceremony before a route matters
- agent role list before useful output

## Layout

Default window:

```text
+--------------------------------------------------------------------------------+
| Toolbar: Project   Search   Route Auto   Offline/Online   Model   Approvals    |
+----------------------+--------------------------------+------------------------+
| Sidebar              | Thread                         | Inspector              |
|                      |                                |                        |
| Projects             | Jini                           | Progress               |
| - current repo       | Describe the task.             | Diffs                  |
| - recent repo        | >                              | Artifacts              |
|                      |                                | Route Evidence         |
| Sessions             | Compact answers, plans,        | Approvals              |
| - release audit      | receipts, asks, and running    | Diagnostics            |
| - route setup        | command output stream here.    | Sources                |
|                      |                                |                        |
+----------------------+--------------------------------+------------------------+
| Composer: attach, paste, dictate, route profile, approval mode, send            |
+--------------------------------------------------------------------------------+
```

Responsive density:

- Wide: three panes visible.
- Medium: inspector collapses into a segmented right drawer.
- Narrow: sidebar and inspector are overlays; thread stays primary.
- Full screen: preserve pane ratios and keyboard shortcuts.

## Navigation Model

Primary navigation:

- Projects: local folders and repos the user has opened.
- Sessions: durable work threads scoped to the selected project.
- Search: sessions, artifacts, files, and route receipts.

Secondary navigation:

- Artifacts: generated or edited outputs.
- Diffs: file changes by task turn.
- Routes: CLI handoff, provider, gateway, and local/offline route status.
- Diagnostics: install, trust, model runtime, sync, and support bundle state.

Navigation rules:

- Opening the app does not auto-select a saved work thread unless the user
  launched from a specific notification, file, deep link, or CLI handoff.
- Recent sessions are visible but passive.
- Selecting a session resumes it directly; no Start/Keep decision.
- Search results open in context and keep the current project visible.

## Core Surfaces

### Toolbar

Required controls:

- Project selector
- Search
- route mode: auto, selected CLI, provider, local, offline
- power posture: normal, low battery, plugged in when available
- model/profile selector
- approval mode
- sync/offline indicator

Toolbar rules:

- Show route and power state as compact labels, not banners.
- Escalate to banners only when work is blocked.
- Never expose secrets, raw provider keys, or full local paths in toolbar text.

### Sidebar

Sections:

- Projects
- Sessions
- Pinned artifacts
- Setup issues

Sidebar rules:

- Keep labels user-facing: project name, session title, latest meaningful
  receipt.
- Do not show internal work ids by default.
- Do not show current work on launch as a blocking decision.
- Setup issues appear only when they affect the selected project or route.

### Thread

The thread is answer-first and receipt-first.

Message types:

- compact answer
- action receipt
- plan
- approval request
- command output summary
- bounded clarification ask
- route handoff receipt
- error with recovery

Thread rules:

- Simple answers stay one line when possible.
- Plans show only when the task requires multiple actions.
- Running work streams progress only after the user has seen the plan or first
  action.
- Long command output is summarized in the thread and expandable in inspector.
- A clarification ask must name the exact ambiguity and candidate resolution.

### Composer

Controls:

- text input
- attachment button
- paste/import
- microphone or dictation hook when platform-ready
- route/profile popover
- approval mode popover
- send/stop

Composer rules:

- Enter sends; Shift-Enter inserts a newline.
- Slash commands may exist only as familiar power shortcuts and must never be
  required for first value.
- Placeholder stays direct: `Describe the task.`
- Disabled states explain the reason in one short line.

### Inspector

Tabs:

- Progress
- Diffs
- Artifacts
- Route
- Approvals
- Diagnostics
- Sources

Inspector rules:

- Inspector is evidence, not ceremony.
- Default tab follows task type: Diffs for file edits, Route for route setup,
  Progress for running work, Artifacts for generated outputs.
- Every tab has an empty state that teaches one action, not a product tour.

## Key User Flows

### First Launch

1. User opens Jini.
2. App shows task prompt and optional `Open Project` affordance.
3. If no project is selected and the task requires files, ask for project
   selection.
4. If the task is a simple question, answer without forcing project selection.

Pass condition: `what is the capital of france?` returns `Paris.` with no
project, artifact, or saved-work frame.

### Open Project

1. User selects a local folder.
2. App shows project name in toolbar and recent sessions in sidebar.
3. Center remains on the task prompt.
4. Route inspector shows configured CLI/local/provider health on demand.

Pass condition: opening a project does not create a session.

### Direct File Edit

1. User asks for a clear file edit.
2. Jini resolves file in project.
3. If safe, app asks for approval only when policy requires it.
4. Jini applies edit through Go core.
5. Thread shows changed-file receipt.
6. Inspector opens Diffs.

Pass condition: no Working Draft, no generic artifact, no unnecessary plan.

### Ambiguous File Edit

1. User asks for an edit but multiple files match.
2. Thread asks one exact question with candidates.
3. User selects candidate or types filename.
4. Jini performs edit and shows receipt.

Pass condition: ambiguity is resolved in one turn when possible.

### Coding Task With Commands

1. User gives a repo task.
2. Jini presents a short plan.
3. App requests approval before shell commands, file writes, commit, push, or
   network escalation based on policy.
4. Progress streams as receipts, not raw logs.
5. Diffs, tests, blockers, and route evidence stay inspectable.

Pass condition: user can answer "what changed and what ran" without reading a
transcript.

### Route Setup

1. User opens route controls or types a route task.
2. Route inspector lists configured routes by category:
   CLI handoff, provider API, gateway, local/offline.
3. Missing routes show exact setup command or env var without secrets.
4. Gatekeeper-rejected CLIs fail closed and point to `jini doctor`.

Pass condition: route named `codex` never silently becomes provider API.

### Offline Resume

1. Device is offline or user chooses offline route.
2. Toolbar shows offline state.
3. Existing sessions and artifacts remain openable.
4. Available local routes are listed.
5. Remote routes are marked unavailable.
6. New local events append sync debt.

Pass condition: offline work does not fork the session.

### Diagnostics Export

1. User opens Diagnostics.
2. App previews what will be included.
3. Secrets, prompts, artifact bodies, full paths, and provider payloads are
   excluded by default.
4. User exports bundle.

Pass condition: support can debug route/install/sync issues without private
content leakage.

## Visual Direction

Personality:

- precise
- calm
- native
- low-noise
- evidence-oriented

Visual rules:

- Prefer native macOS spacing, sidebars, toolbars, sheets, and popovers.
- Use a neutral base with one Jini accent for route/action state.
- Diffs use conventional green/red semantics with accessible contrast.
- Route states use labels plus color; color is never the only signal.
- Avoid gradients, decorative illustrations, and marketing panels inside the
  work surface.
- Avoid purple-on-white defaults and generic AI glow treatments.

Typography:

- Use platform-native text sizing unless the HLD chooses a custom design
  system with measurable readability benefits.
- Monospace only for commands, paths, diffs, logs, and code.
- Thread text should optimize for scanning: short paragraphs, receipts, and
  small plans.

Motion:

- Use motion for state changes: inspector open, approval sheet, command
  progress, route switch.
- Respect reduced-motion settings.
- Do not animate token streaming in ways that slow scanning.

## Approval UX

Approval request anatomy:

- action
- target
- route
- risk
- exact command or file list when applicable
- allow once
- always allow for this project or route when policy permits
- deny

Approval sheet example:

```text
Approve command?

Command
go test ./internal/app -run TestRouteHelp

Why
Verify route setup help regression.

Route
local shell

Choices
Allow once    Always allow tests in this project    Deny
```

Approval rules:

- Destructive operations default to deny unless the user explicitly approves.
- Commit and push approvals are separate.
- Network escalation names destination route or provider.
- Long-running local model work shows battery/power posture before start.

## Empty And Error States

Empty states:

- No project: `Open a project or ask a general question.`
- No sessions: `No saved sessions yet. Start with a task.`
- No diffs: `No file changes in this session.`
- No route issues: `Configured routes are ready or not selected.`

Error states:

- Missing CLI: name executable and setup path.
- Gatekeeper rejected CLI: do not run; explain trust check and recovery.
- Offline blocked route: name unavailable route and local alternatives.
- Ambiguous file: list candidates.
- Permission denied: show target, operation, and recovery.

## Keyboard And Menu Contract

Required shortcuts:

- Command-N: new task
- Command-O: open project
- Command-P: command palette
- Command-F: search current project/session
- Command-Shift-F: global search
- Command-1: focus sidebar
- Command-2: focus thread
- Command-3: focus inspector
- Command-Return: approve focused safe action when allowed
- Escape: close sheet, popover, or stop focus trap

Menu groups:

- File: New Task, Open Project, Open Artifact, Export Diagnostics
- Edit: Copy, Paste, Find
- View: Sidebar, Inspector, Route, Diffs, Artifacts
- Routes: Auto, Offline, Local Model, CLI Handoff, Provider, Route Help
- Help: Doctor, Logs, Privacy, Keyboard Shortcuts

## Accessibility

Requirements:

- Full keyboard navigation across sidebar, thread, inspector, composer, and
  approval sheets.
- VoiceOver labels for route state, approval risk, diff additions/deletions,
  offline state, and sync debt.
- Minimum contrast meets platform accessibility expectations.
- Reduced motion honored.
- Text scales without clipping at larger accessibility sizes.
- Focus order follows visual order.
- Inspector tabs expose selected state and unread changes.

## Copy Rules

Use compact, literal copy.

Prefer:

- `Edited pear fellow script.txt`
- `Approval needed: run tests`
- `Codex CLI is blocked by macOS trust checks`
- `Offline: local work available, sync paused`

Avoid:

- `Result ready`
- `Task Snapshot`
- `Your first draft is ready`
- `Working with`
- `Let's get started`
- `Unlock your productivity`

## Implementation Trace

| UX Area | PRD Requirement |
| --- | --- |
| first-minute prompt | MAC-P0-1 |
| same session graph | MAC-P0-2, MAC-P0-12 |
| project/session sidebar | MAC-P0-3 |
| compact factual answers | MAC-P0-4 |
| direct file edit and ambiguity | MAC-P0-5 |
| diff and artifact inspector | MAC-P0-6 |
| approval sheets | MAC-P0-7 |
| route inspector | MAC-P0-8 |
| offline resume | MAC-P0-9 |
| diagnostics export | MAC-P0-11 |
| native shortcuts and menu | MAC-P1-7 |

## UX Gates

Before implementation:

- HLD maps every UX surface to Go core responsibilities and IPC boundaries.
- LLD defines view models for projects, sessions, thread turns, approvals,
  diffs, route evidence, offline debt, and diagnostics.
- Prototype validates the three-pane layout at wide, medium, and narrow widths.

Before alpha:

- Golden transcripts pass for simple answer, direct edit, ambiguous edit,
  coding task approval, route setup, and offline resume.
- Keyboard-only walkthrough completes first launch, open project, direct edit,
  diff review, route help, approval, and diagnostics export.
- VoiceOver smoke covers toolbar state, thread messages, inspector tabs,
  approvals, and diffs.
- Visual QA rejects Start/Keep, Switch, Task Snapshot, Working Draft, and
  dashboard-first launch.

Before release:

- Codex desktop-caliber scorecard passes.
- Clean-machine install and first launch pass.
- Gatekeeper-rejected downstream CLI state is visible and safe.
- Token-frugal route evidence is inspectable without cluttering simple answers.
- Public docs and screenshots match the shipped app.
