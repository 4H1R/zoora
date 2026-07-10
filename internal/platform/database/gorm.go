package database

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// PoolConfig sizes the underlying *sql.DB connection pool. Zero-valued fields
// fall back to conservative defaults so callers can pass a partial config.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (p PoolConfig) withDefaults() PoolConfig {
	if p.MaxOpenConns <= 0 {
		p.MaxOpenConns = 25
	}
	if p.MaxIdleConns <= 0 {
		p.MaxIdleConns = 10
	}
	if p.ConnMaxLifetime <= 0 {
		p.ConnMaxLifetime = 5 * time.Minute
	}
	if p.ConnMaxIdleTime <= 0 {
		p.ConnMaxIdleTime = 1 * time.Minute
	}
	return p
}

// NewConnection opens the primary Postgres pool. When replicaURL is non-empty a
// gorm dbresolver is registered so reads are routed to the replica and writes to
// the primary; empty replicaURL keeps all traffic on the primary (the default).
func NewConnection(databaseURL, replicaURL string, pool PoolConfig, slogLogger *slog.Logger, logQueries bool) (*gorm.DB, error) {
	pool = pool.withDefaults()

	logLevel := logger.Error
	if logQueries {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true, // ErrRecordNotFound is handled in repos, not a real error
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if replicaURL != "" {
		// dbresolver sends writes (and any explicitly-marked statements) to the
		// primary and reads to the replica. Pool limits are mirrored onto the
		// resolver-managed connections.
		resolver := dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(replicaURL)},
			Policy:   dbresolver.RandomPolicy{},
		}).
			SetMaxOpenConns(pool.MaxOpenConns).
			SetMaxIdleConns(pool.MaxIdleConns).
			SetConnMaxLifetime(pool.ConnMaxLifetime).
			SetConnMaxIdleTime(pool.ConnMaxIdleTime)
		if err := db.Use(resolver); err != nil {
			return nil, fmt.Errorf("registering read-replica resolver: %w", err)
		}
		if slogLogger != nil {
			slogLogger.Info("database read-replica resolver enabled")
		}
	}

	if slogLogger != nil {
		slogLogger.Info("database connection established")
	}
	return db, nil
}
