package models

type UserRole string

const (
	AdminRole     UserRole = "admin"
	SuperUserRole UserRole = "super_user"
)

type Admin struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token string `json:"token"`
	Admin Admin  `json:"admin"`
}
