package libraryservice

import (
	"context"
	"time"

	"go-base-project/internal/models"

	"github.com/xtgo/uuid"
)

// Repository contains non-transactional data access used by Service.
// It lives in the consumer package so storage implementations only depend on
// the contract the service actually needs.
type Repository interface {
	WithTransaction(context.Context, func(Transaction) error) error

	CreateUser(context.Context, models.User) error
	GetUser(context.Context, uuid.UUID) (models.User, error)
	CreateBook(context.Context, models.Book, []uuid.UUID) error
	GetBook(context.Context, uuid.UUID) (models.Book, error)
	ListBooks(context.Context, models.BookFilter) ([]models.Book, error)
	ListBorrowedBooks(context.Context, uuid.UUID) ([]models.Book, error)
}

type Transaction interface {
	GetUser(context.Context, uuid.UUID) (models.User, error)
	GetBook(context.Context, uuid.UUID) (models.Book, error)
	ActiveBorrowCount(context.Context, uuid.UUID) (int, error)
	FindActiveLedger(context.Context, uuid.UUID, uuid.UUID) (models.LibraryLedgerEntry, error)
	UpdateBookAvailability(context.Context, uuid.UUID, int) error
	InsertLedger(context.Context, models.LibraryLedgerEntry) error
	MarkReturned(context.Context, uuid.UUID, time.Time) error
}
