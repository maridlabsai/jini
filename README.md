# Jini

**In plain words:** Jini helps you finish work without losing track of what
matters.

It keeps five things clear:

- what you are working on
- what Jini is doing
- what is ready now
- what is still missing
- what to do next

Most AI tools are good at getting work started. Jini is for the part where
people lose time:

- after a meeting
- before sending a plan forward
- before choosing one option
- before calling something finished

## The Public Shape

Jini should feel small.

Install once with the normal path:

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
```

Supported install hosts today: macOS and Linux.

Desktop and mobile are part of the public product direction, but they should
not be treated as generally available installs until those surfaces are shipped
and pass the same work-thread and trust checks as the CLI.

If you already cloned the repo, the local equivalent is:

```bash
./install.sh
```

Then most people should only need one command:

```bash
jini
```

That opens Jini.

Inside the shell you can work naturally:

- paste messy notes
- say what outcome you want
- if setup is missing, type `Use Auto`
- type `Connect Claude`, `Connect Bedrock`, `Connect Azure OpenAI`, or `Connect Local SLM` if you want to steer the route
- choose `Show what's ready`
- choose `Show what is missing`
- choose `Help me plan this` when the work needs a structured plan before execution
- choose `Switch project` when more than one project is active

The first screen should keep one obvious move in front of you:

- `Paste notes or type what you want finished.`

If you are not sure how to start, type:

```text
help me finish this
```

`jini check`, `jini open`, and `jini run` still exist for scripts and power
users. They are not the normal user model.

If multiple projects are already in flight, Jini should show `Active work`
first, let you pick one, and keep sibling work visible under `Other active work`
while one project stays in focus.

Before Jini starts new work, it should also show a short decision card:

- `Tool`
- `Provider`
- `How chosen`
- `Model`
- `Effort level`
- `Why this route`
- `Continuity` when Jini keeps the current coding route to preserve context instead of churning tools
- for multimodal work, Jini now says when it is learning separately from screenshot, scanned PDF, or audio/transcript evidence
- when that learning exists, Jini also shows a compact `Multimodal learning` block in the shell so you can inspect those separate buckets without leaving the work

That card should appear before the first draft so the user can see what Jini is
about to do instead of guessing after the fact.

For multi-step work, Jini now plans quietly before drafting. The user sees the
result through better structure and clearer `Doing now` / `Next` state, not
through an exposed planner transcript. When route confidence is weak or the
work is higher-stakes, Jini may also run one focused refine pass before showing
the first artifact instead of always returning the very first draft unchanged.
For `extra high` work, Jini can also run a selective consistency check with a
second independent draft and keep the better result.
For coding work, Jini now also tries to preserve a suitable current route when
the quality gap is small, and it can learn from repeated manual route choices
on similar coding work instead of churning tools unnecessarily.

## Routing And Providers

Most people should ignore the routing knobs and start with:

1. install Jini
2. run `jini`
3. paste the work you want finished
4. if Jini says setup is missing, type `Use Auto`

Jini now has three setup knobs behind that simple path:

- tool: `JINI_TOOL`
- provider: `JINI_PROVIDER`
- model: `JINI_MODEL`

If you leave them unset, Jini stays in the simplest path. Inside the shell,
`Use Auto` means: Jini picks the cheapest suitable route by default, and
switches to a stronger route only when the request clearly asks for deeper work.

Supported tool choices today:

- `auto`
- `claude-code`
- `bedrock-sonnet`
- `azure-openai`
- `chatgpt`
- `codex`
- `local-fast`
- `local-workhorse`
- `local-deep`
- `local-multimodal`
- `local-preview`

Supported provider choices:

- `auto`
- `claude` or `anthropic`
- `bedrock`
- `azure-openai`
- `local-slm`
- `local-preview`

For setup or troubleshooting:

```bash
jini provider doctor
```

The doctor tells you what Jini will use, what `auto` resolved to, what is
missing, and never prints API keys, AWS secret keys, profile values, or model
IDs.

For multimodal work, doctor also shows the separate learning buckets Jini keeps
for screenshot work, scanned PDF/document work, and audio/transcript work.

When tool routing is active, Jini also tells you which tool it chose and which
provider it routed through.

If Jini asks for a secret inside the shell, it saves it in this repo's `.jini`
folder so the next run can stay simple. Do not commit `.jini`.

### Direct Claude API

```bash
JINI_PROVIDER=claude
ANTHROPIC_API_KEY=...
JINI_MODEL=sonnet
jini provider doctor
```

`JINI_MODEL=sonnet` resolves to Claude Sonnet 4 on the direct Claude API.

### Amazon Bedrock

```bash
JINI_PROVIDER=bedrock
AWS_REGION=us-east-1
AWS_PROFILE=...
JINI_MODEL=sonnet-4.6
jini provider doctor
```

`JINI_MODEL=sonnet-4.6` resolves to Claude Sonnet 4.6 on Bedrock. You can also
set `BEDROCK_MODEL_ID` directly.

### Azure OpenAI

```bash
JINI_PROVIDER=azure-openai
AZURE_OPENAI_ENDPOINT=...
AZURE_OPENAI_API_KEY=...
AZURE_OPENAI_DEPLOYMENT=...
JINI_MODEL=sonnet
jini provider doctor
```

On Azure, the deployment decides the actual model. `JINI_MODEL` is only a user
hint for the shell.

### Local SLM

```bash
JINI_PROVIDER=local-slm
JINI_TOOL=auto
JINI_MODEL=auto
JINI_LOCAL_SLM_ENDPOINT=http://127.0.0.1:11434/v1
JINI_LOCAL_SLM_MODEL=qwen3:8b
jini provider doctor
```

Optional local profile overrides:

```bash
JINI_LOCAL_SLM_FAST_MODEL=phi4-mini
JINI_LOCAL_SLM_WORKHORSE_MODEL=qwen3:8b-instruct
JINI_LOCAL_SLM_DEEP_MODEL=qwen3:14b
JINI_LOCAL_SLM_MULTIMODAL_MODEL=gemma3:12b
```

Inside Jini, the easiest path is `Connect Local SLM`.

Jini now treats these local slots as device-aware. It probes the OS, version,
architecture, memory, accelerator class, and local runtime shape, refreshes
that profile when the Jini/OS/runtime shape drifts, and uses backend readiness
plus the device profile to decide whether `fast`, `workhorse`, `deep`, or
`multimodal` is actually suitable.

When you run `jini provider doctor` on a Local SLM setup, Jini also records a
small measured capability report for the local profiles. That report captures
real request success, warm latency, cold-start cost, token throughput, and
structured-output reliability so auto routing can prefer the fastest
good-enough local path instead of relying only on machine class heuristics.
If a Local SLM endpoint is already configured, Jini also starts warming that
report in the background during interactive shell startup so the first normal
work request can often use measured routing without needing doctor first.
Jini also keeps a short rolling history for each local adapter so routing can
notice regressions or improving trends instead of trusting only the latest
sample. Trend penalties are confidence-weighted, so one noisy warm-up does
not overcorrect the route as hard as sustained degradation does. Older
regressions decay over time so the router can recover after a local runtime
or model fix. Strong rebound after a degraded streak is also detected so a
recovered local route can win back score faster. Jini also scales that
benchmark signal by current work shape, so recovered structured-drafting
performance helps planning more than unrelated code work. Within planning,
Jini now narrows that again by artifact family, so recovered checklist-style
evidence helps readiness checks more than trip-itinerary requests. When Jini
sees real Local SLM completions for a cohort like `trip-itinerary` or
`sendable-followup`, it now records that direct cohort evidence and uses it
ahead of transferred adapter evidence for later similar requests. Existing
`Model upvote` and `Model downvote` signals now feed that same cohort memory,
so later local routing can learn from whether people found a cohort’s output
actually useful, not just well-formed. Jini also accepts richer artifact
feedback for Local SLM cohorts: `Accepted as is`, `Needed light edits`, and
`Not useful`. It now also learns passively from edit distance on the saved
artifact, so “accepted after tiny cleanup” and “substantive rewrite needed”
do not get treated as the same signal. That passive learning is now
section-aware, so title/header cleanup, supporting-section edits, and core
content rewrites are weighted differently. Inside core sections, Jini now also
distinguishes wording-only edits from actual decision or recommendation
changes. Jini also learns from downstream outcome signals like `Used this`,
`Shared this`, and `Replaced this`, so real artifact adoption counts more than
edit patterns alone. It also records passive workflow signals from repeated
artifact opens, export opens, and substantive reopen-after-rewrite events, so
Local SLM routing can keep learning even when the user never labels the
outcome explicitly. For downstream work outside Jini, you can opt in with
`jini observe add <external-file>`, and Jini will scan that external copy on
normal work loads to learn from reuse or substantive replacement there too.

### Full Auto

```bash
JINI_TOOL=auto
JINI_PROVIDER=auto
JINI_MODEL=auto
jini provider doctor
```

If you ask for `JINI_MODEL=sonnet-4.6` with `JINI_TOOL=auto`, Jini prefers a
Bedrock Sonnet route. If only Claude direct is ready, Jini can choose Claude
Code. If only Azure is ready, Jini can choose an Azure-backed route. If nothing
cloud-backed is configured, it falls back to local preview.

In auto mode, Jini also looks at the work itself:

- tool choice, model choice, and effort choice are all first-class runtime decisions
- by default, Jini chooses the cheapest suitable route for the work
- the target shape is for commercially usable local SLMs to become the normal front line for low and medium work, with paid remote routes used only when quality or policy requires escalation
- that local path is now a real routed pool, with lighter local profiles for intake and cleanup and stronger local profiles for drafting, reasoning, or multimodal work
- you can now connect an OpenAI-compatible local model server with `Connect Local SLM`
- Jini also judges an effort level for each request: `low`, `medium`, `high`, or `extra high`
- trips, follow-ups, and decision-writing usually stay on cheaper Azure-backed writing routes
- code-heavy requests usually stay on cheaper Azure-backed code routes
- if the request explicitly asks for deep, rigorous, comprehensive, or high-rigor work, Jini switches to a best-tool-first route such as `Claude Code` or `Bedrock Sonnet`
- the chosen route is saved with the work so later screens keep showing the same `Working with` label

Current provider support in the Go binary:

- Claude direct can generate the first useful draft through the Messages API.
- Azure OpenAI can generate the first useful draft through deployment chat completions.
- Amazon Bedrock can generate the first useful draft through the Converse API.
- OpenAI-compatible local SLM servers can generate the first useful draft through chat completions.
- Local preview remains deterministic and offline.
- Setup and cloud errors stay visible without printing secrets.

## What The Screen Should Tell You

When Jini is helping, the screen should read like this:

```text
Goal
Research to PRD handoff

Working with
- Latest PRD draft and review comments

AI route
Azure OpenAI

How chosen
Automatic

Model
gpt-4o-prod

Effort level
Medium

Why this route
Auto mode prefers the cheapest suitable tool for this kind of work.

Continuity
Kept the current coding route to preserve context continuity because the quality gap was not material.

Just finished
- Build-readiness draft created
- Missing build blockers identified

Doing now
Separating what is ready from what still blocks build

Up next
Open Missing Pieces Before Build

Now
Checking assumptions and approval gaps

Done
- Build-readiness draft created
- Missing build blockers identified

Need
Name the approval owner and confirm the first implementation slice.

Why this matters
The readiness check is useful now, but build should not start until approval and the first slice are explicit.

Options
- Set approval owner
- Set first slice
- Skip for now

If you skip this
- Jini will keep approval and first-slice gaps visible instead of treating the plan as build-ready.

Next
Open Build-readiness check

Ready now
- Build-Readiness Check
- Handoff Brief

Blocked
- Product approval
- Rollback note

Not sure about
- Whether approval was already granted in the review thread

Safe to do
Nothing has been sent yet. You can review before sharing.
```

That is the product. Not a wall of commands. Not a file tree. Not a hidden
chat state.

It should also feel like a live work thread, not only a static summary. Users
should be able to see what changed this turn, what Jini is doing now, what is
next, and the one high-impact clarification Jini is waiting on.

## The Problems Jini Is Being Tuned For

### After a meeting

You want:

- a sendable follow-up
- clear owners
- decisions made
- open questions

### Before sending a plan forward

You want:

- a clear answer on whether it is ready
- a short list of what is still missing
- a safe next step

### Before choosing one option

You want:

- a recommendation you can defend
- a short tradeoff summary
- fewer repeat debates

### Before calling an incident closed

You want:

- a closure checklist
- proof that the important follow-up happened
- less cleanup later

## Install Once

If you already cloned the repo, run:

```bash
./install.sh
```

If the installer says your shell cannot see the new command yet, add the PATH
line it prints, then restart the shell.

After that, start Jini with:

```bash
jini
```

If you want the public site version of this story instead of the repo version,
start here:

- [Home](./docs/index.md)
- [Simple Guide](./docs/simple.md)
- [Examples](./docs/examples.md)
- [Install](./docs/install.md)

## What Stays Public

The core tool stays public:

- runtime
- docs
- examples
- tests

Paid work later is for:

- adoption help
- implementation help
- team onboarding
- custom integrations
- enterprise support

## Support

- [GitHub Issues](https://github.com/maridlabsai/jini/issues) for bugs and clear gaps
- [GitHub Discussions](https://github.com/maridlabsai/jini/discussions) for questions and workflow feedback
- `maridlabsai@gmail.com` for commercial contact
