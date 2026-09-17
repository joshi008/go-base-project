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

	// Creation POST request for branches, cars, pricing
	carController := controllers.NewCarController(db) // Pass nil for strategy for now
	mux.HandleFunc("/branches", carController.CreateBranch)
	mux.HandleFunc("/cars", carController.CreateCar)
	mux.HandleFunc("/pricing", carController.CreatePricing)

	// Updation for pricing
	mux.HandleFunc("/pricing-update", carController.UpdatePricing)

	// Get Car for booking
	mux.HandleFunc("/booking", carController.GetCar)

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
