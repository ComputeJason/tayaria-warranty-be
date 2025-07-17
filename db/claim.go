package db

import (
	"context"
	"fmt"
	"log"
	"time"

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

	// First verify if the shop exists
	var exists bool
	err := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM shops WHERE id = $1)", shopID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check shop existence: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("shop with ID %s does not exist", shopID)
	}

	// Parse shopID into UUID
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop ID format: %v (shop_id: %s)", err, shopID)
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

	log.Printf("Creating claim with params: claimID=%s, warrantyID=%s, shopID=%s", claimID, warranty.ID, shopUUID)

	row := db.QueryRow(context.Background(), query,
		claimID,
		nil,
		shopUUID,
		"unacknowledged", // Default status
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
		SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason, 
		       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate, 
		       c.created_at, c.updated_at
		FROM claims c
		LEFT JOIN shops s ON c.shop_id = s.id
		WHERE c.shop_id = $1
		ORDER BY c.created_at DESC
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
		var shopName, contact pgtype.Text

		err := rows.Scan(
			&claim.ID,
			&claim.WarrantyID,
			&claim.ShopID,
			&shopName,
			&contact,
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
		if shopName.Valid {
			claim.ShopName = shopName.String
		}
		if contact.Valid {
			claim.Contact = contact.String
		}
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
		SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason, 
		       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate, 
		       c.created_at, c.updated_at
		FROM claims c
		LEFT JOIN shops s ON c.shop_id = s.id
		WHERE c.id = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, claimID)

	row := db.QueryRow(context.Background(), query, claimID)

	var claim models.Claim
	var rejectionReason pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp
	var shopName, contact pgtype.Text

	err := row.Scan(
		&claim.ID,
		&claim.WarrantyID,
		&claim.ShopID,
		&shopName,
		&contact,
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

	// Check if shop exists (should not be null for valid claims)
	if !shopName.Valid {
		return nil, fmt.Errorf("shop not found for claim %s", claimID)
	}

	// Convert pgtype values to Go types
	claim.ShopName = shopName.String
	if contact.Valid {
		claim.Contact = contact.String
	}
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

	// Get shop information
	shopQuery := `SELECT shop_name, contact FROM shops WHERE id = $1`
	var shopName, contact pgtype.Text
	err = db.QueryRow(context.Background(), shopQuery, claim.ShopID).Scan(&shopName, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop information: %v", err)
	}

	if shopName.Valid {
		claim.ShopName = shopName.String
	}
	if contact.Valid {
		claim.Contact = contact.String
	}

	return &claim, nil
}

// UpdateClaimWarrantyID updates the warranty_id of a claim
func UpdateClaimWarrantyID(claimID string, warrantyID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	tx, err := db.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	// 1. Update the claim's warranty_id
	claimUpdateQuery := `
		UPDATE claims 
		SET warranty_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, warranty_id, shop_id, status, rejection_reason, date_settled, date_closed,
		          customer_name, phone_number, email, car_plate, created_at, updated_at
	`
	row := tx.QueryRow(context.Background(), claimUpdateQuery, claimID, warrantyID)

	var claim models.Claim
	var rejectionReason pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

	err = row.Scan(
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
			tx.Rollback(context.Background())
			return nil, fmt.Errorf("claim not found")
		}
		tx.Rollback(context.Background())
		return nil, fmt.Errorf("failed to update claim warranty: %v", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
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

	// Get shop information
	shopQuery := `SELECT shop_name, contact FROM shops WHERE id = $1`
	var shopName, contact pgtype.Text
	err = db.QueryRow(context.Background(), shopQuery, claim.ShopID).Scan(&shopName, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop information: %v", err)
	}

	if shopName.Valid {
		claim.ShopName = shopName.String
	}
	if contact.Valid {
		claim.Contact = contact.String
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

// GetClaimsByStatus retrieves claims based on status type
func GetClaimsByStatus(statusType string) ([]models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var query string
	var args []interface{}

	switch statusType {
	case "unacknowledged":
		query = `
			SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason, 
			       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate, 
			       c.created_at, c.updated_at
			FROM claims c
			LEFT JOIN shops s ON c.shop_id = s.id
			WHERE c.status = 'unacknowledged'
			ORDER BY c.created_at DESC
		`
	case "pending":
		query = `
			SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason, 
			       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate, 
			       c.created_at, c.updated_at
			FROM claims c
			LEFT JOIN shops s ON c.shop_id = s.id
			WHERE c.status = 'pending'
			ORDER BY c.created_at DESC
		`
	case "history":
		query = `
			SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason, 
			       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate, 
			       c.created_at, c.updated_at
			FROM claims c
			LEFT JOIN shops s ON c.shop_id = s.id
			WHERE c.status IN ('approved', 'rejected')
			ORDER BY c.created_at DESC
		`
	default:
		return nil, fmt.Errorf("invalid status type: %s", statusType)
	}

	log.Printf("Executing SQL query: %s", query)

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %v", err)
	}
	defer rows.Close()

	claims := []models.Claim{} // Initialize empty slice
	for rows.Next() {
		var claim models.Claim
		var rejectionReason pgtype.Text
		var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp
		var shopName, contact pgtype.Text

		err := rows.Scan(
			&claim.ID,
			&claim.WarrantyID,
			&claim.ShopID,
			&shopName,
			&contact,
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
			return nil, fmt.Errorf("failed to scan claim: %v", err)
		}

		// Check if shop exists (should not be null for valid claims)
		if !shopName.Valid {
			return nil, fmt.Errorf("shop not found for claim %s", claim.ID)
		}

		// Convert pgtype values to Go types
		claim.ShopName = shopName.String
		if contact.Valid {
			claim.Contact = contact.String
		}
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
		return nil, fmt.Errorf("error iterating claims: %v", err)
	}

	return claims, nil
}

// AcceptClaim changes claim status to approved and adds tyre details
func AcceptClaim(claimID string, tyreDetails []models.TyreDetail) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Start transaction
	tx, err := db.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	// Update claim status to approved
	updateClaimQuery := `
		UPDATE claims 
		SET status = $2, 
		    updated_at = CURRENT_TIMESTAMP,
		    date_settled = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'pending'
		RETURNING id`

	var claimIDResult string
	err = tx.QueryRow(context.Background(), updateClaimQuery, claimID, models.ApprovedStatus).Scan(&claimIDResult)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("claim not found or not in pending status")
		}
		return nil, fmt.Errorf("failed to update claim status: %v", err)
	}

	// Insert tyre details
	insertTyreQuery := `
		INSERT INTO tyre_details (claim_id, brand, size, tread_pattern)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	for i := range tyreDetails {
		var tyreID string
		var createdAt time.Time
		err = tx.QueryRow(context.Background(), insertTyreQuery,
			claimID,
			tyreDetails[i].Brand,
			tyreDetails[i].Size,
			tyreDetails[i].TreadPattern,
		).Scan(&tyreID, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to insert tyre detail: %v", err)
		}
		tyreDetails[i].ID = tyreID
		tyreDetails[i].ClaimID = claimID
		tyreDetails[i].CreatedAt = createdAt
	}

	// Commit transaction
	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Get updated claim with tyre details
	return GetClaimWithTyreDetails(claimID)
}

// GetClaimWithTyreDetails retrieves a claim with its tyre details
func GetClaimWithTyreDetails(claimID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Get claim
	claim, err := GetClaimByID(claimID)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, nil
	}

	// Get tyre details
	query := `
		SELECT id, claim_id, brand, size, tread_pattern, created_at
		FROM tyre_details
		WHERE claim_id = $1
		ORDER BY created_at ASC`

	rows, err := db.Query(context.Background(), query, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tyre details: %v", err)
	}
	defer rows.Close()

	var tyreDetails []models.TyreDetail
	for rows.Next() {
		var td models.TyreDetail
		err := rows.Scan(
			&td.ID,
			&td.ClaimID,
			&td.Brand,
			&td.Size,
			&td.TreadPattern,
			&td.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tyre detail: %v", err)
		}
		tyreDetails = append(tyreDetails, td)
	}

	claim.TyreDetails = tyreDetails
	return claim, nil
}

// RejectClaim changes claim status to rejected with a reason
func RejectClaim(claimID string, reason string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE claims 
		SET status = $2, 
		    rejection_reason = $3,
		    updated_at = CURRENT_TIMESTAMP,
		    date_settled = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'pending'
		RETURNING id, warranty_id, shop_id, status, rejection_reason, 
		          date_settled, date_closed,
		          customer_name, phone_number, email, car_plate,
		          created_at, updated_at`

	var claim models.Claim
	var warrantyID, rejectionReason pgtype.Text
	var dateSettled, dateClosed pgtype.Timestamp

	err := db.QueryRow(context.Background(), query, claimID, models.RejectedStatus, reason).Scan(
		&claim.ID,
		&warrantyID,
		&claim.ShopID,
		&claim.Status,
		&rejectionReason,
		&dateSettled,
		&dateClosed,
		&claim.CustomerName,
		&claim.PhoneNumber,
		&claim.Email,
		&claim.CarPlate,
		&claim.CreatedAt,
		&claim.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("claim not found or not in pending status")
		}
		return nil, fmt.Errorf("failed to update claim: %v", err)
	}

	if warrantyID.Valid {
		claim.WarrantyID = &warrantyID.String
	}
	if rejectionReason.Valid {
		claim.RejectionReason = rejectionReason.String
	}
	if dateSettled.Valid {
		claim.DateSettled = &dateSettled.Time
	}
	if dateClosed.Valid {
		claim.DateClosed = &dateClosed.Time
	}

	return &claim, nil
}

// CloseClaim updates the date_closed field of a claim
func CloseClaim(claimID string) (*models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE claims 
		SET date_closed = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('approved', 'rejected')
		RETURNING id, warranty_id, shop_id, status, rejection_reason, 
		          date_settled, date_closed,
		          customer_name, phone_number, email, car_plate,
		          created_at, updated_at`

	log.Printf("Executing SQL query: %s with params: [%s]", query, claimID)

	var claim models.Claim
	var warrantyID, rejectionReason pgtype.Text
	var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp

	err := db.QueryRow(context.Background(), query, claimID).Scan(
		&claim.ID,
		&warrantyID,
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
			return nil, fmt.Errorf("claim not found or not in approved/rejected status")
		}
		return nil, fmt.Errorf("failed to close claim: %v", err)
	}

	// Convert pgtype values to Go types
	if warrantyID.Valid {
		claim.WarrantyID = &warrantyID.String
	}
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
