package environment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
)

type GaugeRepository interface {
	List(ctx context.Context) ([]environment.Gauge, error)
	FindByLocation(ctx context.Context, location string) (environment.Gauge, error)
}

type ReadingRepository interface {
	Record(ctx context.Context, r environment.SensorReading) (environment.SensorReading, error)
	LatestByLocation(ctx context.Context, location string) (environment.SensorReading, error)
	ListTrend(ctx context.Context, location string, limit int) ([]environment.SensorReading, error)
}

// AlertRepository.FindOpenByLocation returns (nil, nil) when no alert is
// currently open for the location - that's the normal case, not an error.
type AlertRepository interface {
	Create(ctx context.Context, a environment.EnvAlert) (environment.EnvAlert, error)
	FindOpenByLocation(ctx context.Context, location string) (*environment.EnvAlert, error)
	List(ctx context.Context) ([]environment.EnvAlert, error)
	Resolve(ctx context.Context, id int64) error
}

// AlertBroadcaster pushes a newly created/escalated alert to any real-time
// subscribers (e.g. the WebSocket hub). Kept separate from Notifier since
// this is a transient, best-effort push, not a persisted notification.
type AlertBroadcaster interface {
	Broadcast(a environment.EnvAlert)
}

// PartnerDeviceRepository is the persistence port for the Serial<->Location
// mapping (admin-managed, see CONTEXT.md#environment). There is
// deliberately no Delete beyond what Update(Active: false) already covers -
// deactivating pauses polling without losing the mapping.
type PartnerDeviceRepository interface {
	Create(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error)
	Update(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error)
	FindBySerial(ctx context.Context, serial string) (environment.PartnerDevice, error)
	List(ctx context.Context) ([]environment.PartnerDevice, error)
}

// PartnerDeviceBroadcaster pushes a freshly polled PartnerDeviceSnapshot to
// any real-time subscribers (the SSE stream). Same shape/purpose as
// AlertBroadcaster - transient, best-effort push, not persisted.
type PartnerDeviceBroadcaster interface {
	Broadcast(s environment.PartnerDeviceSnapshot)
}
