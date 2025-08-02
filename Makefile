# Makefile
.PHONY: migrate-up migrate-down migrate-version migrate-create migrate-force build run clean dev-setup deploy-db

# Variables
# Define the path to  main server binary and migration binary
SERVER_BIN = ./bin/server
MIGRATE_BIN = ./bin/migrate

# Migration Commands (using the dedicated migrate binary)
migrate-up: build-migrate
	@echo "Applying database migrations..."
	@$(MIGRATE_BIN) -up

migrate-down: build-migrate
	@echo "Rolling back last database migration..."
	@$(MIGRATE_BIN) -down

migrate-version: build-migrate
	@echo "Checking database migration version..."
	@$(MIGRATE_BIN) -version

migrate-force: build-migrate
	@echo "Forcing database version..."
	@powershell -Command "$$version = Read-Host 'Enter version to force (e.g., 1): '; $(MIGRATE_BIN) -force $$version"

# Migration File Creation (interactive)
migrate-create:
	@echo "Creating new migration..."
	@powershell -Command " \
		$$name = Read-Host 'Enter migration name (e.g., create_users_table) '; \
		if ([string]::IsNullOrWhiteSpace($$name)) { \
			Write-Error 'Error: Migration name cannot be empty'; \
			exit 1 \
		}; \
		New-Item -ItemType Directory -Force \"internal/infrastructure/postgres/migrations\" | Out-Null; \
		$$timestamp = (Get-Date -Format 'yyyyMMddHHmmss'); \
		$$up_file = \"internal/infrastructure/postgres/migrations/$${timestamp}_$${name}.up.sql\"; \
		$$down_file = \"internal/infrastructure/postgres/migrations/$${timestamp}_$${name}.down.sql\"; \
		Set-Content -Path $$up_file -Value \"-- Migration: $${name}`n-- Created at: $$((Get-Date))`n`n-- Add your UP migration SQL here`n\"; \
		Set-Content -Path $$down_file -Value \"-- Rollback migration: $${name}`n-- Created at: $$((Get-Date))`n`n-- Add your DOWN migration SQL here`n\"; \
		Write-Host 'Created migration files:'; \
		Write-Host \"   $${up_file}\"; \
		Write-Host \"   $${down_file}\"; \
		Write-Host ''; \
		Write-Host 'Next steps:'; \
		Write-Host '1. Edit the .up.sql file with your migration SQL'; \
		Write-Host '2. Edit the .down.sql file with rollback SQL'; \
		Write-Host '3. Run ''make migrate-up'' to apply the migration'"

# Build Commands
build-server:
	@echo "Building server binary..."
	@go build -o $(SERVER_BIN) cmd/server/main.go

build-migrate:
	@echo "Building migration binary..."
	@go build -o $(MIGRATE_BIN) cmd/migrate/main.go

build: build-server build-migrate
	@echo "All binaries built."

# Run Commands
run: build-server
	@echo "Starting server..."
	@$(SERVER_BIN)

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