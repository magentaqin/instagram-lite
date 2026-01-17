package main

import (
	"instagram-lite-backend/config"
	"instagram-lite-backend/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	config.InitDB()

	// Create Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router)

	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
