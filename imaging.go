// ABOUTME: Bounded image processing with accurate MIME extensions and aspect-preserving thumbnails.
// ABOUTME: Content-addressed originals allow safe, repeatable media staging before metadata commits.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"omnicollect/storage"
	"os"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type processedImageData struct {
	Filename      string
	OriginalFile  string
	ThumbFile     string
	OriginalBytes []byte
	ThumbBytes    []byte
	Width         int
	Height        int
	Format        string
}

const maxImageFileSize = 30 * 1024 * 1024
const maxImagePixels int64 = 24_000_000

var imageWorkers = make(chan struct{}, 2)
var imageExtensions = map[string]string{"jpeg": ".jpg", "png": ".png", "gif": ".gif", "webp": ".webp"}

func validateImageConfig(reader io.Reader) (string, int, int, error) {
	cfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return "", 0, 0, fmt.Errorf("not a valid image: %w", err)
	}
	if imageExtensions[format] == "" {
		return "", 0, 0, fmt.Errorf("unsupported image format: %s", format)
	}
	if cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width) > maxImagePixels/int64(cfg.Height) {
		return "", 0, 0, fmt.Errorf("image exceeds the 24 megapixel limit")
	}
	return format, cfg.Width, cfg.Height, nil
}

func acquireImageWorker(ctx context.Context) (func(), error) {
	select {
	case imageWorkers <- struct{}{}:
		return func() { <-imageWorkers }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func persistProcessedImage(ctx context.Context, media storage.MediaStore, data processedImageData) (ProcessImageResult, error) {
	if err := media.SaveOriginal(ctx, data.OriginalFile, data.OriginalBytes); err != nil {
		return ProcessImageResult{}, fmt.Errorf("saving original: %w", err)
	}
	if err := media.SaveThumbnail(ctx, data.ThumbFile, data.ThumbBytes); err != nil {
		return ProcessImageResult{}, fmt.Errorf("saving thumbnail: %w", err)
	}
	return ProcessImageResult{Filename: data.Filename, OriginalPath: data.OriginalFile, ThumbnailPath: data.ThumbFile, Width: data.Width, Height: data.Height, Format: data.Format}, nil
}

func processImageFile(ctx context.Context, sourcePath string) (processedImageData, error) {
	release, err := acquireImageWorker(ctx)
	if err != nil {
		return processedImageData{}, err
	}
	defer release()
	file, err := os.Open(sourcePath)
	if err != nil {
		return processedImageData{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxImageFileSize+1))
	if err != nil {
		return processedImageData{}, err
	}
	return decodeImage(ctx, data)
}

func processImageBytes(ctx context.Context, data []byte) (processedImageData, error) {
	release, err := acquireImageWorker(ctx)
	if err != nil {
		return processedImageData{}, err
	}
	defer release()
	return decodeImage(ctx, data)
}

func decodeImage(ctx context.Context, data []byte) (processedImageData, error) {
	if len(data) > maxImageFileSize {
		return processedImageData{}, fmt.Errorf("image exceeds the 30 MB limit")
	}
	if err := ctx.Err(); err != nil {
		return processedImageData{}, err
	}
	format, _, _, err := validateImageConfig(bytes.NewReader(data))
	if err != nil {
		return processedImageData{}, err
	}
	source, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return processedImageData{}, err
	}
	if err := ctx.Err(); err != nil {
		return processedImageData{}, err
	}
	thumb := imaging.Fit(source, 400, 400, imaging.Lanczos)
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, thumb, &jpeg.Options{Quality: 80}); err != nil {
		return processedImageData{}, err
	}
	if err := ctx.Err(); err != nil {
		return processedImageData{}, err
	}
	digest := sha256.Sum256(data)
	name := fmt.Sprintf("%x%s", digest, imageExtensions[format])
	return processedImageData{Filename: name, OriginalFile: name, ThumbFile: name, OriginalBytes: data,
		ThumbBytes: buffer.Bytes(), Width: source.Bounds().Dx(), Height: source.Bounds().Dy(), Format: format}, nil
}
