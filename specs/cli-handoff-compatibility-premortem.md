# CLI Hand-off Compatibility Premortem

Chain-of-evidence review of every hand-off route (like the empirical claude-code
validation), covering how each CLI behaves non-interactively and how Jini's
plan/semi/autonomous posture must map to its real flags. Sources are cited; the
non-claude CLIs are not installed here, so their flags are **doc-verified**
(pending empirical `--help` confirmation once installed), and marked as such in
code.

## Evidence — per-CLI non-interactive + approval flags

| CLI | prompt token style | read-only (plan) | semi (edits, no arbitrary cmds) | autonomous (edits+cmds) |
| --- | --- | --- | --- | --- |
| claude-code | positional (`--print X`) | `--print` default ✓ (empirical) | `--permission-mode acceptEdits` | `--dangerously-skip-permissions` |
| codex | positional (`exec X`) | `exec` default = read-only sandbox ✓ | (no clean edits-only mode) | `--dangerously-bypass-approvals-and-sandbox` (alias `--yolo`) |
| gemini-cli | **value** (`-p X`) | default prompts → auto-cancel ~ | `--approval-mode auto_edit` | `--yolo` |
| aider | **value** (`--message X`) | **DEFAULT WRITES** ✗ → needs `--dry-run` | `--yes-always` (aider runs no arbitrary shell) | `--yes-always` |
| opencode | positional (`run X`) | default `edit: ask` → auto-cancel ~ | (config-based, no clean flag) | `--auto` |

## Findings (severity-ordered)

1. **[HIGH — safety] Aider plan posture writes and commits.** Jini's plan
   posture forwards the descriptor's default args unchanged; aider's default
   `--message` **applies edits and auto-commits** ([aider docs]). So a user in
   Ask mode / an untrusted dir who routes to aider gets edits + a git commit
   with no consent — plan is supposed to be a safe, read-only preview. Fix: plan
   posture for aider must pass `--dry-run` (dry run, no file modification).

2. **[HIGH — correctness] Posture-arg insertion corrupts value-style prompts.**
   `applyPostureArgs` inserts posture args immediately before the `{{prompt}}`
   token. That is correct only when `{{prompt}}` is positional (claude `--print
   X`, codex `exec X`, opencode `run X`). For gemini (`-p X`) and aider
   (`--message X`), `{{prompt}}` is a **flag value**, so inserting before it
   yields e.g. `gemini -p --approval-mode auto_edit <prompt>` — the flag becomes
   the value and the prompt becomes a stray positional. This blocks adding
   correct gemini/aider posture args.

3. **[MED — parity] Only claude-code has posture args.** codex/gemini/aider/
   opencode have no semi/autonomous args, so those grants silently degrade to
   plan (safe, and hinted by `postureDegradedHintForDir`, but not "compatible
   like claude"). Now that flags are doc-verified, wire them.

## Fix design — explicit `{{posture}}` slot + `PlanArgs`

Replace the position-guessing with an explicit marker so each descriptor states
exactly where posture args belong, and give plan its own args (for read-only
enforcement):

- Add `{{posture}}` to each `DefaultArgs` at the correct spot:
  - claude: `["--print", "{{posture}}", "{{prompt}}"]`
  - codex: `["exec", "{{posture}}", "{{prompt}}"]`
  - gemini: `["{{posture}}", "-p", "{{prompt}}"]`
  - aider: `["{{posture}}", "--message", "{{prompt}}"]`
  - opencode: `["run", "{{posture}}", "{{prompt}}"]`
- Add `PlanArgs []string` alongside `SemiArgs`/`AutonomousArgs`. Per-CLI:
  - claude Plan `[]`, Semi `--permission-mode acceptEdits`, Auto `--dangerously-skip-permissions`
  - codex Plan `[]`, Semi `nil`, Auto `--dangerously-bypass-approvals-and-sandbox`
  - gemini Plan `[]`, Semi `--approval-mode auto_edit`, Auto `--yolo`
  - aider Plan `--dry-run`, Semi `--yes-always`, Auto `--yes-always`
  - opencode Plan `[]`, Semi `nil`, Auto `--auto`
- `applyPostureArgs` substitutes `{{posture}}` with the selected posture's args
  (dropping the token when empty), then substitutes `{{prompt}}` as today. The
  `JINI_*_ARGS` override still replaces args wholesale (posture is a no-op then),
  unchanged.
- Posture resolution is unchanged: a grant degrades to the highest posture the
  descriptor has verified args for; plan always applies (now including aider's
  `--dry-run`, so plan is read-only for every route).

Sources:
- [Codex flags](https://www.vincentschmalbach.com/how-codex-cli-flags-actually-work-full-auto-sandbox-and-bypass/), [Codex bypass](https://allthings.how/codex-cli-skip-permissions-how-to-bypass-approvals-and-sandbox/)
- [Gemini approval modes](https://deepwiki.com/char8x/gemini-cli/8.2-approval-modes), [Gemini config](https://google-gemini.github.io/gemini-cli/docs/get-started/configuration.html)
- [Aider options](https://aider.chat/docs/config/options.html), [Aider scripting](https://aider.chat/docs/scripting.html)
- [OpenCode CLI](https://opencode.ai/docs/cli/), [OpenCode permissions](https://opencode.ai/docs/permissions)
