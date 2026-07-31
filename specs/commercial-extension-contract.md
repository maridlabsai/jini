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

## Known gaps (extend when the commercial repo starts)

- The current seam expresses Hold vs Decline. **Route-switching mechanics**
  (actually re-attempting on a different route mid-survival) and **dodge
  recording** into the ledger's `ThrottleDodged`/`Dodges` are not yet wired
  through the survival loop — the `Decision.Dodged` field exists in the
  contract but the public loop does not yet act on a switch. Add a
  switch-and-record mechanism to `runWithThrottleSurvival` (behind the same
  registered-strategy seam) as the first commercial-integration task.
