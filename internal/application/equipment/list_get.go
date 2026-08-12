package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	portequipment "github.com/efangly/thanes-lims-backend/internal/ports/equipment"
)

type ListEquipmentUseCase struct {
	equipment portequipment.Repository
}

func NewListEquipmentUseCase(equipment portequipment.Repository) *ListEquipmentUseCase {
	return &ListEquipmentUseCase{equipment: equipment}
}

func (uc *ListEquipmentUseCase) Execute(ctx context.Context) ([]equipment.Equipment, error) {
	return uc.equipment.List(ctx)
}

type GetEquipmentUseCase struct {
	equipment portequipment.Repository
}

func NewGetEquipmentUseCase(equipment portequipment.Repository) *GetEquipmentUseCase {
	return &GetEquipmentUseCase{equipment: equipment}
}

func (uc *GetEquipmentUseCase) Execute(ctx context.Context, id string) (equipment.Equipment, error) {
	return uc.equipment.FindByID(ctx, id)
}
