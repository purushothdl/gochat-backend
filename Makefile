# Makefile
.PHONY: migrate-up migrate-down migrate-version migrate-create

migrate-up:
	go run cmd/migrate/main.go -up

migrate-down:
	go run cmd/migrate/main.go -down

migrate-version:
	go run cmd/migrate/main.go -version

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

# Development
dev-setup: migrate-up
	go run cmd/server/main.go

# Production deployment
deploy-db: migrate-version migrate-up
	@echo "Database migration completed"