---
title: Simple Guide
description: The shortest explanation of what Jini is for and what to type first.
---

<p class="page-lead">Jini is for the awkward middle of work: after the meeting, before the handoff, before the recommendation, and before calling something done. It should reduce stress, not add process.</p>

<div class="section-card" markdown="1">
  <span class="section-kicker">Start here</span>
  <h2>The shortest first run</h2>
  <div class="steps-grid">
    <div class="step-card">
      <span class="step-number">1</span>
      <h3>Install Jini</h3>
      <p>Run the installer once from any terminal.</p>
    </div>
    <div class="step-card">
      <span class="step-number">2</span>
      <h3>Start with <code>jini</code></h3>
      <p>That is the front door. It should be enough for normal use.</p>
    </div>
    <div class="step-card">
      <span class="step-number">3</span>
      <h3>Paste the work</h3>
      <p>Give Jini the messy notes, rough ask, screenshot, or draft you already have.</p>
    </div>
  </div>

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
jini
```

  <p>If setup is missing, Jini should say so in the shell. Then type <code>Auto</code>. If your company needs one strict route, use the matching setup path on the <a href="./install.html">Install</a> page instead.</p>

  <p><code>jini</code> is the front door.</p>

`jini` is the front door.
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Use cases</span>
  <h2>Think about normal problems</h2>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3>After a meeting</h3>
      <p>The follow-up is fuzzy, owners are implied, and the sendable version does not exist yet.</p>
    </div>
    <div class="checklist-card">
      <h3>Before a handoff</h3>
      <p>The plan looks finished, but you are not sure it is safe to hand off or build from.</p>
    </div>
    <div class="checklist-card">
      <h3>Before a decision</h3>
      <p>The choice was made, but the reasoning, tradeoffs, or open questions are getting lost.</p>
    </div>
    <div class="checklist-card">
      <h3>Before real closure</h3>
      <p>The main pain stopped, but the actual aftercare work is still easy to skip.</p>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Inside Jini</span>
  <h2>What the smallest flow should feel like</h2>
  <div class="pill-list">
    <span>Paste messy notes</span>
    <span>Ask for the outcome you want</span>
    <span>Auto only if setup is missing</span>
    <span>Open</span>
    <span>Missing</span>
    <span>Plan</span>
    <span>Switch</span>
  </div>

  <p>Before Jini starts a new piece of work, it should show a short decision card with the route, model, effort level, and reason. When you choose <code>Plan</code>, it should slow down and structure the work into goal, requirements, design, steps, and run.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">One concrete path</span>
  <h2>If you use Claude</h2>

```text
Claude
```

  <p>Jini should ask only for the missing API key, save it in the repo-local <code>.jini</code> folder, and let you continue. If you do not know how to begin, type <code>help me finish this</code> and then say something plain like <code>turn these meeting notes into a follow-up I can send</code>.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Output quality</span>
  <h2>What good help looks like</h2>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>Visible context</h3>
      <ul class="compact-list">
        <li>the goal</li>
        <li>the working inputs</li>
        <li>the chosen route when it matters</li>
        <li>what else is already active</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Visible progress</h3>
      <ul class="compact-list">
        <li>what just finished</li>
        <li>what is happening now</li>
        <li>what will happen next</li>
        <li>what is already done</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Visible risk</h3>
      <ul class="compact-list">
        <li>what still needs attention</li>
        <li>why that missing thing matters</li>
        <li>what Jini is not sure about</li>
        <li>whether the result is still safe to review before sharing</li>
      </ul>
    </div>
  </div>

  <p>The things Jini gives back should feel like work you can use: a follow-up you can send, a recommendation you can explain, a closure checklist, a plan check, or a trip itinerary. If it only gives you a prettier status wall, it is failing.</p>
</div>

<div class="section-card section-card-cta" markdown="1">
  <h2>If you only remember one rule</h2>
  <p>Jini should make the work easier to finish. If it makes you think harder about the tool than about the work, it is failing.</p>
  <div class="page-intro-links">
    <a href="./examples.html">Examples</a>
    <a href="./install.html">Install</a>
    <a href="./cli.html">CLI Guide</a>
  </div>
</div>
