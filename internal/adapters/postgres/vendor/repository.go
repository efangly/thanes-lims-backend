package vendor

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/vendor"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) vendor.Vendor {
	return vendor.Vendor{
		ID:           m.ID,
		Name:         m.Name,
		ContactName:  m.ContactName,
		ContactPhone: m.ContactPhone,
		ContactEmail: m.ContactEmail,
		Address:      m.Address,
	}
}

func toModel(v vendor.Vendor) Model {
	return Model{
		ID:           v.ID,
		Name:         v.Name,
		ContactName:  v.ContactName,
		ContactPhone: v.ContactPhone,
		ContactEmail: v.ContactEmail,
		Address:      v.Address,
	}
}

func (r *Repository) Create(ctx context.Context, v vendor.Vendor) (vendor.Vendor, error) {
	m := toModel(v)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return vendor.Vendor{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (vendor.Vendor, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return vendor.Vendor{}, shared.ErrNotFound
	}
	if err != nil {
		return vendor.Vendor{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (vendor.Vendor, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return vendor.Vendor{}, shared.ErrNotFound
	}
	if err != nil {
		return vendor.Vendor{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]vendor.Vendor, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]vendor.Vendor, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, v vendor.Vendor) (vendor.Vendor, error) {
	m := toModel(v)
	// Select every column so blanking an optional field (e.g. clearing an
	// Address) is persisted - GORM's Updates on a struct skips zero values.
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", v.ID).
		Select("name", "contact_name", "contact_phone", "contact_email", "address").
		Updates(&m).Error; err != nil {
		return vendor.Vendor{}, err
	}
	return r.FindByID(ctx, v.ID)
}
