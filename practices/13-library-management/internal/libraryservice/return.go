package libraryservice

import (
	"context"
	"errors"

	"github.com/xtgo/uuid"
)

func (s *Service) Return(ctx context.Context, userID, bookID uuid.UUID) error {
	unlock := s.locks.lock(userID, bookID)
	defer unlock()

	return s.repository.WithTransaction(ctx, func(tx Transaction) error {
		if _, err := tx.GetUser(ctx, userID); err != nil {
			return err
		}
		if _, err := tx.GetBook(ctx, bookID); err != nil {
			return err
		}

		entry, err := tx.FindActiveLedger(ctx, userID, bookID)
		if errors.Is(err, ErrNotFound) {
			return ErrNotBorrowed
		}
		if err != nil {
			return err
		}

		now := s.now().UTC()
		if err = tx.MarkReturned(ctx, entry.ID, now); err != nil {
			return err
		}
		return tx.UpdateBookAvailability(ctx, bookID, 1)
	})
}
