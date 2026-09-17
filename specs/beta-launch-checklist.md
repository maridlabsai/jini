# Jini Beta Launch Checklist

The gated path from **today** (main = the matured product, nightly builds green on
every platform) to a **public beta**. Each item is a checkable state, not a vibe.
Ordered by dependency: nothing in a later phase starts until its gate is met.

Legend: `[x]` done · `[ ]` open · `⛳` blocked on an escalated human input.

## Phase 0 — Engineering readiness (mostly done)

- [x] `main` carries the matured product (PR #12, 191 commits ahead of v0.1.2)
- [x] Cross-platform build matrix green: darwin arm64/amd64, linux arm64/amd64, windows
- [x] Nightly channel publishes 10/10 assets with `.sha256` checksums
- [x] `install.sh` installs the prebuilt binary on macOS (incl. Apple-Silicon ad-hoc), Linux, Windows
- [x] `jini version` / `jini update` channel plumbing in place
- [x] Self-maintenance live: `health.yml`, `triage-digest.yml`, `competitive-intel.yml`, Dependabot (gomod/actions/bundler)
- [x] Model onboarding trilogy: `jini models` → `jini check model` → catalog auto-merge
- [ ] Repo hardening: branch protection (require CI green) + auto-merge (⛳ needs Admin grant; script ready)
- [ ] All open Dependabot PRs merged (rebased onto current main, CI green)

## Phase 1 — Beta blockers (the real gate)

- [ ] ⛳ **Apple Developer ID cert** → cut a **signed + notarized stable release**
      (nightly works unsigned; beta testers should not hit a Gatekeeper wall)
- [ ] Verify the signed release installs clean on a **fresh** macOS machine
      (`JINI_REQUIRE_SIGNED=1` path passes) and a fresh Linux box
- [ ] `jini update` verified to move a machine from an older build to the new one
- [ ] **Free-provider capacity sanity check** — the load-bearing GTM risk: confirm
      Cerebras/Groq/Gemini free tiers behave under, say, 50–100 concurrent testers.
      Have the "bring your own free key" + local-floor fallback message ready.
- [ ] A one-command uninstall / rollback exists and is documented

## Phase 2 — Beta assets & messaging

- [ ] Landing section leads with the wedge: *"the AI coding CLI that never rations
      you"* — savings receipt + throttle-survival demo above the fold
- [ ] Quickstart: install → first task → `jini savings` in under 5 minutes
- [ ] `jini feedback` path tested end-to-end (files a GitHub issue, upvotable)
- [ ] A short demo clip (the automated video pipeline, or one hand-made) showing a
      rival throttling and Jini not
- [ ] CHANGELOG / release notes for the beta build

## Phase 3 — Launch execution (soft, seeded)

- [ ] Seed list: ~50–100 devs sourced from the exact threads where Codex/Kiro users
      complain about credits/throttling (r/codex, r/kiroIDE, the OpenAI forum) —
      invite, don't spam; lead with "here's the throttle problem, solved"
- [ ] Waitlist / referral loop live (Robinhood-style bump) — optional for a small beta
- [ ] Instrumentation decided **before** launch: activation (install → first green
      task), savings receipts shared, D1/D7 retention, throttle-survival events
- [ ] A public place for beta feedback (GitHub Discussions or the triage digest)
- [ ] On-call plan: who watches `health.yml` failures and escalations to
      sharmas@outlook.com during the beta window

## Phase 4 — Beta → GA exit criteria

- [ ] D7 retention clears an agreed bar
- [ ] Crash-free / gate-clean across the beta cohort
- [ ] **Free-tier capacity held** at beta scale (or the BYO-key story proven as the
      graceful fallback) — this decides whether the whole "un-metered" wedge scales
- [ ] Early free→paid signal observed (validates the revenue model before spend)
- [ ] CAC-per-channel measured on the seeded channels; kill the losers before
      committing the 50%-of-revenue marketing budget

## Rollback / kill switch (must exist before Phase 1 ships)

- [ ] A bad release can be pulled: delete/deprecate the release, and the install
      one-liner falls back to the previous good channel build
- [ ] `jini update` cannot brick a working install (verified on a downgrade path)

---

**Right now, the single highest-leverage unlock is the Apple Developer ID cert**
(Phase 1) — it turns the green nightly pipeline into a signed stable release that
beta testers can install without friction. Everything above it in Phase 0 is done
or one grant away; everything below it waits on it.
