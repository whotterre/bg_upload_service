package repositories

import (
	"time"
	"whotterre/img_service/internal/models"

	"gorm.io/gorm"
)

// JobRepository defines the interface for job-related database operations
type JobRepository interface {
	Create(job *models.Job) error
	GetByID(id string) (*models.Job, error)
	UpdateStatus(id string, status string, progress int) error
	UpdateError(id string, errorMessage string) error
	Complete(id string) error
	Delete(id string) error
	GetAllPending() ([]models.Job, error)
	GetAllProcessing() ([]models.Job, error)
}

type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new instance of JobRepository
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

// Create creates a new job in the database
func (r *jobRepository) Create(job *models.Job) error {
	return r.db.Create(job).Error
}

// GetByID retrieves a job by its ID
func (r *jobRepository) GetByID(id string) (*models.Job, error) {
	var job models.Job
	err := r.db.Where("id = ?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateStatus updates the job status and progress
func (r *jobRepository) UpdateStatus(id string, status string, progress int) error {
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   status,
			"progress": progress,
		}).Error
}

// UpdateError updates the job with an error message and sets status to failed
func (r *jobRepository) UpdateError(id string, errorMessage string) error {
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.StatusFailed,
			"error_message": errorMessage,
		}).Error
}

// Complete marks a job as completed with timestamp
func (r *jobRepository) Complete(id string) error {
	now := time.Now()
	return r.db.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.StatusCompleted,
			"progress":     100,
			"completed_at": &now,
		}).Error
}

// Delete soft deletes a job
func (r *jobRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Job{}).Error
}

// GetAllPending retrieves all pending jobs
func (r *jobRepository) GetAllPending() ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.Where("status = ?", models.StatusPending).
		Order("created_at ASC").
		Find(&jobs).Error
	return jobs, err
}

// GetAllProcessing retrieves all currently processing jobs
func (r *jobRepository) GetAllProcessing() ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.Where("status = ?", models.StatusProcessing).
		Order("created_at ASC").
		Find(&jobs).Error
	return jobs, err
}
