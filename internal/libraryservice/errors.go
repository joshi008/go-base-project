package libraryservice

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrBorrowLimit     = errors.New("user already has the maximum number of borrowed books")
	ErrBookUnavailable = errors.New("book is unavailable")
	ErrAlreadyBorrowed = errors.New("user already borrowed this book")
	ErrNotBorrowed     = errors.New("user has not borrowed this book")
)
