package health

import "time"

// HealthStatus represents the overall health status
type HealthStatus struct {
    Status      string            `json:"status"`      // "healthy", "unhealthy", "degraded"
    Timestamp   time.Time         `json:"timestamp"`
    Database    DatabaseHealth    `json:"database"`
    Checks      map[string]Check  `json:"checks"`
}

// DatabaseHealth contains database-specific health metrics
type DatabaseHealth struct {
    Connected       bool          `json:"connected"`
    ResponseTime    int64         `json:"response_time_ms"`
    ActiveConns     int32         `json:"active_connections"`
    IdleConns       int32         `json:"idle_connections"`
    MaxConns        int32         `json:"max_connections"`
    SchemaVersion   uint          `json:"schema_version"`
    SchemaDirty     bool          `json:"schema_dirty"`
}

// Check represents an individual health check result
type Check struct {
    Status    string        `json:"status"`
    Message   string        `json:"message,omitempty"`
    Duration  int64         `json:"duration_ms"`
    Error     string        `json:"error,omitempty"`
}
