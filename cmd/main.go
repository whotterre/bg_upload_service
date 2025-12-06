package main

import (
	"log"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/initializers"
	"whotterre/img_service/internal/repositories"
	"whotterre/img_service/internal/routes"
	"whotterre/img_service/internal/workers"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func main() {

	config.LoadConfig()

	gin.SetMode(config.AppConfig.GinMode)

	app := gin.Default()

	err := initializers.ConnectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db := initializers.GetDB()

	redisOpt := asynq.RedisClientOpt{
		Addr:     config.AppConfig.RedisAddr,
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
	}

	taskDistributor := workers.NewTaskDistributor(redisOpt)

	routes.SetupRoutes(app, db, taskDistributor)

	jobRepo := repositories.NewJobRepository(db)
	imageRepo := repositories.NewUploadRepository(db)
	taskProcessor := workers.NewRedisTaskProcessor(redisOpt, jobRepo, imageRepo)

	go func() {
		log.Println("Starting async task processor...")
		if err := taskProcessor.Start(); err != nil {
			log.Fatalf("Failed to start task processor: %v", err)
		}
	}()

	port := ":" + config.AppConfig.Port
	log.Printf("Starting image processing service on port %s", config.AppConfig.Port)

	if err := app.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
