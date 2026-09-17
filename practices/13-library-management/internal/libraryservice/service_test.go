package libraryservice_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"go-base-project/internal/libraryservice"
	"go-base-project/internal/models"
	"go-base-project/internal/repository/memory"
)

func TestBorrowSucceedsAndDecrementsAvailability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, repository := newService()
	user := createUser(t, ctx, service, "Ada")
	book := createBook(t, ctx, service, "Concurrency in Go", 1)

	if err := service.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow() error = %v", err)
	}

	got, err := repository.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook() error = %v", err)
	}
	if got.AvailabilityQuantity != 0 {
		t.Fatalf("availability = %d, want 0", got.AvailabilityQuantity)
	}
}

func TestBorrowRejectsFourthActiveBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, _ := newService()
	user := createUser(t, ctx, service, "Grace")
	books := make([]models.Book, 4)
	for index := range books {
		books[index] = createBook(t, ctx, service, "Book", 1)
	}
	for _, book := range books[:3] {
		if err := service.Borrow(ctx, user.ID, book.ID); err != nil {
			t.Fatalf("Borrow() setup error = %v", err)
		}
	}

	err := service.Borrow(ctx, user.ID, books[3].ID)
	if !errors.Is(err, libraryservice.ErrBorrowLimit) {
		t.Fatalf("Borrow() error = %v, want ErrBorrowLimit", err)
	}
}

func TestBorrowRejectsDuplicateBookForUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, repository := newService()
	user := createUser(t, ctx, service, "Linus")
	book := createBook(t, ctx, service, "Operating Systems", 2)

	if err := service.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow() setup error = %v", err)
	}
	err := service.Borrow(ctx, user.ID, book.ID)
	if !errors.Is(err, libraryservice.ErrAlreadyBorrowed) {
		t.Fatalf("Borrow() error = %v, want ErrAlreadyBorrowed", err)
	}
	got, getErr := repository.GetBook(ctx, book.ID)
	if getErr != nil {
		t.Fatalf("GetBook() error = %v", getErr)
	}
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("availability = %d, want rollback to preserve 1", got.AvailabilityQuantity)
	}
}

func TestConcurrentBorrowOfLastCopyAllowsExactlyOne(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, repository := newService()
	firstUser := createUser(t, ctx, service, "First")
	secondUser := createUser(t, ctx, service, "Second")
	book := createBook(t, ctx, service, "One Copy", 1)

	start := make(chan struct{})
	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for _, userID := range []models.User{firstUser, secondUser} {
		waitGroup.Add(1)
		go func(user models.User) {
			defer waitGroup.Done()
			<-start
			results <- service.Borrow(ctx, user.ID, book.ID)
		}(userID)
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	unavailable := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, libraryservice.ErrBookUnavailable):
			unavailable++
		default:
			t.Fatalf("unexpected Borrow() error = %v", err)
		}
	}
	if successes != 1 || unavailable != 1 {
		t.Fatalf("results: successes=%d unavailable=%d, want 1 and 1", successes, unavailable)
	}
	got, err := repository.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook() error = %v", err)
	}
	if got.AvailabilityQuantity != 0 {
		t.Fatalf("availability = %d, want 0", got.AvailabilityQuantity)
	}
}

func TestReturnRestoresAvailabilityAndAllowsBorrowAgain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, repository := newService()
	user := createUser(t, ctx, service, "Margaret")
	book := createBook(t, ctx, service, "Compilers", 1)
	if err := service.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow() setup error = %v", err)
	}

	if err := service.Return(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Return() error = %v", err)
	}
	got, err := repository.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetBook() error = %v", err)
	}
	if got.AvailabilityQuantity != 1 {
		t.Fatalf("availability = %d, want 1", got.AvailabilityQuantity)
	}
	if err = service.Borrow(ctx, user.ID, book.ID); err != nil {
		t.Fatalf("Borrow() after return error = %v", err)
	}
}

func TestReturnRejectsBookNotBorrowedByUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, _ := newService()
	user := createUser(t, ctx, service, "Barbara")
	book := createBook(t, ctx, service, "Networks", 1)

	err := service.Return(ctx, user.ID, book.ID)
	if !errors.Is(err, libraryservice.ErrNotBorrowed) {
		t.Fatalf("Return() error = %v, want ErrNotBorrowed", err)
	}
}

func TestListBorrowedBooksOnlyIncludesActiveEntries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, _ := newService()
	user := createUser(t, ctx, service, "Ken")
	returned := createBook(t, ctx, service, "Returned", 1)
	active := createBook(t, ctx, service, "Active", 1)
	for _, book := range []models.Book{returned, active} {
		if err := service.Borrow(ctx, user.ID, book.ID); err != nil {
			t.Fatalf("Borrow() setup error = %v", err)
		}
	}
	if err := service.Return(ctx, user.ID, returned.ID); err != nil {
		t.Fatalf("Return() setup error = %v", err)
	}

	books, err := service.ListBorrowedBooks(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListBorrowedBooks() error = %v", err)
	}
	if len(books) != 1 || books[0].ID != active.ID {
		t.Fatalf("ListBorrowedBooks() = %+v, want only active book %s", books, active.ID)
	}
}

func newService() (*libraryservice.Service, *memory.Repository) {
	repository := memory.New()
	return libraryservice.New(repository), repository
}

func createUser(
	t *testing.T,
	ctx context.Context,
	service *libraryservice.Service,
	name string,
) models.User {
	t.Helper()
	user, err := service.CreateUser(ctx, name)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	return user
}

func createBook(
	t *testing.T,
	ctx context.Context,
	service *libraryservice.Service,
	name string,
	quantity int,
) models.Book {
	t.Helper()
	book, err := service.CreateBook(ctx, name, nil, quantity)
	if err != nil {
		t.Fatalf("CreateBook() error = %v", err)
	}
	return book
}
