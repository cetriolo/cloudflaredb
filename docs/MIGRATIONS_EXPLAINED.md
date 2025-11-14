# How Migrations Work - Complete Explanation

This document explains exactly how database migrations work in this application.

## The Short Answer

**✅ Migrations run automatically when the application starts!**

You don't need to do anything manually. Just deploy your Docker container or run the app, and migrations will execute automatically.

## How It Works

### 1. On Application Startup

When you start the application (via Docker, docker-compose, or directly with `go run`), this happens:

```
main.go (lines 42-48)
    ↓
db.MigrateFromFiles(ctx)
    ↓
Reads all .sql files from migrations/ folder
    ↓
Sorts them alphabetically (001, 002, 003...)
    ↓
Executes each migration against the database
    ↓
Migration complete! App starts serving
```

### 2. The Code

In `cmd/api/main.go`:

```go
// Run database migrations to ensure schema is up to date
ctx := context.Background()
if err := db.MigrateFromFiles(ctx); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}

log.Println("Database migrations completed")
```

This code runs **every time** the application starts, before accepting any HTTP requests.

### 3. What MigrateFromFiles Does

Located in `internal/database/migrations.go`:

```go
func (db *DB) MigrateFromFiles(ctx context.Context) error {
    // 1. Read all .sql files from migrations/ folder
    entries, err := fs.ReadDir(os.DirFS("migrations"), ".")

    // 2. Sort them by filename
    sort.Strings(migrationFiles)

    // 3. Execute each migration in order
    for _, filename := range migrationFiles {
        content, _ := os.ReadFile(filepath.Join("migrations", filename))
        _, err = db.ExecContext(ctx, string(content))
    }
}
```

## Idempotent Migrations

All migrations use `IF NOT EXISTS` clauses:

```sql
-- ✅ SAFE: Can run multiple times
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ...
);

-- ✅ SAFE: Can run multiple times
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

This means:
- **First run**: Table is created
- **Second run**: Nothing happens (table exists, skipped)
- **Third run**: Still nothing happens

**No errors, no duplicates, just safe execution!**

## What You See in Logs

### Successful Migration

```
2025/11/14 11:43:32 Starting application in production mode
2025/11/14 11:43:32 Using database driver: cfd1
2025/11/14 11:43:32 Database connection established
2025/11/14 11:43:32 Running database migrations from files...
2025/11/14 11:43:32 Applying migration: 001_create_users_table.sql
2025/11/14 11:43:32 Successfully applied migration: 001_create_users_table.sql
2025/11/14 11:43:32 Applying migration: 002_create_rooms_table.sql
2025/11/14 11:43:33 Successfully applied migration: 002_create_rooms_table.sql
2025/11/14 11:43:33 All migrations completed successfully
2025/11/14 11:43:33 Database migrations completed
2025/11/14 11:43:33 Server starting on port 8080
```

### Migration Already Applied (Still Success!)

```
2025/11/14 11:45:00 Running database migrations from files...
2025/11/14 11:45:00 Applying migration: 001_create_users_table.sql
2025/11/14 11:45:00 Successfully applied migration: 001_create_users_table.sql
                    ↑ No error! IF NOT EXISTS prevented duplicate creation
```

## In Different Environments

### Docker / Docker Compose

```bash
docker-compose up -d
```

**What happens:**
1. Container starts
2. Application binary (`./api`) runs
3. **Migrations execute automatically** ✅
4. HTTP server starts
5. Container is ready

**You see:**
```
cloudflaredb-app-1  | Running database migrations from files...
cloudflaredb-app-1  | Applying migration: 001_create_users_table.sql
cloudflaredb-app-1  | Successfully applied migration: 001_create_users_table.sql
cloudflaredb-app-1  | Server starting on port 8080
```

### Local Development

```bash
go run cmd/api/main.go
```

**What happens:**
1. Go compiles and runs the code
2. **Migrations execute automatically** ✅
3. HTTP server starts

### Production Deployment

Whether you deploy to:
- Cloud Run
- AWS ECS
- Kubernetes
- Any other platform

**Migrations always run automatically on startup!** ✅

## When to Use Manual Migrations

You might want to run migrations manually with `wrangler` in these cases:

### 1. Pre-Seeding Production Database

Before your first deployment:

```bash
# Create tables before deploying app
npx wrangler d1 execute cloudflaredb --remote --file=migrations/001_create_users_table.sql
npx wrangler d1 execute cloudflaredb --remote --file=migrations/002_create_rooms_table.sql

# Now deploy app (migrations will run again but do nothing)
docker-compose up -d
```

**Why?** Separates database setup from app deployment for peace of mind.

### 2. Testing Migrations Locally

```bash
# Test migration on local D1 instance
npx wrangler d1 execute cloudflaredb --local --file=migrations/003_new_migration.sql

# Verify it works
npx wrangler d1 execute cloudflaredb --local --command "SELECT * FROM new_table;"

# Then commit and deploy (app will run it automatically)
git add migrations/003_new_migration.sql
git commit -m "Add new migration"
git push
```

**Why?** Catch errors before deployment.

### 3. CI/CD Pipeline

In your GitHub Actions workflow:

```yaml
- name: Run Migrations
  run: |
    npm install -g wrangler
    npx wrangler d1 execute ${{ secrets.CLOUDFLARE_DB_NAME }} \
      --remote \
      --file=migrations/001_create_users_table.sql

- name: Deploy App
  run: docker-compose up -d
```

**Why?** Separate migration execution from deployment for better control and rollback options.

### 4. Database-Only Updates

If you only want to update the database without deploying the app:

```bash
# Add new migration file
cat > migrations/003_add_posts_table.sql << 'EOF'
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ...
);
EOF

# Run manually without deploying app
npx wrangler d1 execute cloudflaredb --remote --file=migrations/003_add_posts_table.sql

# Later, deploy app (migration will run again, but safely skipped)
```

## Common Questions

### Q: Do I need to run migrations manually before deploying?

**A: No!** The app runs them automatically. Manual execution is optional.

### Q: What if I deploy before running migrations manually?

**A: That's fine!** The app will run them automatically on startup.

### Q: What if I run migrations manually AND the app runs them?

**A: No problem!** The `IF NOT EXISTS` clauses make them safe to run multiple times.

### Q: Can migrations run concurrently if I scale to multiple containers?

**A: Yes, it's safe!** D1 uses SQLite's locking mechanism. If two containers try to run migrations simultaneously:
- One acquires the lock and executes
- The other waits
- Both succeed because of `IF NOT EXISTS`

However, for cleaner logs, you might want to run migrations once before scaling.

### Q: What if a migration fails?

**A: The app won't start!** It will log the error and exit:

```
2025/11/14 12:00:00 Running database migrations from files...
2025/11/14 12:00:00 Applying migration: 003_bad_migration.sql
2025/11/14 12:00:00 Failed to run migrations: syntax error near "BAAD"
```

This is intentional - the app won't serve requests with a broken database schema.

### Q: How do I check what migrations have run?

Check your database schema:

```bash
# Via wrangler
npx wrangler d1 execute cloudflaredb --remote --command "SELECT name FROM sqlite_master WHERE type='table';"

# Via the app (if it's running)
curl http://localhost:8080/users  # If users table exists, migration ran
```

### Q: Can I disable automatic migrations?

**A: Not recommended**, but you could comment out these lines in `main.go`:

```go
// if err := db.MigrateFromFiles(ctx); err != nil {
//     log.Fatalf("Failed to run migrations: %v", err)
// }
```

**Why not recommended?**
- Easy to forget to run migrations
- Different environments might have different schemas
- Automatic migrations ensure consistency

## Migration Execution Flow Chart

```
┌─────────────────────────┐
│   Application Starts    │
└───────────┬─────────────┘
            │
            ↓
┌─────────────────────────┐
│  Load Configuration     │
└───────────┬─────────────┘
            │
            ↓
┌─────────────────────────┐
│  Connect to Database    │
│  (SQLite or D1)         │
└───────────┬─────────────┘
            │
            ↓
┌─────────────────────────┐
│  MigrateFromFiles()     │
│  ┌───────────────────┐  │
│  │ Read *.sql files  │  │
│  │ Sort by name      │  │
│  │ Execute each      │  │
│  └───────────────────┘  │
└───────────┬─────────────┘
            │
            ↓
      ┌─────────┐
      │ Success?│
      └────┬────┘
           │
    ┌──────┴──────┐
    │             │
   YES            NO
    │             │
    ↓             ↓
┌─────────┐   ┌───────┐
│  Start  │   │ Crash │
│  HTTP   │   │  App  │
│ Server  │   └───────┘
└─────────┘
```

## Summary

| Scenario | Migrations Run? | Manual Required? |
|----------|----------------|------------------|
| `docker-compose up` | ✅ Automatically | ❌ No |
| `go run cmd/api/main.go` | ✅ Automatically | ❌ No |
| Production deployment | ✅ Automatically | ❌ No |
| First-time setup | ✅ Automatically | ❌ No (optional for testing) |
| Adding new migration | ✅ Automatically on next startup | ❌ No (optional for testing) |

## Best Practices

1. **Let the app handle migrations** - Don't overcomplicate with manual steps
2. **Test migrations locally** - Use `wrangler d1 execute --local` to test before committing
3. **Use `IF NOT EXISTS`** - Always make migrations idempotent
4. **Version your migrations** - Use numbered prefixes (001, 002, 003...)
5. **Never modify old migrations** - Create new ones instead
6. **Keep migrations small** - One logical change per file

## Additional Resources

- [Docker D1 Setup Guide](DOCKER_D1_SETUP.md)
- [Migrations README](../migrations/README.md)
- [GitHub Secrets Setup](GITHUB_SECRETS_SETUP.md)
