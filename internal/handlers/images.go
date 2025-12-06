package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/models"
	"whotterre/img_service/internal/repositories"
	"whotterre/img_service/internal/workers"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandlers struct {
	jobRepo         repositories.JobRepository
	imageRepo       repositories.UploadRepository
	taskDistributor workers.TaskDistributor
	s3Client        *s3.Client
}

func NewImageHandlers(
	jobRepo repositories.JobRepository,
	imageRepo repositories.UploadRepository,
	taskDistributor workers.TaskDistributor,
	s3Client *s3.Client,
) *ImageHandlers {
	return &ImageHandlers{
		jobRepo:         jobRepo,
		imageRepo:       imageRepo,
		taskDistributor: taskDistributor,
		s3Client:        s3Client,
	}
}

func (h *ImageHandlers) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// 2. Validate size
	if header.Size > config.AppConfig.MaxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large"})
		return
	}

	// 3. Validate content type (basic check)
	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file format"})
		return
	}

	jobID := uuid.New().String()
	imageID := uuid.New().String()
	originalPath := fmt.Sprintf("originals/%s%s", jobID, ext)

	_, err = h.s3Client.PutObject(c.Request.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.S3BucketName),
		Key:         aws.String(originalPath),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to storage"})
		return
	}

	// 6. Create Job record
	job := &models.Job{
		ID:       jobID,
		Status:   models.StatusPending,
		Progress: 0,
	}
	if err := h.jobRepo.Create(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}

	// 7. Create Image record
	originalURL := workers.GetS3URL(originalPath)
	imgRecord := &models.UploadedImage{
		ID:           imageID,
		JobID:        jobID,
		OriginalPath: originalPath,
		OriginalURL:  originalURL,
		OriginalSize: header.Size,
		Format:       ext[1:], // remove dot
	}
	if err := h.imageRepo.CreateImage(imgRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create image record"})
		return
	}

	// 8. Enqueue Task
	payload := &workers.ProcessImagePayload{
		JobID:        jobID,
		ImageID:      imageID,
		OriginalPath: originalPath,
	}
	if err := h.taskDistributor.DistributeProcessImageTask(c.Request.Context(), payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue processing task"})
		return
	}

	// 9. Return Response
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  models.StatusPending,
		"message": "Image uploaded and queued for processing",
		"links": gin.H{
			"status": fmt.Sprintf("/upload/%s/status", jobID),
			"result": fmt.Sprintf("/upload/%s/result", jobID),
		},
	})
}

func (h *ImageHandlers) GetJobStatus(c *gin.Context) {
	id := c.Param("id")
	job, err := h.jobRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":        job.ID,
		"status":        job.Status,
		"progress":      job.Progress,
		"error_message": job.ErrorMessage,
		"created_at":    job.CreatedAt,
		"completed_at":  job.CompletedAt,
	})
}

func (h *ImageHandlers) GetJobResult(c *gin.Context) {
	id := c.Param("id")
	job, image, err := h.imageRepo.GetJobWithImage(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job or image not found"})
		return
	}

	if job.Status != models.StatusCompleted {
		c.JSON(http.StatusOK, gin.H{
			"job_id":  job.ID,
			"status":  job.Status,
			"message": "Processing not yet completed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id": job.ID,
		"status": job.Status,
		"original": gin.H{
			"url":  image.OriginalURL,
			"size": image.OriginalSize,
		},
		"processed": gin.H{
			"resized_url":     image.ResizedURL,
			"compressed_url":  image.CompressedURL,
			"thumbnail_url":   image.ThumbnailURL,
			"compressed_size": image.CompressedSize,
		},
		"metadata": gin.H{
			"width":  image.Width,
			"height": image.Height,
			"format": image.Format,
		},
	})
}
