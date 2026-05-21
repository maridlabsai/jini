---
title: Install
description: Install Jini once, then use `jini`.
---

<p class="page-lead">The normal path should be simple: install once, run <code>jini</code>, paste the work you want finished.</p>

<div class="section-card" markdown="1">
  <span class="section-kicker">Recommended</span>
  <h2>Install once from any terminal</h2>

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
```

  <div class="pill-list">
    <span>macOS</span>
    <span>Linux</span>
    <span>HTTPS enforced</span>
    <span>User-space install</span>
  </div>

  <p>The installer first tries to install a matching release binary. If needed, it falls back to a source-backed runtime, verifies that the command launches, and prints one PATH fix only when your shell still cannot see <code>jini</code>.</p>

```bash
jini
```

  <p>If the installer prints a PATH line, run it once in the current shell and add it to your shell profile later.</p>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">First run</span>
  <h2>What should happen next</h2>
  <div class="steps-grid">
    <div class="step-card">
      <span class="step-number">1</span>
      <h3>Run <code>jini</code></h3>
      <p>Do not start with provider jargon unless Jini tells you setup is missing.</p>
    </div>
    <div class="step-card">
      <span class="step-number">2</span>
      <h3>Paste the work</h3>
      <p>Start from the notes, draft, screenshot, transcript, or rough ask you already have.</p>
    </div>
    <div class="step-card">
      <span class="step-number">3</span>
      <h3>Auto if needed</h3>
      <p>If setup is missing, type <code>Auto</code> and let Jini help you connect the best available route.</p>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Choose your path</span>
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
</div>

<div class="section-card" markdown="1">
  <span class="section-kicker">Copy-paste paths</span>
  <h2>Strict route setup</h2>

  <div class="scenario-grid">
    <div class="scenario-card" markdown="1">
      <h3>Claude direct</h3>
```bash
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
jini doctor
jini
```
      <p>Use this when your team already gave you a direct Anthropic key and Claude should be the fixed route.</p>
    </div>

    <div class="scenario-card" markdown="1">
      <h3>Amazon Bedrock</h3>
```bash
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
jini doctor
jini
```
      <p>Use this when AWS policy or Bedrock access already exists. If you know the exact model id, you can set <code>BEDROCK_MODEL_ID</code> instead.</p>
    </div>

    <div class="scenario-card" markdown="1">
      <h3>Azure OpenAI</h3>
```bash
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
jini doctor
jini
```
      <p>Use this when your company requires Azure OpenAI only. On Azure, the deployment decides the actual model.</p>
    </div>

    <div class="scenario-card" markdown="1">
      <h3>Auto route</h3>
```bash
export JINI_TOOL=auto
export JINI_PROVIDER=auto
export JINI_MODEL=auto
jini doctor
jini
```
      <p>Use this if you want Jini to choose the cheapest suitable route, model, and effort level for each request.</p>
    </div>

    <div class="scenario-card" markdown="1">
      <h3>Local SLM</h3>
```bash
export JINI_PROVIDER=local-slm
export JINI_TOOL=auto
export JINI_MODEL=auto
export JINI_LOCAL_SLM_ENDPOINT="http://127.0.0.1:11434/v1"
export JINI_LOCAL_SLM_MODEL="qwen3:8b"
jini doctor
jini
```
      <p>Use this when you already run an OpenAI-compatible local model endpoint and want local workhorse routing.</p>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
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

<div class="section-card section-card-cta" markdown="1">
  <h2>Most people should stop here</h2>
  <p>Install Jini, run <code>jini</code>, and paste the work you want finished. Only drop to the strict route blocks when policy or debugging requires it.</p>
  <div class="page-intro-links">
    <a href="./simple.html">Simple Guide</a>
    <a href="./cli.html">CLI Guide</a>
    <a href="./examples.html">Examples</a>
  </div>
</div>
