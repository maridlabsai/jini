# GitHub Copilot Target Shim

This shim represents the GitHub Copilot-specific activation surface for Jini bundles.

The intended pattern is:

- install shared payloads once
- bind Copilot-specific instructions separately
- keep destination-specific behavior explicit and auditable

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
