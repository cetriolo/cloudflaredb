# Database Setup Guide

This guide provides detailed instructions for setting up the database for both local development and production environments.

## Table of Contents

- [Local Development with SQLite](#local-development-with-sqlite)
- [Production with Cloudflare D1](#production-with-cloudflare-d1)
- [Database Schema](#database-schema)
- [Migrations](#migrations)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Optimization](#performance-optimization)

## Local Development with SQLite

### Quick Start

SQLite is used for local development and requires minimal setup.

1. **No installation required** - SQLite is embedded in the Go application via the `mattn/go-sqlite3` driver.

2. **Configuration**

Create a `.env` file:

```env
DATABASE_DRIVER=sqlite3
DATABASE_DSN=./local.db
```

3. **Start the application**

```bash
go run cmd/api/main.go
```

The database file `local.db` will be created automatically in your project root.

### SQLite Configuration Options

The DSN (Data Source Name) for SQLite supports various options:

```env
# In-memory database (data lost on restart)
DATABASE_DSN=:memory:

# File-based database with options
DATABASE_DSN=./local.db?cache=shared&mode=rwc
```

Common options:
- `cache=shared` - Enable shared cache mode
- `mode=rwc` - Read, write, and create if not exists
- `_journal_mode=WAL` - Use Write-Ahead Logging for better concurrency

### SQLite Best Practices

1. **Enable WAL mode** for better write concurrency:
```sql
PRAGMA journal_mode=WAL;
```

2. **Regular vacuuming** to optimize database size:
```sql
VACUUM;
```

3. **Monitor database size**:
```bash
ls -lh local.db
```

## Production with Cloudflare D1

### Prerequisites

- Cloudflare account
- Wrangler CLI installed
- Node.js 16 or higher

### Step 1: Install Wrangler

```bash
npm install -g wrangler
# or
npx wrangler --version
```

### Step 2: Authenticate with Cloudflare

```bash
npx wrangler login
```

This will open your browser to authenticate with Cloudflare.

### Step 3: Create D1 Database

```bash
npx wrangler d1 create cloudflaredb
```

**Output:**
```
✅ Successfully created DB 'cloudflaredb'

[[d1_databases]]
binding = "DB"
database_name = "cloudflaredb"
database_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

**Important:** Save the `database_id` - you'll need it!

### Step 4: Get Your Account ID

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Select your domain (or Workers & Pages)
3. Your Account ID is in the right sidebar

Or use Wrangler:
```bash
npx wrangler whoami
```

### Step 5: Create API Token

1. Go to [API Tokens](https://dash.cloudflare.com/profile/api-tokens)
2. Click "Create Token"
3. Use "Edit Cloudflare Workers" template or create custom token:
   - **Permissions:**
     - Account > D1 > Edit
     - Account > Workers Scripts > Edit
   - **Account Resources:**
     - Include > Your Account

4. Copy the generated token immediately (it won't be shown again!)

### Step 6: Configure Application

Create or update `.env`:

```env
DATABASE_DRIVER=cfd1
CLOUDFLARE_ACCOUNT_ID=your_account_id_here
CLOUDFLARE_API_TOKEN=your_api_token_here
CLOUDFLARE_DB_NAME=cloudflaredb
```

**Security Note:** Never commit `.env` to version control!

### Step 7: Initialize Database Schema

The application automatically runs migrations on startup. However, you can also run them manually:

```bash
npx wrangler d1 execute cloudflaredb --remote --command "
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
"
```

### Step 8: Verify Setup

Start the application:

```bash
go run cmd/api/main.go
```

Check the logs for successful connection:
```
2025/11/13 10:00:00 Starting application in production mode
2025/11/13 10:00:00 Using database driver: cfd1
2025/11/13 10:00:00 Database connection established
2025/11/13 10:00:00 Database migrations completed
2025/11/13 10:00:00 Server starting on port 8080
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**Columns:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY, AUTOINCREMENT | Unique user identifier |
| `email` | TEXT | NOT NULL, UNIQUE | User email address |
| `name` | TEXT | NOT NULL | User full name |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

**Indexes:**
- `idx_users_email` on `email` - Improves email lookup performance

## Migrations

### Automatic Migrations

The application automatically runs migrations on startup. Migrations are defined in:
- `internal/database/database.go` - `Migrate()` function

### Manual Migrations (Cloudflare D1)

#### Create Migration File

Create `migrations/001_initial_schema.sql`:

```sql
-- Create users table
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

#### Execute Migration

```bash
# Remote (production)
npx wrangler d1 execute cloudflaredb --remote --file=migrations/001_initial_schema.sql

# Local (for testing with wrangler)
npx wrangler d1 execute cloudflaredb --local --file=migrations/001_initial_schema.sql
```

### Migration Best Practices

1. **Use IF NOT EXISTS** - Makes migrations idempotent
2. **Version your migrations** - Use numbered prefixes (001_, 002_, etc.)
3. **Test locally first** - Always test migrations on local/dev before production
4. **Backup before migrations** - Use D1 Time Travel for rollback capability
5. **Keep migrations small** - One logical change per migration

## Backup and Recovery

### SQLite (Local Development)

#### Backup

```bash
# Simple file copy
cp local.db local.db.backup

# Or use SQLite backup command
sqlite3 local.db ".backup local.db.backup"
```

#### Restore

```bash
cp local.db.backup local.db
```

### Cloudflare D1 (Production)

D1 uses **Time Travel** for point-in-time recovery.

#### View Database Info

```bash
npx wrangler d1 info cloudflaredb
```

#### Query Historical Data

D1 automatically maintains backups for 30 days. You can restore to any point in time using the Cloudflare Dashboard or API.

#### Export Data

```bash
# Export to JSON
npx wrangler d1 execute cloudflaredb --remote --command "SELECT * FROM users" --json > backup.json

# Export specific table
npx wrangler d1 execute cloudflaredb --remote --command "SELECT * FROM users WHERE created_at > '2025-01-01'" --json > recent_users.json
```

#### Import Data

```bash
# From SQL file
npx wrangler d1 execute cloudflaredb --remote --file=import.sql
```

## Performance Optimization

### Indexing Strategy

Current indexes:
- `idx_users_email` - Optimizes email lookups

Add more indexes based on query patterns:

```sql
-- If you frequently query by name
CREATE INDEX idx_users_name ON users(name);

-- Composite index for complex queries
CREATE INDEX idx_users_created_email ON users(created_at, email);
```

### Query Optimization Tips

1. **Use EXPLAIN QUERY PLAN**

```sql
EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = 'test@example.com';
```

2. **Limit result sets**
```sql
SELECT * FROM users LIMIT 100;
```

3. **Use prepared statements** - Already implemented in repository layer

4. **Avoid SELECT ***
```sql
-- Good
SELECT id, email, name FROM users;

-- Avoid
SELECT * FROM users;
```

### Connection Pool Settings

Configured in `internal/database/database.go`:

```go
db.SetMaxOpenConns(25)     // Max open connections
db.SetMaxIdleConns(5)      // Max idle connections
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(2 * time.Minute)
```

Adjust based on your workload:
- **High traffic**: Increase `MaxOpenConns` to 50-100
- **Low traffic**: Decrease to 10-25 to save resources

### D1-Specific Optimizations

1. **Batch operations** - Group multiple inserts/updates
2. **Use Read Replicas** - Enable global read replication for read-heavy workloads
3. **Cache frequently accessed data** - Consider adding Redis/KV for hot data
4. **Monitor query performance** - Use Cloudflare Analytics

## Testing Database Connections

### SQLite

```bash
# Open database
sqlite3 local.db

# Run query
sqlite> SELECT * FROM users;

# Check schema
sqlite> .schema users

# Exit
sqlite> .quit
```

### Cloudflare D1

```bash
# List databases
npx wrangler d1 list

# Query database
npx wrangler d1 execute cloudflaredb --remote --command "SELECT COUNT(*) FROM users"

# Interactive mode
npx wrangler d1 execute cloudflaredb --remote
```

## Troubleshooting

### Common Issues

#### SQLite: Database is locked

**Cause:** Multiple processes accessing the same database file

**Solution:**
```bash
# Stop all instances
pkill -f "go run cmd/api/main.go"

# Remove WAL files
rm -f *.db-shm *.db-wal

# Restart
go run cmd/api/main.go
```

#### D1: Authentication failed

**Cause:** Invalid API token or account ID

**Solution:**
1. Verify credentials in `.env`
2. Check token permissions
3. Regenerate token if needed

#### D1: Database not found

**Cause:** Incorrect database name or ID

**Solution:**
```bash
# List all D1 databases
npx wrangler d1 list

# Verify the name matches your .env
```

#### Connection timeout

**Cause:** Network issues or rate limiting

**Solution:**
1. Check internet connection
2. Verify Cloudflare status
3. Implement retry logic (already in place)

## Security Considerations

### SQLite

1. **File permissions**
```bash
chmod 600 local.db
```

2. **Don't expose database files** - Add to `.gitignore`

3. **Use parameterized queries** - Prevents SQL injection (already implemented)

### Cloudflare D1

1. **Rotate API tokens regularly**
2. **Use least privilege** - Only grant necessary permissions
3. **Monitor access logs** - Use Cloudflare Analytics
4. **Enable audit logging** - Track database changes
5. **Never commit credentials** - Use environment variables

## Additional Resources

- [Cloudflare D1 Documentation](https://developers.cloudflare.com/d1/)
- [D1 Best Practices](https://developers.cloudflare.com/d1/best-practices/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Go database/sql Package](https://pkg.go.dev/database/sql)
