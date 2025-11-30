package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateToken creates a JWT token for a user
func GenerateToken(userID primitive.ObjectID, email, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "defaultsecret"
	}

	claims := jwt.MapClaims{
		"user_id": userID.Hex(),
		"email":   email,
		"role":    role,
		"exp":     jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		"iat":     jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ExtractUserIDAndRole safely extracts the user ID and role from Gin context
func ExtractUserIDAndRole(c *gin.Context) (primitive.ObjectID, string) {
	claims, exists := c.Get("user")
	if !exists {
		return primitive.NilObjectID, ""
	}

	userClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, ""
	}

	// Extract user_id
	uidStr, ok := userClaims["user_id"].(string)
	if !ok {
		fmt.Println("user_id not found in token claims")
		return primitive.NilObjectID, ""
	}

	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		fmt.Println("invalid user_id in token claims:", uidStr)
		return primitive.NilObjectID, ""
	}

	// Extract role
	role, _ := userClaims["role"].(string)
	return userID, role
}
