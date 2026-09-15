package environment

import "time"

// PartnerDeviceSnapshot is the polled, cached view of one Partner Device:
// the latest metadata + reading fetched from the Partner API, plus the
// Level derived from the reading against the linked Gauge. Never persisted
// as a Sensor Reading - SMtrack (the partner) stays the system of record
// for the raw history, this is a transient cache entry (see ADR 0011).
type PartnerDeviceSnapshot struct {
	Serial          string
	Location        string
	Name            string
	Status          bool
	Firmware        string
	Online          bool
	TempDisplay     float64
	HumidityDisplay float64
	SendTime        time.Time
	Level           Level
	FetchedAt       time.Time
	// Stale is true when this snapshot was served from cache because a
	// live poll of the Partner API failed (e.g. 429/timeout) - see ADR
	// 0011 and the stale-cache-fallback decision.
	Stale bool
}
