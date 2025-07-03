package models

type Warranty struct {
	ID             string `json:"id"`
	CustomerName   string `json:"customer_name"`
	CarPlate       string `json:"car_plate"`
	PhoneNumber    string `json:"phone_number"`
	Email          string `json:"email"`
	PurchaseDate   string `json:"purchase_date"`
	ReceiptURL     string `json:"receipt_url"`
	Status         string `json:"status"`
	CreatedDate    string `json:"created_date"`
	ExpirationDate string `json:"expiration_date"`
	TaggedClaimID  string `json:"tagged_claim_id"`
}

type CreateWarrantyRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
	CarPlate     string `json:"car_plate" binding:"required"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	PurchaseDate string `json:"purchase_date" binding:"required"`
	ReceiptURL   string `json:"receipt_url" binding:"required"`
}
