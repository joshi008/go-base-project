package model

type User struct {
	ID int
	Name string
	Expenses Expenses[]
}

type Expense struct {
	ID int
	Amount int
	LenderUserId int
	SplitType string
	Splits []Split
}

type Split struct {
	ID int
	ExpenseId int
	BrowerUserId int
	Amount int
	LenderUserId int
	RawVal int
}

type SplitInternal struct {
	UserI User
	SplitVal int
}
