---
title: Install
description: Install Jini once, then use `jini`.
---

Install once from any terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
```

Supported host targets today:

- macOS
- Linux

Windows is not first-class yet. The current installer is a Bash installer, and
the real CI matrix runs on macOS and Linux hosts.

That installer:

- first tries to install a matching release binary
- falls back to a source build only when needed
- installs a stable `jini` command into a user bin directory
- verifies that the command launches
- prints one PATH fix only if your shell still cannot see `jini`

If the installer falls back to a source build, it needs Go and Git.

If you already cloned the repo, the local equivalent is:

```bash
./install.sh
```

After install, the normal command is:

```bash
jini
```

## Recommended First Run

For most people, the first run should be:

```bash
jini
```

Then paste the work you want finished.

If Jini says setup is missing, type:

```text
Use Auto
```

That means: Jini picks for you. It chooses the cheapest suitable route by
default, and only moves to a stronger route when the request clearly asks for
deeper work.

If Jini asks for a secret, it saves it in this repo's `.jini` folder so the
next run can stay simple. Do not commit `.jini`.

Use the copy-paste setup blocks below only if:

- your company already gave you the exact values
- you need one strict route
- you are debugging setup

## Copy-Paste Setup Paths

### I use Claude

Copy this:

```bash
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
jini provider doctor
jini
```

Replace `paste-your-key-here` with your real Anthropic API key.

In Jini, `claude` and `anthropic` mean the same provider. Most people should
type `claude`.

### I use Amazon Bedrock

Copy this:

```bash
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
jini provider doctor
jini
```

`JINI_MODEL=sonnet-4.6` means Claude Sonnet 4.6 on Bedrock.

If you already know the exact Bedrock model id, you can set
`BEDROCK_MODEL_ID` instead. When both are set, `BEDROCK_MODEL_ID` wins.

### My team uses Azure OpenAI

If your company already uses Azure OpenAI, Jini can use that same setup.

Copy this:

```bash
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
jini provider doctor
jini
```

In Azure OpenAI, the deployment is the name your team gave the model endpoint.
That deployment decides the actual model. `JINI_MODEL` is optional on Azure and
acts only as a hint for the shell.

Ask your Azure admin for these four values:

- Azure OpenAI endpoint URL
- Azure OpenAI API key
- Azure OpenAI deployment name
- Azure OpenAI API version

If company policy requires Azure OpenAI only, use this Azure path or type
`Connect Azure OpenAI` inside Jini. Do not use Auto mode for that requirement.

### I want Jini to choose automatically

Auto mode is the easiest start.

```bash
export JINI_TOOL=auto
export JINI_PROVIDER=auto
export JINI_MODEL=auto
jini provider doctor
jini
```

Auto mode:

- checks what tools are actually usable
- checks what providers are actually ready
- chooses the cheapest suitable option by default
- chooses the model that fits that route
- judges a request-specific effort level: `low`, `medium`, `high`, or `extra high`
- tells you what it picked
- falls back to local preview if no cloud provider is ready

If you hint `JINI_MODEL=sonnet-4.6`, Jini prefers a Bedrock Sonnet route.

For example:

- trip planning or meeting follow-up can stay on a cheaper Azure-backed writing route
- code-heavy work can stay on a cheaper Azure-backed code route
- if you ask for deep, rigorous, or comprehensive work, Jini can switch to a stronger route such as `Claude Code` or `Bedrock Sonnet`

If you need one strict route, do not use Auto. Force the route you require.

Once Jini picks a route for a piece of work, it saves that choice with the work
so later screens keep showing the same `Working with` label.

### I want to keep most work on a local SLM

If you already run an OpenAI-compatible local model server, Jini can use that
as the cheap-first front line.

Copy this:

```bash
export JINI_PROVIDER=local-slm
export JINI_TOOL=auto
export JINI_MODEL=auto
export JINI_LOCAL_SLM_ENDPOINT="http://127.0.0.1:11434/v1"
export JINI_LOCAL_SLM_MODEL="qwen3:8b"
jini provider doctor
jini
```

Optional profile overrides:

```bash
export JINI_LOCAL_SLM_FAST_MODEL="phi4-mini"
export JINI_LOCAL_SLM_WORKHORSE_MODEL="qwen3:8b-instruct"
export JINI_LOCAL_SLM_DEEP_MODEL="qwen3:14b"
export JINI_LOCAL_SLM_MULTIMODAL_MODEL="gemma3:12b"
```

Inside Jini, the easier path is:

```text
Connect Local SLM
```

Jini will ask only for the endpoint and default model, then keep the profile
slots in the same repo-local `.jini` folder.

Jini also keeps a versioned device profile in `.jini` so local routing can
adjust to the machine, OS version, and runtime capabilities instead of staying
frozen to one generic local model choice.

## What `provider doctor` Really Checks

`jini provider doctor` is a local setup check.

It tells you:

- what Jini will use
- what auto mode resolved to
- how Jini handles effort level
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

- `Show what's ready`
- `Show what is missing`
- `Help me plan this`

## Inside Jini

When Jini opens, if setup is missing or you need one route first, type:

```text
Use Auto
Connect Claude
Connect Bedrock
Connect Azure OpenAI
Connect Local SLM
```

Jini will ask only for the missing details and save them in this repo's `.jini`
folder.

`Use Auto` is the recommended recovery path when setup is missing. It tells
Jini to choose the cheapest suitable tool, provider, and model for the work,
and only switch to a stronger route when the request clearly asks for deeper
work.

After that, the normal start stays:

```bash
jini
```
