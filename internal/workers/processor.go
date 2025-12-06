package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/models"
	"whotterre/img_service/internal/repositories"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	ProcessImageTask(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server    *asynq.Server
	jobRepo   repositories.JobRepository
	imageRepo repositories.UploadRepository
	s3Client  *s3.Client
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, jobRepo repositories.JobRepository, imageRepo repositories.UploadRepository) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.AppConfig.AWSRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AppConfig.AWSAccessKeyID,
			config.AppConfig.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	return &RedisTaskProcessor{
		server:    server,
		jobRepo:   jobRepo,
		imageRepo: imageRepo,
		s3Client:  s3Client,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskProcessImage, processor.ProcessImageTask)
	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) ProcessImageTask(ctx context.Context, task *asynq.Task) error {
	var payload ProcessImagePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Processing image for job: %s", payload.JobID)

	if err := processor.jobRepo.UpdateStatus(payload.JobID, models.StatusProcessing, 10); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	resizedPath, err := ResizeImage(ctx, processor.s3Client, payload.OriginalPath, payload.JobID)
	if err != nil {
		processor.jobRepo.UpdateError(payload.JobID, fmt.Sprintf("resize failed: %v", err))
		return fmt.Errorf("resize failed: %w", err)
	}
	processor.jobRepo.UpdateStatus(payload.JobID, models.StatusProcessing, 40)

	compressedPath, compressedSize, err := CompressImage(ctx, processor.s3Client, payload.OriginalPath, payload.JobID)
	if err != nil {
		processor.jobRepo.UpdateError(payload.JobID, fmt.Sprintf("compress failed: %v", err))
		return fmt.Errorf("compress failed: %w", err)
	}
	processor.jobRepo.UpdateStatus(payload.JobID, models.StatusProcessing, 70)

	thumbnailPath, err := GenerateThumbnail(ctx, processor.s3Client, payload.OriginalPath, payload.JobID)
	if err != nil {
		processor.jobRepo.UpdateError(payload.JobID, fmt.Sprintf("thumbnail failed: %v", err))
		return fmt.Errorf("thumbnail failed: %w", err)
	}
	processor.jobRepo.UpdateStatus(payload.JobID, models.StatusProcessing, 90)

	resizedURL := GetS3URL(resizedPath)
	compressedURL := GetS3URL(compressedPath)
	thumbnailURL := GetS3URL(thumbnailPath)

	if err := processor.imageRepo.UpdateImageURLs(payload.ImageID, resizedURL, compressedURL, thumbnailURL); err != nil {
		return fmt.Errorf("failed to update image URLs: %w", err)
	}

	if err := processor.imageRepo.UpdateImagePaths(payload.ImageID, resizedPath, compressedPath, thumbnailPath); err != nil {
		return fmt.Errorf("failed to update image paths: %w", err)
	}

	image, err := processor.imageRepo.GetImageByID(payload.ImageID)
	if err == nil {
		processor.imageRepo.UpdateImageURLs(payload.ImageID, resizedURL, compressedURL, thumbnailURL)
		if image.CompressedSize == 0 {
			processor.imageRepo.GetImageByID(payload.ImageID)
		}
		_ = compressedSize
	}

	if err := processor.jobRepo.Complete(payload.JobID); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	log.Printf("Successfully processed image for job: %s", payload.JobID)
	return nil
}
