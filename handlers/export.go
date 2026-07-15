package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/utils"

	"github.com/gin-gonic/gin"
)

// GET /api/master/export/warranties
func ExportWarrantiesCSV(c *gin.Context) {
	// Get all warranties from database
	warranties, err := db.GetAllWarranties()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch warranties: %v", err)})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=warranties.csv")

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"id",
		"name",
		"phone_number",
		"email",
		"purchase_date",
		"expiry_date",
		"car_plate",
		"receipt",
		"created_at",
		"updated_at",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write warranty data
	for _, warranty := range warranties {
		receiptURL := warranty.Receipt
		if s3Key := utils.ExtractS3Key(receiptURL); s3Key != receiptURL {
			if presigned, err := utils.GeneratePresignedURL(s3Key, 7*24*time.Hour); err == nil {
				receiptURL = presigned
			}
		}

		record := []string{
			warranty.ID,
			warranty.Name,
			warranty.PhoneNumber,
			warranty.Email,
			warranty.PurchaseDate.Format("2006-01-02T15:04:05Z07:00"),
			warranty.ExpiryDate.Format("2006-01-02T15:04:05Z07:00"),
			warranty.CarPlate,
			receiptURL,
			warranty.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			warranty.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write warranty record"})
			return
		}
	}
}

// GET /api/master/export/claims
func ExportClaimsCSV(c *gin.Context) {
	// Get all claims from database
	claims, err := db.GetAllClaimsForExport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch claims: %v", err)})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=claims.csv")

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"id",
		"warranty_id",
		"shop_id",
		"shop_name",
		"contact",
		"status",
		"rejection_reason",
		"date_settled",
		"date_closed",
		"customer_name",
		"phone_number",
		"email",
		"car_plate",
		"supporting_doc_url",
		"created_at",
		"updated_at",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write claim data
	for _, claim := range claims {
		// Handle nullable fields
		warrantyID := ""
		if claim.WarrantyID != nil {
			warrantyID = *claim.WarrantyID
		}

		dateSettled := ""
		if claim.DateSettled != nil {
			dateSettled = claim.DateSettled.Format("2006-01-02T15:04:05Z07:00")
		}

		dateClosed := ""
		if claim.DateClosed != nil {
			dateClosed = claim.DateClosed.Format("2006-01-02T15:04:05Z07:00")
		}

		supportingDocURL := ""
		if claim.SupportingDocURL != nil {
			rawURL := *claim.SupportingDocURL
			if s3Key := utils.ExtractS3Key(rawURL); s3Key != rawURL {
				if presigned, err := utils.GeneratePresignedURL(s3Key, 7*24*time.Hour); err == nil {
					rawURL = presigned
				}
			}
			supportingDocURL = rawURL
		}

		record := []string{
			claim.ID,
			warrantyID,
			claim.ShopID,
			claim.ShopName,
			claim.Contact,
			string(claim.Status),
			claim.RejectionReason,
			dateSettled,
			dateClosed,
			claim.CustomerName,
			claim.PhoneNumber,
			claim.Email,
			claim.CarPlate,
			supportingDocURL,
			claim.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			claim.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write claim record"})
			return
		}
	}
}
