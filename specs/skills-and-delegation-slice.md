# Skills And Delegation Boundary

Updated: 2026-06-07

## Purpose

This public spec is a boundary handoff for skills and delegated work in Jini.
It is not a free-tier implementation slice.

## Tier Boundary

Commercial tier owns the agent and skills OS productivity suite.

Free tier must not ship `skills`, `delegate`, developer agents, tester agents,
or a skills-based OS productivity suite.

## Free Tier Rule

The free tier can keep natural intake, routing, local/BYO model setup, context
compression, basic artifacts, and useful continuation.

The free tier must not include the file-backed skill runtime,
specialist developer/tester helpers, explicit delegation commands,
organization skill libraries, or commercial productivity-learning automation.

## Commercial Handoff

Implementation of the full suite belongs in the commercial tier.

Commercial Jini may define:

- skills and delegation commands
- project-scoped and user-scoped skill libraries
- developer agents and tester agents
- work-attached run records
- approval, audit, and rollback controls
- paid productivity learning from repeated workflows

The public repo should only preserve the product boundary and avoid free-tier
surface leakage.

## UX Contract

This is a P1 simplification slice, not a reason to teach another interaction
model. Natural `jini` intake remains the default path.

Commercial helper capability can be mentioned through progressive-disclosure
controls for users who need them. Basic free-tier capability suggestions must
stay plain-language and must not require users to learn commercial command
vocabulary.

The default user path remains:

- `jini`
- paste the work
- get a useful result

Free-tier users should not need skills.

Commercial Jini can suggest a relevant capability in plain language, such as
"Use reviewer on this draft?", without requiring the user to learn the word
`delegate` first.

Commercial agent output must look like normal Jini work:

- a named artifact
- a short result summary
- what changed
- what still needs attention

It must not show agent trees, role theater, or step-by-step orchestration logs
by default.

## Public Regression Rule

Reject any public/free change that:

- adds `skills` or `delegate` as a free-tier command
- defines file-backed skill discovery as a free-tier runtime requirement
- adds developer agents or tester agents to the free tier
- makes skills or agent vocabulary a prerequisite for normal use
- exposes visible agent trees or orchestration logs by default

## Success Criteria

This boundary is preserved when:

- the free tier remains useful through natural `jini` intake, local/BYO routes,
  context compression, artifacts, and continuation
- the commercial repo owns the agent and skills OS productivity suite
- public docs and readiness gates do not describe the commercial suite as a
  free-tier runtime requirement
