# Jini

**In plain words:** Jini is a CLI-first AI work router and durable session
layer.

It helps people who use multiple AI CLIs, models, and local runtimes finish work
without losing route control, context, or cost discipline.

It keeps five things clear:

- what you are working on
- what Jini is doing
- what is ready now
- what is still missing
- what to do next

Most AI tools are good at getting work started. Jini is for the part where
developers and operators lose time:

- switching among configured provider, model, and local routes
- avoiding throttling and quota dead ends
- continuing without replaying transcripts
- editing the right local files from the current folder
- knowing what route, artifact, and next action are current

Meeting, plan-readiness, travel, and vendor-comparison flows are proof
scenarios. They are not the product identity. The hard positioning and tiering
decisions live in
[specs/product-settling-decisions.md](specs/product-settling-decisions.md).

## The Public Shape

Jini should feel small.

Published release: [`v0.1.2`](https://github.com/maridlabsai/jini/releases/tag/v0.1.2).

Current source contract for the next release:

- CLI-first: `jini` opens a compact task prompt, not a saved-work dashboard.
- Simple questions, explicit local file edits, setup guidance, and route
  inspection stay direct.
- No `Start/Keep` modal, `Task Snapshot`, or generic scaffold appears for
  simple input.
- Routes are claimed only when Jini can invoke the configured provider, local
  runtime, or installed CLI handoff, or fail closed with setup guidance.
- External sends, bookings, payments, commits, pushes, destructive changes,
  credential changes, and paid managed automation require visible approval.

Install once with the normal path:

```bash
curl -fsSL https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh | bash
```

Supported install hosts today: macOS and Linux.

Desktop and mobile are part of the public product direction, but they should
not be treated as generally available installs until those surfaces are shipped
and pass the same work-thread and trust checks as the CLI.

The public website execution tracker lives in
[docs/website-execution-tracker.md](docs/website-execution-tracker.md).

If you already cloned the repo, the local equivalent is:

```bash
./install.sh
```

Then most people should only need one command:

```bash
jini
```

That opens Jini.

The first run should look like this:

```text
Jini
>
```

From there, direct work should stay direct:

```text
> what is the capital of france
Paris.

> add a line saying "hello from Jini" in the pear fellow script .txt file in this folder
Updated pear fellow script.txt
- Added line: hello from Jini
- Location: /path/to/pear fellow script.txt
```

If Jini cannot safely choose a file, it should fail closed and list the
candidate filenames instead of guessing.

If you want the small public command catalog without dropping into the shell,
run:

```bash
jini commands
```

If you maintain routes, bundles, or release plumbing, the deeper inventory
lives under `jini admin help`.

Inside the shell you can work naturally:

- paste messy notes
- say what outcome you want
- ask `what is blocked?`
- ask `open the latest artifact`
- ask `plan this change` when the work needs a structured path before execution
- type the saved project title when you want to resume another active thread
- if setup is missing, Jini should tell you what to connect instead of making you memorize route phrases

The first screen should stay light: no saved-work dashboard, no tutorial block,
and no artifact/status frame before the user asks for work.

If you are not sure how to start, type:

```text
help me finish this
```

If you need the public command list, use `jini commands`. If you maintain
routes, bundles, or release plumbing, use `jini admin help`. Those paths exist
to support the product, not replace the normal `jini` front door.

The public support surface outside the front door should stay short:

- `jini status`
- `jini continue`
- `jini open`
- `jini route`
- `jini doctor`

Everything else should stay behind `jini help --admin` unless Jini explicitly
points you there.

If multiple projects are already in flight, bare `jini` should still start with
the same task prompt. Saved work belongs behind explicit inspection commands
such as `jini status`, `jini continue`, and `jini open`, or natural title
matching when the user types the saved project name.

When route choice matters, Jini should keep route evidence inspectable without
making users learn routing before value:

- `Tool`
- `Provider`
- `How chosen`
- `Model`
- `Effort level`
- `Why this route`
- `Continuity` when Jini keeps the current coding route to preserve context instead of churning tools
- for multimodal work, Jini now says when it is learning separately from screenshot, scanned PDF, or audio/transcript evidence
- when that learning exists, Jini also shows a compact `Multimodal learning` block in the shell so you can inspect those separate buckets without leaving the work

That evidence should support the work instead of becoming the first-run
experience. The default flow should produce a useful result first, then make
route, model, effort, and continuity easy to inspect through `jini route`,
`jini status`, or the saved work state.

When you need to steer the route directly, keep it explicit and reversible:

```bash
jini route list
jini route help
jini route dogfood
jini route smoke codex
jini route validate codex --real-cli --checks all
jini route set azure-code
jini route auto
```

That gives developers a clear escape hatch for framework switching without
making every task start with routing setup. `codex` and `claude-code` are
reserved for real installed-CLI handoff: Jini invokes the installed CLI when it
is present and trusted, and fails closed with setup guidance when it is not.
Use `jini route help` when you need the env vars for Codex, Claude Code,
Gemini CLI, Aider, OpenCode, provider routes, or local SLM routes.
Use `jini route dogfood` before release validation to see which installed CLI
routes still need real auth, approval, output-shape, and receipt-privacy
evidence in `.jini/cli-dogfood.json`.
After a real installed CLI smoke succeeds, use
`jini route validate <route> --real-cli --checks all` to record that evidence
without hand-editing JSON.

For multi-step work, Jini now plans quietly before drafting. The user sees the
result through better structure and clearer `Doing now` / `Next` state, not
through an exposed planner transcript. When route confidence is weak or the
work is higher-stakes, Jini may also run one focused refine pass before showing
the first artifact instead of always returning the very first generated result unchanged.
For `extra high` work, Jini can also run a selective consistency check with a
second independent draft and keep the better result.
For coding work, Jini now also tries to preserve a suitable current route when
the quality gap is small, and it can learn from repeated manual route choices
on similar coding work instead of churning tools unnecessarily.

## Routing And Providers

Most people should ignore the routing knobs and start with:

1. install Jini
2. run `jini`
3. describe the task or paste the notes, files, screenshot, transcript, or rough ask
4. if Jini says setup is missing, type `auto`

Jini now has three setup knobs behind that simple path:

- tool: `JINI_TOOL`
- provider: `JINI_PROVIDER`
- model: `JINI_MODEL`

If you leave them unset, Jini stays in the simplest path. Inside the shell,
`auto` means: Jini picks the cheapest suitable route by default, and
switches to a stronger route only when the request clearly asks for deeper work.

Supported tool choices today:

- `auto`
- `claude-api`
- `bedrock-sonnet`
- `azure-openai`
- `chatgpt`
- `azure-code`
- `local-fast`
- `local-workhorse`
- `local-deep`
- `local-multimodal`
- `local-preview`

CLI handoff routes:

- `codex`
- `claude-code`
- `gemini-cli`
- `aider`
- `opencode`

Adapter support waves:

- Wave 0: shared handoff contract, route receipts, `doctor` detection, and
  fail-closed setup guidance
- Wave 1: Codex, Claude Code, Gemini CLI, Aider, and OpenCode
- Wave 2: Ollama, LM Studio, OpenRouter, and LiteLLM-compatible gateways
- Wave 3: Continue, Cline/Roo, Cursor, Windsurf, and GitHub Copilot coding
  agent

Wave 0 and Wave 1 are runtime-supported when the downstream CLI is installed
and trusted by the OS. Missing or rejected CLIs fail closed with setup guidance
instead of silently falling back to a provider API alias.
Wave 2 and Wave 3 are planned targets, not shipped claims.
Gemini API and Vertex AI provider routes are planned only after
`gemini-cli` dogfood validates the CLI route. When added, they will be labeled
as provider API routes, not as `gemini-cli`.

Supported provider choices:

- `auto`
- `claude` or `anthropic`
- `bedrock`
- `azure-openai`
- `local-slm`
- `local-preview`

For setup or troubleshooting:

```bash
jini route help
jini route dogfood
jini route smoke codex
jini route validate codex --real-cli --checks all
jini doctor
```

Route help shows what to install or which env vars to set. The doctor tells you
what Jini will use, what `auto` resolved to, what is missing, and never prints
API keys, AWS secret keys, profile values, or model IDs.
Route dogfood shows the Wave 1 CLI validation checklist and a copyable evidence
template without marking any route validated automatically. Route validate
records evidence only after you explicitly confirm a real installed CLI smoke.
Route smoke runs the configured CLI with a harmless prompt and reports receipt
metadata without printing prompt or output bodies.

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
jini doctor
```

`JINI_MODEL=sonnet` resolves to Claude Sonnet 4 on the direct Claude API.

### Amazon Bedrock

```bash
JINI_PROVIDER=bedrock
AWS_REGION=us-east-1
AWS_PROFILE=...
JINI_MODEL=sonnet-4.6
jini doctor
```

`JINI_MODEL=sonnet-4.6` resolves to Claude Sonnet 4.6 on Bedrock. You can also
set `BEDROCK_MODEL_ID` directly. If your team does not use AWS profiles, use
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` instead of `AWS_PROFILE`; add
`AWS_SESSION_TOKEN` when your credentials are temporary.

### Azure OpenAI

```bash
JINI_PROVIDER=azure-openai
AZURE_OPENAI_ENDPOINT=...
AZURE_OPENAI_API_KEY=...
AZURE_OPENAI_DEPLOYMENT=...
JINI_MODEL=sonnet
jini doctor
```

On Azure, the deployment decides the actual model. `JINI_MODEL` is only a user
hint for the shell.

### Local SLM

```bash
JINI_PROVIDER=local-slm
JINI_TOOL=auto
JINI_MODEL=auto
```

In default `auto` mode, if an OpenAI-compatible local server is already running
on a common loopback port, Jini can discover it and choose the model
automatically before spending remote tokens. Today it checks:

```text
http://127.0.0.1:11434/v1   # Ollama OpenAI-compatible endpoint
http://127.0.0.1:1234/v1    # LM Studio
http://127.0.0.1:8080/v1    # llama.cpp-style local servers
http://127.0.0.1:8000/v1    # local OpenAI-compatible servers
```

Use explicit settings only when auto-discovery is not enough:

```bash
JINI_LOCAL_SLM_ENDPOINT=http://127.0.0.1:11434/v1
JINI_LOCAL_SLM_MODEL=qwen3:8b
jini doctor
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
architecture, memory, accelerator class, local runtime shape, and power state,
refreshes that profile when the Jini/OS/runtime shape drifts, and uses backend
readiness plus the device profile to decide whether `fast`, `workhorse`,
`deep`, or `multimodal` is actually suitable. Explicit model settings still
win, but without them Jini scores the discovered local models by task, device
class, and low-battery posture.

Auto-discovered mobile routes are intentionally narrow: Jini only treats
lightweight fine-tuned local models as eligible on iOS/Android-class devices.
Laptop routes split into light and pro policy tiers from measured device SKU
signals, so light laptops avoid pro-sized local models while pro laptops can
use stronger local workhorse models when the device can handle them.

When you run `jini doctor` on a Local SLM setup, Jini also records a
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

### Full auto

```bash
JINI_TOOL=auto
JINI_PROVIDER=auto
JINI_MODEL=auto
jini doctor
```

If you ask for `JINI_MODEL=sonnet-4.6` with `JINI_TOOL=auto`, Jini prefers a
Bedrock Sonnet route. If only Claude direct is ready, Jini can choose the
Claude provider route. If only Azure is ready, Jini can choose an Azure-backed
route. If nothing cloud-backed is configured, it falls back to local preview.

In auto mode, Jini also looks at the work itself:

- tool choice, model choice, and effort choice are all first-class runtime decisions
- by default, Jini chooses the cheapest suitable route for the work
- the target shape is for commercially usable local SLMs to become the normal front line for low and medium work, with paid remote routes used only when quality or policy requires escalation
- that local path is now a real routed pool, with lighter local profiles for intake and cleanup and stronger local profiles for drafting, reasoning, or multimodal work
- you can now connect an OpenAI-compatible local model server with `Connect Local SLM`
- Jini also judges an effort level for each request: `low`, `medium`, `high`, or `extra high`
- trips, follow-ups, and decision-writing usually stay on cheaper Azure-backed writing routes
- code-heavy requests usually stay on cheaper Azure-backed code routes
- if the request explicitly asks for deep, rigorous, comprehensive, or high-rigor work, Jini switches to the strongest ready route such as `Bedrock Sonnet` or another configured provider/model route
- the chosen route is saved with the work so later screens keep showing the same `Working with` label

Real downstream CLI handoff for names like `codex` and `claude-code` is a P0
contract: Jini invokes the installed trusted CLI or fails closed with setup
guidance. Public route claims must stay limited to routes the current binary
actually invokes.

Current provider support in the Go binary:

- Claude direct can generate the first useful result through the Messages API.
- Azure OpenAI can generate the first useful result through deployment chat completions.
- Amazon Bedrock can generate the first useful result through the Converse API.
- OpenAI-compatible local SLM servers can generate the first useful result through chat completions.
- Local preview remains deterministic and offline.
- Setup and cloud errors stay visible without printing secrets.

## What The Screen Should Tell You

The default screen should stay small: prompt, answer, edit receipt, or one short
setup hint. Durable work details belong behind explicit commands such as
`jini status`, `jini continue`, `jini open`, and `jini route`; they are not the
first-minute product.

For durable work only, requested status can show the goal, route, ready outputs,
missing blockers, and next safe action. It must not appear around simple
questions, obvious local edits, greetings, or unclear input.

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
