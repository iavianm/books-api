package main

import (
	"log"
	"net/http"

	"github.com/iavianm/books-api/internal/config"
	"github.com/iavianm/books-api/internal/database"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting server on port %s", conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+conf.HTTPPort, mux))
}
