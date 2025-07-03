package handlers

import (
	"net/http"

	"tayaria-warranty-be/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUserByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	if uuidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "UUID is required",
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(uuidStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	user, err := db.GetUserByID(uuidStr)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
				"uuid":  uuidStr,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User found",
		"user":    user,
	})
}

func GetUserByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name is required",
		})
		return
	}

	user, err := db.GetUserByFullName(name)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
				"name":  name,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User found",
		"user":    user,
	})
}
