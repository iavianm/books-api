package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/iavianm/books-api/internal/model"
)

var ErrBookNotFound = errors.New("book not found")

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (b BookRepository) Create(ctx context.Context, book model.CreateBookRequest) (*model.Book, error) {
	query := `INSERT INTO books (title, author, genre, year) 
			  VALUES ($1, $2, $3, $4) 
			  RETURNING id, title, author, genre, year, created_at, updated_at`

	row := b.db.QueryRowContext(ctx, query, book.Title, book.Author, book.Genre, book.Year)

	var newBook model.Book

	err := row.Scan(
		&newBook.ID,
		&newBook.Title,
		&newBook.Author,
		&newBook.Genre,
		&newBook.Year,
		&newBook.CreatedAt,
		&newBook.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &newBook, nil
}

func (b BookRepository) Update(ctx context.Context, id int64, book model.UpdateBookRequest) (*model.Book, error) {
	query := `UPDATE books SET title=$2, author=$3, genre=$4, year=$5, updated_at=now() 
              WHERE id=$1 RETURNING id, title, author, genre, year, created_at, updated_at`

	row := b.db.QueryRowContext(ctx, query, id, book.Title, book.Author, book.Genre, book.Year)

	var updatedBook model.Book
	err := row.Scan(
		&updatedBook.ID,
		&updatedBook.Title,
		&updatedBook.Author,
		&updatedBook.Genre,
		&updatedBook.Year,
		&updatedBook.CreatedAt,
		&updatedBook.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return &updatedBook, nil
}

func (b BookRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM books WHERE id = $1`
	result, err := b.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBookNotFound
	}
	return nil
}

func (b BookRepository) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	query := `SELECT id, title, author, genre, year, created_at, updated_at FROM books WHERE id = $1`

	row := b.db.QueryRowContext(ctx, query, id)

	var foundBook model.Book
	err := row.Scan(
		&foundBook.ID,
		&foundBook.Title,
		&foundBook.Author,
		&foundBook.Genre,
		&foundBook.Year,
		&foundBook.CreatedAt,
		&foundBook.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return &foundBook, nil
}

func (b BookRepository) GetAll(ctx context.Context) ([]model.Book, error) {
	query := `SELECT id, title, author, genre, year, created_at, updated_at FROM books ORDER BY id`
	rows, err := b.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	foundBooks := make([]model.Book, 0)

	for rows.Next() {
		var foundBook model.Book
		err := rows.Scan(
			&foundBook.ID,
			&foundBook.Title,
			&foundBook.Author,
			&foundBook.Genre,
			&foundBook.Year,
			&foundBook.CreatedAt,
			&foundBook.UpdatedAt)

		if err != nil {
			return nil, err
		}

		foundBooks = append(foundBooks, foundBook)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foundBooks, nil
}
