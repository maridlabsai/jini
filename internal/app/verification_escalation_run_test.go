package app

import (
	"context"
	"strings"
	"testing"
)

func verifiedReceiptFixture() *cliHandoffReceipt {
	return &cliHandoffReceipt{Verification: &verifyResult{Ran: 2, Verified: true}}
}
func unverifiedReceiptFixture() *cliHandoffReceipt {
	return &cliHandoffReceipt{Verification: &verifyResult{Ran: 2, Failed: 1}}
}

func TestEscalationLadderParsesConsentedRoutes(t *testing.T) {
	t.Setenv("JINI_VERIFY_ESCALATE_TO", " claude-code , openai ")
	l := verificationEscalationLadder()
	if len(l) != 2 {
		t.Fatalf("expected 2 consented routes, got %v", l)
	}
}

func TestEscalationLadderEmptyWhenUnset(t *testing.T) {
	t.Setenv("JINI_VERIFY_ESCALATE_TO", "")
	if verificationEscalationLadder() != nil {
		t.Fatalf("no config → nil ladder (never escalate silently)")
	}
}

func TestRunEscalationSkipsVerifiedStart(t *testing.T) {
	start := verifiedReceiptFixture()
	got, trail := runVerificationEscalation(context.Background(), start, []string{"strong"}, 2,
		func(context.Context, string) *cliHandoffReceipt { t.Fatal("must not re-run a verified result"); return nil })
	if got != start || len(trail) != 0 {
		t.Fatalf("verified start must not escalate")
	}
}

func TestRunEscalationSkipsWhenVerificationNotRun(t *testing.T) {
	start := &cliHandoffReceipt{} // Verification nil → can't judge → don't escalate
	got, trail := runVerificationEscalation(context.Background(), start, []string{"strong"}, 2,
		func(context.Context, string) *cliHandoffReceipt { t.Fatal("must not re-run when verification did not run"); return nil })
	if got != start || len(trail) != 0 {
		t.Fatalf("unrun verification must not escalate")
	}
}

func TestRunEscalationAdoptsVerifiedStrongerResult(t *testing.T) {
	start := unverifiedReceiptFixture()
	strong := verifiedReceiptFixture()
	got, trail := runVerificationEscalation(context.Background(), start, []string{"strong"}, 2,
		func(_ context.Context, route string) *cliHandoffReceipt {
			if route == "strong" {
				return strong
			}
			return unverifiedReceiptFixture()
		})
	if got != strong {
		t.Fatalf("must adopt the verified stronger result")
	}
	if len(trail) != 1 || !trail[0].Verified {
		t.Fatalf("trail must record the verified escalation, got %+v", trail)
	}
}

func TestRunEscalationKeepsOriginalAndStopsAtCap(t *testing.T) {
	start := unverifiedReceiptFixture()
	calls := 0
	got, trail := runVerificationEscalation(context.Background(), start, []string{"a", "b", "c"}, 2,
		func(context.Context, string) *cliHandoffReceipt {
			calls++
			return unverifiedReceiptFixture() // never verifies
		})
	if got != start {
		t.Fatalf("must keep the original when no stronger route verifies")
	}
	if calls != 2 {
		t.Fatalf("must stop at the cap (2 re-runs), got %d", calls)
	}
	if len(trail) != 2 {
		t.Fatalf("trail must record the 2 capped attempts, got %d", len(trail))
	}
}

func TestEscalationReasonIsHonest(t *testing.T) {
	won := appendVerificationEscalationReason("base.", []verificationEscalationAttempt{{Route: "openai", Verified: true}})
	if !strings.Contains(won, "verified on openai") {
		t.Fatalf("a won escalation must name the verifying route: %q", won)
	}
	lost := appendVerificationEscalationReason("", []verificationEscalationAttempt{{Route: "a"}, {Route: "b"}})
	if !strings.Contains(lost, "still unverified") {
		t.Fatalf("a failed escalation must say so honestly: %q", lost)
	}
}
