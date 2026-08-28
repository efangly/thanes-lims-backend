package equipment

import (
	"context"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type UpdateEquipmentUseCase struct {
	equipment portequipment.Repository
	vendors   portequipment.VendorDirectory
	locations portequipment.LocationDirectory
}

func NewUpdateEquipmentUseCase(
	equipment portequipment.Repository,
	vendors portequipment.VendorDirectory,
	locations portequipment.LocationDirectory,
) *UpdateEquipmentUseCase {
	return &UpdateEquipmentUseCase{equipment: equipment, vendors: vendors, locations: locations}
}

// UpdateEquipmentInput carries a partial update: a nil field is left
// untouched, a non-nil field is written (an empty string / empty pointer
// clears the underlying column). LastCalibratedAt / NextCalibrationDue are
// deliberately not editable here - they move only through RecordCalibration.
type UpdateEquipmentInput struct {
	ID               string
	Name             *string
	TypeCode         *string
	SerialNumber     *string
	Category         *string
	Manufacturer     *string
	Model            *string
	InstallationDate *time.Time
	ClearInstallDate bool
	// VendorID / LocationID: nil = unchanged, non-nil "" = clear, non-nil
	// value = set (validated).
	VendorID   *string
	LocationID *string
}

func (uc *UpdateEquipmentUseCase) Execute(ctx context.Context, in UpdateEquipmentInput) (equipment.Equipment, error) {
	e, err := uc.equipment.FindByID(ctx, in.ID)
	if err != nil {
		return equipment.Equipment{}, err
	}

	if in.Name != nil {
		e.Name = strings.TrimSpace(*in.Name)
	}
	if in.TypeCode != nil {
		e.TypeCode = strings.TrimSpace(*in.TypeCode)
	}
	if in.SerialNumber != nil {
		e.SerialNumber = strings.TrimSpace(*in.SerialNumber)
	}
	if in.Category != nil {
		e.Category = strings.TrimSpace(*in.Category)
	}
	if in.Manufacturer != nil {
		e.Manufacturer = strings.TrimSpace(*in.Manufacturer)
	}
	if in.Model != nil {
		e.Model = strings.TrimSpace(*in.Model)
	}
	if in.ClearInstallDate {
		e.InstallationDate = nil
	} else if in.InstallationDate != nil {
		e.InstallationDate = in.InstallationDate
	}

	if in.VendorID != nil {
		ref := optionalRef(*in.VendorID)
		if err := validateVendor(ctx, uc.vendors, ref); err != nil {
			return equipment.Equipment{}, err
		}
		e.VendorID = ref
	}
	if in.LocationID != nil {
		ref := optionalRef(*in.LocationID)
		if err := validateLocation(ctx, uc.locations, ref); err != nil {
			return equipment.Equipment{}, err
		}
		e.LocationID = ref
	}

	return uc.equipment.Update(ctx, e)
}
