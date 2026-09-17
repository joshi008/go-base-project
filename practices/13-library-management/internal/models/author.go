package models

import "github.com/xtgo/uuid"

type Author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AuthorBookMapping struct {
	ID       uuid.UUID `json:"id"`
	BookID   uuid.UUID `json:"bookId"`
	AuthorID uuid.UUID `json:"authorId"`
}
