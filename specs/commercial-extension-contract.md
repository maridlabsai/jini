# Commercial Extension Contract

How the separate paid module (`../jini-commercial`) builds against this public
repo without importing its `internal/` packages. See
[[paid-autopilot-boundary]] for the tier decision this implements.

## Why a contract package exists

Go forbids a different module from importing another module's `internal/`
packages. So the extension seam cannot live in `internal/app`; it lives in
non-internal packages the commercial module can import:

- `github.com/maridlabsai/jini/autopilot` — the extension contract + registry.
- `github.com/maridlabsai/jini/runner` — a public entrypoint wrapper so the
  commercial module can build its own binary running the same core.

The public build never registers a strategy, so Autopilot is inert in it and
the free hold/self-resume behavior is unchanged. This split keeps MECHANISM +
seam public and POLICY (predictive avoidance, when/where to switch, savings
optimization) in the commercial module.

## What the commercial module implements

```go
package autopilotimpl

import (
    "context"
    "github.com/maridlabsai/jini/autopilot"
)

type Strategy struct{ /* entitlement client, route intelligence, etc. */ }

func (s Strategy) OnThrottle(ctx context.Context, e autopilot.ThrottleEvent) (autopilot.Decision, error) {
    // Paid policy: predictively avoid, switch routes, auto-resume, optimize
    // savings. Return Decision{Action: autopilot.Hold|Decline, Dodged: ...}.
}

func init() { autopilot.Register(Strategy{ /* ... */ }) }
```

## How the commercial binary is composed

```go
// cmd/jini-pro/main.go in ../jini-commercial
package main

import (
    "os"
    "github.com/maridlabsai/jini/runner"
    _ "example.com/jini-commercial/autopilotimpl" // init() registers the strategy
)

func main() {
    os.Exit(runner.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

At runtime the public core consults `autopilot.Registered()` at the throttle
point via `entitledAutopilotApprover()` (internal), engaging the strategy only
when a strategy is registered AND `currentSubscriptionTier() == "commercial"`.

## Current seam surface (this repo)

- `autopilot.Strategy` / `ThrottleEvent` / `Decision` / `Action` / `Register` /
  `Registered` — the contract.
- `runner.Run` — public entrypoint wrapper over `app.RunInteractive`.
- Internal bridge: `autopilotApprover` adapts a `Strategy` to the throttle
  approver seam; consulted in `configureThrottleApproverForEntry`.

## Switch-and-record mechanism (scaffolded)

The public loop now implements throttle-aware switching behind the seam, so the
commercial side is a pure policy drop-in:

- `autopilot.Decision.SwitchToRoute` names a fallback route; `Dodged` marks a
  dodge for the ledger.
- Internal: `autopilotApprover.ResolveThrottle` carries the switch target into
  `runWithThrottleSurvival`, which — on a granted switch — calls
  `throttleSurvivalOptions.switchAttempt` once, and on success returns the
  fallback answer WITHOUT waiting out the reset (a dodge), recording
  `report.Dodged`/`SwitchedTo`. A failed switch falls back to holding the
  original route. `recordSavingsOnDecision` flags the ledger entry
  `ThrottleDodged` and re-attributes it to the switched route.
- The switch executor `attemptOnRoute` re-routes via the proven
  `detectRouteForToolMode`/`enrichRouteDecisionForRequest` path (CLI handoff or
  provider), single-shot, no recursion.
- A pure public build registers no strategy, so `switchAttempt` is never
  invoked and the free hold/self-resume path is byte-identical (verified by the
  unchanged `TestRunWithThrottleSurvival*` suite).

## Remaining for the commercial repo

- The paid **policy** itself: predictive avoidance, which route to switch to
  and when, savings optimization — implemented as `autopilot.Strategy` in
  `../jini-commercial`.
- Refinement: dodge route/class attribution currently reuses the switch target
  label; the commercial integration can enrich pricing for the actual answering
  model if it differs.
