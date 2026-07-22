package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestThrottleParkRoundTrip(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := writeThrottlePark("refactor the parser", "Claude API route"); err != nil {
		t.Fatal(err)
	}
	updateThrottleParkError("throttled: status 429", "local-fast")
	park := loadThrottlePark()
	if park == nil {
		t.Fatal("park not loaded")
	}
	if park.Prompt != "refactor the parser" || park.FallbackHint != "local-fast" {
		t.Fatalf("bad park: %+v", park)
	}
	if _, err := time.Parse(time.RFC3339, park.ParkedAt); err != nil {
		t.Fatalf("bad timestamp: %v", err)
	}
	info, err := os.Stat(filepath.Join(sessionStateRoot(), "throttle-park.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %o", perm)
	}
	clearThrottlePark()
	if loadThrottlePark() != nil {
		t.Fatal("park not cleared")
	}
}

func TestThrottleParkCorruptFileLoadsNil(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionStateRoot(), "throttle-park.json"), []byte("{corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if loadThrottlePark() != nil {
		t.Fatal("corrupt park must load as nil")
	}
}

func TestThrottleParkAgeDisclosure(t *testing.T) {
	old := throttlePark{SchemaVersion: "0.1.0", ContextType: "JiniThrottlePark", Prompt: "old task", ParkedAt: time.Now().Add(-72 * time.Hour).Format(time.RFC3339)}
	line := throttleParkResumeLine(&old)
	if !strings.Contains(line, "parked 3 days ago") {
		t.Fatalf("age not disclosed: %q", line)
	}
	fresh := throttlePark{ParkedAt: time.Now().Format(time.RFC3339), Prompt: "new task"}
	if strings.Contains(throttleParkResumeLine(&fresh), "parked") {
		t.Fatal("fresh park must not disclose age")
	}
}
