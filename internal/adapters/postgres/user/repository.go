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

// knownRoles is every Role this Go enum supports - see
// internal/domain/user/role.go and migrations/000018_create_roles_table.up.sql.
var knownRoles = []user.Role{
	user.RoleAdmin,
	user.RoleLabManager,
	user.RoleQA,
	user.RoleScientist,
	user.RoleGeneral,
}

// roleByDisplayName maps a roles.name row back to the Role enum, built from
// knownRoles so there's exactly one place (Role.DisplayName) that owns the
// enum<->name mapping.
var roleByDisplayName = func() map[string]user.Role {
	m := make(map[string]user.Role, len(knownRoles))
	for _, r := range knownRoles {
		m[r.DisplayName()] = r
	}
	return m
}()

func toDomain(m Model) user.User {
	role, ok := roleByDisplayName[m.RoleName]
	if !ok {
		role = user.Role(m.RoleName)
	}
	return user.User{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// findWithRole runs cond/args against users joined to roles, so the
// resulting Model has RoleName populated for toDomain.
func (r *Repository) findWithRole(ctx context.Context, cond string, args ...any) (Model, error) {
	var m Model
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where(cond, args...).
		First(&m).Error
	return m, err
}

// roleIDByRole resolves the roles.id a Role enum value corresponds to -
// needed to persist users.role_id on Create/Update, since the domain User
// only carries the Role enum, not the numeric FK.
func (r *Repository) roleIDByRole(ctx context.Context, role user.Role) (int64, error) {
	name := role.DisplayName()
	var id int64
	err := r.db.WithContext(ctx).Table("roles").Select("id").Where("name = ?", name).Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, shared.ErrValidation
	}
	return id, nil
}

func (r *Repository) Create(ctx context.Context, u user.User) (user.User, error) {
	roleID, err := r.roleIDByRole(ctx, u.Role)
	if err != nil {
		return user.User{}, err
	}

	m := Model{
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       roleID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return user.User{}, err
	}

	out := u
	out.ID = m.ID
	return out, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (user.User, error) {
	m, err := r.findWithRole(ctx, "users.id = ?", id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user.User{}, shared.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return toDomain(m), nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	m, err := r.findWithRole(ctx, "users.email = ?", email)
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
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = users.role_id").
		Order("users.id").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	users := make([]user.User, len(models))
	for i, m := range models {
		users[i] = toDomain(m)
	}
	return users, nil
}

func (r *Repository) Update(ctx context.Context, u user.User) (user.User, error) {
	roleID, err := r.roleIDByRole(ctx, u.Role)
	if err != nil {
		return user.User{}, err
	}

	m := Model{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       roleID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", u.ID).Updates(&m).Error; err != nil {
		return user.User{}, err
	}
	return r.FindByID(ctx, u.ID)
}

// CountByRole counts Users currently holding role - used by the last-admin
// guard in the UpdateUser use case (see ADR 0002).
func (r *Repository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", role.DisplayName()).
		Count(&count).Error
	return count, err
}
