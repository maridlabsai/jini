package app

import (
	"context"
	"errors"
	"testing"
	"time"

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

// switchingStrategy asks the loop to dodge to a named route.
type switchingStrategy struct{ target string }

func (s switchingStrategy) OnThrottle(context.Context, autopilot.ThrottleEvent) (autopilot.Decision, error) {
	return autopilot.Decision{Action: autopilot.Hold, SwitchToRoute: s.target, Dodged: true}, nil
}

func TestSurvivalLoopSwitchesAndRecordsDodge(t *testing.T) {
	oldSleep := throttleSleep
	throttleSleep = func(context.Context, time.Duration) error { return nil }
	defer func() { throttleSleep = oldSleep }()

	switched := false
	opts := throttleSurvivalOptions{
		approver: autopilotApprover{strategy: switchingStrategy{target: "local-fast"}},
		switchAttempt: func(_ context.Context, mode string) (string, error) {
			switched = true
			if mode != "local-fast" {
				t.Fatalf("unexpected switch target %q", mode)
			}
			return "answered on fallback", nil
		},
	}
	text, report, err := runWithThrottleSurvival(context.Background(), "Claude API route", nil, opts, func() (string, error) {
		return "", &throttledRouteError{label: "Claude API route", underlying: errors.New("rate limit")}
	})
	if err != nil {
		t.Fatalf("switch should succeed, got %v", err)
	}
	if !switched || text != "answered on fallback" {
		t.Fatalf("switch executor not used: switched=%v text=%q", switched, text)
	}
	if !report.Dodged || report.SwitchedTo != "local-fast" {
		t.Fatalf("dodge not recorded in report: %+v", report)
	}
	if report.Holds != 0 {
		t.Fatalf("a dodge must not hold, got %d holds", report.Holds)
	}
}

func TestSurvivalLoopFailedSwitchFallsBackToHold(t *testing.T) {
	oldSleep := throttleSleep
	throttleSleep = func(context.Context, time.Duration) error { return nil }
	defer func() { throttleSleep = oldSleep }()

	attempts := 0
	opts := throttleSurvivalOptions{
		approver: autopilotApprover{strategy: switchingStrategy{target: "local-fast"}},
		switchAttempt: func(context.Context, string) (string, error) {
			return "", errors.New("fallback not ready")
		},
	}
	text, report, err := runWithThrottleSurvival(context.Background(), "route", nil, opts, func() (string, error) {
		attempts++
		if attempts >= 2 {
			return "recovered on original", nil
		}
		return "", &throttledRouteError{label: "route", underlying: errors.New("rate limit")}
	})
	if err != nil || text != "recovered on original" {
		t.Fatalf("failed switch must fall back to holding original route: text=%q err=%v", text, err)
	}
	if report.Dodged {
		t.Fatal("a failed switch is not a dodge")
	}
	if report.Holds != 1 {
		t.Fatalf("expected one hold after failed switch, got %d", report.Holds)
	}
}

func TestRecordSavingsMarksDodgeInLedger(t *testing.T) {
	withSavingsHome(t)
	decision := recordSavingsOnDecision(subscriptionDecision(), providerConfig{}, 4000, 4000,
		providerGenerationRequest{Title: "work task"},
		throttleSurvivalReport{Dodged: true, SwitchedTo: "Codex CLI handoff"})
	if decision.SavingsEntry == nil || !decision.SavingsEntry.ThrottleDodged {
		t.Fatalf("dodge not flagged on entry: %+v", decision.SavingsEntry)
	}
	ledger := loadSavingsLedger()
	if ledger == nil || ledger.Totals.Dodges != 1 {
		t.Fatalf("dodge not counted in ledger totals: %+v", ledger)
	}
}
