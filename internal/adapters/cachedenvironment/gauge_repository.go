// Package cachedenvironment decorates a ports/environment.GaugeRepository,
// adding a Redis read-through Cache in front of FindByLocation, per
// docs/adr/0006-redis-cache-for-rbac-permissions-and-gauge-thresholds.md.
// List passes straight through - it isn't the hot path (called once per
// sensor reading via FindByLocation, not per gauge listing).
package cachedenvironment

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// gaugeTTL matches the other Redis cache entries: no endpoint creates or
// updates a gauges row (only List), so a short, fixed staleness window
// needs no invalidation path. Fail-open: any cache miss or error, including
// Redis being unreachable, falls back to Postgres exactly as before (see
// ADR 0006).
const gaugeTTL = 15 * time.Minute

type CachedGaugeRepository struct {
	next  portenvironment.GaugeRepository
	cache cache.Cache
}

func NewCachedGaugeRepository(next portenvironment.GaugeRepository, c cache.Cache) *CachedGaugeRepository {
	return &CachedGaugeRepository{next: next, cache: c}
}

func gaugeCacheKey(location string) string {
	return "env:gauge:" + location
}

func (r *CachedGaugeRepository) List(ctx context.Context) ([]environment.Gauge, error) {
	return r.next.List(ctx)
}

// FindByLocation returns the cached gauge when present; any cache miss or
// error falls straight back to Postgres, matching pre-cache behavior.
func (r *CachedGaugeRepository) FindByLocation(ctx context.Context, location string) (environment.Gauge, error) {
	key := gaugeCacheKey(location)

	if data, err := r.cache.Get(ctx, key); err == nil {
		var g environment.Gauge
		if decErr := gob.NewDecoder(bytes.NewReader(data)).Decode(&g); decErr == nil {
			return g, nil
		}
	}

	g, err := r.next.FindByLocation(ctx, location)
	if err != nil {
		return environment.Gauge{}, err
	}

	var buf bytes.Buffer
	if encErr := gob.NewEncoder(&buf).Encode(g); encErr != nil {
		log.Printf("cachedenvironment: encode gauge for cache: %v", encErr)
		return g, nil
	}
	if err := r.cache.Set(ctx, key, buf.Bytes(), gaugeTTL); err != nil {
		log.Printf("cachedenvironment: populate gauge cache: %v", err)
	}
	return g, nil
}
