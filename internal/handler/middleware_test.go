package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovererTurnsPanicInto500(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]string
		m["boom"] = "x"
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/books", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("Connection"); got != "close" {
		t.Errorf("Connection = %q, want close", got)
	}
}

func TestRecovererPassesThroughNormalRequests(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/books", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", rec.Code)
	}
}

func TestStatusWriterCapturesExplicitStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: http.StatusOK}

	sw.WriteHeader(http.StatusNotFound)

	if sw.status != http.StatusNotFound {
		t.Errorf("captured status = %d, want 404", sw.status)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("underlying writer got %d, want 404", rec.Code)
	}
}

func TestStatusWriterDefaultsTo200(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: http.StatusOK}

	if _, err := sw.Write([]byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if sw.status != http.StatusOK {
		t.Errorf("status = %d, want 200 when WriteHeader is not called", sw.status)
	}
}

func TestLoggingPassesRequestThrough(t *testing.T) {
	called := false
	h := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/books", nil))

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}
