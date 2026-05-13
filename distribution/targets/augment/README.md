# Augment Target Shim

This shim represents the Augment-specific activation surface for Jini bundles.

The intended pattern is:

- separate universal assets from target bindings
- use a minimal shim for Augment-specific activation
- keep the shim reviewable and removable

This repo exposes the shim through `plan-install`, `doctor-install`, and `activate-runtime-target`.
