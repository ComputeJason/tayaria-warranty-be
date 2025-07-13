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

// CreateClaim creates a new claim in the database
func CreateClaim(claim models.CreateClaimRequest, shopID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// First, validate that the car plate has a valid warranty
	warranty, err := GetValidWarrantyByCarPlate(claim.CarPlate)
	if err != nil {
		return nil, fmt.Errorf("failed to check warranty: %v", err)
	}
	if warranty == nil {
		return nil, fmt.Errorf("no valid warranty found for car plate %s", claim.CarPlate)
	}

	// Generate UUID for claim ID
	claimID := uuid.New().String()

	query := `
		INSERT INTO claims (id, warranty_id, shop_id, status, customer_name, phone_number, email, car_plate)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, warranty_id, shop_id, status, rejection_reason, date_settled, date_closed, 
		          customer_name, phone_number, email, car_plate, created_at, updated_at
	`

	log.Printf("Executing SQL query: %s with params: [%s, %s, %s, %s, %s, %s, %s, %s]",
		query, claimID, warranty.ID, shopID, "pending", claim.CustomerName, claim.PhoneNumber, claim.Email, claim.CarPlate)

	row := db.QueryRow(context.Background(), query,
		claimID,
		warranty.ID,
		shopID,
		"pending", // Default status
		claim.CustomerName,
		claim.PhoneNumber,
		claim.Email,
		claim.CarPlate,
	)

	var result models.Claim
	var rejectionReason pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

	err = row.Scan(
		&result.ID,
		&result.WarrantyID,
		&result.ShopID,
		&result.Status,
		&rejectionReason,
		&dateSettled,
		&dateClosed,
		&result.CustomerName,
		&result.PhoneNumber,
		&result.Email,
		&result.CarPlate,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create claim: %v", err)
	}

	// Convert pgtype values to Go types
	if rejectionReason.Valid {
		result.RejectionReason = rejectionReason.String
	}
	if dateSettled.Valid {
		result.DateSettled = &dateSettled.Time
	}
	if dateClosed.Valid {
		result.DateClosed = &dateClosed.Time
	}
	if createdAt.Valid {
		result.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		result.UpdatedAt = updatedAt.Time
	}

	return &result, nil
}

// GetShopClaims retrieves claims for a specific shop
func GetShopClaims(shopID string) ([]models.Claim, error) {
	if db == nil {
		return []models.Claim{}, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, warranty_id, shop_id, status, rejection_reason, date_settled, date_closed,
		       customer_name, phone_number, email, car_plate, created_at, updated_at
		FROM claims
		WHERE shop_id = $1
		ORDER BY created_at DESC
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, shopID)

	rows, err := db.Query(context.Background(), query, shopID)
	if err != nil {
		return []models.Claim{}, fmt.Errorf("failed to query claims: %v", err)
	}
	defer rows.Close()

	claims := []models.Claim{} // Initialize empty slice
	for rows.Next() {
		var claim models.Claim
		var rejectionReason pgtype.Text
		var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

		err := rows.Scan(
			&claim.ID,
			&claim.WarrantyID,
			&claim.ShopID,
			&claim.Status,
			&rejectionReason,
			&dateSettled,
			&dateClosed,
			&claim.CustomerName,
			&claim.PhoneNumber,
			&claim.Email,
			&claim.CarPlate,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return []models.Claim{}, fmt.Errorf("failed to scan claim: %v", err)
		}

		// Convert pgtype values to Go types
		if rejectionReason.Valid {
			claim.RejectionReason = rejectionReason.String
		}
		if dateSettled.Valid {
			claim.DateSettled = &dateSettled.Time
		}
		if dateClosed.Valid {
			claim.DateClosed = &dateClosed.Time
		}
		if createdAt.Valid {
			claim.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			claim.UpdatedAt = updatedAt.Time
		}

		claims = append(claims, claim)
	}

	if err = rows.Err(); err != nil {
		return []models.Claim{}, fmt.Errorf("error iterating claims: %v", err)
	}

	return claims, nil
}

// GetClaimByID retrieves a claim by its ID
func GetClaimByID(claimID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, warranty_id, shop_id, status, rejection_reason, date_settled, date_closed,
		       customer_name, phone_number, email, car_plate, created_at, updated_at
		FROM claims
		WHERE id = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, claimID)

	row := db.QueryRow(context.Background(), query, claimID)

	var claim models.Claim
	var rejectionReason pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&claim.ID,
		&claim.WarrantyID,
		&claim.ShopID,
		&claim.Status,
		&rejectionReason,
		&dateSettled,
		&dateClosed,
		&claim.CustomerName,
		&claim.PhoneNumber,
		&claim.Email,
		&claim.CarPlate,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Claim not found
		}
		return nil, fmt.Errorf("failed to get claim: %v", err)
	}

	// Convert pgtype values to Go types
	if rejectionReason.Valid {
		claim.RejectionReason = rejectionReason.String
	}
	if dateSettled.Valid {
		claim.DateSettled = &dateSettled.Time
	}
	if dateClosed.Valid {
		claim.DateClosed = &dateClosed.Time
	}
	if createdAt.Valid {
		claim.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		claim.UpdatedAt = updatedAt.Time
	}

	return &claim, nil
}

// UpdateClaimStatus updates the status of a claim
func UpdateClaimStatus(claimID string, status models.ClaimStatus, rejectionReason string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE claims 
		SET status = $2, rejection_reason = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, warranty_id, shop_id, status, rejection_reason, date_settled, date_closed,
		          customer_name, phone_number, email, car_plate, created_at, updated_at
	`

	log.Printf("Executing SQL query: %s with params: [%s, %s, %s]", query, claimID, status, rejectionReason)

	row := db.QueryRow(context.Background(), query, claimID, status, rejectionReason)

	var claim models.Claim
	var rejectionReasonDB pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&claim.ID,
		&claim.WarrantyID,
		&claim.ShopID,
		&claim.Status,
		&rejectionReasonDB,
		&dateSettled,
		&dateClosed,
		&claim.CustomerName,
		&claim.PhoneNumber,
		&claim.Email,
		&claim.CarPlate,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, fmt.Errorf("failed to update claim status: %v", err)
	}

	// Convert pgtype values to Go types
	if rejectionReasonDB.Valid {
		claim.RejectionReason = rejectionReasonDB.String
	}
	if dateSettled.Valid {
		claim.DateSettled = &dateSettled.Time
	}
	if dateClosed.Valid {
		claim.DateClosed = &dateClosed.Time
	}
	if createdAt.Valid {
		claim.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		claim.UpdatedAt = updatedAt.Time
	}

	return &claim, nil
}

// CheckWarrantyExists checks if a warranty exists
func CheckWarrantyExists(warrantyID string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT EXISTS(SELECT 1 FROM warranties WHERE id = $1)`

	var exists bool
	err := db.QueryRow(context.Background(), query, warrantyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check warranty existence: %v", err)
	}

	return exists, nil
}
