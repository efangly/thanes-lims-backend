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
	// Battery is the device's battery level in percent (0-100).
	Battery int
	// Plug is true when the device is on external/mains power.
	Plug bool
	// Door1/Door2/Door3 are true when that door is open - most devices
	// only use Door1 (single-door cabinet); the others stay false.
	Door1 bool
	Door2 bool
	Door3 bool
	// ExtMemory is true when the device's SD/external memory card is
	// present.
	ExtMemory bool
	FetchedAt time.Time
	// Stale is true when this snapshot was served from cache because a
	// live poll of the Partner API failed (e.g. 429/timeout) - see ADR
	// 0011 and the stale-cache-fallback decision.
	Stale bool
}
