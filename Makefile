.PHONY: run dev db-up db-down tidy

run:
	go run ./cmd/api

dev:
	go tool air

db-up:
	docker compose up -d

db-down:
	docker compose down

tidy:
	go mod tidy