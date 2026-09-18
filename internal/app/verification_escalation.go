package app

import "fmt"

// Confidence-based escalation — #1 Phase 2 of specs/pre-viral-readiness.md. When a
// coding change FAILS objective verification (jini verify), route UP the quality
// ladder instead of shipping unverified output. This is the decision core: pure,
// bounded, and boundary-safe. The live loop (Phase 2b) supplies the ladder + attempt
// count and applies the decision; the receipt logs every escalation honestly.
//
// Two invariants keep frugality and the free/commercial boundary intact:
//   - the ladder is built by the route layer to contain ONLY routes stronger than
//     the current one that the user has actually consented to (a paid route never
//     appears unless the commercial tier / a BYO premium key is configured), and
//   - a cap bounds escalations so a single task never walks the whole ladder.
//
// Confidence is EARNED by verification, never asserted by the model — so the only
// trigger here is an objective verification failure.

// verificationEscalationCap bounds escalations per task. Small on purpose: the
// point is one step up on failure, not burning premium capacity.
const verificationEscalationCap = 2

type escalationDecision struct {
	Escalate  bool
	NextRoute string
	Reason    string
}

// decideVerificationEscalation decides whether to route up after a verify verdict.
// ladder is the ordered escalation targets (weakest→strongest), each already known
// to be stronger than the current route and consented to by the user. attemptsUsed
// counts escalations already taken for this task; capLimit bounds them.
func decideVerificationEscalation(verified bool, ladder []string, attemptsUsed, capLimit int) escalationDecision {
	if verified {
		return escalationDecision{Reason: "verified; no escalation needed"}
	}
	if attemptsUsed >= capLimit {
		return escalationDecision{Reason: fmt.Sprintf("verification failed, but the escalation cap (%d) is reached", capLimit)}
	}
	if attemptsUsed >= len(ladder) {
		return escalationDecision{Reason: "verification failed, but no stronger route is available or consented to"}
	}
	next := ladder[attemptsUsed]
	return escalationDecision{
		Escalate:  true,
		NextRoute: next,
		Reason:    fmt.Sprintf("verification failed; escalating to %s (attempt %d of %d)", next, attemptsUsed+1, capLimit),
	}
}
