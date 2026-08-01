package app

// Bridge from the public autopilot extension contract to the internal throttle
// approver seam. When the commercial module has registered a Strategy AND the
// session is entitled, a throttle consults that strategy; otherwise the free
// hold/self-resume behavior is used unchanged.

import (
	"context"

	"github.com/maridlabsai/jini/autopilot"
)

type autopilotApprover struct{ strategy autopilot.Strategy }

func (a autopilotApprover) Approve(ctx context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error) {
	decision, err := a.strategy.OnThrottle(ctx, autopilot.ThrottleEvent{
		RouteLabel:   req.Label,
		Wait:         req.Wait,
		FallbackHint: req.FallbackHint,
		Hold:         req.Hold,
		TaskTitle:    req.TaskTitle,
	})
	if err != nil {
		return approvalDeclined, err
	}
	if decision.Action == autopilot.Decline {
		return approvalDeclined, nil
	}
	return approvalGranted, nil
}

// ResolveThrottle is the richer seam: it carries the strategy's fallback-route
// choice and dodge flag through to the survival loop, which the plain Approve
// path cannot express.
func (a autopilotApprover) ResolveThrottle(ctx context.Context, req throttleApprovalRequest) (throttleResolution, error) {
	decision, err := a.strategy.OnThrottle(ctx, autopilot.ThrottleEvent{
		RouteLabel:   req.Label,
		Wait:         req.Wait,
		FallbackHint: req.FallbackHint,
		Hold:         req.Hold,
		TaskTitle:    req.TaskTitle,
	})
	if err != nil {
		return throttleResolution{decision: approvalDeclined}, err
	}
	if decision.Action == autopilot.Decline {
		return throttleResolution{decision: approvalDeclined}, nil
	}
	return throttleResolution{
		decision:      approvalGranted,
		switchToRoute: decision.SwitchToRoute,
		dodged:        decision.Dodged,
	}, nil
}

// entitledAutopilotApprover returns the paid approver when a strategy is
// registered and the session is entitled; ok=false leaves the free path.
func entitledAutopilotApprover() (throttleApprover, bool) {
	strategy, registered := autopilot.Registered()
	if !registered || currentSubscriptionTier() != "commercial" {
		return nil, false
	}
	return autopilotApprover{strategy: strategy}, true
}
