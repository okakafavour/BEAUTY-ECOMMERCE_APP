package utils

import (
	"os"
	"time"

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
