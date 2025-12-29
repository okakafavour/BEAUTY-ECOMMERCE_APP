package main

import (
	"fmt"
	"log"
	"os"

	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/controllers"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/routes"
	"beauty-ecommerce-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set JWT secret
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set")

	// Connect to MongoDB
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// Initialize email service (Brevo)
	fmt.Println("✅ Email system ready")

	// Add temporary test order (optional)
	utils.AddTestOrder(&utils.Order{
		ID:              "order_123",
		PaymentIntentID: "pi_3SajFMRhIgDY5Lro1wAWon5R",
		Status:          "pending",
	})

	// Initialize Gin router
	router := gin.Default()

	// Payment webhook
	router.POST("/payment/webhook", controllers.StripeWebhook)

	// Security headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "no-referrer-when-downgrade")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Set up all routes
	routes.SetUpRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on PORT:", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
