package app

import (
	"context"
	"fmt"
	"strings"
)

// Phase 2b of the confidence-routing + verification design
// (specs/pre-viral-readiness.md): after a coding handoff that actually changed
// code and succeeded, run objective verification (jini verify) on the working
// directory and stamp the receipt with an honest verdict — so Jini reports
// "verified ✓ / UNVERIFIED" instead of trusting the model's word.
//
// Opt-in during rollout (JINI_VERIFY_AFTER_TASK) so no one is surprised by build
// latency; once validated on real repos it can default on. The escalation loop
// (Phase 2c) reads receipt.Verification to route up the ladder on a failure.

// cliHandoffVerifyEnabled reports whether post-task verification runs.
func cliHandoffVerifyEnabled() bool {
	return strings.TrimSpace(configValue("JINI_VERIFY_AFTER_TASK")) != ""
}

// cliHandoffVerifyIncludeTests adds the (slower) test checks when the opt-in is
// "tests" or "full"; plain truthy values run buildability only.
func cliHandoffVerifyIncludeTests() bool {
	switch strings.ToLower(strings.TrimSpace(configValue("JINI_VERIFY_AFTER_TASK"))) {
	case "tests", "full", "test":
		return true
	default:
		return false
	}
}

// annotateCLIHandoffVerification runs objective verification on the handoff's
// working directory when it changed code AND succeeded, stamping the receipt.
// It deliberately does nothing on a failed handoff (there is no trustworthy
// result to verify) or when nothing changed (nothing to verify).
func annotateCLIHandoffVerification(ctx context.Context, receipt *cliHandoffReceipt) {
	if receipt == nil || !cliHandoffVerifyEnabled() {
		return
	}
	if receipt.ExitStatus != 0 || receipt.SideEffectCount == 0 {
		return
	}
	dir := strings.TrimSpace(receipt.CWD)
	if dir == "" {
		return
	}
	result := verifyWorkspace(ctx, dir, cliHandoffVerifyIncludeTests())
	receipt.Verification = &result
}

// cliHandoffVerificationLine renders the one-line verdict for the receipt summary,
// or "" when verification was not run.
func cliHandoffVerificationLine(receipt *cliHandoffReceipt) string {
	if receipt == nil || receipt.Verification == nil {
		return ""
	}
	v := receipt.Verification
	switch {
	case v.Ran == 0:
		return "Verification: unverifiable (no recognized build system)"
	case v.Verified:
		return fmt.Sprintf("Verification: verified ✓ (%d objective checks passed)", v.Ran)
	default:
		for _, c := range v.Checks {
			if !c.Skipped && !c.Passed {
				return fmt.Sprintf("Verification: UNVERIFIED — %q failed (%d of %d checks failed)", c.Name, v.Failed, v.Ran)
			}
		}
		return fmt.Sprintf("Verification: UNVERIFIED (%d of %d checks failed)", v.Failed, v.Ran)
	}
}
