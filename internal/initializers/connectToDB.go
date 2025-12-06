package initializers

import (
	"fmt"
	"log"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectToDB initializes the database connection and runs migrations
func ConnectToDB() error {
	var err error

	dsn := config.AppConfig.DatabaseURL
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set in configuration")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Failed to connect to Postgres database: %v", err)
		return err
	}

	log.Println("Database connection established successfully")

	// Run auto-migrations
	if err := DB.AutoMigrate(&models.Job{}, &models.UploadedImage{}); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
