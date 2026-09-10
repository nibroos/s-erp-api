package ai

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// pngOfSize builds a real encoded PNG so Bounded exercises actual decoding.
func pngOfSize(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x += 2 {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func decodedSize(t *testing.T, data []byte) (int, int) {
	t.Helper()
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width, cfg.Height
}

// A 4K screenshot is ~13,000 visual tokens and overflows an 8,192 context, so
// oversized images must come back within the cap.
func TestBoundedScalesDownAndKeepsAspectRatio(t *testing.T) {
	got := Image{Data: pngOfSize(t, 3840, 2160), MimeType: "image/png"}.Bounded()
	w, h := decodedSize(t, got.Data)
	if w != maxImageDimension {
		t.Errorf("longest side is %d, want %d", w, maxImageDimension)
	}
	if want := 2160 * maxImageDimension / 3840; h != want {
		t.Errorf("height is %d, want %d — aspect ratio was not preserved", h, want)
	}
	if got.MimeType != "image/png" {
		t.Errorf("png was re-encoded as %q; lossless matters for text in screenshots", got.MimeType)
	}
}

// A tall image must be bounded on its height, not its width.
func TestBoundedUsesTheLongestSide(t *testing.T) {
	got := Image{Data: pngOfSize(t, 500, 2000), MimeType: "image/png"}.Bounded()
	w, h := decodedSize(t, got.Data)
	if h != maxImageDimension {
		t.Errorf("height is %d, want %d", h, maxImageDimension)
	}
	if w >= 500 {
		t.Errorf("width %d was not reduced alongside the height", w)
	}
}

// Re-encoding a small image would cost quality for nothing.
func TestBoundedLeavesSmallImagesUntouched(t *testing.T) {
	data := pngOfSize(t, 800, 600)
	got := Image{Data: data, MimeType: "image/png"}.Bounded()
	if !bytes.Equal(got.Data, data) {
		t.Error("an in-bounds image was re-encoded")
	}
}

// Better to send something the model may still handle than to drop the
// attachment because it could not be parsed.
func TestBoundedPassesThroughUndecodableData(t *testing.T) {
	data := []byte("not an image at all")
	got := Image{Data: data, MimeType: "image/png"}.Bounded()
	if !bytes.Equal(got.Data, data) {
		t.Error("undecodable data should be passed through unchanged")
	}
}

func TestBoundedKeepsJPEGAsJPEG(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2000, 1500)), nil); err != nil {
		t.Fatal(err)
	}
	got := Image{Data: buf.Bytes(), MimeType: "image/jpeg"}.Bounded()
	if got.MimeType != "image/jpeg" {
		t.Errorf("mime type became %q", got.MimeType)
	}
	if w, _ := decodedSize(t, got.Data); w != maxImageDimension {
		t.Errorf("width is %d, want %d", w, maxImageDimension)
	}
}

// Each model server words the overflow differently, so the retry must key on
// the shape of the message rather than one exact string.
func TestIsContextOverflow(t *testing.T) {
	overflows := []string{
		`{"error":{"code":400,"message":"request (13609 tokens) exceeds the available context size (8192 tokens), try increasing it","type":"exceed_context_size_error"}}`,
		"This model's maximum context length is 8192 tokens",
		"prompt is too long for the context window",
		"too many tokens in request",
	}
	for _, msg := range overflows {
		if !isContextOverflow(errors.New(msg)) {
			t.Errorf("not recognised as an overflow: %s", msg)
		}
	}
	for _, msg := range []string{"connection refused", "model not found", "invalid api key"} {
		if isContextOverflow(errors.New(msg)) {
			t.Errorf("wrongly treated as an overflow: %s", msg)
		}
	}
}

func TestRecentTurns(t *testing.T) {
	msgs := []Message{
		{Content: "1"}, {Content: "2"}, {Content: "3"}, {Content: "4"},
	}
	got := recentTurns(msgs, 2)
	if len(got) != 2 || got[0].Content != "3" || got[1].Content != "4" {
		t.Fatalf("expected the newest two turns, got %+v", got)
	}
	if got := recentTurns(msgs, 10); len(got) != 4 {
		t.Fatalf("asking for more turns than exist should return all, got %d", len(got))
	}
}
