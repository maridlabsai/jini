---
title: Install
description: Build the current Go-based Jini preview and start with the provider path that matches your setup.
---

This preview is source-built today.

Use these commands from the Jini repo root.

Before you start:

- install Go
- open a terminal in the Jini repo folder
- run `go version` first if you are not sure Go is available

Build the local binary:

```bash
go build -o jini ./cmd/jini
```

After that, use `./jini` in this preview.

If you later install Jini onto your `PATH`, the command becomes `jini`.

## The Easy Start

Start Jini first:

```bash
go build -o jini ./cmd/jini
./jini
```

Inside Jini, if you want a connected provider, type one of these:

```text
Use Claude
Use Bedrock
Use Azure
Use Auto
```

Jini will ask only for the missing setup details and save them in this repo's
`.jini` folder. After that, your normal start stays:

```bash
./jini
```

## Copy-Paste Setup Paths

### I use Claude

Copy this:

```bash
go build -o jini ./cmd/jini
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
./jini provider doctor
./jini
```

Replace `paste-your-key-here` with your real Anthropic API key.

In Jini, `claude` and `anthropic` mean the same provider. Most people should
type `claude`.

### I use Amazon Bedrock

Copy this:

```bash
go build -o jini ./cmd/jini
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
./jini provider doctor
./jini
```

`JINI_MODEL=sonnet-4.6` means Claude Sonnet 4.6 on Bedrock.

If you already know the exact Bedrock model id, you can set
`BEDROCK_MODEL_ID` instead. When both are set, `BEDROCK_MODEL_ID` wins.

### My team uses Azure OpenAI

If your company already uses Azure OpenAI, Jini can use that same setup.

Copy this:

```bash
go build -o jini ./cmd/jini
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
./jini provider doctor
./jini
```

In Azure OpenAI, the deployment is the name your team gave the model endpoint.
That deployment decides the actual model. `JINI_MODEL` is optional on Azure and
acts only as a hint for the shell.

### I want Jini to choose automatically

Auto mode is the easiest start.

```bash
go build -o jini ./cmd/jini
./jini provider doctor
./jini
```

If `JINI_PROVIDER` and `JINI_MODEL` are unset, Jini uses auto mode.

Auto mode:

- checks what providers are actually ready
- chooses the best available option it can use
- tells you what it picked
- falls back to local preview if no cloud provider is ready

If you hint `JINI_MODEL=sonnet-4.6`, Jini prefers a compatible provider such as
Bedrock.

If you want to force auto explicitly:

```bash
export JINI_PROVIDER=auto
export JINI_MODEL=auto
./jini provider doctor
./jini
```

## What `provider doctor` Really Checks

`./jini provider doctor` is a local setup check.

It tells you:

- what Jini will use
- what auto mode resolved to
- what settings are missing

It does not print secret values.

It does not prove:

- your AWS auth is valid
- Bedrock model access is enabled
- your Azure key is accepted by the service
- your Claude account is allowed to use a model

So the rule is:

- if `provider doctor` says `needs setup`, fix that first
- if it says `ok` and calls still fail, check real provider access next

## First Thing To Do In Jini

When Jini opens, paste the work you already have and ask for the outcome you
want.

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

If Jini shows actions instead of a plain prompt, choose:

- `Open ready work`
- `See what is still missing`
- `Plan this first`

## One Command Versus Two

On the website, you will often see `jini`.

For this source-built preview, use:

```bash
./jini
```

Use plain `jini` only after you install the binary onto your `PATH`.

## What This Preview Is

Today’s Go runtime is the new front door.

It already does the important part:

- one interactive entry point
- visible provider choice
- visible work state
- useful outputs before status recaps

It is still a preview, so the rough edges are real.
