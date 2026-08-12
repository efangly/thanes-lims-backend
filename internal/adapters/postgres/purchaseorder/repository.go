package purchaseorder

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/purchaseorder"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) purchaseorder.PurchaseOrder {
	return purchaseorder.PurchaseOrder{
		ID:        m.ID,
		ItemID:    m.ItemID,
		Quantity:  m.Quantity,
		Vendor:    m.Vendor,
		OrderDate: m.OrderDate,
		Status:    purchaseorder.Status(m.Status),
	}
}

func toModel(po purchaseorder.PurchaseOrder) Model {
	return Model{
		ID:        po.ID,
		ItemID:    po.ItemID,
		Quantity:  po.Quantity,
		Vendor:    po.Vendor,
		OrderDate: po.OrderDate,
		Status:    string(po.Status),
	}
}

func (r *Repository) Create(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	m := toModel(po)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (purchaseorder.PurchaseOrder, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return purchaseorder.PurchaseOrder{}, shared.ErrNotFound
	}
	if err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]purchaseorder.PurchaseOrder, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("order_date DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]purchaseorder.PurchaseOrder, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, po purchaseorder.PurchaseOrder) (purchaseorder.PurchaseOrder, error) {
	m := toModel(po)
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", po.ID).Updates(&m).Error; err != nil {
		return purchaseorder.PurchaseOrder{}, err
	}
	return r.FindByID(ctx, po.ID)
}
