package models

import "time"

type UserRole string

const (
	AdminRole  UserRole = "admin"
	MasterRole UserRole = "master"
)

type Shop struct {
	ID        string    `json:"id" db:"id"`
	ShopName  string    `json:"shop_name" db:"shop_name"`
	Address   string    `json:"address" db:"address"`
	Contact   string    `json:"contact" db:"contact"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateRetailAccountRequest struct {
	ShopName string `json:"shop_name" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Contact  string `json:"contact"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateRetailAccountResponse struct {
	ID        string    `json:"id"`
	ShopName  string    `json:"shop_name"`
	Address   string    `json:"address"`
	Contact   string    `json:"contact"`
	Username  string    `json:"username"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ShopLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ShopLoginResponse struct {
	Token string `json:"token"`
	Shop  Shop   `json:"shop"`
}

// TODO: Add middleware authentication for master routes
// TODO: Implement JWT token generation and validation
// TODO: Add proper session management
