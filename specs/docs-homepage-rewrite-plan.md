# Docs Homepage Rewrite Plan

## Goal

Rewrite the public docs homepage so it feels like a product front door, not an
internal docs tree.

The new homepage should help a first-time visitor answer five questions fast:

1. What is Jini?
2. Why would I use it instead of a raw model tool?
3. How do I install it?
4. What do I type first?
5. Why should I trust the cost, routing, and stored state?

## Reference Shape

Use the Buddy site as a structure reference, not a branding reference:

- short hero
- one install command
- one quickstart path
- visible compatibility strip
- plain-language trust section
- command cheat sheet
- concrete output section

Jini should keep its own product posture:

- one shell
- cheapest suitable route by default
- visible artifacts and readiness
- local-first where suitable
- measurable route and command evidence

## First-Pass Changes

### Homepage

The homepage should be reorganized into:

1. Hero
   - one-sentence value proposition
   - primary install CTA
   - secondary quickstart/examples CTA
   - install command

2. Works With
   - Claude Code
   - Codex
   - Bedrock
   - Azure OpenAI
   - Local models

3. Quickstart
   - install
   - run `jini`
   - paste the work

4. Why Jini
   - one shell
   - cheaper by default
   - resumable work
   - outputs you can use

5. Commands
   - `jini`
   - `jini setup`
   - `jini doctor`
   - `jini status`
   - `jini open`
   - `jini metrics`

6. What Jini Writes
   - deliverables
   - work state
   - route evidence
   - readiness signals

7. Trust And Cost
   - what is stored
   - what is shown before send
   - why route choice is visible
   - where measured metrics live

### Navigation

Rename visible nav labels to be more product-facing:

- `Simple Guide` -> `Quickstart`
- `What Jini Shows` -> `Outputs`
- `Proof` -> `Trust`

File paths can stay stable for now.

### Supporting Pages

Light-touch alignment only in the first pass:

- update page titles/descriptions to match the new labels
- update CLI wording where metrics still sounds proxy-only
- keep examples and install pages largely intact

## Distribution Notes

NuGet is not part of this first pass and should not be added just for badge
value.

NuGet only becomes worthwhile if Jini grows a real .NET or PowerShell-native
distribution story. Current higher-ROI distribution paths remain:

- Homebrew
- pipx / PyPI
- winget / scoop

## Out Of Scope

Do not do these in the first pass:

- redesign the whole visual theme
- add a new docs generator
- add a separate marketing site
- add screenshots that do not already exist
- add package channels that are not real install paths

## Success Criteria

The first-pass homepage is successful when:

- the first screen explains value and install in under 10 seconds
- the command story is stable and canonical
- compatibility is visible without jargon overload
- trust/cost signals are visible without reading deep docs
- the site feels smaller and more buyable
