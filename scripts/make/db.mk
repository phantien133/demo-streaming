# Database utility targets

.PHONY: db-schema-snapshot db-reset-hard

db-schema-snapshot:
	./scripts/db/dump-schema.sh ./internal/database/schema.sql

db-reset-hard:
	./scripts/db/hard-reset.sh
