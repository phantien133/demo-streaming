# Database utility targets

.PHONY: db-schema-snapshot

db-schema-snapshot:
	./scripts/db/dump-schema.sh ./internal/database/schema.sql
