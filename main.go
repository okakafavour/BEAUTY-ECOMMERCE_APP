package main

import (
	"fmt"
	"log"
	"os"

	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("❌ Could not load .env file:", err)
	} else {
		log.Println("✅ .env loaded successfully")
	}

	// Debug print Cloudinary URL
	fmt.Println("CLOUDINARY_URL:", os.Getenv("CLOUDINARY_URL"))

	// Set middleware JWT secret
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set:", string(middlewares.JwtSecret))

	// Connect to DB
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// Create Gin router
	router := gin.Default()

	// Register all routes
	routes.SetUpRoutes(router)

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
