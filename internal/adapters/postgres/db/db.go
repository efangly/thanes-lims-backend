package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New opens a GORM connection against Postgres and tunes the underlying
// connection pool. Schema is owned entirely by migrations/ - GORM AutoMigrate
// is never called here.
func New(dsn string) (*gorm.DB, error) {
	// IgnoreRecordNotFoundError: repositories return shared.ErrNotFound as
	// an expected, handled outcome (e.g. "no open alert for this location")
	// - GORM's default logger would otherwise log every one of those as an
	// error and drown out real problems.
	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
	})

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("db: underlying sql.DB: %w", err)
	}

	// Cloud demo Postgres instances tend to cap connections aggressively.
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return gdb, nil
}
