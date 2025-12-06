package main

import (
	"fmt"
	"log"
	"os"

	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/routes"
	"beauty-ecommerce-backend/utils"

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

	// Example order for testing (remove in production)
	utils.AddTestOrder(&utils.Order{
		ID:              "order_123",
		PaymentIntentID: "pi_3SajFMRhIgDY5Lro1wAWon5R",
		Status:          "pending",
	})

	// Create Gin router
	router := gin.Default()

	// Register all routes (including payment/webhook)
	routes.SetUpRoutes(router)

	// ❌ REMOVE THIS → it causes duplicate route
	// router.POST("/payment/webhook", utils.StripeWebhookHandler)

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
