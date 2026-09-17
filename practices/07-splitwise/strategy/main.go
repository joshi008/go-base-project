package strategy

import model "go-base-project/models"

func PickSplitStrategy(e model.Expense) model.Expense {
	if e.SplitType == "EQUAL" {
		return equalStrategy(e)
	} else if e.SplitType == "PERCENTAGE" {
		return percentStrategy(e)
	} else if e.SplitType == "EXACT" {
		return exactStrategy(e)
	}

	return e
}

func equalStrategy(e model.Expense) model.Expense {
	count := len(e.Splits)
	equalAmount := e.Amount / count
	for i := 0; i < count; i++ {
		e.Splits[i].Amount = equalAmount
	}

	return e
}

func percentStrategy(e model.Expense) model.Expense {
	count := len(e.Splits)
	// Validates the split is correct on every split
	//
	// Set the amount in every split
	//
}

// func exactStrategy(e model.Expense) model.Expense {

// }
//
//
// Reconcile UserLevelBalancingSheet
