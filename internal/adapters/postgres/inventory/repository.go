package inventory

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) inventory.InventoryItem {
	return inventory.InventoryItem{
		ID:            m.ID,
		Name:          m.Name,
		Category:      m.Category,
		Quantity:      m.Quantity,
		Unit:          m.Unit,
		Min:           m.Min,
		Max:           m.Max,
		DefaultVendor: m.DefaultVendor,
	}
}

func toModel(i inventory.InventoryItem) Model {
	return Model{
		ID:            i.ID,
		Name:          i.Name,
		Category:      i.Category,
		Quantity:      i.Quantity,
		Unit:          i.Unit,
		Min:           i.Min,
		Max:           i.Max,
		DefaultVendor: i.DefaultVendor,
	}
}

func (r *Repository) Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	m := toModel(i)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (inventory.InventoryItem, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.InventoryItem{}, shared.ErrNotFound
	}
	if err != nil {
		return inventory.InventoryItem{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]inventory.InventoryItem, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]inventory.InventoryItem, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryItem, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("quantity", quantity).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("default_vendor", vendor).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	return r.FindByID(ctx, id)
}
