package environment

import (
	"context"

	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// GetPartnerDeviceTimeseriesUseCase reads the full trailing 1h telemetry
// window for a serial straight from the Partner API, for on-demand chart
// rendering. Unlike GetPartnerDeviceSnapshotUseCase (which only ever serves
// the single latest point, from cache), this always calls out to SMtrack
// live and is never cached - only wired up when PARTNER_API_ENABLED, same
// as DiscoverPartnerDevicesByWardUseCase.
type GetPartnerDeviceTimeseriesUseCase struct {
	client portenvironment.PartnerAPIClient
}

func NewGetPartnerDeviceTimeseriesUseCase(client portenvironment.PartnerAPIClient) *GetPartnerDeviceTimeseriesUseCase {
	return &GetPartnerDeviceTimeseriesUseCase{client: client}
}

func (uc *GetPartnerDeviceTimeseriesUseCase) Execute(ctx context.Context, serial string) ([]portenvironment.PartnerDeviceReading, error) {
	return uc.client.FetchTimeseries(ctx, serial)
}
