// Package redis implements the internal/ports/cache.Cache port against a
// real Redis instance via go-redis.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	goredis "github.com/redis/go-redis/v9"
)

type Adapter struct {
	client *goredis.Client
}

// New parses redisURL (e.g. "redis://user:pass@host:6379") and verifies
// connectivity with a PING - callers should treat a non-nil error as fatal
// at boot, matching how the Postgres connection is handled.
func New(ctx context.Context, redisURL string) (*Adapter, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}

	client := goredis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return &Adapter{client: client}, nil
}

// Ping verifies Redis is reachable. It's used by the /health endpoint so
// external monitoring can alert on a Redis outage: per ADR 0005 the refresh
// path is fail-closed, so a Redis outage forces every user to re-login once
// their access token expires (~15m) and must be treated as an incident.
func (a *Adapter) Ping(ctx context.Context) error {
	if err := a.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping: %w", err)
	}
	return nil
}

func (a *Adapter) Close() error {
	return a.client.Close()
}

func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := a.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("redis: get: %w", err)
	}
	return val, nil
}

func (a *Adapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := a.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis: set: %w", err)
	}
	return nil
}

func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := a.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis: delete: %w", err)
	}
	return nil
}
