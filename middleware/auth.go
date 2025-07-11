package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Remove AuthMiddleware, keep AdminMiddleware only
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, you would check the user's role from the JWT token
		// For mock purposes, we'll just check if the token contains "admin"
		authHeader := c.GetHeader("Authorization")
		if !strings.Contains(authHeader, "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
