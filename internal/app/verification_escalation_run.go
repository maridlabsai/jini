package app

import (
	"context"
	"fmt"
	"strings"
)

// Phase 2c of the verification design (specs/pre-viral-readiness.md): when a
// coding handoff fails objective verification, re-run the task on a stronger route
// the user has consented to, until it verifies or the ladder/cap is exhausted.
//
// Boundary + safety invariants (this is the riskiest mechanism, so they matter):
//   - The ladder comes ONLY from explicit user config (JINI_VERIFY_ESCALATE_TO) —
//     a stronger route is never adopted silently, preserving the free/commercial
//     line and the no-silent-paid-cost rule. Empty ladder → no escalation.
//   - A cap bounds re-runs so a single task never walks the whole ladder.
//   - Only a *verified* escalation is adopted; a still-unverified stronger attempt
//     leaves the original result standing and is reported honestly. Re-runs are
//     non-destructive here (they do not auto-revert the prior attempt); the receipt
//     names both attempts so the user can review via the existing rollback hint.

// verificationEscalationLadder returns the ordered stronger routes the user has
// consented to escalate to (weakest→strongest), or nil when none are configured.
func verificationEscalationLadder() []string {
	raw := strings.TrimSpace(configValue("JINI_VERIFY_ESCALATE_TO"))
	if raw == "" {
		return nil
	}
	var routes []string
	for _, r := range strings.Split(raw, ",") {
		if r = normalizeToolMode(strings.TrimSpace(r)); r != "" {
			routes = append(routes, r)
		}
	}
	return routes
}

type verificationEscalationAttempt struct {
	Route    string
	Verified bool
}

// shouldEscalateReceipt is true only when verification actually RAN and FAILED —
// not when it was skipped (opt-in off / unverifiable), which we can't judge.
func shouldEscalateReceipt(r *cliHandoffReceipt) bool {
	return r != nil && r.Verification != nil && r.Verification.Ran > 0 && !r.Verification.Verified
}

func receiptVerified(r *cliHandoffReceipt) bool {
	return r != nil && r.Verification != nil && r.Verification.Verified
}

// runVerificationEscalation re-runs the task up the ladder while the current
// result stays unverified, bounded by capLimit. execute runs the task on a route
// and returns the resulting receipt (with its own verification). Only a verified
// escalation is adopted; otherwise the original stands. Returns the final receipt
// and the attempt trail (for the honest receipt reason).
func runVerificationEscalation(
	ctx context.Context,
	start *cliHandoffReceipt,
	ladder []string,
	capLimit int,
	execute func(context.Context, string) *cliHandoffReceipt,
) (*cliHandoffReceipt, []verificationEscalationAttempt) {
	current := start
	var trail []verificationEscalationAttempt
	for attemptsUsed := 0; shouldEscalateReceipt(current); attemptsUsed++ {
		d := decideVerificationEscalation(false, ladder, attemptsUsed, capLimit)
		if !d.Escalate {
			break
		}
		next := execute(ctx, d.NextRoute)
		verified := receiptVerified(next)
		trail = append(trail, verificationEscalationAttempt{Route: d.NextRoute, Verified: verified})
		if verified {
			return next, trail // adopt the verified stronger result
		}
	}
	return current, trail
}

// appendVerificationEscalationReason records the escalation outcome on the receipt
// reason so it is never silent.
func appendVerificationEscalationReason(reason string, trail []verificationEscalationAttempt) string {
	if len(trail) == 0 {
		return reason
	}
	var b strings.Builder
	if strings.TrimSpace(reason) != "" {
		b.WriteString(reason)
		b.WriteString(" ")
	}
	last := trail[len(trail)-1]
	if last.Verified {
		b.WriteString(fmt.Sprintf("Verification failed on the initial route; escalated and verified on %s.", last.Route))
	} else {
		routes := make([]string, 0, len(trail))
		for _, a := range trail {
			routes = append(routes, a.Route)
		}
		b.WriteString(fmt.Sprintf("Verification failed; escalated to %s but still unverified — review the change.", strings.Join(routes, " → ")))
	}
	return strings.TrimSpace(b.String())
}
