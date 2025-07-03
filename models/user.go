package models

type User struct {
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
