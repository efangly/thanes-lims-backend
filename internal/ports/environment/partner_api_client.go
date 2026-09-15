package environment

import (
	"context"
	"time"
)

// PartnerDeviceMetadata is the curated device snapshot returned by
// GET /log/partner/devices/{serial} - see docs/partner-api-guide.md.
type PartnerDeviceMetadata struct {
	Serial   string
	Name     string
	Status   bool
	Firmware string
	Online   bool
}

// PartnerDeviceReading is one telemetry row from
// GET /log/partner/devices/{serial}/timeseries.
type PartnerDeviceReading struct {
	Serial          string
	SendTime        time.Time
	TempDisplay     float64
	HumidityDisplay float64
}

// PartnerAPIClient is the outbound port to SMtrack's Partner API (see
// docs/partner-api-guide.md). Implementations are responsible for
// X-API-Key auth and for translating the Partner API's error envelope
// (401/404/429/400) into errors the poller can distinguish - see the
// partnerapi adapter's APIError.
type PartnerAPIClient interface {
	FetchMetadata(ctx context.Context, serial string) (PartnerDeviceMetadata, error)
	// FetchLatestReading returns the most recent reading in the last 24h.
	// found is false (with a nil error) when the device has no reading in
	// that window - not the same as an API error.
	FetchLatestReading(ctx context.Context, serial string) (reading PartnerDeviceReading, found bool, err error)
}

// RetryableError distinguishes a transient PartnerAPIClient failure (rate
// limited, a 5xx, or a network/timeout error) from a permanent one
// (401/404) - see ADR 0011's stale-cache-fallback decision. Adapters
// implement it on their error types (see partnerapi.APIError); the poller
// uses errors.As against this interface so it never imports the adapter
// package directly.
type RetryableError interface {
	error
	Retryable() bool
}
