# Release Readiness Countdown

Readiness is reported **only** from executed CI evidence — each claim must carry a
commit, environment, command, exit status, and result digest. Nothing below is
"passing" or "green" until such an artifact exists.

## Current readiness: UNKNOWN

During the trust-first recovery (JINI-R0), every prior readiness claim is treated
as an **unverified historical snapshot**. Current release readiness is **unknown**
and is not asserted here as a count, percentage, or date. It will be regenerated
from executed CI evidence (commit, environment, command, exit status, result
digest) before any gate is reported as passing.

---

## Historical snapshot — UNVERIFIED, NOT current readiness

> The tables and estimates below are kept for context only. They were derived from
> in-repo status labels and a scorecard that checks for a test *name*, not from
> executed test artifacts. **Do not read any mark below as current truth.** Each
> item must be regenerated from executed CI evidence before it can be reported as
> passing. The status column is deliberately neutralized to `unverified`.

### Free tier — the public release (historical snapshot)

| # | Release gate | Status | Historical note (unverified) |
| --- | --- | --- | --- |
| 1 | Core CLI behavior (P0 requirements) | unverified | snapshot label claimed trace 16/16; not backed by an executed artifact |
| 2 | Commit gate (build/tests/lint/drift/scorecard) | unverified | intended to run each commit; needs an executed-CI digest |
| 3 | Push gate (publish-readiness + real CLI dogfood evidence) | unverified | snapshot referenced `.jini/cli-smoke.json` + `cli-dogfood.json` for claude-code |
| 4 | Functional harness + `jini check functional` self-check | unverified | snapshot label claimed 17/17 offline; needs an executed-CI digest |
| 5 | Maturity guardrails (security/accessibility/readability/tone/citations) | unverified | respguard + corpus present in-repo; not independently verified here |
| 6 | Attachments incl. multimodal (Anthropic + OpenAI vision) | unverified | bedrock Converse deferred; follow-up, not a release claim |
| 7 | Hand-off CLI compatibility | unverified | claude-code empirically exercised (`posture verified`); other CLIs labeled `posture experimental (doc-verified)`. Behavioral posture harness (`jini route validate <route> --posture`) runs plan/semi/autonomous in a scratch dir (plan=read-only; escalations apply edits; fake-CLI tested). To regenerate: run it per real CLI on install (codex/gemini/aider/opencode) and record version-pinned executed evidence |
| 8 | Packaging: signed/notarized asset + `install.sh` on a clean machine | unverified | `tools/release_macos.sh` builds→codesigns (hardened runtime)→notarizes→packages tar.gz + `.sha256` (identity/profile from env, never in repo). `install.sh` enforces SHA-256 and checks the macOS signature with `codesign --verify -R="anchor apple generic"`. To regenerate: run `release_macos.sh` with the real Apple Developer ID cert and confirm Gatekeeper acceptance on a clean machine (needs the signing certificate) |
| 9 | BYO credential validation + Grok fixtures (typed errors, keychain) | unverified | typed credential taxonomy (401/403/429/404/5xx) across remote providers; `jini provider validate [shape]` makes one live read-only call; OpenAI-compatible BYO family (openai/xai(grok)/groq/deepseek/mistral) validates and routes from one shape registry; macOS login-keychain storage with 0600-dotfile fallback. To regenerate: Linux/Windows secret backends and per-shape receipt-denomination pricing |
| 10 | Cross-platform TTFV + first-task success (macOS/Linux/Win) + public quickstart docs | unverified | public quickstart/install/examples docs live; cross-compiles for windows/amd64+arm64, linux/arm64, darwin; CI has a windows build+commands smoke alongside macOS+Linux installer smokes. To regenerate: measure real TTFV and first-task success on fresh macOS/Linux/Windows machines (needs real multi-OS runs; the full Go suite assumes Unix shell/exec, so Windows CI is build+smoke) |

Follow-after items (not release claims, and not a basis for shipping): real
metered-usage capture (savings is imputed-only today), residual hardening, bedrock
vision. These are deferred work, not evidence that the gates above pass.

The recovery gaps surfaced by the trust-first program — execution safety, durable
continuation, default verification, honest economics — are **feature work**, not
only packaging, and must be reflected whenever readiness is regenerated.

### Commercial tier — follows the free release (historical snapshot)

> Same caveat: unverified snapshot, not current readiness. No "N of M green" count
> is asserted as current.

| # | Release gate | Status | Historical note (unverified) |
| --- | --- | --- | --- |
| 1 | `../jini-commercial` repo bootstrap (separate Go module on the public seams: autopilot/, runner/, agentloop/) | unverified | snapshot described a separate module + `cmd/jini-pro` building against the public `autopilot`/`runner` seams via a local replace; cannot import `internal/` |
| 2 | Paid Autopilot policy (predictive throttle avoidance, route switching, savings optimization) | unverified | snapshot described a route-switching ladder (`ladder/` implementing `autopilot.Strategy`). Remaining: predictive avoidance + savings optimization, and a live throttle→switch capture |
| 3 | Plus/Pro entitlements + billing (annual-first; Plus serverless) | unverified | paywall wiring; prices stay out of the public repo |
| 4 | Managed native loop (Plus), Continuity (Pro), decision-tree backtrack (Pro) | unverified | Pro backtrack depends on the agentloop recorder/checkpointer seams |
| 5 | Pricing finalized (Free / Plus $2.99 / Pro $5.99) + customer-facing messaging | unverified | kept simple; commercial-repo only |

The commercial tier timing is not estimated here; it depends on standing up the
new repo plus billing/entitlements and managed infrastructure, and is regenerated
from executed evidence like everything else.
