package routes

import (
	"context"
	"log"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/handlers"
	"whotterre/img_service/internal/repositories"
	"whotterre/img_service/internal/workers"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(app *gin.Engine, db *gorm.DB, taskDistributor workers.TaskDistributor) {
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

	jobRepo := repositories.NewJobRepository(db)
	imageRepo := repositories.NewUploadRepository(db)

	imageHandlers := handlers.NewImageHandlers(jobRepo, imageRepo, taskDistributor, s3Client)

	app.POST("/upload", imageHandlers.UploadImage)
	app.GET("/upload/:id/status", imageHandlers.GetJobStatus)
	app.GET("/upload/:id/result", imageHandlers.GetJobResult)
}
