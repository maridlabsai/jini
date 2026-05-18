---
title: CLI Guide
description: The small public command surface and the setup check that sits behind it.
---

Jini should feel small.

Most people should start with one command:

```bash
jini
```

You do not need to know Python to use Jini.

## Install First

Use [Install](./install.html) once, then the normal command is just `jini`.

## The Main Entry

Use this first:

```bash
jini
```

It should:

- continue the thing you were already working on
- show `Active work` first if more than one project is already in flight
- or show simple choices if nothing is active yet
- let you type naturally, with one obvious prompt: `Paste notes or type what you want finished.`
- show a short decision card before new work starts: `Tool`, `Provider`, `How chosen`, `Model`, `Effort level`, `Why this route`, and `Continuity` when Jini stays on the current coding route to preserve context
- for multimodal work, make the explanation explicit about subtype-scoped learning, so users can see that screenshot, scanned PDF, and audio/transcript work are learned separately
- when that learning exists, show a small `Multimodal learning` block in the preflight card and current work screen so the user can inspect those separate buckets without leaving the shell
- plan quietly before drafting when the request is clearly multi-step
- spend one extra refine pass only when route confidence is weak or the work is higher-risk
- use a selective consistency check only for `extra high` work
- use subtype-aware multimodal routing, consistency checks, and local benchmark memory, so PDF/scan, screenshot, and audio work are neither routed, judged, nor learned with one generic rubric
- for coding work, prefer continuity when the current route is still suitable and the quality gap is small
- let repeated manual route choices on similar coding work bias later auto routing instead of forcing fresh route churn every time
- expose `Show what's ready`, `Show what is missing`, and `Help me plan this`
- let you type `Use Auto` when setup help is needed
- let you type `Connect Claude`, `Connect Bedrock`, `Connect Azure OpenAI`, or `Connect Local SLM` to steer the route inline

## In-Shell Actions

Use these inside Jini:

- `Keep going`
- `Show what's ready`
- `Show what is missing`
- `Help me plan this`
- `Switch project`
- `Start new work`

If more than one project is active, Jini should keep the current one open and
show the others under `Other active work`.

If there is no current focus but work already exists, Jini should open with an
`Active work` list so you can pick the project you want without knowing file
paths or internal ids.

`Help me plan this` is the structured mode. Use it when the work is still fuzzy
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
jini open
jini open prd
jini open follow-up
```

The important rule is simple:

- open useful things
- not file paths
- not internal labels

`jini run` remains for explicit or scripted use, but it is not the public
front door.

### `jini provider doctor`

Use this when setting up Claude direct, Azure OpenAI, Amazon Bedrock, a local
SLM server, or the local preview.

Most people should start by pasting the work they want finished. If setup help
is needed, use auto mode:

```bash
jini provider doctor
jini
```

Leave provider and model unset first. Let Jini tell you what it can use.

Inside Jini itself, you can also type:

```text
Use Auto
Connect Claude
Connect Bedrock
Connect Azure OpenAI
Connect Local SLM
```

Jini will ask only for the missing details and save them in this repo's
`.jini` folder. Do not commit `.jini`.

The setup knobs are:

- `JINI_TOOL=auto|claude-code|bedrock-sonnet|chatgpt|codex|azure-openai|local-fast|local-workhorse|local-deep|local-multimodal|local-preview`
- `JINI_PROVIDER=auto|claude|bedrock|azure-openai|local-slm|local-preview`
- `JINI_MODEL=auto|sonnet|sonnet-4.6|...`

If `JINI_TOOL` is unset, Jini stays in legacy provider mode until you choose a
tool route in the shell. `Use Auto` saves all three as `auto`.

The doctor checks required environment variables and reports only safe presence
information. It tells you what Jini will use, what `auto` resolved to, and
what is missing. It does not print API keys, AWS secret keys, profile values,
or model IDs.

For Local SLM, the doctor also shows the machine view Jini is using for
routing:

- `DEVICE_CLASS`
- `DEVICE_OS`
- `LOCAL_ACCELERATOR`
- `LOCAL_RUNTIME_CLASS`

That is how Jini decides whether `fast`, `workhorse`, `deep`, or
`multimodal` local work is appropriate on this device.

For multimodal work, doctor now also shows the separate learning buckets Jini
keeps for:

- screenshot work
- scanned PDF/document work
- audio/transcript work

If the Local SLM endpoint is ready, doctor also refreshes a small measured
capability report and shows benchmark summary lines for the local profiles.
That lets auto routing account for real load success, warm latency,
cold-start cost, token throughput, and structured-output reliability instead
of using only static device-class guesses.
When you start the interactive shell with a ready Local SLM setup, Jini also
tries to warm that report in the background so the next request can benefit
from measured local routing sooner.
Jini keeps a short rolling history for each local profile as well, so route
scoring can react to drift over time instead of trusting only the latest run.
Those trend penalties are confidence-weighted so a single noisy warm-up does
not count as heavily as repeated regression. Older regressions also decay so
fresh recovery can win back the route faster. When a local route rebounds
strongly after a degraded streak, Jini marks that as recovered and restores
score more aggressively. Jini also weights that benchmark signal by the
current work shape, so structured-draft recovery helps planning more than
unrelated code work. Within planning, Jini narrows that further by artifact
family, so readiness-check evidence does not get full credit on trip
itinerary requests. Jini also records direct Local SLM cohort evidence from
real completions like `trip-itinerary` and `sendable-followup`, so later
requests can use measured task-family evidence instead of only transferred
adapter evidence. `Model upvote` and `Model downvote` also feed that cohort
memory for Local SLM routes, so matching future work can learn from explicit
usefulness signals as well as format and latency. Jini also accepts richer
artifact feedback for matching Local SLM cohorts: `Accepted as is`, `Needed
light edits`, and `Not useful`. It also uses passive edit-distance learning on
the artifact itself, so tiny cleanup and substantive rewrite do not collapse
into one usefulness bucket. That passive learning is section-aware, so header
cleanup does not get treated like a core content rewrite. Within core content,
Jini also separates wording edits from actual decision changes. It also learns
from downstream artifact outcomes like `Used this`, `Shared this`, and
`Replaced this`. It also records passive workflow signals from repeated opens,
export opens, and substantive reopen-after-rewrite events. For downstream work
outside Jini, you can opt in with `jini observe add <external-file>`, and Jini
will scan that external copy on normal work loads to learn from reuse or
substantive replacement there too.

It is a local setup check. It does not guarantee that cloud access is valid.

Direct Claude API:

```bash
export JINI_PROVIDER=claude
export ANTHROPIC_API_KEY="paste-your-key-here"
export JINI_MODEL=sonnet
jini provider doctor
jini
```

Amazon Bedrock:

```bash
export JINI_PROVIDER=bedrock
export AWS_REGION=us-east-1
export AWS_PROFILE="your-profile"
export JINI_MODEL=sonnet-4.6
jini provider doctor
jini
```

Azure OpenAI:

```bash
export JINI_PROVIDER=azure-openai
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="paste-your-key-here"
export AZURE_OPENAI_DEPLOYMENT="your-deployment-name"
export AZURE_OPENAI_API_VERSION=2024-10-21
jini provider doctor
jini
```

On Azure, the deployment decides the actual model.

On Bedrock, `sonnet-4.6` maps to Claude Sonnet 4.6.

On direct Claude, `sonnet` maps to Claude Sonnet 4.

Local SLM:

```bash
export JINI_PROVIDER=local-slm
export JINI_TOOL=auto
export JINI_MODEL=auto
export JINI_LOCAL_SLM_ENDPOINT="http://127.0.0.1:11434/v1"
export JINI_LOCAL_SLM_MODEL="qwen3:8b"
jini provider doctor
jini
```

Optional local profile overrides:

```bash
export JINI_LOCAL_SLM_FAST_MODEL="phi4-mini"
export JINI_LOCAL_SLM_WORKHORSE_MODEL="qwen3:8b-instruct"
export JINI_LOCAL_SLM_DEEP_MODEL="qwen3:14b"
export JINI_LOCAL_SLM_MULTIMODAL_MODEL="gemma3:12b"
```

Auto mode:

```bash
export JINI_TOOL=auto
export JINI_PROVIDER=auto
export JINI_MODEL=auto
jini provider doctor
jini
```

In auto mode, Jini chooses from the tools and providers that are actually ready
in your environment, and it biases the choice toward the kind of work you gave
it.

Examples:

- by default, planning and writing work can prefer a cheaper Azure-backed writing route
- by default, code-heavy work can prefer a cheaper Azure-backed code route
- if you ask for deep, rigorous, or comprehensive work, Jini can move to a stronger route such as `Claude Code` or `Bedrock Sonnet`
- Jini also chooses the model that fits the selected route
- Jini also judges a request-level effort: `low`, `medium`, `high`, or `extra high`

If you need one strict route for policy, data, or cost reasons, do not use
Auto. Use the matching explicit route instead.

After Jini picks a route for a work item, it saves that route with the work so
later screens keep the same `Working with` label.

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
    <a href="./simple.html">Simple Guide</a>
    <a href="./examples.html">Examples</a>
    <a href="./state-and-artifacts.html">What Jini Shows</a>
  </div>
</div>
