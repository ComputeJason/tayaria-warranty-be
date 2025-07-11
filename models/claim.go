package models

import "time"

type ClaimStatus string

const (
	UnacknowledgedStatus ClaimStatus = "unacknowledged"
	PendingStatus        ClaimStatus = "pending"
	ApprovedStatus       ClaimStatus = "approved"
	RejectedStatus       ClaimStatus = "rejected"
)

type Claim struct {
	ID           string      `json:"id"`
	WarrantyID   string      `json:"warranty_id"`
	Description  string      `json:"description"`
	Status       ClaimStatus `json:"status"`
	AdminComment string      `json:"admin_comment"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type CreateClaimRequest struct {
	WarrantyID  string `json:"warranty_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateClaimStatusRequest struct {
	Status       ClaimStatus `json:"status" binding:"required"`
	AdminComment string      `json:"admin_comment,omitempty"`
}

type TagWarrantyRequest struct {
	WarrantyID string `json:"warranty_id" binding:"required"`
}
