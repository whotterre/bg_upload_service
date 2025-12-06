package workers

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"whotterre/img_service/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
)

// GenerateThumbnail creates a thumbnail version of the image
func GenerateThumbnail(ctx context.Context, s3Client *s3.Client, originalPath, jobID string) (string, error) {
	// Download image from S3
	img, format, err := downloadImageFromS3(ctx, s3Client, originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	// Create thumbnail with configured dimensions
	thumbnail := imaging.Thumbnail(img, config.AppConfig.ThumbnailWidth, config.AppConfig.ThumbnailHeight, imaging.Lanczos)

	// Encode thumbnail
	var buf bytes.Buffer
	if err := encodeImage(&buf, thumbnail, format); err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	// Upload to S3
	thumbnailPath := fmt.Sprintf("thumbnails/%s%s", jobID, filepath.Ext(originalPath))
	if err := uploadToS3(ctx, s3Client, thumbnailPath, buf.Bytes(), getContentType(format)); err != nil {
		return "", fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	return thumbnailPath, nil
}
