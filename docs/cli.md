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
  <p>If you only want the smallest useful command set, begin with grouped help, one example, and one real outcome check.</p>
</div>

## Most People Start Here

```bash
jini start --harness codex
jini example research-prd
jini outcome
```

Use those three commands to:

- install the starter harness path
- feel the product on a familiar workflow
- inspect the current state of real work before pushing toward an outcome

## Daily Workflow Commands

Use these when you are working on one pack or handoff:

```bash
jini plan /path/to/work --repo /path/to/repo --intent wiki
jini handoff --repo /path/to/repo --harness codex
jini activate --repo /path/to/repo --harness codex
jini run --repo /path/to/repo --harness codex
```

Those commands let Jini act like a harness orchestration CLI:

- `plan`: choose the right execution posture
- `handoff`: stage a harness-ready bundle
- `activate`: materialize the selected harness surface
- `run`: execute the flow through the chosen harness

## Outcome Commands

Use these when you want to keep the work itself coherent:

```bash
jini outcome
jini artifacts
jini show prd
jini next --repo /path/to/repo --intent verify
jini resume --repo /path/to/repo --intent verify --max-chars 900
jini advance-pack /path/to/work
```

They answer:

- what state the work is in
- what happens next
- what is still missing
- what context to hand to the next agent or person

If you want the plain-language explanation of why those surfaces matter, read
[State And Artifacts](./state-and-artifacts.md).

## Install And Harnesses

Use these when you are setting up Jini in a shell or agent environment:

```bash
jini start --harness codex
jini harnesses
jini guide --harness codex
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
