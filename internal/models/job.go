package models

import (
	"time"

	"gorm.io/gorm"
)

// Job represents the processing job status
type Job struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Status       string         `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Progress     int            `gorm:"type:int;default:0" json:"progress"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}


func (Job) TableName() string {
	return "jobs"
}

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)
