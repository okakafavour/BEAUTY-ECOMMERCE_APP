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
	// Load environment variables
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("⚠️ Could not load .env file, relying on environment variables")
	// }

	// Set JWT secret
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set")

	// Connect to MongoDB
	config.ConnectDB()
	fmt.Println("✅ Database connected")

	// Start email worker for async emails
	utils.StartEmailWorker()
	fmt.Println("✅ Email worker started")

	// --- TEST ADMIN EMAIL (non-blocking) ---
	go func() {
		if err := utils.SendTestAdminEmail(); err != nil {
			log.Println("⚠️ Admin test email failed:", err)
		} else {
			log.Println("✅ Admin test email sent! Check inbox.")
		}
	}()

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
		AllowOrigins: []string{
			"http://localhost:3000", // Local dev
			"*",                     // Allow all for now (change later)
		},
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
	fmt.Println("SMTP_HOST:", os.Getenv("SMTP_HOST"))
	fmt.Println("SMTP_PORT:", os.Getenv("SMTP_PORT"))
	fmt.Println("SMTP_USERNAME:", os.Getenv("SMTP_USERNAME"))
	fmt.Println("SMTP_PASSWORD:", os.Getenv("SMTP_PASSWORD") != "")
	fmt.Println("SMTP_FROM:", os.Getenv("SMTP_FROM"))

}
