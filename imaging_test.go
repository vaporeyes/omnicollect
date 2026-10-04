// ABOUTME: Verifies content-addressed filenames, real MIME, aspect ratios, and decode budgets.
// ABOUTME: Constructs bounded image fixtures without relying on user files or huge allocations.
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"net/http"
	"strings"
	"testing"
)

func pngFixture(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
func TestImagePreservesFormatAspectAndStableIdentity(t *testing.T) {
	data := pngFixture(t, 800, 400)
	first, err := processImageBytes(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	second, err := processImageBytes(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	if first.Filename != second.Filename || !strings.HasSuffix(first.Filename, ".png") || len(first.Filename) != 68 {
		t.Fatalf("wrong content identity: %q", first.Filename)
	}
	if first.Format != "png" || !bytes.Equal(first.OriginalBytes, data) {
		t.Fatal("original modified or MIME mismatched")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(first.ThumbBytes))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" || config.Width != 400 || config.Height != 200 {
		t.Fatalf("thumbnail cropped or wrong format: %+v %s", config, format)
	}
	if http.DetectContentType(first.ThumbBytes) != "image/jpeg" {
		t.Fatal("thumbnail MIME incorrect")
	}
}
func TestImageRejectsExcessiveDimensionsBeforeDecode(t *testing.T) {
	data := pngFixture(t, 1, 1)
	binary.BigEndian.PutUint32(data[16:20], 100000)
	binary.BigEndian.PutUint32(data[20:24], 100000)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	_, err := processImageBytes(context.Background(), data)
	if err == nil || !strings.Contains(err.Error(), "megapixel") {
		t.Fatalf("unsafe dimensions accepted: %v", err)
	}
}
func TestImageBoundsAndCancellation(t *testing.T) {
	if _, err := processImageBytes(context.Background(), []byte("<svg>not an accepted image</svg>")); err == nil {
		t.Fatal("non-image accepted")
	}
	if _, err := processImageBytes(context.Background(), make([]byte, maxImageFileSize+1)); err == nil {
		t.Fatal("oversized image accepted")
	}
	for i := 0; i < cap(imageWorkers); i++ {
		imageWorkers <- struct{}{}
	}
	defer func() {
		for i := 0; i < cap(imageWorkers); i++ {
			<-imageWorkers
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := processImageBytes(ctx, pngFixture(t, 1, 1)); !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting image worker ignored cancellation: %v", err)
	}
}
