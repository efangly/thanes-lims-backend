package environment

import (
	"context"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// DiscoverPartnerDevicesByWardUseCase browses SMtrack's own device
// inventory for a given ward, to help an admin find a device's serial
// before creating a PartnerDevice (Serial<->Location) mapping. This is
// read-through to the Partner API - nothing here is persisted (see ADR
// 0011's "proxy, don't persist" decision, extended to this new read).
type DiscoverPartnerDevicesByWardUseCase struct {
	client portenvironment.PartnerAPIClient
}

func NewDiscoverPartnerDevicesByWardUseCase(client portenvironment.PartnerAPIClient) *DiscoverPartnerDevicesByWardUseCase {
	return &DiscoverPartnerDevicesByWardUseCase{client: client}
}

type DiscoverPartnerDevicesByWardInput struct {
	Ward  string
	Page  int
	Limit int
}

// Execute lists SMtrack's devices in in.Ward. Page/Limit are passed through
// as-is (the Partner API documents its own defaults of 1/20 and a max of
// 100) except Limit is defensively clamped to that max here, since a caller
// requesting an unbounded page size is a client bug worth catching before
// it reaches SMtrack.
func (uc *DiscoverPartnerDevicesByWardUseCase) Execute(ctx context.Context, in DiscoverPartnerDevicesByWardInput) (portenvironment.PartnerDeviceListing, error) {
	ward := strings.TrimSpace(in.Ward)
	if ward == "" {
		return portenvironment.PartnerDeviceListing{}, shared.ErrValidation
	}

	limit := in.Limit
	if limit > 100 {
		limit = 100
	}

	return uc.client.ListDevicesByWard(ctx, ward, in.Page, limit)
}
