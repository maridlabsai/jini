# Jini macOS App

This directory contains the Phase 1 macOS shell for Jini. It is a Tauri 2 app
that launches the Go core as a sidecar:

```bash
jini app serve --stdio --surface macos
```

The app is intentionally thin:

- Tauri owns the window, renderer, native bundle, and sidecar process.
- Go owns intent, routing, sessions, approvals, diagnostics, and side effects.
- The renderer may not read or write project files directly.

## Current State

This is an internal dogfood scaffold, not a signed public release. It provides:

- project/session browser layout
- task composer with compact answer rendering
- route, diff, artifact, approval, and diagnostics panels
- scoped sidecar bridge to the Go protocol
- diagnostics export through the Go sidecar

Still required before user release:

- project open and CLI/app session continuity
- direct file-edit approvals and diff projections
- offline resume and sync debt projection
- signing, notarization, stapling, and Gatekeeper smoke
- public `.dmg` or `.app` release asset

## Development

```bash
npm install
npm run dev
```

`npm run prepare:sidecar` builds the Go sidecar into
`src-tauri/binaries/jini-sidecar-<target-triple>` for Tauri packaging.
