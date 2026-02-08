package strategy

import (
	"go-base-project/models"
)

// ExpenseStrategy defines how we validate and calculate splits
type ExpenseStrategy interface {
	Validate(expense *models.Expense) error
	CalculateSplits(expense *models.Expense) ([]*models.Split, error)
}
