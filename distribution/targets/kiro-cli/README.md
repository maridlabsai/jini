# Kiro CLI Target Shim

This shim represents the Kiro CLI-specific activation surface for Jini bundles.

The intended pattern is:

- keep bundle payloads universal
- bind Kiro CLI-specific behavior through a thin shim
- make workflow-default changes visible before activation

This repo documents the target shim as metadata. Use `jini doctor --format json` and `jini publish-readiness --format json` for native Go health checks today.
