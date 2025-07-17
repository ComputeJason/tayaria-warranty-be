package handlers

import (
	"fmt"
	"log"
	"net/http"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// POST /api/user/claim
func CreateClaim(c *gin.Context) {
	var req models.CreateClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get shop_id from context (set by AdminMiddleware)
	shopID, exists := c.Get("shop_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Shop ID not found in context"})
		return
	}

	log.Printf("Creating claim for shop ID: %v (type: %T)", shopID, shopID)

	// Create a complete request with shop_id from context
	completeReq := models.CreateClaimRequest{
		CustomerName: req.CustomerName,
		PhoneNumber:  req.PhoneNumber,
		Email:        req.Email,
		CarPlate:     req.CarPlate,
	}

	// Create claim in database (includes warranty validation)
	claim, err := db.CreateClaim(completeReq, shopID.(string))
	if err != nil {
		if err.Error() == "no valid warranty found for car plate "+req.CarPlate {
			c.JSON(http.StatusNotFound, gin.H{"error": "No valid warranty found for this car plate"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, claim)
}

// GET /api/admin/claims
func GetShopClaims(c *gin.Context) {
	// Get shop_id from context (set by AdminMiddleware)
	shopID, exists := c.Get("shop_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Shop ID not found in context"})
		return
	}

	claims, err := db.GetShopClaims(shopID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claims)
}

// POST /api/user/claim/:id/tag-warranty
func TagWarrantyToClaim(c *gin.Context) {
	var req models.TagWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimID := c.Param("id")

	// Get the claim
	claim, err := db.GetClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// Check if claim is in pending status
	if claim.Status != models.PendingStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only tag warranty to claims in pending status"})
		return
	}

	// Update the claim with the warranty ID
	updatedClaim, err := db.UpdateClaimWarrantyID(claimID, req.WarrantyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/user/claim/:id/change-status
func ChangeClaimStatus(c *gin.Context) {
	var req models.UpdateClaimStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimID := c.Param("id")

	// Get the current claim
	claim, err := db.GetClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// Validate status transition
	validTransitions := map[models.ClaimStatus][]models.ClaimStatus{
		models.UnacknowledgedStatus: {models.PendingStatus},
		models.PendingStatus:        {models.ApprovedStatus, models.RejectedStatus},
	}

	allowed := false
	for _, next := range validTransitions[claim.Status] {
		if req.Status == next {
			allowed = true
			break
		}
	}

	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
		return
	}

	// Additional validation for rejected status
	if req.Status == models.RejectedStatus && req.RejectionReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason required for rejected status"})
		return
	}

	// Update claim status
	updatedClaim, err := db.UpdateClaimStatus(claimID, req.Status, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/user/claim/:id/close
func CloseClaim(c *gin.Context) {
	claimID := c.Param("id")

	// Close the claim
	claim, err := db.CloseClaim(claimID)
	if err != nil {
		if err.Error() == "claim not found or not in approved/rejected status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only close approved or rejected claims"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claim)
}

// GET /api/master/claims
func GetAllClaims(c *gin.Context) {
	// Get status from query parameter
	status := c.Query("status")

	// Validate status parameter
	if status != "unacknowledged" && status != "pending" && status != "history" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter. Must be one of: unacknowledged, pending, history"})
		return
	}

	// Get claims based on status
	claims, err := db.GetClaimsByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claims)
}

// GET /api/master/claim/:id
func GetClaimInfoByID(c *gin.Context) {
	claimID := c.Param("id")

	// Get the claim with tyre details
	claim, err := db.GetClaimWithTyreDetails(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// If claim has a warranty_id, get the warranty details
	var warranty *models.Warranty
	if claim.WarrantyID != nil {
		warranties, err := db.GetWarrantiesByCarPlate(claim.CarPlate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get warranty details: %v", err)})
			return
		}
		// Find the specific warranty
		for _, w := range warranties {
			if w.ID == *claim.WarrantyID {
				warranty = &w
				break
			}
		}
	}

	// Return combined response
	c.JSON(http.StatusOK, gin.H{
		"claim":    claim,
		"warranty": warranty,
	})
}

// POST /api/master/claim/:id/pending
func ChangeClaimStatusToPending(c *gin.Context) {
	claimID := c.Param("id")

	// Get the current claim
	claim, err := db.GetClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// Check if claim is in unacknowledged status
	if claim.Status != models.UnacknowledgedStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only change unacknowledged claims to pending"})
		return
	}

	// Update claim status to pending
	updatedClaim, err := db.UpdateClaimStatus(claimID, models.PendingStatus, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/master/claim/:id/accept
func ChangeClaimStatusToAccepted(c *gin.Context) {
	var req models.AcceptClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimID := c.Param("id")

	// Get the current claim
	claim, err := db.GetClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// Check if claim is in pending status
	if claim.Status != models.PendingStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only accept claims in pending status"})
		return
	}

	// Convert request tyre details to model
	tyreDetails := make([]models.TyreDetail, len(req.TyreDetails))
	for i, td := range req.TyreDetails {
		tyreDetails[i] = models.TyreDetail{
			Brand:        td.Brand,
			Size:         td.Size,
			TreadPattern: td.TreadPattern,
		}
	}

	// Update claim status and add tyre details
	updatedClaim, err := db.AcceptClaim(claimID, tyreDetails)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/master/claim/:id/reject
func ChangeClaimStatusToRejected(c *gin.Context) {
	var req models.RejectClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimID := c.Param("id")

	// Get the current claim
	claim, err := db.GetClaimByID(claimID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if claim == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}

	// Check if claim is in pending status
	if claim.Status != models.PendingStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only reject claims in pending status"})
		return
	}

	// Update claim status with rejection reason
	updatedClaim, err := db.RejectClaim(claimID, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}
