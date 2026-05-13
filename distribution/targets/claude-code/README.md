# Claude Code Target Shim

This shim represents the Claude Code-specific activation surface for Jini bundles.

The intended pattern is:

- keep canonical payloads universal
- isolate Claude Code-specific bindings in a thin shim
- review shim behavior before activation

This repo exposes the shim through `plan-install`, `doctor-install`, and `activate-runtime-target`.
