---
title: Commercial
description: The shell stays free. Planned app downloads stay free. The paid layer should exist only when Jini can prove it saves money or keeps work moving.
---

<p class="page-lead">Jini should be easy to adopt and hard to overpay for. The shell stays free. Planned app downloads stay free. The paid layer should exist only when Jini can show real savings or continuity proof that a team would miss without it.</p>

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>What stays free</h3>
    <p>CLI, runtime, docs, examples, tests, local SLM use, BYO provider use, resumable work state, and inspectable artifacts stay in the public product surface.</p>
  </div>
  <div class="checklist-card">
    <h3>What downloads are free</h3>
    <p>macOS, Windows, iOS, and Android app shells are planned to be free downloads so users can review and resume the same session anywhere once release blockers are cleared.</p>
  </div>
  <div class="checklist-card">
    <h3>What the $1 license unlocks</h3>
    <p>Planned Commercial License features: provider-limit forecasting, throttle avoidance, automatic fallback, and automatic resume after limits or throttles hit. Everyone should get a 30-day free trial before asking for the $1/month subscription.</p>
  </div>
</div>

<div class="section-card" markdown="1">
## The short version

- use the free shell when you want one stable place to run and resume work
- expect planned desktop and mobile apps to be free downloads when they are actually ready
- start with a 30-day free trial of the commercial optimizer
- pay only if Jini can prove that it saved money or prevented stalled work
</div>

<div class="section-card" markdown="1">
## What is not open source

The app downloads can be free without making the app implementation public.

- desktop/mobile app source code lives in the commercial repo only
- native wrappers, host manifests, release automation, and store-delivery code stay private
- the public repo may describe the app surfaces, but it should not ship the app code itself
</div>

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>The rule</h3>
    <p><strong>Do not charge for the shell. Do not charge for app downloads. Give everyone a 30-day free trial before asking for the $1/month subscription. Charge only for the cost-saver and continuity layer that proves its value at runtime.</strong></p>
  </div>
  <div class="checklist-card">
    <h3>Why upgrade</h3>
    <p>Upgrade when provider limits, throttles, or expensive routes are a real operating problem and you want Jini to keep work moving automatically instead of forcing manual babysitting. The paywall prompt should appear before downgrade, not after the account is already squeezed into constrained free mode.</p>
  </div>
</div>

<div class="section-card" markdown="1">
## Distribution rule of thumb

- distribute directly from the website first when platform policy allows it and store fees do not buy something essential
- keep macOS and Windows direct-first
- keep Android direct-first when policy allows, with Play Store secondary only if trust or reach justify it
- accept that iOS remains App Store constrained
</div>

<div class="section-card" markdown="1">
## Free app surfaces, when ready

- **macOS and Windows:** deeper review, artifact opening, session continuation, and renewal-proof inspection
- **iOS and Android:** quick session review, approval/defer flows, and interruption-safe continuation
- **CLI remains first-class:** the apps are another surface over the same session, not a separate product

The commercial apps should help a wider user base without turning Jini into a different workflow for each device class.
</div>

<div class="section-card" markdown="1">
## Current readiness and payment status

| Surface | Status | Planned activation |
|---|---|---|
| CLI | Available now | None |
| macOS app shell | Preview only. Not downloadable yet | Start with a 30-day free trial, then buy on the website and sign in |
| Windows app shell | Preview only. Not downloadable yet | Start with a 30-day free trial, then buy on the website and sign in |
| iOS companion app | Preview only. Not on the App Store yet | Sign in with an existing paid account |
| Android companion app | Preview only. Not downloadable yet. Direct-first when policy allows, with Play Store secondary | Sign in with an existing paid account |
| Commercial License checkout | Planned. Not live yet | Start with a 30-day free trial, then website checkout + account entitlement |

Payment integration is not live yet. The current product work defines the shared entitlement model and distribution plan, but not a production checkout or store-billing pipeline.

Current posture: preview and planning slices are implemented, but the apps are not yet release-ready for direct public download or store rollout.
</div>

<div class="section-card" markdown="1">
## What the paid layer must prove before renewal

If the paid layer is worth keeping, it should show evidence that the free shell alone could not provide:

- month-to-date token savings
- provider headroom preserved
- throttles avoided
- throttles auto-recovered
- sessions resumed without manual babysitting

If Jini cannot prove that value, it should not expect renewal.
</div>
