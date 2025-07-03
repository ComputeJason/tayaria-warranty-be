package handlers

import (
	"net/http"
	"time"

	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Mock user data
var mockUsers = map[string]models.User{
	"000": {
		PhoneNumber: "000",
		Name:        "Admin User",
		Email:       "admin@example.com",
	},
	"111": {
		PhoneNumber: "111",
		Name:        "Customer User",
		Email:       "customer@example.com",
	},
}

// Mock retailer/shop data
var mockShop = models.Shop{
	ID:           "shop-1",
	ShopName:     "Tayaria Main Shop",
	Address:      "123 Main St",
	Contact:      "+60123456789",
	Username:     "retailer",
	PasswordHash: "retailerpass",
}

func LoginUser(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := mockUsers[req.PhoneNumber]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"phoneNumber": user.PhoneNumber,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("mock-secret-key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: tokenString,
		User:  user,
	})
}

func LogoutUser(c *gin.Context) {
	// In a real application, you would invalidate the token
	// For mock purposes, we'll just return success
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func GetUserInformation(c *gin.Context) {
	phoneNumber := c.Param("phoneNumber")
	user, exists := mockUsers[phoneNumber]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func EditUserInformation(c *gin.Context) {
	phoneNumber := c.Param("phoneNumber")
	user, exists := mockUsers[phoneNumber]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var updatedUser models.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user information
	user.Name = updatedUser.Name
	user.Email = updatedUser.Email
	user.Address = updatedUser.Address
	mockUsers[phoneNumber] = user

	c.JSON(http.StatusOK, user)
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
	if req.Username != mockShop.Username || req.Password != mockShop.PasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// TODO: Generate real JWT
	c.JSON(http.StatusOK, gin.H{
		"token": "mock-jwt-token",
		"shop":  mockShop,
	})
}
