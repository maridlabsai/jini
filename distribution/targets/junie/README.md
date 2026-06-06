# Junie Target Shim

This shim represents the Junie-specific activation surface for Jini bundles.

The intended pattern is:

- preserve one canonical Jini payload
- expose Junie-specific bindings through a minimal shim
- review automation-sensitive surfaces before activation

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
