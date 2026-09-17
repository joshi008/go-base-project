package controllers

import (
	"encoding/json"
	"net/http"

	"go-base-project/models"
	"go-base-project/services"
)

type ExpenseController struct {
	service services.ExpenseService
}

func NewExpenseController(service services.ExpenseService) *ExpenseController {
	return &ExpenseController{
		service: service,
	}
}

// CreateExpense handles POST /expenses - creates a new expense with calculated splits
func (c *ExpenseController) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var expense models.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if expense.TotalAmount <= 0 {
		http.Error(w, "Total amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if expense.Lender == nil || expense.Lender.ID <= 0 {
		http.Error(w, "Lender ID is required", http.StatusBadRequest)
		return
	}

	if len(expense.Splits) == 0 {
		http.Error(w, "At least one split is required", http.StatusBadRequest)
		return
	}

	// Service handles: validation -> calculation -> database insertion
	if err := c.service.CreateExpense(&expense, expense.ExpenseType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

// GetExpenses handles GET /expenses - fetches all expenses
func (c *ExpenseController) GetExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := c.service.GetExpenses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}
