package routes

import (
	"net/http"
	"whotterre/img_service/internal/workers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(app *gin.Engine, db *gorm.DB, taskDistributor workers.TaskDistributor) {
	app.POST("/upload", dummyHandler)
	app.GET("/upload/{id}/status", dummyHandler)
	app.GET("/upload/{id}/result", dummyHandler)
}

func dummyHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Cooking beans"})
}
