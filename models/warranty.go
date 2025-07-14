package models

import "time"

type Warranty struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	PurchaseDate time.Time `json:"purchase_date"`
	ExpiryDate   time.Time `json:"expiry_date"`
	CarPlate     string    `json:"car_plate"`
	Receipt      string    `json:"receipt"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsUsed       bool      `json:"is_used"`
}

type CreateWarrantyRequest struct {
	Name         string    `json:"name" binding:"required"`
	PhoneNumber  string    `json:"phone_number" binding:"required"`
	Email        string    `json:"email"`
	PurchaseDate time.Time `json:"purchase_date" binding:"required"`
	CarPlate     string    `json:"car_plate" binding:"required"`
	Receipt      string    `json:"receipt" binding:"required"`
}
