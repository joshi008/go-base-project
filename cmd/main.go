package main

import (
	"fmt"
	model "go-base-project/models"
	"go-base-project/strategy"
)

func main() { //orchestration layer
	fmt.Println("Starting of the program")

	user1 := model.User {
		ID: 123,
		Name: "Hrishabh",
	}

	user2 := model.User {
		ID: 124,
		Name: "Hrishabh",
	}

	si = model.SplitInternal{
		ID: 123,
		SplitVal: 50,
	}
	sj = []model.SplitInternal[
		{
		ID: 123,
		SplitVal: 50,
	}]

	AddExpense(user1, 2000, "EQUAL",  )
}


func AddExpense(user model.User, amount int, splitType string, split []model.SplitInternal ) (model.Expense) {
	e := model.Expense{
		ID: 2390,
		Amount: amount,
		LenderUserId: user.ID,
		SplitType: splitType,
		Splits: [],
	}

	for i:=0;i<len(split);i++ {
		s := model.Split{
			ID: 31+i,
			ExpenseId: 2390,
			BrowerUserId: split[i].UserI.ID,
			RawVal: split[i].SplitVal,
			Amount: -1,
			LenderUserId: user.ID,
		}
		e.Splits = append(e.Splits, s)
	}

	strategy.PickSplitStrategy(e);


}
