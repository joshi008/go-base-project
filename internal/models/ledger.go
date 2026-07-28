package models

import (
	"time"

	"github.com/xtgo/uuid"
)

type ActionType string

const Borrow ActionType = "BORROW"

type LibraryLedgerEntry struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"userId"`
	BookID     uuid.UUID  `json:"bookId"`
	ActionType ActionType `json:"actionType"`
	ReturnedAt *time.Time `json:"returnedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
