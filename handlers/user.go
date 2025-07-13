package handlers

import (
	"net/http"

	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// Mock retailer/shop data
var mockShop = models.Shop{
	ID:       "shop-1",
	ShopName: "Tayaria Main Shop",
	Address:  "123 Main St",
	Contact:  "+60123456789",
	Username: "retailer",
	Password: "retailerpass",
}

// POST /retailer/login
func RetailerLogin(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username != mockShop.Username || req.Password != mockShop.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// TODO: Generate real JWT
	c.JSON(http.StatusOK, gin.H{
		"token": "mock-jwt-token",
		"shop":  mockShop,
	})
}
