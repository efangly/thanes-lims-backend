package inventory

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/inventory"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

// LotRepository persists InventoryLots. It is a separate type from
// Repository (items) but shares the package and *gorm.DB.
type LotRepository struct {
	db *gorm.DB
}

func NewLotRepository(db *gorm.DB) *LotRepository {
	return &LotRepository{db: db}
}

func lotToDomain(m LotModel) inventory.InventoryLot {
	return inventory.InventoryLot{
		ID:         m.ID,
		ItemID:     m.ItemID,
		LotNo:      m.LotNo,
		ExpireDate: m.ExpireDate,
		Quantity:   m.Quantity,
	}
}

func lotToModel(l inventory.InventoryLot) LotModel {
	return LotModel{
		ID:         l.ID,
		ItemID:     l.ItemID,
		LotNo:      l.LotNo,
		ExpireDate: l.ExpireDate,
		Quantity:   l.Quantity,
	}
}

func (r *LotRepository) Create(ctx context.Context, l inventory.InventoryLot) (inventory.InventoryLot, error) {
	m := lotToModel(l)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return inventory.InventoryLot{}, err
	}
	return lotToDomain(m), nil
}

func (r *LotRepository) FindByID(ctx context.Context, id string) (inventory.InventoryLot, error) {
	var m LotModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.InventoryLot{}, shared.ErrNotFound
	}
	if err != nil {
		return inventory.InventoryLot{}, err
	}
	return lotToDomain(m), nil
}

func (r *LotRepository) FindByItemAndLotNo(ctx context.Context, itemID, lotNo string) (inventory.InventoryLot, error) {
	var m LotModel
	err := r.db.WithContext(ctx).First(&m, "item_id = ? AND lot_no = ?", itemID, lotNo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.InventoryLot{}, shared.ErrNotFound
	}
	if err != nil {
		return inventory.InventoryLot{}, err
	}
	return lotToDomain(m), nil
}

func (r *LotRepository) ListByItem(ctx context.Context, itemID string) ([]inventory.InventoryLot, error) {
	var models []LotModel
	if err := r.db.WithContext(ctx).Where("item_id = ?", itemID).Order("expire_date NULLS LAST, lot_no").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]inventory.InventoryLot, len(models))
	for i, m := range models {
		out[i] = lotToDomain(m)
	}
	return out, nil
}

func (r *LotRepository) UpdateQuantity(ctx context.Context, id string, quantity int) (inventory.InventoryLot, error) {
	if err := r.db.WithContext(ctx).Model(&LotModel{}).Where("id = ?", id).Update("quantity", quantity).Error; err != nil {
		return inventory.InventoryLot{}, err
	}
	return r.FindByID(ctx, id)
}
