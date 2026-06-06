# Claude Code Target Shim

This shim represents the Claude Code-specific activation surface for Jini bundles.

The intended pattern is:

- keep canonical payloads universal
- isolate Claude Code-specific bindings in a thin shim
- review shim behavior before activation

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
