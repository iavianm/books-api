.PHONY: build run test fmt db-up db-down db-down migrate-up migrate-down migrate-create

MIGRATIONS_DIR := migrations
DATABASE_URL ?= postgres://postgres_user:postgres_password@localhost:5437/books?sslmode=disable

build:
	go build -o ./.bin/api ./cmd/api

run: build
	set -a && . ./.env && set +a && ./.bin/api

test:
	go test ./... -v

fmt:
	gofmt -w .
	goimports -w .

db-up:
	docker-compose up -d

db-down:
	docker-compose down

db-psql:
	docker-compose exec postgres psql -U postgres -d books

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
