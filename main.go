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
	"github.com/joho/godotenv"
)

func main() {

	// Load .env only for local development (optional)
	_ = godotenv.Load()

	// Set JWT secret
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set")

	// Connect to database
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// TEMP TEST ORDER (optional)
	utils.AddTestOrder(&utils.Order{
		ID:              "order_123",
		PaymentIntentID: "pi_3SajFMRhIgDY5Lro1wAWon5R",
		Status:          "pending",
	})

	// Start router
	router := gin.Default()

	// Register Stripe webhook BEFORE middleware
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

	// CORS Config
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000", // Local dev
			"*",                     // Allow all for now (change later)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// App routes
	routes.SetUpRoutes(router)

	// Get PORT from Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local
	}

	fmt.Println("🚀 Server running on PORT:", port)

	// Start server
	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
