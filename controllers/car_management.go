package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go-base-project/models"
	"go-base-project/services"
)

// CarController handles car-related requests
type CarController struct {
	DB *sql.DB
}

func NewCarController(db *sql.DB) *CarController {
	return &CarController{DB: db}
}

// Creation for branches, cars, pricing
func (cc *CarController) CreateBranch(w http.ResponseWriter, r *http.Request) {
	var branch models.Branches
	if err := json.NewDecoder(r.Body).Decode(&branch); err != nil {
		fmt.Println("Error decoding branch:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Insert branch into the database
	query := "INSERT INTO branches (name) VALUES ($1) RETURNING id"
	err := cc.DB.QueryRow(query, branch.Name).Scan(&branch.ID)
	if err != nil {
		http.Error(w, "Failed to create branch", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(branch)
}

func (cc *CarController) CreateCar(w http.ResponseWriter, r *http.Request) {
	var car models.Cars
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received car creation request: %+v\n", car)

	// Insert car into the database
	query := "INSERT INTO cars (id, category_id, branch_id) VALUES ($1, $2, $3) RETURNING id"
	err := cc.DB.QueryRow(query, car.ID, car.CategoryId, car.BranchID).Scan(&car.ID)
	if err != nil {
		fmt.Println("Error inserting car into database:", err)
		http.Error(w, "Failed to create car", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func (cc *CarController) CreatePricing(w http.ResponseWriter, r *http.Request) {
	var pricing models.Pricing
	if err := json.NewDecoder(r.Body).Decode(&pricing); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Insert pricing into the database
	query := "INSERT INTO pricing (price, branch_id, category_id) VALUES ($1, $2, $3) RETURNING id"
	err := cc.DB.QueryRow(query, pricing.Price, pricing.BranchID, pricing.CategoryId).Scan(&pricing.ID)
	if err != nil {
		http.Error(w, "Failed to create pricing", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pricing)
}

// Updation for pricing
func (cc *CarController) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	var pricing models.Pricing
	if err := json.NewDecoder(r.Body).Decode(&pricing); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update pricing in the database
	query := "UPDATE pricing SET price = $1 WHERE id = $2"
	_, err := cc.DB.Exec(query, pricing.Price, pricing.ID)
	if err != nil {
		http.Error(w, "Failed to update pricing", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pricing)
}

// /get-car - CategoryId, StartTime, EndTime -> CarId
func (cc *CarController) GetCar(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	categoryId := r.URL.Query().Get("categoryId")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")

	// Validate required parameters
	if categoryId == "" || startTime == "" || endTime == "" {
		http.Error(w, "Missing required parameters: categoryId, startTime, endTime", http.StatusBadRequest)
		return
	}

	// Convert categoryId to int
	intCategoryId, err := strconv.Atoi(categoryId)
	if err != nil {
		http.Error(w, "Invalid categoryId format", http.StatusBadRequest)
		return
	}

	// CreateBooking
	bookingService := services.NewBookingService(cc.DB)
	carId, err := bookingService.CreateBooking(intCategoryId, startTime, endTime)

	if err != nil {
		fmt.Printf("Booking error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Car booked successfully", 
		"carId": carId,
		"categoryId": intCategoryId,
		"startTime": startTime,
		"endTime": endTime,
	})
}
