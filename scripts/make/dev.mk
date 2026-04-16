# Development targets

.PHONY: run transcode-run test docker-up docker-down docker-up-transcode seed seed-reset swagger-install swagger-gen swagger-db local-ip
.PHONY: dev

run:
	go run ./cmd/api

transcode-run:
	go run ./cmd/transcode

dev:
	air

test:
	go test ./...

docker-up:
	docker compose up --build

docker-up-transcode:
	docker compose --profile transcode up --build

docker-down:
	docker compose down

seed: migrate-up
	go run ./cmd/seed

seed-reset: migrate-up
	go run ./cmd/seed -reset

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-gen:
	swag init -g ./cmd/api/main.go -o ./docs

swagger-db:
	docker compose up -d --build
	make migrate-up
	make seed

local-ip:
	./scripts/net/local-ip.sh
