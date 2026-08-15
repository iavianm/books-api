package main

import (
	"log"
	"net/http"

	"github.com/iavianm/books-api/internal/config"
	"github.com/iavianm/books-api/internal/database"
	"github.com/iavianm/books-api/internal/handler"
	"github.com/iavianm/books-api/internal/repository"
	"github.com/iavianm/books-api/internal/service"
)

func main() {
	conf := config.LoadConfig()

	db, err := database.Connect("pgx", conf.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := database.RunMigrations(db, conf.DB.Name); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied")

	repo := repository.NewBookRepository(db)
	srv := service.NewBookService(repo)
	h := handler.NewHandler(srv)

	mux := h.Routes()

	log.Printf("Starting server on port %s", conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+conf.HTTPPort, mux))
}
