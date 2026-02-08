package strategy

import (
	"fmt"
	"math"

	"go-base-project/models"
)

type ExactStrategy struct{}

func (s *ExactStrategy) Validate(expense *models.Expense) error {
	numSplits := len(expense.Splits)
	if numSplits == 0 {
		return fmt.Errorf("no splits provided")
	}

	sumAmount := 0.0
	for _, split := range expense.Splits {
		sumAmount += split.Amount
	}

	if math.Abs(sumAmount-expense.TotalAmount) >= 0.01 {
		return fmt.Errorf("total split amount does not match total expense amount")
	}

	return nil
}

func (s *ExactStrategy) CalculateSplits(expense *models.Expense) ([]*models.Split, error) {
	numSplits := len(expense.Splits)
	if numSplits == 0 {
		return nil, fmt.Errorf("no splits provided")
	}

	sumAmount := 0.0
	for _, split := range expense.Splits {
		sumAmount += split.Amount
	}

	if math.Abs(sumAmount-expense.TotalAmount) >= 0.01 {
		return nil, fmt.Errorf("total split amount does not match total expense amount")
	}

	return expense.Splits, nil
}
