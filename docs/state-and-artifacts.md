---
title: What Jini Shows
description: The shortest explanation of what Jini should keep visible while work is moving.
---

Jini should not make you guess.

When it is helping, it should keep these things visible:

- the goal
- the working inputs
- what other active work exists
- the AI route
- what model it chose
- what effort level it judged for this request
- why it made that route choice
- what device class it sees when Local SLM routing matters
- what local runtime and accelerator it sees when Local SLM routing matters
- what it just finished
- what it is doing now
- what it will do next
- what it is doing right now
- what is already done
- what it still needs
- why that missing thing matters
- what options the user has for resolving it
- what comes next
- what is ready now
- what is still missing
- what it is not sure about
- whether it is still safe to review before sharing

## The Command That Matters

```bash
jini
```

That opens Jini after install.

From there, the user should be able to open ready work, see what is missing, or
plan first without learning file paths or command names.

If more than one project is already active, Jini should show `Active work`
first, let the user pick one, and keep the rest visible as sibling work instead
of hiding them behind the filesystem.

## What The Shell Should Tell You

It should read like this:

```text
Goal
Weekly product review follow-up

Working with
- Meeting notes.txt (processed)
- Hotel screenshot.png (attached)

AI route
Amazon Bedrock (chosen automatically)

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

Now
Turning notes into owners and next steps

Done
- Sendable follow-up drafted
- Owners and due points pulled out

Need
Confirm any missing owner or due date before sending this follow-up.

Why this matters
The note is usable now, but it becomes truly sendable only when ownership and timing are explicit.

Options
- Add missing owner
- Add due date
- Skip for now

If you skip this
- Jini will keep the follow-up in draft form and leave missing owner or date gaps visible.

Next
Open Sendable Follow-up

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

That is the model.

The important part is not only the provider name. The user should also be able
to tell whether Jini chose that provider automatically or because they forced it.

It should also be able to tell the user whether this is a low, medium, high, or
extra-high effort request, and why that route was chosen before the first draft begins.

When Local SLM routing is active, Jini should also be able to explain:

- device class
- local runtime class
- accelerator class
- why that pushed the work toward `fast`, `workhorse`, `deep`, or `multimodal`

## What `Show what's ready` Should Feel Like

It should show useful things, not file paths:

- `Sendable Follow-up`
- `Build-Readiness Check`
- `Handoff Brief`
- `Recommendation Memo`
- `Closure Checklist`
- `7 Day Paris Trip`

If a user has to learn the internal storage model before getting value, the
tool is failing.

## What Each Label Means

- `Goal`: the current thing the user is trying to finish
- `Working with`: the visible inputs for this thread, including text, files, images, audio, or links
- `AI route`: the tool/provider route Jini is actually using right now
- `How chosen`: whether the route was automatic, user-locked, or a fallback
- `Model`: the model Jini chose for this request
- `Effort level`: how hard Jini judged this request to be
- `Why this route`: the route policy or explicit user choice behind the tool/provider decision
- `Just finished`: the durable changes from the latest turn
- `Doing now`: the current active step in plain words
- `Up next`: the next artifact or move Jini expects to take
- `Now`: the current step in plain words
- `Done`: what Jini has already completed
- `Need`: the one highest-impact thing Jini still needs
- `Why this matters`: why Jini is asking for that one thing
- `Options`: the bounded ways the user can resolve the active ask
- `If you skip this`: the assumptions or draft limits Jini will preserve if the user does not answer yet
- `Next`: the one most sensible move from here
- `Other active work`: sibling projects you can switch to without losing context
- `Ready now`: things you can open and use immediately
- `Blocked`: blockers that still matter
- `Not sure about`: uncertainty Jini could not safely guess through
- `Safe to do`: whether anything has been sent or changed yet
