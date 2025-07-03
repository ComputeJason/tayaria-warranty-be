package models

type Shop struct {
	ID           string `json:"id"`
	ShopName     string `json:"shop_name"`
	Address      string `json:"address"`
	Contact      string `json:"contact"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type CreateShopRequest struct {
	ShopName     string `json:"shop_name" binding:"required"`
	Address      string `json:"address" binding:"required"`
	Contact      string `json:"contact" binding:"required"`
	Username     string `json:"username" binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required"`
}

type UpdateShopRequest struct {
	ShopName     string `json:"shop_name,omitempty"`
	Address      string `json:"address,omitempty"`
	Contact      string `json:"contact,omitempty"`
	Username     string `json:"username,omitempty"`
	PasswordHash string `json:"password_hash,omitempty"`
}

// ShopWithAdmin combines shop information with its admin details
type ShopWithAdmin struct {
	Shop  Shop  `json:"shop"`
	Admin Admin `json:"admin"`
}
