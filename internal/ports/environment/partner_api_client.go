package environment

import (
	"context"
	"time"
)

// PartnerDeviceMetadata is the curated device snapshot returned by the
// Partner API's GetDeviceSnapshot/ListDevicesByWard RPCs - see
// docs/partner-api-guide.md and proto/partner/partner.proto.
type PartnerDeviceMetadata struct {
	Serial   string
	Name     string
	Status   bool
	Firmware string
	Online   bool
}

// PartnerDeviceReading is one telemetry point, as returned by
// GetDeviceSnapshot's timeseries (trailing 1h, newest first) or
// ListDevicesByWard's per-device latest reading.
type PartnerDeviceReading struct {
	Serial          string
	SendTime        time.Time
	TempDisplay     float64
	HumidityDisplay float64
}

// PartnerDeviceWithReading pairs a device (as returned by ListDevicesByWard)
// with its single latest reading, if it has one. Named to avoid colliding
// with the domain.PartnerDevice type (the admin-managed Serial<->Location
// mapping) - this is a read-through result from SMtrack, never persisted.
type PartnerDeviceWithReading struct {
	Device    PartnerDeviceMetadata
	Latest    PartnerDeviceReading
	HasLatest bool // false when the device has no telemetry yet
}

// PartnerDeviceListing is one page of ListDevicesByWard's results.
type PartnerDeviceListing struct {
	Devices []PartnerDeviceWithReading
	Total   int
	Page    int
	Limit   int
}

// PartnerAPIClient is the outbound port to SMtrack's Partner API (see
// docs/partner-api-guide.md, proto/partner/partner.proto). Implementations
// are responsible for "x-api-key" auth and for translating the Partner
// API's gRPC status codes into errors the poller can distinguish - see the
// partnergrpc adapter's RPCError.
type PartnerAPIClient interface {
	// FetchSnapshot returns the device's metadata plus its most recent
	// reading in the trailing 1h window (GetDeviceSnapshot's
	// timeseries[0], newest first). found is false (with a nil error)
	// when the device has no reading in that window - not the same as an
	// API error.
	FetchSnapshot(ctx context.Context, serial string) (meta PartnerDeviceMetadata, reading PartnerDeviceReading, found bool, err error)

	// ListDevicesByWard returns every device in ward that the configured
	// API key is scoped to, each with its single latest reading if it has
	// one. Used to help an admin discover a device's serial before
	// creating a PartnerDevice (Serial<->Location) mapping - read-through,
	// never persisted (see ADR 0011).
	ListDevicesByWard(ctx context.Context, ward string, page, limit int) (PartnerDeviceListing, error)
}

// RetryableError distinguishes a transient PartnerAPIClient failure (rate
// limited, unavailable, timed out, or a network error) from a permanent one
// (unauthenticated/not found/invalid argument) - see ADR 0011's
// stale-cache-fallback decision. Adapters implement it on their error types
// (see partnergrpc.RPCError); the poller uses errors.As against this
// interface so it never imports the adapter package directly.
type RetryableError interface {
	error
	Retryable() bool
}
