package models

type User struct {
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
}
