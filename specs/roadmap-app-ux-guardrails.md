# Roadmap App UX Guardrails (GUI Surfaces)

Updated: 2026-07-12
Status: ROADMAP-STAGE. These guardrails apply to Jini's post-v1 GUI surfaces
(macOS desktop first, mobile continuation surface later) — NOT the v1 terminal
CLI. They activate only when the roadmap item activates and a decision-record
update brings that surface into scope (see
[number-one-platform-prd.md](./number-one-platform-prd.md) Roadmap and Tier
Boundary).

The v1 terminal product carries the CLI-idiomatic translation of these same
intents in the PRD's "Terminal experience bar" (UX Contract section). This
document is the literal, native-GUI form for when there is a screen to render.

Precedence: nothing here overrides the PRD, the free/commercial boundary, or
the honesty rules. A premium surface that renders an unlabeled or inflated
savings number is a defect (§4.5, §5), the same as in the CLI.

## 1. Editor's-Choice design guardrails

- **Platform-idiomatic, never a generic cross-platform port.** iOS/macOS: native
  Swift surfaces — Live Activities and Dynamic Island for long-running or
  background agent tasks (a multi-file refactor, a running route, a savings
  milestone), fluid Home Screen / menu-bar widgets. Android: Material You
  dynamic color, rich contextual glanceable widgets. The desktop app follows
  the macOS HLD/LLD's Go-core + native-shell boundary — the GUI is a
  presentation surface over the same session model, never a second product.
- **Micro-interactions & haptics.** Custom haptic profiles (CoreHaptics /
  Android HapticFeedback) for critical developer moments: a crisp snap when a
  task completes and the receipt lands, a subtle rumble on a throttle/failure
  alert, physics-based transitions on the session timeline. Motion is
  purposeful and interruptible; respect reduce-motion.
- **Frictionless time-to-value.** Premium anonymous onboarding: the user
  reaches the core session/timeline layout within three taps, no account and
  no tier selection forced upfront. An account exists only for Continuity sync
  (paid), consistent with the v1 anonymous-by-default rule.

## 2. Refreshing-UX system patterns

- **Adaptive state engine — morph by work state** (the GUI analog of the CLI's
  state-adaptive surface): Idle/planning (compact intake, recent sessions,
  running savings counter), Active-task (high-density progress: route status,
  live token/cost, streamed diffs, approval prompts), Completed/review (clean
  receipt, savings summary, diff review, share affordance). State transitions
  are animated and legible.
- **Elegant asymmetry & editorial typography.** No enterprise card-grid
  clutter: high-contrast type, purposeful whitespace, a magazine-style
  editorial layout for the savings dashboard and session summaries. This is
  the surface where the CLI's `jini savings --report` visual ambition becomes
  the primary experience.
- **Contextual skeletoning.** While the agent works or context loads, use
  polished custom shimmer skeletons that hint at incoming content (files about
  to change, receipt about to render) — never a bare spinner.

## 3. Deliverable rule for GUI build prompts

Every GUI build prompt (Lovable/SwiftUI/Compose or equivalent) must explicitly
specify, per feature: exact layout, the micro-interactions and haptics,
accessibility (WCAG AA, Dynamic Type / text scaling, VoiceOver/TalkBack labels,
reduce-motion, color-independent meaning), and the platform-native components
that make that feature feel high-end. No pseudo-code, no placeholder UI.

## Honesty and boundary carry-over (non-negotiable)

- Savings shown in any denomination carry the literal/imputed label at every
  size, including glances, widgets, and Live Activities (§4.5).
- Paid features (Autopilot, Continuity) fail closed without entitlement; a free
  equivalent path exists; no pricing/SKU data ships in a public client (§4.9).
- The GUI is a surface over the Go core's session model; it renders and
  requests, it does not own product logic or perform side effects directly
  (macOS HLD/LLD boundary).
