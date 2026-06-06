# Codex Target Shim

This shim represents the Codex-specific activation surface for Jini bundles.

The intended pattern is:

- install universal Jini payloads once
- bind Codex-specific instructions or pointers separately
- keep the shim lightweight and reviewable

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
