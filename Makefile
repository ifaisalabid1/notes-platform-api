.PHONY: run dev db-up db-down db-reset migrate-up migrate-down tidy docker-build docker-run

include .env
export

APP_NAME=notes-platform-api
DOCKER_IMAGE=$(APP_NAME):local

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

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm \
		--name $(APP_NAME) \
		-p 8080:8080 \
		--env-file .env \
		-e PORT=8080 \
		$(DOCKER_IMAGE)