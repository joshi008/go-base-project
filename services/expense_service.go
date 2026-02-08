package services

import (
	"database/sql"
	"fmt"

	"go-base-project/models"
	"go-base-project/services/strategy"
)

// ExpenseService defines the interface for expense operations
type ExpenseService interface {
	CreateExpense(expense *models.Expense, expenseType string) error
	GetExpenses() ([]*models.Expense, error)
	GetExpenseByID(id int) (*models.Expense, error)
}

// expenseServiceImpl implements ExpenseService
type expenseServiceImpl struct {
	db *sql.DB
}

// NewExpenseService creates a new expense service
func NewExpenseService(db *sql.DB) ExpenseService {
	return &expenseServiceImpl{db: db}
}

// CreateExpense validates, calculates splits, and saves expense to database
func (s *expenseServiceImpl) CreateExpense(expense *models.Expense, expenseType string) error {
	// Step 1: Create appropriate strategy based on expense type
	var expenseStrategy strategy.ExpenseStrategy
	switch expenseType {
	case "percentage":
		expenseStrategy = &strategy.PercentageStrategy{}
	case "exact":
		expenseStrategy = &strategy.ExactStrategy{}
	case "equal":
		expenseStrategy = &strategy.EqualStrategy{}
	default:
		return fmt.Errorf("invalid expense type: %s", expenseType)
	}

	expense.ExpenseType = expenseType

	// Step 2: Validate the expense
	// if err := expenseStrategy.Validate(expense); err != nil {
	// 	return fmt.Errorf("validation failed: %w", err)
	// }

	// Step 3: Calculate splits based on strategy
	splits, err := expenseStrategy.CalculateSplits(expense)
	if err != nil {
		return fmt.Errorf("calculation failed: %w", err)
	}
	expense.Splits = splits

	// Step 4: Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 5: Insert expense into database
	var expenseID int
	err = tx.QueryRow(
		"INSERT INTO expenses (description, total_amount, expense_type, lender_id) VALUES ($1, $2, $3, $4) RETURNING id",
		expense.Description, expense.TotalAmount, expenseType, expense.Lender.ID,
	).Scan(&expenseID)
	if err != nil {
		return fmt.Errorf("failed to insert expense: %w", err)
	}

	// Step 6: Insert splits into database
	for _, split := range expense.Splits {
		_, err := tx.Exec(
			"INSERT INTO splits (expense_id, user_id, amount, value) VALUES ($1, $2, $3, $4)",
			expenseID, split.User.ID, split.Amount, split.Value,
		)
		if err != nil {
			return fmt.Errorf("failed to insert split: %w", err)
		}
	}

	// Step 7: Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	expense.ID = expenseID
	return nil
}

// GetExpenses fetches all expenses from database
func (s *expenseServiceImpl) GetExpenses() ([]*models.Expense, error) {
	rows, err := s.db.Query(`
		SELECT id, description, total_amount, expense_type, lender_id 
		FROM expenses
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*models.Expense
	for rows.Next() {
		var expense models.Expense
		var lenderID int
		if err := rows.Scan(&expense.ID, &expense.Description, &expense.TotalAmount, &expense.ExpenseType, &lenderID); err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expense.Lender = &models.User{ID: lenderID}
		expenses = append(expenses, &expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expenses: %w", err)
	}

	return expenses, nil
}

// GetExpenseByID fetches a single expense by ID
func (s *expenseServiceImpl) GetExpenseByID(id int) (*models.Expense, error) {
	var expense models.Expense
	var lenderID int

	err := s.db.QueryRow(`
		SELECT id, description, total_amount, expense_type, lender_id 
		FROM expenses 
		WHERE id = $1
	`, id).Scan(&expense.ID, &expense.Description, &expense.TotalAmount, &expense.ExpenseType, &lenderID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}

	expense.Lender = &models.User{ID: lenderID}
	return &expense, nil
}
