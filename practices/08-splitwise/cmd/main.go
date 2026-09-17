package main

import "fmt"

// Models

type User struct {
	ID           string
	Transactions []Transaction
}

type Transaction struct {
	ID           string
	LenderUserId string
	SplitType    SplitType
	Amount       int
	Splits       []Split
}

type Split struct {
	ID             string
	LenderUserId   string
	BorrowerUserId string
	RawAmount      int
	Amount         int
}

type SplitType int

const (
	EQUAL SplitType = iota
	EXACT
	PERCENTAGE
)

// Constructor
func NewUser(id string) *User {
	return &User{
		ID: id,
	}
}
func NewTransaction(id string, user *User, splitType SplitType, amount int) *Transaction {
	return &Transaction{
		ID:           id,
		LenderUserId: user.ID,
		SplitType:    splitType,
		Amount:       amount,
	}
}
func NewSplit(id string, lenderUserId string, borrowerUserId string, rawAmount int) *Split {
	return &Split{
		ID:             id,
		LenderUserId:   lenderUserId,
		BorrowerUserId: borrowerUserId,
		RawAmount:      rawAmount,
	}
}

// Storage
type Storage struct {
	users        map[string]*User
	transactions map[string]*Transaction
	splits       map[string]*Split
}

func NewStorage() *Storage {
	return &Storage{
		users:        make(map[string]*User),
		transactions: make(map[string]*Transaction),
		splits:       make(map[string]*Split),
	}
}
func (s *Storage) AddUser(user *User) error {
	if _, exists := s.users[user.ID]; exists {
		return fmt.Errorf("User already exists")
	}
	s.users[user.ID] = user
	fmt.Println("User added ", user.ID)
	return nil
}
func (s *Storage) GetUser(id string) (*User, error) {
	if _, exists := s.users[id]; !exists {
		return nil, fmt.Errorf("User does not exist")
	}
	return s.users[id], nil
}
func (s *Storage) GetSplitsForUser(userId string) []*Split {
	hereSplit := []*Split{}
	for _, split := range s.splits {
		if split.BorrowerUserId == userId || split.LenderUserId == userId {
			hereSplit = append(hereSplit, split)
		}
	}
	return hereSplit
}

// Service
type SplitwiseService struct {
	db *Storage
}

func NewSplitwiseService(db *Storage) *SplitwiseService {
	return &SplitwiseService{
		db: db,
	}
}

func (s *SplitwiseService) AddExpense(lenderId string, amount int, borrowers []string, splitStrategy SplitStrategy, splitVals []int) error {
	lender, err := s.db.GetUser(lenderId)
	if err != nil {
		return fmt.Errorf("lender not found: %v", err)
	}

	txId := fmt.Sprintf("tx-%d", len(s.db.transactions)+1)
	tx := NewTransaction(txId, lender, EQUAL, amount)
	s.db.transactions[txId] = tx

	vals, err := splitStrategy.Calculate(amount, borrowers, splitVals)

	for i, borrower := range borrowers {
		if _, err := s.db.GetUser(borrower); err != nil {
			return fmt.Errorf("Not found")
		}

		splitId := fmt.Sprintf("sp-%d-%d", len(s.db.splits)+1, i)
		split := NewSplit(splitId, lender.ID, borrower, vals[borrower])

		s.db.splits[splitId] = split

		tx.Splits = append(tx.Splits, *split)
	}

	fmt.Printf("Success: %s paid %d, split equally among %d people.\n", lenderId, amount, len(borrowers))
	return nil
}

func (s *SplitwiseService) ShowBalances(userId string) (map[string]int, error) {
	_, err := s.db.GetUser(userId)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	splits := s.db.GetSplitsForUser(userId)
	balance := make(map[string]int)
	for _, s := range splits {
		// fmt.Println(s)
		if s.LenderUserId == s.BorrowerUserId {
			// Do nothing
		} else if s.LenderUserId == userId {
			balance[s.BorrowerUserId] -= s.RawAmount
		} else {
			balance[s.LenderUserId] += s.RawAmount
		}
	}

	return balance, nil
}

// Split Strategy
type SplitStrategy interface {
	Calculate(amount int, borrowers []string, vals []int) (map[string]int, error)
}
type EqualSplit struct{}

func (e *EqualSplit) Calculate(amount int, borrowers []string, vals []int) (map[string]int, error) {
	eqAmount := amount / len(borrowers)

	splits := make(map[string]int)
	for _, bor := range borrowers {
		splits[bor] = eqAmount
	}
	return splits, nil
}

type ExactSplit struct{}

func main() {
	fmt.Println("Starting Program!")

	db := NewStorage()
	db.AddUser(NewUser("u1"))
	db.AddUser(NewUser("u2"))
	db.AddUser(NewUser("u3"))

	splitwise := NewSplitwiseService(db)
	involvedUsers := []string{"u1", "u2", "u3"}
	equalStrategy := &EqualSplit{}
	err := splitwise.AddExpense("u1", 900, involvedUsers, equalStrategy, make([]int, 0))
	if err != nil {
		fmt.Println("Split Error")
	}

	involvedUsers = []string{"u1", "u2"}
	err = splitwise.AddExpense("u2", 600, involvedUsers, equalStrategy, make([]int, 0))
	if err != nil {
		fmt.Println("Split Error")
	}

	bal, _ := splitwise.ShowBalances("u1")
	fmt.Println(bal)
}
