.PHONY: run build test cover migrate-up migrate-down seed fresh swagger tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api && go build -o bin/migrate ./cmd/migrate

test:
	go test ./... -count=1

cover:
	go test ./internal/service/ -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

seed:
	go run ./cmd/migrate seed

fresh:
	go run ./cmd/migrate fresh

swagger:
	swag init -g cmd/api/main.go -o docs --parseInternal

tidy:
	go mod tidy
