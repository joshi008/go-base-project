package libraryservice

import (
	"context"
	"errors"

	"go-base-project/internal/models"

	"github.com/xtgo/uuid"
)

func (s *Service) Borrow(ctx context.Context, userID, bookID uuid.UUID) error {
	unlock := s.locks.lock(userID, bookID)
	defer unlock()

	return s.repository.WithTransaction(ctx, func(tx Transaction) error {
		if _, err := tx.GetUser(ctx, userID); err != nil {
			return err
		}

		count, err := tx.ActiveBorrowCount(ctx, userID)
		if err != nil {
			return err
		}
		if count >= maxActiveBorrows {
			return ErrBorrowLimit
		}

		if _, err = tx.FindActiveLedger(ctx, userID, bookID); err == nil {
			return ErrAlreadyBorrowed
		} else if !errors.Is(err, ErrNotFound) {
			return err
		}

		book, err := tx.GetBook(ctx, bookID)
		if err != nil {
			return err
		}
		if book.AvailabilityQuantity <= 0 {
			return ErrBookUnavailable
		}

		if err = tx.UpdateBookAvailability(ctx, bookID, -1); err != nil {
			return err
		}
		now := s.now().UTC()
		return tx.InsertLedger(ctx, models.LibraryLedgerEntry{
			ID:         uuid.NewRandom(),
			UserID:     userID,
			BookID:     bookID,
			ActionType: models.Borrow,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	})
}
