package handlers

import (
	"net/http"
	"time"

	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// Mock warranty data
var mockWarranties = map[string]models.Warranty{}

// POST /warranties
func RegisterWarranty(c *gin.Context) {
	var req models.CreateWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	warrantyID := "WAR-" + time.Now().Format("20060102150405")
	purchaseDate, _ := time.Parse("2006-01-02", req.PurchaseDate)
	expirationDate := purchaseDate.AddDate(0, 0, 15) // +15 days

	warranty := models.Warranty{
		ID:             warrantyID,
		CustomerName:   req.CustomerName,
		CarPlate:       req.CarPlate,
		PhoneNumber:    req.PhoneNumber,
		Email:          req.Email,
		PurchaseDate:   req.PurchaseDate,
		ReceiptURL:     req.ReceiptURL,
		Status:         "active",
		CreatedDate:    time.Now().Format("2006-01-02"),
		ExpirationDate: expirationDate.Format("2006-01-02"),
		TaggedClaimID:  "",
	}
	mockWarranties[warrantyID] = warranty
	c.JSON(http.StatusCreated, warranty)
}

// GET /warranties?car_plate=XXX
func GetWarrantiesByCarPlate(c *gin.Context) {
	carPlate := c.Query("car_plate")
	var warranties []models.Warranty
	for _, w := range mockWarranties {
		if w.CarPlate == carPlate {
			warranties = append(warranties, w)
		}
	}
	c.JSON(http.StatusOK, warranties)
}

// GET /warranties/valid?car_plate=XXX
func GetValidWarrantyByCarPlate(c *gin.Context) {
	carPlate := c.Query("car_plate")
	for _, w := range mockWarranties {
		if w.CarPlate == carPlate && w.Status == "active" {
			c.JSON(http.StatusOK, gin.H{"valid": true, "warranty": w})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"valid": false, "warranty": nil})
}

// GET /warranties/:id/receipt
func GetWarrantyReceipt(c *gin.Context) {
	warrantyID := c.Param("id")
	warranty, exists := mockWarranties[warrantyID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warranty not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"receipt_url": warranty.ReceiptURL})
}
