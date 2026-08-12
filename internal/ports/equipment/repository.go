package equipment

import (
	"context"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
)

type Repository interface {
	Create(ctx context.Context, e equipment.Equipment) (equipment.Equipment, error)
	FindByID(ctx context.Context, id string) (equipment.Equipment, error)
	List(ctx context.Context) ([]equipment.Equipment, error)
	Update(ctx context.Context, e equipment.Equipment) (equipment.Equipment, error)
}
