package environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
)

// GetPartnerDeviceSnapshotUseCase serves the last polled snapshot straight
// from cache - it never calls the Partner API itself (PollPartnerDevicesJob
// owns that on its own schedule). Returns shared.ErrNotFound if the device
// has never been polled or its cache entry has passed staleMax (see ADR
// 0011).
type GetPartnerDeviceSnapshotUseCase struct {
	cache    cache.Cache
	cacheTTL time.Duration
}

func NewGetPartnerDeviceSnapshotUseCase(c cache.Cache, cacheTTL time.Duration) *GetPartnerDeviceSnapshotUseCase {
	return &GetPartnerDeviceSnapshotUseCase{cache: c, cacheTTL: cacheTTL}
}

func (uc *GetPartnerDeviceSnapshotUseCase) Execute(ctx context.Context, serial string) (environment.PartnerDeviceSnapshot, error) {
	data, err := uc.cache.Get(ctx, partnerSnapshotCacheKey(serial))
	if err != nil {
		return environment.PartnerDeviceSnapshot{}, shared.ErrNotFound
	}
	var snap environment.PartnerDeviceSnapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return environment.PartnerDeviceSnapshot{}, shared.ErrNotFound
	}
	if time.Since(snap.FetchedAt) > uc.cacheTTL {
		snap.Stale = true
	}
	return snap, nil
}
