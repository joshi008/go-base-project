package main

import (
	"fmt"
	"strconv"
)

type SplitType int

const (
	EQUAL SplitType = iota
	EXACT
)

type User struct {
	ID string
}
type Expense struct {
	ID           string
	Amount       int64
	LenderUserId string
	Splits       []*Split
}
type Split struct {
	ID             string
	LenderUserId   string
	BorrowerUserId string
	Amount         int64
	RawSplit       int64
}

// Constructors
func NewUser(ID string) *User {
	return &User{
		ID: ID,
	}
}
func NewExpense(ID string, Amount int64, LenderUserId string, Splits []*Split) *Expense {
	return &Expense{
		ID:           ID,
		Amount:       Amount,
		LenderUserId: LenderUserId,
		Splits:       Splits,
	}
}
func NewSplit(ID string, LenderUserId string, BorrowerUserId string, Amount int64, RawSplit int64) *Split {
	return &Split{
		ID:             ID,
		LenderUserId:   LenderUserId,
		BorrowerUserId: BorrowerUserId,
		Amount:         Amount,
		RawSplit:       RawSplit,
	}
}

// Storage
type DB struct {
	Users    map[string]*User
	Expenses map[string]*Expense
	Splits   map[string]*Split
}

func NewStorage() *DB {
	return &DB{
		Users:    make(map[string]*User),
		Expenses: make(map[string]*Expense),
		Splits:   make(map[string]*Split),
	}
}
func (d *DB) NewUser(ID string) *User {
	nu := NewUser(ID)
	d.Users[ID] = nu
	return nu
}

// Expense Service
type ExpenseManager struct {
	DB *DB
}

func NewExpenseManager(db *DB) *ExpenseManager {
	return &ExpenseManager{
		DB: db,
	}
}

func (e *ExpenseManager) AddExpense(Amount int64, LenderUserId string, BorrowerSplit map[string]int64, SplitType SplitType) {
	var rawSplits map[string]int64
	switch SplitType {
	case EXACT:
		exactStrat := &ExactStrategy{}
		rawSplits = exactStrat.Evaluate(BorrowerSplit, Amount)
	case EQUAL:
		equalStrat := &EqualStrategy{}
		rawSplits = equalStrat.Evaluate(BorrowerSplit, Amount)
	}

	var splits []*Split

	for sp, _ := range rawSplits {
		newId := strconv.Itoa(len(e.DB.Splits) + 1)
		s := NewSplit(newId, LenderUserId, sp, rawSplits[sp], BorrowerSplit[sp])
		splits = append(splits, s)
		e.DB.Splits[newId] = s
	}

	newExId := strconv.Itoa(len(e.DB.Expenses) + 1)
	expense := NewExpense(newExId, Amount, LenderUserId, splits)
	e.DB.Expenses[newExId] = expense
}

func (e *ExpenseManager) GetMyBalance(ID string) map[string]int64 {
	bal := make(map[string]int64)
	for s, _ := range e.DB.Splits {
		sp := e.DB.Splits[s]
		if ID == sp.BorrowerUserId && ID == sp.LenderUserId {
			continue
		} else if ID == sp.BorrowerUserId {
			bal[sp.LenderUserId] -= sp.Amount
		} else if ID == sp.LenderUserId {
			bal[sp.BorrowerUserId] += sp.Amount
		}
	}

	return bal
}

// Split Service
type SplitStrategy interface {
	Evaluate(BorrowerSplit map[string]int64, Amount int64) map[string]int64
}
type ExactStrategy struct{}

func (e *ExactStrategy) Evaluate(BorrowerSplit map[string]int64, Amount int64) map[string]int64 {
	return BorrowerSplit
}

type EqualStrategy struct{}

func (e *EqualStrategy) Evaluate(BorrowerSplit map[string]int64, Amount int64) map[string]int64 {
	splits := make(map[string]int64)
	split := Amount / int64(len(BorrowerSplit))
	extra := Amount - (split * int64(len(BorrowerSplit)))
	for c, i := range BorrowerSplit {
		splits[c] = split
		if i == 0 {
			splits[c] += extra
			extra = 0
		}
	}
	return splits
}

func main() {
	fmt.Println("--- California Burrito Split ---")

	db := NewStorage()

	// Registering the group
	hrishabh := db.NewUser("Hrishabh")
	suruthiga := db.NewUser("Suruthiga")
	aniket := db.NewUser("Aniket")
	omkar := db.NewUser("Omkar")
	bhumika := db.NewUser("Bhumika")

	e := NewExpenseManager(db)

	// Creating the exact split payload based on the receipt
	BorrowerSplit := make(map[string]int64)
	BorrowerSplit[hrishabh.ID] = 340
	BorrowerSplit[suruthiga.ID] = 241
	BorrowerSplit[aniket.ID] = 174
	BorrowerSplit[omkar.ID] = 303
	BorrowerSplit[bhumika.ID] = 140

	// Hrishabh pays the full amount of 1198
	e.AddExpense(1198, hrishabh.ID, BorrowerSplit, EXACT)

	// Fetch balances for Hrishabh
	bal := e.GetMyBalance(hrishabh.ID)

	fmt.Println("\nAmount owed to Hrishabh:")
	for user, amount := range bal {
		fmt.Printf("- %s owes: ₹%d\n", user, amount)
	}
}
