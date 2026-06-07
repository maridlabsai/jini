# App Platform Shipping Playbook

Updated: 2026-06-06

This document compresses current platform research into the shipping rules for
Jini's web, desktop, and mobile apps.

It is a specialized implementation playbook, not the top-precedence product and
operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this playbook conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, automation posture, app-shipping order, or route
policy, the canonical PRD wins and this playbook should be updated.

Read alongside:

- [platform-offline-strategy.md](./platform-offline-strategy.md)
- [cross-surface-session-platform-prd.md](./cross-surface-session-platform-prd.md)
- [cross-surface-session-system-and-dev-design.md](./cross-surface-session-system-and-dev-design.md)
- [engineering-gate-matrix.md](./engineering-gate-matrix.md)
- [local-model-support-matrix.md](./local-model-support-matrix.md)

## Product Rule

Jini apps should be thin, observable, secure platform views over one durable
session graph.

The costly mistake to avoid is treating each app as a separate product with its
own state, update loop, logging shape, and security model.

## Default Stack Decision

Use the smallest stack that preserves the shared session graph and gives the
platform enough native control.

| Surface | Default | Use when | Avoid when |
| --- | --- | --- | --- |
| Web | Next.js App Router, TypeScript, server-first rendering, OpenTelemetry instrumentation | marketing, docs, account, admin, web continuation, public artifact review | the surface must run fully offline without a browser runtime |
| macOS desktop | Tauri 2 shell over the Go core, with native signing, notarization, and sandbox review | we need a small trusted desktop shell, local files, local models, and cross-platform reuse | deep Apple-only UI or heavy AppKit/SwiftUI integration is the product differentiator |
| Windows desktop | Tauri 2 shell over the Go core, packaged with MSIX when package identity, clean updates, or enterprise deployment matter | desktop continuity, local artifact store, local model supervision, enterprise install | the app requires heavy Windows-only UI primitives; then use Windows App SDK and WinUI |
| iOS | Expo React Native for continuity/review, with native modules only where secure storage, local model runtime, or platform permission boundaries require them | quick review, approval, annotate, hand off, resume | deep native inference or complex local artifact editing becomes the core product |
| Android | Expo React Native for continuity/review, with Kotlin modules for secure storage, local model runtime, Play Integrity, and performance-critical paths | same mobile surface as iOS plus Android-specific attestation and Baseline Profiles | Android becomes the main authoring host; that is a desktop job |

Framework exceptions must be explicit:

- Choose native SwiftUI or Jetpack Compose when platform-specific security,
  model runtime, accessibility, or OS integration is more important than shared
  UI velocity.
- Choose Flutter only if a single compiled mobile UI matters more than sharing
  TypeScript app logic with the web surface.
- Choose Electron only when the product needs a mature Node or Chromium
  extension ecosystem that Tauri cannot provide. If Electron is used, the
  Electron security checklist is a hard gate, not guidance.

## Costly Mistakes To Avoid

- Do not ship separate session stores for CLI, desktop, web, and mobile.
- Do not make mobile pretend to be desktop inference parity.
- Do not ship unsigned or unverifiable auto-updates.
- Do not embed client secrets and call them protected.
- Do not log prompts, artifacts, tokens, API keys, local file paths, or model
  payloads without explicit redaction.
- Do not accept webview defaults. Every webview needs a permission, navigation,
  external-link, and IPC policy.
- Do not let app-store privacy declarations, Android permissions, Windows
  package identity, macOS entitlements, or notarization become a release-week
  chore.
- Do not use debug build size or debug performance as a proxy for production.
- Do not rely only on lab performance. Field telemetry must validate real
  devices, network conditions, and user journeys.
- Do not let each platform invent a different log vocabulary. Diagnostics must
  correlate through one session id, route id, artifact id, device class, app
  version, runtime version, and model profile.

## Security Baseline

Security is a product feature for Jini because the product handles work
context, local files, route evidence, and model payloads.

Minimum cross-platform requirements:

- Use OWASP MASVS as the mobile security verification baseline.
- Store secrets only in platform-provided secure storage.
- Keep API keys and provider credentials out of mobile binaries, desktop
  bundles, update manifests, logs, and crash reports.
- Encrypt local session projections and artifact caches when the platform
  provides a durable key store.
- Require explicit user intent before connector writes, publication, send, or
  irreversible local file mutation.
- Separate user-visible route evidence from private diagnostic payloads.
- Redact by default and make raw diagnostic upload opt-in.
- Treat screenshots, PDFs, audio transcripts, prompts, and local file paths as
  sensitive data.

Platform-specific gates:

- macOS: sign, notarize, staple, review hardened runtime, review sandbox
  entitlements, and keep the entitlement set explainable.
- Windows: prefer MSIX package identity when notifications, background tasks,
  clean uninstall, Store, enterprise deployment, or Windows AI APIs matter.
- Android: enforce release signing, Play Integrity for abuse-sensitive actions,
  least-privilege permissions, non-exported components by default, and secure
  network configuration.
- iOS: enforce App Store privacy declarations, permission minimization, keychain
  storage, privacy manifest review, and no background capability without a
  product reason.
- Web: ship strict Content Security Policy, secure cookies, server-side auth
  checks, dependency scanning, and no secret-bearing client environment values.
- Tauri: define capabilities per window and platform; do not expose broad
  commands to web content.
- Electron exception path: no Node integration for remote content, context
  isolation on, process sandboxing on, restrictive CSP, sender validation for
  IPC, and no untrusted `shell.openExternal`.

## Performance And Optimization Baseline

Jini should optimize for time to useful session state, not just generic app
startup.

Required measures:

- Web: track Core Web Vitals in the field. The web surface should meet the
  current web.dev thresholds at p75: LCP at or below 2.5s, INP at or below
  200ms, CLS at or below 0.1.
- Android: ship Baseline Profiles and startup profiles for app launch, session
  resume, artifact open, approval, and handoff flows.
- iOS: collect production app size, cold start, memory, battery, and crash
  evidence from TestFlight and release builds, not simulator-only runs.
- Desktop: measure cold start, warm resume, local model route readiness,
  artifact open latency, memory, update size, and install success.
- Flutter exception path: measure release bundle size and use DevTools size
  analysis; debug builds are not representative.
- Local model flows: measure model load time, first-token time, tokens per
  second, memory, battery or thermal pressure, crash rate, and route-regret
  rate by device class.

Optimization policy:

- Optimize the critical path first: session list, current session resume,
  latest artifact, missing items, next action, and route evidence.
- Defer heavy model/runtime initialization until a route can actually use it,
  unless prewarming is measured to improve first useful result without hurting
  battery or startup.
- Cache bounded session projections on mobile; cache richer artifact state on
  desktop.
- Keep mobile payloads small; do not download full desktop context just to show
  a review screen.
- Prefer incremental updates and signed over-the-air updates only when rollback
  and runtime compatibility are enforced.

## Logging, Diagnostics, And Observability

Use one observability vocabulary across all platforms.

Minimum emitted fields:

- `session_id`
- `surface`
- `platform`
- `app_version`
- `runtime_version`
- `route_id`
- `model_profile`
- `provider_family`
- `artifact_id`
- `operation`
- `offline_state`
- `sync_state`
- `result`
- `duration_ms`
- `error_class`
- `privacy_redaction_state`

Operating rules:

- Use OpenTelemetry concepts for traces, metrics, and logs, while leaving the
  backend vendor replaceable.
- Correlate client events with server or CLI events through `session_id` and
  `route_id`.
- Keep platform-native quality feeds in the loop: Android vitals, App Store and
  TestFlight crash/performance reports, Windows package/install telemetry, and
  web real-user monitoring.
- Emit local-only diagnostics when the user is offline, then sync redacted
  summaries after reconnection.
- Provide a local diagnostics bundle export that includes versions, route
  evidence, sync state, and redacted logs, never raw secrets or full prompts by
  default.
- Separate operational logs from product learning signals. A crash stack is not
  an implicit permission to train or tune behavior.

## Update And Release Policy

Every shipped app must support safe rollback.

Required release fields:

- app version
- build number
- platform
- release channel
- native runtime version
- web bundle version
- local model registry version
- schema version
- minimum compatible session schema
- rollout cohort
- rollback target

Rules:

- Signed updates are mandatory for mobile and desktop update paths.
- Expo EAS Update may be used only with end-to-end code signing and explicit
  runtime versioning.
- Desktop auto-update must verify signatures and preserve rollback.
- Windows app updates should prefer MSIX or an equally inspectable signed update
  path.
- macOS direct distribution must pass signing, notarization, and stapling before
  release.
- Schema migrations must be reversible or guarded by an explicit compatibility
  check.
- Model registry updates must follow the offline model promotion loop before
  becoming defaults.

## Development And Maintenance

The development system should make app shipping boring.

Required developer ergonomics:

- One app contract package for session envelope, artifact metadata, route
  evidence, sync events, and diagnostic event names.
- Golden fixtures for session resume, offline event append, sync reconciliation,
  route evidence, and approval boundaries.
- Platform smoke tests that launch the app, open latest artifact, append an
  offline event, reconnect, and verify the same session id.
- CI jobs for type checks, unit tests, accessibility checks, bundle-size checks,
  dependency audit, signed-update verification, and publish-readiness.
- Local dev command that prints route, provider, model profile, app version,
  sync state, and logging endpoint without exposing secrets.
- App-store and package metadata generated from checked-in source, not edited by
  hand in release week.

## App Shipping Gates

Before any app ships publicly, it must pass:

- session continuity gate
- offline and sync gate
- route evidence gate
- security and privacy gate
- update signing and rollback gate
- platform packaging gate
- performance budget gate
- observability and diagnostics gate
- accessibility gate
- support bundle gate

No gate can depend on tribal knowledge. Each gate needs a checked-in command,
fixture, or release checklist.

## Source-Backed Inputs

- Apple App Sandbox:
  [developer.apple.com/documentation/security/protecting-user-data-with-app-sandbox](https://developer.apple.com/documentation/security/protecting-user-data-with-app-sandbox)
- Apple notarization:
  [developer.apple.com/documentation/security/notarizing-macos-software-before-distribution](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution)
- Android security:
  [developer.android.com/privacy-and-security/security-best-practices](https://developer.android.com/privacy-and-security/security-best-practices)
- Android Baseline Profiles:
  [developer.android.com/topic/performance/baselineprofiles/overview](https://developer.android.com/topic/performance/baselineprofiles/overview)
- Android vitals:
  [developer.android.com/topic/performance/vitals](https://developer.android.com/topic/performance/vitals)
- Play Integrity:
  [developer.android.com/google/play/integrity](https://developer.android.com/google/play/integrity)
- Windows MSIX:
  [learn.microsoft.com/windows/msix/overview](https://learn.microsoft.com/windows/msix/overview)
- Windows packaging:
  [learn.microsoft.com/windows/apps/package-and-deploy/packaging](https://learn.microsoft.com/windows/apps/package-and-deploy/packaging/)
- OWASP MASVS:
  [mas.owasp.org/MASVS](https://mas.owasp.org/MASVS/)
- Tauri security:
  [v2.tauri.app/security](https://v2.tauri.app/security/)
- Electron security:
  [electronjs.org/docs/latest/tutorial/security](https://www.electronjs.org/docs/latest/tutorial/security)
- Expo EAS Update code signing:
  [docs.expo.dev/eas-update/code-signing](https://docs.expo.dev/eas-update/code-signing/)
- Flutter app size:
  [docs.flutter.dev/perf/app-size](https://docs.flutter.dev/perf/app-size)
- Web Vitals:
  [web.dev/articles/vitals](https://web.dev/articles/vitals)
- Next.js instrumentation:
  [nextjs.org/docs/app/guides/instrumentation](https://nextjs.org/docs/app/guides/instrumentation)
- OpenTelemetry:
  [opentelemetry.io/docs/what-is-opentelemetry](https://opentelemetry.io/docs/what-is-opentelemetry/)
