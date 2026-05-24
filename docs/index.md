---
title: Jini
description: Free orchestration for AI work across providers today. Pay only when Jini can prove it saves money or keeps work moving.
---

<div class="hero-panel hero-panel-marketing">
  <p class="hero-kicker">AI work that has to survive week two</p>
  <h1 class="hero-title">Turn messy AI work into something you can actually send.</h1>
  <p class="hero-summary">Jini turns rough notes, transcripts, screenshots, and drafts into follow-ups, readiness checks, and decision memos that survive handoff. The core shell stays open. Start in the CLI today, then carry the same work forward as desktop and mobile come online. The paid layer stays narrow: it only enters when Jini can prove it saved money or kept work moving.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html' | relative_url }}">Install Free</a>
    <a class="cta-button cta-button-secondary" href="{{ '/examples.html' | relative_url }}">See Examples</a>
    <a class="cta-button cta-button-secondary" href="{{ '/simple.html' | relative_url }}">Quickstart</a>
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
  <div class="hero-scene">
    <div class="hero-scene-card hero-scene-card-live">
      <span class="hero-scene-label">Live now</span>
      <strong>CLI thread</strong>
      <p>Paste rough work into <code>jini</code>, keep route evidence visible, and leave with a usable next artifact instead of a chat dead-end.</p>
      <span class="hero-scene-chip">installable today</span>
    </div>
    <div class="hero-scene-connector" aria-hidden="true">
      <span></span>
    </div>
    <div class="hero-scene-card hero-scene-card-future">
      <span class="hero-scene-label">Next surfaces</span>
      <strong>Desktop and mobile continuity</strong>
      <p>Carry the same thread forward as each app surface comes online instead of starting over from scratch.</p>
      <span class="hero-scene-chip">when each surface is live</span>
    </div>
    <div class="hero-scene-connector" aria-hidden="true">
      <span></span>
    </div>
    <div class="hero-scene-card hero-scene-card-output">
      <span class="hero-scene-label">What leaves the thread</span>
      <strong>Sendable output</strong>
      <p>Follow-up, readiness check, or decision memo with enough reasoning attached that someone else can act on it.</p>
      <span class="hero-scene-chip">artifact first</span>
    </div>
  </div>
</div>

<div class="section-card hero-decision-frame" markdown="1">
  <span class="section-kicker">Choose the lightest layer that still leaves behind usable work.</span>
  <h2>Jini earns the right to exist when the work has to leave chat with a sendable artifact, a safer handoff, and reasoning you can still explain later.</h2>
  <div class="signal-grid">
    <div class="signal-card">
      <h3>After the meeting</h3>
      <p>Turn messy notes into a sendable follow-up with owners, decisions, and open questions someone else can act on.</p>
    </div>
    <div class="signal-card">
      <h3>Before the handoff</h3>
      <p>Check whether a plan, spec, or task bundle is actually ready before someone else has to build on top of it.</p>
    </div>
    <div class="signal-card">
      <h3>Before the decision</h3>
      <p>Keep the reasoning, tradeoffs, and unanswered questions attached to the recommendation instead of buried in chat history.</p>
    </div>
  </div>
  <div class="quote-strip">
    <strong>If you only need a one-off answer, use a raw model shell.</strong>
    <p>If the work needs a sendable artifact, a safer handoff, or reasoning you can still explain later, use Jini.</p>
  </div>
  <p class="page-lead">Use the raw shell for one-shot answers, the free Jini shell for real work, and the paid layer only after it can prove it saved money or kept work moving.</p>

  <div class="offer-grid">
    <div class="offer-card offer-card-plain">
      <span class="workflow-meta">One-shot work</span>
      <h3>Use the raw shell for one-shot answers.</h3>
      <p>Fastest path when no artifact, continuation, or handoff matters after the answer lands.</p>
    </div>
    <div class="offer-card offer-card-core">
      <span class="workflow-meta">Default choice</span>
      <h3>Use the free Jini shell when the work has to survive handoff.</h3>
      <p><strong>Free orchestration core</strong> for work that needs a usable artifact, visible route choices, and resumable state instead of another throwaway answer.</p>
    </div>
    <div class="offer-card offer-card-paid">
      <span class="workflow-meta">Only after proof</span>
      <h3>Add the paid optimizer only when the proof can be measured.</h3>
      <p><strong>Pay only for proof</strong>: when checkout is live, the planned 30-day free trial should come before the $1/month subscription, and only after Jini can show savings, preserved headroom, or recovered work.</p>
    </div>
  </div>
<p class="editorial-note">Best fit when the work needs a sendable artifact, a safer handoff, or reasoning you can still explain later. If you only need one answer, stay in a raw shell.</p>
</div>

<div class="section-card section-card-story showcase-story" markdown="1">
## See the product surface

<p class="page-lead">Show the product before the rest of the operating brief: first the active surface, then the artifacts that leave the thread.</p>
<p class="editorial-note">These frames are current example captures from the shipped CLI and public example set. They show interface shape and artifact posture, not live checkout, store rollout, or signed-download completion.</p>
<p class="editorial-note">Each card carries a capture id and source so reused frames stay explicit and auditable instead of reading like independent screenshots.</p>
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

<div class="section-card section-card-story" markdown="1">
## What Jini leaves behind

<p class="page-lead">After the surface and outputs are visible, judge the product by what it actually leaves behind: usable deliverables, resumable state, visible evidence, and work that survives a device switch without reconstruction.</p>

<div class="workflow-grid workflow-grid-story">
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
  <div class="proof-card">
    <h3>One session, not four different products</h3>
    <p>Jini should preserve the same session identity across macOS, Windows, mobile, and CLI instead of treating each surface like a separate workflow.</p>
  </div>
  <div class="proof-card">
    <h3>Resume without reconstruction</h3>
    <p>The latest deliverable, missing items, route evidence, and next action should travel with the session so device switches do not force context rebuilding.</p>
  </div>
  <div class="proof-card">
    <h3>Review on one surface, continue on another</h3>
    <p>Users should be able to inspect outputs on one device and continue the same work on another without losing the thread.</p>
  </div>
  <div class="proof-card">
    <h3>Cheap continuity</h3>
    <p>Continuation should be cheaper than restart. Jini should reuse session state before it spends money rebuilding the same context again.</p>
  </div>
</div>

<div class="proof-grid proof-grid-story">
  <div class="proof-card">
    <h3>One stable front door</h3>
    <p>Users learn one shell and one set of actions while providers, models, and tools can change underneath.</p>
  </div>
  <div class="proof-card">
    <h3>Cheaper routing by default</h3>
    <p>Jini keeps routine work on the cheapest suitable route and escalates only when depth, verification, or policy requires it.</p>
  </div>
  <div class="proof-card">
    <h3>Outputs you can use</h3>
    <p>The first useful result should be a follow-up, memo, readiness check, or itinerary, not a prettier status wall.</p>
  </div>
</div>

<p class="editorial-note">The free shell should already be enough to finish serious work. The paid layer should stay later and narrower: only after Jini can prove it saved money or prevented stalled work.</p>
</div>

<div class="section-card section-card-soft proof-signal-panel" markdown="1">
## Proof, kept brief

<p class="page-lead">Trust should support the pitch, not bury it. Show the key signals, make the release posture explicit, then get back to the work.</p>

<div class="proof-hero-card">
  <span class="proof-hero-kicker">{{ site.data.public_proof.hero.eyebrow }}</span>
  <h3>{{ site.data.public_proof.hero.headline }}</h3>
  <p>{{ site.data.public_proof.hero.body }}</p>
</div>

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

<div class="quote-strip">
  <strong>Free first. Paid only if the savings story is measurable.</strong>
  <p>No live store claims before release. No fake telemetry. No hidden preview posture.</p>
</div>
</div>

<div class="section-card section-card-story" markdown="1">
## Where you can use it now

<p class="page-lead">Keep rollout truth, download posture, and activation timing in one place so the homepage reads like a current release brief instead of a promise stack.</p>

<div class="checklist-grid rollout-grid">
  <div class="checklist-card rollout-card">
    <h3>Today</h3>
    <p>The CLI is the live installable surface right now. Desktop and mobile are still in release preparation and not yet publicly downloadable.</p>
  </div>
  <div class="checklist-card rollout-card">
    <h3>Free downloads when live</h3>
    <p>macOS, Windows, iOS, and Android stay free to download when each surface is ready. Desktop and Android should distribute directly first where policy allows, while iOS remains App Store constrained.</p>
  </div>
  <div class="checklist-card rollout-card">
    <h3>Paid only after proof</h3>
    <p>When checkout is live, the Commercial License should begin with the planned 30-day free trial and become $1/month subscription pricing for provider-limit forecasting, throttle avoidance, automatic fallback, and automatic resume.</p>
  </div>
</div>

<div class="pill-list">
  {% for surface in site.data.public_surfaces.surfaces %}
  <span>{{ surface.name }}: {{ surface.badge }}</span>
  {% endfor %}
</div>

| Surface | Status | Current state |
|---|---|---|
{% for surface in site.data.public_surfaces.surfaces %}
| {{ surface.name }} | {{ surface.badge }} | {{ surface.current_state }} |
{% endfor %}

<p class="availability-note">Desktop launch plan: buy on the website, then sign in to unlock paid features. Android should prefer direct download first where policy allows, with store distribution secondary. iOS remains store-constrained. Mobile should not be the first place users subscribe.</p>

<p>This homepage status block is fed from <code>docs/_data/public_surfaces.json</code>, a sanitized snapshot built from the commercial release packets.</p>

## Small front door. Clear paid boundary.

<p class="page-lead">Normal work should start with one command, expose evidence when needed, and keep the paid boundary narrow and explicit.</p>

<div class="quote-strip">
  <strong>Start with <code>jini</code>.</strong>
  <p>If you need the small public command list, use <code>jini commands</code>. Setup and route debugging exist when the work needs them, but they are support tools, not the product story.</p>
</div>

<div class="economics-table" markdown="1">
| Need | Free Jini shell | Planned paid layer |
|---|---|---|
| Start and finish work | Included | Not required |
| Use local models or your own provider accounts | Included | Not required |
| Keep work resumable and inspectable | Included | Not required |
| Forecast provider limits before they hit | Not included | Planned |
| Avoid throttles with automatic fallback | Not included | Planned |
| Resume automatically after interruptions | Manual today | Planned |

<p>Free Jini should already be enough to start, finish, and resume serious work. The paid layer should only exist to save money or keep work moving when providers become the bottleneck.</p>
</div>

<div class="proof-grid proof-grid-story">
  <div class="proof-card">
    <h3>Plain files and visible state</h3>
    <p>Jini keeps artifacts, work state, and route evidence legible instead of hiding everything in a proprietary cloud memory layer.</p>
  </div>
  <div class="proof-card">
    <h3>Stored locally, not as product magic</h3>
    <p>Artifacts, resume state, route evidence, and repo-local setup state under <code>.jini</code> stay inspectable. No hidden cloud memory, no opaque routing claims without measurable evidence, and no auto-share behavior that skips review.</p>
  </div>
  <div class="proof-card">
    <h3>Measured trust signals</h3>
    <p><code>jini metrics</code> reports command timings and route evidence so efficiency is inspectable, not implied. Strict routes still exist when policy requires them. The paid layer should eventually show month-to-date savings, headroom preserved, throttles avoided, and sessions resumed.</p>
  </div>
</div>
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
