BIN=bin/api
GOLANGCI_CONFIG=backend/.golangci.yaml

.PHONY: dev build test lint lint-fix fmt migrate sqlc gen openapi setup
dev:
	go run ./backend/cmd/api

build:
	mkdir -p bin && go build -o $(BIN) ./backend/cmd/api

test:
	go test ./...

test-v:
	go test -v -count=1 ./...

test-nc:
	go test -count=1 ./...

lint:
	golangci-lint run -c $(GOLANGCI_CONFIG)

lint-fix:
	golangci-lint run -c $(GOLANGCI_CONFIG) --fix

fmt:
	golangci-lint fmt -c $(GOLANGCI_CONFIG)

migrate:
	MIGRATION_DIR=backend/migrations go run ./backend/cmd/migrate

sqlc:
	sqlc generate

openapi:
	oapi-codegen -generate types,gin,spec -o backend/internal/api/gen.go -package api openapi/openapi.yaml
	pnpm --dir frontend gen:types

seed:
	go run ./backend/cmd/seeduser

pre-commit-setup:
	git config core.hooksPath .githooks
