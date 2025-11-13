# Troubleshooting Guide

Common issues and their solutions for the CloudflareDB API project.

## Build and Compilation Issues

### Go Version Mismatch Error

**Error:**
```
compile: version "go1.24.2" does not match go tool version "go1.25.1"
```

**Cause:** Cached compiled packages from a previous Go version are conflicting with the current Go toolchain.

**Solution:**

1. **Clear all Go caches:**
```bash
go clean -cache
go clean -modcache  # This will require re-downloading dependencies
go clean -testcache
```

2. **Remove cached build artifacts:**
```bash
rm -rf ~/.cache/go-build
rm -rf $GOPATH/pkg  # Caution: requires re-downloading all modules
```

3. **Force rebuild:**
```bash
CGO_ENABLED=1 go build -a -o bin/api cmd/api/main.go
```

4. **If the issue persists, reinstall Go:**
```bash
# Download and install the latest Go version from https://golang.org/dl/
# Then verify:
go version
```

### CGO_ENABLED Required

**Error:**
```
package github.com/mattn/go-sqlite3: build constraints exclude all Go files
```

**Cause:** SQLite driver requires CGO to be enabled.

**Solution:**
```bash
export CGO_ENABLED=1
go build ./...
```

Or for a single command:
```bash
CGO_ENABLED=1 go build -o bin/api cmd/api/main.go
```

### Missing Compiler (gcc)

**Error:**
```
gcc: command not found
```

**Cause:** SQLite driver requires a C compiler.

**Solution:**

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install build-essential
```

**macOS:**
```bash
xcode-select --install
```

**Alpine (Docker):**
```bash
apk add --no-cache gcc musl-dev
```

## Runtime Issues

### Database Locked Error

**Error:**
```
database is locked
```

**Cause:** Multiple processes are trying to access the SQLite database simultaneously.

**Solution:**

1. **Stop all running instances:**
```bash
pkill -f "go run cmd/api/main.go"
pkill -f "bin/api"
```

2. **Remove WAL files:**
```bash
rm -f *.db-shm *.db-wal
```

3. **Enable WAL mode for better concurrency:**
```sql
PRAGMA journal_mode=WAL;
```

### Port Already in Use

**Error:**
```
bind: address already in use
```

**Cause:** Another process is using port 8080.

**Solution:**

1. **Find the process:**
```bash
lsof -i :8080
# or
netstat -tlnp | grep 8080
```

2. **Kill the process:**
```bash
kill -9 <PID>
```

3. **Or use a different port:**
```bash
PORT=8081 go run cmd/api/main.go
```

## Cloudflare D1 Issues

### Authentication Failed

**Error:**
```
failed to connect to database: authentication failed
```

**Cause:** Invalid Cloudflare credentials.

**Solution:**

1. **Verify environment variables:**
```bash
echo $CLOUDFLARE_ACCOUNT_ID
echo $CLOUDFLARE_API_TOKEN
echo $CLOUDFLARE_DB_NAME
```

2. **Check token permissions:**
- Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
- Verify the token has D1 edit permissions
- Regenerate token if necessary

3. **Test with Wrangler:**
```bash
npx wrangler d1 list
```

### Database Not Found

**Error:**
```
database not found
```

**Cause:** Incorrect database name or database doesn't exist.

**Solution:**

1. **List available databases:**
```bash
npx wrangler d1 list
```

2. **Verify database name in `.env`:**
```env
CLOUDFLARE_DB_NAME=cloudflaredb
```

3. **Create database if it doesn't exist:**
```bash
npx wrangler d1 create cloudflaredb
```

### Rate Limiting

**Error:**
```
too many requests
```

**Cause:** Cloudflare API rate limit exceeded.

**Solution:**

1. **Implement retry logic** (already included in the app)
2. **Add exponential backoff**
3. **Cache frequently accessed data**
4. **Use read replicas for read-heavy workloads**

## Testing Issues

### Tests Fail to Run

**Error:**
```
no test files
```

**Cause:** Test files not found or not properly named.

**Solution:**

1. **Verify test files exist:**
```bash
find . -name "*_test.go"
```

2. **Run tests with verbose output:**
```bash
go test -v ./...
```

### Tests Pass Locally but Fail in CI

**Cause:** Environment differences between local and CI.

**Solution:**

1. **Check Go version in CI:**
```yaml
# .github/workflows/ci.yml
- uses: actions/setup-go@v5
  with:
    go-version: '1.25'  # Match local version
```

2. **Ensure CGO is enabled in CI:**
```yaml
env:
  CGO_ENABLED: 1
```

3. **Install build dependencies:**
```yaml
- name: Install dependencies
  run: |
    sudo apt-get update
    sudo apt-get install -y build-essential
```

## Docker Issues

### Docker Build Fails

**Error:**
```
error building image
```

**Cause:** Missing dependencies or incorrect Dockerfile.

**Solution:**

1. **Check Docker version:**
```bash
docker version
```

2. **Rebuild without cache:**
```bash
docker build --no-cache -t cloudflaredb-api .
```

3. **Check build logs:**
```bash
docker build -t cloudflaredb-api . 2>&1 | tee build.log
```

### Container Exits Immediately

**Cause:** Application crashes on startup.

**Solution:**

1. **Check container logs:**
```bash
docker logs <container-id>
```

2. **Run interactively:**
```bash
docker run -it cloudflaredb-api /bin/sh
```

3. **Check environment variables:**
```bash
docker run -e DATABASE_DRIVER=sqlite3 cloudflaredb-api
```

## Dependency Issues

### Module Not Found

**Error:**
```
module not found
```

**Cause:** Dependencies not downloaded.

**Solution:**
```bash
go mod download
go mod tidy
```

### Version Conflicts

**Error:**
```
module requires Go >= X.Y
```

**Cause:** Go version too old.

**Solution:**

1. **Update Go:**
```bash
# Download from https://golang.org/dl/
```

2. **Or update go.mod:**
```go
go 1.25  // Match your installed version
```

## Performance Issues

### Slow Query Performance

**Cause:** Missing indexes or inefficient queries.

**Solution:**

1. **Add indexes:**
```sql
CREATE INDEX idx_users_name ON users(name);
```

2. **Analyze queries:**
```sql
EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = ?;
```

3. **Use prepared statements** (already implemented)

### High Memory Usage

**Cause:** Connection pool misconfiguration.

**Solution:**

Adjust connection pool settings in `internal/database/database.go`:
```go
db.SetMaxOpenConns(10)  // Reduce if memory is limited
db.SetMaxIdleConns(2)
```

## Environment Issues

### .env File Not Loaded

**Cause:** File doesn't exist or is in wrong location.

**Solution:**

1. **Create from example:**
```bash
cp .env.example .env
```

2. **Verify location:**
```bash
ls -la .env
```

3. **Check file permissions:**
```bash
chmod 644 .env
```

## Migration Issues

### Migration Already Applied

**Cause:** Attempting to run the same migration twice.

**Solution:**

The migrations use `IF NOT EXISTS`, so this shouldn't cause issues. If it does:

1. **Check database schema:**
```bash
sqlite3 local.db .schema
```

2. **Manually verify tables:**
```sql
SELECT name FROM sqlite_master WHERE type='table';
```

### Migration Fails

**Cause:** Invalid SQL or database permissions.

**Solution:**

1. **Check migration SQL:**
```bash
cat migrations/*.sql
```

2. **Test locally first:**
```bash
sqlite3 test.db < migrations/001_initial_schema.sql
```

3. **Check database permissions:**
```bash
ls -l *.db
chmod 644 *.db
```

## Getting More Help

### Enable Debug Logging

Add debug logging to `cmd/api/main.go`:
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Check Application Logs

```bash
# Running locally
go run cmd/api/main.go 2>&1 | tee app.log

# Docker
docker-compose logs -f
```

### Report Issues

If you encounter an issue not covered here:

1. **Check existing issues:** [GitHub Issues](https://github.com/your-repo/issues)
2. **Provide details:**
   - Go version (`go version`)
   - OS and version
   - Error messages
   - Steps to reproduce
3. **Create minimal reproduction case**

## Additional Resources

- [Go Official Docs](https://golang.org/doc/)
- [Cloudflare D1 Docs](https://developers.cloudflare.com/d1/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Docker Documentation](https://docs.docker.com/)
