package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/iavianm/books-api/internal/model"
	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

type mockService struct {
	createFn  func(ctx context.Context, req model.CreateBookRequest) (*model.Book, error)
	getByIDFn func(ctx context.Context, id int64) (*model.Book, error)
	getAllFn  func(ctx context.Context) ([]model.Book, error)
	updateFn  func(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error)
	deleteFn  func(ctx context.Context, id int64) error

	gotID     int64
	gotCreate model.CreateBookRequest
}

func (m *mockService) Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error) {
	m.gotCreate = req
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return &model.Book{ID: 1, Title: req.Title}, nil
}

func (m *mockService) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	m.gotID = id
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.Book{ID: id, Title: "Война и мир"}, nil
}

func (m *mockService) GetAll(ctx context.Context) ([]model.Book, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return []model.Book{{ID: 1, Title: "Война и мир"}, {ID: 2, Title: "Идиот"}}, nil
}

func (m *mockService) Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error) {
	m.gotID = id
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return &model.Book{ID: id, Title: req.Title}, nil
}

func (m *mockService) Delete(ctx context.Context, id int64) error {
	m.gotID = id
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func do(t *testing.T, svc BookService, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	rec := httptest.NewRecorder()
	NewHandler(svc).Routes().ServeHTTP(rec, httptest.NewRequest(method, target, reader))
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodGet, "/health", "")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestGetBooks(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodGet, "/books", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var books []model.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &books); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body)
	}
	if len(books) != 2 || books[0].Title != "Война и мир" {
		t.Errorf("books = %+v", books)
	}
}

func TestGetBooksEmptyListIsArray(t *testing.T) {
	svc := &mockService{getAllFn: func(context.Context) ([]model.Book, error) {
		return []model.Book{}, nil
	}}

	rec := do(t, svc, http.MethodGet, "/books", "")

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func TestGetBooksServiceError(t *testing.T) {
	svc := &mockService{getAllFn: func(context.Context) ([]model.Book, error) {
		return nil, errors.New("db is down")
	}}

	rec := do(t, svc, http.MethodGet, "/books", "")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "db is down") {
		t.Errorf("internal details leaked to client: %s", rec.Body)
	}
}

func TestGetBook(t *testing.T) {
	svc := &mockService{}
	rec := do(t, svc, http.MethodGet, "/books/42", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.gotID != 42 {
		t.Errorf("service got id = %d, want 42", svc.gotID)
	}
}

func TestGetBookNotFound(t *testing.T) {
	svc := &mockService{getByIDFn: func(context.Context, int64) (*model.Book, error) {
		return nil, repository.ErrBookNotFound
	}}

	rec := do(t, svc, http.MethodGet, "/books/999", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestInvalidIDReturnsBadRequest(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{"not a number", "/books/abc"},
		{"negative", "/books/-5"},
		{"zero", "/books/0"},
		{"float", "/books/1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, &mockService{}, http.MethodGet, tt.target, "")
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestCreateBook(t *testing.T) {
	svc := &mockService{}
	body := `{"title":"Война и мир","author":"Толстой","genre":"novel","year":1869}`

	rec := do(t, svc, http.MethodPost, "/books", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %s)", rec.Code, rec.Body)
	}
	if svc.gotCreate.Title != "Война и мир" || svc.gotCreate.Year != 1869 {
		t.Errorf("service got %+v", svc.gotCreate)
	}

	var book model.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &book); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if book.ID != 1 {
		t.Errorf("book = %+v", book)
	}
}

func TestCreateBookInvalidJSON(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodPost, "/books", `{"title":`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestCreateBookValidationError(t *testing.T) {
	svc := &mockService{createFn: func(context.Context, model.CreateBookRequest) (*model.Book, error) {
		return nil, fmt.Errorf("%w: title is required", service.ErrValidation)
	}}

	rec := do(t, svc, http.MethodPost, "/books", `{"title":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "title is required") {
		t.Errorf("body = %s, want validation details", rec.Body)
	}
}

func TestUpdateBook(t *testing.T) {
	svc := &mockService{}
	body := `{"title":"Идиот","author":"Достоевский","genre":"novel","year":1869}`

	rec := do(t, svc, http.MethodPut, "/books/7", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if svc.gotID != 7 {
		t.Errorf("service got id = %d, want 7", svc.gotID)
	}
}

func TestUpdateBookNotFound(t *testing.T) {
	svc := &mockService{updateFn: func(context.Context, int64, model.UpdateBookRequest) (*model.Book, error) {
		return nil, repository.ErrBookNotFound
	}}

	rec := do(t, svc, http.MethodPut, "/books/999", `{"title":"X"}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestDeleteBook(t *testing.T) {
	svc := &mockService{}
	rec := do(t, svc, http.MethodDelete, "/books/3", "")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty for 204", rec.Body)
	}
	if svc.gotID != 3 {
		t.Errorf("service got id = %d, want 3", svc.gotID)
	}
}

func TestDeleteBookNotFound(t *testing.T) {
	svc := &mockService{deleteFn: func(context.Context, int64) error {
		return repository.ErrBookNotFound
	}}

	rec := do(t, svc, http.MethodDelete, "/books/999", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestRoutingRejectsWrongMethod(t *testing.T) {
	tests := []struct {
		method string
		target string
	}{
		{http.MethodDelete, "/books"},
		{http.MethodPut, "/books"},
		{http.MethodPost, "/books/1"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			rec := do(t, &mockService{}, tt.method, tt.target, "")
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want 405", rec.Code)
			}
		})
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodGet, "/unknown", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
