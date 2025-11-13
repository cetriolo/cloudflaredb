# CloudflareDB API

A production-ready Go REST API application using Cloudflare D1 database with SQLite compatibility for local development.

## Features

- RESTful API for user management
- Cloudflare D1 integration for production
- SQLite for local development
- Clean architecture with repository pattern
- Comprehensive unit and integration tests
- Docker support
- GitHub Actions CI/CD pipeline
- Graceful shutdown
- Health check endpoint

## Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose (optional, for containerized development)
- Cloudflare account with D1 database (for production)

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   ├── database.go          # Database connection and migrations
│   │   └── database_test.go     # Database tests
│   ├── handlers/
│   │   ├── user_handler.go      # HTTP handlers
│   │   └── user_handler_test.go # Handler tests
│   ├── models/
│   │   └── user.go              # Data models
│   └── repository/
│       ├── user_repository.go      # Data access layer
│       └── user_repository_test.go # Repository tests
├── .github/
│   └── workflows/
│       └── ci.yml               # GitHub Actions workflow
├── docker-compose.yml           # Docker Compose configuration
├── Dockerfile                   # Docker image definition
├── Makefile                     # Build and development commands
├── .env.example                 # Environment variables template
└── README.md                    # This file
```

## Quick Start

### Local Development

1. **Clone the repository**

```bash
git clone <your-repo-url>
cd cloudflaredb
```

2. **Install dependencies**

```bash
go mod download
```

3. **Set up environment variables**

```bash
cp .env.example .env
```

Edit `.env` to configure your local settings (default SQLite configuration works out of the box).

4. **Run the application**

```bash
go run cmd/api/main.go
```

Or use the Makefile:

```bash
make run
```

The API will be available at `http://localhost:8080`.

### Using Docker

1. **Build and run with Docker Compose**

```bash
docker-compose up -d
```

2. **View logs**

```bash
docker-compose logs -f
```

3. **Stop the application**

```bash
docker-compose down
```

## Configuration

The application is configured via environment variables. See `.env.example` for all available options.

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENVIRONMENT` | Application environment | `development` | No |
| `PORT` | HTTP server port | `8080` | No |
| `DATABASE_DRIVER` | Database driver (`sqlite3` or `cfd1`) | `sqlite3` | No |
| `DATABASE_DSN` | Database connection string (for SQLite) | `./local.db` | No |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID | - | Yes (for D1) |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token | - | Yes (for D1) |
| `CLOUDFLARE_DB_NAME` | Cloudflare D1 database name | - | Yes (for D1) |

### Local Development (SQLite)

```env
DATABASE_DRIVER=sqlite3
DATABASE_DSN=./local.db
```

### Production (Cloudflare D1)

```env
DATABASE_DRIVER=cfd1
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_DB_NAME=your_database_name
```

## API Endpoints

### Health Check

```
GET /health
```

Returns the health status of the API.

**Response:**
```json
{
  "status": "healthy"
}
```

### User Management

#### Create User

```
POST /users
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

#### Get User

```
GET /users/{id}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

#### List Users

```
GET /users?limit=10&offset=0
```

**Query Parameters:**
- `limit` (optional): Number of users to return (default: 10)
- `offset` (optional): Number of users to skip (default: 0)

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  }
]
```

#### Update User

```
PUT /users/{id}
Content-Type: application/json

{
  "email": "newemail@example.com",
  "name": "Jane Doe"
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "email": "newemail@example.com",
  "name": "Jane Doe",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:30:00Z"
}
```

#### Delete User

```
DELETE /users/{id}
```

**Response:** `204 No Content`

## Testing

### Run all tests

```bash
make test
```

Or with coverage:

```bash
make test-coverage
```

This will generate a coverage report in `coverage.html`.

### Run specific tests

```bash
# Test a specific package
go test -v ./internal/repository

# Test a specific function
go test -v -run TestUserRepository_Create ./internal/repository
```

### Test with Docker

The tests use an in-memory SQLite database, so they don't require any external dependencies.

## Database Setup

### Local Development with SQLite

No setup required! The application will automatically create a SQLite database file (`local.db`) when it starts.

### Production with Cloudflare D1

1. **Create a D1 database**

```bash
npx wrangler d1 create cloudflaredb
```

2. **Get your database details**

The command will output your database ID. Note this down along with your Cloudflare account ID.

3. **Create an API token**

- Go to [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
- Click "Create Token"
- Use the "Edit Cloudflare Workers" template
- Add D1 edit permissions
- Copy the generated token

4. **Set environment variables**

```env
DATABASE_DRIVER=cfd1
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_DB_NAME=cloudflaredb
```

5. **Run migrations**

The application automatically runs migrations on startup. The database schema is created in `internal/database/database.go`.

### Manual Migrations

If you need to run migrations manually using Wrangler:

```bash
# Create a migration file
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

## Development

### Available Make Commands

```bash
make help          # Display all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make test-coverage # Run tests with coverage report
make clean         # Clean build artifacts
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make docker-down   # Stop Docker Compose
make lint          # Run linter
make deps          # Download and tidy dependencies
```

### Code Structure Best Practices

This project follows Go best practices:

- **Clean Architecture**: Separation of concerns with clear boundaries
- **Repository Pattern**: Data access abstraction
- **Dependency Injection**: Loosely coupled components
- **Error Handling**: Proper error wrapping and context
- **Testing**: Unit and integration tests with >80% coverage
- **Configuration**: Environment-based configuration
- **Logging**: Structured logging with request tracking

## CI/CD

The project includes a GitHub Actions workflow that:

1. Runs tests with race detection
2. Runs linting
3. Builds the application
4. Builds Docker image
5. Uploads coverage reports to Codecov

The workflow runs on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

## Best Practices Implemented

### Database

- Connection pooling configuration
- Query timeouts
- Prepared statements for SQL injection prevention
- Database indexes for performance
- Migrations on startup

### Security

- No sensitive data in logs
- Environment variable configuration
- SQL injection prevention via parameterized queries
- CORS-ready (add middleware as needed)

### Performance

- Efficient database queries with indexes
- Connection pooling
- Graceful shutdown
- Request timeouts

### Reliability

- Health check endpoint
- Graceful shutdown
- Retry logic ready (add as needed)
- Comprehensive error handling

## Troubleshooting

### SQLite locked database

If you encounter "database is locked" errors:

```bash
# Stop all running instances
pkill -f "go run cmd/api/main.go"

# Remove lock files
rm -f *.db-shm *.db-wal
```

### Tests failing

Make sure CGO is enabled (required for SQLite):

```bash
export CGO_ENABLED=1
go test ./...
```

### Docker build fails

Ensure you have the latest Docker version and enough disk space:

```bash
docker system prune -a
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Support

For issues and questions, please open an issue on GitHub.
