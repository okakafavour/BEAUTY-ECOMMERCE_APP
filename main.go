package main

import (
	"fmt"
	"log"
	"os"

	"beauty-ecommerce-backend/config"
	"beauty-ecommerce-backend/middlewares"
	"beauty-ecommerce-backend/routes"
	"beauty-ecommerce-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Could not load .env file:", err)
	}

	// JWT Secret
	middlewares.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("✅ JwtSecret set")

	// Connect DB
	config.ConnectDB()
	fmt.Println("✅ Connected to MongoDB:", config.DB.Name())

	// Send test admin email (non-blocking)
	go func() {
		if err := utils.SendTestAdminEmail(); err != nil {
			log.Println("⚠️ Admin test email failed:", err)
		} else {
			log.Println("✅ Admin test email sent! Check inbox.")
		}
	}()

	// TEMP TEST ORDER DATA
	utils.AddTestOrder(&utils.Order{
		ID:              "order_123",
		PaymentIntentID: "pi_3SajFMRhIgDY5Lro1wAWon5R",
		Status:          "pending",
	})

	// Gin setup
	router := gin.Default()

	// Security headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "no-referrer-when-downgrade")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetUpRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on PORT:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
