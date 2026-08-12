package equipment

import (
	"context"
	"fmt"
	"time"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
)

type CreateEquipmentUseCase struct {
	equipment portequipment.Repository
	idgen     portidgen.SequenceGenerator
}

func NewCreateEquipmentUseCase(equipment portequipment.Repository, idgen portidgen.SequenceGenerator) *CreateEquipmentUseCase {
	return &CreateEquipmentUseCase{equipment: equipment, idgen: idgen}
}

type CreateEquipmentInput struct {
	Name               string
	TypeCode           string
	NextCalibrationDue time.Time
}

// Execute generates EQ-{TYPE}-{seq3}, counted per TypeCode (each equipment
// category gets its own sequence).
func (uc *CreateEquipmentUseCase) Execute(ctx context.Context, in CreateEquipmentInput) (equipment.Equipment, error) {
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
	}

	return uc.equipment.Create(ctx, e)
}
