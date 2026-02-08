package routes

import (
	"database/sql"
	"net/http"

	"go-base-project/controllers"
	"go-base-project/utils"
)

// SetupRoutes configures all application routes
func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Initialize controllers
	healthController := controllers.NewHealthController(db)

	// Health check endpoint
	mux.HandleFunc("/health", healthController.HealthCheck)

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Welcome to Go Base Project API"}`))
	})

	// Wrap with logging middleware
	return utils.LoggingMiddleware(mux)
}
