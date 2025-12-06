package repositories

import (
	"whotterre/img_service/internal/models"

	"gorm.io/gorm"
)

// UploadRepository defines the interface for upload-related database operations
type UploadRepository interface {
	CreateJob(job *models.Job) error
	GetJobByID(id string) (*models.Job, error)
	UpdateJobStatus(id string, status string, progress int) error
	UpdateJobError(id string, errorMessage string) error
	CompleteJob(id string) error

	// Image operations
	CreateImage(image *models.UploadedImage) error
	GetImageByID(id string) (*models.UploadedImage, error)
	GetImageByJobID(jobID string) (*models.UploadedImage, error)
	UpdateImageURLs(id string, resizedURL, compressedURL, thumbnailURL string) error
	UpdateImagePaths(id string, resizedPath, compressedPath, thumbnailPath string) error

	// Combined operations
	GetJobWithImage(jobID string) (*models.Job, *models.UploadedImage, error)
}

type uploadRepository struct {
	db *gorm.DB
}

// NewUploadRepository creates a new instance of UploadRepository
func NewUploadRepository(db *gorm.DB) UploadRepository {
	return &uploadRepository{db: db}
}

// CreateJob creates a new job in the database
func (r *uploadRepository) CreateJob(job *models.Job) error {
	return r.db.Create(job).Error
}

// GetJobByID retrieves a job by its ID
func (r *uploadRepository) GetJobByID(id string) (*models.Job, error) {
	var job models.Job
	err := r.db.Where("id = ?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateJobStatus updates the job status and progress
func (r *uploadRepository) UpdateJobStatus(id string, status string, progress int) error {
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   status,
			"progress": progress,
		}).Error
}

// UpdateJobError updates the job with an error message and sets status to failed
func (r *uploadRepository) UpdateJobError(id string, errorMessage string) error {
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.StatusFailed,
			"error_message": errorMessage,
		}).Error
}

// CompleteJob marks a job as completed
func (r *uploadRepository) CompleteJob(id string) error {
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.StatusCompleted,
			"progress":     100,
			"completed_at": gorm.Expr("NOW()"),
		}).Error
}

// CreateImage creates a new uploaded image record
func (r *uploadRepository) CreateImage(image *models.UploadedImage) error {
	return r.db.Create(image).Error
}

// GetImageByID retrieves an image by its ID
func (r *uploadRepository) GetImageByID(id string) (*models.UploadedImage, error) {
	var image models.UploadedImage
	err := r.db.Where("id = ?", id).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByJobID retrieves an image by job ID
func (r *uploadRepository) GetImageByJobID(jobID string) (*models.UploadedImage, error) {
	var image models.UploadedImage
	err := r.db.Where("job_id = ?", jobID).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// UpdateImageURLs updates the processed image URLs
func (r *uploadRepository) UpdateImageURLs(id string, resizedURL, compressedURL, thumbnailURL string) error {
	return r.db.Model(&models.UploadedImage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"resized_url":    resizedURL,
			"compressed_url": compressedURL,
			"thumbnail_url":  thumbnailURL,
		}).Error
}

// UpdateImagePaths updates the processed image paths
func (r *uploadRepository) UpdateImagePaths(id string, resizedPath, compressedPath, thumbnailPath string) error {
	return r.db.Model(&models.UploadedImage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"resized_path":    resizedPath,
			"compressed_path": compressedPath,
			"thumbnail_path":  thumbnailPath,
		}).Error
}

// GetJobWithImage retrieves a job along with its associated image
func (r *uploadRepository) GetJobWithImage(jobID string) (*models.Job, *models.UploadedImage, error) {
	var job models.Job
	var image models.UploadedImage

	// Get job
	if err := r.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		return nil, nil, err
	}

	// Get associated image
	if err := r.db.Where("job_id = ?", jobID).First(&image).Error; err != nil {
		return &job, nil, err
	}

	return &job, &image, nil
}
