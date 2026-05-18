# UX Cross-Functional Gate Review

Updated: 2026-05-17

## Scope

This review covers the current Jini install path, first-run launcher,
preflight decision card, and current-work shell.

The review was run through three role lenses:

- UX researcher
- UX designer
- application developer

The goal was to check whether Jini feels natural, easy to adopt from other
tools, low-friction, and explicit enough about the context that matters.

## Shared Blocking Findings Before Fixes

1. The launcher still sounded tool-shaped in places.
   - `Jini shell`
   - `If you need setup help, type Use Auto`
   - `If you are not sure, type help me finish this`

2. The decision card was transparent but too diagnostic in tone.
   - `Route policy`
   - `Why this was chosen`
   - `Change route`

3. The current-work action area showed too many commands at once.
   - explicit feedback verbs
   - model voting commands
   - too much visible operator surface for a normal user

4. The same idea was described with more internal wording than needed.
   - `Route policy` was less natural than `How chosen`
   - `Why this was chosen` was less direct than `Why this route`

5. Users coming from Claude Code, Copilot, ChatGPT, or Gemini would likely
   understand the workflow, but the visible copy still asked them to translate
   some platform language into product meaning.

## Changes Applied

1. Simplified launcher wording.
   - removed `Jini shell` from the first-run launcher
   - changed setup help to:
     - `Need setup help? Type Use Auto and Jini will help you connect the best available option.`
   - changed unsure help to:
     - `Not sure? Type help me finish this.`
   - changed `Good inputs` to `Examples`

2. Made the decision card more natural without hiding state.
   - `Route policy` -> `How chosen`
   - `Why this was chosen` -> `Why this route`
   - `Change route` -> `Want a different route?`

3. Reduced action clutter in current work.
   - replaced `Jini shell / What do you want to do?` with `Choose one`
   - kept the main actions as the visible list
   - grouped feedback under `Tell Jini how this draft went`
   - moved low-frequency power commands into one compact advanced line

4. Kept trust-critical state visible.
   - tool
   - provider
   - model
   - how chosen
   - why this route
   - continuity when Jini preserves the current coding route
   - verification level and reason

## Role Verdicts

### UX Researcher

Result: pass

Why:

- the first minute now feels more like relief than configuration
- setup help is available without leading with internal theory
- the visible copy is plainer and safer for low-confidence users
- the user still sees what is safe, what is missing, and what to do next

Remaining watch item:

- the decision card is still denser than the launcher and should stay compact

### UX Designer

Result: pass

Why:

- the screen order remains result-first
- the decision card is more human-readable
- the current-work action area is more clearly prioritized
- the interface keeps the important context visible without expanding the main
  action list further

Remaining watch item:

- feedback actions may eventually deserve a dedicated secondary affordance

### Application Developer

Result: pass

Why:

- the product remains inspectable and explicit
- the new wording is closer to the way users interpret modern AI tools
- continuity is now visible as a first-class reason instead of hidden in
  generic routing copy
- power commands still work without dominating the default path

Remaining watch item:

- strict-route and advanced routing concepts should remain easy to discover but
  secondary to the normal flow

## Gate Summary

### UX Researcher Gate

Pass

### UX Designer Gate

Pass

### Relevant Dogfood Gates

- Beginner trust: pass
- Expert operator: pass
- Product and executive: pass

## Evidence

The updated UX is reflected in:

- `internal/app/app.go`
- `internal/app/app_test.go`
- `internal/app/provider_test.go`
- `README.md`
- `docs/cli.md`
- `docs/simple.md`
- `docs/state-and-artifacts.md`

## Final Judgment

The product now clears this cross-functional gate for the current slice.

It is more natural for people coming from other tools, it keeps the critical
context visible, and it reduces unnecessary learning burden without hiding the
route/model/verification decisions that expert users need to trust.
