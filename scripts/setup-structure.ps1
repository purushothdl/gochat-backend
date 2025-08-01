# Create all directories
$dirs = @(
    "cmd\server", "cmd\migrate",
    "internal\config", "internal\database", 
    "internal\domain\user", "internal\domain\auth",
    "internal\infrastructure\postgres\migrations",
    "internal\transport\http\middleware",
    "internal\container", "internal\shared\types",
    "internal\shared\dto", "internal\shared\response",
    "internal\shared\logger", "internal\shared\validator",
    "pkg\errors", "pkg\auth"
)

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Force -Path $dir
}

# Create all files
$files = @(
    "cmd\server\main.go", "cmd\migrate\main.go",
    "internal\config\config.go", "internal\config\env.go",
    "internal\database\database.go", "internal\database\migrations.go", "internal\database\health.go",
    "internal\domain\user\entity.go", "internal\domain\user\repository.go", "internal\domain\user\service.go",
    "internal\domain\user\handlers.go", "internal\domain\user\requests.go", "internal\domain\user\responses.go",
    "internal\domain\user\interfaces.go", "internal\domain\user\errors.go",
    "internal\domain\auth\entity.go", "internal\domain\auth\repository.go", "internal\domain\auth\service.go",
    "internal\domain\auth\handlers.go", "internal\domain\auth\requests.go", "internal\domain\auth\responses.go",
    "internal\domain\auth\interfaces.go", "internal\domain\auth\errors.go",
    "internal\infrastructure\postgres\user_repository.go", "internal\infrastructure\postgres\auth_repository.go",
    "internal\transport\http\middleware\auth.go", "internal\transport\http\router.go",
    "internal\transport\http\server.go", "internal\transport\http\response.go",
    "internal\container\container.go", "internal\container\providers.go",
    "internal\shared\types\user.go", "internal\shared\types\common.go",
    "internal\shared\dto\user.go", "internal\shared\dto\common.go",
    "internal\shared\response\response.go", "internal\shared\response\pagination.go",
    "internal\shared\logger\logger.go", "internal\shared\validator\validator.go",
    "pkg\errors\errors.go", "pkg\errors\codes.go",
    "pkg\auth\jwt.go", "pkg\auth\password.go", "pkg\auth\tokens.go",
    "infrastructure\postgres\migrations\001_create_users_table.up.sql",
    "infrastructure\postgres\migrations\001_create_users_table.down.sql",
    "infrastructure\postgres\migrations\002_create_refresh_tokens_table.up.sql",
    "infrastructure\postgres\migrations\002_create_refresh_tokens_table.down.sql",
    ".env", "go.mod", "Makefile", "README.md"
)

foreach ($file in $files) {
    New-Item -ItemType File -Force -Path $file
}

Write-Host "Project structure created successfully!" -ForegroundColor Green
