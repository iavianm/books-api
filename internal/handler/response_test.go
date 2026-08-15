package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	if err := writeJSON(rec, http.StatusCreated, map[string]int{"id": 7}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}

	var got map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["id"] != 7 {
		t.Errorf("body = %v", got)
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    int64
		wantErr bool
	}{
		{"positive", "42", 42, false},
		{"one", "1", 1, false},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, true},
		{"not a number", "abc", 0, true},
		{"empty", "", 0, true},
		{"float", "1.5", 0, true},
		{"overflow", "99999999999999999999", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/books/"+tt.id, nil)
			r.SetPathValue("id", tt.id)

			got, err := ParseID(r)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got id = %d", tt.id, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("id = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHandleServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantInBody string
	}{
		{
			name:       "validation error maps to 400",
			err:        fmt.Errorf("%w: title is required", service.ErrValidation),
			wantStatus: http.StatusBadRequest,
			wantInBody: "title is required",
		},
		{
			name:       "not found maps to 404",
			err:        repository.ErrBookNotFound,
			wantStatus: http.StatusNotFound,
			wantInBody: "book not found",
		},
		{
			name:       "wrapped not found maps to 404",
			err:        fmt.Errorf("get book: %w", repository.ErrBookNotFound),
			wantStatus: http.StatusNotFound,
			wantInBody: "book not found",
		},
		{
			name:       "unknown error maps to 500",
			err:        errors.New("connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantInBody: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/books/1", nil)

			handleServiceError(rec, r, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v (body %s)", err, rec.Body)
			}
			if !strings.Contains(body.Message, tt.wantInBody) {
				t.Errorf("message = %q, want it to contain %q", body.Message, tt.wantInBody)
			}
		})
	}
}

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/books/1", nil)

	handleServiceError(rec, r, errors.New(`pq: relation "books" does not exist`))

	if strings.Contains(rec.Body.String(), "relation") {
		t.Errorf("sql details leaked to client: %s", rec.Body)
	}
}
