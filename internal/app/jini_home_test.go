package app

import (
	"os"
	"testing"
)

// JINI_HOME redirects global per-user state (mode/trust/savings) to a sandbox so
// a dogfood or test run never pollutes the real ~/.jini ledger. Without it, the
// OS home is used.
func TestJiniHomeDirHonorsOverride(t *testing.T) {
	t.Setenv("JINI_HOME", "/tmp/jini-home-sandbox")
	got, err := jiniHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/jini-home-sandbox" {
		t.Fatalf("JINI_HOME override = %q, want /tmp/jini-home-sandbox", got)
	}
}

func TestJiniHomeDirFallsBackToOSHome(t *testing.T) {
	t.Setenv("JINI_HOME", "")
	got, err := jiniHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	osHome, _ := os.UserHomeDir()
	if got != osHome {
		t.Fatalf("no override should use OS home %q, got %q", osHome, got)
	}
}

// Whitespace-only override is ignored (treated as unset).
func TestJiniHomeDirIgnoresBlankOverride(t *testing.T) {
	t.Setenv("JINI_HOME", "   ")
	got, err := jiniHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	osHome, _ := os.UserHomeDir()
	if got != osHome {
		t.Fatalf("blank override should fall back to OS home %q, got %q", osHome, got)
	}
}
