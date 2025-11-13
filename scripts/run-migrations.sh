#!/bin/bash

# run-migrations.sh
# Script to run database migrations on Cloudflare D1

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DATABASE_NAME="${CLOUDFLARE_DB_NAME:-cloudflaredb}"
MIGRATIONS_DIR="migrations"
ENVIRONMENT="${1:-remote}"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if wrangler is installed
check_wrangler() {
    if ! command -v npx &> /dev/null; then
        print_error "npx not found. Please install Node.js first."
        exit 1
    fi

    print_info "Checking wrangler installation..."
    if ! npx wrangler --version &> /dev/null; then
        print_error "wrangler not found. Installing..."
        npm install -g wrangler
    fi
}

# Function to validate environment
validate_environment() {
    if [ "$ENVIRONMENT" != "local" ] && [ "$ENVIRONMENT" != "remote" ]; then
        print_error "Invalid environment. Use 'local' or 'remote'"
        echo "Usage: $0 [local|remote]"
        exit 1
    fi
}

# Function to check if migrations directory exists
check_migrations_dir() {
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        print_error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi

    # Count migration files
    migration_count=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f | wc -l)
    if [ "$migration_count" -eq 0 ]; then
        print_warning "No migration files found in $MIGRATIONS_DIR"
        exit 0
    fi

    print_info "Found $migration_count migration file(s)"
}

# Function to run a single migration
run_migration() {
    local file=$1
    local filename=$(basename "$file")

    print_info "Running migration: $filename"

    if [ "$ENVIRONMENT" = "local" ]; then
        npx wrangler d1 execute "$DATABASE_NAME" --local --file="$file"
    else
        npx wrangler d1 execute "$DATABASE_NAME" --remote --file="$file"
    fi

    if [ $? -eq 0 ]; then
        print_info "✓ Successfully applied: $filename"
    else
        print_error "✗ Failed to apply: $filename"
        exit 1
    fi
}

# Function to confirm production migrations
confirm_production() {
    if [ "$ENVIRONMENT" = "remote" ]; then
        print_warning "You are about to run migrations on PRODUCTION database: $DATABASE_NAME"
        read -p "Are you sure you want to continue? (yes/no): " confirm

        if [ "$confirm" != "yes" ]; then
            print_info "Migration cancelled"
            exit 0
        fi
    fi
}

# Function to backup before migration
create_backup() {
    if [ "$ENVIRONMENT" = "remote" ]; then
        print_info "Creating backup before migration..."
        backup_file="backups/backup_$(date +%Y%m%d_%H%M%S).sql"
        mkdir -p backups

        npx wrangler d1 export "$DATABASE_NAME" --remote > "$backup_file" 2>/dev/null || true

        if [ -f "$backup_file" ] && [ -s "$backup_file" ]; then
            print_info "✓ Backup created: $backup_file"
        else
            print_warning "Could not create backup (this is normal if database is empty)"
        fi
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  Cloudflare D1 Migration Runner"
    echo "=========================================="
    echo ""

    print_info "Environment: $ENVIRONMENT"
    print_info "Database: $DATABASE_NAME"
    echo ""

    # Validation
    validate_environment
    check_wrangler
    check_migrations_dir
    confirm_production

    # Create backup
    create_backup

    # Find all migration files and sort them
    print_info "Collecting migration files..."
    mapfile -t migrations < <(find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort)

    # Run migrations
    echo ""
    print_info "Starting migrations..."
    echo ""

    success_count=0
    for migration in "${migrations[@]}"; do
        run_migration "$migration"
        ((success_count++))
    done

    echo ""
    echo "=========================================="
    print_info "Migration completed successfully!"
    print_info "Applied $success_count migration(s)"
    echo "=========================================="
}

# Run main function
main
