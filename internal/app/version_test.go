package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersionPrintsVersionAndChannel(t *testing.T) {
	var out bytes.Buffer
	if code := runVersion(&out); code != 0 {
		t.Fatalf("version exit %d", code)
	}
	got := out.String()
	if !strings.HasPrefix(got, "jini ") || !strings.Contains(got, "(") || !strings.Contains(got, ")") {
		t.Fatalf("expected 'jini <version> (<channel>)', got %q", got)
	}
}

func TestBuildVersionAndChannelOverride(t *testing.T) {
	prevV, prevC := buildVersion, buildChannel
	t.Cleanup(func() { buildVersion, buildChannel = prevV, prevC })
	buildVersion, buildChannel = "9.9.9", "beta"
	if jiniVersion() != "9.9.9" {
		t.Fatalf("jiniVersion() = %q, want 9.9.9", jiniVersion())
	}
	if jiniChannel() != "beta" {
		t.Fatalf("jiniChannel() = %q, want beta", jiniChannel())
	}
}

// A source build (no ldflags) reports the 'source' channel, never empty.
func TestChannelDefaultsToSource(t *testing.T) {
	prevC := buildChannel
	t.Cleanup(func() { buildChannel = prevC })
	buildChannel = ""
	if jiniChannel() != "source" {
		t.Fatalf("empty channel must default to source, got %q", jiniChannel())
	}
}
