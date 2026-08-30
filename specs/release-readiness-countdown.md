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
| 7 | Hand-off CLI compatibility | 🟡→ | claude empirically verified (`posture verified` in `jini route list`); codex/gemini/aider/opencode now labeled **`posture experimental (doc-verified)`** so unverified posture is never release-claimed. To flip them to verified: a behavioral posture-validation harness (plan=no-writes / semi=edits-only / autonomous=commands) run with each CLI installed, version-pinned in `.jini/cli-dogfood.json` |
| 8 | **Packaging: signed/notarized asset + `install.sh` on a clean machine** | ⬜ | macos_bundle_hygiene gate exists; need a real signed release + fresh-install smoke on macOS |
| 9 | **BYO credential validation + Grok fixtures** (typed errors, keychain) | ⬜ | PRD v1 backlog |
| 10 | **Cross-platform TTFV < 5 min + first-task success (macOS/Linux/Win)** + public quickstart docs | ⬜ | needs multi-OS run + alpha docs |

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
