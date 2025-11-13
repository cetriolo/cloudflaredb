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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting application in %s mode", cfg.Environment)
	log.Printf("Using database driver: %s", cfg.DatabaseDriver)

	// Initialize database
	db, err := database.New(cfg.DatabaseDriver, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Run migrations from files
	ctx := context.Background()
	if err := db.MigrateFromFiles(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	roomRepo := repository.NewRoomRepository(db.DB)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	roomHandler := handlers.NewRoomHandler(roomRepo)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Serve static files (API testing page)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// User endpoints
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

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/users/")
		parts := strings.Split(path, "/")

		// Check for /users/{id}/rooms
		if len(parts) >= 2 && parts[1] == "rooms" {
			roomHandler.GetUserRooms(w, r)
			return
		}

		// Regular user endpoints /users/{id}
		if strings.Contains(path, "/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

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

	// Room endpoints
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

	mux.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rooms/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Extract path parts
		path := strings.TrimPrefix(r.URL.Path, "/rooms/")
		parts := strings.Split(path, "/")

		// /rooms/{id}/users - Get users in room or assign user to room
		if len(parts) >= 2 && parts[1] == "users" {
			if len(parts) == 2 {
				// /rooms/{id}/users
				switch r.Method {
				case http.MethodGet:
					roomHandler.GetRoomUsers(w, r)
				case http.MethodPost:
					roomHandler.AssignUserToRoom(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if len(parts) == 3 {
				// /rooms/{id}/users/{userId}
				if r.Method == http.MethodDelete {
					roomHandler.RemoveUserFromRoom(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
			return
		}

		// Regular room endpoints /rooms/{id}
		if strings.Contains(path, "/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

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

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("API Tester available at: http://localhost:%s", cfg.Port)
		log.Printf("API endpoints available at: http://localhost:%s/users", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
