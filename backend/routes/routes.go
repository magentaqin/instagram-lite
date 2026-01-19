package routes

import (
	"net/http"

	"instagram-lite-backend/config"
	"instagram-lite-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
  // Health check
  router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
      "status":  "ok",
      "message": "Instagram Lite Backend is running!",
    })
  })

  // API v1 routes
  v1 := router.Group("/api/v1")

  // Upload route
  if config.Uploader != nil {
    uploadHandler := handlers.NewUploadHandler(config.Uploader)
    v1.POST("/upload", uploadHandler.Upload)
  } else {
    v1.POST("/upload", func(c *gin.Context) {
      c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
    })
  }

  // Post routes
  postsHandler := handlers.NewPostsHandler(config.DB) 
  v1.POST("/posts", postsHandler.CreatePost)
  v1.GET("/posts", postsHandler.ListPosts)
}