package environment

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"gorm.io/gorm"
)

type PartnerDeviceRepository struct {
	db *gorm.DB
}

func NewPartnerDeviceRepository(db *gorm.DB) *PartnerDeviceRepository {
	return &PartnerDeviceRepository{db: db}
}

func partnerDeviceToDomain(m PartnerDeviceModel) environment.PartnerDevice {
	return environment.PartnerDevice{Serial: m.Serial, Location: m.Location, Active: m.Active}
}

func (r *PartnerDeviceRepository) Create(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error) {
	m := PartnerDeviceModel{Serial: d.Serial, Location: d.Location, Active: d.Active}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return environment.PartnerDevice{}, err
	}
	return partnerDeviceToDomain(m), nil
}

func (r *PartnerDeviceRepository) Update(ctx context.Context, d environment.PartnerDevice) (environment.PartnerDevice, error) {
	m := PartnerDeviceModel{Serial: d.Serial, Location: d.Location, Active: d.Active}
	if err := r.db.WithContext(ctx).Model(&PartnerDeviceModel{}).Where("serial = ?", d.Serial).Updates(map[string]any{
		"location": m.Location,
		"active":   m.Active,
	}).Error; err != nil {
		return environment.PartnerDevice{}, err
	}
	return d, nil
}

func (r *PartnerDeviceRepository) FindBySerial(ctx context.Context, serial string) (environment.PartnerDevice, error) {
	var m PartnerDeviceModel
	err := r.db.WithContext(ctx).First(&m, "serial = ?", serial).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return environment.PartnerDevice{}, shared.ErrNotFound
	}
	if err != nil {
		return environment.PartnerDevice{}, err
	}
	return partnerDeviceToDomain(m), nil
}

func (r *PartnerDeviceRepository) List(ctx context.Context) ([]environment.PartnerDevice, error) {
	var models []PartnerDeviceModel
	if err := r.db.WithContext(ctx).Order("serial").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]environment.PartnerDevice, len(models))
	for i, m := range models {
		out[i] = partnerDeviceToDomain(m)
	}
	return out, nil
}
