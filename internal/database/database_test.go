package database

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		driver     string
		dsn        string
		wantErr    bool
		skipReason string
	}{
		{
			name:    "sqlite3 in-memory",
			driver:  "sqlite3",
			dsn:     ":memory:",
			wantErr: false,
		},
		{
			name:    "invalid driver",
			driver:  "invalid_driver",
			dsn:     "some_dsn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			db, err := New(tt.driver, tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if db == nil {
					t.Error("Expected non-nil database")
				}
				defer func(db *DB) {
					err := db.Close()
					if err != nil {
						t.Errorf("Failed to close database: %v", err)
					}
				}(db)

				// Verify connection
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := db.PingContext(ctx); err != nil {
					t.Errorf("Failed to ping database: %v", err)
				}
			}
		})
	}
}

func TestDB_Migrate(t *testing.T) {
	db, err := New("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func(db *DB) {
		err := db.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}(db)

	ctx := context.Background()

	// Run migrations
	if err := db.Migrate(ctx); err != nil {
		t.Errorf("Migrate() error = %v", err)
	}

	// Verify table exists
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name='users'"
	var tableName string
	err = db.QueryRowContext(ctx, query).Scan(&tableName)
	if err != nil {
		t.Errorf("Failed to verify table creation: %v", err)
	}

	if tableName != "users" {
		t.Errorf("Expected table 'users', got '%s'", tableName)
	}

	// Verify index exists
	indexQuery := "SELECT name FROM sqlite_master WHERE type='index' AND name='idx_users_email'"
	var indexName string
	err = db.QueryRowContext(ctx, indexQuery).Scan(&indexName)
	if err != nil {
		t.Errorf("Failed to verify index creation: %v", err)
	}

	if indexName != "idx_users_email" {
		t.Errorf("Expected index 'idx_users_email', got '%s'", indexName)
	}

	// Run migrations again (should be idempotent)
	if err := db.Migrate(ctx); err != nil {
		t.Errorf("Migrate() second run error = %v", err)
	}
}
