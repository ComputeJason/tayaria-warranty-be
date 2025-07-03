package handlers

import (
	"net/http"
	"time"

	"tayaria-warranty-be/models"

	"github.com/gin-gonic/gin"
)

// Mock claim data
var mockClaims = map[string]models.Claim{}

// POST /claims
func CreateClaim(c *gin.Context) {
	var req models.CreateClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claimID := "CLAIM-" + time.Now().Format("20060102150405")
	claim := models.Claim{
		ID:               claimID,
		CustomerName:     req.CustomerName,
		PhoneNumber:      req.PhoneNumber,
		Email:            req.Email,
		CarPlate:         req.CarPlate,
		Description:      req.Description,
		Status:           "unacknowledged",
		CreatedAt:        time.Now().Format(time.RFC3339),
		TaggedWarrantyID: "",
	}
	mockClaims[claimID] = claim
	c.JSON(http.StatusCreated, claim)
}

// GET /claims
func GetClaims(c *gin.Context) {
	carPlate := c.Query("car_plate")
	status := c.Query("status")
	var claims []models.Claim
	for _, cl := range mockClaims {
		if carPlate != "" && cl.CarPlate != carPlate {
			continue
		}
		if status != "" && cl.Status != status {
			continue
		}
		claims = append(claims, cl)
	}
	c.JSON(http.StatusOK, claims)
}

// PATCH /claims/:id/tag_warranty
func TagWarrantyToClaim(c *gin.Context) {
	type TagWarrantyRequest struct {
		WarrantyID string `json:"warranty_id" binding:"required"`
	}
	var req TagWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claimID := c.Param("id")
	claim, exists := mockClaims[claimID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}
	if claim.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only tag warranty in pending status"})
		return
	}
	warranty, wExists := mockWarranties[req.WarrantyID]
	if !wExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warranty not found"})
		return
	}
	if warranty.TaggedClaimID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Warranty already tagged to another claim"})
		return
	}
	claim.TaggedWarrantyID = req.WarrantyID
	warranty.TaggedClaimID = claimID
	mockClaims[claimID] = claim
	mockWarranties[req.WarrantyID] = warranty
	c.JSON(http.StatusOK, claim)
}

// PATCH /claims/:id/status
func ChangeClaimStatus(c *gin.Context) {
	type ChangeClaimStatusRequest struct {
		Status          string `json:"status" binding:"required"`
		RejectionReason string `json:"rejection_reason,omitempty"`
	}
	var req ChangeClaimStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claimID := c.Param("id")
	claim, exists := mockClaims[claimID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}
	validTransitions := map[string][]string{
		"unacknowledged": {"pending"},
		"pending":        {"approved", "rejected"},
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
	if req.Status == "approved" && claim.TaggedWarrantyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot approve claim without tagged warranty"})
		return
	}
	if req.Status == "rejected" && req.RejectionReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason required"})
		return
	}
	claim.Status = req.Status
	claim.DateSettled = time.Now().Format(time.RFC3339)
	if req.Status == "rejected" {
		claim.RejectionReason = req.RejectionReason
	} else {
		claim.RejectionReason = ""
	}
	mockClaims[claimID] = claim
	if (req.Status == "approved" || req.Status == "rejected") && claim.TaggedWarrantyID != "" {
		warranty, wExists := mockWarranties[claim.TaggedWarrantyID]
		if wExists {
			warranty.Status = "used"
			mockWarranties[claim.TaggedWarrantyID] = warranty
		}
	}
	c.JSON(http.StatusOK, claim)
}

// PATCH /claims/:id/close
func CloseClaim(c *gin.Context) {
	claimID := c.Param("id")
	claim, exists := mockClaims[claimID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Claim not found"})
		return
	}
	if claim.Status != "approved" && claim.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only close approved or rejected claims"})
		return
	}
	if claim.DateClosed != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Claim already closed"})
		return
	}
	claim.DateClosed = time.Now().Format(time.RFC3339)
	mockClaims[claimID] = claim
	c.JSON(http.StatusOK, claim)
}
