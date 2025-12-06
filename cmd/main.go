package main

import (
	"log"
	"whotterre/img_service/internal/config"
	"whotterre/img_service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	config.LoadConfig()

	gin.SetMode(config.AppConfig.GinMode)

	app := gin.Default()

	routes.SetupRoutes(app)

	port := ":" + config.AppConfig.Port
	log.Printf("Starting image processing service on port %s", config.AppConfig.Port)

	if err := app.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
