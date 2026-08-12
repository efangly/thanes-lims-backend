package inventory

import "github.com/efangly/thanes-lims-backend/internal/domain/inventory"

type CreateItemRequest struct {
	Name     string `json:"name" validate:"required"`
	Category string `json:"category" validate:"required"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit" validate:"required"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" validate:"gte=0"`
}

type ItemResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Pct      int    `json:"pct"`
	Status   string `json:"status"`
}

func toResponse(i inventory.InventoryItem) ItemResponse {
	return ItemResponse{
		ID:       i.ID,
		Name:     i.Name,
		Category: i.Category,
		Quantity: i.Quantity,
		Unit:     i.Unit,
		Min:      i.Min,
		Max:      i.Max,
		Pct:      i.Pct(),
		Status:   string(i.DerivedStatus()),
	}
}
