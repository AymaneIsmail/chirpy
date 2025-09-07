DSN=postgres://postgres:password@localhost:5432/chirpy?sslmode=disable
MIG_DIR=sql/schema

.PHONY: migration-down migration-up
migration-down:
	goose -dir $(MIG_DIR) postgres "$(DSN)" down; \
	
migration-up:
	goose -dir $(MIG_DIR) postgres "$(DSN)" up
