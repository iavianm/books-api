package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iavianm/books-api/internal/model"
)

type mockRepo struct {
	createFn  func(ctx context.Context, req model.CreateBookRequest) (*model.Book, error)
	getByIDFn func(ctx context.Context, id int64) (*model.Book, error)
	getAllFn  func(ctx context.Context) ([]model.Book, error)
	updateFn  func(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error)
	deleteFn  func(ctx context.Context, id int64) error

	createCalls int
	updateCalls int
	deleteCalls int
	gotCreate   model.CreateBookRequest
	gotUpdate   model.UpdateBookRequest
	gotID       int64
}

func (m *mockRepo) Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error) {
	m.createCalls++
	m.gotCreate = req
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return &model.Book{ID: 1, Title: req.Title, Author: req.Author, Genre: req.Genre, Year: req.Year}, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	m.gotID = id
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &model.Book{ID: id}, nil
}

func (m *mockRepo) GetAll(ctx context.Context) ([]model.Book, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return []model.Book{{ID: 1}, {ID: 2}}, nil
}

func (m *mockRepo) Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error) {
	m.updateCalls++
	m.gotUpdate = req
	m.gotID = id
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return &model.Book{ID: id, Title: req.Title}, nil
}

func (m *mockRepo) Delete(ctx context.Context, id int64) error {
	m.deleteCalls++
	m.gotID = id
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func validCreateReq() model.CreateBookRequest {
	return model.CreateBookRequest{Title: "Война и мир", Author: "Толстой", Genre: "novel", Year: 1869}
}

func TestCreateValidation(t *testing.T) {
	nextYear := time.Now().Year() + 1

	tests := []struct {
		name    string
		req     model.CreateBookRequest
		wantErr bool
	}{
		{"valid", validCreateReq(), false},
		{"empty title", model.CreateBookRequest{Title: "", Author: "A", Genre: "G", Year: 2000}, true},
		{"blank title", model.CreateBookRequest{Title: "   ", Author: "A", Genre: "G", Year: 2000}, true},
		{"empty author", model.CreateBookRequest{Title: "T", Author: "", Genre: "G", Year: 2000}, true},
		{"empty genre", model.CreateBookRequest{Title: "T", Author: "A", Genre: "", Year: 2000}, true},
		{"year too old", model.CreateBookRequest{Title: "T", Author: "A", Genre: "G", Year: 1449}, true},
		{"year zero", model.CreateBookRequest{Title: "T", Author: "A", Genre: "G", Year: 0}, true},
		{"year too far", model.CreateBookRequest{Title: "T", Author: "A", Genre: "G", Year: nextYear + 1}, true},
		{"earliest allowed year", model.CreateBookRequest{Title: "T", Author: "A", Genre: "G", Year: 1450}, false},
		{"next year allowed", model.CreateBookRequest{Title: "T", Author: "A", Genre: "G", Year: nextYear}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			_, err := NewBookService(repo).Create(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, ErrValidation) {
					t.Errorf("err = %v, want ErrValidation", err)
				}
				if repo.createCalls != 0 {
					t.Error("repository must not be called on invalid input")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.createCalls != 1 {
				t.Errorf("createCalls = %d, want 1", repo.createCalls)
			}
		})
	}
}

func TestUpdateValidation(t *testing.T) {
	repo := &mockRepo{}
	_, err := NewBookService(repo).Update(context.Background(), 1,
		model.UpdateBookRequest{Title: "", Author: "A", Genre: "G", Year: 2000})

	if !errors.Is(err, ErrValidation) {
		t.Errorf("err = %v, want ErrValidation", err)
	}
	if repo.updateCalls != 0 {
		t.Error("repository must not be called on invalid input")
	}
}

func TestCreateTrimsSpaces(t *testing.T) {
	repo := &mockRepo{}
	_, err := NewBookService(repo).Create(context.Background(), model.CreateBookRequest{
		Title:  "  Война и мир  ",
		Author: "\tТолстой\n",
		Genre:  " novel ",
		Year:   1869,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.gotCreate.Title != "Война и мир" {
		t.Errorf("Title = %q, want %q", repo.gotCreate.Title, "Война и мир")
	}
	if repo.gotCreate.Author != "Толстой" {
		t.Errorf("Author = %q, want %q", repo.gotCreate.Author, "Толстой")
	}
	if repo.gotCreate.Genre != "novel" {
		t.Errorf("Genre = %q, want %q", repo.gotCreate.Genre, "novel")
	}
}

func TestUpdateTrimsSpaces(t *testing.T) {
	repo := &mockRepo{}
	_, err := NewBookService(repo).Update(context.Background(), 7, model.UpdateBookRequest{
		Title:  "  Идиот  ",
		Author: "  Достоевский  ",
		Genre:  "  novel  ",
		Year:   1869,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.gotUpdate.Title != "Идиот" || repo.gotUpdate.Author != "Достоевский" || repo.gotUpdate.Genre != "novel" {
		t.Errorf("not trimmed: %+v", repo.gotUpdate)
	}
	if repo.gotID != 7 {
		t.Errorf("id = %d, want 7", repo.gotID)
	}
}

func TestRepositoryErrorIsPropagated(t *testing.T) {
	repoErr := errors.New("db is down")
	repo := &mockRepo{
		createFn: func(context.Context, model.CreateBookRequest) (*model.Book, error) {
			return nil, repoErr
		},
	}

	book, err := NewBookService(repo).Create(context.Background(), validCreateReq())
	if book != nil {
		t.Errorf("book = %+v, want nil", book)
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("err = %v, want %v", err, repoErr)
	}
}

func TestReadMethodsDelegate(t *testing.T) {
	repo := &mockRepo{}
	s := NewBookService(repo)
	ctx := context.Background()

	book, err := s.GetByID(ctx, 42)
	if err != nil || book.ID != 42 {
		t.Errorf("GetByID = %+v, %v", book, err)
	}
	if repo.gotID != 42 {
		t.Errorf("repo got id = %d, want 42", repo.gotID)
	}

	books, err := s.GetAll(ctx)
	if err != nil || len(books) != 2 {
		t.Errorf("GetAll = %+v, %v", books, err)
	}

	if err := s.Delete(ctx, 5); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if repo.deleteCalls != 1 || repo.gotID != 5 {
		t.Errorf("deleteCalls = %d, id = %d", repo.deleteCalls, repo.gotID)
	}
}
