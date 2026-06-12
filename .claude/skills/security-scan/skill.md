# Security Scan

Use this skill for public Jini security review.

## Procedure

1. Scan changed files for secrets, credentials, private URLs, customer details, pricing strategy, and commercial rollout material.
2. Run `bash tools/security_configuration_gate.sh`.
3. Check install and release changes for supply-chain, signing, and Gatekeeper risk.
4. Confirm any scanner or dependency claim points to a real command.

## Output

Return blockers first with file references and verification commands.
