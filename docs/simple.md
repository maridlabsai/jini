---
title: Quickstart
description: The shortest explanation of what Jini is for and what to type first.
eyebrow: First useful run
context_line: This page is the shortest path from “what is Jini for?” to “I can tell whether this helps my real work.” It should feel like product motion, not setup ceremony.
highlights:
  - One front door
  - Artifact first
  - After the meeting
  - Before the handoff
quick_links:
  - label: Install
    href: /install.html
  - label: Examples
    href: /examples.html
  - label: Outputs
    href: /state-and-artifacts.html
---

<p class="page-lead">Jini is for the awkward middle of work: after the meeting, before the handoff, before the recommendation, and before calling something done. It should reduce stress, not add process.</p>

<p class="page-lead"><code>jini</code> is the front door. Describe the task, then use <code>jini open</code>, <code>jini continue</code>, or <code>jini status</code> when Jini points you there.</p>

<div class="section-card">
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
      <h3>Describe the task</h3>
      <p>Give Jini the messy notes, rough ask, screenshot, or draft you already have.</p>
    </div>
  </div>

<pre><code class="language-bash">curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
jini</code></pre>

<pre><code class="language-text">Jini
Describe the task.
Type `help` for examples and commands.
&gt; what is the capital of france
Paris.

&gt; add a line saying "hello from Jini" in the pear fellow script .txt file in this folder
Updated pear fellow script.txt
- Added line: hello from Jini
- Location: /path/to/pear fellow script.txt</code></pre>

  <p>If setup is missing, Jini should say so in the shell. Then type <code>auto</code>. If your company needs one strict route, use the matching setup path on the <a href="{{ '/install.html' | relative_url }}">Install</a> page instead.</p>

  <p><code>jini</code> is the front door. If you want the small public command list before you start, run <code>jini commands</code>. If you maintain routes, bundles, or release plumbing, the deeper inventory lives under <code>jini admin help</code>. The public support list should stay short: <code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code>.</p>
</div>

<div class="section-card">
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

<div class="section-card">
  <span class="section-kicker">Inside Jini</span>
  <h2>What the smallest flow should feel like</h2>
  <div class="pill-list">
    <span>Paste messy notes</span>
    <span>Ask for the outcome you want</span>
    <span>auto only if setup is missing</span>
    <span>jini open</span>
    <span>jini continue</span>
    <span>jini status</span>
    <span>route only when it matters</span>
  </div>

  <p>Jini should keep route, model, effort level, and reason inspectable without making them the first thing a new user has to learn. When work needs structure, a plain ask like <code>plan this change</code> should trigger that path.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Route setup</span>
  <h2>Only configure a route when Jini asks</h2>

  <p>Most first runs should not start with provider setup. If Jini needs help, it should ask for the missing route, key, or local model endpoint and keep the task visible.</p>
  <p>If you do not know how to begin, type <code>help me finish this</code> and then say something plain like <code>turn these meeting notes into a follow-up I can send</code>.</p>
</div>

<div class="section-card">
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

<div class="section-card section-card-cta">
  <span class="section-kicker">Next step</span>
  <h2>If you only remember one rule</h2>
  <p>Jini should make the work easier to finish. If it makes you think harder about the tool than about the work, it is failing.</p>
  <div class="page-intro-links">
    <a href="{{ '/examples.html' | relative_url }}">Examples</a>
    <a href="{{ '/install.html' | relative_url }}">Install</a>
    <a href="{{ '/cli.html' | relative_url }}">Command Catalog</a>
  </div>
</div>
