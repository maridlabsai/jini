# Friction Reduction Research

## Purpose

This research defines how Jini removes user friction compared with Codex,
ChatGPT, and Claude without copying their product shapes blindly. The goal is
not to make Jini look like any one competitor. The goal is to make Jini feel
lighter to start, easier to resume, safer to trust, and richer when work becomes
too complex for plain terminal text.

## Source Set

The research uses current public product documentation and help materials:

- OpenAI Codex plan and cloud task docs: https://help.openai.com/en/articles/11369540/
- OpenAI Codex cloud docs: https://platform.openai.com/docs/codex
- OpenAI Codex mobile continuation: https://openai.com/index/work-with-codex-from-anywhere/
- ChatGPT projects: https://help.openai.com/en/articles/10169521-using-projects-in-chatgpt
- ChatGPT canvas: https://help.openai.com/en/articles/9930697-what-is-the-canvas-feature-in-chatgpt-and-how-do-i-use-it
- ChatGPT tasks: https://help.openai.com/en/articles/10291617-scheduled-tasks-in-chatgpt
- ChatGPT memory sources: https://help.openai.com/en/articles/10303002-how-does-memory-use-past-conversations
- Claude Code overview: https://docs.anthropic.com/en/docs/claude-code/overview
- Claude Code slash commands: https://docs.anthropic.com/en/docs/claude-code/slash-commands
- Claude Code hooks: https://docs.anthropic.com/en/docs/claude-code/hooks
- Claude Code memory: https://docs.anthropic.com/en/docs/claude-code/memory
- Claude artifacts: https://support.anthropic.com/en/articles/9487310-what-are-artifacts-and-how-do-i-use-them
- Claude AI-powered artifacts: https://support.anthropic.com/en/articles/11649438-prototype-ai-powered-apps-with-claude-artifacts

## Competitive Lessons

### Codex

Codex reduces friction by meeting developers where they already work, then
making task delegation portable. The core lesson is not "be another coding
agent." The lesson is that users adopt faster when local work, cloud work,
reviews, approvals, and remote check-ins feel like one continuous loop.

Jini should consume this as a continuity requirement:

- start from one command and one obvious path
- keep the user's repo or work context attached without forcing re-entry
- let users inspect progress and approve risky steps from any surface
- preserve a pull-down path from delegated work back to local work
- expose model/route choice only when it changes cost, latency, or risk

### ChatGPT

ChatGPT reduces friction by making the first action conversational, then moving
larger work into persistent spaces and richer editors. Projects keep context,
files, custom instructions, and memory together. Canvas makes writing and code
editable, section-addressable, restorable, and separate from the chat stream.
Tasks move appropriate work out of the active session and notify the user later.
Memory sources make personalization more legible.

Jini should consume this as a surface escalation requirement:

- do not force every input into a work template
- promote complex or reusable outputs into artifacts instead of long text dumps
- keep project/work memory scoped and inspectable
- offer follow-up work as resumable tasks when waiting is useful
- show why context was used when the answer depends on prior work

### Claude

Claude Code reduces coding friction through terminal-native immediacy,
slash-command discoverability, memory files, hooks, MCP, permissions, and
project/user/enterprise scoping. Claude artifacts reduce creative and planning
friction by moving substantial standalone content into a dedicated side panel
with reuse, versioning, publishing, remixing, and rapid iteration.

Jini should consume this as a command and artifact ergonomics requirement:

- keep help and advanced controls discoverable only when needed
- support reusable commands without making them mandatory for beginners
- use hooks and setup doctors to self-heal environment problems
- switch from chat text to artifact surfaces for standalone content
- make version, share, publish, and rollback state visible before risky action

## Jini Friction Principles

### 1. One Prompt Before Taxonomy

The first user input must be accepted in plain language. Jini can infer intent,
ask one or two clarifying questions, or start a small useful pass. It must not
force the user to choose a pack, route, model, persona, or schema before value
is visible.

### 2. No Empty-Shell Noise

Starting `jini` without a task should not print a stale default work summary.
The shell should show a compact prompt and defer full help, examples, setup
status, and route details until the user asks or a problem requires it.

### 3. Useful Result First

When enough scope exists, the first output should contain the useful result
before status metadata. State, route, cost, safety, and missing-context
sections are supporting context, not the lead.

### 4. Artifact Escalation

When content is substantial, reusable, editable, or multi-part, Jini should
create an artifact envelope that a richer surface can render as a document,
itinerary, checklist, decision table, code review, timeline, map, or task board.
The terminal can summarize, but it should not be the only experience.

### 5. Continue Anywhere

Work state should be addressable across CLI, desktop, mobile, and commercial
sync surfaces. "Continue" should not require remembering a file path, command,
conversation ID, or model route.

### 6. Setup That Fixes Itself

Every common first-run failure should map to a human-readable doctor check and
a next action. PATH problems, missing runtime, unavailable provider, missing
token, blocked accelerator, unsupported OS version, and stale cache should be
diagnosed before the user has to search docs.

### 7. Visible Trust Without Repeated Ceremony

Jini should require confirmation for risky actions, but it should not ask for
the same obvious permission repeatedly. Trust decisions should be scoped,
inspectable, expirable, and reversible.

### 8. Best Productivity With Least Expense

Route selection must optimize for productivity per unit cost. Local and cheaper
routes should be preferred for drafting, extraction, summarization, and other
low-risk work. Deep or premium routes should be used when quality, safety,
context size, or external action risk justifies the cost.

## Required Product Workstreams

### A. First Minute

The first-minute path must be benchmarked against Codex, ChatGPT, and Claude:

- install or launch
- first successful task
- first clarification
- first artifact
- first safe external action
- first resume
- first provider/local-route diagnosis

Exit criterion: a new user can reach a useful result without learning internal
Jini terminology.

### B. Natural Intake

Jini must maintain a generic intake layer that detects intent and uncertainty
without use-case-specific hard coding. Travel, meetings, research, vendors, and
code are profiles on top of a shared semantic envelope, not separate command
trees.

Exit criterion: greeting, underspecified request, scoped request, pasted notes,
and complex multi-artifact request all choose different interaction shapes.

### C. Rich Surfaces

The CLI remains the universal front door, but detailed work should graduate to
artifact surfaces. Public Jini can emit local artifacts and smart links.
Commercial Jini can add sync, native apps, mobile continuation, subscription
state, and hosted artifact collaboration.

Exit criterion: the same work object can render as terminal summary, markdown
artifact, desktop artifact, and mobile review card.

### D. Cross-Surface Resume

Jini needs a single work identity that can be resumed from any supported
surface. Resume should bring back the goal, ready artifacts, current blockers,
route/cost posture, last action, next safe action, and pending approvals.

Exit criterion: "continue" works without requiring a path when one active work
item is obvious, and asks a minimal chooser when multiple active items exist.

### E. Cost And Route Transparency

Users should not need to understand provider plumbing to use Jini, but they
must be able to inspect route decisions. Route labels should be API/provider
or capability labels, not marketing shorthand.

Exit criterion: every route decision can explain model/API/provider, why it was
chosen, cheaper fallback, stronger fallback, and local/offline fallback when
available.

## What Jini Should Not Do

- Do not copy Codex by becoming only a coding shell.
- Do not copy ChatGPT by hiding all structure inside opaque chat.
- Do not copy Claude by turning every advanced feature into slash-command
taxonomy.
- Do not add native app surfaces before the work object and artifact envelope
are stable.
- Do not show cost, model, provider, or state metadata before the first useful
result unless the user asks or safety requires it.

## Measurement

Jini should track friction with product metrics, not only implementation
presence:

- time to first useful result
- commands before first useful result
- clarification quality score
- stale-context recovery rate
- artifact open/reuse rate
- resume success rate
- setup-doctor self-recovery rate
- repeated-permission prompt rate
- local/cheap route adoption rate
- premium-route regret rate after user edits

## Research Verdict

Jini's best posture is a lightweight work-continuity layer above models and
surfaces. Codex is strong at developer execution, ChatGPT is strong at everyday
conversation plus persistent spaces, and Claude is strong at terminal-native
coding plus rich artifacts. Jini wins only if it reduces switching cost across
all three patterns: natural intake, durable work state, artifact surfaces, and
least-expense route selection.
