package models

import (
	"time"

	"gorm.io/gorm"
)

type UploadedImage struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	JobID          string         `gorm:"type:varchar(36);index" json:"job_id"`
	Job            *Job           `gorm:"foreignKey:JobID" json:"job,omitempty"`
	OriginalURL    string         `gorm:"type:text" json:"original_url"`
	OriginalPath   string         `gorm:"type:text" json:"original_path"`
	OriginalSize   int64          `gorm:"type:bigint" json:"original_size"`
	ResizedURL     string         `gorm:"type:text" json:"resized_url,omitempty"`
	ResizedPath    string         `gorm:"type:text" json:"resized_path,omitempty"`
	CompressedURL  string         `gorm:"type:text" json:"compressed_url,omitempty"`
	CompressedPath string         `gorm:"type:text" json:"compressed_path,omitempty"`
	CompressedSize int64          `gorm:"type:bigint" json:"compressed_size,omitempty"`
	ThumbnailURL   string         `gorm:"type:text" json:"thumbnail_url,omitempty"`
	ThumbnailPath  string         `gorm:"type:text" json:"thumbnail_path,omitempty"`
	Width          int            `gorm:"type:int" json:"width"`
	Height         int            `gorm:"type:int" json:"height"`
	Format         string         `gorm:"type:varchar(10)" json:"format"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UploadedImage) TableName() string {
	return "uploaded_images"
}
