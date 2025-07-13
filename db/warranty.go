package db

import (
	"context"
	"fmt"
	"log"

	"tayaria-warranty-be/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateWarranty creates a new warranty in the database
func CreateWarranty(warranty models.CreateWarrantyRequest) (*models.Warranty, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Calculate expiry date (6 months from purchase date)
	expiryDate := warranty.PurchaseDate.AddDate(0, 6, 0)

	// Generate UUID for warranty ID
	warrantyID := uuid.New().String()

	// TODO: Implement file upload to S3 or Supabase storage
	// For now, use a placeholder URL
	receiptURL := warranty.Receipt
	if receiptURL == "" {
		receiptURL = "https://placeholder.com/receipt.pdf"
	}

	query := `
		INSERT INTO warranties (id, name, phone_number, email, purchase_date, expiry_date, car_plate, receipt)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, phone_number, email, purchase_date, expiry_date, car_plate, receipt, created_at, updated_at
	`

	log.Printf("Executing SQL query: %s with params: [%s, %s, %s, %s, %s, %s, %s, %s]",
		query, warrantyID, warranty.Name, warranty.PhoneNumber, warranty.Email,
		warranty.PurchaseDate.Format("2006-01-02"), expiryDate.Format("2006-01-02"), warranty.CarPlate, receiptURL)

	row := db.QueryRow(context.Background(), query,
		warrantyID,
		warranty.Name,
		warranty.PhoneNumber,
		warranty.Email,
		warranty.PurchaseDate,
		expiryDate,
		warranty.CarPlate,
		receiptURL,
	)

	var result models.Warranty
	var purchaseDate, expiryDateDB, createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.PhoneNumber,
		&result.Email,
		&purchaseDate,
		&expiryDateDB,
		&result.CarPlate,
		&result.Receipt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create warranty: %v", err)
	}

	// Convert pgtype.Timestamp to time.Time
	if purchaseDate.Valid {
		result.PurchaseDate = purchaseDate.Time
	}
	if expiryDateDB.Valid {
		result.ExpiryDate = expiryDateDB.Time
	}
	if createdAt.Valid {
		result.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		result.UpdatedAt = updatedAt.Time
	}

	return &result, nil
}

// GetWarrantiesByCarPlate retrieves all warranties for a given car plate
func GetWarrantiesByCarPlate(carPlate string) ([]models.Warranty, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, name, phone_number, email, purchase_date, expiry_date, car_plate, receipt, created_at, updated_at
		FROM warranties
		WHERE car_plate = $1
		ORDER BY created_at DESC
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, carPlate)

	rows, err := db.Query(context.Background(), query, carPlate)
	if err != nil {
		return nil, fmt.Errorf("failed to query warranties: %v", err)
	}
	defer rows.Close()

	var warranties []models.Warranty
	for rows.Next() {
		var warranty models.Warranty
		var purchaseDate, expiryDate, createdAt, updatedAt pgtype.Timestamp

		err := rows.Scan(
			&warranty.ID,
			&warranty.Name,
			&warranty.PhoneNumber,
			&warranty.Email,
			&purchaseDate,
			&expiryDate,
			&warranty.CarPlate,
			&warranty.Receipt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan warranty: %v", err)
		}

		// Convert pgtype.Timestamp to time.Time
		if purchaseDate.Valid {
			warranty.PurchaseDate = purchaseDate.Time
		}
		if expiryDate.Valid {
			warranty.ExpiryDate = expiryDate.Time
		}
		if createdAt.Valid {
			warranty.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			warranty.UpdatedAt = updatedAt.Time
		}

		warranties = append(warranties, warranty)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warranties: %v", err)
	}

	return warranties, nil
}

// GetValidWarrantyByCarPlate retrieves the active warranty for a given car plate
func GetValidWarrantyByCarPlate(carPlate string) (*models.Warranty, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT w.id, w.name, w.phone_number, w.email, w.purchase_date, w.expiry_date, w.car_plate, w.receipt, w.created_at, w.updated_at
		FROM warranties w
		LEFT JOIN claims c ON w.id = c.warranty_id
		WHERE w.car_plate = $1 
		AND w.expiry_date >= CURRENT_DATE
		AND c.warranty_id IS NULL  -- Only get warranties not tagged to any claim
		ORDER BY w.expiry_date DESC
		LIMIT 1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, carPlate)

	row := db.QueryRow(context.Background(), query, carPlate)

	var warranty models.Warranty
	var purchaseDate, expiryDate, createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&warranty.ID,
		&warranty.Name,
		&warranty.PhoneNumber,
		&warranty.Email,
		&purchaseDate,
		&expiryDate,
		&warranty.CarPlate,
		&warranty.Receipt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No valid warranty found
		}
		return nil, fmt.Errorf("failed to get valid warranty: %v", err)
	}

	// Convert pgtype.Timestamp to time.Time
	if purchaseDate.Valid {
		warranty.PurchaseDate = purchaseDate.Time
	}
	if expiryDate.Valid {
		warranty.ExpiryDate = expiryDate.Time
	}
	if createdAt.Valid {
		warranty.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		warranty.UpdatedAt = updatedAt.Time
	}

	return &warranty, nil
}

// GetAllValidWarrantiesForCarPlate retrieves all valid warranties for a car plate that can be tagged to a claim
func GetAllValidWarrantiesForCarPlate(carPlate string) ([]models.Warranty, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT w.id, w.name, w.phone_number, w.email, w.purchase_date, w.expiry_date, w.car_plate, w.receipt, w.created_at, w.updated_at
		FROM warranties w
		LEFT JOIN claims c ON w.id = c.warranty_id
		WHERE w.car_plate = $1 
		AND w.expiry_date >= CURRENT_DATE
		AND c.warranty_id IS NULL  -- Only get warranties not tagged to any claim
		ORDER BY w.expiry_date DESC
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, carPlate)

	rows, err := db.Query(context.Background(), query, carPlate)
	if err != nil {
		return nil, fmt.Errorf("failed to query warranties: %v", err)
	}
	defer rows.Close()

	var warranties []models.Warranty
	for rows.Next() {
		var warranty models.Warranty
		var purchaseDate, expiryDate, createdAt, updatedAt pgtype.Timestamp

		err := rows.Scan(
			&warranty.ID,
			&warranty.Name,
			&warranty.PhoneNumber,
			&warranty.Email,
			&purchaseDate,
			&expiryDate,
			&warranty.CarPlate,
			&warranty.Receipt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan warranty: %v", err)
		}

		// Convert pgtype.Timestamp to time.Time
		if purchaseDate.Valid {
			warranty.PurchaseDate = purchaseDate.Time
		}
		if expiryDate.Valid {
			warranty.ExpiryDate = expiryDate.Time
		}
		if createdAt.Valid {
			warranty.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			warranty.UpdatedAt = updatedAt.Time
		}

		warranties = append(warranties, warranty)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warranties: %v", err)
	}

	return warranties, nil
}

// GetWarrantyReceipt retrieves the receipt URL for a warranty
func GetWarrantyReceipt(warrantyID string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT receipt
		FROM warranties
		WHERE id = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, warrantyID)

	row := db.QueryRow(context.Background(), query, warrantyID)

	var receipt string
	err := row.Scan(&receipt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("warranty not found")
		}
		return "", fmt.Errorf("failed to get warranty receipt: %v", err)
	}

	return receipt, nil
}
