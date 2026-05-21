---
title: CLI Guide
description: The small public command surface and the setup check that sits behind it.
---

<p class="page-lead">Jini should feel small. Most people should learn one command, not a command tree.</p>

<p>You do not need to know Python to use Jini.</p>

<div class="section-card" markdown="1">
  <span class="section-kicker">Front door</span>
  <h2>Start here</h2>

```bash
jini
```

  <p>That should be the normal entry. Jini should either continue the thing you were already working on, show <code>Active work</code> when several projects are in flight, or offer simple choices if nothing is active yet.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Inside Jini</span>
  <h2>Main in-shell actions</h2>
  <div class="pill-list">
    <span>Continue</span>
    <span>Open</span>
    <span>Missing</span>
    <span>Plan</span>
    <span>Switch</span>
    <span>Start</span>
  </div>

  <p><code>Plan</code> is the structured mode. Use it when the work is still fuzzy and you want Jini to turn it into goal, requirements, design, steps, and run.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Decision card</span>
  <h2>What Jini should show before new work starts</h2>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>Route context</h3>
      <ul class="compact-list">
        <li>Tool</li>
        <li>Provider</li>
        <li>How chosen</li>
        <li>Continuity when Jini stays on the current coding route</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Model context</h3>
      <ul class="compact-list">
        <li>Model</li>
        <li>Effort level</li>
        <li>Why this route</li>
        <li>Verification depth when that matters</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Multimodal context</h3>
      <p>For screenshot, scanned PDF, and audio work, Jini should make the subtype-specific learning and route reason explicit instead of using one generic multimodal explanation.</p>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Scriptable surface</span>
  <h2>Commands that stay available for automation and power users</h2>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3><code>jini setup</code></h3>
      <p>Materializes one safe starter setup for a harness when you need an explicit local setup path.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini status</code></h3>
      <p>Shows a calm work summary: what you are working on, what is ready, what is missing, and the next step.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini open</code></h3>
      <p>Opens useful outputs like a follow-up, memo, or check instead of internal file paths.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini doctor</code></h3>
      <p>Local setup check for Claude, Bedrock, Azure OpenAI, Local SLM, or local preview.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini metrics</code></h3>
      <p>Shows the lean-platform command count, command timings, active provider, and measured route cost and trend evidence when available.</p>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Setup check</span>
  <h2>What <code>doctor</code> is for</h2>

```bash
jini doctor
jini
```

  <p>Most people should still start by pasting the work they want finished. Use doctor when setup help is needed, when you need one strict route, or when you are debugging access.</p>

  <div class="pill-list">
    <span>Auto</span>
    <span>Claude</span>
    <span>Bedrock</span>
    <span>Azure</span>
    <span>Local</span>
  </div>

  <p>Doctor reports what Jini will use, what auto mode resolved to, and what is missing. It does not print secret values.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Local SLM</span>
  <h2>What extra context doctor shows for local routing</h2>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>Machine view</h3>
      <ul class="compact-list">
        <li><code>DEVICE_CLASS</code></li>
        <li><code>DEVICE_OS</code></li>
        <li><code>LOCAL_ACCELERATOR</code></li>
        <li><code>LOCAL_RUNTIME_CLASS</code></li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Learning buckets</h3>
      <p>Jini keeps separate learning for screenshot work, scanned PDF work, and audio/transcript work.</p>
    </div>
    <div class="signal-card">
      <h3>Measured signals</h3>
      <p>When the local endpoint is ready, doctor can refresh benchmark summaries for load success, warm latency, cold start cost, token throughput, and structured-output reliability.</p>
    </div>
  </div>
</div>

<div class="section-card section-card-cta" markdown="1">
  <h2>The public rule</h2>
  <p>The normal path is still small: install once, run <code>jini</code>, paste the work you want finished. The rest of the CLI exists to support that experience, not replace it.</p>
  <div class="page-intro-links">
    <a href="./install.html">Install</a>
    <a href="./simple.html">Simple Guide</a>
    <a href="./state-and-artifacts.html">What Jini Shows</a>
  </div>
</div>
