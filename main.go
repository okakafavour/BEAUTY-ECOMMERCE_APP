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
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("❌ Could not load .env file:", err)
	} else {
		log.Println("✅ .env loaded successfully")
	}

	// Set JWT secret for middleware
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set:", string(middlewares.JwtSecret))

	// Connect to database
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// TEMP: example order (remove later)
	utils.AddTestOrder(&utils.Order{
		ID:              "order_123",
		PaymentIntentID: "pi_3SajFMRhIgDY5Lro1wAWon5R",
		Status:          "pending",
	})

	// Create router
	router := gin.Default()

	// ===== REGISTER STRIPE WEBHOOK BEFORE MIDDLEWARE =====
	router.POST("/payment/webhook", controllers.StripeWebhook)

	// ==== SECURITY HEADERS ====
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "no-referrer-when-downgrade")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// ==== CORS CONFIG ====
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",            // local dev
			"https://your-frontend-domain.com", // production
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	//=== SET UP ROUTES =====
	routes.SetUpRoutes(router)

	//==== START SERVER =====
	fmt.Println("🚀 Server running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
