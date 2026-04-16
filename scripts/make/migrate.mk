# Migration targets

.PHONY: migrate-version migrate-create migrate-up migrate-down migrate-force migrate-status

migrate-version:
	go run ./cmd/migrate -action version

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=create_users_table"; exit 1; fi
	go run ./cmd/migrate -action create -name "$(name)"

migrate-up:
	go run ./cmd/migrate -action up

migrate-down:
	go run ./cmd/migrate -action down

migrate-force:
	@if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=1"; exit 1; fi
	go run ./cmd/migrate -action force -version $(version)

migrate-status: migrate-version
