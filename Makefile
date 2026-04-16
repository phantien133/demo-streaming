SHELL := /bin/sh

ifneq (,$(wildcard .env))
include .env
export
endif

PORT ?= 8080
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= streaming
DB_PASSWORD ?= streaming
DB_NAME ?= streaming
DB_SSLMODE ?= disable
MIGRATIONS_DIR ?= migrations

include scripts/make/dev.mk
include scripts/make/migrate.mk
include scripts/make/db.mk

.PHONY: help

help:
	@echo "Development (scripts/make/dev.mk):"
	@echo "  make run              # go run ./cmd/api"
	@echo "  make transcode-run    # go run ./cmd/transcode"
	@echo "  make dev              # hot reload via air"
	@echo "  make test             # go test ./..."
	@echo "  make docker-up        # docker compose up --build"
	@echo "  make docker-up-transcode # docker compose --profile transcode up --build"
	@echo "  make docker-down      # docker compose down"
	@echo "  make seed             # seed local fake data into Postgres"
	@echo "  make seed-reset       # truncate tables then seed"
	@echo "  make swagger-install  # install swag CLI"
	@echo "  make swagger-gen      # generate Swagger docs into ./docs"
	@echo "  make swagger-db       # docker up + migrate + seed (for Swagger demo)"
	@echo "  make local-ip         # print local LAN IP (macOS)"
	@echo ""
	@echo "Migrations (scripts/make/migrate.mk):"
	@echo "  make migrate-version"
	@echo "  make migrate-create name=create_users_table"
	@echo "  make migrate-up"
	@echo "  make migrate-down"
	@echo "  make migrate-force version=1"
	@echo "  make migrate-status   # alias of migrate-version"
	@echo "  make migrate-up DB_HOST=postgres   # override for docker network"
	@echo ""
	@echo "Database utilities (scripts/make/db.mk):"
	@echo "  make db-schema-snapshot   # write schema.sql to internal/database/"
	@echo "  make db-reset-hard        # drop+create DB then migrate+seed"
