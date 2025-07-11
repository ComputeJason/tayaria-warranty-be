package handlers

import (
	"net/http"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// POST /api/user/warranty
func RegisterWarranty(c *gin.Context) {
	var req models.CreateWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create warranty in database
	warranty, err := db.CreateWarranty(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, warranty)
}

// GET /api/user/warranties/car-plate/:carPlate
func GetWarrantiesByCarPlate(c *gin.Context) {
	carPlate := c.Param("carPlate")

	warranties, err := db.GetWarrantiesByCarPlate(carPlate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, warranties)
}

// GET /api/user/warranties/valid/:carPlate
func GetValidWarrantyByCarPlate(c *gin.Context) {
	carPlate := c.Param("carPlate")

	warranty, err := db.GetValidWarrantyByCarPlate(carPlate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if warranty == nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "warranty": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true, "warranty": warranty})
}

// GET /api/user/warranty/receipt/:id
func GetWarrantyReceipt(c *gin.Context) {
	warrantyID := c.Param("id")

	receiptURL, err := db.GetWarrantyReceipt(warrantyID)
	if err != nil {
		if err.Error() == "warranty not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Warranty not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"receipt_url": receiptURL})
}
