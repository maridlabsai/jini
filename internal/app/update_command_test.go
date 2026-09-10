package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveUpdateChannel(t *testing.T) {
	prevV, prevC := buildVersion, buildChannel
	t.Cleanup(func() { buildVersion, buildChannel = prevV, prevC })

	// Explicit wins.
	buildChannel = "beta"
	if got := resolveUpdateChannel("nightly"); got != "nightly" {
		t.Fatalf("explicit channel should win, got %q", got)
	}
	// Build channel used when no explicit + not source.
	if got := resolveUpdateChannel(""); got != "beta" {
		t.Fatalf("build channel should be used, got %q", got)
	}
	// Source build with no explicit + no receipt → stable default.
	buildChannel = ""
	if got := resolveUpdateChannel(""); got != "stable" {
		t.Fatalf("should default to stable, got %q", got)
	}
	// Invalid explicit is ignored (falls through), not treated as a channel.
	buildChannel = "beta"
	if got := resolveUpdateChannel("bogus"); got != "beta" {
		t.Fatalf("invalid explicit should fall through to build channel, got %q", got)
	}
}

func TestUpdateInstallCommandUsesChannelAndURL(t *testing.T) {
	t.Setenv("JINI_UPDATE_INSTALL_URL", "file:///tmp/install.sh")
	got := updateInstallCommand("beta")
	for _, want := range []string{"file:///tmp/install.sh", "--channel beta", "--force"} {
		if !strings.Contains(got, want) {
			t.Fatalf("command %q missing %q", got, want)
		}
	}
}

func TestRunUpdateRejectsBadChannel(t *testing.T) {
	var out bytes.Buffer
	if code := runUpdate([]string{"--channel", "bogus"}, &out, &out); code != 1 {
		t.Fatalf("bad channel should exit 1, got %d", code)
	}
	if !strings.Contains(out.String(), "Unknown channel") {
		t.Fatalf("expected unknown-channel message, got %q", out.String())
	}
}

func TestRunUpdateRejectsUnknownArg(t *testing.T) {
	var out bytes.Buffer
	if code := runUpdate([]string{"--wat"}, &out, &out); code != 1 {
		t.Fatalf("unknown arg should exit 1, got %d", code)
	}
}
