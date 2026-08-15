package handler

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type errorResponse struct {
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	response := errorResponse{Message: message}
	if err := writeJSON(w, status, response); err != nil {
		log.Println("Error writing response:", err)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "book not found")
}

func ServerError(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusBadRequest, "bad request")
}

func ParseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("id must be positive")
	}
	return id, nil
}

func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrBookNotFound):
		NotFound(w, r)
	default:
		slog.Error("internal error", "err", err, "path", r.URL.Path)
		ServerError(w, r)
	}
}
