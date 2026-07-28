# Library Book Management System — Design

Status: Approved
Date: 2026-07-28

## Objective

Backend service (Go + Postgres) to manage books, users, and borrow/return
workflows, per `readme.md`.

## Constraints

- A user may hold at most 3 concurrently-borrowed books.
- A user may not hold 2 copies of the same BookID at once.
- Books have multiple copies, tracked as `AvailabilityQuantity`. Borrowing a
  book decrements the count; returning increments it. A borrow fails if the
  count is already 0.
- Concurrency safety is enforced with in-memory Go mutexes (`sync.Mutex`),
  scoped to a single process — not distributed-safe. Acceptable for this
  project per explicit instruction.

## Architecture & folder layout

```
go-base-project/
  cmd/main.go                  # wiring: DB connection, repo, service, http server, graceful shutdown
  internal/
    models/                    # plain structs + enums, no behavior
      book.go  author.go  user.go  ledger.go
    repository/                # data access
      repository.go            # interface(s) consumed by libraryservice
      postgres/                 # pgx-backed implementation
      memory/                   # in-memory fake used only by unit tests
    libraryservice/            # business rules + locking
      service.go  borrow.go  return.go  books.go  users.go  locks.go
    httpserver/                # HTTP layer
      server.go  routes.go  books_handler.go  users_handler.go  borrow_handler.go
```

`internal/` keeps these packages non-importable outside the module. The
repository interface(s) are defined in `libraryservice` (consumer defines the
interface); `postgres` and `memory` implementations both satisfy it.

## Data model

Postgres schema (assumes a local Postgres instance is already running;
`cmd/main.go` reads `DATABASE_URL` env var with a localhost default, and runs
`CREATE TABLE IF NOT EXISTS` statements on startup — no migration tooling):

```sql
users(id UUID PK, name TEXT, created_at TIMESTAMPTZ)
authors(id UUID PK, name TEXT)
books(id UUID PK, name TEXT, availability_quantity INT, created_at TIMESTAMPTZ)
author_book_mappings(id UUID PK, book_id FK, author_id FK)
library_ledger(
  id UUID PK, user_id FK, book_id FK,
  action_type TEXT,        -- 'BORROW' (only stored value; return is an update, see below)
  returned_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
```

A borrow inserts one `library_ledger` row with `returned_at = NULL`. A return
**updates that same row's** `returned_at` rather than inserting a second row.
This makes "currently borrowed books for a user" a simple
`WHERE user_id = ? AND returned_at IS NULL` query, and "does this user already
hold this book" a similar per-book check.

Go structs in `internal/models` mirror these tables 1:1: `Book`, `Author`,
`AuthorBookMapping`, `User`, `LibraryLedgerEntry`, plus a small `ActionType`
enum (currently just `Borrow`).

Driver: `pgx`. HTTP routing: standard library `net/http` (Go 1.22+ ServeMux
method + path-param routing), no third-party router.

## Business rules & locking (in `libraryservice`)

A `lockManager` holds `map[uuid]*sync.Mutex` for users and a separate one for
books, guarded by a small internal mutex for map/entry creation.

- **Borrow(userID, bookID)**: lock user → lock book → (in one DB transaction)
  count user's active borrows (reject if ≥ 3) → check no existing active
  ledger row for this exact user+book (reject if found) → check
  `availability_quantity > 0` (reject if not) → decrement quantity → insert
  ledger row → commit → unlock book → unlock user.
- **Return(userID, bookID)**: lock user → lock book → find the active ledger
  row for user+book (error if none) → set `returned_at` → increment quantity
  → commit → unlock.
- Locks are always acquired user-then-book and released in reverse order, to
  avoid deadlock ordering issues.

Business errors (not found, capacity exceeded, unavailable, duplicate-borrow)
are typed sentinel errors in `libraryservice`, mapped to HTTP status codes in
`httpserver`. The service layer stays HTTP-agnostic.

## HTTP API

| Method & Path | Purpose |
|---|---|
| `POST /books` | Add a new book (name, authorIDs, availabilityQuantity) |
| `GET /books/{id}` | Get book details by ID |
| `GET /books?available=true&author={id}` | List books with filters |
| `POST /users` | Register a new user (name) |
| `GET /users/{id}` | Fetch user details |
| `POST /users/{id}/borrow` | Borrow a book (body: bookId) |
| `POST /users/{id}/return` | Return a book (body: bookId) |
| `GET /users/{id}/books` | List books currently borrowed by user |

## Testing strategy (TDD)

- `libraryservice` unit tests run against `internal/repository/memory` (an
  in-memory fake) — fast, no DB required. Written before implementation,
  covering at minimum:
  - Borrow succeeds and decrements availability.
  - 4th borrow by the same user is rejected (3-book limit).
  - Borrowing the same BookID twice by the same user is rejected.
  - Concurrent borrows on a book with `availability_quantity = 1` from two
    different users — exactly one succeeds (goroutines + WaitGroup,
    `go test -race`).
  - Return succeeds, increments availability, book becomes borrowable again.
  - Returning a book not currently borrowed by that user errors.
  - Listing a user's borrowed books reflects only active (non-returned)
    ledger entries.
- `httpserver` handler tests use `httptest` against a fake
  `libraryservice`-shaped interface (table-driven).
- `postgres` repository gets a small integration test suite, skipped
  automatically if it can't connect, so `go test ./...` doesn't hard-fail
  without Postgres running.
