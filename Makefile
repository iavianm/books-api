.PHONY: build run test test-integration cover fmt up down logs db-up db-down db-psql migrate-down migrate-create

-include .env
export

MIGRATIONS_DIR := migrations
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

build:
	go build -o ./.bin/api ./cmd/api

run: build
	./.bin/api

test:
	go test ./... -v

test-integration:
	TEST_DATABASE_DSN="host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)" \
		go test -tags=integration ./internal/repository/ -v

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

fmt:
	gofmt -w .
	goimports -w .

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f app

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-psql:
	docker compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
