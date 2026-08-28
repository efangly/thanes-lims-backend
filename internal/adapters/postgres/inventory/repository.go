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

// toDomain leaves Quantity at zero; callers overlay the derived sum of the
// item's InventoryLots (see sumByItem).
func toDomain(m Model) inventory.InventoryItem {
	return inventory.InventoryItem{
		ID:            m.ID,
		Name:          m.Name,
		Category:      m.Category,
		Unit:          m.Unit,
		Min:           m.Min,
		Max:           m.Max,
		DefaultVendor: m.DefaultVendor,

		CustodianUserID: m.CustodianUserID,
		Manufacturer:    m.Manufacturer,
		VendorID:        m.VendorID,
		LocationID:      m.LocationID,
	}
}

func toModel(i inventory.InventoryItem) Model {
	return Model{
		ID:            i.ID,
		Name:          i.Name,
		Category:      i.Category,
		Unit:          i.Unit,
		Min:           i.Min,
		Max:           i.Max,
		DefaultVendor: i.DefaultVendor,

		CustodianUserID: i.CustodianUserID,
		Manufacturer:    i.Manufacturer,
		VendorID:        i.VendorID,
		LocationID:      i.LocationID,
	}
}

// sumByItem returns the derived on-hand quantity (sum of InventoryLot
// quantities) per item id. Items with no lots are simply absent from the
// map, which reads back as zero.
func (r *Repository) sumByItem(ctx context.Context, itemIDs ...string) (map[string]int, error) {
	type row struct {
		ItemID string
		Total  int
	}
	q := r.db.WithContext(ctx).Model(&LotModel{}).
		Select("item_id, COALESCE(SUM(quantity), 0) AS total").
		Group("item_id")
	if len(itemIDs) > 0 {
		q = q.Where("item_id IN ?", itemIDs)
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, x := range rows {
		out[x.ItemID] = x.Total
	}
	return out, nil
}

func (r *Repository) Create(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	m := toModel(i)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	// A freshly created item has no lots yet, so Quantity is zero.
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
	sums, err := r.sumByItem(ctx, id)
	if err != nil {
		return inventory.InventoryItem{}, err
	}
	item := toDomain(m)
	item.Quantity = sums[id]
	return item, nil
}

func (r *Repository) List(ctx context.Context) ([]inventory.InventoryItem, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	sums, err := r.sumByItem(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]inventory.InventoryItem, len(models))
	for i, m := range models {
		item := toDomain(m)
		item.Quantity = sums[m.ID]
		out[i] = item
	}
	return out, nil
}

// Update uses Save (not Updates) so cleared optional fields - a removed
// VendorID/LocationID - actually persist; GORM's Updates skips nil values.
// Callers load the full item via FindByID first and overlay changes.
func (r *Repository) Update(ctx context.Context, i inventory.InventoryItem) (inventory.InventoryItem, error) {
	m := toModel(i)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	return r.FindByID(ctx, i.ID)
}

func (r *Repository) UpdateDefaultVendor(ctx context.Context, id string, vendor string) (inventory.InventoryItem, error) {
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("default_vendor", vendor).Error; err != nil {
		return inventory.InventoryItem{}, err
	}
	return r.FindByID(ctx, id)
}
