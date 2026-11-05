package app

import (
	"strings"
	"testing"
)

func TestR1APolicyFailsClosed(t *testing.T) {
	base := r1aOperationRequest{
		SchemaVersion: r1aSchemaVersion,
		OperationID:   "op_00000000000000000000000000000000",
		RunID:         "run_00000000000000000000000000000000",
		Mode:          r1aModeSupervised,
		OpClass:       r1aOpClassSingleFileWrite,
		Capability:    r1aCapFSWriteSingle,
		RelTarget:     "a/b.txt",
		WriteScope:    []string{"a/b.txt"},
		BeforeHash:    "sha256:x",
		AfterHash:     "sha256:y",
	}
	if dec := r1aEvaluatePolicy(base); !dec.Allow {
		t.Fatalf("baseline request should be allowed: %s", dec.Reason)
	}

	cases := []struct {
		name string
		mut  func(r *r1aOperationRequest)
	}{
		{"bad schema", func(r *r1aOperationRequest) { r.SchemaVersion = "r1a.v999" }},
		{"auto mode", func(r *r1aOperationRequest) { r.Mode = r1aModeAuto }},
		{"plan mode", func(r *r1aOperationRequest) { r.Mode = r1aModePlan }},
		{"unknown class", func(r *r1aOperationRequest) { r.OpClass = "file.write.multi" }},
		{"unknown capability", func(r *r1aOperationRequest) { r.Capability = "fs.write.all" }},
		{"malformed op id", func(r *r1aOperationRequest) { r.OperationID = "nope" }},
		{"absent target", func(r *r1aOperationRequest) { r.RelTarget = "  "; r.WriteScope = []string{"  "} }},
		{"absolute target", func(r *r1aOperationRequest) { r.RelTarget = "/etc/passwd"; r.WriteScope = []string{"/etc/passwd"} }},
		{"traversal target", func(r *r1aOperationRequest) { r.RelTarget = "../x"; r.WriteScope = []string{"../x"} }},
		{"root-equivalent", func(r *r1aOperationRequest) { r.RelTarget = "."; r.WriteScope = []string{"."} }},
		{"scope mismatch", func(r *r1aOperationRequest) { r.WriteScope = []string{"other.txt"} }},
		{"scope too wide", func(r *r1aOperationRequest) { r.WriteScope = []string{"a/b.txt", "c.txt"} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := base
			r.WriteScope = append([]string(nil), base.WriteScope...)
			c.mut(&r)
			if dec := r1aEvaluatePolicy(r); dec.Allow {
				t.Fatalf("%s should fail closed", c.name)
			}
		})
	}
}

func TestR1AIDsAreStableAndValidated(t *testing.T) {
	id, err := r1aNewID("op")
	if err != nil {
		t.Fatal(err)
	}
	if !r1aValidID("op", id) {
		t.Fatalf("generated id %q should validate", id)
	}
	if r1aValidID("run", id) {
		t.Fatal("op id must not validate as run id")
	}
	for _, bad := range []string{"", "op_", "op_zz", "op_" + strings.Repeat("g", 32), "run_00000000000000000000000000000000"} {
		if r1aValidID("op", bad) {
			t.Fatalf("%q should be rejected as an op id", bad)
		}
	}
}

func TestR1AHashIsStable(t *testing.T) {
	a := r1aHashBytes([]byte("hello"))
	b := r1aHashBytes([]byte("hello"))
	c := r1aHashBytes([]byte("world"))
	if a != b {
		t.Fatal("hash must be deterministic")
	}
	if a == c {
		t.Fatal("different content must hash differently")
	}
	if !strings.HasPrefix(a, "sha256:") {
		t.Fatalf("hash must be prefixed: %q", a)
	}
}

func TestR1ABindingMatchAndWellFormed(t *testing.T) {
	req := r1aOperationRequest{
		OperationID: "op_00000000000000000000000000000000",
		RelTarget:   "a.txt",
		BeforeHash:  "sha256:a",
		AfterHash:   "sha256:b",
		WriteScope:  []string{"a.txt"},
	}
	good := r1aApprovalBinding{OperationID: req.OperationID, RelTarget: "a.txt", BeforeHash: "sha256:a", AfterHash: "sha256:b", WriteScope: []string{"a.txt"}, Approved: true}
	if !r1aBindingWellFormed(good) || !r1aBindingMatches(req, good) {
		t.Fatal("matching binding should be accepted")
	}
	bad := good
	bad.AfterHash = "sha256:c"
	if r1aBindingMatches(req, bad) {
		t.Fatal("mismatched after-hash must not match")
	}
	malformed := good
	malformed.OperationID = "nope"
	if r1aBindingWellFormed(malformed) {
		t.Fatal("malformed op id should not be well-formed")
	}
}

func TestR1ARedactionMasksSecretsNotProse(t *testing.T) {
	// A deliberately FAKE token (never the real compromised key).
	in := "route ok; token=abcd1234efgh and key sk-FAKEabcdefghij1234567890 plus normal task text"
	out := r1aRedact(in)
	if strings.Contains(out, "sk-FAKEabcdefghij1234567890") || strings.Contains(out, "token=abcd1234efgh") {
		t.Fatalf("secrets not redacted: %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("expected redaction marker: %q", out)
	}
	if !strings.Contains(out, "normal task text") {
		t.Fatalf("ordinary prose must survive: %q", out)
	}
}

func TestR1AOutcomesValid(t *testing.T) {
	for _, o := range []r1aOutcome{
		r1aCompletedVerified, r1aCompletedUnverified, r1aDenied, r1aRejected,
		r1aFailedNoMutation, r1aFailedRolledBack, r1aRecoveryRequired,
		r1aConflict, r1aCancelled, r1aInterrupted,
	} {
		if !o.valid() {
			t.Fatalf("%q should be a valid outcome", o)
		}
	}
	if r1aOutcome("failed_partial").valid() {
		t.Fatal("failed_partial is not an Increment-1 outcome")
	}
}

func TestR1AEventLogOrdered(t *testing.T) {
	var l r1aEventLog
	for i := 0; i < 5; i++ {
		l.append(r1aEvent{SchemaVersion: r1aSchemaVersion, EventType: "e"})
	}
	events := l.snapshotEvents()
	if len(events) != 5 {
		t.Fatalf("want 5 events, got %d", len(events))
	}
	for i, e := range events {
		if e.Sequence != int64(i+1) {
			t.Fatalf("event %d has sequence %d", i, e.Sequence)
		}
		if e.SchemaVersion != r1aSchemaVersion {
			t.Fatalf("event %d missing schema version", i)
		}
	}
}
