package main

import (
	"errors"
	"fmt"
	"sort"
)

// ==========================================
// 1. ENUMS (Constants using iota)
// Extremely common for Statuses, Types, etc.
// ==========================================
type SplitType int

const (
	EQUAL   SplitType = iota // 0
	EXACT                    // 1
	PERCENT                  // 2
)

// ==========================================
// 2. STRUCTS & POINTERS
// ==========================================
type User struct {
	ID      string
	Name    string
	Balance float64
}

// Constructor pattern
func NewUser(id, name string) *User {
	return &User{
		ID:      id,
		Name:    name,
		Balance: 0,
	}
}

// ==========================================
// 3. MAPS (The In-Memory DB)
// ==========================================
type Storage struct {
	users map[string]*User
}

func NewStorage() *Storage {
	return &Storage{
		users: make(map[string]*User), // NEVER FORGET 'make'
	}
}

func (s *Storage) AddUser(u *User) error {
	// Check existence
	if _, exists := s.users[u.ID]; exists {
		return errors.New("user already exists")
	}
	s.users[u.ID] = u
	return nil
}

func (s *Storage) GetUser(id string) (*User, error) {
	u, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

// ==========================================
// 4. INTERFACES (Strategy Pattern)
// Essential for open/closed principle (SOLID)
// ==========================================
type SplitStrategy interface {
	Validate(amount float64, splits []float64) bool
	Calculate(amount float64, numPeople int) []float64
}

// Implement the interface for EQUAL
type EqualSplit struct{}

func (e *EqualSplit) Validate(amount float64, splits []float64) bool {
	return true // Equal splits usually don't need custom validation
}

func (e *EqualSplit) Calculate(amount float64, numPeople int) []float64 {
	splitAmt := amount / float64(numPeople)
	res := make([]float64, numPeople)
	for i := range res {
		res[i] = splitAmt
	}
	return res
}

// ==========================================
// 5. SLICES & CUSTOM SORTING
// Often asked for "Top K" or "Leaderboard" features
// ==========================================
func SortUsersByBalance(users []*User) {
	// Sort descending (highest balance first)
	sort.Slice(users, func(i, j int) bool {
		return users[i].Balance > users[j].Balance
	})
}

// ==========================================
// 6. MAIN & DRIVER FLOW
// ==========================================
func main() {
	// 1. Initialize
	db := NewStorage()

	// 2. Load Data
	db.AddUser(NewUser("u1", "Alice"))
	db.AddUser(NewUser("u2", "Bob"))

	// 3. Handle Errors Gracefully
	user, err := db.GetUser("u1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 4. Switch Statements for Factory/Logic routing
	sType := EQUAL
	var strategy SplitStrategy

	switch sType {
	case EQUAL:
		strategy = &EqualSplit{}
	case EXACT:
		// strategy = &ExactSplit{}
	default:
		fmt.Println("Unknown split type")
	}

	// 5. Execute Logic & Slice Iteration
	splits := strategy.Calculate(100.0, 2)

	// Appending to slices
	var results []string
	for i, amt := range splits {
		results = append(results, fmt.Sprintf("Person %d pays %.2f", i+1, amt))
	}

	fmt.Printf("User %s processed: %v\n", user.Name, results)
}
