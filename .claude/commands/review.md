# Review Current Jini Work

Review the current diff for product, implementation, test, and release risk.

## Procedure

1. Read `git status --short --branch`.
2. Read `git diff --cached` if staged changes exist; otherwise read `git diff`.
3. Use the relevant project agents mentally:
   - `jini-cli-quality-reviewer`
   - `jini-route-runtime-reviewer`
   - `jini-prd-drift-reviewer`
   - `jini-release-gate-reviewer`
   - `jini-security-boundary-reviewer`
4. Prioritize concrete release blockers over style feedback.

## Output

Return findings first, ordered by severity. Include file references and test gaps.

$ARGUMENTS
