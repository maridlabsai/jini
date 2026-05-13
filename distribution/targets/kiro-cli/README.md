# Kiro CLI Target Shim

This shim represents the Kiro CLI-specific activation surface for Jini bundles.

The intended pattern is:

- keep bundle payloads universal
- bind Kiro CLI-specific behavior through a thin shim
- make workflow-default changes visible before activation

This repo exposes the shim through `plan-install`, `doctor-install`, and `activate-runtime-target`.
