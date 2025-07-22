package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/models"
	"tayaria-warranty-be/utils"

	"github.com/gin-gonic/gin"
)

// POST /api/user/warranty
func RegisterWarranty(c *gin.Context) {
	// Parse FormData
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receipt file is required and must be PDF, JPG, or PNG"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	phoneNumber := c.PostForm("phone_number")
	email := c.PostForm("email")
	purchaseDateStr := c.PostForm("purchase_date")
	carPlate := c.PostForm("car_plate")

	// Parse purchase date
	purchaseDate, err := time.Parse(time.RFC3339, purchaseDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase_date format. Use ISO8601 (RFC3339)"})
		return
	}

	// Upload to S3 (or fallback)
	receiptURL, uploadErr := utils.UploadReceiptToS3(file, header, carPlate)
	if uploadErr != nil {
		log.Printf("[WARN] S3 upload failed: %v. Using fallback URL.", uploadErr)
	}

	// Create warranty in database
	warranty, err := db.CreateWarranty(models.CreateWarrantyRequest{
		Name:         name,
		PhoneNumber:  phoneNumber,
		Email:        email,
		PurchaseDate: purchaseDate,
		CarPlate:     carPlate,
		Receipt:      receiptURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send confirmation email to user (non-blocking)
	if email != "" {
		go func() {
			if err := utils.SendWarrantyConfirmationEmail(*warranty); err != nil {
				log.Printf("Failed to send warranty confirmation email: %v", err)
			} else {
				log.Printf("Warranty confirmation email sent successfully to %s", email)
			}
		}()
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
func HasValidWarrantyByCarPlate(c *gin.Context) {
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

	// Extract S3 key from the URL (remove the base URL)
	s3Key := receiptURL
	if strings.HasPrefix(receiptURL, utils.S3BaseURL) {
		s3Key = strings.TrimPrefix(receiptURL, utils.S3BaseURL)
	}

	presignedURL, err := utils.GeneratePresignedURL(s3Key, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate pre-signed URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"receipt_url": presignedURL})
}

// GET /api/master/warranties/valid/:carPlate
func GetValidWarrantiesForTagging(c *gin.Context) {
	carPlate := c.Param("carPlate")

	warranties, err := db.GetAllValidWarrantiesForCarPlate(carPlate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(warranties) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    "No valid warranties found for this car plate",
			"warranties": []models.Warranty{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"warranties": warranties,
	})
}
