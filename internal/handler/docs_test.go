package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestDocsPage(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodGet, "/docs", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "/openapi.yaml") {
		t.Error("docs page must point at the spec URL")
	}
}

func TestOpenAPISpec(t *testing.T) {
	rec := do(t, &mockService{}, http.MethodGet, "/openapi.yaml", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("Content-Type = %q, want application/yaml", ct)
	}

	body := rec.Body.String()
	if !strings.HasPrefix(body, "openapi: 3.1.0") {
		t.Errorf("spec does not start with the openapi version: %.40s", body)
	}
	for _, path := range []string{"/books", "/books/{id}", "/health"} {
		if !strings.Contains(body, path) {
			t.Errorf("spec is missing path %q", path)
		}
	}
}
