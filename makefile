
include .env
export $(shell sed 's/=.*//' .env)

# Variables
#Default replacement word; can be overridden via command line
	REPLACE_WITH ?= void
# Define the directories to search; defaults to the current directory
	DIRS ?= .
# Define the file patterns to include in the search
	FILE_PATTERN ?= *.go

migration-create:
	@echo "**************************** migration create ***************************************"
	goose -dir migrations create $(NAME) -ext sql
	@echo "******************************************************************************"

migrate-up:
	@echo "**************************** migration up ***************************************"
	@command="goose -dir migrations postgres \"user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}\" up"; \
	echo $$command; \
	result=$$(eval $$command); \
	echo "$$result"
	@echo "******************************************************************************"
migrate-down:
	@echo "**************************** migration down ***************************************"
	@command="goose -dir migrations postgres \"user=${PG_USER} password=${PG_PASSWORD} dbname=${PG_DB} sslmode=${PG_SSLMODE}\" down"; \
	echo $$command; \
	result=$$(eval $$command); \
	echo "$$result"
	@echo "******************************************************************************"

bootstrap:
	@for dir in $(DIRS); do \
		if [ "$(shell uname)" = "Darwin" ]; then \
			find "$$dir" -type f -name "$(FILE_PATTERN)" -exec sed -i "" 's/gobp/$(REPLACE_WITH)/g' {} +; \
		else \
			find "$$dir" -type f -name "$(FILE_PATTERN)" -exec sed -i 's/gobp/$(REPLACE_WITH)/g' {} +; \
		fi \
	done
	rm -f go.mod
	go mod init $(REPLACE_WITH)
	go mod tidy
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
# to run: make bootstrap REPLACE_WITH=example DIRS="src include" FILE_PATTERN="*.go"

sqlc:
	@echo "**************************** sqlc generate ***************************************"#
	cd pkg/db/sqlc && sqlc generate && cd ../../../ 
	@echo "******************************************************************************"

dev:
	go build brainwars

DEV_COMPOSE_FILE=./docker/docker-compose-dev.yml
DEBUG_COMPOSE_FILE=./docker/docker-compose-debug.yml

.PHONY: compose-up
compose-up:
	docker compose -f $(DEV_COMPOSE_FILE) up -d

.PHONY: compose-up-build
compose-up-build:
	docker compose -f $(DEV_COMPOSE_FILE) up --build

.PHONY: compose-up-debug
compose-up-debug:
	docker compose -f $(DEV_COMPOSE_FILE) -f $(DEBUG_COMPOSE_FILE) up -d
	
.PHONY: compose-up-debug-build
compose-up-debug-build: # using multiple docker compose file and building together will stack up the compose.
	docker compose -f $(DEV_COMPOSE_FILE) -f $(DEBUG_COMPOSE_FILE) up --build
