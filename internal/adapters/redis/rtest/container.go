//go:build integration

// Package rtest spins up a disposable Redis container (via testcontainers-go)
// for use by adapter integration tests, mirroring
// internal/adapters/postgres/pgtest for Postgres.
package rtest

import (
	"context"
	"testing"

	redisadapter "github.com/efangly/thanes-lims-backend/internal/adapters/redis"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// SetupRedis starts a Redis container and returns a ready-to-use
// *redisadapter.Adapter. The container is terminated automatically via
// t.Cleanup.
func SetupRedis(t *testing.T) *redisadapter.Adapter {
	t.Helper()

	ctx := context.Background()

	ctr, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("rtest: start container: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(context.Background()); err != nil {
			t.Logf("rtest: terminate container: %v", err)
		}
	})

	connStr, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("rtest: connection string: %v", err)
	}

	adapter, err := redisadapter.New(ctx, connStr)
	if err != nil {
		t.Fatalf("rtest: connect: %v", err)
	}
	t.Cleanup(func() {
		_ = adapter.Close()
	})

	return adapter
}
