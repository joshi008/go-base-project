package models

import (
	"time"

	"github.com/xtgo/uuid"
)

type Book struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	AvailabilityQuantity int       `json:"availabilityQuantity"`
	CreatedAt            time.Time `json:"createdAt"`
}

type BookFilter struct {
	Available *bool
	AuthorID  *uuid.UUID
}
