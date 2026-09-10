package ai

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"

	"golang.org/x/image/draw"

	// Registered so Decode recognises what users actually upload.
	_ "image/gif"
)

// maxImageDimension bounds the longest side of an image before it is sent to a
// vision model.
//
// Cost is driven by pixels, not bytes: qwen2.5-VL turns roughly every 28x28
// block into a token, so a 4K screenshot alone is over 13,000 tokens and
// overflows an 8,192-token context before the conversation is even added. At
// 1024px the same screenshot costs a few hundred tokens, and text in a
// screenshot stays legible — which is the thing that must survive.
const maxImageDimension = 1024

// jpegQuality is high enough that downscaled text stays sharp for OCR.
const jpegQuality = 90

// Bounded returns the image scaled down so its longest side is at most
// maxImageDimension, re-encoded in place. An image already within bounds, or
// one that cannot be decoded, is returned unchanged — a request that might work
// is better than dropping the attachment outright.
func (i Image) Bounded() Image {
	src, format, err := image.Decode(bytes.NewReader(i.Data))
	if err != nil {
		log.Printf("chat AI: could not decode image (%s), sending as-is: %v", i.MimeType, err)
		return i
	}

	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxImageDimension && h <= maxImageDimension {
		return i
	}

	// Preserve the aspect ratio; the longest side lands on the cap.
	scale := float64(maxImageDimension) / float64(w)
	if h > w {
		scale = float64(maxImageDimension) / float64(h)
	}
	dstW, dstH := int(float64(w)*scale), int(float64(h)*scale)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	// CatmullRom keeps small text readable; a cheaper filter smears it into
	// something the model reads as the wrong digits.
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	var buf bytes.Buffer
	mime := "image/jpeg"
	if format == "png" {
		// Screenshots and documents stay lossless, where flat colour keeps PNG
		// small anyway and crisp edges matter most for reading text.
		if err := png.Encode(&buf, dst); err != nil {
			return i
		}
		mime = "image/png"
	} else if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return i
	}

	log.Printf("chat AI: scaled image %dx%d -> %dx%d (%d -> %d bytes)",
		w, h, dstW, dstH, len(i.Data), buf.Len())
	return Image{Data: buf.Bytes(), MimeType: mime}
}
