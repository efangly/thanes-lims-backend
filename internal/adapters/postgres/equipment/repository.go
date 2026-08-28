package equipment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/equipment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) equipment.Equipment {
	return equipment.Equipment{
		ID:                 m.ID,
		Name:               m.Name,
		TypeCode:           m.TypeCode,
		LastCalibratedAt:   m.LastCalibratedAt,
		NextCalibrationDue: m.NextCalibrationDue,
		UsageHours:         m.UsageHours,
		SerialNumber:       m.SerialNumber,
		Category:           m.Category,
		Manufacturer:       m.Manufacturer,
		Model:              m.Model,
		InstallationDate:   m.InstallationDate,
		VendorID:           m.VendorID,
		LocationID:         m.LocationID,
	}
}

func toModel(e equipment.Equipment) Model {
	return Model{
		ID:                 e.ID,
		Name:               e.Name,
		TypeCode:           e.TypeCode,
		LastCalibratedAt:   e.LastCalibratedAt,
		NextCalibrationDue: e.NextCalibrationDue,
		UsageHours:         e.UsageHours,
		SerialNumber:       e.SerialNumber,
		Category:           e.Category,
		Manufacturer:       e.Manufacturer,
		Model:              e.Model,
		InstallationDate:   e.InstallationDate,
		VendorID:           e.VendorID,
		LocationID:         e.LocationID,
	}
}

func (r *Repository) Create(ctx context.Context, e equipment.Equipment) (equipment.Equipment, error) {
	m := toModel(e)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return equipment.Equipment{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (equipment.Equipment, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return equipment.Equipment{}, shared.ErrNotFound
	}
	if err != nil {
		return equipment.Equipment{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]equipment.Equipment, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]equipment.Equipment, len(models))
	for i, m := range models {
		out[i] = toDomain(m)
	}
	return out, nil
}

// Update uses Save (not Updates) so cleared optional fields - an emptied
// SerialNumber, a removed VendorID/LocationID - actually persist; GORM's
// Updates skips zero/nil values on a struct arg. Callers always load the
// full Equipment via FindByID first and overlay the changed fields.
func (r *Repository) Update(ctx context.Context, e equipment.Equipment) (equipment.Equipment, error) {
	m := toModel(e)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return equipment.Equipment{}, err
	}
	return r.FindByID(ctx, e.ID)
}
