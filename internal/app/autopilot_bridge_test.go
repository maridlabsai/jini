package app

import (
	"context"
	"testing"

	"github.com/maridlabsai/jini/autopilot"
)

type stubStrategy struct{ action autopilot.Action }

func (s stubStrategy) OnThrottle(context.Context, autopilot.ThrottleEvent) (autopilot.Decision, error) {
	return autopilot.Decision{Action: s.action}, nil
}

func TestEntitledAutopilotApproverRequiresRegistrationAndEntitlement(t *testing.T) {
	autopilot.Register(nil)
	t.Cleanup(func() { autopilot.Register(nil) })

	// Nothing registered → free path.
	t.Setenv("JINI_TIER", "commercial")
	if _, ok := entitledAutopilotApprover(); ok {
		t.Fatal("no strategy registered → must not engage autopilot")
	}

	autopilot.Register(stubStrategy{action: autopilot.Hold})

	// Registered but free tier → free path (fails closed on entitlement).
	t.Setenv("JINI_TIER", "free")
	if _, ok := entitledAutopilotApprover(); ok {
		t.Fatal("free tier must not engage paid autopilot even if a strategy is present")
	}

	// Registered + entitled → engaged.
	t.Setenv("JINI_TIER", "commercial")
	approver, ok := entitledAutopilotApprover()
	if !ok {
		t.Fatal("registered + entitled must engage autopilot")
	}
	if _, isAP := approver.(autopilotApprover); !isAP {
		t.Fatalf("expected autopilotApprover, got %T", approver)
	}
}

func TestConfigureThrottleApproverUsesAutopilotWhenEntitled(t *testing.T) {
	withExecutionModeHome(t)
	autopilot.Register(stubStrategy{action: autopilot.Hold})
	t.Cleanup(func() { autopilot.Register(nil) })
	old := throttleApproverForProcess
	t.Cleanup(func() { throttleApproverForProcess = old })

	t.Setenv("JINI_TIER", "commercial")
	configureThrottleApproverForEntry(true)
	if _, ok := throttleApproverForProcess.(autopilotApprover); !ok {
		t.Fatalf("entitled + registered must install autopilotApprover, got %T", throttleApproverForProcess)
	}
}

func TestAutopilotApproverMapsDeclineToDeclined(t *testing.T) {
	a := autopilotApprover{strategy: stubStrategy{action: autopilot.Decline}}
	decision, err := a.Approve(context.Background(), throttleApprovalRequest{Label: "route"})
	if err != nil || decision != approvalDeclined {
		t.Fatalf("Decline must map to approvalDeclined, got %v %v", decision, err)
	}
	grant := autopilotApprover{strategy: stubStrategy{action: autopilot.Hold}}
	decision, err = grant.Approve(context.Background(), throttleApprovalRequest{Label: "route"})
	if err != nil || decision != approvalGranted {
		t.Fatalf("Hold must map to approvalGranted, got %v %v", decision, err)
	}
}
