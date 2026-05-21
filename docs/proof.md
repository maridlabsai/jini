---
title: Trust
description: Trust comes from usable outputs, visible missing pieces, and measurable route evidence.
---

<p class="page-lead">The proof is not a theory. It is whether Jini can turn something messy into something usable while keeping the remaining risk visible.</p>

<div class="section-card">
  <span class="section-kicker">The core test</span>
  <h2>What proof should look like</h2>
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
  <span class="section-kicker">User trust</span>
  <h2>What must also be obvious</h2>
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
