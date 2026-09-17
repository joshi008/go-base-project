package httpserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-base-project/internal/httpserver"
	"go-base-project/internal/libraryservice"
	"go-base-project/internal/repository/memory"
)

func TestBorrowWorkflow(t *testing.T) {
	handler := httpserver.New(libraryservice.New(memory.New())).Handler()

	user := performJSON(t, handler, http.MethodPost, "/users", `{"name":"Ada"}`,
		http.StatusCreated)
	book := performJSON(t, handler, http.MethodPost, "/books",
		`{"name":"Concurrency in Go","availabilityQuantity":1}`,
		http.StatusCreated)
	performJSON(t, handler, http.MethodPost, "/users/"+user["id"].(string)+"/borrow",
		`{"bookId":"`+book["id"].(string)+`"}`, http.StatusNoContent)

	request := httptest.NewRequest(
		http.MethodGet, "/users/"+user["id"].(string)+"/books", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET borrowed books status = %d, body = %s", response.Code, response.Body)
	}
	var books []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&books); err != nil {
		t.Fatalf("decode borrowed books: %v", err)
	}
	if len(books) != 1 || books[0]["id"] != book["id"] {
		t.Fatalf("borrowed books = %#v, want book %v", books, book["id"])
	}
}

func TestInvalidPathUUIDReturnsBadRequest(t *testing.T) {
	handler := httpserver.New(libraryservice.New(memory.New())).Handler()
	request := httptest.NewRequest(http.MethodGet, "/books/not-a-uuid", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func performJSON(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body string,
	wantStatus int,
) map[string]any {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s status = %d, body = %s", method, path, response.Code, response.Body)
	}
	if wantStatus == http.StatusNoContent {
		return nil
	}
	var value map[string]any
	if err := json.NewDecoder(response.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return value
}
