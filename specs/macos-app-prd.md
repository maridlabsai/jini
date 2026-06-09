# macOS App PRD

Updated: 2026-06-09

This is a specialized PRD for the Jini macOS app. It is subordinate to
[number-one-platform-prd.md](./number-one-platform-prd.md) and
[product-settling-decisions.md](./product-settling-decisions.md).

The macOS app is active P1 planning work. It does not make Windows, mobile,
commercial agent suites, or broad company automation active scope.

## Product Goal

Build a macOS app at the Codex desktop-caliber UX bar while preserving Jini's
settled identity: a CLI-first AI work router and durable session layer.

The app is not a new product, not a dashboard-first workflow, and not a new
conversation style. It is the richer desktop surface for the same session model as the CLI:
task in, useful result out, route evidence available, side effects approved,
artifacts and diffs easy to inspect.

## Research Basis

Current benchmark inputs:

- OpenAI describes the Codex app as a desktop command center for agents with
  project/thread continuity, parallel work, diff review, worktrees, skills, and
  session history shared with CLI and IDE surfaces:
  https://openai.com/index/introducing-the-codex-app/
- OpenAI's Codex CLI docs anchor the familiar terminal-agent expectation:
  local execution in a selected directory, file edits, command execution,
  slash commands, approval modes, subagents, web search, and code review:
  https://developers.openai.com/codex/cli
- Apple requires direct-distribution macOS apps to use Developer ID signing,
  hardened runtime, notarization, launch testing under Gatekeeper, and a
  verified distribution path:
  https://help.apple.com/xcode/mac/current/en.lproj/dev033e997ca.html
- Tauri 2 provides the current repo's default desktop shell path, but its
  capability model and updater signing requirements must be treated as hard
  security gates, not defaults to accept blindly:
  https://v2.tauri.app/security/capabilities/
  https://v2.tauri.app/plugin/updater/

## Tenets

- Task first, not dashboard first.
- Same session model as the CLI.
- Claude/Codex familiarity beats Jini-specific workflow invention.
- Offline is a route state, not a separate product.
- Token frugality is visible when useful and silent when not useful.
- Route evidence, file diffs, approvals, and receipts are inspectable.
- Native trust is product quality: signed, notarized, sandbox-reviewed,
  hardened runtime enabled, secrets protected, diagnostics redacted.
- Focused implementation wins: no surface expands without a PRD, HLD, LLD, and
  gate update.

## P0 Requirements

P0 is the minimum credible internal dogfood bar.

| ID | Requirement | Acceptance Bar |
| --- | --- | --- |
| MAC-P0-1 | Launch into a familiar task prompt | First screen supports opening a project and typing a task without Start/Keep, Switch, or saved-work ceremony. |
| MAC-P0-2 | Reuse the CLI session graph | The app reads and writes the same session id, artifact ids, route evidence, approval state, and offline debt as the Go CLI core. |
| MAC-P0-3 | Project and thread continuity | Users can select a local project, see recent sessions for that project, and resume one without replaying full transcripts. |
| MAC-P0-4 | Compact direct answers | Simple factual questions answer directly in the thread and do not create saved work, snapshots, or artifact shells. |
| MAC-P0-5 | Direct local file work | Clear local file-edit requests produce file changes, diff preview, and receipt; ambiguous edits fail closed with exact candidates. |
| MAC-P0-6 | Diff and artifact review | Users can inspect changed files, rendered artifacts, generated notes, and route receipts in one right-side review surface. |
| MAC-P0-7 | Side-effect approvals | File writes, shell commands, external sends, commits, pushes, and network escalation require visible approval policy and receipt. |
| MAC-P0-8 | Route control | Users can see current route, why it was chosen, offline/online state, token posture, configured CLI health, and setup guidance. |
| MAC-P0-9 | Offline continuation | Existing sessions and local artifacts open offline; local model routes can continue when configured; sync debt is visible. |
| MAC-P0-10 | Secure local trust | Direct builds must be Developer ID signed, notarized, stapled, hardened-runtime enabled, sandbox-entitlement reviewed, and Gatekeeper-smoked. |
| MAC-P0-11 | Redacted diagnostics | Support bundles include app/core versions, session ids, route ids, sync state, command classes, and redacted logs; no secrets or raw prompts by default. |
| MAC-P0-12 | No duplicate state | The app must not keep a private session store that can diverge from the CLI session graph. |

## P1 Requirements

P1 is the public alpha bar.

| ID | Requirement | Acceptance Bar |
| --- | --- | --- |
| MAC-P1-1 | Local-to-CLI handoff | A session can move between macOS app and `jini` CLI with the same current artifact and route evidence. |
| MAC-P1-2 | Downstream CLI bridge | Configured Codex, Claude Code, Gemini CLI, Aider, and OpenCode routes show available, missing, rejected, and failed handoff states. |
| MAC-P1-3 | Worktree-safe coding tasks | Coding tasks can isolate changes from the main working tree or clearly explain why isolation is unavailable. |
| MAC-P1-4 | Commit and PR receipt | Coding work can present files changed, tests run, blockers, commit readiness, and push/PR handoff state. |
| MAC-P1-5 | Battery and power posture | Route choices account for low battery, plugged-in mode, thermal pressure when available, and local model load cost. |
| MAC-P1-6 | Token savings ledger | Users can inspect why context was reused, compressed, skipped, or escalated without seeing route ceremony on every answer. |
| MAC-P1-7 | Native productivity | Keyboard shortcuts, command palette, notifications, menu commands, and file-open affordances follow macOS expectations. |
| MAC-P1-8 | Update and rollback | Signed updates verify signatures, preserve rollback, and block incompatible session schema upgrades. |

## P2 Requirements

P2 remains gated by future decision-record updates.

| ID | Requirement | Acceptance Bar |
| --- | --- | --- |
| MAC-P2-1 | Commercial managed skills | Commercial users can manage governed skills without exposing free-tier agent-role theater. |
| MAC-P2-2 | Managed multi-agent work | Commercial users can supervise parallel developer/tester agents with policy, audit, and worktree isolation. |
| MAC-P2-3 | Team policy and audit | Commercial teams can govern approvals, route policy, secrets, diagnostics, and retention. |
| MAC-P2-4 | Automations | Background work and scheduled follow-up exist only behind explicit commercial policy and receipts. |

## Non-Goals

- No Start/Keep model.
- No visible Switch startup control.
- No saved-work dashboard on bare launch.
- No generic drafting shell for simple questions or obvious file edits.
- No new slash-command grammar required before value.
- No free-tier skills-based OS productivity suite.
- No public agent tree, role theater, or orchestration log as the default UX.
- No separate macOS-only session model.
- No Mac App Store commitment before direct distribution is signed,
  notarized, and trusted.

## UX Contract

The first minute should feel familiar to a Codex or Claude Code user.

Default shell:

```text
Jini
Describe the task.
>
```

Simple answer:

```text
> what is the capital of france?
Paris.
```

Clear file edit:

```text
> add "jini was here" to pear fellow script.txt
Edited pear fellow script.txt
```

Ambiguous file edit:

```text
> add "jini was here" to the script txt file
I found multiple matching .txt files:
- pear fellow script.txt
- pear vc script.txt

Which one should I edit?
```

Coding task:

```text
> add tests for route setup help
Plan
- update route help fixture
- add regression test
- run focused test

Approval needed: run go test ./internal/app -run TestRouteHelp
```

The app may show richer panels, but the transcript remains compact. Panels
should explain work state, not replace useful output.

## App Surfaces

- Left rail: projects, recent sessions, explicit search.
- Center: lightweight task thread with answer or action receipt first.
- Right inspector: progress, diffs, artifacts, route evidence, approvals,
  diagnostics, and sources when relevant.
- Bottom or command palette: route, model/profile, approval mode, offline mode,
  and power posture controls.

Every surface must answer: what changed, what ran, what route acted, what is
ready, what needs approval, and how to resume from CLI.

## Technical Direction

Default implementation direction remains a Tauri 2 shell over the Go core.

The HLD must compare Tauri 2 with native SwiftUI/AppKit before implementation.
Tauri wins only if it preserves native trust, accessibility, performance,
secure IPC, updater signing, and a small entitlement set. Native SwiftUI/AppKit
wins if the Codex-caliber desktop bar requires deeper macOS integration than a
webview shell can safely provide.

Non-negotiable architecture constraints:

- Go core owns intent, routing, session graph, artifact metadata, approvals,
  offline debt, diagnostics, and gates.
- macOS UI owns presentation, local project selection, native menus,
  notifications, file-open affordances, and approval surfaces.
- IPC is allowlisted per window and per command.
- No prompt, artifact body, provider key, access token, or local file path is
  logged without explicit redaction policy.
- The app must support deterministic golden transcript fixtures before public
  alpha.

## Metrics

- install success rate
- first launch success rate
- time to first useful answer
- time to resume latest session
- direct file-edit success rate
- ambiguous-edit fail-closed rate
- route handoff success rate
- token savings shown versus context replay avoided
- offline session open success rate
- diff open latency
- crash-free sessions
- Gatekeeper launch success on clean machines
- diagnostics export success without secret leakage

## Release Gates

No macOS release ships unless these gates pass:

- PRD/HLD/LLD trace covers every P0 requirement.
- Golden transcripts cover simple question, direct file edit, ambiguous file
  edit, coding task with approval, route handoff, offline resume, and compact
  saved-work behavior.
- App smoke launches from a signed, notarized, stapled build on a clean
  Gatekeeper-enabled Mac.
- Security review covers sandbox entitlements, hardened runtime exceptions,
  IPC allowlist, updater signatures, secret storage, diagnostics redaction, and
  local file access prompts.
- Session compatibility tests prove CLI-to-app and app-to-CLI continuity.
- Offline tests prove local open, local append, route evidence, and sync debt.
- Performance tests record cold start, warm resume, artifact open, diff open,
  memory, and local route readiness.
- Accessibility review covers keyboard-only use, VoiceOver labels, focus order,
  contrast, and reduced-motion behavior.
- Competitor scorecard includes Codex desktop parity checks and deletion
  checks for Jini-only ceremony.

## Roadmap

Phase 0, design:

- approve this PRD
- write macOS UX design
- write macOS HLD and LLD
- choose Tauri 2 versus SwiftUI/AppKit with a prototype-backed decision
- add macOS PRD trace and golden transcript fixtures

Phase 1, internal dogfood:

- launch app shell
- open local project
- show sessions from Go session graph
- submit compact task
- show artifacts, diffs, route evidence, and approvals

Phase 2, alpha:

- direct file edits with diff review
- downstream CLI handoff health
- offline session open and sync-debt surface
- signed, notarized direct installer
- redacted diagnostics bundle

Phase 3, beta:

- worktree-safe coding tasks
- commit/push/PR handoff receipts
- token savings ledger
- battery-aware route policy
- signed updates and rollback

Phase 4, release:

- Codex desktop-caliber scorecard green
- competitor regression transcripts green
- clean-machine install and Gatekeeper launch green
- public docs match shipped behavior

## Open Decisions

- Tauri 2 versus SwiftUI/AppKit final implementation stack.
- Direct-only distribution versus future Mac App Store path.
- Exact local model runtime packaging and update policy.
- Commercial skills and multi-agent UI timing after free macOS alpha proves the
  core desktop surface.
