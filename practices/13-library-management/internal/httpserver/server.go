package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-base-project/internal/libraryservice"
	"go-base-project/internal/models"

	"github.com/xtgo/uuid"
)

type LibraryService interface {
	CreateUser(context.Context, string) (models.User, error)
	GetUser(context.Context, uuid.UUID) (models.User, error)
	CreateBook(context.Context, string, []uuid.UUID, int) (models.Book, error)
	GetBook(context.Context, uuid.UUID) (models.Book, error)
	ListBooks(context.Context, models.BookFilter) ([]models.Book, error)
	Borrow(context.Context, uuid.UUID, uuid.UUID) error
	Return(context.Context, uuid.UUID, uuid.UUID) error
	ListBorrowedBooks(context.Context, uuid.UUID) ([]models.Book, error)
}

type Server struct {
	service LibraryService
	mux     *http.ServeMux
}

func New(service LibraryService) *Server {
	server := &Server{service: service, mux: http.NewServeMux()}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /books", s.createBook)
	s.mux.HandleFunc("GET /books", s.listBooks)
	s.mux.HandleFunc("GET /books/{id}", s.getBook)
	s.mux.HandleFunc("POST /users", s.createUser)
	s.mux.HandleFunc("GET /users/{id}", s.getUser)
	s.mux.HandleFunc("POST /users/{id}/borrow", s.borrow)
	s.mux.HandleFunc("POST /users/{id}/return", s.returnBook)
	s.mux.HandleFunc("GET /users/{id}/books", s.listBorrowedBooks)
}

func (s *Server) createUser(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(writer, request, &input) {
		return
	}
	user, err := s.service.CreateUser(request.Context(), input.Name)
	s.respond(writer, http.StatusCreated, user, err)
}

func (s *Server) getUser(writer http.ResponseWriter, request *http.Request) {
	id, ok := pathUUID(writer, request, "id")
	if !ok {
		return
	}
	user, err := s.service.GetUser(request.Context(), id)
	s.respond(writer, http.StatusOK, user, err)
}

func (s *Server) createBook(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Name                 string   `json:"name"`
		AuthorIDs            []string `json:"authorIds"`
		AvailabilityQuantity int      `json:"availabilityQuantity"`
	}
	if !decodeJSON(writer, request, &input) {
		return
	}
	authorIDs := make([]uuid.UUID, len(input.AuthorIDs))
	for index, rawID := range input.AuthorIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			writeError(writer, http.StatusBadRequest, "authorIds must contain valid UUIDs")
			return
		}
		authorIDs[index] = id
	}
	book, err := s.service.CreateBook(
		request.Context(), input.Name, authorIDs, input.AvailabilityQuantity)
	s.respond(writer, http.StatusCreated, book, err)
}

func (s *Server) getBook(writer http.ResponseWriter, request *http.Request) {
	id, ok := pathUUID(writer, request, "id")
	if !ok {
		return
	}
	book, err := s.service.GetBook(request.Context(), id)
	s.respond(writer, http.StatusOK, book, err)
}

func (s *Server) listBooks(writer http.ResponseWriter, request *http.Request) {
	var filter models.BookFilter
	if rawAvailable := request.URL.Query().Get("available"); rawAvailable != "" {
		available, err := strconv.ParseBool(rawAvailable)
		if err != nil {
			writeError(writer, http.StatusBadRequest, "available must be true or false")
			return
		}
		filter.Available = &available
	}
	if rawAuthor := request.URL.Query().Get("author"); rawAuthor != "" {
		authorID, err := uuid.Parse(rawAuthor)
		if err != nil {
			writeError(writer, http.StatusBadRequest, "author must be a valid UUID")
			return
		}
		filter.AuthorID = &authorID
	}
	books, err := s.service.ListBooks(request.Context(), filter)
	s.respond(writer, http.StatusOK, books, err)
}

func (s *Server) borrow(writer http.ResponseWriter, request *http.Request) {
	s.changeBorrowState(writer, request, s.service.Borrow)
}

func (s *Server) returnBook(writer http.ResponseWriter, request *http.Request) {
	s.changeBorrowState(writer, request, s.service.Return)
}

func (s *Server) changeBorrowState(
	writer http.ResponseWriter,
	request *http.Request,
	action func(context.Context, uuid.UUID, uuid.UUID) error,
) {
	userID, ok := pathUUID(writer, request, "id")
	if !ok {
		return
	}
	var input struct {
		BookID string `json:"bookId"`
	}
	if !decodeJSON(writer, request, &input) {
		return
	}
	bookID, err := uuid.Parse(input.BookID)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "bookId must be a valid UUID")
		return
	}
	if err = action(request.Context(), userID, bookID); err != nil {
		s.respond(writer, 0, nil, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (s *Server) listBorrowedBooks(writer http.ResponseWriter, request *http.Request) {
	userID, ok := pathUUID(writer, request, "id")
	if !ok {
		return
	}
	books, err := s.service.ListBorrowedBooks(request.Context(), userID)
	s.respond(writer, http.StatusOK, books, err)
}

func (s *Server) respond(writer http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		switch {
		case errors.Is(err, libraryservice.ErrInvalidInput):
			status = http.StatusBadRequest
		case errors.Is(err, libraryservice.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, libraryservice.ErrBorrowLimit),
			errors.Is(err, libraryservice.ErrBookUnavailable),
			errors.Is(err, libraryservice.ErrAlreadyBorrowed),
			errors.Is(err, libraryservice.ErrNotBorrowed):
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
			err = errors.New("internal server error")
		}
		writeError(writer, status, err.Error())
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if encodeErr := json.NewEncoder(writer).Encode(responseValue(value)); encodeErr != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
	}
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type bookResponse struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	AvailabilityQuantity int       `json:"availabilityQuantity"`
	CreatedAt            time.Time `json:"createdAt"`
}

func responseValue(value any) any {
	switch typed := value.(type) {
	case models.User:
		return userResponse{
			ID: typed.ID.String(), Name: typed.Name, CreatedAt: typed.CreatedAt,
		}
	case models.Book:
		return toBookResponse(typed)
	case []models.Book:
		response := make([]bookResponse, len(typed))
		for index, book := range typed {
			response[index] = toBookResponse(book)
		}
		return response
	default:
		return value
	}
}

func toBookResponse(book models.Book) bookResponse {
	return bookResponse{
		ID:                   book.ID.String(),
		Name:                 book.Name,
		AvailabilityQuantity: book.AvailabilityQuantity,
		CreatedAt:            book.CreatedAt,
	}
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, target any) bool {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func pathUUID(
	writer http.ResponseWriter,
	request *http.Request,
	name string,
) (uuid.UUID, bool) {
	id, err := uuid.Parse(request.PathValue(name))
	if err != nil {
		writeError(writer, http.StatusBadRequest, fmt.Sprintf("%s must be a valid UUID", name))
		return uuid.UUID{}, false
	}
	return id, true
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]string{"error": message})
}
