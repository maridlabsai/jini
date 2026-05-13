# GitHub Copilot Target Shim

This shim represents the GitHub Copilot-specific activation surface for Jini bundles.

The intended pattern is:

- install shared payloads once
- bind Copilot-specific instructions separately
- keep destination-specific behavior explicit and auditable

This repo exposes the shim through `plan-install`, `doctor-install`, and `activate-runtime-target`.
