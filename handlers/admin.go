package handlers

import (
	"net/http"
	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// Mock admin data
var mockAdmin = models.Admin{
	ID:           "admin-1",
	Username:     "admin",
	PasswordHash: "adminpass", // In production, use a hashed password
}

// POST /admin/login
func AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username != mockAdmin.Username || req.Password != mockAdmin.PasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// TODO: Generate real JWT
	c.JSON(http.StatusOK, models.AdminLoginResponse{
		Token: "mock-jwt-token",
		Admin: mockAdmin,
	})
}
