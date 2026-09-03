# Release Readiness Countdown

Evidence-based tracker of what remains before release. No fabricated dates —
progress is measured in **release gates green / total**, with honest effort
bands. Update as gates flip.

## Free tier — the public release

**Countdown: 7 of 10 release gates green. 3 to go.**

| # | Release gate | Status | Evidence / what's left |
| --- | --- | --- | --- |
| 1 | Core CLI behavior (P0 requirements) | ✅ | trace 16/16 = 100% |
| 2 | Commit gate green (build/tests/lint/drift/scorecard) | ✅ | run each commit |
| 3 | Push gate green (publish-readiness + real CLI dogfood evidence) | ✅ | exit 0; `.jini/cli-smoke.json` + `cli-dogfood.json` for claude-code |
| 4 | Functional harness + `jini check functional` self-check | ✅ | 17/17 offline |
| 5 | Maturity guardrails (security/accessibility/readability/tone/citations) | ✅ | respguard + corpus |
| 6 | Attachments incl. multimodal (Anthropic + OpenAI vision) | ✅ | bedrock Converse deferred (non-blocker) |
| 7 | Hand-off CLI compatibility | 🟡→ | claude empirically verified (`posture verified`); others labeled **`posture experimental (doc-verified)`** so unverified posture is never release-claimed. **Behavioral posture harness built** (`jini route validate <route> --posture`): runs plan/semi/autonomous in a scratch dir and proves plan=read-only (live-confirmed on claude) + escalations apply edits; fake-CLI tested (catches a plan-write or a no-op escalation). Remaining to fully close: run it per real CLI on install (codex/gemini/aider/opencode) and record version-pinned evidence |
| 8 | **Packaging: signed/notarized asset + `install.sh` on a clean machine** | 🟡→ | **`tools/release_macos.sh`** builds→codesigns (hardened runtime)→notarizes→packages tar.gz + `.sha256` (identity/profile from env, never in repo). **install.sh now verifies integrity**: SHA-256 checksum is enforced (tampered asset is refused — proven end-to-end locally), and macOS signature is checked with `codesign --verify -R="anchor apple generic"` (ad-hoc/self-signed rejected; unsigned allowed only pre-signing, or `JINI_REQUIRE_SIGNED=1` to enforce). CI already smokes source+release install on macOS+Linux. Remaining: run `release_macos.sh` with the real Apple Developer ID cert to produce a notarized asset and confirm Gatekeeper acceptance on a clean machine (needs the signing certificate) |
| 9 | **BYO credential validation + Grok fixtures** (typed errors, keychain) | 🟡→ | **Typed credential taxonomy** (401 invalid / 403 forbidden / 429 rate-limited / 404 model / 5xx) wired into all remote providers, and 429 still cooperates with throttle survival. **`jini provider validate [shape]`** makes one live read-only call and reports a typed result. **OpenAI-compatible BYO family** (openai/xai(grok)/groq/deepseek/mistral) validates AND routes from one shape registry — a pasted key works with a default model (no extra config); explicit-select only (no ambient auto-adopt, frugality). Per-shape fixtures pass (a shape without a passing fixture is not claimed). Remaining: OS keychain storage (keys are in 0600 `provider.json` today) and per-shape receipt-denomination pricing |
| 10 | **Cross-platform TTFV < 5 min + first-task success (macOS/Linux/Win)** + public quickstart docs | 🟡→ | Public quickstart/install/examples docs are live (docs/index, install, simple, examples, cli). jini **cross-compiles clean** for windows/amd64+arm64, linux/arm64, darwin (verified locally); CI now adds a **windows-latest build+commands smoke** alongside the existing macOS+Linux installer smokes. Remaining: measure real TTFV < 5 min + first-task success on fresh macOS/Linux/Windows machines (needs real multi-OS runs; note the full Go suite still assumes Unix shell/exec, so Windows CI is build+smoke, not the whole suite) |

Non-blockers (fine to ship without, follow after): real metered-usage capture
(savings is imputed-only today), the 2 residual-hardening items, bedrock vision.

**Honest estimate:** the *product code* is essentially done — gates 8–10 are
**packaging, validation, and docs**, not feature work. Realistic band: **~1–2
weeks**, dominated by cross-platform testing + signing/notarization pace, not
engineering. Gate 7 can be closed instantly by labeling non-claude routes
experimental in release notes, or when those CLIs are installed for a real
`--help`/dogfood pass.

## Commercial tier — follows the free release

**Countdown: 0 of 5 gates green (starts after free ships).** Larger scope: a new
repo, billing, and managed infra.

| # | Release gate | Status | What's left |
| --- | --- | --- | --- |
| 1 | `../jini-commercial` repo bootstrap (separate Go module on the public seams: autopilot/, runner/, agentloop/) | ⬜ | does not exist yet |
| 2 | Paid Autopilot policy (predictive throttle avoidance, route switching, savings optimization) | ⬜ | public keeps only the fail-closed gate + free equivalent |
| 3 | Plus/Pro entitlements + billing (annual-first to dodge the Stripe floor; Plus serverless) | ⬜ | paywall wiring; prices stay out of the public repo |
| 4 | Managed native loop (Plus), Continuity (Pro), decision-tree backtrack (Pro) | ⬜ | Pro backtrack depends on the shipped agentloop recorder/checkpointer seams |
| 5 | Pricing finalized (Free / Plus $2.99 / Pro $5.99) + customer-facing messaging | ⬜ | kept simple; commercial-repo only |

**Honest estimate:** **weeks after** free tier — dominated by standing up the new
repo + billing/entitlements + managed infrastructure, not by the public seams
(already exported and tested).
