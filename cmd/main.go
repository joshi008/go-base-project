package main

import "fmt"

type SplitType int

const (
	EQUAL SplitType = iota
	EXACT
	PERCENTAGE
)

type User struct {
	ID string
}

type Expense struct {
	LenderId  string
	SplitType SplitType
	Amount    int
	Splits    []*Split
}

type Split struct {
	RawAmount  int
	BorrowerId string
	LenderId   string
	Amount     int
}

// constructor
func NewUserC(ID string) *User {
	return &User{
		ID: ID,
	}
}
func NewExpense(LenderId string, SplitType SplitType, Amount int, Splits []*Split) *Expense {
	return &Expense{
		LenderId:  LenderId,
		SplitType: SplitType,
		Amount:    Amount,
		Splits:    Splits,
	}
}
func NewSplit(RawAmount int, BorrowerId string, LenderId string, Amount int) *Split {
	return &Split{
		RawAmount:  RawAmount,
		BorrowerId: BorrowerId,
		LenderId:   LenderId,
		Amount:     Amount,
	}
}

// Storage
type DB struct {
	Users    map[string]*User
	Expenses map[string]*Expense
	Splits   map[string]*Split
}

func NewDBStorage() *DB {
	fmt.Println("DB Initialised")
	return &DB{
		Users:    make(map[string]*User),
		Expenses: make(map[string]*Expense),
		Splits:   make(map[string]*Split),
	}
}
func (d *DB) NewUser(user *User) *User {
	if d.Users[user.ID] != nil {
		fmt.Println("User Already Exists")
	}
	d.Users[user.ID] = user
	fmt.Println("New User Created: ", user.ID)
	return user
}

// Split Service
type SplitService struct {
	db *DB
}

func NewSplitService(db *DB) *SplitService {
	return &SplitService{
		db: db,
	}
}

func (s *SplitService) AddExpense(LenderId string, SplitType SplitType, BorrowerMap map[string]int, Amount int) *Expense {
	actualSplit := make(map[string]int)
	switch SplitType {
	case EQUAL:
		strat := &EqualSplit{}
		actualSplit = strat.Calculate(Amount, BorrowerMap)
	default:
	}

	splits := make([]*Split, 0)
	for i, _ := range actualSplit {
		st := NewSplit(BorrowerMap[i], i, LenderId, actualSplit[i])
		s.db.Splits[string(len(splits)+1)] = st
		splits = append(splits, st)
	}

	e := NewExpense(LenderId, SplitType, Amount, splits)
	s.db.Expenses[string(len(s.db.Expenses)+1)] = e

	return e
}

func (s *SplitService) CheckBalance(UserId string) map[string]int {
	balanceList := make(map[string]int)
	for p, _ := range s.db.Splits {
		if s.db.Splits[p].BorrowerId == UserId && s.db.Splits[p].LenderId == UserId {
			balanceList[s.db.Splits[p].LenderId] = 0
		} else if s.db.Splits[p].BorrowerId == UserId {
			balanceList[s.db.Splits[p].LenderId] += s.db.Splits[p].Amount
		} else if s.db.Splits[p].LenderId == UserId {
			balanceList[s.db.Splits[p].BorrowerId] -= s.db.Splits[p].Amount
		}
	}
	fmt.Print("Checking Balance for", UserId, balanceList)
	return balanceList
}

type SplitStrat interface {
	Calculate(Amount int, BorrowerMap map[string]int) map[string]int
}
type EqualSplit struct{}

func (e *EqualSplit) Calculate(Amount int, BorrowerMap map[string]int) map[string]int {
	length := len(BorrowerMap)
	sp := Amount / length
	for b, _ := range BorrowerMap {
		BorrowerMap[b] = sp
	}
	return BorrowerMap
}

func main() {
	fmt.Println("Welcome to my program")

	db := NewDBStorage()
	u1 := db.NewUser(NewUserC("Hrishabh"))
	u2 := db.NewUser(NewUserC("Tushar"))
	u3 := db.NewUser(NewUserC("Maithilee"))

	s := NewSplitService(db)
	equalSplitBorrowerMap := make(map[string]int)
	equalSplitBorrowerMap[u1.ID] = 1
	equalSplitBorrowerMap[u2.ID] = 1
	equalSplitBorrowerMap[u3.ID] = 1
	s.AddExpense(u1.ID, EQUAL, equalSplitBorrowerMap, 333)

	s.CheckBalance(u2.ID)

	return
}
