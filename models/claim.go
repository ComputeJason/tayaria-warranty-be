package models

import "time"

type ClaimStatus string

const (
	UnacknowledgedStatus ClaimStatus = "unacknowledged"
	PendingStatus        ClaimStatus = "pending"
	ApprovedStatus       ClaimStatus = "approved"
	RejectedStatus       ClaimStatus = "rejected"
)

type TyreDetail struct {
	ID        string    `json:"id"`
	ClaimID   string    `json:"claim_id"`
	Brand     string    `json:"brand"`
	Size      string    `json:"size"`
	Cost      float64   `json:"cost"`
	CreatedAt time.Time `json:"created_at"`
}

type Claim struct {
	ID              string      `json:"id"`
	WarrantyID      *string     `json:"warranty_id"`
	ShopID          string      `json:"shop_id"`
	Status          ClaimStatus `json:"status"`
	RejectionReason string      `json:"rejection_reason"`
	DateSettled     *time.Time  `json:"date_settled"`
	DateClosed      *time.Time  `json:"date_closed"`
	TotalCost       float64     `json:"total_cost"`
	// Customer info
	CustomerName string    `json:"customer_name"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	CarPlate     string    `json:"car_plate"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// Optional field for when we need to include tyre details
	TyreDetails []TyreDetail `json:"tyre_details,omitempty"`
}

// Request models
type CreateClaimRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
	Email        string `json:"email"`
	CarPlate     string `json:"car_plate" binding:"required"`
}

type TagWarrantyRequest struct {
	WarrantyID string `json:"warranty_id" binding:"required"`
}

type UpdateClaimStatusRequest struct {
	Status          ClaimStatus `json:"status" binding:"required"`
	RejectionReason string      `json:"rejectionReason"`
}

type AcceptClaimRequest struct {
	TyreDetails []struct {
		Brand string  `json:"brand" binding:"required"`
		Size  string  `json:"size" binding:"required"`
		Cost  float64 `json:"cost" binding:"required,min=0"`
	} `json:"tyre_details" binding:"required,min=1,max=4"`
}

type RejectClaimRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}
