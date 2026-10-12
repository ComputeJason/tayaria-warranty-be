package db

import (
	"context"
	"fmt"

	"tayaria-warranty-be/models"

	"github.com/jackc/pgx/v5/pgtype"
)

// GetAllWarranties retrieves warranties from the database, optionally filtered to those
// purchased within [from, to] (inclusive, "YYYY-MM-DD"). An empty string leaves that side
// unbounded, so calling with both empty returns every warranty (the full export).
func GetAllWarranties(from, to string) ([]models.Warranty, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, name, phone_number, email, purchase_date, expiry_date, car_plate, brand, receipt, created_at, updated_at
		FROM warranties
		WHERE ($1 = '' OR purchase_date >= $1::date)
		AND ($2 = '' OR purchase_date <= $2::date)
		ORDER BY created_at DESC
	`

	rows, err := db.Query(context.Background(), query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query warranties: %v", err)
	}
	defer rows.Close()

	var warranties []models.Warranty
	for rows.Next() {
		var warranty models.Warranty
		var purchaseDate, expiryDate, createdAt, updatedAt pgtype.Timestamp
		var email pgtype.Text

		err := rows.Scan(
			&warranty.ID,
			&warranty.Name,
			&warranty.PhoneNumber,
			&email,
			&purchaseDate,
			&expiryDate,
			&warranty.CarPlate,
			&warranty.Brand,
			&warranty.Receipt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan warranty: %v", err)
		}

		// Convert pgtype values to Go types
		if email.Valid {
			warranty.Email = email.String
		}
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

// GetAllClaimsForExport retrieves claims from the database, optionally filtered to those
// created within [from, to] (inclusive, "YYYY-MM-DD"). An empty string leaves that side
// unbounded, so calling with both empty returns every claim (the full export). The "to" bound
// is widened to end-of-day since created_at carries a time component, not just a date.
func GetAllClaimsForExport(from, to string) ([]models.Claim, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT c.id, c.warranty_id, c.shop_id, s.shop_name, s.contact, c.status, c.rejection_reason,
		       c.date_settled, c.date_closed, c.customer_name, c.phone_number, c.email, c.car_plate,
		       c.supporting_doc_url, c.brand, c.quantity, c.created_at, c.updated_at
		FROM claims c
		LEFT JOIN shops s ON c.shop_id = s.id
		WHERE ($1 = '' OR c.created_at >= $1::date)
		AND ($2 = '' OR c.created_at < ($2::date + INTERVAL '1 day'))
		ORDER BY c.created_at DESC
	`

	rows, err := db.Query(context.Background(), query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query claims: %v", err)
	}
	defer rows.Close()

	var claims []models.Claim
	for rows.Next() {
		var claim models.Claim
		var rejectionReason, brand pgtype.Text
		var quantity pgtype.Int4
		var dateSettled, dateClosed, createdAt, updatedAt pgtype.Timestamp
		var shopName, contact pgtype.Text
		var supportingDocURL pgtype.Text
		var warrantyID pgtype.Text

		err := rows.Scan(
			&claim.ID,
			&warrantyID,
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
			&supportingDocURL,
			&brand,
			&quantity,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %v", err)
		}

		// Convert pgtype values to Go types
		if warrantyID.Valid {
			claim.WarrantyID = &warrantyID.String
		}
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
		if brand.Valid {
			claim.Brand = &brand.String
		}
		if quantity.Valid {
			q := int(quantity.Int32)
			claim.Quantity = &q
		}
		if createdAt.Valid {
			claim.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			claim.UpdatedAt = updatedAt.Time
		}
		if supportingDocURL.Valid {
			claim.SupportingDocURL = &supportingDocURL.String
		}

		requestedTyreDetails, err := getRequestedTyreDetails(claim.ID)
		if err != nil {
			return nil, err
		}
		claim.RequestedTyreDetails = requestedTyreDetails

		claims = append(claims, claim)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating claims: %v", err)
	}

	return claims, nil
}
