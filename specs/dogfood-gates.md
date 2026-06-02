# Dogfood Gates

Updated: 2026-05-15

## Purpose

This document is a standing release gate for any change that touches:

- install
- first-run setup
- provider, route, model, or effort configuration
- launcher copy
- `provider doctor`
- shell usage guidance
- homepage, install page, simple guide, CLI guide, or state guidance

Jini does not pass these slices by being internally correct. It passes only if
simulated end users can succeed without translating Jini's internals for
themselves.

If a slice fails one of these persona gates, it should not ship until the
failure is fixed or an explicit exception is recorded.

The standing persona catalog lives in:

- [dogfood-personas.yaml](./dogfood-personas.yaml)

## Persona Panel

Install, configuration, and usage work should always be reviewed against the
full panel below:

### Beginner Trust

1. Low-literacy first-time user
2. Pragmatic “just make it work” user
3. College student user
4. High school student user
5. Household manager user

### Platform And Policy

6. AWS Bedrock user
7. Enterprise Azure user
8. Claude user
9. Codex user
10. ChatGPT user
11. Gemini user

### Expert Operator

12. Power user
13. Software engineer user
14. Hardcore developer
15. AI engineer
16. QA tester
17. Architect user

### Product And Executive

18. AI PM
19. Software VP user

### Domain-Specific

20. Realtor user
21. Travel advisor user

These are not optional research personas. They are merge blockers for install,
configuration, and usage work.

## Shared Rules

- One recommended install command must be obvious.
- One recommended first run path must be obvious.
- `jini` must remain the normal front door after install.
- `Auto` must be explained in plain language, not only as system state.
- If a stronger or cheaper route is chosen, Jini must say what it chose and
  why.
- If a model is chosen, Jini must say which model it chose and why.
- If an effort level is judged, Jini must show the level in a way users can
  understand.
- If a strict route matters, the docs must say how to force it.
- Safety and storage behavior must be stated plainly before asking for secrets.

## Persona Gates

### A. Beginner Trust Gate

Pass only if all are true:

- One install command is clearly labeled as the normal path.
- The next exact command after install is clearly `jini`.
- `Auto` is explained as “Jini picks for you” before any tool/provider/model theory.
- One exact first sentence to type into Jini is shown consistently.
- It is clear that Jini will not send anything before review.
- The user does not need env vars to get started.

Fail if any are true:

- The user must choose between `curl ... | bash` and `./install.sh` without guidance.
- The user sees three different `auto` knobs before first success.
- The user must read “shell,” “provider doctor,” “route,” or “effort” before
  they know what to type.

### B. Platform And Policy Gate

Pass only if all are true:

- There is one obvious strict path for Bedrock, Azure, Claude, Codex, ChatGPT,
  and Gemini-style users where relevant.
- The difference between forced route and auto mode is explicit.
- If a route or model matters, the docs explain when Jini will stay on it and
  when it may not.
- Jini promises a visible route readout before real work starts.
- The route readout includes tool, provider, model, and reason.
- The docs honestly state what `provider doctor` verifies and what it does not.
- Azure-only guidance is explicit for policy-constrained users.

Fail if any are true:

- “Prefers X” is used where the user actually needs “will use X unless you
  choose otherwise.”
- The route readout omits tool, provider, model, or reason.
- Auto mode is presented as the safest enterprise path when policy requires a
  fixed route.

### C. Expert Operator Gate

Pass only if all are true:

- Power users can still force a route, inspect what was chosen, and understand
  why.
- Developers can predict the cheapest-vs-best routing behavior from the docs.
- QA users can derive pass/fail expectations from the surface without reading
  implementation code.
- Architects and AI engineers can see that tool/model/effort selection is a
  core policy, not hidden behavior.

Fail if any are true:

- Jini hides a heuristic that materially changes route, model, or effort.
- Auto mode behaves like magic instead of a visible policy.

### D. Product And Executive Gate

Pass only if all are true:

- An AI PM can explain the first-run path in one short paragraph.
- A software VP can see how Jini keeps cost discipline while preserving quality
  escalation for deeper work.
- The docs make the product promise clearer than the system internals.

Fail if any are true:

- The setup story is internally correct but externally fragmented.
- Cost, confidence, and escalation behavior are not legible.

### E. Domain-Specific Gate

Pass only if all are true:

- Domain users such as a travel advisor can understand the first-run path
  without software vocabulary.
- Domain users such as a realtor can trust offline capture, follow-up, and
  privacy boundaries without learning provider jargon.
- Examples stay grounded in real outcomes, not system terminology.

Fail if any are true:

- Domain users need platform jargon before they can judge whether Jini helps
  them.

## Required Outputs For Persona Reviews

Each dogfood review should return:

- top blocking confusions
- top strengths
- explicit pass/fail per persona gate
- concrete recommended wording or flow changes

## Mechanical Merge Gate

For any install/config/usage slice, answer all of these before merge:

- [ ] Did the full standing persona panel review the current slice?
- [ ] Did each persona group produce pass/fail criteria rather than only opinions?
- [ ] Did the docs converge to one recommended first-run path?
- [ ] Is `Auto` explained in plain language?
- [ ] If Jini saves secrets or routing state, does the user learn that in plain language?
- [ ] If a route is auto-chosen, does Jini show what it chose and why?
- [ ] If a model is auto-chosen, does Jini show what it chose and why?
- [ ] If an effort level is auto-judged, does Jini show the level clearly?
- [ ] Are strict-route users told how to force their route?
- [ ] Are unresolved persona failures recorded as explicit non-goals for the slice?

## Current Findings To Preserve

The current standing persona findings are:

- the first-run story can still drift across README and docs unless reviewed explicitly
- low-confidence users still get confused when too many “first things” appear
- strict-route users do not trust “prefers” language when they need guarantees
- Azure users need explicit Azure-only guidance and plain deployment wording
- “saved in `.jini`” is too vague unless storage and trust boundaries are stated clearly
- auto-selected tool, model, and effort need visible reasons or advanced users will treat the product as magical

These findings should be assumed open until the docs and product copy clearly
remove them.
