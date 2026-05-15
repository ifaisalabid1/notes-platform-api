.PHONY: run dev db-up db-down db-reset migrate-up migrate-down tidy

include .env
export

run:
	go run ./cmd/api

dev:
	go tool air

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

tidy:
	go mod tidy