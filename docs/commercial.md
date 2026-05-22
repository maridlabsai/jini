---
title: Commercial
description: The core shell stays open. Desktop/mobile app shells are preview-only today and planned to be free downloads once released. Planned pricing for the Commercial License is $1/month.
---

<p class="page-lead">The core Jini shell stays open. Today only the CLI is installable. Desktop and mobile app shells are preview-only today, planned to be free downloads once released, and planned to pair with a $1/month Commercial License once checkout and entitlement activation are live.</p>

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>What stays open</h3>
    <p>CLI, runtime, docs, examples, tests, local SLM use, BYO provider use, compaction, checkpointed resume, and projection-first continuity stay in the public product surface.</p>
  </div>
  <div class="checklist-card">
    <h3>What downloads are free</h3>
    <p>macOS, Windows, iOS, and Android app shells are planned to be free downloads so users can review and resume the same session anywhere once release blockers are cleared.</p>
  </div>
  <div class="checklist-card">
    <h3>What the $1 license unlocks</h3>
    <p>Planned Commercial License features: cost optimization, provider-limit forecasting, throttle avoidance, automatic fallback, and automatic resume after limits or throttles hit.</p>
  </div>
</div>

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>The rule</h3>
    <p><strong>Do not charge for access to the open shell or for downloading the apps. Charge for the adaptive optimizer and the provider-aware continuity layer.</strong></p>
  </div>
  <div class="checklist-card">
    <h3>Why upgrade</h3>
    <p>Upgrade when you routinely get close to provider limits, want 30-50% token savings, and need Jini to keep work moving automatically when hosted tools throttle or hit subscription ceilings.</p>
  </div>
</div>

<div class="section-card" markdown="1">
## Commercial app surfaces

- **macOS and Windows:** deeper review, artifact opening, session continuation, and renewal-proof inspection
- **iOS and Android:** quick session review, approval/defer flows, and interruption-safe continuation
- **CLI remains first-class:** the apps are another surface over the same session, not a separate product

The commercial apps should help a wider user base without turning Jini into a different workflow for each device class.
</div>

<div class="section-card" markdown="1">
## Distribution and payment status

| Surface | Status | Planned activation |
|---|---|---|
| CLI | Available now | None |
| macOS app shell | Preview only. Not downloadable yet | Buy on the website, then sign in |
| Windows app shell | Preview only. Not downloadable yet | Buy on the website, then sign in |
| iOS companion app | Preview only. Not on the App Store yet | Sign in with an existing paid account |
| Android companion app | Preview only. Not on the Play Store yet | Sign in with an existing paid account |
| Commercial License checkout | Planned. Not live yet | Website checkout + account entitlement |

Payment integration is not live yet. The current product work defines the shared entitlement model and distribution plan, but not a production checkout or store-billing pipeline.

Current posture: preview and planning slices are implemented, but the apps are not yet release-ready for direct public store rollout.
</div>

<div class="section-card" markdown="1">
## Why people keep renewing each month

The subscription has to prove itself with runtime evidence:

- month-to-date token savings
- provider headroom preserved
- throttles avoided
- throttles auto-recovered
- sessions resumed without manual babysitting

If Jini cannot prove that value, it should not expect renewal.
</div>
