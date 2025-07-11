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
func CreateClaim(claim models.CreateClaimRequest) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Generate UUID for claim ID
	claimID := uuid.New().String()

	query := `
		INSERT INTO claims (id, warranty_id, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, warranty_id, description, status, admin_comment, created_at, updated_at
	`

	log.Printf("Executing SQL query: %s with params: [%s, %s, %s, %s]",
		query, claimID, claim.WarrantyID, claim.Description, "pending")

	row := db.QueryRow(context.Background(), query,
		claimID,
		claim.WarrantyID,
		claim.Description,
		"pending", // Default status
	)

	var result models.Claim
	var createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&result.ID,
		&result.WarrantyID,
		&result.Description,
		&result.Status,
		&result.AdminComment,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create claim: %v", err)
	}

	// Convert pgtype.Timestamp to time.Time
	if createdAt.Valid {
		result.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		result.UpdatedAt = updatedAt.Time
	}

	return &result, nil
}

// GetClaims retrieves claims with optional filtering
func GetClaims(warrantyID string, status string) ([]models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var query string
	var args []interface{}
	argCount := 0

	query = `
		SELECT id, warranty_id, description, status, admin_comment, created_at, updated_at
		FROM claims
		WHERE 1=1
	`

	if warrantyID != "" {
		argCount++
		query += fmt.Sprintf(" AND warranty_id = $%d", argCount)
		args = append(args, warrantyID)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	log.Printf("Executing SQL query: %s with params: %v", query, args)

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %v", err)
	}
	defer rows.Close()

	var claims []models.Claim
	for rows.Next() {
		var claim models.Claim
		var createdAt, updatedAt pgtype.Timestamp

		err := rows.Scan(
			&claim.ID,
			&claim.WarrantyID,
			&claim.Description,
			&claim.Status,
			&claim.AdminComment,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %v", err)
		}

		// Convert pgtype.Timestamp to time.Time
		if createdAt.Valid {
			claim.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			claim.UpdatedAt = updatedAt.Time
		}

		claims = append(claims, claim)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating claims: %v", err)
	}

	return claims, nil
}

// GetClaimByID retrieves a claim by its ID
func GetClaimByID(claimID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, warranty_id, description, status, admin_comment, created_at, updated_at
		FROM claims
		WHERE id = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, claimID)

	row := db.QueryRow(context.Background(), query, claimID)

	var claim models.Claim
	var createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&claim.ID,
		&claim.WarrantyID,
		&claim.Description,
		&claim.Status,
		&claim.AdminComment,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Claim not found
		}
		return nil, fmt.Errorf("failed to get claim: %v", err)
	}

	// Convert pgtype.Timestamp to time.Time
	if createdAt.Valid {
		claim.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		claim.UpdatedAt = updatedAt.Time
	}

	return &claim, nil
}

// UpdateClaimStatus updates the status of a claim
func UpdateClaimStatus(claimID string, status models.ClaimStatus, adminComment string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE claims 
		SET status = $2, admin_comment = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, warranty_id, description, status, admin_comment, created_at, updated_at
	`

	log.Printf("Executing SQL query: %s with params: [%s, %s, %s]", query, claimID, status, adminComment)

	row := db.QueryRow(context.Background(), query, claimID, status, adminComment)

	var claim models.Claim
	var createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(
		&claim.ID,
		&claim.WarrantyID,
		&claim.Description,
		&claim.Status,
		&claim.AdminComment,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, fmt.Errorf("failed to update claim status: %v", err)
	}

	// Convert pgtype.Timestamp to time.Time
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
