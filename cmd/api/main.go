package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iavianm/books-api/internal/config"
	"github.com/iavianm/books-api/internal/database"
	"github.com/iavianm/books-api/internal/handler"
	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	conf := config.LoadConfig()

	db, err := database.Connect("pgx", conf.DB.DSN())
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := database.RunMigrations(db, conf.DB.Name); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	slog.Info("migrations applied")

	repo := repository.NewBookRepository(db)
	srv := service.NewBookService(repo)
	h := handler.NewHandler(srv)

	serv := &http.Server{
		Addr:              ":" + conf.HTTPPort,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", conf.HTTPPort)
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := serv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}
