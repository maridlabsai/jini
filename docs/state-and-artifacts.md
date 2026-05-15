---
title: What Jini Shows
description: The shortest explanation of what Jini should keep visible while work is moving.
---

Jini should not make you guess.

When it is helping, it should keep these things visible:

- what you are working on
- what it is using
- what provider it is working with
- what it is doing
- what is ready now
- what is still missing
- what it is not sure about
- what to do next
- whether it is still safe to review before sharing

## The Command That Matters

```bash
./jini
```

That opens Jini in this source-built preview.

From there, the user should be able to open ready work, see what is missing, or
plan first without learning file paths or command names.

## What The Shell Should Tell You

It should read like this:

```text
You're working on
Weekly product review follow-up

Working with
Amazon Bedrock (chosen automatically)

Jini picked this because
- Bedrock credentials are ready
- the requested model works there

Jini is using
Meeting notes and follow-up tasks

Jini is doing
Turning notes into owners and next steps
2 of 4 steps done

Ready now
- Sendable Follow-up
- Owners and Due Points

Still missing
- Owner confirmation

Not sure about
- Whether every action item has a clear owner

Next step
Open Sendable Follow-up

Safe to do
Nothing has been sent yet. You can review before sharing.
```

That is the model.

The important part is not only the provider name. The user should also be able
to tell whether Jini chose that provider automatically or because they forced it.

## What `Open ready work` Should Feel Like

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

- `Working with`: the provider Jini is actually using right now
- `Jini is using`: the notes, draft, or context it is pulling from
- `Jini is doing`: the current step in plain words
- `Ready now`: things you can open and use immediately
- `Still missing`: blockers that still matter
- `Not sure about`: uncertainty Jini could not safely guess through
- `Next step`: the one most sensible move from here
- `Safe to do`: whether anything has been sent or changed yet
