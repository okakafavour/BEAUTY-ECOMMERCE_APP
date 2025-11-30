package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"beauty-ecommerce-backend/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Ensure JWT_SECRET matches your server
	os.Setenv("JWT_SECRET", "defaultsecret") // or the value in your .env

	// Use a real user ID from MongoDB, or generate a new one
	userID, err := primitive.ObjectIDFromHex("692aa2554d544abd4b4288a6")
	if err != nil {
		log.Fatal("Invalid user ID:", err)
	}

	email := "user@example.com"
	role := "CUSTOMER"

	// Generate JWT token
	token, err := utils.GenerateToken(userID, email, role)
	if err != nil {
		log.Fatal("Failed to generate token:", err)
	}

	fmt.Println("✅ JWT Token (copy this into Postman):")
	fmt.Println(token)
	fmt.Println()
	fmt.Println("Authorization header example:")
	fmt.Printf("Authorization: Bearer %s\n", token)
	fmt.Println()
	fmt.Println("Token expires in 72 hours from now:", time.Now().Add(72*time.Hour))
}
