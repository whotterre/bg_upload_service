package workers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"whotterre/img_service/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func downloadImageFromS3(ctx context.Context, s3Client *s3.Client, key string) (image.Image, string, error) {
	// Download from S3
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.AppConfig.S3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the image data
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return img, format, nil
}


func uploadToS3(ctx context.Context, s3Client *s3.Client, key string, data []byte, contentType string) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.S3BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	return nil
}

// GetS3URL generates the S3 URL for a given key
func GetS3URL(key string) string {
	region := config.AppConfig.AWSRegion
	bucket := config.AppConfig.S3BucketName
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
}
