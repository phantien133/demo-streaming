# Development targets

.PHONY: run test docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
