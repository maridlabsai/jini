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
  <p>When there is no repo context and no existing Jini work, the fallback should stay just as calm: interactive <code>jini</code> should still open the live shell. If the user gives a task there, Jini can answer briefly with <code>Run this from the repo or folder that needs work.</code>, but it should stay open instead of exiting. Non-interactive fallback output should stay just as calm and avoid banners, start cards, mini catalogs, diagnostics, or adoption guidance.</p>
  <p>That repo-aware start surface should stay light: skip the generic empty-state sentence, show one calm repo-context line, let the direct task suggestions stand on their own, use a soft cue like <code>Useful here:</code> above one or two useful commands Jini found in the repo, and keep at most one quiet adoption hint for existing Jini work.</p>
  <p>Internal diagnostics like <code>repo-map</code> and setup surfaces like <code>doctor</code> should stay off the first screen. A Claude Code, Codex, or GitHub CLI user should see task suggestions first, then at most a small <code>Already have current work?</code> note with <code>jini status</code> when existing Jini work actually needs attention.</p>
  <p>The three starter suggestions should also stay brief and action-first: <code>Review the repo and suggest the next move.</code>, <code>Fix the failing tests in this repo.</code>, and <code>Review the current branch and call out risks.</code></p>
  <p>Inside a repo, the first suggestions should be concrete asks like <code>jini review this repo</code>, <code>jini fix failing tests</code>, and <code>jini review this branch</code>, not examples and setup trivia.</p>
  <p>In a real terminal, bare <code>jini</code> should not print a launcher card and exit. It should stay open and accept the first task line immediately.</p>
  <p>That same calm shell shape should appear when Jini resumes remembered work too, so zero-arg <code>jini</code> still feels like one shell instead of one prompt for fresh repo work and another prompt for active work.</p>
  <p>For parity with Claude Code, Codex, and GitHub CLI expectations, the shell should drop the version banner, repo receipt, active-work receipt, and startup coaching line. It should skip the full outcome report before the prompt and open directly at <code>jini&gt;</code>.</p>
  <p>After the first answer it should keep the <code>jini&gt;</code> prompt open, let you type follow-up turns, and keep the controls in the background instead of teaching them before the task starts. If you need them, <code>commands</code>, <code>doctor</code>, <code>admin help</code>, and <code>exit</code> should still work inside the same live session, but they should answer with concise in-shell summaries instead of relaunching the full catalog or operator cards. Habitual prefixed input like <code>jini commands</code> or <code>jini status</code> typed inside the shell should still recover cleanly instead of turning into fake tasks.</p>
  <p>Once the session is already in flow, Jini should only print one short steering line when it is genuinely pointing you toward a concrete action. Otherwise it should acknowledge the task and return to the prompt.</p>

<pre><code class="language-bash">$ jini
jini&gt; fix failing tests

Working on: fix failing tests
Start with `make test`.
jini&gt; doctor

DOCTOR   Local preview [ok]</code></pre>
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
      <p>Shows route cost and continuity. Use <code>jini route list</code>, <code>jini route set codex</code>, or <code>jini route auto</code> when you need explicit framework switching.</p>
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
