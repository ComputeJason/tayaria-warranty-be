package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
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
