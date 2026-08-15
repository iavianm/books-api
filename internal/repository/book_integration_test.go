//go:build integration

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/iavianm/books-api/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5437 user=postgres_user password=postgres dbname=books sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v (is the database running? make db-up)", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRepositoryCRUD(t *testing.T) {
	repo := NewBookRepository(testDB(t))
	ctx := context.Background()

	created, err := repo.Create(ctx, model.CreateBookRequest{
		Title:  "Интеграционный тест",
		Author: "Автор",
		Genre:  "test",
		Year:   2024,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, created.ID) })

	if created.ID == 0 {
		t.Error("Create must return generated id")
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Error("Create must return timestamps")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Интеграционный тест" || got.Year != 2024 {
		t.Errorf("GetByID = %+v", got)
	}

	books, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(books) == 0 {
		t.Error("GetAll returned empty list after Create")
	}
	for i := 1; i < len(books); i++ {
		if books[i-1].ID > books[i].ID {
			t.Errorf("GetAll is not ordered by id: %d before %d", books[i-1].ID, books[i].ID)
			break
		}
	}

	updated, err := repo.Update(ctx, created.ID, model.UpdateBookRequest{
		Title:  "Обновлённый",
		Author: "Автор",
		Genre:  "test",
		Year:   2025,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Обновлённый" || updated.Year != 2025 {
		t.Errorf("Update = %+v", updated)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Errorf("updated_at was not refreshed: %v -> %v", created.UpdatedAt, updated.UpdatedAt)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := repo.GetByID(ctx, created.ID); !errors.Is(err, ErrBookNotFound) {
		t.Errorf("GetByID after Delete: err = %v, want ErrBookNotFound", err)
	}
}

func TestRepositoryNotFound(t *testing.T) {
	repo := NewBookRepository(testDB(t))
	ctx := context.Background()

	const missingID = 999999999

	if _, err := repo.GetByID(ctx, missingID); !errors.Is(err, ErrBookNotFound) {
		t.Errorf("GetByID: err = %v, want ErrBookNotFound", err)
	}

	_, err := repo.Update(ctx, missingID, model.UpdateBookRequest{
		Title: "X", Author: "A", Genre: "G", Year: 2000,
	})
	if !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Update: err = %v, want ErrBookNotFound", err)
	}

	if err := repo.Delete(ctx, missingID); !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Delete: err = %v, want ErrBookNotFound", err)
	}
}
