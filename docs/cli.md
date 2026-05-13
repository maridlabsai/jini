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

## Most People Start Here

```bash
jini try-example research-prd
jini get-started --target codex
jini status-pack /path/to/work
```

Use those three commands to:

- feel the product on a familiar workflow
- see the smallest install and target path
- inspect the current state of real work

## Daily Workflow Commands

Use these when you are working on one pack or handoff:

```bash
jini status-pack /path/to/work
jini execution-checklist /path/to/work --repo /path/to/repo --intent verify
jini compact-context /path/to/work --repo /path/to/repo --intent verify --max-chars 900
jini advance-pack /path/to/work
```

They answer:

- what state the work is in
- what happens next
- what is still missing
- what context to hand to the next agent or person

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
jini get-started --target codex
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
jini --help
```

That is the authoritative source for every installed command.
