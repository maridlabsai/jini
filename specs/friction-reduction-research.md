# Friction Reduction Research

## Purpose

This research defines how Jini removes user friction compared with Codex,
ChatGPT, and Claude without copying their product shapes blindly. The goal is
not to make Jini look like any one competitor. The goal is to make Jini feel
lighter to start, easier to resume, safer to trust, and richer when work becomes
too complex for plain terminal text.

This research should be read through one stricter charter:

- Jini is a cost optimizer first
- Jini UX must be second to none
- Jini must preserve one resumable session across macOS, Windows, mobile, and
  CLI

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

## UX Research And Design Anchor

UX research and UX design should reject any surface that fails one of these
tests:

- it does not make the cheapest suitable path feel like the normal path
- it makes users learn product theory before they can act
- it makes device switching feel like starting over
- it makes review, readiness, or next action harder to understand
- it improves internal neatness more than user outcome

The design target is not "feature rich AI UX." The design target is:

- cheaper to finish
- easier to understand
- easier to resume
- safer to hand off

That anchor should govern:

- first-minute research
- screen-order decisions
- copy choices
- continuation design
- cross-surface parity decisions

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

Compatibility requirement: common tactical commands from adjacent AI tools must
be accepted as low-risk aliases. `/help`, `/status`, `/doctor`, `/model`,
`/init`, `/memory`, `/permissions`, and `/cost` should explain state, setup,
memory, permission, or route posture without creating a `First Useful Pass` or
mutating current work. These surfaces should stay compact and have benchmark
coverage because they are used repeatedly during ramp-up and recovery.

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

## Prompt Surface Comparison

The important comparison is not "which product has the prettiest prompt." The
important comparison is where each product asks the user to carry product
theory instead of simply stating the work.

| Dimension | Claude Code | ChatGPT | Codex | Jini today | Jini gap |
| --- | --- | --- | --- | --- | --- |
| First turn | Terminal-native task intake with advanced control living behind commands, memory files, and skills | Conversational first turn; projects and canvas appear when work needs durable context or editing | Natural task intake across CLI, app, web, and mobile-connected continuation | Compact paste-first shell exists, but richer launcher still exposes more scaffolding | Jini still explains itself too early |
| Clarification | Typically asks only when blocked; repo and tool context already carry a lot | Clarifies selectively, then keeps the conversation flowing | Clarifies around approvals, repo state, or task direction while preserving active work state | Profile-based clarification can be helpful, but it is still partly template-driven | Clarification is not yet driven by one generic scope engine |
| Tactical help | Commands, skills, memory, and hooks are discoverable without blocking first work | Tools are present, but everyday use starts from chat | Coding surfaces expose approvals, clients, plugins, and active work without making them mandatory upfront | `/help`, `/status`, `/doctor`, `/memory`, and related aliases are safe and tested | Good parity here, but the launcher still teaches too much before the task |
| Rich editing | Artifacts, skills, IDE, desktop, and browser workflows | Canvas, projects, saved project sources, tasks, and memory | Diffs, screenshots, approvals, terminal output, and mobile review/approval loops | Artifact envelope direction is right, but many flows still resolve to markdown-first terminal behavior | Artifact escalation is still narrower and more implicit than competitors |
| Resume and continuity | Memory files, auto memory, remote control, desktop, web, and IDE | Projects keep chats, files, instructions, and project memory together | Active threads, approvals, plugins, and project context flow across machines and phone | Current-work resume is strong in CLI, but cross-surface identity is still mostly architectural intent | Jini continuity is ahead conceptually, behind in shipped surface depth |
| Context visibility | Project memory and CLAUDE.md are explicit, inspectable, and layered | Project memory, project sources, and saved responses are visible containers | Connected machines keep local files and credentials in place while live state syncs to other surfaces | Work state is visible, but pack/profile structure still leaks through the experience | Jini needs user-facing context containers, not internal pack vocabulary |
| Route and model visibility | Available when useful, not mandatory before task intake | Largely abstracted unless tools/settings matter | Model, approvals, plugins, and client surfaces are inspectable during active work | Public launcher still prints `Working with` immediately in richer help modes | Route/provider labels still appear too early |

## Prompt Archetype Comparison

The fastest way to see prompt adoption gaps is to compare what a user can type
without learning the product first, and how much product shape comes back
before the first useful move.

### 1. Greeting Or Tiny Social Input

User input examples:

- `hello`
- `hi`
- `good morning`

Observed competitor pattern:

- ChatGPT stays conversational.
- Claude stays conversational unless the surrounding IDE or terminal context
  makes a work request obvious.
- Codex stays minimal and usually waits for concrete work rather than
  manufacturing a durable artifact.

Jini target:

- greet back briefly
- do not create work
- do not print a framework
- do not introduce route, provider, pack, or artifact language

Current Jini deviation:

- this has improved, but any regression here is extremely damaging because it
  teaches users that Jini is a workflow engine before it proves it is an
  assistant

### 2. Vague Task Intake

User input examples:

- `plan me a trip`
- `help me with this`
- `I need to follow up on a meeting`

Observed competitor pattern:

- ChatGPT usually asks one or two clarifying questions or offers a short
  scaffold while keeping the tone conversational.
- Claude often asks for the missing high-yield context and does not invent a
  large template unless the user asks for one.
- Codex tends to ask for repo, file, or task specifics when the action surface
  is ambiguous.

Jini target:

- detect uncertainty
- ask only the minimum high-yield questions
- avoid printing a domain template before scope is understood
- prefer a generic scope planner over profile-specific questionnaires

Current Jini deviation:

- clarification quality is better than before, but it is still too visibly
  shaped by starter profiles

### 3. Already-Scoped Request

User input examples:

- `plan a 7 day Paris trip for two adults in October with a mid-range budget`
- `turn these meeting notes into a sendable follow-up`
- `summarize this doc for the leadership team in bullet points`

Observed competitor pattern:

- ChatGPT drafts immediately when the ask is sufficiently scoped.
- Claude drafts or acts immediately when context is present.
- Codex proceeds directly when the instruction is concrete and the execution
  path is clear.

Jini target:

- draft immediately
- avoid follow-up questions unless the missing detail materially changes the
  result
- keep route and state metadata behind the result, not ahead of it

Current Jini deviation:

- Jini is closer here, but it still sometimes over-values additional scoping
  instead of rewarding the user for already doing the work of being specific

### 4. Tactical Command Or Tooling Input

User input examples:

- `/help`
- `/doctor`
- `/memory`
- `show what's ready`

Observed competitor pattern:

- Claude exposes advanced controls in slash commands, but they do not become
  required in normal use.
- Codex exposes active-work, model, approval, and tooling state when asked.
- ChatGPT rarely makes tactical controls the main entry posture.

Jini target:

- accept adjacent-tool command habits
- answer tactically
- never create fake work
- keep command output compact and benchmarked for latency

Current Jini deviation:

- this is one of Jini's stronger competitive surfaces, but it must stay compact
  and must not bleed back into the launcher

### 5. Revise This / Transform This

User input examples:

- `make this shorter`
- `turn this into a checklist`
- `make it more executive`

Observed competitor pattern:

- ChatGPT Canvas and Claude Artifacts make revision feel natural and local to
  the current artifact.
- Codex keeps transformation attached to the active diff, output, or artifact
  instead of reopening the whole intake experience.

Jini target:

- treat revision as an edit of the current artifact
- preserve versions
- expose undo
- keep the transform local, not as a restart of the workflow

Current Jini deviation:

- the product direction is now correct, but artifact revision still needs to
  feel more native than command-driven

### 6. Resume / Continue

User input examples:

- `continue`
- `resume this`
- `what's left`

Observed competitor pattern:

- Codex is strongest here because active work, approvals, terminal state,
  screenshots, and diffs travel together across surfaces.
- ChatGPT Projects are strong for persistent knowledge work and files.
- Claude memory and project context reduce restart cost, but continuation is
  still more environment-specific than Codex.

Jini target:

- continue obvious work immediately
- show blockers, ready artifacts, and next safe action
- avoid making the user remember file paths, IDs, or internal work terms

Current Jini deviation:

- CLI continuation is strong, but cross-surface continuation is still more
  promise than shipped experience

### 7. Long-Running Or Multi-Artifact Work

User input examples:

- `compare these vendors and give me a memo plus shortlist`
- `review this PR and draft the team update`
- `plan the trip and make a packing list`

Observed competitor pattern:

- ChatGPT escalates long-form work into projects, files, canvas, and tasks.
- Claude escalates substantial standalone outputs into artifacts.
- Codex keeps active execution state attached while exposing outputs, diffs,
  logs, and approvals.

Jini target:

- split work into a shared work object with multiple artifacts
- keep the shell concise
- provide smart links into the richer outputs
- make "what is ready", "what is blocked", and "what changed" inspectable on
  demand

Current Jini deviation:

- Jini still leans too heavily on markdown-first terminal rendering for work
  that deserves richer artifact-first presentation

### What The Official Product Patterns Say

- Claude Code documents a terminal, IDE, desktop, and browser posture, then
  layers memory via `CLAUDE.md`, auto memory, and optional skills instead of
  requiring them before the first request.
- ChatGPT projects document a persistent workspace for chats, files,
  instructions, and project memory, while canvas is a separate editing surface
  that can open when the work becomes longer or more revision-heavy.
- Codex documents one account across app, CLI, IDE extension, web, and mobile
  continuation. The mobile continuation write-up is especially important:
  active work, approvals, screenshots, terminal output, diffs, and context stay
  attached to the running environment instead of being re-entered manually.

Jini should consume these as interaction principles, not mimicry targets:

- natural language first
- durable context second
- rich editing when complexity rises
- inspectable control surfaces on demand
- seamless continuation across devices and sessions

## Current Jini Deviations

### Adoption-Positive

- The compact `jini` shell is materially lighter than the earlier default state
  dump and now behaves more like a paste-first front door.
- Tactical inputs such as `/help`, `/status`, `/doctor`, `/memory`, and
  punctuated variants are already treated as non-work control inputs with
  regression coverage.
- Greeting-only input now stays conversational instead of forcing a fake work
  object.
- Jini already has the right instinct around "least expense first" routing and
  durable work identity.

### Adoption-Negative

- The richer launcher still opens with a teaching surface rather than a plain
  work box. `What do you need help finishing?`, `Working with`, examples, and
  command lists are still more instructional than Claude, ChatGPT, or Codex.
- `Working with` and provider labels still leak before the user has received
  value. Competitors generally reveal this kind of infrastructure only when it
  matters for action, cost, or trust.
- `help me finish this` and `First Useful Pass` are still internal Jini
  concepts. They may be reasonable implementation concepts, but they are not
  concepts users of Claude, ChatGPT, or Codex arrive expecting.
- Starter packs still encode visible product behavior through use-case profiles,
  hard-coded detect signals, hard-coded headings, and in some cases hard-coded
  destination links. That creates useful demos, but it does not create the
  feeling of a broadly capable assistant.
- Clarification is still profile-specific. Travel shows the issue clearly:
  there is a reasonably good scoped-question path, but it is still built as a
  travel profile rather than a generic scope planner that happens to work for
  travel.
- Hidden provider prompts still over-shape the work through pack-specific
  guidance. This improves consistency, but it also risks making output feel
  canned when users expect a more adaptive assistant.
- Jini still has more visible internal taxonomy than the competitors. The
  target experience should be "stateful assistant with good artifacts," not
  "starter-pack workflow engine."
- Jini still lacks one shared prompt contract across CLI, desktop, and mobile.
  Competitors differ in surface design, but they all preserve a more stable
  mental model for what `continue`, `revise`, `help`, and a scoped ask will do.

## Major Work Items To Bridge The Gap

### 1. Replace Pack-First Intake With A Shared Work Envelope

Introduce one generic intake and planning schema:

- user goal
- current material
- output intent
- missing high-value scope
- risk level
- artifact need
- continuation need

Travel, follow-up, research, vendor choice, and code should be profile lenses
on top of this envelope, not primary branching structures in the first-turn
experience.

### 2. Demote Product Teaching In The First Minute

Default launcher behavior should move closer to:

- one short invitation to paste the work
- optional help only when requested
- no provider/route label before the user asks or a safety/cost issue requires
  it
- no example wall unless the user is blocked

This is the single largest adoption gap because it affects every new user and
every context switch.

### 3. Remove Internal Product Terms From User-Facing Default Flows

`First Useful Pass` can remain an internal artifact class if needed, but it
should not be a headline concept in the user journey. Users should see artifact
names that match the work itself: itinerary, follow-up, decision memo, bug
brief, handoff note, task list.

### 4. Generalize Clarification Into A Scope Planner

Replace profile-specific clarification builders with a generic planner that:

- detects what is already known
- ranks missing inputs by output value
- asks only the smallest set of high-yield questions
- decides when a generic first draft is better than more questioning
- works across travel, planning, writing, research, code, and mixed requests

### 5. Generalize Artifact Planning And Smart Linking

Artifact shaping should come from a shared artifact planner and capability
registry, not from pack-local heading scripts and destination-specific smart
links. Domain packs may contribute optional hints, but the base planner should
own:

- artifact type selection
- section planning
- link preservation rules
- readiness and missing-input summaries
- cross-surface renderability

### 6. Make Continuation Feel Native, Not Bolted On

Codex and ChatGPT show the value of seamless continuation across clients and
devices. Jini should keep work identity stable and allow:

- current work resume
- lightweight mobile/desktop review
- artifact review before share
- approval and clarification check-ins without reopening the whole workflow

### 7. Benchmark Prompt Friction As A First-Class Quality Signal

The regression suite should compare Jini against prompt classes, not only local
goldens:

- greeting
- vague request
- already-scoped request
- revise-this artifact request
- tactical command
- continue/resume request
- long-running delegated work check-in

The goal is not to clone competitor wording. The goal is to keep Jini's
required user effort at or below the benchmark set by Claude, ChatGPT, and
Codex for the same class of ask.

### 8. Prune Demo-Grade Hard Coding From Core Runtime

Hard-coded profile signals, destination-specific links, and pack-specific
section templates should be treated as temporary scaffolding unless they can be
proven to generalize. Anything that survives in core must justify itself as a
domain-agnostic capability or as a clearly isolated optional profile.

### 9. Establish One Cross-Surface Prompt Contract

The same classes of input should behave the same way across CLI, desktop, and
mobile review surfaces:

- greeting stays conversational
- vague ask triggers minimal scoping
- scoped ask drafts immediately
- revise transforms the active artifact
- continue resumes active work
- tactical input reveals control surfaces without mutating work

This is the cleanest way to reduce ramp-up friction for users coming from
Codex, ChatGPT, or Claude, because it removes surprise rather than adding more
documentation.

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
- percent of already-scoped requests that draft on first response
- clarification quality score
- clarification burden per vague request
- stale-context recovery rate
- artifact open/reuse rate
- resume success rate
- setup-doctor self-recovery rate
- repeated-permission prompt rate
- local/cheap route adoption rate
- premium-route regret rate after user edits
- prompt archetype parity score across greeting, vague ask, scoped ask, revise,
  continue, and tactical help

## Research Verdict

Jini's best posture is a lightweight work-continuity layer above models and
surfaces. Codex is strong at developer execution, ChatGPT is strong at everyday
conversation plus persistent spaces, and Claude is strong at terminal-native
coding plus rich artifacts. Jini wins only if it reduces switching cost across
all three patterns: natural intake, durable work state, artifact surfaces, and
least-expense route selection.
