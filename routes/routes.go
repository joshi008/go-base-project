package routes

import (
	"database/sql"
	"net/http"

	"go-base-project/controllers"
	"go-base-project/services"
	"go-base-project/utils"
)

// SetupRoutes configures all application routes
func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Initialize controllers and services
	healthController := controllers.NewHealthController(db)
	
	// Expense service and controller
	expenseService := services.NewExpenseService(db)
	expenseController := controllers.NewExpenseController(expenseService)

	// Health check endpoint
	mux.HandleFunc("/health", healthController.HealthCheck)
	
	// Expense endpoints
	mux.HandleFunc("/expenses", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			expenseController.GetExpenses(w, r)
		case http.MethodPost:
			expenseController.CreateExpense(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// User endpoints
	userController := controllers.NewUserController(db)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userController.GetUsers(w, r)
		case http.MethodPost:
			userController.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

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
