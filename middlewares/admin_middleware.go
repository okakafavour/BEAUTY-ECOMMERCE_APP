package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware checks if the user is admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Example: expecting role in header (replace with your JWT logic)
		role := c.GetHeader("Role")
		if strings.ToLower(role) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
