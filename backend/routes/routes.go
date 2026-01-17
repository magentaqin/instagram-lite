package routes

import (
	"instagram-lite-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Instagram Lite Backend is running！",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Post routes
		v1.POST("/posts", controllers.CreatePost)
	}
}
