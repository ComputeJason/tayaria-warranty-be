package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db *pgxpool.Pool
)

func Init() error {
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Disable prepared statement cache
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	db, err = pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return err
	}

	// Test the connection
	log.Printf("Testing database connection...")
	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to database (using pool)")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

// GetUserByFullName searches for a user by their full name
func GetUserByFullName(fullName string) (map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, name, email, phone_number, created_at, updated_at
		FROM users
		WHERE name = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, fullName)

	row := db.QueryRow(context.Background(), query, fullName)

	// Create a map to store the result
	result := make(map[string]interface{})

	// Scan the result into variables
	var id, fullNameResult, email, phoneNumber string
	var createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(&id, &fullNameResult, &email, &phoneNumber, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user data: %v", err)
	}

	// Populate the result map
	result["id"] = id
	result["full_name"] = fullNameResult
	result["email"] = email
	result["phone_number"] = phoneNumber
	if createdAt.Valid {
		result["created_at"] = createdAt.Time
	}
	if updatedAt.Valid {
		result["updated_at"] = updatedAt.Time
	}

	return result, nil
}

// GetUserByID searches for a user by their ID
func GetUserByID(userID string) (map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, full_name, email, phone_number, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	log.Printf("Executing SQL query: %s with params: [%s]", query, userID)

	row := db.QueryRow(context.Background(), query, userID)

	// Create a map to store the result
	result := make(map[string]interface{})

	// Scan the result into variables
	var id, fullName, email, phoneNumber string
	var createdAt, updatedAt pgtype.Timestamp

	err := row.Scan(&id, &fullName, &email, &phoneNumber, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user data: %v", err)
	}

	// Populate the result map
	result["id"] = id
	result["full_name"] = fullName
	result["email"] = email
	result["phone_number"] = phoneNumber
	if createdAt.Valid {
		result["created_at"] = createdAt.Time
	}
	if updatedAt.Valid {
		result["updated_at"] = updatedAt.Time
	}

	return result, nil
}
