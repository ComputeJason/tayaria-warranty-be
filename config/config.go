package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL   string
	SupabaseKey   string
	Environment   string
	DatabaseURL   string
	StorageBucket string
}

var AppConfig Config

func Init() error {
	// Load environment variables based on environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		log.Fatal("APP_ENV environment variable is not set")
	}

	// Only try to load .env file if not in production (Render uses environment variables)
	if env != "production" {
		// Load the appropriate .env file
		envFile := ".env." + env
		log.Printf("Loading environment file: %s", envFile)

		// Look for .env file in the current directory
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		}
	} else {
		log.Printf("Production environment detected, using system environment variables")
	}

	// Debug: Print all environment variables
	log.Printf("Environment variables loaded:")
	log.Printf("SUPABASE_URL: %s", os.Getenv("SUPABASE_URL"))
	log.Printf("SUPABASE_KEY: %s", os.Getenv("SUPABASE_KEY"))
	log.Printf("DATABASE_URL: %s", os.Getenv("DATABASE_URL"))
	log.Printf("STORAGE_BUCKET: %s", os.Getenv("STORAGE_BUCKET"))

	AppConfig = Config{
		SupabaseURL:   os.Getenv("SUPABASE_URL"),
		SupabaseKey:   os.Getenv("SUPABASE_KEY"),
		Environment:   env,
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		StorageBucket: os.Getenv("STORAGE_BUCKET"),
	}

	// Validate required fields
	if AppConfig.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is not set")
	}
	if AppConfig.SupabaseKey == "" {
		return fmt.Errorf("SUPABASE_KEY is not set")
	}

	return nil
}

func IsProduction() bool {
	return strings.ToLower(AppConfig.Environment) == "production"
}
