# Launcher Intake Design

Updated: 2026-05-14

## Goal

Make `jini` the real first-run product:

1. run `jini`
2. paste what needs to be finished
3. answer only the highest-value blocking question when needed
4. get a first useful artifact fast
5. keep missing information and next step visible without overwhelming the user

## Public launcher rules

- default to paste-first input
- accept rough or messy context
- avoid exposing tool, provider, or model theory before value appears
- ask at most one high-impact clarification before first output when the request
  is too underspecified to produce a good first draft
- produce the artifact before the long explanation
- keep `Start new` and `Continue` as plain-language actions

## Current-work rules

When work already exists, startup should show a compact resume card:

- `Goal`
- `Working with`
- `AI route`
- `Up next`
- `Ready now`
- `Blocked`

Detailed rationale such as `Why this matters`, `Options`, or `If you skip this`
should remain available behind explicit actions like `Show missing` or
`jini check`.

## Starter design rules

- starter behavior must come from shared profile metadata, not scattered
  use-case branches
- clarification prompts should ask only for missing high-value dimensions
- starter output should reflect the parsed request shape instead of one default
  canned example

Use this stable path for public review and future links.
