package environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/ports/cache"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

func partnerSnapshotCacheKey(serial string) string {
	return "env:partnerdevice:snapshot:" + serial
}

// PollPartnerDeviceUseCase polls one Partner Device's metadata + latest
// reading from the Partner API, evaluates it through the same
// EvaluateThresholdsUseCase pipeline as any other Location, caches the
// resulting snapshot, and broadcasts it over SSE. See CONTEXT.md#environment
// and ADR 0011 (no raw persistence, proxy + short-TTL cache only).
type PollPartnerDeviceUseCase struct {
	client      portenvironment.PartnerAPIClient
	cache       cache.Cache
	evaluate    *EvaluateThresholdsUseCase
	broadcaster portenvironment.PartnerDeviceBroadcaster
	// cacheTTL is the freshness window: a snapshot older than this is
	// marked Stale when served, but is still usable as a stale-cache
	// fallback until staleMax.
	cacheTTL time.Duration
	// staleMax is the Redis TTL on the cache entry itself - the hard
	// cutoff after which a failed poll has nothing left to fall back to
	// and Execute returns a real error (decided as 5 minutes).
	staleMax time.Duration
}

func NewPollPartnerDeviceUseCase(
	client portenvironment.PartnerAPIClient,
	c cache.Cache,
	evaluate *EvaluateThresholdsUseCase,
	broadcaster portenvironment.PartnerDeviceBroadcaster,
	cacheTTL, staleMax time.Duration,
) *PollPartnerDeviceUseCase {
	return &PollPartnerDeviceUseCase{client: client, cache: c, evaluate: evaluate, broadcaster: broadcaster, cacheTTL: cacheTTL, staleMax: staleMax}
}

// Execute polls, caches and broadcasts a fresh snapshot. On a Retryable
// failure (rate limited, network/timeout, 5xx) it falls back to the last
// cached snapshot marked Stale instead of failing outright; a permanent
// failure (401/404), or a Retryable failure with nothing cached to fall
// back to, returns an error.
func (uc *PollPartnerDeviceUseCase) Execute(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDeviceSnapshot, error) {
	snap, err := uc.fetch(ctx, d)
	if err == nil {
		uc.store(ctx, d.Serial, snap)
		uc.broadcaster.Broadcast(snap)
		return snap, nil
	}

	if !isRetryable(err) {
		return environment.PartnerDeviceSnapshot{}, err
	}

	cached, ok := uc.load(ctx, d.Serial)
	if !ok {
		return environment.PartnerDeviceSnapshot{}, err
	}
	cached.Stale = true
	uc.broadcaster.Broadcast(cached)
	return cached, nil
}

func (uc *PollPartnerDeviceUseCase) fetch(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDeviceSnapshot, error) {
	meta, reading, found, err := uc.client.FetchSnapshot(ctx, d.Serial)
	if err != nil {
		return environment.PartnerDeviceSnapshot{}, err
	}

	snap := environment.PartnerDeviceSnapshot{
		Serial:    d.Serial,
		Location:  d.Location,
		Name:      meta.Name,
		Status:    meta.Status,
		Firmware:  meta.Firmware,
		Online:    meta.Online,
		FetchedAt: time.Now(),
	}
	// found == false (no reading in the last 24h) leaves TempDisplay/
	// HumidityDisplay/SendTime/Level/Battery/Plug/Door*/ExtMemory at their
	// zero values - there is nothing to evaluate an alert against yet.
	if found {
		snap.TempDisplay = reading.TempDisplay
		snap.HumidityDisplay = reading.HumidityDisplay
		snap.SendTime = reading.SendTime
		snap.Battery = reading.Battery
		snap.Plug = reading.Plug
		snap.Door1 = reading.Door1
		snap.Door2 = reading.Door2
		snap.Door3 = reading.Door3
		snap.ExtMemory = reading.ExtMemory

		alert, err := uc.evaluate.Execute(ctx, d.Location, reading.TempDisplay)
		if err != nil {
			return environment.PartnerDeviceSnapshot{}, err
		}
		snap.Level = environment.LevelOK
		if alert != nil {
			snap.Level = alert.Level
		}
	}
	return snap, nil
}

func (uc *PollPartnerDeviceUseCase) store(ctx context.Context, serial string, snap environment.PartnerDeviceSnapshot) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		log.Printf("partnerdevice: encode snapshot for cache: %v", err)
		return
	}
	if err := uc.cache.Set(ctx, partnerSnapshotCacheKey(serial), buf.Bytes(), uc.staleMax); err != nil {
		log.Printf("partnerdevice: populate snapshot cache: %v", err)
	}
}

func (uc *PollPartnerDeviceUseCase) load(ctx context.Context, serial string) (environment.PartnerDeviceSnapshot, bool) {
	data, err := uc.cache.Get(ctx, partnerSnapshotCacheKey(serial))
	if err != nil {
		return environment.PartnerDeviceSnapshot{}, false
	}
	var snap environment.PartnerDeviceSnapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return environment.PartnerDeviceSnapshot{}, false
	}
	return snap, true
}

// isRetryable defaults to true for any error that doesn't implement
// RetryableError (a plain network/timeout error from the http.Client, for
// instance) - only a Partner API response we can positively identify as
// permanent (401/404 via partnerapi.APIError) returns false.
func isRetryable(err error) bool {
	var re portenvironment.RetryableError
	if errors.As(err, &re) {
		return re.Retryable()
	}
	return true
}
