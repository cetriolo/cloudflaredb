#!/bin/bash

# create-migration.sh
# Script to create a new database migration file

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

MIGRATIONS_DIR="migrations"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get next migration number
get_next_number() {
    local last_migration=$(ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort | tail -n 1)

    if [ -z "$last_migration" ]; then
        echo "001"
    else
        local last_number=$(basename "$last_migration" | cut -d'_' -f1)
        local next_number=$((10#$last_number + 1))
        printf "%03d" "$next_number"
    fi
}

# Function to convert description to snake_case
to_snake_case() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '_' | tr -cd '[:alnum:]_'
}

# Main execution
main() {
    echo "=========================================="
    echo "  Create New Migration"
    echo "=========================================="
    echo ""

    # Check if migrations directory exists
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        mkdir -p "$MIGRATIONS_DIR"
        print_info "Created migrations directory"
    fi

    # Get migration description
    if [ -z "$1" ]; then
        read -p "Enter migration description: " description
    else
        description="$1"
    fi

    if [ -z "$description" ]; then
        print_error "Description cannot be empty"
        exit 1
    fi

    # Generate filename
    migration_number=$(get_next_number)
    snake_case_desc=$(to_snake_case "$description")
    filename="${migration_number}_${snake_case_desc}.sql"
    filepath="$MIGRATIONS_DIR/$filename"

    # Create migration file
    cat > "$filepath" << EOF
-- Migration: $description
-- Created: $(date +%Y-%m-%d)
-- Description: TODO: Add detailed description

-- Add your SQL statements here

-- Example: Create a new table
-- CREATE TABLE IF NOT EXISTS table_name (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     name TEXT NOT NULL,
--     created_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );

-- Example: Add a column
-- ALTER TABLE users ADD COLUMN new_column TEXT;

-- Example: Create an index
-- CREATE INDEX IF NOT EXISTS idx_table_column ON table_name(column_name);

-- Remember:
-- 1. Always use IF NOT EXISTS for idempotency
-- 2. Create indexes for foreign keys and frequently queried columns
-- 3. Test locally before running in production
EOF

    print_info "Created migration file: $filepath"
    echo ""
    echo "Next steps:"
    echo "  1. Edit the migration file and add your SQL statements"
    echo "  2. Test locally: sqlite3 test.db < $filepath"
    echo "  3. Run in app: go run cmd/api/main.go"
    echo "  4. Deploy to D1: ./scripts/run-migrations.sh remote"
    echo ""
}

# Run main function
main "$@"
