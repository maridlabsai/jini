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

Most people should only need one command:

```bash
jini
```

That enters the Jini shell.

Inside the shell you can work naturally:

- paste messy notes
- say what outcome you want
- choose `Open ready work`
- choose `See what is still missing`
- choose `Plan this first` when the work needs a structured plan before execution

`jini check`, `jini open`, and `jini run` still exist for scripts and power
users. They are not the normal user model.

## Model Providers

Jini should work with the provider your team already has. The user choice is:

- provider: `JINI_PROVIDER`
- model: `JINI_MODEL`

If you leave either unset, Jini treats it as `auto`.

Supported provider choices:

- `auto`
- `claude` or `anthropic`
- `bedrock`
- `azure-openai`
- `local-preview`

For setup or troubleshooting:

```bash
jini provider doctor
```

The doctor tells you what Jini will use, what `auto` resolved to, what is
missing, and never prints API keys, AWS secret keys, profile values, or model
IDs.

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

### Full Auto

```bash
JINI_PROVIDER=auto
JINI_MODEL=auto
jini provider doctor
```

If you ask for `JINI_MODEL=sonnet-4.6` with `JINI_PROVIDER=auto`, Jini prefers
Bedrock. If nothing cloud-backed is configured, it falls back to local preview.

Current provider support in the Go binary:

- Claude direct can generate the first useful draft through the Messages API.
- Azure OpenAI can generate the first useful draft through deployment chat completions.
- Amazon Bedrock can generate the first useful draft through the Converse API.
- Local preview remains deterministic and offline.
- Setup and cloud errors stay visible without printing secrets.

## What The Screen Should Tell You

When Jini is helping, the screen should read like this:

```text
You're working on
Research to PRD handoff

Working with
Azure OpenAI

Jini is using
Latest PRD draft and review comments

Jini is doing
Checking assumptions and approval gaps
2 of 4 steps done

Ready now
- Build-Readiness Check
- Handoff Brief

Still missing
- Product approval
- Rollback note

Not sure about
- Whether approval was already granted in the review thread

Next step
Open Build-readiness check

Safe to do
Nothing has been sent yet. You can review before sharing.
```

That is the product. Not a wall of commands. Not a file tree. Not a hidden
chat state.

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

## Current Preview

The rewrite is moving to a Go-based binary CLI.

Today’s preview is source-built, but the user-facing shape is already being
cut over.

Build the current preview locally:

```bash
go build ./cmd/jini
./jini
```

Or run it without a build step:

```bash
go run ./cmd/jini
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
