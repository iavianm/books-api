package cache

import (
	"context"
	"strconv"
	"time"

	wrpool "github.com/iavianm/memory_cache_wrpool"
	wrcache "github.com/iavianm/memory_cache_wrpool/cache"
	"github.com/iavianm/memory_cache_wrpool/workerpool"

	"github.com/iavianm/books-api/internal/model"
)

const (
	listKey = "books:all"

	poolWorkers   = 4
	poolQueueSize = 64
)

type BookRepo interface {
	Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	GetAll(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

type BookRepository struct {
	repo BookRepo

	bookCache *wrcache.Cache[model.Book]
	bookPool  *workerpool.Pool[model.Book]
	bookExec  *wrpool.Executor[model.Book]

	listCache *wrcache.Cache[[]model.Book]
	listPool  *workerpool.Pool[[]model.Book]
	listExec  *wrpool.Executor[[]model.Book]
}

func NewBookRepository(repo BookRepo, ttl, cleanup time.Duration) *BookRepository {
	bookCache := wrcache.New[model.Book](ttl, cleanup)
	bookPool := workerpool.New[model.Book](poolWorkers, poolQueueSize)

	listCache := wrcache.New[[]model.Book](ttl, cleanup)
	listPool := workerpool.New[[]model.Book](poolWorkers, poolQueueSize)

	return &BookRepository{
		repo:      repo,
		bookCache: bookCache,
		bookPool:  bookPool,
		bookExec:  wrpool.New(bookCache, bookPool),
		listCache: listCache,
		listPool:  listPool,
		listExec:  wrpool.New(listCache, listPool),
	}
}

func bookKey(id int64) string {
	return "book:" + strconv.FormatInt(id, 10)
}

func (c *BookRepository) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	book, err := c.bookExec.Do(ctx, bookKey(id), func(ctx context.Context) (model.Book, error) {
		b, err := c.repo.GetByID(ctx, id)
		if err != nil {
			return model.Book{}, err
		}
		return *b, nil
	})
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (c *BookRepository) GetAll(ctx context.Context) ([]model.Book, error) {
	books, err := c.listExec.Do(ctx, listKey, func(ctx context.Context) ([]model.Book, error) {
		return c.repo.GetAll(ctx)
	})
	if err != nil {
		return nil, err
	}

	out := make([]model.Book, len(books))
	copy(out, books)

	return out, nil
}

func (c *BookRepository) Create(ctx context.Context, req model.CreateBookRequest) (*model.Book, error) {
	book, err := c.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	c.listCache.Delete(listKey)

	return book, nil
}

func (c *BookRepository) Update(ctx context.Context, id int64, req model.UpdateBookRequest) (*model.Book, error) {
	book, err := c.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	c.bookCache.Delete(bookKey(id))
	c.listCache.Delete(listKey)

	return book, nil
}

func (c *BookRepository) Delete(ctx context.Context, id int64) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return err
	}

	c.bookCache.Delete(bookKey(id))
	c.listCache.Delete(listKey)

	return nil
}

func (c *BookRepository) Close() {
	c.bookPool.Stop()
	c.listPool.Stop()
	c.bookCache.Close()
	c.listCache.Close()
}
