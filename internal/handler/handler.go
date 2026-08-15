package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/iavianm/books-api/internal/model"
)

type BookService interface {
	Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	GetAll(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct {
	service BookService
}

func NewHandler(service BookService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", h.getBooks)
	mux.HandleFunc("GET /books/{id}", h.getBook)
	mux.HandleFunc("POST /books", h.createBook)
	mux.HandleFunc("PUT /books/{id}", h.updateBook)
	mux.HandleFunc("DELETE /books/{id}", h.deleteBook)
	mux.HandleFunc("GET /health", h.health)
	return Recoverer(Logging(mux))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	err := writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	if err != nil {
		ServerError(w, r)
		return
	}
}

func (h *Handler) getBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.GetAll(r.Context())
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = writeJSON(w, http.StatusOK, books)
	if err != nil {
		log.Printf("failed to write response: %v", err)
		ServerError(w, r)
		return
	}
}

func (h *Handler) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := ParseID(r)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		BadRequest(w, r)
		return
	}

	book, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = writeJSON(w, http.StatusOK, book)
	if err != nil {
		log.Printf("failed to write response: %v", err)
		ServerError(w, r)
		return
	}
}

func (h *Handler) createBook(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookRequest
	err := readJSON(r, &req)
	if err != nil {
		log.Printf("failed to read request: %v", err)
		BadRequest(w, r)
		return
	}

	create, err := h.service.Create(r.Context(), req)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, create)
	if err != nil {
		log.Printf("failed to write response: %v", err)
		ServerError(w, r)
		return
	}
}

func (h *Handler) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := ParseID(r)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		BadRequest(w, r)
		return
	}

	var req model.UpdateBookRequest
	err = readJSON(r, &req)
	if err != nil {
		log.Printf("failed to read request: %v", err)
		BadRequest(w, r)
		return
	}

	update, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	err = writeJSON(w, http.StatusOK, update)
	if err != nil {
		log.Printf("failed to write response: %v", err)
		ServerError(w, r)
		return
	}
}

func (h *Handler) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := ParseID(r)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		BadRequest(w, r)
		return
	}
	err = h.service.Delete(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
