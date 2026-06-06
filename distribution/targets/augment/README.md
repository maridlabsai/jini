# Augment Target Shim

This shim represents the Augment-specific activation surface for Jini bundles.

The intended pattern is:

- separate universal assets from target bindings
- use a minimal shim for Augment-specific activation
- keep the shim reviewable and removable

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
