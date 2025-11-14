# Build stage
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
# CGO is required for SQLite support (optional, for local dev)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite compatibility
# The binary supports both SQLite (local) and D1 (via HTTP API)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
# - ca-certificates: Required for HTTPS connections to Cloudflare D1 API
# - sqlite-libs: Optional, for local SQLite support
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/api .

# Copy web assets and migrations
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]
