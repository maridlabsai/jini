---
title: Install
description: Install Jini once, then use `jini`.
eyebrow: Live install path
context_line: The CLI is the public installable surface today. The rest of this page exists to get you to a real first run fast, then show the honest surface boundary for everything else.
highlights:
  - Install once
  - "Run `jini`"
  - CLI live now
  - Desktop and mobile when live
quick_links:
  - label: Quickstart
    href: /simple.html
  - label: Examples
    href: /examples.html
  - label: Command Catalog
    href: /cli.html
---

<p class="page-lead">The normal path should be simple: install once, run <code>jini</code>, paste the work you want finished.</p>

<div class="section-card section-card-soft surface-story">
  <span class="section-kicker">Current availability</span>
  <h2>What you get today</h2>
  <p>The CLI is the live installable surface today. Desktop and mobile are the next release surfaces, stay free to download when live, and are not publicly downloadable yet. Desktop and Android should distribute directly first where policy allows, while iOS remains App Store constrained. Commercial pricing is planned to start with a 30-day free trial and become $1/month once checkout and entitlement activation are live.</p>

  <!-- | {{ surface.name }} | {{ surface.current_state }} | {{ surface.next_step }} | -->
  <p>In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both.</p>

  <table>
    <thead>
      <tr>
        <th scope="col">Surface</th>
        <th scope="col">Current state</th>
        <th scope="col">What to expect next</th>
      </tr>
    </thead>
    <tbody>
      {% for surface in site.data.public_surfaces.surfaces %}
      <tr>
        <th scope="row">{{ surface.name }}</th>
        <td>{{ surface.current_state }}</td>
        <td>{{ surface.next_step }}</td>
      </tr>
      {% endfor %}
    </tbody>
  </table>

  <p class="editorial-note">Install the CLI today. Treat this matrix as the source of truth for every other surface until the release packet says that surface is live. It is fed from <code>docs/_data/public_surfaces.json</code>, a sanitized snapshot built from the commercial release packets.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Recommended</span>
  <h2>Install once from any terminal</h2>

<pre><code class="language-bash">curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash</code></pre>

  <div class="pill-list">
    <span>macOS</span>
    <span>Linux</span>
    <span>HTTPS enforced</span>
    <span>User-space install</span>
  </div>

  <p>The installer first tries to install a matching release binary. If needed, it falls back to a source-backed runtime, verifies that the command launches, and prints one PATH fix only when your shell still cannot see <code>jini</code>.</p>

  <p>After install, Jini also prints one short provenance line for support and troubleshooting, such as <code>- install source: release binary</code>, <code>- install source: source runtime (explicit source)</code>, <code>- install source: source runtime (release-unavailable)</code>, or <code>- install source: source fallback (release validation failed: unsupported-public-command-surface)</code>.</p>

  <p>The same install writes <code>install-receipt.txt</code> with the machine-readable fields <code>source_reason=</code> and <code>release_validation=</code>. When the install needs follow-up, the receipt also records <code>next_step=</code>.</p>
  <p>In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both.</p>
  <p>On source-path installs that need follow-up, the terminal now prints this exact support handoff line: <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>.</p>

  <p>When support asks for install details on a source-path install, send the support receipt path plus only these receipt keys. Treat <code>next_step=</code> as the actionable follow-up field for this source install.</p>
  <ul>
    <li><code>install-receipt.txt</code> path from the printed <code>- support receipt: ...</code> line</li>
    <li><code>version=</code></li>
    <li><code>source_reason=</code></li>
    <li><code>release_validation=</code></li>
    <li><code>next_step=</code></li>
  </ul>
  <p>If the install output shows only <code>- install source: release binary</code> followed by <code>jini</code>, support can ignore this checklist.</p>
  <p>If <code>next_step=</code> is missing on a source-path install, that source path likely completed without extra follow-up. The healthy <code>release binary</code> path does not use this source-path handoff at all.</p>

  <table>
    <thead>
      <tr>
        <th scope="col">Install source line</th>
        <th scope="col">What it usually means</th>
        <th scope="col">What to do next</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <th scope="row"><code>- install source: release binary</code></th>
        <td>Matching published release asset was available, passed Jini's public command check, and was accepted without any source fallback.</td>
        <td>No action needed unless you expected a source install for local development. The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path, and a missing <code>next_step=</code> field is only meaningful on source-path installs.</td>
      </tr>
      <tr>
        <th scope="row"><code>- install source: source runtime (explicit source)</code></th>
        <td>You pointed the installer at a local checkout on purpose.</td>
        <td>Keep using that checkout and rerun from the repo when you want the local source tree to stay in control. If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>. Treat <code>next_step=</code> as the actionable follow-up field for this source install.</td>
      </tr>
      <tr>
        <th scope="row"><code>- install source: source runtime (release-unavailable)</code></th>
        <td>No matching release asset was available for this machine or release channel.</td>
        <td>If this machine should have had a published release, file a release issue and include the receipt. If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>. Treat <code>next_step=</code> as the actionable follow-up field for this source install.</td>
      </tr>
      <tr>
        <th scope="row"><code>- install source: source fallback (release validation failed: unsupported-public-command-surface)</code></th>
        <td>Downloaded release binary did not support the current public command surface.</td>
        <td>Keep the source install, attach install-receipt.txt, and flag the stale release artifact. If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>. Treat <code>next_step=</code> as the actionable follow-up field for this source install.</td>
      </tr>
    </tbody>
  </table>

  <p>Example release-binary success output:</p>
<pre><code class="language-text">- install source: release binary
jini
</code></pre>
  <p>The healthy release-binary path stops there.</p>
  <p>The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path, and a missing <code>next_step=</code> field is only meaningful on source-path installs.</p>
  <p>Unlike the source-path follow-up output below, <code>next_step=</code> appears only on source-path installs that need follow-up.</p>

  <p>Example source-path follow-up output:</p>
<pre><code class="language-text">- install source: source runtime (release-unavailable)
- support receipt: /Users/you/.local/bin/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)
- next step: If this machine should have had a published release, file a release issue and include the receipt.
</code></pre>
  <p>Treat <code>next_step=</code> as the actionable follow-up field for this source install.</p>

<pre><code class="language-bash">jini</code></pre>

  <p>If the installer prints a PATH line, run it once in the current shell and add it to your shell profile later.</p>
</div>

<div class="section-card">
  <span class="section-kicker">First run</span>
  <h2>What should happen next</h2>
  <div class="steps-grid">
    <div class="step-card">
      <span class="step-number">1</span>
      <h3>Run <code>jini</code></h3>
      <p>Do not start with provider jargon unless Jini tells you setup is missing. Inside a repo, Jini should immediately steer toward task-first asks like <code>jini review this repo</code> or <code>jini fix failing tests</code>.</p>
      <p>That repo-aware start surface should stay light: skip the generic empty-state sentence, show one calm repo-context line, let the direct task suggestions stand on their own, use a soft cue like <code>Useful here:</code> above one or two useful commands Jini found in the repo, and keep at most one quiet adoption hint for existing Jini work.</p>
      <p>Internal diagnostics like <code>repo-map</code> and setup surfaces like <code>doctor</code> should stay off the first screen. A Claude Code, Codex, or GitHub CLI user should see task suggestions first, then at most a small <code>Already have Jini work?</code> note with <code>status /path/to/work</code> when an existing Jini work path actually needs adoption.</p>
      <p>The three starter suggestions should stay brief and action-first: <code>Review the repo and suggest the next move.</code>, <code>Fix the failing tests in this repo.</code>, and <code>Review the current branch and call out risks.</code></p>
      <p>In a real terminal, bare <code>jini</code> should stay open, show one calm repo-context line, ask <code>What do you want Jini to do?</code>, and let you type the first task directly at the <code>jini&gt;</code> prompt.</p>
      <p>That same calm task cue should appear when Jini resumes remembered work too, so zero-arg <code>jini</code> still feels like one shell instead of one prompt for fresh repo work and another prompt for active work.</p>
      <p>That prompt should remain open for follow-up turns. The task should stay primary, and the controls should stay in the background until you need them. If you do, <code>commands</code>, <code>doctor</code>, <code>help --admin</code>, and <code>exit</code> should still work as in-session escape hatches instead of forcing a relaunch.</p>
      <p>Once the session is already in flow, Jini should only print a <code>NEXT</code> line when it is actually steering you to a concrete action. Otherwise it should acknowledge the task and go straight back to the prompt.</p>
    </div>
    <div class="step-card">
      <span class="step-number">2</span>
      <h3>Paste the work</h3>
      <p>Start from the notes, draft, screenshot, transcript, or rough ask you already have.</p>
      <p>After the first answer, keep typing plain follow-up asks like <code>fix failing tests</code>, <code>what is blocked?</code>, or <code>open the latest artifact</code> instead of learning Jini-specific action words.</p>
    </div>
    <div class="step-card">
      <span class="step-number">3</span>
      <h3>Auto if needed</h3>
      <p>If setup is missing, type <code>Auto</code> and let Jini help you connect the best available route.</p>
    </div>
  </div>

  <p>If you want the small public command list before doing anything else, run <code>jini commands</code>. If you maintain routes, bundles, or release plumbing, the deeper inventory lives under <code>jini admin help</code>.</p>
  <p>That public list should stay deliberately short: <code>jini</code> first, then <code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code> only when the live session needs them.</p>
  <p>Direct task text belongs on the normal front door. Inside a repo, <code>jini review this repo</code>, <code>jini fix failing tests</code>, and <code>jini review this branch</code> should surface repo-aware intake instead of dropping into argparse usage output.</p>
  <p>Help surfaces and <code>jini commands</code> are catalogs, not request entrypoints. If you paste a work request after <code>help</code>, <code>--help</code>, or <code>commands</code>, Jini should reject that tail text and point you back to starting with <code>jini</code> for the start surface.</p>
  <ul class="compact-list">
    <li><code>jini commands me edit pear fellow script.txt</code> should start with <code>ERROR `jini commands` shows the public command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini --help me edit pear fellow script.txt</code> should start with <code>ERROR `jini --help` shows the CLI overview; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini admin help me edit pear fellow script.txt</code> should start with <code>ERROR `jini admin help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini provider help me edit pear fellow script.txt</code> should start with <code>ERROR `jini provider help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
    <li><code>jini provider --help me edit pear fellow script.txt</code> should start with <code>ERROR `jini provider --help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".</code></li>
  </ul>
  <p>Only the first line changes. The redirect stays the same: start with <code>jini</code> to resume active work or see the start options, then use <code>jini status /path/to/work</code> once if you already have work to adopt.</p>

<pre><code class="language-bash">$ jini --help me edit pear fellow script.txt
ERROR `jini --help` shows the CLI overview; it does not take a request like "me edit pear fellow script.txt".
Start with `jini` to resume active work or see the start options.
If you already have work to adopt, use `jini status /path/to/work` once.</code></pre>

  <p>Use that <code>--help</code> example when someone pastes work after the overview flag on first run. It should reject the request and send them back to the normal start path.</p>

<pre><code class="language-bash">$ jini provider help me edit pear fellow script.txt
ERROR `jini provider help` shows the admin command inventory; it does not take a request like "me edit pear fellow script.txt".
Start with `jini` to resume active work or see the start options.
If you already have work to adopt, use `jini status /path/to/work` once.</code></pre>

  <p>Use that provider-help example when someone drifts into the admin/provider tree during first run. It should reject the request and send them back to the normal start path.</p>
  <p>The contrast is simple: <code>--help</code> stays on the CLI-overview path, while <code>provider help</code> crosses into the admin-inventory path, even though both redirects send the user back to <code>jini</code>.</p>
</div>

<div class="section-card">
  <span class="section-kicker">Strict route fallback</span>
  <h2>Only use the strict setup blocks when you actually need them</h2>

  <div class="checklist-grid">
    <div class="checklist-card">
      <h3>Auto</h3>
      <p>Best for most people. Jini chooses the cheapest suitable route by default and escalates only when the request clearly needs deeper work.</p>
    </div>
    <div class="checklist-card">
      <h3>Claude</h3>
      <p>Best when you already use Anthropic directly and want Claude as the strict route.</p>
    </div>
    <div class="checklist-card">
      <h3>Bedrock</h3>
      <p>Best when AWS policy or existing Bedrock access decides the route.</p>
    </div>
    <div class="checklist-card">
      <h3>Azure</h3>
      <p>Best when company policy requires Azure OpenAI only.</p>
    </div>
    <div class="checklist-card">
      <h3>Local</h3>
      <p>Best when you already run an OpenAI-compatible local model server and want cheap-first local routing.</p>
    </div>
  </div>

  <p>The short rule: stay on the normal path unless policy or debugging forces a strict route. When it does, use one of these copy-paste setups and return to <code>jini</code>.</p>

  <div class="scenario-grid">
    <div class="scenario-card">
      <h3>Claude direct</h3>
<pre><code class="language-bash">
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
jini doctor
jini
</code></pre>
      <p>Use this when your team already gave you a direct Anthropic key and Claude should be the fixed route.</p>
    </div>

    <div class="scenario-card">
      <h3>Amazon Bedrock</h3>
<pre><code class="language-bash">
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
jini doctor
jini
</code></pre>
      <p>Use this when AWS policy or Bedrock access already exists. If you know the exact model id, you can set <code>BEDROCK_MODEL_ID</code> instead.</p>
    </div>

    <div class="scenario-card">
      <h3>Azure OpenAI</h3>
<pre><code class="language-bash">
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
jini doctor
jini
</code></pre>
      <p>Use this when your company requires Azure OpenAI only. On Azure, the deployment decides the actual model.</p>
    </div>

    <div class="scenario-card">
      <h3>Auto route</h3>
<pre><code class="language-bash">
export JINI_TOOL=auto
export JINI_PROVIDER=auto
export JINI_MODEL=auto
jini doctor
jini
</code></pre>
      <p>Use this if you want Jini to choose the cheapest suitable route, model, and effort level for each request.</p>
    </div>

    <div class="scenario-card">
      <h3>Local SLM</h3>
<pre><code class="language-bash">
export JINI_PROVIDER=local-slm
export JINI_TOOL=auto
export JINI_MODEL=auto
export JINI_LOCAL_SLM_ENDPOINT="http://127.0.0.1:11434/v1"
export JINI_LOCAL_SLM_MODEL="qwen3:8b"
jini doctor
jini
</code></pre>
      <p>Use this when you already run an OpenAI-compatible local model endpoint and want local workhorse routing.</p>
    </div>
  </div>
</div>

<div class="section-card">
  <span class="section-kicker">Setup check</span>
  <h2>What <code>jini doctor</code> actually tells you</h2>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>What it does</h3>
      <ul class="compact-list">
        <li>shows what Jini will use</li>
        <li>shows what auto mode resolved to</li>
        <li>shows what settings are still missing</li>
        <li>does not print secret values</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>What it does not prove</h3>
      <ul class="compact-list">
        <li>your AWS auth is valid</li>
        <li>your Bedrock model access is enabled</li>
        <li>your Azure key is accepted by the service</li>
        <li>your Claude account is allowed to use a model</li>
      </ul>
    </div>
    <div class="signal-card">
      <h3>Local SLM extras</h3>
      <p>When local routing is active, doctor also shows the machine and runtime view Jini is using, including device class, local runtime class, and accelerator context.</p>
    </div>
  </div>
</div>

<div class="section-card section-card-cta">
  <span class="section-kicker">Next step</span>
  <h2>Most people should stop here</h2>
  <p>Install Jini, run <code>jini</code>, and paste the work you want finished. Only drop to the strict route blocks when policy or debugging requires it.</p>
  <div class="page-intro-links">
    <a href="{{ '/simple.html' | relative_url }}">Quickstart</a>
    <a href="{{ '/cli.html' | relative_url }}">Command Catalog</a>
    <a href="{{ '/examples.html' | relative_url }}">Examples</a>
  </div>
</div>
