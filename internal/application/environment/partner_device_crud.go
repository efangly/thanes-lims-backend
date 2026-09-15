package environment

import (
	"context"
	"errors"
	"strings"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

type CreatePartnerDeviceUseCase struct {
	devices portenvironment.PartnerDeviceRepository
	gauges  portenvironment.GaugeRepository
}

func NewCreatePartnerDeviceUseCase(devices portenvironment.PartnerDeviceRepository, gauges portenvironment.GaugeRepository) *CreatePartnerDeviceUseCase {
	return &CreatePartnerDeviceUseCase{devices: devices, gauges: gauges}
}

type CreatePartnerDeviceInput struct {
	Serial   string
	Location string
	Active   bool
}

// Execute creates a Partner Device mapping. Location must reference an
// existing Gauge - a Partner Device is never allowed to auto-create one
// (see CONTEXT.md#environment).
func (uc *CreatePartnerDeviceUseCase) Execute(ctx context.Context, in CreatePartnerDeviceInput) (environment.PartnerDevice, error) {
	d := environment.PartnerDevice{
		Serial:   strings.TrimSpace(in.Serial),
		Location: strings.TrimSpace(in.Location),
		Active:   in.Active,
	}
	if err := d.Validate(); err != nil {
		return environment.PartnerDevice{}, err
	}

	if _, err := uc.gauges.FindByLocation(ctx, d.Location); err != nil {
		return environment.PartnerDevice{}, err
	}

	if _, err := uc.devices.FindBySerial(ctx, d.Serial); err == nil {
		return environment.PartnerDevice{}, shared.ErrConflict
	} else if !errors.Is(err, shared.ErrNotFound) {
		return environment.PartnerDevice{}, err
	}

	return uc.devices.Create(ctx, d)
}

type UpdatePartnerDeviceUseCase struct {
	devices portenvironment.PartnerDeviceRepository
	gauges  portenvironment.GaugeRepository
}

func NewUpdatePartnerDeviceUseCase(devices portenvironment.PartnerDeviceRepository, gauges portenvironment.GaugeRepository) *UpdatePartnerDeviceUseCase {
	return &UpdatePartnerDeviceUseCase{devices: devices, gauges: gauges}
}

type UpdatePartnerDeviceInput struct {
	Serial   string
	Location string
	Active   bool
}

// Execute updates the Location and/or Active flag of an existing Partner
// Device. Active: false pauses polling without deleting the mapping.
func (uc *UpdatePartnerDeviceUseCase) Execute(ctx context.Context, in UpdatePartnerDeviceInput) (environment.PartnerDevice, error) {
	existing, err := uc.devices.FindBySerial(ctx, in.Serial)
	if err != nil {
		return environment.PartnerDevice{}, err
	}

	existing.Location = strings.TrimSpace(in.Location)
	existing.Active = in.Active
	if err := existing.Validate(); err != nil {
		return environment.PartnerDevice{}, err
	}

	if _, err := uc.gauges.FindByLocation(ctx, existing.Location); err != nil {
		return environment.PartnerDevice{}, err
	}

	return uc.devices.Update(ctx, existing)
}

type ListPartnerDevicesUseCase struct {
	devices portenvironment.PartnerDeviceRepository
}

func NewListPartnerDevicesUseCase(devices portenvironment.PartnerDeviceRepository) *ListPartnerDevicesUseCase {
	return &ListPartnerDevicesUseCase{devices: devices}
}

func (uc *ListPartnerDevicesUseCase) Execute(ctx context.Context) ([]environment.PartnerDevice, error) {
	return uc.devices.List(ctx)
}

type GetPartnerDeviceUseCase struct {
	devices portenvironment.PartnerDeviceRepository
}

func NewGetPartnerDeviceUseCase(devices portenvironment.PartnerDeviceRepository) *GetPartnerDeviceUseCase {
	return &GetPartnerDeviceUseCase{devices: devices}
}

func (uc *GetPartnerDeviceUseCase) Execute(ctx context.Context, serial string) (environment.PartnerDevice, error) {
	return uc.devices.FindBySerial(ctx, serial)
}
