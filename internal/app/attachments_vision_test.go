package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	for _, id := range []string{"anthropic", "azure-openai", "local-slm"} {
		if !routeSupportsVision(providerConfig{ID: id}) {
			t.Fatalf("%s must support vision", id)
		}
	}
	for _, id := range []string{"local-preview", "bedrock"} {
		if routeSupportsVision(providerConfig{ID: id}) {
			t.Fatalf("%s must not be treated as vision-capable yet", id)
		}
	}
}

func TestOpenAIVisionContent(t *testing.T) {
	// No images → single text block (byte-identical to text-only).
	if c := openaiVisionContent("hi", nil); len(c) != 1 || c[0]["type"] != "text" || c[0]["text"] != "hi" {
		t.Fatalf("text-only content wrong: %+v", c)
	}
	dir := t.TempDir()
	png := filepath.Join(dir, "p.png")
	_ = os.WriteFile(png, []byte("\x89PNGdata"), 0o644)
	heic := filepath.Join(dir, "x.heic")
	_ = os.WriteFile(heic, []byte("d"), 0o644)
	big := filepath.Join(dir, "b.png")
	_ = os.WriteFile(big, bytes.Repeat([]byte("A"), maxImageAttachmentB+1), 0o644)

	c := openaiVisionContent("describe", []attachmentRef{
		{Path: png, Kind: "image"},
		{Path: filepath.Join(dir, "gone.png"), Kind: "image"}, // missing
		{Path: heic, Kind: "image"},                           // unsupported
		{Path: big, Kind: "image"},                            // oversize
	})
	if len(c) != 2 || c[1]["type"] != "image_url" {
		t.Fatalf("expected text + one image_url (others skipped): %+v", c)
	}
	iu, _ := c[1]["image_url"].(map[string]any)
	url, _ := iu["url"].(string)
	if !strings.HasPrefix(url, "data:image/png;base64,") {
		t.Fatalf("image_url must be a png data URI, got %q", url)
	}
}

func TestOpenAIChatMessages_TextVsVision(t *testing.T) {
	// No images → user content is a plain string (unchanged payload shape).
	msgs := openaiChatMessages("azure-openai", providerGenerationRequest{}, "sys", "hello")
	if _, ok := msgs[1]["content"].(string); !ok {
		t.Fatalf("no-image user content must be a string, got %T", msgs[1]["content"])
	}
	// Images + vision route → user content is the multimodal array.
	dir := t.TempDir()
	png := filepath.Join(dir, "p.png")
	_ = os.WriteFile(png, []byte("\x89PNG"), 0o644)
	req := providerGenerationRequest{Images: []attachmentRef{{Path: png, Kind: "image"}}}
	msgs = openaiChatMessages("azure-openai", req, "sys", "look")
	if _, ok := msgs[1]["content"].([]map[string]any); !ok {
		t.Fatalf("image request must yield array content, got %T", msgs[1]["content"])
	}
	// Non-vision route ignores images (stays a string).
	msgs = openaiChatMessages("local-preview", req, "sys", "look")
	if _, ok := msgs[1]["content"].(string); !ok {
		t.Fatalf("non-vision route must keep string content, got %T", msgs[1]["content"])
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
