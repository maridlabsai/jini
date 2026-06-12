# Commit Gate

Use this skill before committing public Jini changes.

## Procedure

1. Inspect `git status --short --branch`.
2. Run `git diff --check` and `git diff --cached --check`.
3. Run `bash tools/run_required_gates.sh commit`.
4. If Markdown changed and a language scanner is available, run it; otherwise use a conservative local text scan.
5. Review the diff before committing.

## Output

Return gate output summary, exact failure if any, and remaining risk.
