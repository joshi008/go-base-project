package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Expense struct {
	ID          int       `json:"id"`
	TotalAmount float64   `json:"total_amount"`
	Description string    `json:"description"`
	Lender      *User     `json:"lender"`
	Splits      []*Split  `json:"splits"`
	CreatedAt   time.Time `json:"created_at"`
	ExpenseType string    `json:"expense_type"`
}

type Split struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`
	User   *User   `json:"user"`
	// If Type == PERCENT, this stores 40.0, 20.0, etc.
	// If Type == EQUAL, this can be empty (or 1.0).
	// If Type == EXACT, this equals Amount.
	Value float64 `json:"value"`
}
