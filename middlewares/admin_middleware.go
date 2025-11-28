package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Check if claims exist
		claimsRaw, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Claims missing"})
			c.Abort()
			return
		}

		// 2. Convert safely
		claims, ok := claimsRaw.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}

		// 3. Extract role safely
		roleValue, ok := claims["role"]
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role value"})
			c.Abort()
			return
		}

		// 4. Check for ADMIN
		if strings.ToUpper(role) != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		// 5. Continue
		c.Next()
	}
}
