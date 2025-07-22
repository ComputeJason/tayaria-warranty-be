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
	log.Printf("AWS_ACCESS_KEY_ID: %s", os.Getenv("AWS_ACCESS_KEY_ID"))
	log.Printf("AWS_SECRET_ACCESS_KEY: %s", maskSecret(os.Getenv("AWS_SECRET_ACCESS_KEY")))
	log.Printf("AWS_REGION: %s", os.Getenv("AWS_REGION"))

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

	// Validate AWS credentials for S3 functionality
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		log.Printf("Warning: AWS_ACCESS_KEY_ID is not set - S3 functionality may not work")
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("Warning: AWS_SECRET_ACCESS_KEY is not set - S3 functionality may not work")
	}
	if os.Getenv("AWS_REGION") == "" {
		log.Printf("Warning: AWS_REGION is not set - S3 functionality may not work")
	}

	return nil
}

// maskSecret masks a secret string for logging (shows first 4 and last 4 characters)
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

func IsProduction() bool {
	return strings.ToLower(AppConfig.Environment) == "production"
}
