---
title: Command Catalog
description: The small public command surface and the setup check that sits behind it.
eyebrow: Keep the shell small
context_line: "The command story should stay simple: one front door for most people, one calm catalog when needed, and deeper plumbing only for operators who actually maintain routes."
highlights:
  - "Start with `jini`"
  - "`jini commands`"
  - Five support commands only
  - Admin help for operators
quick_links:
  - label: Install
    href: /install.html
  - label: Quickstart
    href: /simple.html
  - label: Examples
    href: /examples.html
---

<p class="page-lead">Jini should feel small. Most people should learn one command, not a command tree.</p>

<p>Jini runs as a native command-line app; the normal user experience is one small <code>jini</code> front door.</p>

<div class="section-card">
  <span class="section-kicker">Front door</span>
  <h2>Start here</h2>

<pre><code class="language-bash">jini</code></pre>

  <p>That should be the normal entry. Jini should always start task-first; saved work should stay behind explicit inspection commands or natural title matching.</p>
  <p>The shipped <code>v0.1.2</code> start shape is compact and literal:</p>

<pre><code class="language-bash">$ jini
Jini
&gt; what is the capital of france
Paris.

&gt; add a line saying "hello from Jini" in notes.txt
Updated notes.txt
- Added line: hello from Jini
- Location: /path/to/notes.txt</code></pre>

  <p>If a file edit is ambiguous, Jini should stop and list candidate filenames instead of guessing. Setup, route, and status commands should support that flow, not appear before the task.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Command catalog</span>
  <h2>See the public command surface without the internal inventory</h2>

<pre><code class="language-bash">jini commands</code></pre>

  <p>Use this when you want the small product-facing command catalog. If you maintain routes, bundles, or release plumbing, use <code>jini admin help</code> instead of treating the full internal parser as the public surface.</p>
  <p>A professional platform does not teach every command it happens to have. The public catalog should stay narrow: <code>jini</code>, then <code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code> when the live session actually needs them.</p>
  <p>Help surfaces and <code>jini commands</code> are catalogs, not request entrypoints. If you type <code>jini help me edit notes.txt</code>, <code>jini commands me edit notes.txt</code>, <code>jini --help me edit notes.txt</code>, or <code>jini provider help me edit notes.txt</code>, Jini should reject that tail text and tell you to start with <code>jini</code> for the start surface instead.</p>

<pre><code class="language-bash">$ jini help me edit pear fellow script.txt
ERROR `jini help` shows the CLI overview; it does not take a request like "me edit pear fellow script.txt".
Start with `jini` to resume active work or see the start options.
If you already have current work, use `jini status` once.</code></pre>

  <p>That corrective output is the expected terminal shape for a help request tail. It should reject the request and redirect people to the start surface instead of dumping the overview.</p>
  <p>Direct task intake should still work on the normal front door. Inside a repo, <code>jini fix failing tests</code> and <code>jini review this repo</code> should use the same calm shell language instead of an argparse usage wall or a separate intake report.</p>
  <p>Outside a repo, the same direct task ask should stay concise too: acknowledge the task, then tell the user to run it from the repo or folder that needs work. It should not fall back to startup cards, examples, or setup teaching.</p>
  <p>The same redirect shape applies across the other help-entry variants too:</p>
  <ul class="compact-list">
    <li><code>jini commands me edit pear fellow script.txt</code> starts with <code>ERROR `jini commands` shows the public command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini --help me edit pear fellow script.txt</code> starts with <code>ERROR `jini --help` shows the CLI overview; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini admin help me edit pear fellow script.txt</code> starts with <code>ERROR `jini admin help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini provider help me edit pear fellow script.txt</code> starts with <code>ERROR `jini provider help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini provider --help me edit pear fellow script.txt</code> starts with <code>ERROR `jini provider --help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
  </ul>
  <p>Only the first line changes. The redirect lines stay the same: start with <code>jini</code> to resume active work or see the start options, then use <code>jini status</code> once if you already have current work.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Inside Jini</span>
  <h2>Keep typing the next thing</h2>
  <div class="pill-list">
    <span>fix failing tests</span>
    <span>open the latest artifact</span>
    <span>what is blocked?</span>
    <span>plan this change</span>
    <span>resume the other repo review</span>
    <span>start from these notes</span>
  </div>

  <p>Daily use should feel like one calm conversation. You should not have to memorize product words like <code>Missing</code> or saved-work controls before Jini becomes useful.</p>
  <p>If the work is fuzzy, plain asks like <code>plan this change</code> should still trigger the structured path without forcing you to translate into Jini-specific vocabulary first.</p>
</div>

<div class="section-card">
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

<div class="section-card">
  <span class="section-kicker">Catalog entries</span>
  <h2>Support commands when Jini points you there</h2>
  <p>These are the only support commands that still deserve public teaching. They support the product front door, and most people should reach them because Jini asked for one of them, not because they are navigating a command tree by hand.</p>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3><code>jini status</code></h3>
      <p>Shows a calm work summary when you want the current work state outside the shell.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini continue</code></h3>
      <p>Shows the next useful artifact without rebuilding the current work state from scratch.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini open</code></h3>
      <p>Opens the richer artifact view when the shell preview is not enough.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini route</code></h3>
      <p>Shows route cost and continuity. Use <code>jini route list</code>, <code>jini route set azure-code</code>, or <code>jini route auto</code> when you need explicit provider/model switching. <code>codex</code> and <code>claude-code</code> are reserved for real installed-CLI handoff. Use <code>jini route dogfood</code> before release validation to see the Wave 1 CLI evidence checklist and copyable <code>.jini/cli-dogfood.json</code> template.</p>
    </div>
    <div class="checklist-card">
      <h3><code>jini doctor</code></h3>
      <p>Local setup check when Jini needs route help or you are debugging access.</p>
    </div>
  </div>
  <p>Commands like <code>try-example</code>, <code>get-started</code>, <code>show</code>, <code>expand</code>, <code>context</code>, <code>resume</code>, <code>metrics</code>, and <code>harnesses</code> should not sit in the public command story.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Setup check</span>
  <h2>What <code>doctor</code> is for</h2>

<pre><code class="language-bash">jini doctor
jini</code></pre>

  <p>Most people should still start by describing the task or pasting the rough notes they already have. Use doctor when setup help is needed, when you need one strict route, or when you are debugging access.</p>

  <div class="pill-list">
    <span>auto</span>
    <span>Claude</span>
    <span>Bedrock</span>
    <span>Azure</span>
    <span>Local</span>
  </div>

  <p>Doctor reports what Jini will use, what auto mode resolved to, and what is missing. It does not print secret values.</p>
</div>

<div class="section-card">
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

<div class="section-card section-card-cta">
  <span class="section-kicker">Next step</span>
  <h2>The public rule</h2>
  <p>The normal path is still small: install once, run <code>jini</code>, then describe the task or paste the rough notes you already have. When you need a public command list, use <code>jini commands</code>. Use <code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code> when Jini points you there; the rest of the CLI should stay out of the way.</p>
  <div class="page-intro-links">
    <a href="{{ '/install.html' | relative_url }}">Install</a>
    <a href="{{ '/simple.html' | relative_url }}">Quickstart</a>
    <a href="{{ '/state-and-artifacts.html' | relative_url }}">Outputs</a>
  </div>
</div>
