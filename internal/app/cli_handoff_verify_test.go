package app

import (
	"context"
	"strings"
	"testing"
)

func TestHandoffVerificationSkipsWhenOptInUnset(t *testing.T) {
	r := &cliHandoffReceipt{CWD: writeGoModule(t), ExitStatus: 0, SideEffectCount: 1}
	annotateCLIHandoffVerification(context.Background(), r)
	if r.Verification != nil {
		t.Fatalf("must not verify when JINI_VERIFY_AFTER_TASK is unset")
	}
}

func TestHandoffVerificationRunsOnSuccessfulCodeChange(t *testing.T) {
	t.Setenv("JINI_VERIFY_AFTER_TASK", "1")
	stubVerifyRunner(t, func(_ []string) (bool, string) { return true, "" })
	r := &cliHandoffReceipt{CWD: writeGoModule(t), ExitStatus: 0, SideEffectCount: 2}
	annotateCLIHandoffVerification(context.Background(), r)
	if r.Verification == nil || !r.Verification.Verified {
		t.Fatalf("must verify (and pass) after a successful code-changing handoff, got %+v", r.Verification)
	}
}

func TestHandoffVerificationSkipsFailedHandoff(t *testing.T) {
	t.Setenv("JINI_VERIFY_AFTER_TASK", "1")
	r := &cliHandoffReceipt{CWD: writeGoModule(t), ExitStatus: 1, SideEffectCount: 2}
	annotateCLIHandoffVerification(context.Background(), r)
	if r.Verification != nil {
		t.Fatalf("must not verify a failed handoff — there is no trustworthy result to check")
	}
}

func TestHandoffVerificationSkipsWhenNothingChanged(t *testing.T) {
	t.Setenv("JINI_VERIFY_AFTER_TASK", "1")
	r := &cliHandoffReceipt{CWD: writeGoModule(t), ExitStatus: 0, SideEffectCount: 0}
	annotateCLIHandoffVerification(context.Background(), r)
	if r.Verification != nil {
		t.Fatalf("must not verify when the handoff changed nothing")
	}
}

func TestHandoffVerificationLineRendersEachVerdict(t *testing.T) {
	verified := &cliHandoffReceipt{Verification: &verifyResult{Ran: 2, Verified: true}}
	if got := cliHandoffVerificationLine(verified); !strings.Contains(got, "verified ✓") {
		t.Fatalf("verified line wrong: %q", got)
	}
	failed := &cliHandoffReceipt{Verification: &verifyResult{Ran: 2, Failed: 1, Checks: []verifyCheck{{Name: "go build", Passed: false}}}}
	if got := cliHandoffVerificationLine(failed); !strings.Contains(got, "UNVERIFIED") {
		t.Fatalf("failed line wrong: %q", got)
	}
	unverifiable := &cliHandoffReceipt{Verification: &verifyResult{Ran: 0}}
	if got := cliHandoffVerificationLine(unverifiable); !strings.Contains(got, "unverifiable") {
		t.Fatalf("unverifiable line wrong: %q", got)
	}
	if got := cliHandoffVerificationLine(&cliHandoffReceipt{}); got != "" {
		t.Fatalf("no verification → empty line, got %q", got)
	}
}
