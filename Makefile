.PHONY: build run test fmt db-up db-down db-psql migrate-down migrate-create

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

fmt:
	gofmt -w .
	goimports -w .

db-up:
	docker compose up -d

db-down:
	docker compose down

db-psql:
	docker compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
