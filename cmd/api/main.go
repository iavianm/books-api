package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/iavianm/books-api/internal/config"
	"github.com/iavianm/books-api/internal/database"
	"github.com/iavianm/books-api/internal/handler"
	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	conf := config.LoadConfig()

	db, err := database.Connect("pgx", conf.DB.DSN())
	if err != nil {
		slog.Error("connect to database", "err", err)
		os.Exit(1)
	}

	defer func() { _ = db.Close() }()

	if err := database.RunMigrations(db, conf.DB.Name); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	repo := repository.NewBookRepository(db)
	srv := service.NewBookService(repo)
	h := handler.NewHandler(srv)

	mux := h.Routes()

	slog.Info("starting server", "port", conf.HTTPPort)

	if err := http.ListenAndServe(":"+conf.HTTPPort, mux); err != nil {
		slog.Error("http server", "err", err)
		os.Exit(1)
	}
}
