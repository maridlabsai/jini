// Package autopilot is the PUBLIC extension contract for paid Jini Autopilot.
//
// The paid Autopilot implementation lives in a separate commercial module
// (../jini-commercial), which cannot import Jini's internal/ packages. This
// package is the non-internal seam both sides share: the commercial module
// implements Strategy and calls Register in its process init; the public
// runtime consults Registered() at the throttle point, but only when the
// session is entitled. In a pure public build nothing registers, so Autopilot
// is never active and the free hold/self-resume behavior is unchanged.
//
// Split of responsibilities: this repo owns the MECHANISM and the seam; the
// commercial module owns the POLICY (predictive avoidance, when/where to
// switch routes, savings optimization). No paid policy lives here.
package autopilot

import (
	"context"
	"time"
)

// ThrottleEvent describes a throttle the runtime hit on a route.
type ThrottleEvent struct {
	RouteLabel   string
	Wait         time.Duration
	FallbackHint string // a viable fallback route the runtime already resolved, if any
	// ReadyFallbacks is the ranked list of routes the runtime resolved as ready
	// to answer right now (configured BYO providers that differ from the
	// throttled route, then the local floor). A ladder strategy switches to the
	// first suitable one; empty means nothing is ready and the strategy should
	// Hold. The runtime executes the chosen route once and falls back to holding
	// if that attempt fails.
	ReadyFallbacks []string
	Hold           int // 1-based index of this hold within the survival loop
	TaskTitle      string
}

// Action is what the runtime should do about a throttle.
type Action int

const (
	// Hold resumes on the same route when capacity returns (the free default).
	Hold Action = iota
	// Decline parks the work rather than resuming (fail-closed supervision).
	Decline
)

// Decision is a strategy's answer for one throttle event.
type Decision struct {
	Action Action
	// SwitchToRoute, when non-empty and Action is Hold, asks the runtime to
	// resume on this fallback route instead of waiting out the throttled one
	// (throttle-aware switching). The runtime re-attempts on it once; if that
	// attempt fails or no switch executor is wired, it falls back to holding
	// the original route. FallbackHint on the event names a route the runtime
	// already resolved as viable.
	SwitchToRoute string
	// Dodged marks a throttle DODGE — work continued without waiting out the
	// reset — for the savings ledger's dodge counter. Only meaningful for
	// paid switching; the free runtime never sets it.
	Dodged bool
}

// Strategy is implemented by the commercial Autopilot module.
type Strategy interface {
	OnThrottle(ctx context.Context, event ThrottleEvent) (Decision, error)
}

// registered holds the process-wide strategy. Register exactly once during
// program init, before the CLI runs; the runtime only reads it thereafter.
var registered Strategy

// Register installs the paid Autopilot strategy for this process. Passing nil
// clears it (useful in tests).
func Register(s Strategy) { registered = s }

// Registered returns the installed strategy, if any.
func Registered() (Strategy, bool) { return registered, registered != nil }
