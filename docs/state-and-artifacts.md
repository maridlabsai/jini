---
title: Outputs
description: The shortest explanation of what Jini writes and keeps visible while work is moving.
---

<div class="value-strip">
  <p><strong>Jini should not make you guess.</strong></p>
  <p>While work is moving, the shell should keep the right context visible: the goal, the current inputs, the chosen route, the current step, what is already usable, and what still needs attention.</p>
</div>

<div class="proof-grid">
  <div class="proof-card">
    <h3>Keep The Work Legible</h3>
    <p>Show the goal, the active inputs, and any sibling work so the user always knows what thread they are in.</p>
  </div>
  <div class="proof-card">
    <h3>Keep The AI Choice Honest</h3>
    <p>Show the route, model, effort level, and why Jini chose them. If local routing matters, also show the device and runtime context behind that choice.</p>
  </div>
  <div class="proof-card">
    <h3>Keep Progress Concrete</h3>
    <p>Show what just finished, what is happening now, and what comes next instead of vague “thinking” language.</p>
  </div>
  <div class="proof-card">
    <h3>Keep Risk Visible</h3>
    <p>Show what is ready, what is missing, what is still uncertain, and whether the draft is still safe to review before anything is shared.</p>
  </div>
</div>

<div class="section-card">
  <span class="section-kicker">Why this matters</span>
  <h2>Outputs are part of the buyability story</h2>
  <p>Jini should feel better than plain chat because it gives teams usable deliverables, visible missing pieces, and a continuation surface they can understand without rebuilding state from scratch.</p>
  <p>The same output and session model should travel across macOS, Windows, mobile, and CLI. A device switch should not create a second-class copy of the work.</p>
</div>

<div class="section-card" markdown="1">
  <h2>The Command That Matters</h2>

  ```bash
  jini
  ```

  <p>That should be enough after install. From there, the user should be able to open ready work, see what is missing, or plan first without learning file paths or internal command names.</p>

  <p>If more than one project is active, Jini should show <code>Active work</code> first, let the user pick one, and keep the rest visible as sibling work instead of hiding them behind the filesystem.</p>
</div>

<div class="section-card" markdown="1">
  <h2>What Should Stay Visible</h2>

  <div class="workflow-grid">
    <div class="workflow-card">
      <span class="workflow-meta">Work</span>
      <h3>What am I working on?</h3>
      <p>Keep the goal, the current inputs, and any other active work visible so the user never loses the thread.</p>
      <code>Goal · Working with · Other active work</code>
    </div>
    <div class="workflow-card">
      <span class="workflow-meta">Route</span>
      <h3>Why is Jini using this AI stack?</h3>
      <p>Show the route, model, effort, and route reason. When Local SLM routing matters, also show the device class, local runtime, and accelerator context.</p>
      <code>AI route · Model · Effort · Why this route</code>
    </div>
    <div class="workflow-card">
      <span class="workflow-meta">Progress</span>
      <h3>Where are we now?</h3>
      <p>Make progress readable in plain language by showing the latest completed step, the current step, and the next step.</p>
      <code>Just finished · Doing now · Up next</code>
    </div>
    <div class="workflow-card">
      <span class="workflow-meta">Decision</span>
      <h3>What does Jini need from me?</h3>
      <p>Keep one active ask visible, explain why it matters, and show bounded ways to resolve it or defer it safely.</p>
      <code>Need · Why this matters · Options · If you skip this</code>
    </div>
    <div class="workflow-card">
      <span class="workflow-meta">Output</span>
      <h3>What can I use already?</h3>
      <p>Surface the deliverables that are already ready, the blockers that still matter, and any uncertainty Jini could not safely guess through.</p>
      <code>Ready now · Blocked · Not sure about</code>
    </div>
    <div class="workflow-card">
      <span class="workflow-meta">Safety</span>
      <h3>Can I review safely?</h3>
      <p>Tell the user whether anything has already been sent or changed, so drafts can be reviewed before handoff.</p>
      <code>Safe to do</code>
    </div>
  </div>
</div>

<div class="section-card" markdown="1">
  <h2>What The Shell Should Feel Like</h2>

  <p>The shell should read like a working surface, not like a diagnostic panel. A buyer should be able to see that Jini is organized around deliverables, continuation, and explicit risk rather than model theater.</p>

```text
Goal
Weekly product review follow-up

Working with
- Meeting notes.txt (processed)
- Hotel screenshot.png (attached)

AI route
Amazon Bedrock

How chosen
Automatic

Model
Claude Sonnet 4.6

Effort level
High

Why this route
Auto mode prefers the best planning tool when the request asks for deep or high-rigor work.

Just finished
- Sendable follow-up drafted
- Owners and due points pulled out

Doing now
Tightening owners, due dates, and open questions

Up next
Open Owners and Due Points

Need
Confirm any missing owner or due date before sending this follow-up.

Why this matters
The note is usable now, but it becomes truly sendable only when ownership and timing are explicit.

Options
- Add missing owner
- Add due date
- Skip

If you skip this
- Jini will keep the follow-up in draft form and leave missing owner or date gaps visible.

Ready now
- Sendable Follow-up
- Owners and Due Points

Blocked
- Owner confirmation

Not sure about
- Whether every action item has a clear owner

Safe to do
Nothing has been sent yet. You can review before sharing.

Other active work
- Pricing vendor review
- Paris trip
```

  <p>The important part is not only the provider name. The user should also be able to tell whether Jini chose that route automatically or because they forced it, and how much effort Jini judged the request to need before the first draft begins.</p>
</div>

<div class="section-card" markdown="1">
  <h2>What “Open” Should Feel Like</h2>

  <p>It should open deliverables, not storage concepts. The product surface should make the useful thing obvious before the user has to care where it lives on disk.</p>

  <div class="steps-grid">
    <div class="step-card">
      <span class="step-number">1</span>
      <h3>Open useful work</h3>
      <p>The first thing the user sees should be the memo, checklist, follow-up, plan, or recommendation they can actually review.</p>
    </div>
    <div class="step-card">
      <span class="step-number">2</span>
      <h3>Keep missing context nearby</h3>
      <p>Any blocker, uncertainty, or missing proof should stay visible next to the deliverable instead of being hidden in another command.</p>
    </div>
    <div class="step-card">
      <span class="step-number">3</span>
      <h3>Make the next move obvious</h3>
      <p>The user should always know whether to review, clarify, approve, or continue, without learning Jini’s internal storage model.</p>
    </div>
  </div>

  <p>Examples of the right shelf labels:</p>

  <div class="compat-row">
    <span class="compat-pill">Sendable Follow-up</span>
    <span class="compat-pill">Build-Readiness Check</span>
    <span class="compat-pill">Handoff Brief</span>
    <span class="compat-pill">Recommendation Memo</span>
    <span class="compat-pill">Closure Checklist</span>
    <span class="compat-pill">7 Day Paris Trip</span>
  </div>
</div>

<div class="section-card" markdown="1">
  <h2>What Each Label Should Mean</h2>

  <ul>
    <li><code>Goal</code>: the thing the user is trying to finish right now</li>
    <li><code>Working with</code>: the visible inputs for the thread, including text, files, images, audio, or links</li>
    <li><code>AI route</code>: the tool and provider route Jini is using now</li>
    <li><code>How chosen</code>: whether the route was automatic, user-locked, or a fallback</li>
    <li><code>Model</code>: the model Jini chose for this request</li>
    <li><code>Effort level</code>: how hard Jini judged the request to be</li>
    <li><code>Why this route</code>: the route policy or user choice behind the decision</li>
    <li><code>Just finished</code>: the durable changes from the latest turn</li>
    <li><code>Doing now</code>: the current active step in plain words</li>
    <li><code>Up next</code>: the next artifact or move Jini expects to take</li>
    <li><code>Need</code>: the one highest-impact thing Jini still needs</li>
    <li><code>Why this matters</code>: why Jini is asking for that one thing</li>
    <li><code>Options</code>: the bounded ways the user can resolve the active ask</li>
    <li><code>If you skip this</code>: the draft limits or assumptions Jini will preserve if the user does not answer yet</li>
    <li><code>Ready now</code>: things the user can open and use immediately</li>
    <li><code>Blocked</code>: blockers that still materially matter</li>
    <li><code>Not sure about</code>: uncertainty Jini could not safely guess through</li>
    <li><code>Safe to do</code>: whether anything has already been sent or changed</li>
    <li><code>Other active work</code>: sibling projects the user can switch to without losing context</li>
  </ul>
</div>

<div class="section-card">
  <span class="section-kicker">Design rule</span>
  <h2>What should never happen</h2>
  <div class="checklist-grid">
    <div class="checklist-card">
      <h3>No storage-first labels</h3>
      <p>Users should not have to think in terms of internal files, folders, or hidden state before they can review useful work.</p>
    </div>
    <div class="checklist-card">
      <h3>No diagnostic-first output</h3>
      <p>Status and route details matter, but they should support the deliverable instead of replacing it.</p>
    </div>
    <div class="checklist-card">
      <h3>No invisible risk</h3>
      <p>Anything still missing, uncertain, or blocked should stay near the deliverable instead of being buried behind another step.</p>
    </div>
  </div>
</div>
