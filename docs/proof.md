---
title: Trust
description: Free value should be obvious quickly. Paid value should be visible as savings, recovered work, and route evidence before anyone is asked to renew.
---

<p class="page-lead">Proof has two jobs. First, the free shell should make its value obvious fast. Second, any paid layer should prove that it saved money or prevented stalled work before anyone is asked to renew.</p>

<div class="section-card">
  <span class="section-kicker">Free value</span>
  <h2>What the free shell should make obvious</h2>
  <div class="shell-panel">
<pre>You're working on
Research to PRD handoff

Ready now
- Build-Readiness Check
- Handoff Brief

Still missing
- Product approval

Next step
Open Build-Readiness Check</pre>
  </div>

  <p>That is enough for a user to understand the value. Something useful is already ready, the blocker is visible, and the next move is obvious.</p>
</div>

<div class="section-card">
  <span class="section-kicker">First-minute trust</span>
  <h2>What a buyer should be able to verify quickly</h2>
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
  <span class="section-kicker">Storage boundary</span>
  <h2>What Jini stores and what it does not</h2>
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
</div>

<div class="section-card">
  <span class="section-kicker">Trust posture</span>
  <h2>What should make Jini feel safer than plain chat</h2>
  <div class="proof-grid">
    <div class="proof-card">
      <h3>Deliverables before summaries</h3>
      <p>The first useful thing should be a follow-up, memo, checklist, or readiness check, not a status recap.</p>
    </div>
    <div class="proof-card">
      <h3>Inspectability instead of product magic</h3>
      <p>Users should be able to inspect artifacts, route evidence, and stored state instead of being asked to trust an invisible memory layer.</p>
    </div>
    <div class="proof-card">
      <h3>Review before send</h3>
      <p>The work should stay reviewable until the user deliberately shares or acts on it.</p>
    </div>
  </div>
</div>
