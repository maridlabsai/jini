---
title: CLI Guide
description: The small public command surface and the setup check that sits behind it.
---

Jini should feel small.

Most people should start with one command:

```bash
./jini
```

If Jini is installed on your `PATH`, that command becomes `jini`.

You do not need to know Python to use Jini.

## `./jini` or `jini`

Use `./jini` when you just built the preview from source.

Use `jini` only when the binary is already installed on your `PATH`.

## The Main Entry

Use this first:

```bash
./jini
```

It should:

- continue the thing you were already working on
- or show simple choices if nothing is active yet
- let you type naturally
- expose `Open ready work`, `See what is still missing`, and `Plan this first`
- let you type `Use Claude`, `Use Bedrock`, `Use Azure`, or `Use Auto` to connect a provider inline

## In-Shell Actions

Use these inside Jini:

- `Keep going`
- `Open ready work`
- `See what is still missing`
- `Plan this first`
- `Start something else`

`Plan this first` is the structured mode. Use it when the work is still fuzzy
and you want Jini to turn it into clear steps before execution:

- goal
- requirements
- design
- steps
- run

## Scriptable Commands

These stay available for automation, tests, and power users. They are not the
normal front door.

### `jini check`

Use this when you want a calm work summary.

It should show:

- what you are working on
- what Jini is using
- what Jini is doing
- what is ready now
- what is still missing
- what it is not sure about
- the next step

### `jini open`

Use this when you want to see what Jini already made.

Examples:

```bash
./jini open
./jini open prd
./jini open follow-up
```

The important rule is simple:

- open useful things
- not file paths
- not internal labels

`jini run` remains for explicit or scripted use, but it is not the public
front door.

### `jini provider doctor`

Use this when setting up Claude direct, Azure OpenAI, Amazon Bedrock, or the
local preview.

Most people should start with auto mode:

```bash
./jini provider doctor
./jini
```

Leave provider and model unset first. Let Jini tell you what it can use.

Inside Jini itself, you can also type:

```text
Use Claude
Use Bedrock
Use Azure
Use Auto
```

Jini will ask only for the missing details and save them in this repo's
`.jini` folder.

The setup knobs are:

- `JINI_PROVIDER=auto|claude|bedrock|azure-openai|local-preview`
- `JINI_MODEL=auto|sonnet|sonnet-4.6|...`

If unset, either value defaults to `auto`.

The doctor checks required environment variables and reports only safe presence
information. It tells you what Jini will use, what `auto` resolved to, and
what is missing. It does not print API keys, AWS secret keys, profile values,
or model IDs.

It is a local setup check. It does not guarantee that cloud access is valid.

Direct Claude API:

```bash
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
./jini provider doctor
./jini
```

Amazon Bedrock:

```bash
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
./jini provider doctor
./jini
```

Azure OpenAI:

```bash
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
./jini provider doctor
./jini
```

On Azure, the deployment decides the actual model.

On Bedrock, `sonnet-4.6` maps to Claude Sonnet 4.6.

On direct Claude, `sonnet` maps to Claude Sonnet 4.

Auto mode:

```bash
./jini provider doctor
./jini
```

In auto mode, Jini chooses from the providers that are actually ready in your
environment.

If you hint `sonnet-4.6`, Jini prefers a compatible provider such as Bedrock.

If no cloud provider is ready, it falls back to local preview and tells you
that directly.

## The Shape Jini Is Leaving Behind

Jini is moving away from:

- long command lists
- separate public verbs for every internal step
- path-driven normal use
- builder-first language

That older surface is not the user model anymore.

<div class="section-card">
  <h3>Go next</h3>
  <div class="on-this-page">
    <a href="./simple.md">Simple Guide</a>
    <a href="./examples.md">Examples</a>
    <a href="./state-and-artifacts.md">What Jini Shows</a>
  </div>
</div>
