package db

import (
	"context"
	"log"
	"tayaria-warranty-be/models"

	"github.com/jackc/pgx/v5"
)

func GetShopByUsername(username string) (*models.Shop, error) {
	ctx := context.Background()

	query := `
		SELECT id, shop_name, address, contact, username, password, role, created_at, updated_at
		FROM shops
		WHERE username = $1
	`

	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var shop models.Shop
	err = conn.Conn().QueryRow(ctx, query, username).Scan(
		&shop.ID,
		&shop.ShopName,
		&shop.Address,
		&shop.Contact,
		&shop.Username,
		&shop.Password,
		&shop.Role,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Return nil for not found instead of error
		}
		log.Printf("Error querying shop: %v", err)
		return nil, err
	}

	return &shop, nil
}

func CreateShop(req *models.CreateRetailAccountRequest) (*models.Shop, error) {
	ctx := context.Background()

	query := `
		INSERT INTO shops (shop_name, address, contact, username, password, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, shop_name, address, contact, username, password, role, created_at, updated_at
	`

	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var shop models.Shop
	err = conn.Conn().QueryRow(ctx, query,
		req.ShopName,
		req.Address,
		req.Contact,
		req.Username,
		req.Password,
		models.AdminRole, // Default role for retail accounts
	).Scan(
		&shop.ID,
		&shop.ShopName,
		&shop.Address,
		&shop.Contact,
		&shop.Username,
		&shop.Password,
		&shop.Role,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating shop: %v", err)
		return nil, err
	}

	return &shop, nil
}

func GetAllShops() ([]models.Shop, error) {
	ctx := context.Background()

	query := `
		SELECT id, shop_name, address, contact, username, password, role, created_at, updated_at
		FROM shops
		WHERE role = $1
		ORDER BY created_at DESC
	`

	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Conn().Query(ctx, query, models.AdminRole)
	if err != nil {
		log.Printf("Error querying shops: %v", err)
		return nil, err
	}
	defer rows.Close()

	var shops []models.Shop
	for rows.Next() {
		var shop models.Shop
		err = rows.Scan(
			&shop.ID,
			&shop.ShopName,
			&shop.Address,
			&shop.Contact,
			&shop.Username,
			&shop.Password,
			&shop.Role,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning shop: %v", err)
			continue
		}
		shops = append(shops, shop)
	}

	return shops, nil
}
