package main

import (
	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/routes"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("❌ Could not load .env file:", err)
	} else {
		log.Println("✅ .env loaded successfully")
	}

	// Connect to DB
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// Create Gin router
	router := gin.Default()

	// Register all routes
	routes.SetUpRoutes(router)

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	router.Run(":8080")
}
