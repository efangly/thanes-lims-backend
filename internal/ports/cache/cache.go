// Package cache defines the port for the Redis-backed read-through Cache
// described in docs/adr/0005-redis-cache-for-refresh-tokens-and-location-full-path.md.
// Postgres remains the source of truth for every cached value - this port
// exists purely to reduce read load, never as the only place a value lives.
package cache

import (
	"context"
	"time"
)

// Cache is a generic key-value port shared by every cache use case (Refresh
// Token fast-path lookups, Location Full Path). Values are opaque bytes -
// callers own their own (de)serialization.
type Cache interface {
	// Get returns shared.ErrNotFound if key does not exist (including on
	// TTL expiry). Callers must treat that as a cache miss, not a failure,
	// except where the caller's own fail-closed policy says otherwise.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores value under key with the given TTL. A ttl of 0 means no
	// expiry - only used by entries invalidated explicitly via Delete.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes key. Deleting a key that doesn't exist is not an error.
	Delete(ctx context.Context, key string) error
}
