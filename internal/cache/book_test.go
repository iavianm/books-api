package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iavianm/books-api/internal/model"
)

var errNotFound = errors.New("book not found")

type countingRepo struct {
	getByIDCalls atomic.Int64
	getAllCalls  atomic.Int64

	getByIDDelay time.Duration
	getByIDErr   error
}

func (r *countingRepo) Create(_ context.Context, req model.CreateBookRequest) (*model.Book, error) {
	return &model.Book{ID: 1, Title: req.Title}, nil
}

func (r *countingRepo) GetByID(_ context.Context, id int64) (*model.Book, error) {
	r.getByIDCalls.Add(1)
	if r.getByIDDelay > 0 {
		time.Sleep(r.getByIDDelay)
	}
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	return &model.Book{ID: id, Title: "Война и мир"}, nil
}

func (r *countingRepo) GetAll(_ context.Context) ([]model.Book, error) {
	r.getAllCalls.Add(1)
	return []model.Book{{ID: 1, Title: "Война и мир"}, {ID: 2, Title: "Идиот"}}, nil
}

func (r *countingRepo) Update(_ context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error) {
	return &model.Book{ID: id, Title: req.Title}, nil
}

func (r *countingRepo) Delete(_ context.Context, _ int64) error {
	return nil
}

func newTestCache(t *testing.T, repo BookRepo) *BookRepository {
	t.Helper()

	c := NewBookRepository(repo, time.Minute, time.Minute)
	t.Cleanup(c.Close)
	return c
}

func TestGetByIDIsCached(t *testing.T) {
	repo := &countingRepo{}
	c := newTestCache(t, repo)
	ctx := context.Background()

	first, err := c.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("first GetByID: %v", err)
	}
	second, err := c.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("second GetByID: %v", err)
	}

	if got := repo.getByIDCalls.Load(); got != 1 {
		t.Errorf("repository calls = %d, want 1", got)
	}
	if first.Title != second.Title || first.ID != second.ID {
		t.Errorf("cached value differs: %+v vs %+v", first, second)
	}
}

func TestGetByIDReturnsIndependentCopies(t *testing.T) {
	c := newTestCache(t, &countingRepo{})
	ctx := context.Background()

	first, err := c.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	first.Title = "испорчено вызывающим кодом"

	second, err := c.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if second.Title != "Война и мир" {
		t.Errorf("cache was mutated through returned pointer: %q", second.Title)
	}
}

func TestGetAllIsCached(t *testing.T) {
	repo := &countingRepo{}
	c := newTestCache(t, repo)
	ctx := context.Background()

	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("first GetAll: %v", err)
	}
	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("second GetAll: %v", err)
	}

	if got := repo.getAllCalls.Load(); got != 1 {
		t.Errorf("repository calls = %d, want 1", got)
	}
}

func TestGetAllReturnsIndependentSlices(t *testing.T) {
	c := newTestCache(t, &countingRepo{})
	ctx := context.Background()

	first, err := c.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	first[0].Title = "испорчено вызывающим кодом"

	second, err := c.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if second[0].Title != "Война и мир" {
		t.Errorf("cache was mutated through returned slice: %q", second[0].Title)
	}
}

func TestUpdateInvalidatesCache(t *testing.T) {
	repo := &countingRepo{}
	c := newTestCache(t, repo)
	ctx := context.Background()

	if _, err := c.GetByID(ctx, 1); err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if _, err := c.Update(ctx, 1, model.UpdateBookRequest{Title: "Новое название"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if _, err := c.GetByID(ctx, 1); err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("GetAll after update: %v", err)
	}

	if got := repo.getByIDCalls.Load(); got != 2 {
		t.Errorf("getByID calls = %d, want 2 (cache must be invalidated)", got)
	}
	if got := repo.getAllCalls.Load(); got != 2 {
		t.Errorf("getAll calls = %d, want 2 (list cache must be invalidated)", got)
	}
}

func TestDeleteInvalidatesCache(t *testing.T) {
	repo := &countingRepo{}
	c := newTestCache(t, repo)
	ctx := context.Background()

	if _, err := c.GetByID(ctx, 1); err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if err := c.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := c.GetByID(ctx, 1); err != nil {
		t.Fatalf("GetByID after delete: %v", err)
	}

	if got := repo.getByIDCalls.Load(); got != 2 {
		t.Errorf("getByID calls = %d, want 2", got)
	}
}

func TestCreateInvalidatesListCache(t *testing.T) {
	repo := &countingRepo{}
	c := newTestCache(t, repo)
	ctx := context.Background()

	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if _, err := c.Create(ctx, model.CreateBookRequest{Title: "Новая"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := c.GetAll(ctx); err != nil {
		t.Fatalf("GetAll after create: %v", err)
	}

	if got := repo.getAllCalls.Load(); got != 2 {
		t.Errorf("getAll calls = %d, want 2 (list cache must be invalidated)", got)
	}
}

func TestErrorsAreNotCached(t *testing.T) {
	repo := &countingRepo{getByIDErr: errNotFound}
	c := newTestCache(t, repo)
	ctx := context.Background()

	for i := range 2 {
		if _, err := c.GetByID(ctx, 999); !errors.Is(err, errNotFound) {
			t.Fatalf("call %d: err = %v, want errNotFound", i, err)
		}
	}

	if got := repo.getByIDCalls.Load(); got != 2 {
		t.Errorf("repository calls = %d, want 2 (errors must not be cached)", got)
	}
}

func TestConcurrentGetByIDHitsRepositoryOnce(t *testing.T) {
	repo := &countingRepo{getByIDDelay: 50 * time.Millisecond}
	c := newTestCache(t, repo)
	ctx := context.Background()

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			if _, err := c.GetByID(ctx, 1); err != nil {
				t.Errorf("GetByID: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := repo.getByIDCalls.Load(); got != 1 {
		t.Errorf("repository calls = %d, want 1 (single-flight must collapse concurrent loads)", got)
	}
}
