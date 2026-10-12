package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/utils"

	"github.com/gin-gonic/gin"
)

// parseExportDateRange reads and validates the optional "from"/"to" query params shared by the
// export endpoints. An empty string for either means that side is unbounded.
func parseExportDateRange(c *gin.Context) (from string, to string, err error) {
	from = c.Query("from")
	to = c.Query("to")

	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return "", "", fmt.Errorf("invalid 'from' date, expected YYYY-MM-DD")
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return "", "", fmt.Errorf("invalid 'to' date, expected YYYY-MM-DD")
		}
	}

	return from, to, nil
}

// exportFilename builds a download filename that reflects the applied date range, e.g.
// "claims_2026-01-01_to_2026-03-01.csv", falling back to "claims.csv" when unfiltered.
func exportFilename(base, from, to string) string {
	if from == "" && to == "" {
		return base + ".csv"
	}
	if from == "" {
		return fmt.Sprintf("%s_up_to_%s.csv", base, to)
	}
	if to == "" {
		return fmt.Sprintf("%s_from_%s.csv", base, from)
	}
	return fmt.Sprintf("%s_%s_to_%s.csv", base, from, to)
}

// GET /api/master/export/warranties?from=YYYY-MM-DD&to=YYYY-MM-DD
// Both query params are optional; omitting both exports every warranty. When present, they
// filter to warranties purchased within [from, to] (inclusive).
func ExportWarrantiesCSV(c *gin.Context) {
	from, to, err := parseExportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	warranties, err := db.GetAllWarranties(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch warranties: %v", err)})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s", exportFilename("warranties", from, to)))

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
		"brand",
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
			warranty.Brand,
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

// GET /api/master/export/claims?from=YYYY-MM-DD&to=YYYY-MM-DD
// Both query params are optional; omitting both exports every claim. When present, they filter
// to claims created within [from, to] (inclusive).
func ExportClaimsCSV(c *gin.Context) {
	from, to, err := parseExportDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := db.GetAllClaimsForExport(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch claims: %v", err)})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s", exportFilename("claims", from, to)))

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
		"brand",
		"quantity",
		"requested_tyre_details",
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

		brand := ""
		if claim.Brand != nil {
			brand = *claim.Brand
		}

		quantity := ""
		if claim.Quantity != nil {
			quantity = fmt.Sprintf("%d", *claim.Quantity)
		}

		tyreParts := make([]string, len(claim.RequestedTyreDetails))
		for i, td := range claim.RequestedTyreDetails {
			tyreParts[i] = fmt.Sprintf("%s (%s)", td.Size, td.TreadPattern)
		}
		requestedTyreDetails := strings.Join(tyreParts, "; ")

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
			brand,
			quantity,
			requestedTyreDetails,
			claim.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			claim.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write claim record"})
			return
		}
	}
}
