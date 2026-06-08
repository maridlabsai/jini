---
title: Jini
description: Free orchestration for AI work across providers today. Pay only when Jini can prove it saves money or keeps work moving.
---

**In plain words:** Jini is a CLI-first work router. It keeps the first minute small: ask a simple question, edit the named local file, or hand off larger AI work to the right route without losing context.

<div class="hero-panel hero-panel-marketing">
  <p class="hero-kicker">One small front door for local edits, simple answers, and routed AI work.</p>
  <h1 class="hero-title">Use AI CLIs without relearning the first minute.</h1>
  <p class="hero-category-claim">The open shell for route-aware AI work.</p>
  <p class="hero-summary">Jini starts as a compact CLI: describe the task, ask a question, or edit a named file in the current folder. Bigger work can still become a follow-up, readiness check, or decision memo when that is what you need.</p>
  <p class="hero-summary-support">The free CLI is live today. Desktop and mobile should carry the same work forward only after those surfaces pass the same trust checks.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html' | relative_url }}">Install Free</a>
    <a class="cta-button cta-button-secondary" href="{{ '/examples.html' | relative_url }}">See Examples</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Quickstart</a>
  </div>
  <div class="hero-scene">
    <div class="hero-scene-card hero-scene-card-live hero-scene-card-primary">
      <span class="screen-reader-only">Live now</span>
      <strong>CLI thread</strong>
      <p>Run <code>jini</code>, describe the task, and keep file edits or simple answers direct.</p>
    </div>
    <div class="hero-scene-card hero-scene-card-support">
      <div class="hero-scene-support-row">
        <span class="screen-reader-only">Next surfaces</span>
        <strong>Desktop and mobile continuity</strong>
        <p>Carry the same thread forward as each surface arrives.</p>
      </div>
      <div class="hero-scene-support-row hero-scene-support-row-output">
        <span class="screen-reader-only">What leaves the thread</span>
        <strong>Sendable output</strong>
        <p>Leave with a follow-up, readiness check, or decision memo someone else can use.</p>
      </div>
    </div>
  </div>
  <div class="hero-decision-frame hero-decision-frame-inline">
    <div class="hero-decision-intro">
      <span class="screen-reader-only">Choose the lightest layer that still leaves behind usable work.</span>
      <span class="screen-reader-only">If you only need a one-off answer, use a raw model shell.</span>
    </div>

    <div class="offer-grid">
      <div class="offer-card offer-card-core">
        <span class="offer-card-eyebrow offer-card-eyebrow-core">Free orchestration core</span>
        <h3>Use the free Jini shell when the work has to survive handoff.</h3>
        <div class="offer-card-contexts" aria-label="Flagship jobs">
          <span>Named file edits</span>
          <span>Simple questions</span>
          <span>Routed AI work</span>
        </div>
        <p class="offer-card-note offer-card-note-core">Jini earns the right to exist when the first minute stays obvious and the next route remains inspectable.</p>
      </div>
      <div class="offer-side-rail">
        <div class="offer-card offer-card-plain">
          <span class="offer-card-eyebrow">Fallback path</span>
          <h3>Use the raw shell for one-shot answers.</h3>
          <p class="offer-card-note">Fastest path when the answer can die in the tab.</p>
        </div>
        <div class="offer-card offer-card-paid">
          <span class="offer-card-eyebrow">Optional add-on</span>
          <h3>Add the paid optimizer only when the proof can be measured.</h3>
          <p class="offer-card-note offer-card-note-support">The paid layer stays narrow: it only enters when Jini can prove it saved money or kept work moving.</p>
          <p class="offer-card-note"><strong>Pay only for proof</strong>: when checkout is live, the planned 30-day free trial should come before the $1/month subscription and only after Jini can show savings, preserved headroom, or recovered work.</p>
        </div>
      </div>
    </div>

    <div class="hero-install-rail">
      <span class="command-label">Run once</span>
      <div class="hero-install-main">
        <code>curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash</code>
        <code>jini</code>
        <code>&gt; add a line saying "hello from Jini" in notes.txt</code>
      </div>
      <div class="compat-pill-row" aria-label="Supported routes">
        <span class="compat-pill">Works with Claude Code</span>
        <span class="compat-pill">Works with Codex</span>
        <span class="compat-pill">Works with Bedrock</span>
        <span class="compat-pill">Works with Azure OpenAI</span>
        <span class="compat-pill">Works with Local models</span>
      </div>
    </div>
  </div>
</div>

<div class="section-card section-card-story showcase-story">
<!-- ## See the product surface -->
<h2>See the product surface</h2>

<p class="page-lead">Show the product before the rest of the operating brief: first the active surface, then the artifacts that leave the thread.</p>
<p class="editorial-note">These frames are current example captures from the shipped CLI and public example set. They show interface shape and artifact posture, not live checkout, store rollout, or signed-download completion. Each card carries a capture id and source so reused frames stay explicit and auditable instead of reading like independent screenshots.</p>
<div class="showcase-illustration-frame">
  <img src="{{ '/assets/story/jini-showcase-strip.svg' | relative_url }}" alt="Jini storyboard showing the CLI thread, continuity across upcoming surfaces, and sendable output">
</div>
<p class="editorial-note">This is a checked-in storyboard illustration built from the current public product posture. It is not a screenshot, store mockup, or claim that unreleased app surfaces are already live.</p>

<div class="media-grid media-grid-featured">
  {% for card in site.data.showcase_media.product_surface_cards %}
  <a class="media-card media-card-tone-{{ card.tone }}" href="{{ card.href | relative_url }}">
    <span class="media-window-chrome" aria-hidden="true"></span>
    <span class="media-overlay-tags" aria-hidden="true">
      {% for tag in card.tags %}
      <span>{{ tag }}</span>
      {% endfor %}
    </span>
    <img src="{{ card.image | relative_url }}" alt="{{ card.alt }}">
    <div class="media-artifact-stack" aria-hidden="true">
      {% for artifact in card.artifacts %}
      <span>{{ artifact }}</span>
      {% endfor %}
    </div>
    <div class="media-copy">
      <span class="workflow-meta">{{ card.workflow_meta }}</span>
      <h3>{{ card.title }}</h3>
      <p>{{ card.body }}</p>
      <p class="media-capture-note"><strong>{{ card.capture_id }}</strong> · {{ card.capture_kind }} · {{ card.capture_source }}</p>
      <p class="media-truth-note">{{ card.truth_note }}</p>
      <p class="media-reuse-note">{{ card.reuse_note }}</p>
    </div>
  </a>
  {% endfor %}
</div>

<p class="page-lead">The output story should read like finished work, not a gallery of generic screenshots.</p>
<p class="editorial-note">These are example artifacts from the public scenario set. They are meant to show deliverable shape and continuity, not hidden customer data or fictional production outcomes.</p>
<div class="showcase-illustration-frame showcase-illustration-frame-output">
  <img src="{{ '/assets/story/jini-output-strip.svg' | relative_url }}" alt="Jini storyboard showing a follow-up, readiness check, and recommendation memo leaving the thread">
</div>
<p class="editorial-note">This is a second checked-in storyboard illustration built from the current public examples. It is a product-story visual, not a customer capture or implied live enterprise workflow.</p>

<div class="media-grid media-grid-story">
  {% for card in site.data.showcase_media.output_cards %}
  <a class="media-card media-card-tone-{{ card.tone }}" href="{{ card.href | relative_url }}">
    <span class="media-window-chrome" aria-hidden="true"></span>
    <span class="media-overlay-tags" aria-hidden="true">
      {% for tag in card.tags %}
      <span>{{ tag }}</span>
      {% endfor %}
    </span>
    <img src="{{ card.image | relative_url }}" alt="{{ card.alt }}">
    <div class="media-artifact-stack" aria-hidden="true">
      {% for artifact in card.artifacts %}
      <span>{{ artifact }}</span>
      {% endfor %}
    </div>
    <div class="media-copy">
      <span class="workflow-meta">{{ card.workflow_meta }}</span>
      <h3>{{ card.title }}</h3>
      <p>{{ card.body }}</p>
      <p class="media-capture-note"><strong>{{ card.capture_id }}</strong> · {{ card.capture_kind }} · {{ card.capture_source }}</p>
      <p class="media-truth-note">{{ card.truth_note }}</p>
      <p class="media-reuse-note">{{ card.reuse_note }}</p>
    </div>
  </a>
  {% endfor %}
</div>
</div>

<div class="section-card section-card-story section-card-handoff">
<!-- ## What Jini leaves behind -->
<h2>What Jini leaves behind</h2>

<p class="page-lead">After the surface is visible, the remaining question is simple: does Jini leave behind work, evidence, and continuity that survive handoff without reconstruction?</p>

<div class="workflow-grid workflow-grid-story">
  <div class="workflow-card">
    <span class="workflow-meta">Deliverables</span>
    <h3>Usable artifacts</h3>
    <p>Jini writes the thing you actually need next: a follow-up, memo, readiness check, task list, or recommendation.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">Continuity</span>
    <h3>One session, not four different products</h3>
    <p>Jini should preserve the same session identity across macOS, Windows, mobile, and CLI so continuation stays cheaper than restart.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">Evidence</span>
    <h3>Route evidence</h3>
    <p>Users learn one shell while providers, models, and tools can change underneath. Jini exposes command latency, provider state, route cost, and route trend, and keeps that route evidence inspectable when the cost story matters.</p>
  </div>
  <div class="workflow-card">
    <span class="workflow-meta">Boundary</span>
    <h3>Stored locally, not as product magic</h3>
    <p>The first useful result should be a follow-up, memo, readiness check, or itinerary, not a prettier status wall. Artifacts, resume state, and route evidence stay inspectable instead of hiding in a proprietary cloud memory layer.</p>
  </div>
</div>

<p class="editorial-note">The free shell should already be enough to finish serious work. The paid layer should stay later and narrower: only after Jini can prove it saved money or prevented stalled work.</p>
</div>

<div class="section-card section-card-soft proof-signal-panel surface-story release-confidence-section">
<!-- ## Proof, kept brief -->
<h2>Proof, kept brief</h2>

<p class="page-lead">Trust should support the pitch, not bury it. Keep the proof visible, mark the release posture clearly, then get back to the work.</p>

<p class="editorial-note proof-carousel-note">These are checked-in proof carousel slides from the repo, not mocked app-store shots or implied live-release screenshots.</p>

<div class="proof-carousel-grid">
  {% for slide in site.data.proof_carousel.slides %}
  <div class="proof-carousel-card">
    <img src="{{ slide.image | relative_url }}" alt="{{ slide.alt }}">
    <div class="proof-carousel-copy">
      <span class="workflow-meta">{{ slide.eyebrow }}</span>
      <h3>{{ slide.title }}</h3>
      <p>{{ slide.body }}</p>
      <p class="proof-carousel-truth">{{ slide.truth_note }}</p>
    </div>
  </div>
  {% endfor %}
</div>

<div class="proof-signal-grid">
  {% for card in site.data.public_proof.proof_cards %}
  <div class="proof-signal-card">
    <span class="proof-signal-value">{{ card.value }}</span>
    <p>{{ card.label }}</p>
  </div>
  {% endfor %}
</div>

<div class="trust-architecture-grid">
  <div class="trust-architecture-card">
    <span class="workflow-meta">Inspectability</span>
    <h3>Local by default</h3>
    <p>Artifacts, resume state, and handoff evidence stay readable instead of disappearing into product magic.</p>
  </div>
  <div class="trust-architecture-card">
    <span class="workflow-meta">Route evidence</span>
    <h3>Operational truth stays visible</h3>
    <p>Provider state, command latency, route cost, and route trend stay inspectable when the work needs explanation.</p>
  </div>
  <div class="trust-architecture-card">
    <span class="workflow-meta">Continuity</span>
    <h3>One thread, not repeated restarts</h3>
    <p>The same work should survive the move from CLI to later desktop and mobile surfaces without reconstruction.</p>
  </div>
  <div class="trust-architecture-card">
    <span class="workflow-meta">Paid boundary</span>
    <h3>Optimization must earn its place</h3>
    <p>The paid layer belongs only where savings, preserved headroom, or interruption recovery can actually be shown.</p>
  </div>
</div>

<p class="page-lead proof-boundary-brief"><strong>Free first. Paid only if the savings story is measurable.</strong> No live store claims before release. No fake telemetry. No hidden preview posture.</p>

<div class="release-confidence-band surface-story">
<!-- | {{ surface.name }} | {{ surface.badge }} | {{ surface.current_state }} | -->
<table>
  <thead>
    <tr>
      <th scope="col">Surface</th>
      <th scope="col">Status</th>
      <th scope="col">Current state</th>
    </tr>
  </thead>
  <tbody>
    {% for surface in site.data.public_surfaces.surfaces %}
    <tr>
      <th scope="row">{{ surface.name }}</th>
      <td>{{ surface.badge }}</td>
      <td>{{ surface.current_state }}</td>
    </tr>
    {% endfor %}
  </tbody>
</table>

<p class="availability-note editorial-note">The CLI is the live installable surface right now. Desktop and mobile remain in release preparation and should stay free to download when each surface is live. Desktop and Android should distribute directly first where policy allows, while the paid layer should begin only after the planned 30-day free trial and measurable proof are real. Desktop launch plan: buy on the website, then sign in to unlock paid features. Android should prefer direct download first where policy allows, with store distribution secondary. iOS remains store-constrained. Mobile should not be the first place users subscribe.</p>
</div>
</div>

<div class="section-card section-card-soft release-confidence-followup surface-story">
<!-- ## Where you can use it now -->
<h2>Where you can use it now</h2>

<!-- ## Small front door. Clear paid boundary. -->
<h3>Small front door. Clear paid boundary.</h3>

<p class="page-lead closing-boundary-brief"><strong>Start with <code>jini</code>.</strong> If you need the small public command list, use <code>jini commands</code>. Outside the front door, the public support list should stay short: <code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code>. Everything else is support plumbing, not the product story.</p>

<div class="economics-table surface-story">
<table>
  <thead>
    <tr>
      <th scope="col">Need</th>
      <th scope="col">Free Jini shell</th>
      <th scope="col">Planned paid layer</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row">Start and finish work</th>
      <td>Included</td>
      <td>Not required</td>
    </tr>
    <tr>
      <th scope="row">Use local models or your own provider accounts</th>
      <td>Included</td>
      <td>Not required</td>
    </tr>
    <tr>
      <th scope="row">Keep work resumable and inspectable</th>
      <td>Included</td>
      <td>Not required</td>
    </tr>
    <tr>
      <th scope="row">Forecast provider limits before they hit</th>
      <td>Basic route health visibility</td>
      <td>Throttle Radar and preemptive warning</td>
    </tr>
    <tr>
      <th scope="row">Avoid throttles with automatic fallback</th>
      <td>Manual route switch and manual resume</td>
      <td>Route Autopilot across configured tools, CLIs, providers, and local models</td>
    </tr>
    <tr>
      <th scope="row">Resume automatically after interruptions</th>
      <td>Manual today</td>
      <td>Auto Resume with managed session recovery</td>
    </tr>
    <tr>
      <th scope="row">Prove savings over time</th>
      <td>Inspectable route evidence</td>
      <td>Token Savings Ledger tied to renewal proof</td>
    </tr>
  </tbody>
</table>
</div>

<p class="editorial-note economics-note">Free Jini should already be enough to start, finish, self-monitor throttles, switch routes manually, and resume serious work. The paid layer should only exist when Jini can run the optimization loop automatically and prove savings, preserved headroom, or recovered sessions.</p>
<p class="editorial-note closing-source-note">The homepage status block is fed from <code>docs/_data/public_surfaces.json</code>, a sanitized snapshot built from the commercial release packets.</p>
</div>

<div class="section-card section-card-cta">
  <h2>Start with the one command that matters</h2>
  <p>Install Jini, run <code>jini</code>, and describe the task or paste the rough notes you already have. Learn the rest only when the work actually needs it.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html' | relative_url }}">Install</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Quickstart</a>
    <a class="cta-button cta-button-secondary" href="{{ '/examples.html' | relative_url }}">Examples</a>
  </div>
</div>
