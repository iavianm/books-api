.PHONY: build run test fmt db-up db-down db-down

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
