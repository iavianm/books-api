.PHONY: build run test fmt

build:
	go build -o ./.bin/api ./cmd/api

run: build
	set -a && . ./.env && set +a && ./.bin/api

test:
	go test ./... -v

fmt:
	gofmt -w .
	goimports -w .
