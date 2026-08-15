package main

import (
	"context"
	"errors"
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

	serv := &http.Server{
		Addr:              ":" + conf.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := serv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := serv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}
	slog.Info("server stopped")
}
