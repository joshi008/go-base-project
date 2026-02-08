package strategy

import (
	"fmt"
	"math"

	"go-base-project/models"
)

type PercentageStrategy struct{}

func (s *PercentageStrategy) Validate(expense *models.Expense) error {
	numSplits := len(expense.Splits)
	if numSplits == 0 {
		return fmt.Errorf("no splits provided")
	}

	sumPercentage := 0.0
	sumAmount := 0.0
	for _, split := range expense.Splits {
		sumPercentage += split.Value
		sumAmount += split.Amount
	}

	if math.Abs(sumPercentage-100.0) >= 0.01 {
		return fmt.Errorf("total split percentage must equal 100")
	}
	if math.Abs(sumAmount-expense.TotalAmount) >= 0.01 {
		return fmt.Errorf("total split amount does not match total expense amount")
	}

	return nil
}

func (s *PercentageStrategy) CalculateSplits(expense *models.Expense) ([]*models.Split, error) {
	numSplits := len(expense.Splits)
	if numSplits == 0 {
		return nil, fmt.Errorf("no splits provided")
	}

	sumAmount := 0.0
	for _, split := range expense.Splits {
		split.Amount = math.Round(expense.TotalAmount*(split.Value/100.0)*100) / 100
		sumAmount += split.Amount
	}

	if math.Abs(sumAmount-expense.TotalAmount) >= 0.01 {
		expense.Splits[0].Amount += math.Abs(sumAmount - expense.TotalAmount)
	}

	return expense.Splits, nil
}
