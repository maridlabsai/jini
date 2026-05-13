---
title: Install
description: Use the smallest supported path to get Jini running as a normal CLI.
---

Jini should have one obvious install path.

<div class="section-card">
  <h3>The shortest safe path</h3>
  <p>Install Jini, let it show you the beginner path, then verify the target before you do anything more advanced.</p>
</div>

For the smallest guided path from GitHub:

```bash
pipx install --editable git+https://github.com/maridlabsai/jini.git
jini start --harness codex
jini example research-prd
jini outcome
```

What this does:

- installs Jini as a real CLI while preserving the source checkout it needs for packs and specs
- installs the starter kit for one harness
- verifies that the harness-facing surface is ready
- gives you one proof command immediately after setup

Current distribution boundary:

- `v0.1.0` supports source-backed installs
- the supported public path is editable install from GitHub or a local checkout
- a conventional wheel-only install is not documented yet because Jini still
  couples public runtime assets and writable state to the source-backed runtime
  layout

If you prefer a local source checkout instead of `pipx`, the equivalent smoke path is:

```bash
python3 -m pip install -e .
jini start --harness codex
```

Once Jini is installed, the normal user surface is just `jini ...`. For the
small grouped command set, see [the CLI guide](./cli.md).

If you want the manual trust path instead of the one-command setup:

```bash
jini guide --harness codex
jini plan-install --kit starter-kit --harness codex
jini install-bundles --kit starter-kit --harness codex --prefix /tmp/jini-stage
jini doctor-install --kit starter-kit --harness codex --prefix /tmp/jini-stage
```

<div class="section-card">
  <h3>After install</h3>
  <div class="on-this-page">
    <a href="./proof.md">Run the proof command</a>
    <a href="./examples.md">Try a common workflow</a>
    <a href="./cli.md">See the grouped CLI guide</a>
    <a href="https://github.com/maridlabsai/jini/blob/main/specs/install-packaging.md">Read deeper install details</a>
  </div>
</div>
