# Product Streamline Redline

Updated: 2026-06-09

This document decides when Jini continues streamlining the current codebase and
when it stops patching and rewrites the product kernel from scratch.

## Research Anchors

Current competitor research supports one clear product shape:

- OpenAI Codex CLI is a local terminal coding agent that can read, change, and
  run code in the selected directory.
  Source: https://developers.openai.com/codex/cli
- Claude Code emphasizes real developer tools, explicit permission before file
  changes or commands, access boundaries, and human control.
  Source: https://www.anthropic.com/product/claude-code
- Gemini CLI is an interactive terminal REPL with a client/core/tools split for
  filesystem, shell, web, memory, and extension tools.
  Source: https://google-gemini.github.io/gemini-cli/docs/
- GitHub Copilot cloud agent shows the platform direction: research, plan,
  branch changes, logs, reviewable pull requests, custom agents, memory,
  hooks, MCP, and cross-surface continuity.
  Source: https://docs.github.com/en/copilot/concepts/agents/cloud-agent/about-cloud-agent
- GitHub's agent platform now runs Copilot, Claude, and Codex across GitHub web,
  GitHub Mobile, and VS Code with shared context, governance, memory, and audit.
  Source: https://github.blog/changelog/2026-02-26-claude-and-codex-now-available-for-copilot-business-pro-users/
- Aider reinforces the basics: terminal-first codebase mapping, git integration,
  local/cloud model optionality, and lint/test loops after edits.
  Source: https://aider.chat/

Design implication: Jini must not win by inventing a new conversation grammar.
It can win only by reducing friction across tools, routes, tokens, session
continuation, local/offline fallback, and release-quality gates.

## Design Alternatives

### A. Patch The Current Product

Small fixes continue without changing the architecture contract.

Why rejected: this is how the verbose shell, templates, and PRD drift survived
too long. It is fast per cut but expensive across cuts.

### B. Streamline The Current Kernel

Keep the current Go codebase but force every change through PRD, HLD, LLD,
golden transcripts, and competitor scorecards.

Why selected now: the shell, intent handlers, action handlers, state, and gates
are already separable enough to enforce the lightweight transcript. This path
ships faster than a rewrite while still creating a hard exit if boundaries fail.

### C. Rewrite The Product Kernel

Freeze feature work and rebuild the CLI kernel around shell, intent, action,
state, and gate boundaries from a blank implementation.

Why not selected today: a rewrite is justified only when the current code cannot
preserve the first-minute contract through localized changes and tests. Rewriting
before that proof risks losing GTM time without solving product confusion.

## Selected Approach

Jini continues with Approach B until a rewrite trigger fires.

The active architecture is:

1. Shell: prompt, commands, and receipts.
2. Intent: direct answers, local edits, route controls, current-work controls,
   and task creation.
3. Action: file edits, CLI handoffs, provider/local routes, and fail-closed
   prompts.
4. State: durable work, artifacts, memories, and route receipts.
5. Gates: transcript, PRD drift, scorecard, security, ship, and release checks.

The shell must stay independent of artifacts. The intent boundary must classify
simple questions and direct file edits before any starter pack or work artifact
can run.

## Rewrite Triggers

Stop streamlining and plan a rewrite if any trigger is true:

- Three P0 first-minute transcript incidents recur after a regression test was
  added for that class of incident.
- A simple question, direct file edit, or route-control change cannot be fixed
  without entering starter-pack or artifact-rendering code.
- A new feature needs a default shell frame, modal, or command grammar that a
  Claude Code or Codex user would not expect in the first minute.
- A code change cannot map to a PRD outcome, HLD boundary, LLD contract, and
  executable gate in one short trace.
- The golden transcript gate cannot express the expected behavior without
  brittle, implementation-shaped assertions.
- Product docs require exceptions to explain why the CLI behaves differently
  from the familiar terminal agents users already know.

## Release Redlines

Do not release if any of these fail:

- simple factual question transcript
- direct current-directory file edit transcript
- current-work passivity transcript
- configured CLI handoff or fail-closed setup transcript
- install smoke from release assets
- scorecard PRD completion
- ship check evidence

Architecture quality does not compensate for a bad transcript. A release that
looks clean internally but feels worse than Claude Code, Codex, Gemini CLI, or
Aider in the first minute is not releasable.

## Research Cadence

Every material product cut must do a lightweight competitor refresh:

- classify competitor movement as copy, integrate, watch, reject, or delete
- update the benchmark only when the movement changes a release-critical outcome
- reject any requirement that widens scope without improving the CLI wedge
- add or update a golden transcript before changing behavior
