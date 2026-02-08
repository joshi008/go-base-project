package strategy

import (
	"fmt"
	"math"

	"go-base-project/models"
)

type EqualStrategy struct{}

func (s *EqualStrategy) Validate(expense *models.Expense) error {
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

func (s *EqualStrategy) CalculateSplits(expense *models.Expense) ([]*models.Split, error) {
	numSplits := len(expense.Splits)
	if numSplits == 0 {
		return nil, fmt.Errorf("no splits provided")
	}

	equalAmount := math.Round(expense.TotalAmount/float64(numSplits)*100) / 100
	sumAmount := 0.0
	for _, split := range expense.Splits {
		split.Amount = equalAmount
		split.Value = 1.0 // Indicating equal split
		sumAmount += split.Amount
	}

	if math.Abs(sumAmount-expense.TotalAmount) >= 0.01 {
		expense.Splits[0].Amount += math.Abs(sumAmount - expense.TotalAmount)
	}

	return expense.Splits, nil
}
