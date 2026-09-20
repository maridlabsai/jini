package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func bulkTempFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func stubBulkReadDelegate(t *testing.T, fn func(bundle string) (string, error)) {
	t.Helper()
	old := bulkReadDelegate
	t.Cleanup(func() { bulkReadDelegate = old })
	bulkReadDelegate = func(_ context.Context, bundle string) (string, error) { return fn(bundle) }
}

func TestBulkReadBundleWrapsFilesAndQuestion(t *testing.T) {
	f := bulkTempFile(t, "a.go", "package a\n// hello")
	bundle, err := bulkReadBundle([]string{f}, "what package is this?")
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	for _, want := range []string{"<file path=", "package a", "Question: what package is this?"} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("bundle missing %q:\n%s", want, bundle)
		}
	}
}

func TestRunBulkReadReturnsCheapSummary(t *testing.T) {
	f := bulkTempFile(t, "a.go", "package a")
	stubBulkReadDelegate(t, func(bundle string) (string, error) {
		if !strings.Contains(bundle, "package a") {
			t.Fatalf("the delegate must receive the wrapped file content")
		}
		return "- a.go: declares package a", nil
	})
	var out bytes.Buffer
	if code := runBulkRead([]string{f, "--question", "pkg?"}, &out, &out); code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out.String())
	}
	if !strings.Contains(out.String(), "declares package a") {
		t.Fatalf("expected the cheap summary, got:\n%s", out.String())
	}
}

func TestRunBulkReadNoFilesIsUsageError(t *testing.T) {
	var out bytes.Buffer
	if code := runBulkRead(nil, &out, &out); code != 1 {
		t.Fatalf("no files must exit 1, got %d", code)
	}
}

func TestRunBulkReadEmitsBundleWhenNoCheapRoute(t *testing.T) {
	f := bulkTempFile(t, "a.go", "package a")
	stubBulkReadDelegate(t, func(string) (string, error) { return "", errors.New("no ready cheap route") })
	var out bytes.Buffer
	if code := runBulkRead([]string{f}, &out, &out); code != 0 {
		t.Fatalf("no route must still exit 0 and emit the bundle, got %d", code)
	}
	if !strings.Contains(out.String(), "<file path=") {
		t.Fatalf("must emit the bundle when no cheap route:\n%s", out.String())
	}
}
