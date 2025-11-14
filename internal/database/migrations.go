package database

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrateFromFiles runs all SQL migrations from the migrations directory.
// Migration files are loaded from the root-level migrations/ folder at runtime.
// This allows both the application and wrangler CLI to use the same migrations.
func (db *DB) MigrateFromFiles(ctx context.Context) error {
	log.Println("Running database migrations from files...")

	// Read all migration files from the root migrations folder
	entries, err := fs.ReadDir(os.DirFS("migrations"), ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name (which includes version number)
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration
	for _, filename := range migrationFiles {
		log.Printf("Applying migration: %s", filename)

		// Read migration file from the migrations directory
		content, err := os.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		_, err = db.ExecContext(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	log.Println("All migrations completed successfully")
	return nil
}
