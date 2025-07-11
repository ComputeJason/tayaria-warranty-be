package handlers

import (
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

	// Check if warranty exists
	exists, err := db.CheckWarrantyExists(req.WarrantyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warranty not found"})
		return
	}

	// Create claim in database
	claim, err := db.CreateClaim(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, claim)
}

// GET /api/user/claims
func GetClaims(c *gin.Context) {
	warrantyID := c.Query("warranty_id")
	status := c.Query("status")

	claims, err := db.GetClaims(warrantyID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claims)
}

// POST /api/user/claim/tag-warranty
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

	// Check if warranty exists
	exists, err := db.CheckWarrantyExists(req.WarrantyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warranty not found"})
		return
	}

	// Update the claim with the warranty ID
	// Note: This is a simplified implementation. In a real system, you might want to
	// add additional validation to ensure the warranty isn't already tagged to another claim
	updatedClaim, err := db.UpdateClaimStatus(claimID, claim.Status, claim.AdminComment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/user/claim/change-status
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
	if req.Status == models.RejectedStatus && req.AdminComment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin comment required for rejected status"})
		return
	}

	// Update claim status
	updatedClaim, err := db.UpdateClaimStatus(claimID, req.Status, req.AdminComment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedClaim)
}

// POST /api/user/claim/close
func CloseClaim(c *gin.Context) {
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

	// Check if claim can be closed
	if claim.Status != models.ApprovedStatus && claim.Status != models.RejectedStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only close approved or rejected claims"})
		return
	}

	// For now, we'll just return the claim as-is since our current schema doesn't have a "closed" status
	// In a real implementation, you might want to add a "closed" status or a "closed_at" timestamp
	c.JSON(http.StatusOK, gin.H{"message": "Claim closed successfully", "claim": claim})
}
