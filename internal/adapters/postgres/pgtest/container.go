//go:build integration

// Package pgtest spins up a disposable Postgres container (via
// testcontainers-go) and applies migrations/ against it, for use by
// repository integration tests under internal/adapters/postgres/**.
package pgtest

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// migrationsDir resolves the repo-root migrations/ directory relative to
// this source file, so tests work regardless of the caller's working dir.
func migrationsDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// this file: internal/adapters/postgres/pgtest/container.go
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "migrations")
}

// SetupPostgres starts a Postgres container, applies all migrations, and
// returns a ready-to-use *gorm.DB. The container is terminated automatically
// via t.Cleanup.
func SetupPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("thanes_lims_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("pgtest: start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(context.Background()); err != nil {
			t.Logf("pgtest: terminate container: %v", err)
		}
	})

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("pgtest: connection string: %v", err)
	}

	if err := applyMigrations(dsn); err != nil {
		t.Fatalf("pgtest: apply migrations: %v", err)
	}

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("pgtest: gorm open: %v", err)
	}

	return gdb
}

func applyMigrations(dsn string) error {
	m, err := migrate.New("file://"+migrationsDir(), dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
