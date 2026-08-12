package inventory

import (
	"context"
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	portidgen "github.com/efangly/thanes-lims-backend/internal/ports/idgen"
	portinventory "github.com/efangly/thanes-lims-backend/internal/ports/inventory"
)

type CreateItemUseCase struct {
	items portinventory.Repository
	idgen portidgen.SequenceGenerator
}

func NewCreateItemUseCase(items portinventory.Repository, idgen portidgen.SequenceGenerator) *CreateItemUseCase {
	return &CreateItemUseCase{items: items, idgen: idgen}
}

type CreateItemInput struct {
	Name     string
	Category string
	Quantity int
	Unit     string
	Min      int
	Max      int
}

func (uc *CreateItemUseCase) Execute(ctx context.Context, in CreateItemInput) (inventory.InventoryItem, error) {
	seq, err := uc.idgen.Next(ctx, "inventory", nil)
	if err != nil {
		return inventory.InventoryItem{}, err
	}

	item := inventory.InventoryItem{
		ID:       fmt.Sprintf("INV-%05d", seq),
		Name:     in.Name,
		Category: in.Category,
		Quantity: in.Quantity,
		Unit:     in.Unit,
		Min:      in.Min,
		Max:      in.Max,
	}

	return uc.items.Create(ctx, item)
}
