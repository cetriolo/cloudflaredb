# Docker + Cloudflare D1 Setup Guide

This guide explains how to run the application in Docker while connecting to your Cloudflare D1 database.

## Prerequisites

- Docker and Docker Compose installed
- Cloudflare account with D1 database created
- Cloudflare API token with D1 permissions

## Quick Start

### 1. Create Your Environment File

Copy the example file and fill in your Cloudflare credentials:

```bash
cp .env.example .env
```

Edit `.env` and set your credentials:

```bash
# Required Cloudflare D1 credentials
CLOUDFLARE_ACCOUNT_ID=your_actual_account_id
CLOUDFLARE_API_TOKEN=your_actual_api_token
CLOUDFLARE_DB_NAME=cloudflaredb
```

### 2. Start the Application (Migrations Run Automatically!)

**Good news:** Migrations run automatically when the app starts! Just run:

```bash
docker-compose up -d
```

The application will:
1. Connect to your D1 database
2. Read all `.sql` files from `/migrations` folder
3. Execute them in order (001, 002, etc.)
4. Start serving requests

You'll see this in the logs:
```
Running database migrations from files...
Applying migration: 001_create_users_table.sql
Successfully applied migration: 001_create_users_table.sql
Applying migration: 002_create_rooms_table.sql
Successfully applied migration: 002_create_rooms_table.sql
All migrations completed successfully
```

### 3. Verify It's Working

```bash
# View logs to see migrations running
docker-compose logs -f

# Test the API
curl http://localhost:8080/health
curl http://localhost:8080/users
```

### 4. Stop the Application

```bash
docker-compose down
```

## Configuration Details

### Environment Variables

The docker-compose setup uses these environment variables from your `.env` file:

| Variable | Description | Example |
|----------|-------------|---------|
| `CLOUDFLARE_ACCOUNT_ID` | Your Cloudflare account ID | `a1b2c3d4e5f6...` |
| `CLOUDFLARE_API_TOKEN` | API token with D1 edit permissions | `xyz123abc...` |
| `CLOUDFLARE_DB_NAME` | Name of your D1 database | `cloudflaredb` |
| `ENVIRONMENT` | Application environment | `production` |
| `PORT` | Port to run the app on | `8080` |

### Database Connection

The application connects to D1 using the `cfd1` driver with this DSN format:

```
d1://{CLOUDFLARE_ACCOUNT_ID}:{CLOUDFLARE_API_TOKEN}@{CLOUDFLARE_DB_NAME}
```

Docker Compose automatically constructs this DSN from your environment variables.

## How Migrations Work in Docker

### Automatic Migration Execution

**✅ Migrations run automatically on application startup!**

Here's how it works:

1. **Migration Files Location**: All SQL migrations are in the `/migrations` folder
2. **Docker Copies Them**: The Dockerfile copies `/migrations` into the container at `/root/migrations`
3. **Application Reads Them**: On startup, the app automatically:
   - Reads all `.sql` files from `/migrations` folder
   - Sorts them by filename (001, 002, 003...)
   - Executes them in order against your D1 database
4. **Idempotent Execution**: Migrations use `IF NOT EXISTS` clauses, so they're safe to run multiple times

### What Happens on Startup

```
Application Start
    ↓
Connect to D1 database
    ↓
Read migrations/*.sql files
    ↓
Execute each migration in order
    ↓
All migrations applied successfully
    ↓
Start HTTP server
```

You'll see these logs:
```
2025/11/14 11:43:32 Running database migrations from files...
2025/11/14 11:43:32 Applying migration: 001_create_users_table.sql
2025/11/14 11:43:32 Successfully applied migration: 001_create_users_table.sql
2025/11/14 11:43:32 Applying migration: 002_create_rooms_table.sql
2025/11/14 11:43:33 Successfully applied migration: 002_create_rooms_table.sql
2025/11/14 11:43:33 All migrations completed successfully
```

### Optional: Manual Migration Pre-Run

You can **optionally** run migrations manually using `wrangler` before deploying:

**When might you want this?**
- Testing migrations on a separate database first
- Running migrations in a CI/CD pipeline before deployment
- Separating migration execution from app deployment
- Debugging migration issues

**How?**

```bash
# Test locally first (optional)
npx wrangler d1 execute cloudflaredb --local --file=migrations/001_create_users_table.sql

# Run on production (optional - app will do this automatically)
npx wrangler d1 execute cloudflaredb --remote --file=migrations/001_create_users_table.sql
```

**But this is NOT required!** The app handles it automatically.

## Getting Your Cloudflare Credentials

### 1. Account ID

**Option A - From Dashboard URL:**
1. Go to https://dash.cloudflare.com/
2. Look at the URL: `https://dash.cloudflare.com/{account-id}/...`
3. Copy the account ID from the URL

**Option B - From Sidebar:**
1. Go to https://dash.cloudflare.com/
2. Look at the right sidebar under your account name
3. Copy the Account ID shown there

### 2. API Token

1. Go to https://dash.cloudflare.com/profile/api-tokens
2. Click **"Create Token"**
3. Choose **"Edit Cloudflare Workers"** template (includes D1 permissions)

   Or create a custom token with these permissions:
   - Account > D1 > Edit

4. Click **"Continue to summary"**
5. Click **"Create Token"**
6. **Copy the token immediately** (you can only see it once!)

### 3. Database Name

Find your D1 database name:

```bash
npx wrangler d1 list
```

Output:
```
┌──────────────────────────────────┬──────────────┬─────────┐
│ UUID                             │ Name         │ Created │
├──────────────────────────────────┼──────────────┼─────────┤
│ a1b2c3d4-e5f6-7890-abcd-ef123456 │ cloudflaredb │ 2024-01 │
└──────────────────────────────────┴──────────────┴─────────┘
```

Use either the **Name** (`cloudflaredb`) or **UUID** in your `.env` file.

## Development Workflows

### Using D1 (Production-like)

This is the default configuration. Connect to your real Cloudflare D1 database:

```bash
# Make sure .env has D1 credentials
docker-compose up -d
```

**Best for:**
- Testing production setup locally
- Verifying D1 integration
- Sharing data across team members

### Using SQLite (Local Development)

For completely local development without internet, use the local profile:

```bash
# Start with local SQLite
docker-compose --profile local up -d

# This uses SQLite at /data/local.db inside the container
# and mounts ./data on your host machine
```

**Best for:**
- Offline development
- Fast iteration
- Testing migrations locally

## Troubleshooting

### Migration Errors

**Problem**: Migrations fail with "table already exists"

**Solution**: This shouldn't happen! All migrations use `IF NOT EXISTS` clauses, making them idempotent (safe to run multiple times).

If you see this error:
1. Check the migration file has `CREATE TABLE IF NOT EXISTS`
2. Review the application logs for the actual error
3. Verify your D1 credentials are correct

To verify your schema:
```bash
npx wrangler d1 execute cloudflaredb --remote --command "SELECT name FROM sqlite_master WHERE type='table';"
```

**Problem**: "failed to read migrations directory"

**Solution**: The migrations folder is missing from the Docker image. Rebuild:
```bash
docker-compose build --no-cache
docker-compose up -d
```

### Authentication Errors

**Problem**: Error like "unauthorized" or "invalid credentials"

**Solutions**:
1. Verify your `CLOUDFLARE_API_TOKEN` is correct
2. Make sure the token has D1 permissions
3. Check that `CLOUDFLARE_ACCOUNT_ID` matches your account
4. Recreate the API token if needed

### Connection Errors

**Problem**: Can't connect to D1

**Solutions**:
1. Check your internet connection (D1 is accessed via HTTPS)
2. Verify the `CLOUDFLARE_DB_NAME` is correct
3. Try listing databases: `npx wrangler d1 list`
4. Check Docker container logs: `docker-compose logs`


## Advanced Usage

### Custom Port

Change the port in your `.env`:

```bash
PORT=3000
```

Then update docker-compose port mapping:

```yaml
ports:
  - "3000:3000"  # host:container
```

### Multiple Environments

Create separate environment files:

```bash
# Production
.env.production
CLOUDFLARE_ACCOUNT_ID=prod_account_id
CLOUDFLARE_API_TOKEN=prod_token
CLOUDFLARE_DB_NAME=cloudflaredb-prod

# Staging
.env.staging
CLOUDFLARE_ACCOUNT_ID=staging_account_id
CLOUDFLARE_API_TOKEN=staging_token
CLOUDFLARE_DB_NAME=cloudflaredb-staging
```

Use with:

```bash
# Run with production config
docker-compose --env-file .env.production up -d

# Run with staging config
docker-compose --env-file .env.staging up -d
```

### Viewing Container Logs

```bash
# Follow all logs
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100

# Only show errors
docker-compose logs | grep -i error
```

### Accessing the Container

```bash
# Open a shell in the running container
docker-compose exec app sh

# Check if migrations folder exists
docker-compose exec app ls -la /root/migrations

# Manually test database connection
docker-compose exec app wget -O- http://localhost:8080/health
```

### Health Checks

The container has a built-in health check that pings `/health` every 30 seconds.

View health status:

```bash
docker-compose ps
```

Output shows health status:
```
NAME    COMMAND    STATUS                    PORTS
app     ./api      Up 2 minutes (healthy)    0.0.0.0:8080->8080/tcp
```

## Production Deployment

### Docker Image for Production

Build and tag for production:

```bash
# Build
docker build -t cloudflaredb:latest .

# Tag for registry
docker tag cloudflaredb:latest your-registry.com/cloudflaredb:latest

# Push to registry
docker push your-registry.com/cloudflaredb:latest
```

### Security Best Practices

1. **Never commit `.env` file**
   - Already in `.gitignore`
   - Use secrets management in production

2. **Rotate API tokens regularly**
   - Create new tokens periodically
   - Delete old tokens after updating

3. **Use read-only tokens when possible**
   - If your app only reads data, use a read-only token
   - Only use edit permissions when needed

4. **Limit token scope**
   - Only grant D1 permissions, nothing else
   - Create separate tokens for different environments

## Additional Resources

- [Cloudflare D1 Documentation](https://developers.cloudflare.com/d1/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Wrangler CLI Documentation](https://developers.cloudflare.com/workers/wrangler/)
- [Migration Guide](../migrations/README.md)

## Need Help?

If you encounter issues:

1. Check the [Troubleshooting](#troubleshooting) section above
2. Review Docker container logs: `docker-compose logs`
3. Verify D1 credentials: `npx wrangler d1 list`
4. Test migrations locally first: `npx wrangler d1 execute cloudflaredb --local --file=migrations/001_create_users_table.sql`
