BIN=bin/api

.PHONY: dev build test migrate sqlc gen openapi
dev:
	go run ./backend/cmd/api

build:
	mkdir -p bin && go build -o $(BIN) ./backend/cmd/api

test:
	go test ./...

migrate:
	goose -dir backend/migrations postgres "$$PG_DSN" up

sqlc:
	sqlc generate

openapi:
	oapi-codegen -generate types,gin,spec -o backend/internal/api/gen.go -package api openapi/openapi.yaml

seed:
	go run ./backend/cmd/seeduser
