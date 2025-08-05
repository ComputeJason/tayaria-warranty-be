package handlers

import (
	"net/http"
	"time"

	"tayaria-warranty-be/db"

	"github.com/gin-gonic/gin"
)

// DBPing endpoint for database health check
func DBPing(c *gin.Context) {
	err := db.PingDatabase()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database ping failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Database connection healthy",
		"time":    time.Now().Format(time.RFC3339),
	})
}
