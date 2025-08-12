package imageproc

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/gen2brain/webp"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

type ProcessedImage struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

func (p *Processor) ValidateImage(data []byte) error {
	if !filetype.IsImage(data) {
		return fmt.Errorf("file is not a valid image")
	}

	kind, err := filetype.Match(data)
	if err != nil {
		return fmt.Errorf("failed to detect file type: %w", err)
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowedTypes[kind.MIME.Value] {
		return fmt.Errorf("unsupported image type: %s", kind.MIME.Value)
	}

	return nil
}

// ProcessProfileImage resizes an image and converts it to WebP format.
func (p *Processor) ProcessProfileImage(data []byte) (*ProcessedImage, error) {
    if err := p.ValidateImage(data); err != nil {
        return nil, err
    }

    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, fmt.Errorf("failed to decode image: %w", err)
    }

    resized := imaging.Fill(img, 400, 400, imaging.Center, imaging.Lanczos)

    var buf bytes.Buffer
    options := webp.Options{
        Lossless: false,
        Quality:  85,
    }
    
    if err := webp.Encode(&buf, resized, options); err != nil {
        return nil, fmt.Errorf("failed to encode image to webp: %w", err)
    }

    return &ProcessedImage{
        Data:        buf.Bytes(),
        ContentType: "image/webp",
        Width:       400,
        Height:      400,
    }, nil
}

func (p *Processor) CreateThumbnail(data []byte, size int) (*ProcessedImage, error) {
	// Validate the image first
	if err := p.ValidateImage(data); err != nil {
		return nil, err
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create thumbnail
	thumbnail := imaging.Resize(img, size, size, imaging.Lanczos)

	// Encode as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return &ProcessedImage{
		Data:        buf.Bytes(),
		ContentType: "image/jpeg",
		Width:       size,
		Height:      size,
	}, nil
}

func (p *Processor) GetImageDimensions(data []byte) (int, int, error) {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}
	return img.Width, img.Height, nil
}

func (p *Processor) DetectContentType(data []byte) (string, error) {
	kind, err := filetype.Match(data)
	if err != nil {
		return "", fmt.Errorf("failed to detect content type: %w", err)
	}
	return kind.MIME.Value, nil
}
