package workers

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"path/filepath"
	"whotterre/img_service/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CompressImage compresses an image to reduce file size
func CompressImage(ctx context.Context, s3Client *s3.Client, originalPath, jobID string) (string, int64, error) {
	// Download image from S3
	img, format, err := downloadImageFromS3(ctx, s3Client, originalPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to download image: %w", err)
	}

	// Compress image with configured quality
	var buf bytes.Buffer
	quality := config.AppConfig.CompressQuality
	if quality <= 0 || quality > 100 {
		quality = 85 
	}

	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return "", 0, fmt.Errorf("failed to compress image: %w", err)
	}

	compressedData := buf.Bytes()
	compressedSize := int64(len(compressedData))

	// Upload to S3
	compressedPath := fmt.Sprintf("compressed/%s%s", jobID, filepath.Ext(originalPath))
	if err := uploadToS3(ctx, s3Client, compressedPath, compressedData, getContentType(format)); err != nil {
		return "", 0, fmt.Errorf("failed to upload compressed image: %w", err)
	}

	return compressedPath, compressedSize, nil
}
