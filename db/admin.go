package db

import (
	"context"
	"log"
)

type AdminDB struct {
	ID       string
	Username string
	Password string
	Role     string
	ShopID   string
	ShopName string
}

type SuperUserDB struct {
	ID       string
	Username string
	Password string
	Role     string
}

func GetAdminByUsername(username string) (*AdminDB, error) {
	ctx := context.Background()

	query := `
		SELECT a.id, a.username, a.password, a.role, a.shop_id, s.shop_name
		FROM admins a
		LEFT JOIN shops s ON a.shop_id = s.id
		WHERE a.username = $1
	`

	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var admin AdminDB
	err = conn.Conn().QueryRow(ctx, query, username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Password,
		&admin.Role,
		&admin.ShopID,
		&admin.ShopName,
	)
	if err != nil {
		log.Printf("Error querying admin: %v", err)
		return nil, err
	}

	return &admin, nil
}

func GetSuperUserByUsername(username string) (*SuperUserDB, error) {
	ctx := context.Background()

	query := `
		SELECT a.id, a.username, a.password, a.role
		FROM admins a
		WHERE a.username = $1
	`

	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var superuser SuperUserDB
	err = conn.Conn().QueryRow(ctx, query, username).Scan(
		&superuser.ID,
		&superuser.Username,
		&superuser.Password,
		&superuser.Role,
	)
	if err != nil {
		log.Printf("Error querying admin: %v", err)
		return nil, err
	}

	return &superuser, nil
}
