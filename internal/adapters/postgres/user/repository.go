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
	status := user.Status(m.Status)
	if !status.Valid() {
		status = user.StatusActive
	}
	return user.User{
		ID:           m.ID,
		Name:         m.Name,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         role,
		Status:       status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// findWithRole runs cond/args against users joined to roles, so the
// resulting Model has RoleName populated for toDomain.
func (r *Repository) findWithRole(ctx context.Context, cond string, args ...any) (Model, error) {
	var m Model
	err := r.db.WithContext(ctx).
		Model(&Model{}).
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

	status := u.Status
	if status == "" {
		status = user.StatusActive
	}
	m := Model{
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       roleID,
		Status:       string(status),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return user.User{}, err
	}

	out := u
	out.ID = m.ID
	out.Status = status
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
		Model(&Model{}).
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
		Status:       string(u.Status),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	// Select the columns that a User edit is allowed to touch explicitly, so
	// a zero-value Status/Name never silently skips (struct Updates ignores
	// zero values) and email/role_id can't be smuggled in from a stale load.
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", u.ID).
		Select("name", "password_hash", "role_id", "status", "updated_at").
		Updates(&m).Error; err != nil {
		return user.User{}, err
	}
	return r.FindByID(ctx, u.ID)
}

// Retire soft-deletes the User (sets deleted_at; the row and its FK targets
// stay intact - see CONTEXT.md "Retired"). GORM's soft-delete scoping then
// hides the User from every other query. Returns shared.ErrNotFound if no
// live User has that id.
func (r *Repository) Retire(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&Model{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// CountByRole counts Users currently holding role.
func (r *Repository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Model{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ?", role.DisplayName()).
		Count(&count).Error
	return count, err
}

// CountActiveByRole counts Users holding role whose Status is active - the
// input to the last-active-admin guard shared by role change, suspend and
// retire (see ADR 0010). Retired Users are excluded by GORM's soft-delete
// scope (Model-based query).
func (r *Repository) CountActiveByRole(ctx context.Context, role user.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Model{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.status = ?", role.DisplayName(), string(user.StatusActive)).
		Count(&count).Error
	return count, err
}
