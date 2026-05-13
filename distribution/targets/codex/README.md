# Codex Target Shim

This shim represents the Codex-specific activation surface for Jini bundles.

The intended pattern is:

- install universal Jini payloads once
- bind Codex-specific instructions or pointers separately
- keep the shim lightweight and reviewable

This repo exposes the shim through `plan-install`, `doctor-install`, and `activate-runtime-target`.
