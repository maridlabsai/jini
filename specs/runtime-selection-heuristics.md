# Runtime Selection Heuristics

Updated: 2026-05-15

See also:

- [research-informed-heuristics.md](./research-informed-heuristics.md)

## Purpose

Jini should treat these as first-class runtime decisions on every request:

- which tool to use
- which model to use
- what effort level to apply
- whether a local commercial SLM should handle the request first
- which local SLM profile should handle the request when local is chosen

These are core product behaviors, not hidden implementation details.

They should be interpreted through the stricter product rules in
[research-informed-heuristics.md](./research-informed-heuristics.md), especially:

- hidden planning for multi-step work
- selective refinement instead of always-on critique
- selective consistency checks only for high-risk work
- connector-aware and cohort-aware route learning

## Decision Order

For each request, Jini should decide in this order:

1. classify the work type
2. classify the required depth
3. decide whether the local commercial SLM pool can handle it well enough
4. if local is suitable, choose the local profile
5. choose the cheapest suitable tool route by default
6. choose the model for that route
7. choose the effort level for that specific request
8. show the decision and save it with the work

## Tool Heuristic

Default rule:

- choose the cheapest suitable route for the work
- if a commercially usable local SLM can do the job well enough, it should be
  the default front line
- if local is chosen, Jini should still choose the right local profile rather
  than assuming one local model fits all work

Adjustment rule:

- if the request clearly asks for deeper, more rigorous, or more exhaustive
  work, switch to a stronger route
- if the local SLM would create too much quality risk, escalate visibly
- for coding work, add continuity bias and route-switch cost so Jini does not
  churn routes when the current route is still good enough
- for coding work, factor practical quota headroom and iteration economics, not
  only per-turn model strength
- for coding work, prefer stable continuation unless a materially better route
  justifies the context-switch cost

## Model Heuristic

Default rule:

- model choice follows the chosen route

Examples:

- `Claude Code` -> Claude Sonnet 4
- `Bedrock Sonnet` -> Claude Sonnet 4.6
- Azure-backed routes -> configured Azure deployment

Adjustment rule:

- explicit model constraints or route constraints win

## Effort Heuristic

Jini should judge one of these levels per request:

- `low`
- `medium`
- `high`
- `extra high`

Default rule:

- normal work -> `medium`
- quick/lightweight asks -> `low`
- deeper/rigorous asks -> `high`
- benchmark, architecture, root-cause, release-readiness, or exhaustive asks -> `extra high`

## Visibility Rule

Jini should keep these visible:

- chosen tool
- chosen model
- chosen effort level
- whether the local SLM was used or skipped
- which local profile was used when local was chosen
- why Jini chose them

## Persistence Rule

Once Jini chooses a route for a work item, it should save:

- tool label
- model label
- effort level
- selection reasons

This keeps later screens honest and prevents silent route drift.

For coding-oriented work, persistence should also support:

- continuity of the current route when appropriate
- remembered user override tendencies by coding cohort
- explicit route-switch reasons
