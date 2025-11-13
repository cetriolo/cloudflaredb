package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Environment      string
	Port             string
	DatabaseDriver   string
	DatabaseDSN      string
	CloudflareAccID  string
	CloudflareAPIKey string
	CloudflareDBName string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Environment:      getEnv("ENVIRONMENT", "development"),
		Port:             getEnv("PORT", "8080"),
		DatabaseDriver:   getEnv("DATABASE_DRIVER", "sqlite3"),
		CloudflareAccID:  os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareAPIKey: os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareDBName: os.Getenv("CLOUDFLARE_DB_NAME"),
	}

	// Set DSN based on driver
	if cfg.DatabaseDriver == "cfd1" {
		// Cloudflare D1 driver
		if cfg.CloudflareAccID == "" || cfg.CloudflareAPIKey == "" || cfg.CloudflareDBName == "" {
			return nil, fmt.Errorf("CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_API_TOKEN, and CLOUDFLARE_DB_NAME are required for cfd1 driver")
		}
		cfg.DatabaseDSN = fmt.Sprintf("d1://%s:%s@%s",
			cfg.CloudflareAccID,
			cfg.CloudflareAPIKey,
			cfg.CloudflareDBName,
		)
	} else {
		// SQLite driver for local development
		cfg.DatabaseDSN = getEnv("DATABASE_DSN", "./local.db")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
