package main

import "fmt"

type User struct {
	ID string
}
type SplitType int

const (
	EQUAL SplitType = iota
	EXACT
)

type Expense struct {
	ID        string
	LenderID  string
	Amount    int
	SplitType SplitType
	Splits    []*Split
}
type Split struct {
	ID         string
	RawValue   int
	LenderId   string
	BorrowerId string
	Amount     int
}

// Constructor
func NewUser(ID string) *User {
	return &User{
		ID: ID,
	}
}
func NewExpense(ID string, LenderId string, Amount int, SplitType SplitType, Split []*Split) *Expense {
	return &Expense{
		ID:        ID,
		LenderID:  LenderId,
		Amount:    Amount,
		SplitType: SplitType,
		Splits:    Split,
	}
}
func NewSplit(ID string, RawValue int, LenderId string, BorrowerId string, Amount int) *Split {
	return &Split{
		ID:         ID,
		RawValue:   RawValue,
		LenderId:   LenderId,
		BorrowerId: BorrowerId,
		Amount:     Amount,
	}
}

// DB
type DB struct {
	Users   map[string]*User
	Expense map[string]*Expense
	Split   map[string]*Split
}

func NewDB() *DB {
	return &DB{
		Users:   make(map[string]*User),
		Expense: make(map[string]*Expense),
		Split:   make(map[string]*Split),
	}
}
func (d *DB) NewUser(ID string) *User {
	u := NewUser(ID)
	d.Users[ID] = u
	return u
}

// SplitService
type SplitService struct {
	DB *DB
}

func NewSplitService(DB *DB) *SplitService {
	return &SplitService{
		DB: DB,
	}
}
func (s SplitService) AddExpense(LenderID string, BorrowerSplitList map[string]int, SplitType SplitType, Amount int) *Expense {
	var splitTemp map[string]int = make(map[string]int)
	switch SplitType {
	case EQUAL:
		eq := &EqualStrat{}
		splitTemp, _ = eq.Calculate(BorrowerSplitList, Amount)
	case EXACT:
		ex := &ExactStrat{}
		splitTemp, _ = ex.Calculate(BorrowerSplitList, Amount)
	}

	var splits []*Split
	for k, v := range splitTemp {
		sp := NewSplit(string(len(s.DB.Split)+1), BorrowerSplitList[k], LenderID, k, v)
		s.DB.Split[string(len(s.DB.Split)+1)] = sp
		splits = append(splits, sp)
	}

	e := NewExpense(string(len(s.DB.Expense)+1), LenderID, Amount, SplitType, splits)

	s.DB.Expense[string(len(s.DB.Expense)+1)] = e

	return e
}

func (s SplitService) CheckBalance(UserId string) map[string]int {
	m := make(map[string]int)

	for _, v := range s.DB.Split {
		if v.BorrowerId == UserId && v.LenderId == UserId {
			continue
		} else if v.BorrowerId == UserId {
			m[v.LenderId] += v.Amount
		} else if v.LenderId == UserId {
			m[v.BorrowerId] -= v.Amount
		}
	}

	return m
}

type SplitStrategy interface {
	Calculate(BorrowerSplitList map[string]int, Amount int) (map[string]int, error)
}
type EqualStrat struct{}

func (e *EqualStrat) Calculate(BorrowerSplitList map[string]int, Amount int) (map[string]int, error) {
	perPersonAmount := Amount / len(BorrowerSplitList)
	m := make(map[string]int)
	count := 0
	for key, _ := range BorrowerSplitList {
		count++
		m[key] = perPersonAmount
		if count == len(BorrowerSplitList) {
			m[key] += (Amount - (perPersonAmount * len(BorrowerSplitList)))
		}
	}
	return m, nil
}

type ExactStrat struct{}

func (e *ExactStrat) Calculate(BorrowerSplitList map[string]int, Amount int) (map[string]int, error) {
	sum := 0
	m := make(map[string]int)
	for k, v := range BorrowerSplitList {
		sum += v
		m[k] = v
	}
	if sum == Amount {
		return m, nil
	}
	return nil, fmt.Errorf("")
}

func main() {
	fmt.Println("Start of program")

	db := NewDB()
	u1 := db.NewUser("Hrishabh")
	u2 := db.NewUser("Tushar")
	u3 := db.NewUser("Maithilee")

	s := NewSplitService(db)
	BorrowerList := make(map[string]int)
	BorrowerList[u1.ID] = 0
	BorrowerList[u2.ID] = 0
	BorrowerList[u3.ID] = 0
	s.AddExpense(u1.ID, BorrowerList, EQUAL, 600)
	bal := s.CheckBalance(u1.ID)
	fmt.Println(bal)

	BorrowerList[u1.ID] = 200
	BorrowerList[u2.ID] = 100
	BorrowerList[u3.ID] = 200
	s.AddExpense(u2.ID, BorrowerList, EXACT, 500)

	bal = s.CheckBalance(u1.ID)
	fmt.Println(bal)
}
