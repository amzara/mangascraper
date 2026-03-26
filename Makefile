.PHONY: db-create db-drop migrate-up migrate-down migrate-status migrate-create

# Load .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# put in env
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_USER     ?= postgres
DB_PASSWORD ?=
DB_NAME     ?= mangascraper

# Build DSN
DB_DSN := "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"
DB_ADMIN_DSN := "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=disable"

MIGRATIONS_DIR := "./api/migrations"

db-create:
	psql $(DB_ADMIN_DSN) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
	psql $(DB_ADMIN_DSN) -c "CREATE DATABASE $(DB_NAME);"

db-drop:
	psql $(DB_ADMIN_DSN) -c "DROP DATABASE IF EXISTS $(DB_NAME);"

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql

reset: db-drop db-create migrate-up


//https://dzone.com/articles/goose-as-crucial-tool-for-your-service