# Database Migrations

This directory contains SQL migration files for the CloudflareDB application.

## Migration Files

Migration files are numbered sequentially and executed in order:

- `001_create_users_table.sql` - Initial users table and indexes

## Naming Convention

Migration files follow this naming pattern:

```
<version>_<description>.sql
```

Example: `001_create_users_table.sql`

- **Version**: Three-digit number (001, 002, 003...)
- **Description**: Snake_case description of the migration
- **Extension**: Always `.sql`

## How Migrations Work

### Automatic Migrations (Application Startup)

Migrations are automatically applied when the application starts:

```go
// In cmd/api/main.go
if err := db.MigrateFromFiles(ctx); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

The application:
1. Reads all `.sql` files from the `migrations/` directory
2. Sorts them by filename (alphabetically)
3. Executes each migration in order
4. Uses `IF NOT EXISTS` clauses for idempotency

### Manual Migrations with Wrangler (Cloudflare D1)

For Cloudflare D1 production databases, you can run migrations manually:

#### Run a Single Migration

```bash
npx wrangler d1 execute cloudflaredb \
  --remote \
  --file=migrations/001_create_users_table.sql
```

#### Run All Migrations

```bash
# Run all migration files in order
for file in migrations/*.sql; do
  echo "Running migration: $file"
  npx wrangler d1 execute cloudflaredb --remote --file="$file"
done
```

#### Test Locally First

Always test migrations locally before running in production:

```bash
# Create a local D1 database for testing
npx wrangler d1 execute cloudflaredb \
  --local \
  --file=migrations/001_create_users_table.sql
```

## Creating New Migrations

### Step 1: Create Migration File

Create a new file in the `migrations/` directory with the next sequential number:

```bash
# Example: migrations/002_add_users_age_column.sql
cat > migrations/002_add_users_age_column.sql << 'EOF'
-- Migration: Add age column to users table
-- Created: 2025-11-13
-- Description: Add optional age field for user profiles

ALTER TABLE users ADD COLUMN age INTEGER;

CREATE INDEX IF NOT EXISTS idx_users_age ON users(age);
EOF
```

### Step 2: Test Locally

Test with SQLite:

```bash
# Apply to local database
sqlite3 local.db < migrations/002_add_users_age_column.sql

# Verify
sqlite3 local.db "SELECT sql FROM sqlite_master WHERE name='users';"
```

### Step 3: Run in Application

The migration will automatically run when you start the application:

```bash
go run cmd/api/main.go
```

You should see:
```
Running database migrations from files...
Applying migration: 001_create_users_table.sql
Successfully applied migration: 001_create_users_table.sql
Applying migration: 002_add_users_age_column.sql
Successfully applied migration: 002_add_users_age_column.sql
All migrations completed successfully
```

### Step 4: Deploy to Cloudflare D1

```bash
# Run in production
npx wrangler d1 execute cloudflaredb \
  --remote \
  --file=migrations/002_add_users_age_column.sql
```

## Migration Best Practices

### 1. Use IF NOT EXISTS

Always use `IF NOT EXISTS` clauses to make migrations idempotent:

```sql
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### 2. Avoid Destructive Operations

Never use `DROP TABLE` or `DROP COLUMN` in migrations. If you need to remove something:

```sql
-- BAD: Don't do this
DROP TABLE old_table;

-- GOOD: Mark as deprecated in comments
-- Table 'old_table' is deprecated as of migration 005
-- TODO: Remove after migration 010
```

### 3. Add Comments

Document your migrations:

```sql
-- Migration: Create posts table
-- Created: 2025-11-13
-- Description: Add posts table for user-generated content
-- Related: Issue #123

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 4. Create Indexes

Always create indexes for foreign keys and commonly queried columns:

```sql
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
```

### 5. Test Before Production

Always test migrations on a copy of production data:

```bash
# Export production data
npx wrangler d1 export cloudflaredb --remote > backup.sql

# Create test database
npx wrangler d1 create cloudflaredb-test

# Import data
npx wrangler d1 execute cloudflaredb-test --remote --file=backup.sql

# Test migration
npx wrangler d1 execute cloudflaredb-test --remote --file=migrations/002_new_migration.sql
```

### 6. Keep Migrations Small

One logical change per migration:

```
✓ GOOD: 002_add_user_age.sql
✓ GOOD: 003_create_posts_table.sql

✗ BAD: 002_add_age_and_posts_and_comments.sql
```

## Rollback Strategy

SQLite and Cloudflare D1 don't support transactional DDL for all statements. For rollbacks:

### Option 1: Time Travel (D1 Only)

Cloudflare D1 maintains 30 days of backups:

```bash
# Restore to a point before the migration
# (Use Cloudflare Dashboard for Time Travel)
```

### Option 2: Rollback Migration

Create a reverse migration:

```sql
-- Migration: 002_add_users_age_column.sql
ALTER TABLE users ADD COLUMN age INTEGER;

-- Rollback: 002_rollback_add_users_age_column.sql
ALTER TABLE users DROP COLUMN age;
```

### Option 3: Backup Before Migration

```bash
# Backup before migrating
cp local.db local.db.backup

# Or export D1 data
npx wrangler d1 export cloudflaredb --remote > backup_$(date +%Y%m%d_%H%M%S).sql
```

## Troubleshooting

### Migration Fails with "table already exists"

**Cause**: Missing `IF NOT EXISTS` clause

**Solution**: Update migration to use `IF NOT EXISTS`:
```sql
CREATE TABLE IF NOT EXISTS tablename (...);
```

### Migration Works Locally but Fails in D1

**Cause**: SQLite vs D1 dialect differences

**Solution**: Test with D1 local mode first:
```bash
npx wrangler d1 execute cloudflaredb --local --file=migrations/xxx.sql
```

### Migrations Run Out of Order

**Cause**: Incorrect file naming

**Solution**: Use zero-padded numbers (001, 002, not 1, 2)

### Application Starts but Migrations Don't Run

**Cause**: Embedded filesystem not updated

**Solution**: Rebuild the application:
```bash
go build -o bin/api cmd/api/main.go
```

## Monitoring Migrations

### Check Applied Migrations

```bash
# Local SQLite
sqlite3 local.db ".tables"
sqlite3 local.db ".schema"

# Cloudflare D1
npx wrangler d1 execute cloudflaredb --remote --command "SELECT name FROM sqlite_master WHERE type='table';"
```

### Verify Schema

```bash
# Show table structure
sqlite3 local.db ".schema users"

# Or in D1
npx wrangler d1 execute cloudflaredb --remote --command "SELECT sql FROM sqlite_master WHERE name='users';"
```

## Example Migration Workflow

```bash
# 1. Create migration
cat > migrations/002_add_posts_table.sql << 'EOF'
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
EOF

# 2. Test locally
sqlite3 test.db < migrations/002_add_posts_table.sql
sqlite3 test.db ".schema posts"

# 3. Run in development
go run cmd/api/main.go

# 4. Commit to git
git add migrations/002_add_posts_table.sql
git commit -m "Add posts table migration"

# 5. Deploy to production
npx wrangler d1 execute cloudflaredb --remote --file=migrations/002_add_posts_table.sql

# 6. Verify
npx wrangler d1 execute cloudflaredb --remote --command "SELECT COUNT(*) FROM posts;"
```

## Additional Resources

- [Cloudflare D1 Documentation](https://developers.cloudflare.com/d1/)
- [SQLite ALTER TABLE](https://www.sqlite.org/lang_altertable.html)
- [Database Setup Guide](../docs/DATABASE_SETUP.md)

## Need Help?

If you encounter issues with migrations:
1. Check [Troubleshooting Guide](../docs/TROUBLESHOOTING.md)
2. Review migration logs in application output
3. Test migrations locally first
4. Create a backup before running in production
