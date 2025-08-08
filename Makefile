# Makefile
.PHONY: migrate-up migrate-down migrate-version migrate-create migrate-force build run clean dev-setup deploy-db

# Variables
# Define the path to  main server binary and migration binary
SERVER_BIN = ./bin/server
MIGRATE_BIN = ./bin/migrate
WEBSOCKET_BIN = ./bin/websocket

# Migration Commands (using the dedicated migrate binary)
migrate-up: build-migrate
	@echo "Applying database migrations..."
	@$(MIGRATE_BIN) -up

migrate-down: build-migrate
	@echo "Rolling back last database migration..."
	@$(MIGRATE_BIN) -down

migrate-version: build-migrate
	@echo "Checking database migration status..."
	@$(MIGRATE_BIN) -version

migrate-force: build-migrate
	@echo "Forcing database version..."
	@read -p "Enter version to force: " version; \
	$(MIGRATE_BIN) -force=$$version

migrate-create: build-migrate
	@echo "Creating new migration..."
	@$(MIGRATE_BIN) -create

migrate-create-named: build-migrate
	@echo "Creating new migration with name..."
	@read -p "Enter migration name: " name; \
	$(MIGRATE_BIN) -create -name="$$name"

migrate-list: build-migrate
	@echo "Listing all migrations..."
	@$(MIGRATE_BIN) -list


# Build Commands
build-server:
	@echo "Building server binary..."
	@go build -o $(SERVER_BIN) cmd/server/main.go

build-websocket:
	@echo "Building websocket binary..."
	@go build -o $(WEBSOCKET_BIN) cmd/websocket/main.go

build-migrate:
	@echo "Building migration binary..."
	@go build -o $(MIGRATE_BIN) cmd/migrate/main.go

build: build-server build-websocket build-migrate
	@echo "All binaries built."

# Run Commands
run: build-server
	@echo "Starting server..."
	@$(SERVER_BIN)

run-websocket: build-websocket
	@echo "Starting websocket server..."
	@$(WEBSOCKET_BIN)

# Development Setup (runs migrations, then starts server)
dev-setup: migrate-up run
	@echo "Development setup complete."

# Clean
clean:
	@echo "Cleaning binaries..."
	@rm -rf ./bin
	@echo "Clean complete."

# Production deployment placeholder
deploy-db: migrate-version migrate-up
	@echo "Database migration completed for deployment."