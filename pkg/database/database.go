package database

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// NewClientWithPrefix connects to the PostgreSQL instance selected by the given
// cluster prefix (e.g. DB_IDENTITY, DB_SALES). Each bounded context owns its own
// instance behind its own PgBouncer, so a prefix resolves the full set of
// <prefix>_HOST / _PORT / _USERNAME / _NAME / _PASSWORD keys.
//
// Only DB_USERNAME and DB_PASSWORD keep a generic fallback, because all five
// instances share one credential. The generic DB_HOST / DB_PORT / DB_NAME keys
// are deliberately absent from the environment: without them a missing
// <prefix>_HOST fails the guard below instead of silently connecting to another
// bounded context's database.
func NewClientWithPrefix(logger logger.LoggerInterface, prefix string) (*sqlx.DB, error) {
	if prefix == "" {
		prefix = "DB"
	}

	dbDriver := firstNonEmpty(
		viper.GetString(fmt.Sprintf("%s_DRIVER", prefix)),
		viper.GetString("DB_DRIVER"),
	)
	if dbDriver != "postgres" && dbDriver != "pgx" {
		logger.Error("sqlx only supports PostgreSQL", zap.String("DB_DRIVER", dbDriver))
		return nil, fmt.Errorf("sqlx only supports PostgreSQL, got: %s", dbDriver)
	}

	host := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_HOST", prefix)), viper.GetString("DB_HOST"))
	port := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_PORT", prefix)), viper.GetString("DB_PORT"))
	user := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_USERNAME", prefix)), viper.GetString("DB_USERNAME"))
	dbname := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_NAME", prefix)), viper.GetString("DB_NAME"))
	password := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix)), viper.GetString("DB_PASSWORD"))

	// Each context lives on its own instance, so a missing key must not degrade
	// into pgx's implicit localhost default — that would connect to the wrong
	// database without any error.
	if host == "" || port == "" || dbname == "" {
		logger.Error("Incomplete database configuration",
			zap.String("cluster", prefix),
			zap.String("host", host),
			zap.String("port", port),
			zap.String("dbname", dbname),
		)
		return nil, fmt.Errorf("incomplete database configuration for %s (host=%q port=%q dbname=%q)",
			prefix, host, port, dbname)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password,
	)

	maxOpenConns := viper.GetInt("DB_MAX_OPEN_CONNS")
	if maxOpenConns <= 0 {
		maxOpenConns = 100
	}

	minIdleConns := viper.GetInt("DB_MIN_IDLE_CONNS")
	if minIdleConns <= 0 {
		minIdleConns = 50
	}
	if minIdleConns > maxOpenConns {
		minIdleConns = maxOpenConns
	}

	connMaxLifetime := viper.GetDuration("DB_CONN_MAX_LIFETIME")
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour
	}

	connMaxIdleTime := viper.GetDuration("DB_CONN_MAX_IDLE_TIME")
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 30 * time.Minute
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to database",
			zap.String("cluster", prefix),
			zap.String("dbname", dbname),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbname, err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(minIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	logger.Debug("Database connection pool established successfully",
		zap.String("DB_DRIVER", "pgx"),
		zap.String("cluster", prefix),
		zap.String("dbname", dbname),
		zap.Int("MaxOpenConns", maxOpenConns),
		zap.Int("MaxIdleConns", minIdleConns),
		zap.Duration("ConnMaxLifetime", connMaxLifetime),
		zap.Duration("ConnMaxIdleTime", connMaxIdleTime),
	)

	return db, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
