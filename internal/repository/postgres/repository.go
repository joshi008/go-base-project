package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-base-project/internal/libraryservice"
	"go-base-project/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xtgo/uuid"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) InitSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY, name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL
		);
		CREATE TABLE IF NOT EXISTS authors (
			id UUID PRIMARY KEY, name TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS books (
			id UUID PRIMARY KEY, name TEXT NOT NULL,
			availability_quantity INT NOT NULL CHECK (availability_quantity >= 0),
			created_at TIMESTAMPTZ NOT NULL
		);
		CREATE TABLE IF NOT EXISTS author_book_mappings (
			id UUID PRIMARY KEY,
			book_id UUID NOT NULL REFERENCES books(id),
			author_id UUID NOT NULL REFERENCES authors(id),
			UNIQUE (book_id, author_id)
		);
		CREATE TABLE IF NOT EXISTS library_ledger (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			book_id UUID NOT NULL REFERENCES books(id),
			action_type TEXT NOT NULL CHECK (action_type = 'BORROW'),
			returned_at TIMESTAMPTZ NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS library_ledger_active_user_idx
			ON library_ledger(user_id) WHERE returned_at IS NULL;
		CREATE UNIQUE INDEX IF NOT EXISTS library_ledger_active_user_book_idx
			ON library_ledger(user_id, book_id) WHERE returned_at IS NULL;
	`)
	if err != nil {
		return fmt.Errorf("initialize schema: %w", err)
	}
	return nil
}

func (r *Repository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, created_at) VALUES ($1, $2, $3)`,
		user.ID.String(), user.Name, user.CreatedAt)
	return err
}

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT id::text, name, created_at FROM users WHERE id = $1`, id.String()))
}

func (r *Repository) CreateBook(
	ctx context.Context,
	book models.Book,
	authorIDs []uuid.UUID,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `
		INSERT INTO books (id, name, availability_quantity, created_at)
		VALUES ($1, $2, $3, $4)`,
		book.ID.String(), book.Name, book.AvailabilityQuantity, book.CreatedAt); err != nil {
		return err
	}
	for _, authorID := range authorIDs {
		if _, err = tx.Exec(ctx, `
			INSERT INTO author_book_mappings (id, book_id, author_id)
			VALUES ($1, $2, $3)`,
			uuid.NewRandom().String(), book.ID.String(), authorID.String()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetBook(ctx context.Context, id uuid.UUID) (models.Book, error) {
	return scanBook(r.pool.QueryRow(ctx, `
		SELECT id::text, name, availability_quantity, created_at
		FROM books WHERE id = $1`, id.String()))
}

func (r *Repository) ListBooks(
	ctx context.Context,
	filter models.BookFilter,
) ([]models.Book, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id::text, b.name, b.availability_quantity, b.created_at
		FROM books b
		WHERE ($1::boolean IS NULL OR (b.availability_quantity > 0) = $1)
		  AND ($2::uuid IS NULL OR EXISTS (
			SELECT 1 FROM author_book_mappings abm
			WHERE abm.book_id = b.id AND abm.author_id = $2
		  ))
		ORDER BY b.created_at, b.id`,
		filter.Available, optionalUUID(filter.AuthorID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBooks(rows)
}

func (r *Repository) ListBorrowedBooks(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Book, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id::text, b.name, b.availability_quantity, b.created_at
		FROM books b
		JOIN library_ledger l ON l.book_id = b.id
		WHERE l.user_id = $1 AND l.returned_at IS NULL
		ORDER BY b.created_at, b.id`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBooks(rows)
}

func (r *Repository) WithTransaction(
	ctx context.Context,
	fn func(libraryservice.Transaction) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = fn(&transaction{tx: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type transaction struct {
	tx pgx.Tx
}

func (t *transaction) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	return scanUser(t.tx.QueryRow(ctx,
		`SELECT id::text, name, created_at FROM users WHERE id = $1`, id.String()))
}

func (t *transaction) GetBook(ctx context.Context, id uuid.UUID) (models.Book, error) {
	return scanBook(t.tx.QueryRow(ctx, `
		SELECT id::text, name, availability_quantity, created_at
		FROM books WHERE id = $1`, id.String()))
}

func (t *transaction) ActiveBorrowCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := t.tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM library_ledger
		WHERE user_id = $1 AND returned_at IS NULL`, userID.String()).Scan(&count)
	return count, err
}

func (t *transaction) FindActiveLedger(
	ctx context.Context,
	userID uuid.UUID,
	bookID uuid.UUID,
) (models.LibraryLedgerEntry, error) {
	var entry models.LibraryLedgerEntry
	var id, scannedUserID, scannedBookID, action string
	err := t.tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, book_id::text, action_type,
		       returned_at, created_at, updated_at
		FROM library_ledger
		WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL`,
		userID.String(), bookID.String()).Scan(
		&id, &scannedUserID, &scannedBookID, &action, &entry.ReturnedAt,
		&entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		return models.LibraryLedgerEntry{}, normalizeNotFound(err)
	}
	entry.ID, err = uuid.Parse(id)
	if err != nil {
		return models.LibraryLedgerEntry{}, err
	}
	entry.UserID, _ = uuid.Parse(scannedUserID)
	entry.BookID, _ = uuid.Parse(scannedBookID)
	entry.ActionType = models.ActionType(action)
	return entry, nil
}

func (t *transaction) UpdateBookAvailability(
	ctx context.Context,
	bookID uuid.UUID,
	delta int,
) error {
	tag, err := t.tx.Exec(ctx, `
		UPDATE books SET availability_quantity = availability_quantity + $2
		WHERE id = $1 AND availability_quantity + $2 >= 0`,
		bookID.String(), delta)
	if err == nil && tag.RowsAffected() == 0 {
		return libraryservice.ErrBookUnavailable
	}
	return err
}

func (t *transaction) InsertLedger(
	ctx context.Context,
	entry models.LibraryLedgerEntry,
) error {
	_, err := t.tx.Exec(ctx, `
		INSERT INTO library_ledger
			(id, user_id, book_id, action_type, returned_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID.String(), entry.UserID.String(), entry.BookID.String(),
		entry.ActionType, entry.ReturnedAt, entry.CreatedAt, entry.UpdatedAt)
	return err
}

func (t *transaction) MarkReturned(
	ctx context.Context,
	entryID uuid.UUID,
	returnedAt time.Time,
) error {
	tag, err := t.tx.Exec(ctx, `
		UPDATE library_ledger SET returned_at = $2, updated_at = $2
		WHERE id = $1 AND returned_at IS NULL`, entryID.String(), returnedAt)
	if err == nil && tag.RowsAffected() == 0 {
		return libraryservice.ErrNotBorrowed
	}
	return err
}

type rowScanner interface {
	Scan(...any) error
}

func scanUser(row rowScanner) (models.User, error) {
	var user models.User
	var id string
	if err := row.Scan(&id, &user.Name, &user.CreatedAt); err != nil {
		return models.User{}, normalizeNotFound(err)
	}
	parsed, err := uuid.Parse(id)
	user.ID = parsed
	return user, err
}

func scanBook(row rowScanner) (models.Book, error) {
	var book models.Book
	var id string
	if err := row.Scan(&id, &book.Name, &book.AvailabilityQuantity, &book.CreatedAt); err != nil {
		return models.Book{}, normalizeNotFound(err)
	}
	parsed, err := uuid.Parse(id)
	book.ID = parsed
	return book, err
}

func collectBooks(rows pgx.Rows) ([]models.Book, error) {
	books := make([]models.Book, 0)
	for rows.Next() {
		book, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, rows.Err()
}

func optionalUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func normalizeNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return libraryservice.ErrNotFound
	}
	return err
}
