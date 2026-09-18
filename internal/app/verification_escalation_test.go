package app

import "testing"

func TestEscalationNoneWhenVerified(t *testing.T) {
	d := decideVerificationEscalation(true, []string{"claude-api", "openai"}, 0, verificationEscalationCap)
	if d.Escalate {
		t.Fatalf("must not escalate when verified")
	}
}

func TestEscalationRoutesUpOnFailure(t *testing.T) {
	ladder := []string{"groq", "claude-api"}
	d := decideVerificationEscalation(false, ladder, 0, verificationEscalationCap)
	if !d.Escalate || d.NextRoute != "groq" {
		t.Fatalf("first escalation must go to the weakest ladder target, got %+v", d)
	}
	d2 := decideVerificationEscalation(false, ladder, 1, verificationEscalationCap)
	if !d2.Escalate || d2.NextRoute != "claude-api" {
		t.Fatalf("second escalation must advance the ladder, got %+v", d2)
	}
}

func TestEscalationStopsAtCap(t *testing.T) {
	ladder := []string{"a", "b", "c", "d"}
	d := decideVerificationEscalation(false, ladder, verificationEscalationCap, verificationEscalationCap)
	if d.Escalate {
		t.Fatalf("must not escalate past the cap, got %+v", d)
	}
	if d.Reason == "" {
		t.Fatalf("a capped decision must explain itself for the receipt")
	}
}

func TestEscalationStopsWhenLadderExhausted(t *testing.T) {
	// Ladder shorter than the cap: no stronger route available.
	d := decideVerificationEscalation(false, []string{"only-one"}, 1, verificationEscalationCap)
	if d.Escalate {
		t.Fatalf("must not escalate when the ladder is exhausted, got %+v", d)
	}
}

func TestEscalationEmptyLadderNeverEscalates(t *testing.T) {
	// No consented stronger route (e.g. free tier, no BYO premium): stay honest,
	// report unverified rather than fabricate success or adopt a paid route.
	d := decideVerificationEscalation(false, nil, 0, verificationEscalationCap)
	if d.Escalate {
		t.Fatalf("empty ladder must never escalate (free/commercial boundary), got %+v", d)
	}
}
