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
	ID              string      `json:"id"`
	WarrantyID      string      `json:"warranty_id"`
	ShopID          string      `json:"shop_id"`
	Status          ClaimStatus `json:"status"`
	RejectionReason string      `json:"rejectionReason"`
	DateSettled     *time.Time  `json:"dateSettled"`
	DateClosed      *time.Time  `json:"dateClosed"`
	// Customer info
	CustomerName string    `json:"customerName"`
	PhoneNumber  string    `json:"phoneNumber"`
	Email        string    `json:"email"`
	CarPlate     string    `json:"carPlate"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateClaimRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
	Email        string `json:"email"`
	CarPlate     string `json:"car_plate" binding:"required"`
	ShopID       string `json:"shop_id" binding:"required"`
}

type UpdateClaimStatusRequest struct {
	Status          ClaimStatus `json:"status" binding:"required"`
	RejectionReason string      `json:"rejection_reason,omitempty"`
}

type TagWarrantyRequest struct {
	WarrantyID string `json:"warranty_id" binding:"required"`
}
