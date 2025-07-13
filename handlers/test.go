package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping endpoint for health check
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
