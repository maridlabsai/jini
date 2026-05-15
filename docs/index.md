---
title: Jini
description: Finish work without losing track of what is ready, what is missing, and what to do next.
---

<div class="hero-panel">
  <p class="hero-kicker">Less friction in daily work</p>
  <h1 class="hero-title">Finish work without status hunting or setup guessing.</h1>
  <p class="hero-summary">Jini helps you turn messy work into something usable. It keeps the useful parts visible, shows what is still missing, and makes the next step obvious. It should also be obvious how to start.</p>
  <div class="cta-row">
    <a class="cta-button" href="{{ '/install.html#i-use-claude' | relative_url }}">I Use Claude</a>
    <a class="cta-button cta-button-secondary" href="{{ '/install.html#i-use-amazon-bedrock' | relative_url }}">I Use Bedrock</a>
    <a class="cta-button cta-button-secondary" href="{{ '/install.html#my-team-uses-azure-openai' | relative_url }}">I Use Azure</a>
    <a class="cta-button cta-button-secondary" href="{{ '/install.html#i-want-jini-to-choose-automatically' | relative_url }}">Use Auto Mode</a>
  </div>
</div>

**In plain words:** Jini helps you finish work without losing track of what matters.

It should tell you:

- what you are working on
- what Jini is doing
- what is ready now
- what is still missing
- what to do next

## Start Here

If you are trying the source-built preview, start with:

```bash
go build -o jini ./cmd/jini
./jini
```

If Jini is installed on your `PATH`, the command becomes `jini`.

If you already know your provider, use the matching setup block on the
[Install](./install.md) page:

- [I use Claude](./install.md#i-use-claude)
- [I use Amazon Bedrock](./install.md#i-use-amazon-bedrock)
- [My team uses Azure OpenAI](./install.md#my-team-uses-azure-openai)
- [I want Jini to choose automatically](./install.md#i-want-jini-to-choose-automatically)

If you do not want to think about provider or model settings yet, start with
auto mode. Jini will choose the best ready option it can use and tell you what
it picked.

## What The Screen Should Feel Like

```text
You're working on
Research to PRD handoff

Jini is using
Latest PRD draft and review comments

Jini is doing
Checking assumptions and approval gaps
2 of 4 steps done

Ready now
- Build-Readiness Check
- Handoff Brief

Still missing
- Product approval
- Rollback note

Not sure about
- Whether approval was already granted in the review thread

Next step
Open Build-readiness check

Safe to do
Nothing has been sent yet. You can review before sharing.
```

That is the experience Jini is being rebuilt around.

## First Thing To Type

When Jini opens, do not think about commands first. Paste the work you already
have and ask for the outcome you want.

For example:

```text
turn these meeting notes into a follow-up I can send
```

or:

```text
check whether this plan is ready to hand off
```

or:

```text
plan a 7 day Paris trip with a clear day-by-day itinerary
```

## Start With The Problem You Already Have

<div class="workflow-grid">
  <a class="workflow-card" href="{{ '/examples.html#meeting-followup' | relative_url }}">
    <span class="workflow-meta">After a meeting</span>
    <h3>Turn scattered notes into something you can send.</h3>
    <p>Get a follow-up, owners, decisions, and open questions without rebuilding the meeting later.</p>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#spec-readiness' | relative_url }}">
    <span class="workflow-meta">Before a handoff</span>
    <h3>Check whether a plan is actually ready.</h3>
    <p>See what is safe, what is missing, and what still blocks a real handoff.</p>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#paris-trip' | relative_url }}">
    <span class="workflow-meta">Before a trip</span>
    <h3>Turn travel planning into a usable itinerary.</h3>
    <p>Get the days, likely costs, missing choices, and booking checklist in one place.</p>
  </a>
  <a class="workflow-card" href="{{ '/examples.html#vendor-choice' | relative_url }}">
    <span class="workflow-meta">Before a choice</span>
    <h3>Keep the reasoning attached to the decision.</h3>
    <p>Get a recommendation you can defend instead of a choice you have to re-argue later.</p>
  </a>
</div>

## What You Get Back

Jini should not give you a lecture.

It should give you something you can use:

- a sendable follow-up
- a build-readiness check
- a recommendation memo
- a 7 day trip plan

And then it should show:

- what is still missing
- what it is not sure about
- the next step

<div class="section-card">
  <h3>Go next</h3>
  <div class="on-this-page">
    <a href="./simple.md">Simple Guide</a>
    <a href="./examples.md">Examples</a>
    <a href="./install.md">Install</a>
    <a href="./cli.md">CLI Guide</a>
    <a href="./state-and-artifacts.md">What Jini Shows</a>
    <a href="./contact.md">Contact</a>
  </div>
</div>
