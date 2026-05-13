---
title: Install
description: Use the smallest supported path to get Jini running as a normal CLI.
---

Jini should have one obvious install path.

For the smallest guided path from GitHub:

```bash
pipx install --editable git+https://github.com/maridlabsai/jini.git
jini get-started --target codex
jini plan-install --kit starter-kit --target codex
jini install-bundles --kit starter-kit --target codex --prefix /tmp/jini-stage
jini doctor-install --kit starter-kit --target codex --prefix /tmp/jini-stage
```

What this does:

- installs Jini as a real CLI while preserving the source checkout it needs for packs and specs
- shows the beginner and power-user paths through the same framework
- plans the smallest safe install surface
- materializes the selected bundles
- checks trust, receipts, and activation readiness

Current distribution boundary:

- `v0.1.0` supports source-backed installs
- the supported public path is editable install from GitHub or a local checkout
- a conventional wheel-only install is not documented yet because Jini still
  couples public runtime assets and writable state to the source-backed runtime
  layout

If you prefer a local source checkout instead of `pipx`, the equivalent smoke path is:

```bash
python3 -m pip install -e .
jini get-started --target codex
```

Once Jini is installed, the normal user surface is just `jini ...`. For the
small grouped command set, see [the CLI guide](./cli.md).

For deeper install details, see [specs/install-packaging.md](https://github.com/maridlabsai/jini/blob/main/specs/install-packaging.md).
