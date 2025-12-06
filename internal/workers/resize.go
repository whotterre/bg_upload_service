package workers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"whotterre/img_service/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
)

// ResizeImage resizes an image while maintaining aspect ratio
func ResizeImage(ctx context.Context, s3Client *s3.Client, originalPath, jobID string) (string, error) {
	// Download image from S3
	img, format, err := downloadImageFromS3(ctx, s3Client, originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	// Resize image maintaining aspect ratio
	resized := imaging.Fit(img, config.AppConfig.ResizeWidth, config.AppConfig.ResizeHeight, imaging.Lanczos)

	// Encode resized image
	var buf bytes.Buffer
	if err := encodeImage(&buf, resized, format); err != nil {
		return "", fmt.Errorf("failed to encode resized image: %w", err)
	}

	// Upload to S3
	resizedPath := fmt.Sprintf("resized/%s%s", jobID, filepath.Ext(originalPath))
	if err := uploadToS3(ctx, s3Client, resizedPath, buf.Bytes(), getContentType(format)); err != nil {
		return "", fmt.Errorf("failed to upload resized image: %w", err)
	}

	return resizedPath, nil
}

func encodeImage(buf *bytes.Buffer, img image.Image, format string) error {
	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(buf, img)
	default:
		return jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
	}
}

func getContentType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}
