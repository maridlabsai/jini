package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAttachmentRefs(t *testing.T) {
	got := parseAttachmentRefs("summarize @report.md and @src/main.go, then ping @alice.")
	want := []string{"report.md", "src/main.go", "alice"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", got, want)
	}
	if refs := parseAttachmentRefs("just @ and @@handle"); len(refs) != 1 || refs[0] != "handle" {
		// "@" alone skipped; "@@handle" → TrimPrefix leaves "@handle" → skipped by the
		// leading-@ guard; but "@@handle" first strips one @ to "@handle" then skipped.
		if len(refs) != 0 {
			t.Fatalf("bare @ and @@ must be skipped, got %v", refs)
		}
	}
}

func TestPathLike(t *testing.T) {
	for _, p := range []string{"report.md", "src/main.go", "/abs/x", "a.b"} {
		if !pathLike(p) {
			t.Fatalf("%q should be path-like", p)
		}
	}
	for _, p := range []string{"alice", "team", "handle"} {
		if pathLike(p) {
			t.Fatalf("%q should NOT be path-like", p)
		}
	}
}

func TestResolveAttachments_FoundMissingAndHandles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	found, missing := resolveAttachments("read @report.md, @gone.txt, and ask @alice", dir)
	if len(found) != 1 || found[0].Kind != "file" || filepath.Base(found[0].Path) != "report.md" {
		t.Fatalf("expected report.md found as file, got %+v", found)
	}
	if len(missing) != 1 || missing[0] != "gone.txt" {
		t.Fatalf("expected only path-like gone.txt missing (handle @alice ignored), got %v", missing)
	}
}

func TestResolveAttachments_CountCap(t *testing.T) {
	dir := t.TempDir()
	var b strings.Builder
	for i := 0; i < maxAttachments+5; i++ {
		name := filepath.Join(dir, "f"+string(rune('a'+i))+".txt")
		_ = os.WriteFile(name, []byte("x"), 0o644)
		b.WriteString(" @f" + string(rune('a'+i)) + ".txt")
	}
	found, _ := resolveAttachments(b.String(), dir)
	if len(found) != maxAttachments {
		t.Fatalf("expected count capped at %d, got %d", maxAttachments, len(found))
	}
}

func TestMissingAttachmentError(t *testing.T) {
	if s := missingAttachmentError([]string{"a.txt"}); !strings.Contains(s, `"a.txt"`) || !strings.Contains(s, "couldn't find") {
		t.Fatalf("single message wrong: %q", s)
	}
	if s := missingAttachmentError([]string{"a.txt", "b.png"}); !strings.Contains(s, "these attachments") {
		t.Fatalf("multi message wrong: %q", s)
	}
}

func TestInlineTextAttachments_CapsAndSkipsImages(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.txt")
	_ = os.WriteFile(big, bytes.Repeat([]byte("A"), maxInlineAttachmentB+500), 0o644)
	img := filepath.Join(dir, "pic.png")
	_ = os.WriteFile(img, []byte("PNGDATA"), 0o644)

	found := []attachmentRef{
		{Raw: "@big.txt", Path: big, Kind: "file"},
		{Raw: "@pic.png", Path: img, Kind: "image"},
	}
	out := inlineTextAttachments("do it", found)
	if !strings.Contains(out, "--- big.txt ---") {
		t.Fatal("text attachment should be inlined")
	}
	if !strings.Contains(out, inlineTruncatedNotice) {
		t.Fatal("oversize text file should be truncated with a notice")
	}
	if strings.Contains(out, "PNGDATA") || strings.Contains(out, "--- pic.png ---") {
		t.Fatal("image bytes must not be inlined")
	}
}

func TestIntake_AttachmentAckAndFailClosed(t *testing.T) {
	t.Setenv("JINI_PROVIDER", "local-preview")
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte("Q3 revenue up"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Present attachment → acknowledged, no error.
	var out, errBuf bytes.Buffer
	code := RunInteractive([]string{"summarize", "@report.md"}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("present attachment should not fail: code=%d err=%q", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Attached: report.md (file)") {
		t.Fatalf("expected acknowledgement, got %q", out.String())
	}

	// Missing attachment → fail closed with exact message.
	var out2, err2 bytes.Buffer
	code = RunInteractive([]string{"summarize", "@nope.md"}, strings.NewReader(""), &out2, &err2)
	if code != 1 {
		t.Fatalf("missing attachment must fail closed, got code=%d", code)
	}
	if !strings.Contains(err2.String(), `couldn't find the attachment "nope.md"`) {
		t.Fatalf("expected exact missing message, got %q", err2.String())
	}
}
