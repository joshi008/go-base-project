package libraryservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-base-project/internal/models"

	"github.com/xtgo/uuid"
)

const maxActiveBorrows = 3

type Service struct {
	repository Repository
	locks      *lockManager
	now        func() time.Time
}

func New(repository Repository) *Service {
	return &Service{
		repository: repository,
		locks:      newLockManager(),
		now:        time.Now,
	}
}

func (s *Service) CreateUser(ctx context.Context, name string) (models.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.User{}, fmt.Errorf("%w: user name is required", ErrInvalidInput)
	}
	user := models.User{ID: uuid.NewRandom(), Name: name, CreatedAt: s.now().UTC()}
	if err := s.repository.CreateUser(ctx, user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	return s.repository.GetUser(ctx, id)
}

func (s *Service) CreateBook(
	ctx context.Context,
	name string,
	authorIDs []uuid.UUID,
	availabilityQuantity int,
) (models.Book, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Book{}, fmt.Errorf("%w: book name is required", ErrInvalidInput)
	}
	if availabilityQuantity < 0 {
		return models.Book{}, fmt.Errorf("%w: availability quantity cannot be negative", ErrInvalidInput)
	}
	book := models.Book{
		ID:                   uuid.NewRandom(),
		Name:                 name,
		AvailabilityQuantity: availabilityQuantity,
		CreatedAt:            s.now().UTC(),
	}
	if err := s.repository.CreateBook(ctx, book, authorIDs); err != nil {
		return models.Book{}, err
	}
	return book, nil
}

func (s *Service) GetBook(ctx context.Context, id uuid.UUID) (models.Book, error) {
	return s.repository.GetBook(ctx, id)
}

func (s *Service) ListBooks(ctx context.Context, filter models.BookFilter) ([]models.Book, error) {
	return s.repository.ListBooks(ctx, filter)
}

func (s *Service) ListBorrowedBooks(ctx context.Context, userID uuid.UUID) ([]models.Book, error) {
	if _, err := s.repository.GetUser(ctx, userID); err != nil {
		return nil, err
	}
	return s.repository.ListBorrowedBooks(ctx, userID)
}
