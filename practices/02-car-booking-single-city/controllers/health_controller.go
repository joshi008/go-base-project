package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// HealthController handles health check requests
type HealthController struct {
	DB *sql.DB
}

// NewHealthController creates a new HealthController
func NewHealthController(db *sql.DB) *HealthController {
	return &HealthController{DB: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Message  string `json:"message"`
}

// HealthCheck handles GET /health endpoint
func (hc *HealthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	dbStatus := "connected"
	if err := hc.DB.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	response := HealthResponse{
		Status:   "ok",
		Database: dbStatus,
		Message:  "Server is running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
