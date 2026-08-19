package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestImageMediaType(t *testing.T) {
	cases := map[string]string{
		"a.png": "image/png", "b.JPG": "image/jpeg", "c.jpeg": "image/jpeg",
		"d.gif": "image/gif", "e.webp": "image/webp", "f.heic": "", "g.txt": "",
	}
	for name, want := range cases {
		if got := imageMediaType(name); got != want {
			t.Fatalf("%s → %q, want %q", name, got, want)
		}
	}
}

func TestRouteSupportsVision(t *testing.T) {
	if !routeSupportsVision(providerConfig{ID: "anthropic"}) {
		t.Fatal("anthropic must support vision")
	}
	for _, id := range []string{"local-preview", "local-slm", "bedrock", "azure-openai"} {
		if routeSupportsVision(providerConfig{ID: id}) {
			t.Fatalf("%s must not be treated as vision-capable yet", id)
		}
	}
}

func TestAnthropicUserContent_TextOnlyWhenNoImages(t *testing.T) {
	content := anthropicUserContent("hello", nil)
	if len(content) != 1 || content[0]["type"] != "text" || content[0]["text"] != "hello" {
		t.Fatalf("no-image content must be a single text block, got %+v", content)
	}
}

func TestAnthropicUserContent_EncodesImages(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "pic.png")
	if err := os.WriteFile(png, []byte("\x89PNG\r\n\x1a\nDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A supported, readable, in-budget image → base64 image block appended.
	content := anthropicUserContent("describe", []attachmentRef{{Path: png, Kind: "image"}})
	if len(content) != 2 || content[1]["type"] != "image" {
		t.Fatalf("expected a text + image block, got %+v", content)
	}
	src, ok := content[1]["source"].(map[string]any)
	if !ok || src["media_type"] != "image/png" || src["type"] != "base64" || src["data"] == "" {
		t.Fatalf("image source block malformed: %+v", content[1])
	}
}

func TestAnthropicUserContent_SkipsUnusable(t *testing.T) {
	dir := t.TempDir()
	missing := attachmentRef{Path: filepath.Join(dir, "gone.png"), Kind: "image"}
	heic := filepath.Join(dir, "x.heic")
	_ = os.WriteFile(heic, []byte("data"), 0o644)
	big := filepath.Join(dir, "big.png")
	_ = os.WriteFile(big, bytes.Repeat([]byte("A"), maxImageAttachmentB+1), 0o644)

	content := anthropicUserContent("go", []attachmentRef{
		missing,
		{Path: heic, Kind: "image"}, // unsupported type
		{Path: big, Kind: "image"},  // oversize
	})
	if len(content) != 1 {
		t.Fatalf("missing/unsupported/oversize images must be skipped, got %+v", content)
	}
}
