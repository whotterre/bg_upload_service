package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	app.POST("/upload", dummyHandler)
	app.GET("/upload/{id}/status", dummyHandler)
	app.GET("/upload/{id}/result", dummyHandler)
}

func dummyHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Cooking beans"})
}
