package user

import (
	"context"
	"errors"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/efangly/thanes-lims-backend/internal/domain/user"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func toDomain(m Model) user.User {
	return user.User{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         user.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toModel(u user.User) Model {
	return Model{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (r *Repository) Create(ctx context.Context, u user.User) (user.User, error) {
	m := toModel(u)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return user.User{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (user.User, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.User{}, shared.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	var m Model
	err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.User{}, shared.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]user.User, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Order("id").Find(&models).Error; err != nil {
		return nil, err
	}
	users := make([]user.User, len(models))
	for i, m := range models {
		users[i] = toDomain(m)
	}
	return users, nil
}

func (r *Repository) Update(ctx context.Context, u user.User) (user.User, error) {
	m := toModel(u)
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", u.ID).Updates(&m).Error; err != nil {
		return user.User{}, err
	}
	return r.FindByID(ctx, u.ID)
}
