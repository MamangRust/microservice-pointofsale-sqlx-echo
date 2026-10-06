package server

import "time"

// Config defines the common configuration for all gRPC services
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Port           int
	OtelEndpoint   string

	// DBCluster is the env prefix that selects this service's own PostgreSQL
	// instance (e.g. "DB_SALES" -> DB_SALES_HOST/_PORT/_NAME=pos_sales). Each
	// bounded context runs on its own instance behind its own PgBouncer.
	DBCluster string
}

// Default constants for gRPC server
const (
	DefaultMaxConcurrentConn = 1024
	DefaultWindowSize        = 16 * 1024 * 1024
	DefaultKeepaliveTime     = 20 * time.Second
	DefaultKeepaliveTimeout  = 5 * time.Second
	DefaultMinKeepaliveTime  = 5 * time.Second

	MonitoringInterval     = 30 * time.Second
	CleanupInterval        = 120 * time.Second
	CacheRefCountThreshold = 500

	ShutdownTimeout = 30 * time.Second

	RedisDialTimeout  = 5 * time.Second
	RedisReadTimeout  = 3 * time.Second
	RedisWriteTimeout = 3 * time.Second
	RedisPoolSize     = 10
	RedisMinIdleConns = 3
)
