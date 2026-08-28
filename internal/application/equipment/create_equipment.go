package equipment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

type CreateEquipmentUseCase struct {
	equipment portequipment.Repository
	idgen     portidgen.SequenceGenerator
	vendors   portequipment.VendorDirectory
	locations portequipment.LocationDirectory
}

func NewCreateEquipmentUseCase(
	equipment portequipment.Repository,
	idgen portidgen.SequenceGenerator,
	vendors portequipment.VendorDirectory,
	locations portequipment.LocationDirectory,
) *CreateEquipmentUseCase {
	return &CreateEquipmentUseCase{equipment: equipment, idgen: idgen, vendors: vendors, locations: locations}
}

type CreateEquipmentInput struct {
	Name               string
	TypeCode           string
	NextCalibrationDue time.Time

	SerialNumber     string
	Category         string
	Manufacturer     string
	Model            string
	InstallationDate *time.Time
	VendorID         string
	LocationID       string
}

// Execute generates EQ-{TYPE}-{seq3}, counted per TypeCode (each equipment
// category gets its own sequence). VendorID / LocationID, when supplied,
// must resolve to an existing Vendor and an equipment_storage Location.
func (uc *CreateEquipmentUseCase) Execute(ctx context.Context, in CreateEquipmentInput) (equipment.Equipment, error) {
	vendorID := optionalRef(in.VendorID)
	locationID := optionalRef(in.LocationID)
	if err := validateVendor(ctx, uc.vendors, vendorID); err != nil {
		return equipment.Equipment{}, err
	}
	if err := validateLocation(ctx, uc.locations, locationID); err != nil {
		return equipment.Equipment{}, err
	}

	seq, err := uc.idgen.Next(ctx, "equipment_"+in.TypeCode, nil)
	if err != nil {
		return equipment.Equipment{}, err
	}

	now := time.Now()
	e := equipment.Equipment{
		ID:                 fmt.Sprintf("EQ-%s-%03d", in.TypeCode, seq),
		Name:               in.Name,
		TypeCode:           in.TypeCode,
		LastCalibratedAt:   now,
		NextCalibrationDue: in.NextCalibrationDue,
		SerialNumber:       strings.TrimSpace(in.SerialNumber),
		Category:           strings.TrimSpace(in.Category),
		Manufacturer:       strings.TrimSpace(in.Manufacturer),
		Model:              strings.TrimSpace(in.Model),
		InstallationDate:   in.InstallationDate,
		VendorID:           vendorID,
		LocationID:         locationID,
	}

	return uc.equipment.Create(ctx, e)
}
