---
title: Simple Guide
description: The shortest explanation of what Jini is for and what to type first.
---

Jini helps when work is messy and you do not want to lose time figuring out
what is ready, what is missing, and what to do next.

That is the whole idea.

If you want the shortest first-run path:

1. install Jini
2. run `jini`
3. paste the work you want finished

If setup is missing, Jini should say so in the shell. Then type:

```text
Use Auto
```

If your company or workflow requires one strict route, use the matching path on
the [Install](./install.html) page instead.

## Think About Normal Problems

Jini helps when:

- a meeting ended and the follow-up is fuzzy
- a plan looks finished but may not be safe yet
- a choice was made but the reasoning is getting lost
- a trip is still scattered across tabs and notes
- a problem is fixed but the real closure work is easy to skip

## The Smallest Way To Use It

```bash
jini
```

Install once, then `jini` is the front door.

Inside Jini you can:

- paste messy notes
- ask for the outcome you want
- type `Use Auto` only if Jini says setup is missing
- type `Connect Claude`, `Connect Bedrock`, `Connect Azure OpenAI`, or `Connect Local SLM` if you want to steer the route
- choose `Show what's ready`
- choose `Show what is missing`
- choose `Help me plan this`
- choose `Switch project` when more than one project is active

Before Jini starts a new piece of work, it should also show a short decision
card with:

- tool
- provider
- how chosen
- model
- effort level
- why it made that choice

When you choose `Help me plan this`, Jini should slow down and structure the work
into goal, requirements, design, steps, and run.

## If You Use Claude

The easiest path is:

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
jini
```

Then inside Jini, type:

```text
Connect Claude
```

Jini will ask for the API key and save it in this repo's `.jini` folder.
Do not commit `.jini`.

If you do not know how to begin, type:

```text
help me finish this
```

Then type something plain like:

```text
turn these meeting notes into a follow-up I can send
```

If setup is not the thing you are stuck on, skip straight to the
[Install](./install.html) page for the provider-specific paths.

## What Good Help Looks Like

Jini should quickly tell you:

- the goal
- the working inputs
- the AI route when it matters
- what it just finished
- what it is doing now
- what it will do next
- what else is already active
- what is already done
- what it still needs
- why that missing thing matters
- what your options are
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
    <a href="./examples.html">Examples</a>
    <a href="./install.html">Install</a>
    <a href="./cli.html">CLI Guide</a>
  </div>
</div>
