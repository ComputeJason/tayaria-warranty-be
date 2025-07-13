package handlers

import (
	"net/http"
	"tayaria-warranty-be/db"
	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// POST /admin/login
func AdminLogin(c *gin.Context) {
	var req models.ShopLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get shop by username
	shop, err := db.GetShopByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query shop"})
		return
	}

	if shop == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password (plain text comparison)
	if req.Password != shop.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// TODO: Generate real JWT token
	// TODO: Add proper session management
	// TODO: Add token expiration
	c.JSON(http.StatusOK, models.ShopLoginResponse{
		Token: "mock-jwt-token",
		Shop:  *shop,
	})
}

// POST /api/master/account - Create new retail account
func CreateRetailAccount(c *gin.Context) {
	var req models.CreateRetailAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	existingShop, err := db.GetShopByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
		return
	}

	if existingShop != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Create new shop
	shop, err := db.CreateShop(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create retail account"})
		return
	}

	// Return response without password for security
	response := models.CreateRetailAccountResponse{
		ID:        shop.ID,
		ShopName:  shop.ShopName,
		Address:   shop.Address,
		Contact:   shop.Contact,
		Username:  shop.Username,
		Role:      shop.Role,
		CreatedAt: shop.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GET /api/master/account - Get all retail accounts
func GetRetailAccounts(c *gin.Context) {
	// Get all shops with admin role (retail accounts)
	shops, err := db.GetAllShops()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch retail accounts"})
		return
	}

	// Return the list (will be empty array if no shops found)
	c.JSON(http.StatusOK, shops)
}
