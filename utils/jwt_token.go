package utils

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GenerateToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
