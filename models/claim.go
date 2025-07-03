package models

type ClaimStatus string

const (
	PendingStatus  ClaimStatus = "pending"
	ApprovedStatus ClaimStatus = "approved"
	RejectedStatus ClaimStatus = "rejected"
)

type Claim struct {
	ID               string `json:"id"`
	CustomerName     string `json:"customer_name"`
	PhoneNumber      string `json:"phone_number"`
	Email            string `json:"email"`
	CarPlate         string `json:"car_plate"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	TaggedWarrantyID string `json:"tagged_warranty_id"`
	DateSettled      string `json:"date_settled"`
	RejectionReason  string `json:"rejection_reason"`
	DateClosed       string `json:"date_closed"`
}

type CreateClaimRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	CarPlate     string `json:"car_plate" binding:"required"`
	Description  string `json:"description" binding:"required"`
}

type UpdateClaimStatusRequest struct {
	Status       ClaimStatus `json:"status" binding:"required"`
	AdminComment string      `json:"adminComment,omitempty"`
}
