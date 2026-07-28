package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go-base-project/internal/libraryservice"
	"go-base-project/internal/models"

	"github.com/xtgo/uuid"
)

type Repository struct {
	mu sync.RWMutex

	users       map[uuid.UUID]models.User
	books       map[uuid.UUID]models.Book
	authors     map[uuid.UUID]models.Author
	bookAuthors map[uuid.UUID][]uuid.UUID
	ledger      map[uuid.UUID]models.LibraryLedgerEntry
}

func New() *Repository {
	return &Repository{
		users:       make(map[uuid.UUID]models.User),
		books:       make(map[uuid.UUID]models.Book),
		authors:     make(map[uuid.UUID]models.Author),
		bookAuthors: make(map[uuid.UUID][]uuid.UUID),
		ledger:      make(map[uuid.UUID]models.LibraryLedgerEntry),
	}
}

// AddAuthor supports service tests and local development until author
// management is exposed by the public API.
func (r *Repository) AddAuthor(author models.Author) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.authors[author.ID] = author
}

func (r *Repository) CreateUser(ctx context.Context, user models.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return models.User{}, fmt.Errorf("%w: user %s", libraryservice.ErrNotFound, id)
	}
	return user, nil
}

func (r *Repository) CreateBook(
	ctx context.Context,
	book models.Book,
	authorIDs []uuid.UUID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, authorID := range authorIDs {
		if _, ok := r.authors[authorID]; !ok {
			return fmt.Errorf("%w: author %s", libraryservice.ErrNotFound, authorID)
		}
	}
	r.books[book.ID] = book
	r.bookAuthors[book.ID] = append([]uuid.UUID(nil), authorIDs...)
	return nil
}

func (r *Repository) GetBook(ctx context.Context, id uuid.UUID) (models.Book, error) {
	if err := ctx.Err(); err != nil {
		return models.Book{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	book, ok := r.books[id]
	if !ok {
		return models.Book{}, fmt.Errorf("%w: book %s", libraryservice.ErrNotFound, id)
	}
	return book, nil
}

func (r *Repository) ListBooks(
	ctx context.Context,
	filter models.BookFilter,
) ([]models.Book, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	books := make([]models.Book, 0, len(r.books))
	for _, book := range r.books {
		if filter.Available != nil {
			isAvailable := book.AvailabilityQuantity > 0
			if isAvailable != *filter.Available {
				continue
			}
		}
		if filter.AuthorID != nil && !contains(r.bookAuthors[book.ID], *filter.AuthorID) {
			continue
		}
		books = append(books, book)
	}
	sortBooks(books)
	return books, nil
}

func (r *Repository) ListBorrowedBooks(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Book, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	books := make([]models.Book, 0)
	for _, entry := range r.ledger {
		if entry.UserID == userID && entry.ReturnedAt == nil {
			if book, ok := r.books[entry.BookID]; ok {
				books = append(books, book)
			}
		}
	}
	sortBooks(books)
	return books, nil
}

func (r *Repository) WithTransaction(
	ctx context.Context,
	fn func(libraryservice.Transaction) error,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	tx := &transaction{
		users:  cloneMap(r.users),
		books:  cloneMap(r.books),
		ledger: cloneLedger(r.ledger),
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r.users = tx.users
	r.books = tx.books
	r.ledger = tx.ledger
	return nil
}

type transaction struct {
	users  map[uuid.UUID]models.User
	books  map[uuid.UUID]models.Book
	ledger map[uuid.UUID]models.LibraryLedgerEntry
}

func (tx *transaction) GetUser(
	_ context.Context,
	id uuid.UUID,
) (models.User, error) {
	user, ok := tx.users[id]
	if !ok {
		return models.User{}, fmt.Errorf("%w: user %s", libraryservice.ErrNotFound, id)
	}
	return user, nil
}

func (tx *transaction) GetBook(
	_ context.Context,
	id uuid.UUID,
) (models.Book, error) {
	book, ok := tx.books[id]
	if !ok {
		return models.Book{}, fmt.Errorf("%w: book %s", libraryservice.ErrNotFound, id)
	}
	return book, nil
}

func (tx *transaction) ActiveBorrowCount(
	_ context.Context,
	userID uuid.UUID,
) (int, error) {
	count := 0
	for _, entry := range tx.ledger {
		if entry.UserID == userID && entry.ReturnedAt == nil {
			count++
		}
	}
	return count, nil
}

func (tx *transaction) FindActiveLedger(
	_ context.Context,
	userID uuid.UUID,
	bookID uuid.UUID,
) (models.LibraryLedgerEntry, error) {
	for _, entry := range tx.ledger {
		if entry.UserID == userID && entry.BookID == bookID && entry.ReturnedAt == nil {
			return entry, nil
		}
	}
	return models.LibraryLedgerEntry{}, libraryservice.ErrNotFound
}

func (tx *transaction) UpdateBookAvailability(
	_ context.Context,
	bookID uuid.UUID,
	delta int,
) error {
	book, ok := tx.books[bookID]
	if !ok {
		return fmt.Errorf("%w: book %s", libraryservice.ErrNotFound, bookID)
	}
	if book.AvailabilityQuantity+delta < 0 {
		return libraryservice.ErrBookUnavailable
	}
	book.AvailabilityQuantity += delta
	tx.books[bookID] = book
	return nil
}

func (tx *transaction) InsertLedger(
	_ context.Context,
	entry models.LibraryLedgerEntry,
) error {
	tx.ledger[entry.ID] = entry
	return nil
}

func (tx *transaction) MarkReturned(
	_ context.Context,
	entryID uuid.UUID,
	returnedAt time.Time,
) error {
	entry, ok := tx.ledger[entryID]
	if !ok {
		return fmt.Errorf("%w: ledger entry %s", libraryservice.ErrNotFound, entryID)
	}
	entry.ReturnedAt = &returnedAt
	entry.UpdatedAt = returnedAt
	tx.ledger[entryID] = entry
	return nil
}

func contains(ids []uuid.UUID, target uuid.UUID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func sortBooks(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		if books[i].CreatedAt.Equal(books[j].CreatedAt) {
			return books[i].ID.String() < books[j].ID.String()
		}
		return books[i].CreatedAt.Before(books[j].CreatedAt)
	})
}

func cloneMap[K comparable, V any](source map[K]V) map[K]V {
	cloned := make(map[K]V, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneLedger(
	source map[uuid.UUID]models.LibraryLedgerEntry,
) map[uuid.UUID]models.LibraryLedgerEntry {
	cloned := make(map[uuid.UUID]models.LibraryLedgerEntry, len(source))
	for key, value := range source {
		if value.ReturnedAt != nil {
			returnedAt := *value.ReturnedAt
			value.ReturnedAt = &returnedAt
		}
		cloned[key] = value
	}
	return cloned
}
