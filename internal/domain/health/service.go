package health

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/purushothdl/gochat-backend/internal/database"
)

type Service struct {
    dbPool *pgxpool.Pool
    logger *slog.Logger
}

func NewService(dbPool *pgxpool.Pool, logger *slog.Logger) *Service {
    return &Service{
        dbPool: dbPool,
        logger: logger,
    }
}

// CheckHealth performs comprehensive health checks
func (s *Service) CheckHealth(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Timestamp: time.Now(),
        Checks:    make(map[string]Check),
    }

    // Run all health checks
    status.Database = s.checkDatabaseHealth(ctx, &status)
    s.checkPing(ctx, &status)
    s.checkQuery(ctx, &status)
    s.checkMigrationStatus(&status)

    // Determine overall status
    status.Status = s.determineOverallStatus(status.Checks)

    return status
}

// IsHealthy returns a simple boolean health status
func (s *Service) IsHealthy(ctx context.Context) bool {
    health := s.CheckHealth(ctx)
    return health.Status == "healthy"
}

// checkDatabaseHealth checks basic database connectivity and stats
func (s *Service) checkDatabaseHealth(ctx context.Context, status *HealthStatus) DatabaseHealth {
    start := time.Now()
    dbHealth := DatabaseHealth{}

    if s.dbPool == nil {
        status.Checks["database_connection"] = Check{
            Status:   "unhealthy",
            Message:  "Database pool is nil",
            Duration: time.Since(start).Milliseconds(),
            Error:    "pool not initialized",
        }
        return dbHealth
    }

    // Get pool stats
    stats := s.dbPool.Stat()
    dbHealth.ActiveConns = stats.AcquiredConns()
    dbHealth.IdleConns = stats.IdleConns()
    dbHealth.MaxConns = stats.MaxConns()
    dbHealth.Connected = true

    // Test basic connectivity
    if err := s.dbPool.Ping(ctx); err != nil {
        dbHealth.Connected = false
        status.Checks["database_connection"] = Check{
            Status:   "unhealthy",
            Message:  "Database ping failed",
            Duration: time.Since(start).Milliseconds(),
            Error:    err.Error(),
        }
        return dbHealth
    }

    dbHealth.ResponseTime = time.Since(start).Milliseconds()
    status.Checks["database_connection"] = Check{
        Status:   "healthy",
        Message:  "Database connection successful",
        Duration: dbHealth.ResponseTime,
    }

    return dbHealth
}

func (s *Service) checkPing(ctx context.Context, status *HealthStatus) {
    start := time.Now()
    
    if err := s.dbPool.Ping(ctx); err != nil {
        status.Checks["ping"] = Check{
            Status:   "unhealthy",
            Message:  "Ping failed",
            Duration: time.Since(start).Milliseconds(),
            Error:    err.Error(),
        }
        return
    }

    status.Checks["ping"] = Check{
        Status:   "healthy",
        Message:  "Ping successful",
        Duration: time.Since(start).Milliseconds(),
    }
}

func (s *Service) checkQuery(ctx context.Context, status *HealthStatus) {
    start := time.Now()
    
    var result int
    err := s.dbPool.QueryRow(ctx, "SELECT 1").Scan(&result)
    
    if err != nil {
        status.Checks["query"] = Check{
            Status:   "unhealthy",
            Message:  "Query test failed",
            Duration: time.Since(start).Milliseconds(),
            Error:    err.Error(),
        }
        return
    }

    status.Checks["query"] = Check{
        Status:   "healthy",
        Message:  "Query test successful",
        Duration: time.Since(start).Milliseconds(),
    }
}

func (s *Service) checkMigrationStatus(status *HealthStatus) {
    start := time.Now()
    
    migrationRunner := database.NewMigrationRunner(s.dbPool)
    version, dirty, err := migrationRunner.GetVersion()
    
    if err != nil {
        status.Checks["migrations"] = Check{
            Status:   "degraded",
            Message:  "Could not check migration status",
            Duration: time.Since(start).Milliseconds(),
            Error:    err.Error(),
        }
        return
    }

    status.Database.SchemaVersion = version
    status.Database.SchemaDirty = dirty

    if dirty {
        status.Checks["migrations"] = Check{
            Status:   "unhealthy",
            Message:  "Database schema is in dirty state",
            Duration: time.Since(start).Milliseconds(),
            Error:    "migrations are in inconsistent state",
        }
        return
    }

    status.Checks["migrations"] = Check{
        Status:   "healthy",
        Message:  fmt.Sprintf("Schema version %d", version),
        Duration: time.Since(start).Milliseconds(),
    }
}

func (s *Service) determineOverallStatus(checks map[string]Check) string {
    hasUnhealthy := false
    hasDegraded := false

    for _, check := range checks {
        switch check.Status {
        case "unhealthy":
            hasUnhealthy = true
        case "degraded":
            hasDegraded = true
        }
    }

    if hasUnhealthy {
        return "unhealthy"
    }
    if hasDegraded {
        return "degraded"
    }
    return "healthy"
}
