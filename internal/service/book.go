package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iavianm/books-api/internal/model"
)

var ErrValidation = errors.New("validation failed")

type BookRepo interface {
	Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	GetAll(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

type BookService struct {
	repo BookRepo
}

func NewBookService(repo BookRepo) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Author = strings.TrimSpace(req.Author)
	req.Genre = strings.TrimSpace(req.Genre)

	err := validate(req.Title, req.Author, req.Genre, req.Year)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, req)
}

func (s *BookService) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BookService) GetAll(ctx context.Context) ([]model.Book, error) {
	return s.repo.GetAll(ctx)
}

func (s *BookService) Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Author = strings.TrimSpace(req.Author)
	req.Genre = strings.TrimSpace(req.Genre)

	err := validate(req.Title, req.Author, req.Genre, req.Year)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, req)
}

func (s *BookService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func validate(title, author, genre string, year int) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}

	if strings.TrimSpace(author) == "" {
		return fmt.Errorf("%w: author is required", ErrValidation)
	}

	if strings.TrimSpace(genre) == "" {
		return fmt.Errorf("%w: genre is required", ErrValidation)
	}

	if year < 1450 || year > time.Now().Year()+1 {
		return fmt.Errorf("%w: year must be between 1450 and %d", ErrValidation, time.Now().Year()+1)
	}

	return nil
}
