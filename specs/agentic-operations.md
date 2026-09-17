# Jini Agentic Operations — the solo-maintainer operating model

Maintaining Jini must be least-taxing for a single developer. This is the map of
what runs itself, on what trigger, and where a human is the exception. The
principle: **automate the toil, escalate the judgment.** Right-sized on purpose —
CI for event-driven work, a few scheduled routines for time-driven drift, the
dogfood loop for development, and a precise escalation for the handful of inputs
automation cannot fabricate. Not an agent zoo.

## The three layers

| Layer | Trigger | Carries |
| --- | --- | --- |
| **Event automation** (GitHub Actions) | push / PR / tag | builds, tests, gates, security, releases, site deploy, catalog auto-merge |
| **Scheduled automation** (cron) | time | drift the repo can't see: model freshness, dependency CVEs, competitive/feature signal |
| **Dogfood loop** | the developer's intent | development itself — `jini-pro` proxies to a coding agent under the gates |

Everything a human still touches is either a **judgment call** (approve a
direction, a design) or an **input automation cannot create** — escalated.

## By aspect

| Aspect | How it's carried | State |
| --- | --- | --- |
| **Development** | Dogfood loop (`JINI_HOME=$HOME/.jini-dogfood jini "<task>"` = jini-pro) + the commit gate + pre-commit review/premortem | ✅ live |
| **Validation** | `tools/run_required_gates.sh` on every commit/PR (build/test/lint/PRD-drift/scorecard); `jini check model` for model quality; `health.yml` re-runs both on a daily cron and opens an issue on drift | ✅ live |
| **Builds** | `ci-installer.yml` (mac/linux + windows); `release.yml` cross-compile matrix, ldflags version/channel stamp | ✅ live |
| **Updates / releases** | `release.yml` (tag→stable, `-beta.N`→beta, main→nightly); `jini version` / `jini update`; install.sh `--channel` | ✅ code · ⛳ signed assets need the cert (escalated) |
| **Deployments** | GitHub Releases (assets + checksums); website via GitHub Pages on push | ✅ releases · ⛳ Pages needs the domain (escalated) |
| **Model/provider onboarding** | `jini models` (discovery) → `jini check model` (auto quality gate) → `providers.json` catalog → `catalog-auto-merge.yml` (schema + live gate → auto-merge) | ✅ live |
| **Design** | Occasional, agent-assisted (artifact-design); not a recurring load | judgment |
| **Marketing** | Viral loops built into the product (shareable receipts, throttle demo, OSS); `competitive-intel.yml` scans HN + competitor GitHub issues + Reddit weekly into one living issue, feeding roadmap + messaging | ✅ live |
| **Feedback / roadmap** | `jini feedback [bug\|ask]` files to GitHub (👍 = upvote); `triage-digest.yml` ranks by reactions weekly into one living issue; priority call is human | ✅ live |
| **Payments** | Stripe checkout → webhook → entitlement flips `JINI_SUBSCRIPTION_TIER=commercial` (seam exists in `../jini-commercial`) | ⛳ needs Stripe (escalated) |
| **Security** | `security.yml` (CodeQL SAST, govulncheck, OSV-Scanner, TruffleHog, scheduled + on PR); `dependabot.yml` (gomod + actions, weekly); `security-scan` skill | ✅ live |
| **Community** | Gate-enforced PRs; `catalog-auto-merge.yml`; "no fixture, not claimed"; protected core behind the PRD-drift gate | ✅ live |

Legend: ✅ automated · ⏳ automatable next (no new human input) · ⛳ blocked on an
escalated human input · judgment = stays a human decision by design.

## The exception path — escalation

When automation genuinely cannot proceed (an input only the owner can create),
it does **not** stall — it escalates to **sharmas@outlook.com** with a precise
ask + numbered steps (see the launch escalation). The current human-only inputs:

1. **Apple Developer ID cert** → signed/notarized macOS releases.
2. **Domain + Pages** → the website and install URL.
3. **Stripe account + products** → billing/entitlement → revenue.
4. **Repo settings** (one-time): enable Allow-auto-merge + branch protection to
   the gate checks; add a `<ID>_API_KEY` secret per provider to live-verify
   catalog PRs.

Once those land, the loop is closed: the product builds, validates, ships,
updates, onboards models/providers, and bills — with the developer approving
directions, not doing toil.

## What's left to build

- **Signed remote catalog fetch** (blocked on signing) — makes community
  providers instant instead of shipping in the next release via the embedded
  catalog. Depends on the release-signing key/cert (escalated), so it is not
  purely automatable yet.

The scheduled health workflow, Dependabot, the dependency/CVE audit, and the
competitive-intel digest that this section used to list are **done** — see the
Validation, Security, and Marketing rows above. The last remaining automatable
item is the signed remote catalog fetch, which is gated on the release cert.
