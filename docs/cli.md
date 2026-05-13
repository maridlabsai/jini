---
title: CLI Guide
description: The small grouped command surface most people need once Jini is installed.
---

You do not need to know Python to use Jini.

If Jini is installed, the command is just:

```bash
jini ...
```

If you are not installed yet, start with [install](./install.md).

<div class="section-card">
  <h3>Start with these three</h3>
  <p>If you only want the smallest useful command set, begin with grouped help, one example, and one real pack status check.</p>
</div>

## Most People Start Here

```bash
jini help
jini example research-prd
jini status-pack /path/to/work
```

Use those three commands to:

- see the grouped command surface
- feel the product on a familiar workflow
- inspect the current state of real work before pushing toward an outcome

## Daily Workflow Commands

Use these when you are working on one pack or handoff:

```bash
jini status-pack /path/to/work
jini next /path/to/work --repo /path/to/repo --intent verify
jini resume /path/to/work --repo /path/to/repo --intent verify --max-chars 900
jini advance-pack /path/to/work
```

They answer:

- what state the work is in
- what happens next
- what is still missing
- what context to hand to the next agent or person

If you want the plain-language explanation of why those surfaces matter, read
[State And Artifacts](./state-and-artifacts.md).

## Guided Execution

Use these when you want Jini to stage or run the next workflow step:

```bash
jini recommend-execution /path/to/work --repo /path/to/repo --intent wiki
jini execute-flow /path/to/work --repo /path/to/repo --runtime-target codex
jini run-pack /path/to/work --mode supervised --repo /path/to/repo --runtime-target codex
```

## Install And Targets

Use these when you are setting up Jini in a shell or agent environment:

```bash
jini start --target codex
jini plan-install --kit starter-kit --target codex
jini install-bundles --kit starter-kit --target codex --prefix /tmp/jini-stage
jini doctor-install --kit starter-kit --target codex --prefix /tmp/jini-stage
```

## Publish And Handoffs

Use these when work needs to leave Jini and show up somewhere else:

```bash
jini publish-issues /path/to/work --adapter github --apply-local --format json
jini publish-wiki /path/to/work --adapter markdown --apply-local --format json
jini show-adapters
jini adapter-conformance
```

## System Health

Use these when you want to inspect the framework surface itself:

```bash
jini publish-readiness --format json
jini show-kpis
jini catalog-packs
jini catalog-bundles
```

## Need The Full Surface?

Run:

```bash
jini help
```

Use `jini help --all` when you need the full command inventory.

<div class="section-card">
  <h3>What most people do next</h3>
  <div class="on-this-page">
    <a href="./examples.md">Try another example</a>
    <a href="./proof.md">Re-read the proof screen</a>
    <a href="./contact.md">Ask a question or report a gap</a>
  </div>
</div>
