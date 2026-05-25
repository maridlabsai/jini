---
title: Commercial
description: The shell stays free. App downloads stay free when live. The paid layer should exist only when Jini can prove it saves money or keeps work moving.
eyebrow: Paid only after proof
context_line: The commercial story should be blunt: keep adoption cheap, keep downloads free when surfaces are live, and charge only for the optimizer layer that proves measurable value.
highlights:
  - Shell stays free
  - Free downloads when live
  - Planned 30-day trial
  - $1 only after proof
quick_links:
  - label: Proof
    href: /proof.html
  - label: Install
    href: /install.html
  - label: Contact
    href: /contact.html
---

<p class="page-lead">Jini should be easy to adopt and hard to overpay for. The shell stays free. App downloads stay free when each surface is live. The paid layer should exist only when Jini can show real savings or continuity proof that a team would miss without it.</p>

<div class="section-card" markdown="1">
## The short version

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>What stays free</h3>
    <p>CLI, runtime, docs, examples, tests, local SLM use, BYO provider use, resumable work state, and inspectable artifacts stay in the public product surface.</p>
  </div>
  <div class="checklist-card">
    <h3>What downloads are free</h3>
    <p>macOS, Windows, iOS, and Android app downloads stay free whenever those surfaces are live so users can review and resume the same session anywhere.</p>
  </div>
  <div class="checklist-card">
    <h3>What the $1 license unlocks</h3>
    <p>When live, the Commercial License unlocks provider-limit forecasting, throttle avoidance, automatic fallback, and automatic resume after limits or throttles hit. Everyone should get the planned 30-day free trial before asking for the $1/month subscription.</p>
  </div>
</div>

<div class="quote-strip">
  <strong>Do not charge for the shell. Do not charge for app downloads. Give everyone the planned 30-day free trial before asking for the $1/month subscription. Charge only for the cost-saver and continuity layer that proves its value at runtime.</strong>
  <p>Use the free shell when you want one stable place to run and resume work. Expect desktop and mobile apps to be free downloads when each surface is actually live. Start with the planned 30-day free trial of the commercial optimizer and pay only if Jini can prove that it saved money or prevented stalled work.</p>
</div>

<p>Upgrade when provider limits, throttles, or expensive routes are a real operating problem and you want Jini to keep work moving automatically instead of forcing manual babysitting. The paywall prompt should appear before downgrade, not after the account is already squeezed into constrained free mode.</p>
</div>

<div class="section-card" markdown="1">
## Free shell vs paid optimizer

| Need | Free shell | Paid optimizer |
|---|---|---|
| Start and finish work | Included | Not required |
| Use local models and your own provider accounts | Included | Not required |
| Keep work resumable and inspectable | Included | Not required |
| Planned app downloads when released | Included | Not required |
| Predict provider limits before they block work | Not included | Included after the planned 30-day free trial when checkout is live |
| Avoid throttles automatically | Not included | Included after the planned 30-day free trial when checkout is live |
| Fall back and resume automatically | Not included | Included after the planned 30-day free trial when checkout is live |
| Measured savings and continuity proof | Basic route evidence | Included after the planned 30-day free trial when checkout is live |

The comparison should stay blunt: the free shell must already be useful, and
the paid layer should exist only where automation, continuity, and savings are
meaningfully better than manual babysitting.
</div>

<div class="section-card" markdown="1">
## When the paid layer earns the right to exist

- when throttles or provider limits regularly interrupt active work
- when route cost has become an operating problem instead of a background concern
- when the user would otherwise need manual fallback and manual resume steps
- when the proof can be shown before payment, not explained after payment

If those conditions are not true, the free shell should remain enough.
<p class="page-lead"><strong>What the paid layer must prove before renewal</strong></p>

If the paid layer is worth keeping, it should show evidence that the free shell alone could not provide:

- month-to-date token savings
- provider headroom preserved
- throttles avoided
- throttles auto-recovered
- sessions resumed without manual babysitting

If Jini cannot prove that value, it should not expect renewal.
 </div>

<div class="section-card" markdown="1">
## Current readiness and payment status

### Distribution rule of thumb

- distribute directly from the website first when platform policy allows it and store fees do not buy something essential
- keep macOS and Windows direct-first
- keep Android direct-first when policy allows, with Play Store secondary only if trust or reach justify it
- accept that iOS remains App Store constrained

### Free app surfaces, once live

- **macOS and Windows:** deeper review, artifact opening, session continuation, and renewal-proof inspection
- **iOS and Android:** quick session review, approval/defer flows, and interruption-safe continuation
- **CLI remains first-class:** the apps are another surface over the same session, not a separate product

The commercial apps should help a wider user base without turning Jini into a different workflow for each device class.
<p class="page-lead"><strong>What is not open source</strong></p>
<p>The app downloads can be free without making the app implementation public.</p>

- desktop/mobile app source code lives in the commercial repo only
- native wrappers, host manifests, release automation, and store-delivery code stay private
- the public repo may describe the app surfaces, but it should not ship the app code itself

<div class="pill-list">
  {% for surface in site.data.public_surfaces.surfaces %}
  <span>{{ surface.name }}: {{ surface.badge }}</span>
  {% endfor %}
</div>

| Surface | Badge | Current state | Planned activation |
|---|---|---|---|
{% for surface in site.data.public_surfaces.surfaces %}
| {{ surface.name }} | {{ surface.badge }} | {{ surface.current_state }} | {{ surface.activation }} |
{% endfor %}

Checkout is not live yet. The current product work defines the shared entitlement model and distribution plan, but not a production checkout or store-billing pipeline.

Current posture: preview and planning slices are implemented, but the apps are not yet release-ready for direct public download or store rollout.

This table is fed from <code>docs/_data/public_surfaces.json</code>, which is a sanitized snapshot built from the commercial release packets instead of hand-maintained website copy.
</div>
