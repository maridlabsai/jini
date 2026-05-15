---
title: Simple Guide
description: The shortest explanation of what Jini is for and what to type first.
---

Jini helps when work is messy and you do not want to lose time figuring out
what is ready, what is missing, and what to do next.

That is the whole idea.

If you want the shortest first-run path, start on the [Install](./install.md)
page and copy the setup block that matches your provider.

## Think About Normal Problems

Jini helps when:

- a meeting ended and the follow-up is fuzzy
- a plan looks finished but may not be safe yet
- a choice was made but the reasoning is getting lost
- a trip is still scattered across tabs and notes
- a problem is fixed but the real closure work is easy to skip

## The Smallest Way To Use It

```bash
./jini
```

`./jini` is the front door in this source-built preview.

If Jini is already installed on your `PATH`, the command becomes `jini`.

Inside Jini you can:

- paste messy notes
- ask for the outcome you want
- choose `Open ready work`
- choose `See what is still missing`
- choose `Plan this first`

When you choose `Plan this first`, Jini should slow down and structure the work
into goal, requirements, design, steps, and run.

## If You Use Claude

Copy this first:

```bash
go build -o jini ./cmd/jini
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
./jini provider doctor
./jini
```

Then type something plain like:

```text
turn these meeting notes into a follow-up I can send
```

If setup is not the thing you are stuck on, skip straight to the
[Install](./install.md) page for the provider-specific paths.

## What Good Help Looks Like

Jini should quickly tell you:

- what you are working on
- what it is using
- what it is doing
- what is ready now
- what is still missing
- what it is not sure about
- the next step

It should also reassure you when the output is still a draft and safe to
review before sharing.

## The Kinds Of Things It Should Give Back

- a follow-up you can send
- a short plan check
- a recommendation you can explain
- a trip plan you can actually use
- a closure checklist

Not:

- a confusing status wall
- a path to a file you have to interpret
- a bunch of internal labels

## If You Only Remember One Rule

Jini should reduce stress, not add process.

If it makes you think harder about the tool than about the work, it is failing.

<div class="section-card">
  <h3>Go next</h3>
  <div class="on-this-page">
    <a href="./examples.md">Examples</a>
    <a href="./install.md">Install</a>
    <a href="./cli.md">CLI Guide</a>
  </div>
</div>
