---
title: Jini
description: One shell for AI work. Cheap by default, strong when needed.
---

<div class="hero-panel hero-panel-marketing">
  <p class="hero-kicker">One shell for AI work</p>
  <h1 class="hero-title">Cheap by default. Strong when needed.</h1>
  <p class="hero-summary">Jini gives teams one stable front door for messy work. It keeps progress, artifacts, and the next move visible while it chooses the cheapest suitable route by default.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html' | relative_url }}">Install</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Quickstart</a>
    <a class="cta-button cta-button-secondary" href="{{ '/examples.html' | relative_url }}">Examples</a>
  </div>
  <div class="command-block">
    <span class="command-label">Install once</span>
    <code>curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash</code>
    <code>jini</code>
  </div>
  <div class="compat-row" aria-label="Supported routes">
    <span class="compat-pill">Works with Claude Code</span>
    <span class="compat-pill">Works with Codex</span>
    <span class="compat-pill">Works with Bedrock</span>
    <span class="compat-pill">Works with Azure OpenAI</span>
    <span class="compat-pill">Works with Local models</span>
  </div>
</div>

**In plain words:** Jini helps you get from messy input to a usable output without making you babysit models, prompts, or route choices.

<div class="fact-strip">
  <div class="fact-pill">
    <strong>One shell</strong>
    <p>Install once, then start with <code>jini</code>.</p>
  </div>
  <div class="fact-pill">
    <strong>Cheap by default</strong>
    <p>Routine work stays inexpensive. Escalation happens only when the work justifies it.</p>
  </div>
  <div class="fact-pill">
    <strong>Visible outputs</strong>
    <p>See what is ready, what is missing, and what to open next without status hunting.</p>
  </div>
  <div class="fact-pill">
    <strong>Measured routing</strong>
    <p>Route choice, latency, and local benchmark evidence stay visible instead of hidden behind marketing claims.</p>
  </div>
</div>

## Quickstart

<div class="steps-grid">
  <div class="step-card">
    <span class="step-number">1</span>
    <h3>Install</h3>
    <p>Run the installer once from any terminal.</p>
  </div>
  <div class="step-card">
    <span class="step-number">2</span>
    <h3>Run <code>jini</code></h3>
    <p>That is the front door. You should not need a command tree for normal work.</p>
  </div>
  <div class="step-card">
    <span class="step-number">3</span>
    <h3>Paste the work</h3>
    <p>Start from the notes, screenshot, draft, transcript, or rough ask you already have.</p>
  </div>
</div>

## Why Jini

<div class="proof-grid">
  <div class="proof-card">
    <h3>One stable front door</h3>
    <p>Users learn one shell and one command vocabulary while providers, models, and tools can change underneath.</p>
  </div>
  <div class="proof-card">
    <h3>Cheaper by default</h3>
    <p>Jini keeps routine work on the cheapest suitable route and escalates only when depth, verification, or policy requires it.</p>
  </div>
  <div class="proof-card">
    <h3>Resumable work</h3>
    <p>Artifacts, missing items, and the next step stay attached to the same work object instead of disappearing into chat history.</p>
  </div>
  <div class="proof-card">
    <h3>Outputs you can use</h3>
    <p>The first useful result should be a follow-up, memo, readiness check, or itinerary, not a prettier status wall.</p>
  </div>
</div>

## What Jini writes

<div class="workflow-grid">
  <div class="workflow-card">
    <span class="workflow-meta">Deliverables</span>
    <h3>Usable artifacts</h3>
    <p>Jini writes the thing you actually need next: a follow-up, memo, readiness check, task list, or recommendation.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">State</span>
    <h3>Work you can resume</h3>
    <p>It keeps the active work, the focused artifact, and the next step attached to the same thread so continuation is cheap.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">Evidence</span>
    <h3>Route evidence</h3>
    <p>Jini exposes command latency, provider state, route cost, and route trend instead of asking you to trust the routing story blindly.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">Readiness</span>
    <h3>What is still missing</h3>
    <p>You can see what is ready now, what still blocks safe handoff, and why the missing parts still matter.</p>
  </div>
</div>

## Commands that matter

<div class="checklist-grid">
  <div class="checklist-card">
    <h3><code>jini</code></h3>
    <p>The normal entry point for interactive work.</p>
  </div>
  <div class="checklist-card">
    <h3><code>jini setup</code></h3>
    <p>Materialize one safe local setup path when explicit setup is needed.</p>
  </div>
  <div class="checklist-card">
    <h3><code>jini doctor</code></h3>
    <p>Check route setup without exposing secrets.</p>
  </div>
  <div class="checklist-card">
    <h3><code>jini status</code></h3>
    <p>Show what is active, ready, and still missing.</p>
  </div>
  <div class="checklist-card">
    <h3><code>jini open</code></h3>
    <p>Open the useful output instead of digging for file paths.</p>
  </div>
  <div class="checklist-card">
    <h3><code>jini metrics</code></h3>
    <p>See command count, route evidence, and measured latency/cost signals.</p>
  </div>
</div>

## Why trust it

<div class="media-grid">
  <div class="proof-card">
    <h3>Plain files and visible state</h3>
    <p>Jini keeps artifacts, work state, and route evidence legible instead of hiding everything in a proprietary cloud memory layer.</p>
  </div>
  <div class="proof-card">
    <h3>Cost and latency are measurable</h3>
    <p><code>jini metrics</code> reports real command timings and route evidence so efficiency is inspectable, not implied.</p>
  </div>
  <div class="proof-card">
    <h3>Strict routes still exist</h3>
    <p>If policy requires Claude, Bedrock, Azure, or local-only routing, you can pin that route without losing the same product surface.</p>
  </div>
  <div class="proof-card">
    <h3>Safe before send</h3>
    <p>The work stays reviewable. Jini keeps what is missing and what is uncertain visible before anything is treated as final.</p>
  </div>
</div>

## What Jini stores

<div class="checklist-grid">
  <div class="checklist-card">
    <h3>Stored locally</h3>
    <ul class="compact-list">
      <li>artifacts like follow-ups, memos, and readiness checks</li>
      <li>work state so the same thread can resume cleanly</li>
      <li>route evidence like command timings and local benchmark summaries</li>
      <li>repo-local setup state under <code>.jini</code> when setup needs to stay simple</li>
    </ul>
  </div>
  <div class="checklist-card">
    <h3>Not stored as product magic</h3>
    <ul class="compact-list">
      <li>hidden cloud memory you cannot inspect</li>
      <li>opaque routing claims without measurable evidence</li>
      <li>auto-share behavior that skips review</li>
      <li>secret values printed back into docs or doctor output</li>
    </ul>
  </div>
</div>

## See real outputs

<div class="media-grid">
  <a class="media-card" href="{{ '/examples.html#meeting-followup' | relative_url }}">
    <img src="{{ '/assets/examples/meeting-followup.gif' | relative_url }}" alt="Jini turning meeting notes into a sendable follow-up">
    <div class="media-copy">
      <span class="workflow-meta">After a meeting</span>
      <h3>Turn scattered notes into something you can send.</h3>
      <p>Get a sendable follow-up, owners, and open questions without rebuilding the meeting later.</p>
    </div>
  </a>
  <a class="media-card" href="{{ '/examples.html#spec-readiness' | relative_url }}">
    <img src="{{ '/assets/examples/research-prd.gif' | relative_url }}" alt="Jini checking whether a plan is ready to hand off">
    <div class="media-copy">
      <span class="workflow-meta">Before a handoff</span>
      <h3>Check whether a plan is actually ready.</h3>
      <p>See what is safe, what is missing, and what still blocks a real handoff.</p>
    </div>
  </a>
  <a class="media-card" href="{{ '/examples.html#vendor-choice' | relative_url }}">
    <img src="{{ '/assets/examples/vendor-selection.gif' | relative_url }}" alt="Jini preparing a recommendation memo from a vendor comparison">
    <div class="media-copy">
      <span class="workflow-meta">Before a decision</span>
      <h3>Keep the reasoning attached to the recommendation.</h3>
      <p>Get an output you can explain later instead of a conclusion you have to reconstruct from memory.</p>
    </div>
  </a>
</div>

<div class="section-card section-card-cta">
  <h2>Start with the one command that matters</h2>
  <p>Install Jini, run <code>jini</code>, and paste the work you want finished. Learn the rest only when the work actually needs it.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html' | relative_url }}">Install</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Quickstart</a>
    <a class="cta-button cta-button-secondary" href="{{ '/examples.html' | relative_url }}">Examples</a>
  </div>
</div>
