# Library Management System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement HTTP APIs for a library book management system (books, users, borrow/return) backed by Postgres, with business rules enforced via in-process mutex locking.

**Architecture:** Layered: `internal/models` (plain structs) → `internal/repository` (interface + `memory`/`postgres` implementations, each enforcing borrow/return business rules transactionally) → `internal/libraryservice` (adds per-user/per-book mutex locking on top of the repository) → `internal/httpserver` (net/http handlers) → `cmd/main.go` (wiring).

**Tech Stack:** Go 1.24, standard library `net/http` (Go 1.22+ ServeMux path params), `github.com/jackc/pgx/v5` (Postgres driver), `github.com/google/uuid` (ID generation), standard `testing` package + `go test -race`.

Full design context: `docs/superpowers/specs/2026-07-28-library-management-design.md`.

## Global Constraints

- A user may hold at most 3 concurrently-borrowed books.
- A user may not hold 2 copies of the same BookID at once.
- Borrowing decrements `Book.AvailabilityQuantity`; returning increments it; borrow fails if quantity is already 0.
- Concurrency correctness relies on in-process `sync.Mutex` locking (single-process only, not distributed-safe) — this is intentional and acceptable for this project.
- Package layout lives under `internal/` (models, repository, libraryservice, httpserver) — not importable outside this module.
- Postgres is assumed already running locally; connect via `DATABASE_URL` env var with a localhost default. No migration tooling — `CREATE TABLE IF NOT EXISTS` on startup.

---

### Task 1: Domain models, sentinel errors, and repository interface

**Files:**
- Create: `internal/models/book.go`
- Create: `internal/models/author.go`
- Create: `internal/models/user.go`
- Create: `internal/models/ledger.go`
- Create: `internal/apperrors/errors.go`
- Create: `internal/repository/repository.go`

**Interfaces:**
- Produces: `models.Book{ID, Name, AvailabilityQuantity, CreatedAt}`, `models.Author{ID, Name}`, `models.AuthorBookMapping{ID, BookID, AuthorID}`, `models.User{ID, Name, CreatedAt}`, `models.ActionType` (string enum, `models.ActionBorrow`), `models.LibraryLedgerEntry{ID, UserID, BookID, ActionType, ReturnedAt *time.Time, CreatedAt, UpdatedAt}`.
- Produces: `apperrors.ErrUserNotFound`, `ErrBookNotFound`, `ErrBorrowLimitReached`, `ErrAlreadyBorrowed`, `ErrBookUnavailable`, `ErrNotBorrowed` (all `error` values).
- Produces: `repository.BookFilter{AvailableOnly bool, AuthorID string}` and `repository.Repository` interface (see below) — consumed by every later task.

This task has no behavior to test (plain data + an interface declaration) — its test is that the module builds.

- [ ] **Step 1: Add module dependencies**

Run: `go get github.com/google/uuid && go get github.com/jackc/pgx/v5`

- [ ] **Step 2: Write the model files**

`internal/models/book.go`:
```go
package models

import "time"

type Book struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	AvailabilityQuantity int       `json:"availabilityQuantity"`
	CreatedAt            time.Time `json:"createdAt"`
}
```

`internal/models/author.go`:
```go
package models

type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AuthorBookMapping struct {
	ID       string `json:"id"`
	BookID   string `json:"bookId"`
	AuthorID string `json:"authorId"`
}
```

`internal/models/user.go`:
```go
package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
```

`internal/models/ledger.go`:
```go
package models

import "time"

type ActionType string

const (
	ActionBorrow ActionType = "BORROW"
)

type LibraryLedgerEntry struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	BookID     string     `json:"bookId"`
	ActionType ActionType `json:"actionType"`
	ReturnedAt *time.Time `json:"returnedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
```

- [ ] **Step 3: Write the sentinel errors**

`internal/apperrors/errors.go`:
```go
package apperrors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrBookNotFound       = errors.New("book not found")
	ErrBorrowLimitReached = errors.New("user has reached the 3 book borrow limit")
	ErrAlreadyBorrowed    = errors.New("user already holds a copy of this book")
	ErrBookUnavailable    = errors.New("no copies of this book are currently available")
	ErrNotBorrowed        = errors.New("user does not currently have this book borrowed")
)
```

- [ ] **Step 4: Write the repository interface**

`internal/repository/repository.go`:
```go
package repository

import (
	"context"

	"go-base-project/internal/models"
)

type BookFilter struct {
	AvailableOnly bool
	AuthorID      string
}

// Repository is implemented by internal/repository/memory (tests) and
// internal/repository/postgres (production). Borrow and Return enforce all
// business rules (3-book limit, no duplicate same-book borrow, availability
// check) atomically within a single logical operation, returning the
// apperrors sentinels on rule violations.
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)

	CreateBook(ctx context.Context, book *models.Book, authorIDs []string) error
	GetBook(ctx context.Context, id string) (*models.Book, error)
	ListBooks(ctx context.Context, filter BookFilter) ([]*models.Book, error)

	Borrow(ctx context.Context, userID, bookID string) (*models.LibraryLedgerEntry, error)
	Return(ctx context.Context, userID, bookID string) error
	ListBorrowedBooks(ctx context.Context, userID string) ([]*models.Book, error)
}
```

- [ ] **Step 5: Verify it builds**

Run: `go build ./...`
Expected: no output (success).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/models internal/apperrors internal/repository/repository.go
git commit -m "feat: add domain models, sentinel errors, and repository interface"
```

---

### Task 2: In-memory repository (CRUD + borrow/return business rules)

**Files:**
- Create: `internal/repository/memory/memory.go`
- Test: `internal/repository/memory/memory_test.go`

**Interfaces:**
- Consumes: `models.*` (Task 1), `apperrors.*` (Task 1), `repository.BookFilter` and `repository.Repository` (Task 1).
- Produces: `memory.New() *memory.Repository`, satisfying `repository.Repository`. Used directly by `libraryservice` tests (Task 3) and `httpserver` tests (Task 5/6) as the test-time backing store.

- [ ] **Step 1: Write the failing tests**

`internal/repository/memory/memory_test.go`:
```go
package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-base-project/internal/apperrors"
	"go-base-project/internal/models"
	"go-base-project/internal/repository"
	"go-base-project/internal/repository/memory"
)

func newUser(name string) *models.User {
	return &models.User{ID: uuid.New().String(), Name: name, CreatedAt: time.Now()}
}

func newBook(name string, qty int) *models.Book {
	return &models.Book{ID: uuid.New().String(), Name: name, AvailabilityQuantity: qty, CreatedAt: time.Now()}
}

func TestCreateAndGetUser(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()

	u := newUser("Alice")
	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := repo.GetUser(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Name != "Alice" {
		t.Fatalf("expected Alice, got %q", got.Name)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := memory.New()
	if _, err := repo.GetUser(context.Background(), "missing"); err != apperrors.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateAndGetBook(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()

	b := newBook("Dune", 3)
	if err := repo.CreateBook(ctx, b, nil); err != nil {
		t.Fatalf("CreateBook: %v", err)
	}

	got, err := repo.GetBook(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.Name != "Dune" || got.AvailabilityQuantity != 3 {
		t.Fatalf("unexpected book: %+v", got)
	}
}

func TestListBooks_AvailableOnlyFilter(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()

	available := newBook("Dune", 1)
	unavailable := newBook("Foundation", 0)
	repo.CreateBook(ctx, available, nil)
	repo.CreateBook(ctx, unavailable, nil)

	books, err := repo.ListBooks(ctx, repository.BookFilter{AvailableOnly: true})
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 || books[0].ID != available.ID {
		t.Fatalf("expected only available book, got %+v", books)
	}
}

func TestListBooks_AuthorFilter(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	authorID := uuid.New().String()

	byAuthor := newBook("Dune", 1)
	other := newBook("Foundation", 1)
	repo.CreateBook(ctx, byAuthor, []string{authorID})
	repo.CreateBook(ctx, other, nil)

	books, err := repo.ListBooks(ctx, repository.BookFilter{AuthorID: authorID})
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 || books[0].ID != byAuthor.ID {
		t.Fatalf("expected only byAuthor book, got %+v", books)
	}
}

func TestBorrow_Success(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()

	u := newUser("Alice")
	b := newBook("Dune", 2)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b, nil)

	entry, err := repo.Borrow(ctx, u.ID, b.ID)
	if err != nil {
		t.Fatalf("Borrow: %v", err)
	}
	if entry.UserID != u.ID || entry.BookID != b.ID || entry.ReturnedAt != nil {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	got, _ := repo.GetBook(ctx, b.ID)
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("expected availability 1, got %d", got.AvailabilityQuantity)
	}
}

func TestBorrow_UnknownUserOrBook(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	b := newBook("Dune", 1)
	repo.CreateBook(ctx, b, nil)

	if _, err := repo.Borrow(ctx, "missing-user", b.ID); err != apperrors.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	u := newUser("Alice")
	repo.CreateUser(ctx, u)
	if _, err := repo.Borrow(ctx, u.ID, "missing-book"); err != apperrors.ErrBookNotFound {
		t.Fatalf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBorrow_Unavailable(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	b := newBook("Dune", 0)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b, nil)

	if _, err := repo.Borrow(ctx, u.ID, b.ID); err != apperrors.ErrBookUnavailable {
		t.Fatalf("expected ErrBookUnavailable, got %v", err)
	}
}

func TestBorrow_DuplicateSameBookRejected(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	b := newBook("Dune", 5)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b, nil)

	if _, err := repo.Borrow(ctx, u.ID, b.ID); err != nil {
		t.Fatalf("first borrow: %v", err)
	}
	if _, err := repo.Borrow(ctx, u.ID, b.ID); err != apperrors.ErrAlreadyBorrowed {
		t.Fatalf("expected ErrAlreadyBorrowed, got %v", err)
	}
}

func TestBorrow_FourthBookRejected(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	repo.CreateUser(ctx, u)

	var bookIDs []string
	for i := 0; i < 4; i++ {
		b := newBook("Book", 1)
		repo.CreateBook(ctx, b, nil)
		bookIDs = append(bookIDs, b.ID)
	}

	for i := 0; i < 3; i++ {
		if _, err := repo.Borrow(ctx, u.ID, bookIDs[i]); err != nil {
			t.Fatalf("borrow %d: %v", i, err)
		}
	}
	if _, err := repo.Borrow(ctx, u.ID, bookIDs[3]); err != apperrors.ErrBorrowLimitReached {
		t.Fatalf("expected ErrBorrowLimitReached, got %v", err)
	}
}

func TestReturn_Success(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	b := newBook("Dune", 1)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b, nil)
	repo.Borrow(ctx, u.ID, b.ID)

	if err := repo.Return(ctx, u.ID, b.ID); err != nil {
		t.Fatalf("Return: %v", err)
	}

	got, _ := repo.GetBook(ctx, b.ID)
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("expected availability 1 after return, got %d", got.AvailabilityQuantity)
	}
}

func TestReturn_NotBorrowedRejected(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	b := newBook("Dune", 1)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b, nil)

	if err := repo.Return(ctx, u.ID, b.ID); err != apperrors.ErrNotBorrowed {
		t.Fatalf("expected ErrNotBorrowed, got %v", err)
	}
}

func TestListBorrowedBooks_OnlyActive(t *testing.T) {
	repo := memory.New()
	ctx := context.Background()
	u := newUser("Alice")
	b1 := newBook("Dune", 1)
	b2 := newBook("Foundation", 1)
	repo.CreateUser(ctx, u)
	repo.CreateBook(ctx, b1, nil)
	repo.CreateBook(ctx, b2, nil)

	repo.Borrow(ctx, u.ID, b1.ID)
	repo.Borrow(ctx, u.ID, b2.ID)
	repo.Return(ctx, u.ID, b1.ID)

	borrowed, err := repo.ListBorrowedBooks(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListBorrowedBooks: %v", err)
	}
	if len(borrowed) != 1 || borrowed[0].ID != b2.ID {
		t.Fatalf("expected only b2 borrowed, got %+v", borrowed)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/repository/memory/... -v`
Expected: FAIL — package `memory` (and `New`) does not exist yet.

- [ ] **Step 3: Implement the in-memory repository**

`internal/repository/memory/memory.go`:
```go
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"go-base-project/internal/apperrors"
	"go-base-project/internal/models"
	"go-base-project/internal/repository"
)

// Repository is a thread-safe in-memory implementation of repository.Repository,
// intended for tests only.
type Repository struct {
	mu      sync.Mutex
	users   map[string]*models.User
	books   map[string]*models.Book
	authors map[string][]string // bookID -> authorIDs
	ledger  map[string]*models.LibraryLedgerEntry
}

func New() *Repository {
	return &Repository{
		users:   make(map[string]*models.User),
		books:   make(map[string]*models.Book),
		authors: make(map[string][]string),
		ledger:  make(map[string]*models.LibraryLedgerEntry),
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *user
	r.users[user.ID] = &stored
	return nil
}

func (r *Repository) GetUser(ctx context.Context, id string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, apperrors.ErrUserNotFound
	}
	got := *u
	return &got, nil
}

func (r *Repository) CreateBook(ctx context.Context, book *models.Book, authorIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *book
	r.books[book.ID] = &stored
	r.authors[book.ID] = append([]string{}, authorIDs...)
	return nil
}

func (r *Repository) GetBook(ctx context.Context, id string) (*models.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.books[id]
	if !ok {
		return nil, apperrors.ErrBookNotFound
	}
	got := *b
	return &got, nil
}

func (r *Repository) ListBooks(ctx context.Context, filter repository.BookFilter) ([]*models.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*models.Book
	for _, b := range r.books {
		if filter.AvailableOnly && b.AvailabilityQuantity <= 0 {
			continue
		}
		if filter.AuthorID != "" {
			matches := false
			for _, aid := range r.authors[b.ID] {
				if aid == filter.AuthorID {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}
		got := *b
		result = append(result, &got)
	}
	return result, nil
}

func (r *Repository) Borrow(ctx context.Context, userID, bookID string) (*models.LibraryLedgerEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[userID]; !ok {
		return nil, apperrors.ErrUserNotFound
	}
	book, ok := r.books[bookID]
	if !ok {
		return nil, apperrors.ErrBookNotFound
	}

	activeCount := 0
	for _, entry := range r.ledger {
		if entry.UserID != userID || entry.ReturnedAt != nil {
			continue
		}
		activeCount++
		if entry.BookID == bookID {
			return nil, apperrors.ErrAlreadyBorrowed
		}
	}
	if activeCount >= 3 {
		return nil, apperrors.ErrBorrowLimitReached
	}
	if book.AvailabilityQuantity <= 0 {
		return nil, apperrors.ErrBookUnavailable
	}

	book.AvailabilityQuantity--

	now := time.Now()
	entry := &models.LibraryLedgerEntry{
		ID:         uuid.New().String(),
		UserID:     userID,
		BookID:     bookID,
		ActionType: models.ActionBorrow,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.ledger[entry.ID] = entry

	got := *entry
	return &got, nil
}

func (r *Repository) Return(ctx context.Context, userID, bookID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var active *models.LibraryLedgerEntry
	for _, entry := range r.ledger {
		if entry.UserID == userID && entry.BookID == bookID && entry.ReturnedAt == nil {
			active = entry
			break
		}
	}
	if active == nil {
		return apperrors.ErrNotBorrowed
	}

	book, ok := r.books[bookID]
	if !ok {
		return apperrors.ErrBookNotFound
	}

	now := time.Now()
	active.ReturnedAt = &now
	active.UpdatedAt = now
	book.AvailabilityQuantity++
	return nil
}

func (r *Repository) ListBorrowedBooks(ctx context.Context, userID string) ([]*models.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[userID]; !ok {
		return nil, apperrors.ErrUserNotFound
	}

	var result []*models.Book
	for _, entry := range r.ledger {
		if entry.UserID != userID || entry.ReturnedAt != nil {
			continue
		}
		if book, ok := r.books[entry.BookID]; ok {
			got := *book
			result = append(result, &got)
		}
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/repository/memory/... -v`
Expected: PASS (all tests).

- [ ] **Step 5: Commit**

```bash
git add internal/repository/memory
git commit -m "feat: implement in-memory repository with borrow/return business rules"
```

---

### Task 3: libraryservice — mutex-locked service layer

**Files:**
- Create: `internal/libraryservice/locks.go`
- Create: `internal/libraryservice/service.go`
- Test: `internal/libraryservice/service_test.go`

**Interfaces:**
- Consumes: `repository.Repository`, `repository.BookFilter` (Task 1); `memory.New()` (Task 2, test-only); `models.*`, `apperrors.*` (Task 1).
- Produces: `libraryservice.New(repo repository.Repository) *Service` and methods `RegisterUser(ctx, name string) (*models.User, error)`, `GetUser(ctx, id string) (*models.User, error)`, `AddBook(ctx, name string, authorIDs []string, quantity int) (*models.Book, error)`, `GetBook(ctx, id string) (*models.Book, error)`, `ListBooks(ctx, filter repository.BookFilter) ([]*models.Book, error)`, `Borrow(ctx, userID, bookID string) (*models.LibraryLedgerEntry, error)`, `Return(ctx, userID, bookID string) error`, `ListBorrowedBooks(ctx, userID string) ([]*models.Book, error)` — consumed directly by `httpserver` (Tasks 5/6).

This is the layer the user-facing "greater than 3 books" and "one book, one user at a time" test cases target.

- [ ] **Step 1: Write the failing tests**

`internal/libraryservice/service_test.go`:
```go
package libraryservice_test

import (
	"context"
	"sync"
	"testing"

	"go-base-project/internal/apperrors"
	"go-base-project/internal/libraryservice"
	"go-base-project/internal/repository/memory"
)

func newTestService() *libraryservice.Service {
	return libraryservice.New(memory.New())
}

func TestRegisterAndGetUser(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	user, err := svc.RegisterUser(ctx, "Alice")
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}
	if user.ID == "" || user.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", user)
	}

	got, err := svc.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Name != "Alice" {
		t.Fatalf("expected Alice, got %q", got.Name)
	}
}

func TestAddAndGetBook(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	book, err := svc.AddBook(ctx, "Dune", nil, 2)
	if err != nil {
		t.Fatalf("AddBook: %v", err)
	}
	if book.ID == "" || book.AvailabilityQuantity != 2 {
		t.Fatalf("unexpected book: %+v", book)
	}

	got, err := svc.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.Name != "Dune" {
		t.Fatalf("expected Dune, got %q", got.Name)
	}
}

func TestBorrow_Success(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	user, _ := svc.RegisterUser(ctx, "Alice")
	book, _ := svc.AddBook(ctx, "Dune", nil, 2)

	entry, err := svc.Borrow(ctx, user.ID, book.ID)
	if err != nil {
		t.Fatalf("Borrow: %v", err)
	}
	if entry.UserID != user.ID || entry.BookID != book.ID {
		t.Fatalf("unexpected ledger entry: %+v", entry)
	}

	got, _ := svc.GetBook(ctx, book.ID)
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("expected availability 1, got %d", got.AvailabilityQuantity)
	}
}

func TestBorrow_FourthBookRejected(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	user, _ := svc.RegisterUser(ctx, "Alice")

	var bookIDs []string
	for i := 0; i < 4; i++ {
		book, err := svc.AddBook(ctx, "Book", nil, 1)
		if err != nil {
			t.Fatalf("AddBook: %v", err)
		}
		bookIDs = append(bookIDs, book.ID)
	}

	for i := 0; i < 3; i++ {
		if _, err := svc.Borrow(ctx, user.ID, bookIDs[i]); err != nil {
			t.Fatalf("borrow %d: %v", i, err)
		}
	}

	if _, err := svc.Borrow(ctx, user.ID, bookIDs[3]); err != apperrors.ErrBorrowLimitReached {
		t.Fatalf("expected ErrBorrowLimitReached, got %v", err)
	}
}

func TestBorrow_DuplicateSameBookRejected(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	user, _ := svc.RegisterUser(ctx, "Alice")
	book, _ := svc.AddBook(ctx, "Dune", nil, 5)

	if _, err := svc.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("first borrow: %v", err)
	}
	if _, err := svc.Borrow(ctx, user.ID, book.ID); err != apperrors.ErrAlreadyBorrowed {
		t.Fatalf("expected ErrAlreadyBorrowed, got %v", err)
	}
}

// TestBorrow_ConcurrentSingleCopy_OnlyOneSucceeds: run with `go test -race`.
func TestBorrow_ConcurrentSingleCopy_OnlyOneSucceeds(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	userA, _ := svc.RegisterUser(ctx, "Alice")
	userB, _ := svc.RegisterUser(ctx, "Bob")
	book, _ := svc.AddBook(ctx, "Dune", nil, 1)

	users := []string{userA.ID, userB.ID}
	results := make([]error, len(users))

	var wg sync.WaitGroup
	for i := range users {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.Borrow(ctx, users[i], book.ID)
			results[i] = err
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful borrow, got %d (errors: %v)", successes, results)
	}
}

func TestReturn_Success(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	user, _ := svc.RegisterUser(ctx, "Alice")
	book, _ := svc.AddBook(ctx, "Dune", nil, 1)
	if _, err := svc.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow: %v", err)
	}

	if err := svc.Return(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Return: %v", err)
	}

	got, _ := svc.GetBook(ctx, book.ID)
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("expected availability 1 after return, got %d", got.AvailabilityQuantity)
	}
}

func TestReturn_NotBorrowedRejected(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	user, _ := svc.RegisterUser(ctx, "Alice")
	book, _ := svc.AddBook(ctx, "Dune", nil, 1)

	if err := svc.Return(ctx, user.ID, book.ID); err != apperrors.ErrNotBorrowed {
		t.Fatalf("expected ErrNotBorrowed, got %v", err)
	}
}

func TestListBorrowedBooks_OnlyActive(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	user, _ := svc.RegisterUser(ctx, "Alice")
	book1, _ := svc.AddBook(ctx, "Dune", nil, 1)
	book2, _ := svc.AddBook(ctx, "Foundation", nil, 1)

	if _, err := svc.Borrow(ctx, user.ID, book1.ID); err != nil {
		t.Fatalf("Borrow book1: %v", err)
	}
	if _, err := svc.Borrow(ctx, user.ID, book2.ID); err != nil {
		t.Fatalf("Borrow book2: %v", err)
	}
	if err := svc.Return(ctx, user.ID, book1.ID); err != nil {
		t.Fatalf("Return book1: %v", err)
	}

	borrowed, err := svc.ListBorrowedBooks(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListBorrowedBooks: %v", err)
	}
	if len(borrowed) != 1 || borrowed[0].ID != book2.ID {
		t.Fatalf("expected only book2 borrowed, got %+v", borrowed)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/libraryservice/... -v`
Expected: FAIL — package `libraryservice` does not exist yet.

- [ ] **Step 3: Implement the lock manager**

`internal/libraryservice/locks.go`:
```go
package libraryservice

import "sync"

// lockManager hands out one *sync.Mutex per key (user ID or book ID),
// creating it lazily on first use. It only serializes access within this
// process — see the design doc for why that's sufficient here.
type lockManager struct {
	mu        sync.Mutex
	userLocks map[string]*sync.Mutex
	bookLocks map[string]*sync.Mutex
}

func newLockManager() *lockManager {
	return &lockManager{
		userLocks: make(map[string]*sync.Mutex),
		bookLocks: make(map[string]*sync.Mutex),
	}
}

func (l *lockManager) userLock(id string) *sync.Mutex {
	return l.lockFor(l.userLocks, id)
}

func (l *lockManager) bookLock(id string) *sync.Mutex {
	return l.lockFor(l.bookLocks, id)
}

func (l *lockManager) lockFor(m map[string]*sync.Mutex, id string) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()

	mu, ok := m[id]
	if !ok {
		mu = &sync.Mutex{}
		m[id] = mu
	}
	return mu
}
```

- [ ] **Step 4: Implement the service**

`internal/libraryservice/service.go`:
```go
package libraryservice

import (
	"context"
	"time"

	"github.com/google/uuid"

	"go-base-project/internal/models"
	"go-base-project/internal/repository"
)

type Service struct {
	repo  repository.Repository
	locks *lockManager
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo, locks: newLockManager()}
}

func (s *Service) RegisterUser(ctx context.Context, name string) (*models.User, error) {
	user := &models.User{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *Service) AddBook(ctx context.Context, name string, authorIDs []string, quantity int) (*models.Book, error) {
	book := &models.Book{
		ID:                   uuid.New().String(),
		Name:                 name,
		AvailabilityQuantity: quantity,
		CreatedAt:            time.Now(),
	}
	if err := s.repo.CreateBook(ctx, book, authorIDs); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *Service) GetBook(ctx context.Context, id string) (*models.Book, error) {
	return s.repo.GetBook(ctx, id)
}

func (s *Service) ListBooks(ctx context.Context, filter repository.BookFilter) ([]*models.Book, error) {
	return s.repo.ListBooks(ctx, filter)
}

// Borrow locks user-then-book (always this order, to avoid deadlocks),
// then delegates to the repository, which performs the actual rule
// checks and persistence as one logical/transactional operation.
func (s *Service) Borrow(ctx context.Context, userID, bookID string) (*models.LibraryLedgerEntry, error) {
	userLock := s.locks.userLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	bookLock := s.locks.bookLock(bookID)
	bookLock.Lock()
	defer bookLock.Unlock()

	return s.repo.Borrow(ctx, userID, bookID)
}

func (s *Service) Return(ctx context.Context, userID, bookID string) error {
	userLock := s.locks.userLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	bookLock := s.locks.bookLock(bookID)
	bookLock.Lock()
	defer bookLock.Unlock()

	return s.repo.Return(ctx, userID, bookID)
}

func (s *Service) ListBorrowedBooks(ctx context.Context, userID string) ([]*models.Book, error) {
	return s.repo.ListBorrowedBooks(ctx, userID)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -race ./internal/libraryservice/... -v`
Expected: PASS (all tests). The `-race` flag is required for `TestBorrow_ConcurrentSingleCopy_OnlyOneSucceeds` to meaningfully verify there's no data race.

- [ ] **Step 6: Commit**

```bash
git add internal/libraryservice
git commit -m "feat: add libraryservice with per-user/per-book mutex locking"
```

---

### Task 4: HTTP server scaffolding + error/response helpers

**Files:**
- Create: `internal/httpserver/responses.go`
- Create: `internal/httpserver/server.go`

**Interfaces:**
- Consumes: `libraryservice.Service` (Task 3), `apperrors.*` (Task 1).
- Produces: `httpserver.New(svc *libraryservice.Service) *Server` where `*Server` implements `http.Handler`; internal helpers `writeJSON(w, status, payload)`, `writeError(w, err)`, `writeValidationError(w, msg)` — consumed by handler files in Tasks 5 and 6.

No dedicated test for this task (it's routing/response scaffolding exercised by Tasks 5-6's handler tests); verify with a build.

- [ ] **Step 1: Write the response helpers**

`internal/httpserver/responses.go`:
```go
package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-base-project/internal/apperrors"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, statusForError(err), errorResponse{Error: err.Error()})
}

func writeValidationError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: msg})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrUserNotFound), errors.Is(err, apperrors.ErrBookNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperrors.ErrBorrowLimitReached),
		errors.Is(err, apperrors.ErrAlreadyBorrowed),
		errors.Is(err, apperrors.ErrBookUnavailable),
		errors.Is(err, apperrors.ErrNotBorrowed):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

- [ ] **Step 2: Write the server + routing**

`internal/httpserver/server.go`:
```go
package httpserver

import (
	"net/http"

	"go-base-project/internal/libraryservice"
)

type Server struct {
	svc *libraryservice.Service
	mux *http.ServeMux
}

func New(svc *libraryservice.Service) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /books", s.handleAddBook)
	s.mux.HandleFunc("GET /books/{id}", s.handleGetBook)
	s.mux.HandleFunc("GET /books", s.handleListBooks)

	s.mux.HandleFunc("POST /users", s.handleRegisterUser)
	s.mux.HandleFunc("GET /users/{id}", s.handleGetUser)

	s.mux.HandleFunc("POST /users/{id}/borrow", s.handleBorrow)
	s.mux.HandleFunc("POST /users/{id}/return", s.handleReturn)
	s.mux.HandleFunc("GET /users/{id}/books", s.handleListBorrowedBooks)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
```

- [ ] **Step 3: Verify it builds**

Run: `go build ./...`
Expected: fails with "undefined: s.handleAddBook" etc. — expected, since handlers are written in Tasks 5-6. This confirms the routing table references the right method names for those tasks to implement.

- [ ] **Step 4: Commit**

Do not commit yet — this task doesn't compile on its own (handlers referenced in `routes()` don't exist until Task 5). Combine the commit with Task 5's commit instead. Skip this step here.

---

### Task 5: Book and user HTTP handlers

**Files:**
- Create: `internal/httpserver/books_handler.go`
- Create: `internal/httpserver/users_handler.go`
- Test: `internal/httpserver/books_users_handler_test.go`

**Interfaces:**
- Consumes: `Server`, `writeJSON`/`writeError`/`writeValidationError` (Task 4); `libraryservice.Service` methods `AddBook`, `GetBook`, `ListBooks`, `RegisterUser`, `GetUser` (Task 3); `repository.BookFilter` (Task 1); `memory.New()` (Task 2, test-only).
- Produces: `s.handleAddBook`, `s.handleGetBook`, `s.handleListBooks`, `s.handleRegisterUser`, `s.handleGetUser` — referenced by `routes()` in Task 4.

- [ ] **Step 1: Write the failing tests**

`internal/httpserver/books_users_handler_test.go`:
```go
package httpserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-base-project/internal/httpserver"
	"go-base-project/internal/libraryservice"
	"go-base-project/internal/models"
	"go-base-project/internal/repository/memory"
)

func newTestServer() *httpserver.Server {
	svc := libraryservice.New(memory.New())
	return httpserver.New(svc)
}

func doJSON(t *testing.T, srv http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestAddBook_Success(t *testing.T) {
	srv := newTestServer()
	rec := doJSON(t, srv, "POST", "/books", map[string]any{
		"name": "Dune", "availabilityQuantity": 2,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var book models.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &book); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if book.Name != "Dune" || book.AvailabilityQuantity != 2 {
		t.Fatalf("unexpected book: %+v", book)
	}
}

func TestAddBook_MissingNameRejected(t *testing.T) {
	srv := newTestServer()
	rec := doJSON(t, srv, "POST", "/books", map[string]any{"availabilityQuantity": 2})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetBook_NotFound(t *testing.T) {
	srv := newTestServer()
	rec := doJSON(t, srv, "GET", "/books/does-not-exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListBooks_AvailableFilter(t *testing.T) {
	srv := newTestServer()
	doJSON(t, srv, "POST", "/books", map[string]any{"name": "Dune", "availabilityQuantity": 1})
	doJSON(t, srv, "POST", "/books", map[string]any{"name": "Foundation", "availabilityQuantity": 0})

	rec := doJSON(t, srv, "GET", "/books?available=true", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var books []models.Book
	json.Unmarshal(rec.Body.Bytes(), &books)
	if len(books) != 1 || books[0].Name != "Dune" {
		t.Fatalf("expected only Dune, got %+v", books)
	}
}

func TestRegisterUser_Success(t *testing.T) {
	srv := newTestServer()
	rec := doJSON(t, srv, "POST", "/users", map[string]any{"name": "Alice"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var user models.User
	json.Unmarshal(rec.Body.Bytes(), &user)
	if user.Name != "Alice" || user.ID == "" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	srv := newTestServer()
	rec := doJSON(t, srv, "GET", "/users/does-not-exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/httpserver/... -v`
Expected: FAIL — build fails (`s.handleAddBook` etc. undefined), same as Task 4's expected build failure.

- [ ] **Step 3: Implement the book handlers**

`internal/httpserver/books_handler.go`:
```go
package httpserver

import (
	"encoding/json"
	"net/http"

	"go-base-project/internal/repository"
)

type addBookRequest struct {
	Name                 string   `json:"name"`
	AuthorIDs            []string `json:"authorIds"`
	AvailabilityQuantity int      `json:"availabilityQuantity"`
}

func (s *Server) handleAddBook(w http.ResponseWriter, r *http.Request) {
	var req addBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "invalid request body")
		return
	}
	if req.Name == "" {
		writeValidationError(w, "name is required")
		return
	}
	if req.AvailabilityQuantity < 0 {
		writeValidationError(w, "availabilityQuantity must be >= 0")
		return
	}

	book, err := s.svc.AddBook(r.Context(), req.Name, req.AuthorIDs, req.AvailabilityQuantity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	book, err := s.svc.GetBook(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	filter := repository.BookFilter{
		AvailableOnly: r.URL.Query().Get("available") == "true",
		AuthorID:      r.URL.Query().Get("authorId"),
	}
	books, err := s.svc.ListBooks(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}
```

- [ ] **Step 4: Implement the user handlers**

`internal/httpserver/users_handler.go`:
```go
package httpserver

import (
	"encoding/json"
	"net/http"
)

type registerUserRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "invalid request body")
		return
	}
	if req.Name == "" {
		writeValidationError(w, "name is required")
		return
	}

	user, err := s.svc.RegisterUser(r.Context(), req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	user, err := s.svc.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}
```

Note: `routes()` (Task 4) still references `handleBorrow`, `handleReturn`, `handleListBorrowedBooks`, which don't exist until Task 6 — the build (and this task's tests) will only succeed once Task 6 is also done. Proceed directly to Task 6 before attempting to run tests.

- [ ] **Step 5: Commit**

```bash
git add internal/httpserver/responses.go internal/httpserver/server.go internal/httpserver/books_handler.go internal/httpserver/users_handler.go internal/httpserver/books_users_handler_test.go
git commit -m "feat: add HTTP server scaffolding and book/user handlers"
```

(This commits Task 4 and Task 5 together, since Task 4 doesn't build on its own.)

---

### Task 6: Borrow/return HTTP handlers

**Files:**
- Create: `internal/httpserver/borrow_handler.go`
- Test: `internal/httpserver/borrow_handler_test.go`

**Interfaces:**
- Consumes: `Server` (Task 4), `libraryservice.Service` methods `Borrow`, `Return`, `ListBorrowedBooks` (Task 3), `doJSON`/`newTestServer` test helpers (Task 5).
- Produces: `s.handleBorrow`, `s.handleReturn`, `s.handleListBorrowedBooks` — completes the `routes()` table from Task 4, making the whole `httpserver` package buildable.

- [ ] **Step 1: Write the failing tests**

`internal/httpserver/borrow_handler_test.go`:
```go
package httpserver_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"go-base-project/internal/models"
)

func TestBorrowThenReturn_Flow(t *testing.T) {
	srv := newTestServer()

	rec := doJSON(t, srv, "POST", "/users", map[string]any{"name": "Alice"})
	var user models.User
	json.Unmarshal(rec.Body.Bytes(), &user)

	rec = doJSON(t, srv, "POST", "/books", map[string]any{"name": "Dune", "availabilityQuantity": 1})
	var book models.Book
	json.Unmarshal(rec.Body.Bytes(), &book)

	rec = doJSON(t, srv, "POST", "/users/"+user.ID+"/borrow", map[string]any{"bookId": book.ID})
	if rec.Code != http.StatusCreated {
		t.Fatalf("borrow: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, srv, "GET", "/users/"+user.ID+"/books", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list borrowed: expected 200, got %d", rec.Code)
	}
	var borrowed []models.Book
	json.Unmarshal(rec.Body.Bytes(), &borrowed)
	if len(borrowed) != 1 || borrowed[0].ID != book.ID {
		t.Fatalf("expected book in borrowed list, got %+v", borrowed)
	}

	rec = doJSON(t, srv, "POST", "/users/"+user.ID+"/return", map[string]any{"bookId": book.ID})
	if rec.Code != http.StatusOK {
		t.Fatalf("return: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBorrow_BookUnavailableReturns409(t *testing.T) {
	srv := newTestServer()

	rec := doJSON(t, srv, "POST", "/users", map[string]any{"name": "Alice"})
	var user models.User
	json.Unmarshal(rec.Body.Bytes(), &user)

	rec = doJSON(t, srv, "POST", "/books", map[string]any{"name": "Dune", "availabilityQuantity": 0})
	var book models.Book
	json.Unmarshal(rec.Body.Bytes(), &book)

	rec = doJSON(t, srv, "POST", "/users/"+user.ID+"/borrow", map[string]any{"bookId": book.ID})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestReturn_NotBorrowedReturns409(t *testing.T) {
	srv := newTestServer()

	rec := doJSON(t, srv, "POST", "/users", map[string]any{"name": "Alice"})
	var user models.User
	json.Unmarshal(rec.Body.Bytes(), &user)

	rec = doJSON(t, srv, "POST", "/books", map[string]any{"name": "Dune", "availabilityQuantity": 1})
	var book models.Book
	json.Unmarshal(rec.Body.Bytes(), &book)

	rec = doJSON(t, srv, "POST", "/users/"+user.ID+"/return", map[string]any{"bookId": book.ID})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/httpserver/... -v`
Expected: FAIL — build fails (`s.handleBorrow` etc. undefined).

- [ ] **Step 3: Implement the borrow/return handlers**

`internal/httpserver/borrow_handler.go`:
```go
package httpserver

import (
	"encoding/json"
	"net/http"
)

type bookIDRequest struct {
	BookID string `json:"bookId"`
}

func (s *Server) handleBorrow(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req bookIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "invalid request body")
		return
	}
	if req.BookID == "" {
		writeValidationError(w, "bookId is required")
		return
	}

	entry, err := s.svc.Borrow(r.Context(), userID, req.BookID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) handleReturn(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req bookIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "invalid request body")
		return
	}
	if req.BookID == "" {
		writeValidationError(w, "bookId is required")
		return
	}

	if err := s.svc.Return(r.Context(), userID, req.BookID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "returned"})
}

func (s *Server) handleListBorrowedBooks(w http.ResponseWriter, r *http.Request) {
	books, err := s.svc.ListBorrowedBooks(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}
```

- [ ] **Step 4: Run all httpserver tests to verify they pass**

Run: `go test ./internal/httpserver/... -v`
Expected: PASS (all tests across Tasks 5 and 6).

- [ ] **Step 5: Run the full test suite**

Run: `go test -race ./...`
Expected: PASS for `models`, `apperrors`, `repository/memory`, `libraryservice`, `httpserver`. The `postgres` package doesn't exist yet (Task 7).

- [ ] **Step 6: Commit**

```bash
git add internal/httpserver/borrow_handler.go internal/httpserver/borrow_handler_test.go
git commit -m "feat: add borrow/return HTTP handlers, completing the routing table"
```

---

### Task 7: Postgres repository

**Files:**
- Create: `internal/repository/postgres/schema.go`
- Create: `internal/repository/postgres/postgres.go`
- Test: `internal/repository/postgres/postgres_test.go`

**Interfaces:**
- Consumes: `repository.Repository`, `repository.BookFilter` (Task 1); `models.*`, `apperrors.*` (Task 1).
- Produces: `postgres.Connect(ctx, connString string) (*postgres.Repository, error)` and `(*Repository) Close()`, satisfying `repository.Repository` — consumed by `cmd/main.go` (Task 8).

Integration tests connect to a real local Postgres instance (assumed already running) and skip automatically if unreachable, so `go test ./...` doesn't hard-fail on a machine without Postgres running.

- [ ] **Step 1: Write the failing tests**

`internal/repository/postgres/postgres_test.go`:
```go
package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-base-project/internal/apperrors"
	"go-base-project/internal/models"
	"go-base-project/internal/repository/postgres"
)

func connString() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
}

func newTestRepo(t *testing.T) *postgres.Repository {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repo, err := postgres.Connect(ctx, connString())
	if err != nil {
		t.Skipf("skipping: postgres not reachable at %s: %v", connString(), err)
	}
	return repo
}

func TestCreateAndGetUser(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.Close()
	ctx := context.Background()

	user := &models.User{ID: uuid.New().String(), Name: "Alice", CreatedAt: time.Now()}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := repo.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Name != "Alice" {
		t.Fatalf("expected Alice, got %q", got.Name)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.Close()

	if _, err := repo.GetUser(context.Background(), uuid.New().String()); err != apperrors.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestBorrowAndReturn(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.Close()
	ctx := context.Background()

	user := &models.User{ID: uuid.New().String(), Name: "Bob", CreatedAt: time.Now()}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	book := &models.Book{ID: uuid.New().String(), Name: "Dune", AvailabilityQuantity: 1, CreatedAt: time.Now()}
	if err := repo.CreateBook(ctx, book, nil); err != nil {
		t.Fatalf("CreateBook: %v", err)
	}

	if _, err := repo.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow: %v", err)
	}

	got, err := repo.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.AvailabilityQuantity != 0 {
		t.Fatalf("expected availability 0, got %d", got.AvailabilityQuantity)
	}

	if err := repo.Return(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Return: %v", err)
	}

	got, err = repo.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("expected availability 1 after return, got %d", got.AvailabilityQuantity)
	}

	if err := repo.Return(ctx, user.ID, book.ID); err != apperrors.ErrNotBorrowed {
		t.Fatalf("expected ErrNotBorrowed on double return, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/repository/postgres/... -v`
Expected: FAIL — package `postgres` does not exist yet.

- [ ] **Step 3: Write the schema**

`internal/repository/postgres/schema.go`:
```go
package postgres

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS authors (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS books (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	availability_quantity INT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS author_book_mappings (
	id UUID PRIMARY KEY,
	book_id UUID NOT NULL REFERENCES books(id),
	author_id UUID NOT NULL REFERENCES authors(id)
);

CREATE TABLE IF NOT EXISTS library_ledger (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id),
	book_id UUID NOT NULL REFERENCES books(id),
	action_type TEXT NOT NULL,
	returned_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);
`
```

- [ ] **Step 4: Implement the Postgres repository**

`internal/repository/postgres/postgres.go`:
```go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-base-project/internal/apperrors"
	"go-base-project/internal/models"
	"go-base-project/internal/repository"
)

type Repository struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, connString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, created_at) VALUES ($1, $2, $3)`,
		user.ID, user.Name, user.CreatedAt)
	return err
}

func (r *Repository) GetUser(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateBook(ctx context.Context, book *models.Book, authorIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`INSERT INTO books (id, name, availability_quantity, created_at) VALUES ($1, $2, $3, $4)`,
		book.ID, book.Name, book.AvailabilityQuantity, book.CreatedAt,
	); err != nil {
		return err
	}

	for _, authorID := range authorIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO author_book_mappings (id, book_id, author_id) VALUES ($1, $2, $3)`,
			uuid.New().String(), book.ID, authorID,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetBook(ctx context.Context, id string) (*models.Book, error) {
	var b models.Book
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, availability_quantity, created_at FROM books WHERE id = $1`, id,
	).Scan(&b.ID, &b.Name, &b.AvailabilityQuantity, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrBookNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) ListBooks(ctx context.Context, filter repository.BookFilter) ([]*models.Book, error) {
	query := `SELECT DISTINCT b.id, b.name, b.availability_quantity, b.created_at FROM books b`
	var args []any
	var conditions []string

	if filter.AuthorID != "" {
		query += ` JOIN author_book_mappings m ON m.book_id = b.id`
		args = append(args, filter.AuthorID)
		conditions = append(conditions, fmt.Sprintf("m.author_id = $%d", len(args)))
	}
	if filter.AvailableOnly {
		conditions = append(conditions, "b.availability_quantity > 0")
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Name, &b.AvailabilityQuantity, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, &b)
	}
	return books, rows.Err()
}

func (r *Repository) Borrow(ctx context.Context, userID, bookID string) (*models.LibraryLedgerEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var userExists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID,
	).Scan(&userExists); err != nil {
		return nil, err
	}
	if !userExists {
		return nil, apperrors.ErrUserNotFound
	}

	var availability int
	err = tx.QueryRow(ctx,
		`SELECT availability_quantity FROM books WHERE id = $1 FOR UPDATE`, bookID,
	).Scan(&availability)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrBookNotFound
	}
	if err != nil {
		return nil, err
	}

	var activeCount int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM library_ledger WHERE user_id = $1 AND returned_at IS NULL`, userID,
	).Scan(&activeCount); err != nil {
		return nil, err
	}
	if activeCount >= 3 {
		return nil, apperrors.ErrBorrowLimitReached
	}

	var alreadyBorrowed bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM library_ledger WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL)`,
		userID, bookID,
	).Scan(&alreadyBorrowed); err != nil {
		return nil, err
	}
	if alreadyBorrowed {
		return nil, apperrors.ErrAlreadyBorrowed
	}

	if availability <= 0 {
		return nil, apperrors.ErrBookUnavailable
	}

	if _, err := tx.Exec(ctx,
		`UPDATE books SET availability_quantity = availability_quantity - 1 WHERE id = $1`, bookID,
	); err != nil {
		return nil, err
	}

	now := time.Now()
	entry := &models.LibraryLedgerEntry{
		ID:         uuid.New().String(),
		UserID:     userID,
		BookID:     bookID,
		ActionType: models.ActionBorrow,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO library_ledger (id, user_id, book_id, action_type, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.UserID, entry.BookID, entry.ActionType, entry.CreatedAt, entry.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *Repository) Return(ctx context.Context, userID, bookID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var ledgerID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM library_ledger WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL`,
		userID, bookID,
	).Scan(&ledgerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotBorrowed
	}
	if err != nil {
		return err
	}

	now := time.Now()
	if _, err := tx.Exec(ctx,
		`UPDATE library_ledger SET returned_at = $1, updated_at = $1 WHERE id = $2`, now, ledgerID,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE books SET availability_quantity = availability_quantity + 1 WHERE id = $1`, bookID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) ListBorrowedBooks(ctx context.Context, userID string) ([]*models.Book, error) {
	var userExists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID,
	).Scan(&userExists); err != nil {
		return nil, err
	}
	if !userExists {
		return nil, apperrors.ErrUserNotFound
	}

	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.name, b.availability_quantity, b.created_at
		 FROM books b
		 JOIN library_ledger l ON l.book_id = b.id
		 WHERE l.user_id = $1 AND l.returned_at IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Name, &b.AvailabilityQuantity, &b.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, &b)
	}
	return books, rows.Err()
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/repository/postgres/... -v`
Expected: PASS if a local Postgres is reachable at `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable` (or `$DATABASE_URL` if set); otherwise all tests report `SKIP: skipping: postgres not reachable...` — that's an acceptable outcome for this task, not a failure.

- [ ] **Step 6: Verify the full module builds**

Run: `go build ./... && go vet ./...`
Expected: no output (success).

- [ ] **Step 7: Commit**

```bash
git add internal/repository/postgres
git commit -m "feat: implement Postgres repository with transactional borrow/return"
```

---

### Task 8: Wire everything together in cmd/main.go

**Files:**
- Modify: `cmd/main.go` (currently empty)

**Interfaces:**
- Consumes: `postgres.Connect` (Task 7), `libraryservice.New` (Task 3), `httpserver.New` (Task 4).
- Produces: the runnable binary. No new interfaces for later tasks (this is the top of the dependency graph).

No automated test for `main` (standard for a thin wiring entrypoint) — verified manually by building and running.

- [ ] **Step 1: Write cmd/main.go**

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-base-project/internal/httpserver"
	"go-base-project/internal/libraryservice"
	"go-base-project/internal/repository/postgres"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	repo, err := postgres.Connect(connectCtx, connString)
	cancel()
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer repo.Close()

	svc := libraryservice.New(repo)
	srv := httpserver.New(svc)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	httpSrv := &http.Server{Addr: addr, Handler: srv}

	go func() {
		log.Printf("listening on %s", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
```

- [ ] **Step 2: Build the binary**

Run: `go build ./...`
Expected: no output (success).

- [ ] **Step 3: Manual smoke test (requires local Postgres running)**

Run: `go run ./cmd &` then, once "listening on :8080" appears in the logs:
```bash
curl -s -X POST localhost:8080/users -d '{"name":"Alice"}' | tee /tmp/user.json
curl -s -X POST localhost:8080/books -d '{"name":"Dune","availabilityQuantity":1}' | tee /tmp/book.json
USER_ID=$(jq -r .id /tmp/user.json)
BOOK_ID=$(jq -r .id /tmp/book.json)
curl -s -X POST localhost:8080/users/$USER_ID/borrow -d "{\"bookId\":\"$BOOK_ID\"}"
curl -s localhost:8080/users/$USER_ID/books
curl -s -X POST localhost:8080/users/$USER_ID/return -d "{\"bookId\":\"$BOOK_ID\"}"
```
Expected: each step returns 2xx JSON; the borrowed-books list shows the book before the return call and is empty after. Stop the server afterward (`kill %1` or Ctrl+C in the foreground).

- [ ] **Step 4: Run the entire test suite one final time**

Run: `go test -race ./...`
Expected: PASS across all packages (Postgres tests PASS if local Postgres is reachable, SKIP otherwise).

- [ ] **Step 5: Commit**

```bash
git add cmd/main.go
git commit -m "feat: wire library management HTTP server in cmd/main.go"
```

---

## Self-Review Notes

- **Spec coverage:** add book / get book / list-with-filters (Task 5), register user / get user (Task 5), borrow / return / list-borrowed-by-user (Task 6), Postgres storage (Task 7), mutex locking (Task 3), TDD ordering (every task writes tests before implementation), >3-books and one-book-one-user-at-a-time concurrency tests (Task 3) — all covered.
- **Type consistency:** `repository.Repository` (Task 1) method signatures match exactly what `memory.Repository` (Task 2) and `postgres.Repository` (Task 7) implement, and what `libraryservice.Service` (Task 3) calls. `libraryservice.Service` method signatures match exactly what `httpserver` handlers (Tasks 5-6) call.
- **Known limitation (intentional, per design doc):** locking is in-process only (`sync.Mutex`), so correctness guarantees only hold for a single running instance of this server — acceptable and explicitly requested for this project.
