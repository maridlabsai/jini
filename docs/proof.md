---
title: Trust
description: Free value should be obvious quickly. Paid value should be visible as savings, recovered work, and route evidence before anyone is asked to renew.
eyebrow: Trust without ceremony
context_line: Proof should support the product, not bury it. This page is the explicit trust and renewal boundary for what the public site is allowed to claim.
highlights:
  - Free value first
  - Renewal proof
  - Preview honesty
  - No fake telemetry
quick_links:
  - label: Commercial
    href: /commercial.html
  - label: Install
    href: /install.html
  - label: Examples
    href: /examples.html
---

<div class="section-card">
  <span class="section-kicker">{{ site.data.public_proof.hero.eyebrow }}</span>
  <h2>{{ site.data.public_proof.hero.headline }}</h2>
  <p class="page-lead">Proof has two jobs. First, the free shell should make its value obvious fast. Second, any paid layer should prove that it saved money or prevented stalled work before anyone is asked to pay or renew.</p>
  <p>{{ site.data.public_proof.hero.body }}</p>
  <p>{{ site.data.public_proof.sections[2].bullets[0] }}</p>
  <div class="proof-grid">
    {% for card in site.data.public_proof.proof_cards %}
    <div class="proof-card">
      <h3>{{ card.label }}</h3>
      <p><strong>{{ card.value }}</strong></p>
      {% if forloop.index0 == 0 %}
      <p>One continuity story across CLI, desktop, and mobile surfaces instead of four disconnected product stories.</p>
      {% elsif forloop.index0 == 1 %}
      <p>A public demo is acceptable only when it is tied to measurable token savings, a visible trial-and-renewal story, and an explicit preview boundary.</p>
      {% else %}
      <p>The paid layer should prove that work recovered without the user rebuilding state by hand.</p>
      {% endif %}
    </div>
    {% endfor %}
  </div>
</div>

<div class="section-card">
  <span class="section-kicker">Free value first</span>
  <h2>What the free shell should make obvious</h2>
  <div class="shell-panel">
<pre><code>You're working on
Research to PRD handoff

Ready now
- Build-Readiness Check
- Handoff Brief

Still missing
- Product approval

Next step
Open Build-Readiness Check</code></pre>
  </div>

  <p>That is enough for a user to understand the value. Something useful is already ready, the blocker is visible, and the next move is obvious.</p>
  <span class="screen-reader-only">What a buyer should be able to verify quickly</span>
  <p class="page-lead">A buyer should be able to verify the front door, route choice, and draft posture quickly without reading a second operating manual.</p>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3>What to type</h3>
      <p>A first-time user should be able to see the front door immediately.</p>
    </div>
    <div class="checklist-card">
      <h3>What route Jini chose</h3>
      <p>The chosen provider, model, and route reason should be visible before the draft starts.</p>
    </div>
    <div class="checklist-card">
      <h3>Whether the result is still a draft</h3>
      <p>The user should know when it is still safe to review before anything is shared.</p>
    </div>
  </div>

  <p>If those signals are not obvious, the product still has work to do.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Renewal proof</span>
  <h2>What the paid layer should prove before renewal</h2>
  <p>The same proof should already be visible during the planned trial once checkout is live, before the paywall is presented and before the account is downgraded back to constrained free mode.</p>
  <div class="proof-grid">
    <div class="proof-card">
      <h3>Month-to-date savings</h3>
      <p>The paid layer should show that cheaper routing or compaction actually reduced spend, not just that the feature exists.</p>
    </div>
    <div class="proof-card">
      <h3>Provider headroom preserved</h3>
      <p>Users should be able to see when Jini kept work away from subscription ceilings, provider limits, or expensive emergency routes.</p>
    </div>
    <div class="proof-card">
      <h3>Throttles avoided or recovered</h3>
      <p>The story should include where fallback or recovery kept work moving after a hosted tool stalled, throttled, or became unavailable.</p>
    </div>
    <div class="proof-card">
      <h3>Sessions resumed without babysitting</h3>
      <p>The strongest paid proof is simple: the work kept moving and the user did not have to manually reconstruct state.</p>
    </div>
  </div>
</div>

<div class="section-card">
  <span class="section-kicker">Public trust posture</span>
  <h2>How public proof should be fed</h2>
  <p>The public site should ingest a sanitized proof snapshot, not hand-edited marketing claims. This page now reads from <code>docs/_data/public_proof.json</code>, which should be produced from the commercial proof bundle through a public-safe sync step.</p>
  <div class="proof-grid">
    {% for section in site.data.public_proof.sections %}
    <div class="proof-card">
      <h3>{{ section.headline }}</h3>
      <p>{{ section.bullets[0] }}</p>
    </div>
    {% endfor %}
  </div>
  <p class="page-lead">The trust story should stay inspectable: visible artifacts, visible route evidence, and no hidden cloud-memory magic.</p>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>What it stores</h3>
      <ul class="compact-list">
        <li>artifacts you can open and review</li>
        <li>work state needed for resume and status</li>
        <li>route evidence and local benchmark summaries when available</li>
        <li>repo-local setup state under <code>.jini</code> when setup needs to stay simple</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>What it should not hide</h3>
      <ul class="compact-list">
        <li>secret values in doctor or docs output</li>
        <li>hidden cloud memory the user cannot inspect</li>
        <li>route decisions without visible reasons or metrics</li>
        <li>automatic send/share behavior that skips review</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Why this matters</h3>
      <p>Buyability depends on trust. Jini should feel inspectable, reviewable, and cheap to understand before it asks a team to trust routing and stored state.</p>
    </div>
  </div>
  <p class="page-lead">What keeps proof honest is simple: useful work before summaries, inspectability instead of product magic, and review before send.</p>
  <ul class="compact-list">
    {% for rule in site.data.public_proof.trust_rules %}
    <li>{{ rule }}</li>
    {% endfor %}
  </ul>
</div>
