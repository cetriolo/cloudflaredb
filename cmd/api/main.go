package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloudflaredb/internal/config"
	"cloudflaredb/internal/database"
	"cloudflaredb/internal/handlers"
	"cloudflaredb/internal/repository"
)

func main() {
	// === CONFIGURATION ===
	// Load environment-specific configuration (development/production)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting application in %s mode", cfg.Environment)
	log.Printf("Using database driver: %s", cfg.DatabaseDriver)

	// === DATABASE SETUP ===
	// Initialize database connection with appropriate driver:
	// - SQLite for local development
	// - Cloudflare D1 for production deployment
	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Run database migrations to ensure schema is up to date
	ctx := context.Background()
	if err := db.MigrateFromFiles(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed")

	// === DEPENDENCY INJECTION ===
	// Initialize repositories (data access layer)
	userRepo := repository.NewUserRepository(db.DB)
	roomRepo := repository.NewRoomRepository(db.DB)

	// Initialize HTTP handlers (presentation layer)
	userHandler := handlers.NewUserHandler(userRepo)
	roomHandler := handlers.NewRoomHandler(roomRepo)

	// === HTTP ROUTING SETUP ===
	mux := http.NewServeMux()

	// Serve static files (API testing page)
	// Accessible at the root path for browser-based testing
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	// Health check endpoint for monitoring and load balancers
	// Returns 200 OK with a simple JSON status message
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			log.Printf("Failed to write health check response: %v", err)
		}
	})

	// === USER ENDPOINTS ===
	// /users - Collection endpoint
	// GET: List all users (with pagination)
	// POST: Create a new user
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /users/{id} - Individual user endpoints and sub-resources
	// Handles multiple route patterns:
	//   - /users/{id} - GET/PUT/DELETE for individual user operations
	//   - /users/{id}/rooms - GET to list rooms assigned to a user
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		// Reject empty ID (path like "/users/")
		if r.URL.Path == "/users/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Parse the URL path to determine the endpoint type
		path := strings.TrimPrefix(r.URL.Path, "/users/")
		parts := strings.Split(path, "/")

		// Handle sub-resource: /users/{id}/rooms
		// Returns all rooms that the user is assigned to
		if len(parts) >= 2 && parts[1] == "rooms" {
			roomHandler.GetUserRooms(w, r)
			return
		}

		// Reject invalid nested paths
		if strings.Contains(path, "/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Handle individual user operations: /users/{id}
		// GET: Retrieve user details
		// PUT: Update user information
		// DELETE: Remove user from database
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// === ROOM ENDPOINTS ===
	// /rooms - Collection endpoint
	// GET: List all rooms (with pagination)
	// POST: Create a new room
	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			roomHandler.ListRooms(w, r)
		case http.MethodPost:
			roomHandler.CreateRoom(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /rooms/{id} - Individual room endpoints and sub-resources
	// Handles multiple route patterns:
	//   - /rooms/{id} - GET/PUT/DELETE for individual room operations
	//   - /rooms/{id}/users - GET to list users in a room, POST to assign user to room
	//   - /rooms/{id}/users/{userId} - DELETE to remove user from room
	mux.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		// Reject empty ID (path like "/rooms/")
		if r.URL.Path == "/rooms/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Parse the URL path to determine the endpoint type
		path := strings.TrimPrefix(r.URL.Path, "/rooms/")
		parts := strings.Split(path, "/")

		// Handle sub-resources: /rooms/{id}/users and /rooms/{id}/users/{userId}
		// This implements the many-to-many relationship between rooms and users
		if len(parts) >= 2 && parts[1] == "users" {
			if len(parts) == 2 {
				// /rooms/{id}/users - Collection operations
				// GET: Retrieve all users assigned to this room
				// POST: Assign a user to this room (requires user_id in body)
				switch r.Method {
				case http.MethodGet:
					roomHandler.GetRoomUsers(w, r)
				case http.MethodPost:
					roomHandler.AssignUserToRoom(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if len(parts) == 3 {
				// /rooms/{id}/users/{userId} - Individual relationship operations
				// DELETE: Remove a specific user from this room
				if r.Method == http.MethodDelete {
					roomHandler.RemoveUserFromRoom(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else {
				// Invalid path depth
				http.Error(w, "Not found", http.StatusNotFound)
			}
			return
		}

		// Reject invalid nested paths
		if strings.Contains(path, "/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Handle individual room operations: /rooms/{id}
		// GET: Retrieve room details
		// PUT: Update room information
		// DELETE: Remove room from database
		switch r.Method {
		case http.MethodGet:
			roomHandler.GetRoom(w, r)
		case http.MethodPut:
			roomHandler.UpdateRoom(w, r)
		case http.MethodDelete:
			roomHandler.DeleteRoom(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// === SERVER CONFIGURATION ===
	// Create HTTP server with production-ready timeouts:
	// - ReadTimeout: prevents slow-client attacks
	// - WriteTimeout: prevents slow-write attacks
	// - IdleTimeout: closes idle keep-alive connections
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      loggingMiddleware(mux), // Wrap router with logging middleware
		ReadTimeout:  15 * time.Second,       // Maximum time to read request
		WriteTimeout: 15 * time.Second,       // Maximum time to write response
		IdleTimeout:  60 * time.Second,       // Maximum idle time for keep-alive
	}

	// === SERVER STARTUP ===
	// Start server in a goroutine to allow graceful shutdown handling
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("API Tester available at: http://localhost:%s", cfg.Port)
		log.Printf("API endpoints available at: http://localhost:%s/users", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// === GRACEFUL SHUTDOWN ===
	// Listen for termination signals (SIGINT from Ctrl+C, SIGTERM from orchestrators)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until signal received

	log.Println("Server is shutting down...")

	// Gracefully shutdown the server with a timeout
	// Allows in-flight requests to complete before shutting down
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// loggingMiddleware is an HTTP middleware that logs request information.
// Logs: HTTP method, request URI, and processing duration.
// This helps with debugging and monitoring API usage.
// Example log: "GET /users/123 1.234ms"
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
